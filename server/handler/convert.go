package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/Chen-ce/subconverter/config"
	"github.com/Chen-ce/subconverter/exporter"
	"github.com/Chen-ce/subconverter/fetcher"
	"github.com/Chen-ce/subconverter/merger"
	"github.com/Chen-ce/subconverter/parser"
	"github.com/Chen-ce/subconverter/templates"
)

// Convert 订阅转换处理器
func Convert(c *gin.Context) {
	// 获取参数
	target := c.Query("target")
	urlParam := c.Query("url")
	nodeParams := c.QueryArray("node")
	configName := c.DefaultQuery("config", "default")
	include := c.Query("include")
	exclude := c.Query("exclude")
	
	// 验证必需参数
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing required parameter: target",
		})
		return
	}
	
	if urlParam == "" && len(nodeParams) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one of 'url' or 'node' parameter is required",
		})
		return
	}
	
	// 创建合并器
	m := merger.NewMerger()
	
	// 处理订阅 URL
	if urlParam != "" {
		// 解码 URL
		decodedURL, err := url.QueryUnescape(urlParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid url parameter",
			})
			return
		}
		
		// 分割多个订阅（用 | 分隔）
		urls := strings.Split(decodedURL, "|")
		
		f := fetcher.NewFetcher()
		for _, subURL := range urls {
			subURL = strings.TrimSpace(subURL)
			if subURL == "" {
				continue
			}
			
			// 获取订阅内容
			content, err := f.Fetch(subURL)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"error": fmt.Sprintf("failed to fetch subscription: %v", err),
				})
				return
			}
			
			// 解析订阅
			nodes, err := parseSubscription(content)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("failed to parse subscription: %v", err),
				})
				return
			}
			
			m.Add(nodes)
		}
	}
	
	// 处理单独节点
	for _, nodeURI := range nodeParams {
		node, err := parser.ParseNodeURI(nodeURI)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("failed to parse node: %v", err),
			})
			return
		}
		m.Add([]*parser.Node{node})
	}
	
	// 合并节点
	nodes := m.Merge()
	
	// 应用过滤
	if include != "" || exclude != "" {
		nodes = filterNodes(nodes, include, exclude)
	}
	
	if len(nodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no valid nodes after filtering",
		})
		return
	}
	
	// 导出
	var result string
	var err error
	var contentType string
	
	// 解析 target 参数（处理 surge&ver=X 格式）
	targetParts := strings.Split(target, "&")
	baseTarget := targetParts[0]
	
	// 获取版本参数（用于 Surge）
	version := 4 // 默认版本
	for _, part := range targetParts[1:] {
		if strings.HasPrefix(part, "ver=") {
			fmt.Sscanf(part, "ver=%d", &version)
		}
	}
	
	switch baseTarget {
	case "clash":
		// 使用模板
		cfg := config.Get()
		tmplMgr := templates.NewManager(cfg.Templates.Dir)
		
		if configName == "" {
			configName = cfg.Clash.DefaultRules
		}
		
		clashConfig, err := tmplMgr.ApplyToNodes(configName, nodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to apply template: %v", err),
			})
			return
		}
		
		// 导出为 YAML
		exp := exporter.NewClashExporter()
		result, err = exp.ExportWithConfig(clashConfig)
		contentType = "text/yaml; charset=utf-8"
		
	case "singbox":
		// Sing-box 导出
		exp := exporter.NewSingboxExporter()
		result, err = exp.Export(nodes)
		contentType = "application/json; charset=utf-8"
		
	case "surge":
		// Surge 导出
		exp := exporter.NewSurgeExporter(version)
		result, err = exp.Export(nodes)
		contentType = "text/plain; charset=utf-8"
		
	case "surfboard":
		// Surfboard 使用 Surge 4 格式
		exp := exporter.NewSurgeExporter(4)
		result, err = exp.Export(nodes)
		contentType = "text/plain; charset=utf-8"
		
	case "v2ray", "ss", "ssr", "mixed":
		// 通用格式（Base64 URI）
		exp := exporter.NewBase64Exporter()
		result, err = exp.Export(nodes)
		contentType = "text/plain; charset=utf-8"
		
	case "quanx":
		// Quantumult X 导出
		exp := exporter.NewQuantumultXExporter()
		result, err = exp.Export(nodes)
		contentType = "text/plain; charset=utf-8"
		
	case "loon":
		// Loon 导出
		exp := exporter.NewLoonExporter()
		result, err = exp.Export(nodes)
		contentType = "text/plain; charset=utf-8"
		
	case "quan":
		// Quantumult 暂未实现（使用 Quantumult X 格式）
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Quantumult format is not implemented yet",
			"message": "Please use 'quanx' (Quantumult X) instead",
			"supported": []string{"clash", "singbox", "surge&ver=4", "quanx", "loon", "v2ray", "ss", "ssr", "mixed"},
		})
		return
		
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported target format: %s", target),
			"supported": []string{"clash", "singbox", "surge&ver=4", "surge&ver=3", "surfboard", "quanx", "loon", "v2ray", "ss", "ssr", "mixed"},
		})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to export: %v", err),
		})
		return
	}
	
	// 返回结果
	c.Data(http.StatusOK, contentType, []byte(result))
}

// parseSubscription 解析订阅内容
func parseSubscription(content string) ([]*parser.Node, error) {
	// 先尝试 Base64
	base64Parser := parser.NewBase64Parser()
	nodes, err := base64Parser.Parse(content)
	if err == nil {
		return nodes, nil
	}
	
	// 再尝试 Clash
	clashParser := parser.NewClashParser()
	nodes, err = clashParser.Parse(content)
	if err == nil {
		return nodes, nil
	}
	
	return nil, fmt.Errorf("failed to parse subscription")
}

// filterNodes 过滤节点
func filterNodes(nodes []*parser.Node, include, exclude string) []*parser.Node {
	var result []*parser.Node
	
	includeKeywords := strings.Split(include, "|")
	excludeKeywords := strings.Split(exclude, "|")
	
	for _, node := range nodes {
		// 检查排除条件
		if exclude != "" {
			excluded := false
			for _, keyword := range excludeKeywords {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" && strings.Contains(node.Name, keyword) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		// 检查包含条件
		if include != "" {
			included := false
			for _, keyword := range includeKeywords {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" && strings.Contains(node.Name, keyword) {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		result = append(result, node)
	}
	
	return result
}
