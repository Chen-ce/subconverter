package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/Chen-ce/subconverter/config"
	"github.com/Chen-ce/subconverter/exporter"
	"github.com/Chen-ce/subconverter/fetcher"
	"github.com/Chen-ce/subconverter/merger"
	"github.com/Chen-ce/subconverter/pkg/hash"
	"github.com/Chen-ce/subconverter/pkg/logger"
	"github.com/Chen-ce/subconverter/parser"
	"github.com/Chen-ce/subconverter/templates"
)

// Convert 订阅转换处理器
func Convert(c *gin.Context) {
	// 获取并处理参数
	target := c.Query("target")
	urlParam := c.Query("url")
	nodeParams := c.QueryArray("node")
	configName := c.DefaultQuery("config", "")
	include := c.Query("include")
	exclude := c.Query("exclude")
	ver := c.Query("ver")

	executeConversion(c, target, urlParam, nodeParams, configName, include, exclude, ver)
}

// PureConvert 供外部程序调用的 JSON 转换接口 (POST)
func PureConvert(c *gin.Context) {
	var req struct {
		Target  string   `json:"target" binding:"required"`
		URL     string   `json:"url"`
		Nodes   []string `json:"nodes"`
		Config  string   `json:"config"`
		Include string   `json:"include"`
		Exclude string   `json:"exclude"`
		Ver     string   `json:"ver"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body", "details": err.Error()})
		return
	}

	executeConversion(c, req.Target, req.URL, req.Nodes, req.Config, req.Include, req.Exclude, req.Ver)
}

// executeConversion 执行实际的转换逻辑
func executeConversion(c *gin.Context, target, urlParam string, nodeParams []string, configName, include, exclude, ver string) {
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

	// 1. 尝试从存储加载转换好的缓存 (Tier 1: Result Cache)
	// 如果是通过 ID 调用的，我们可以检查 ID/export_{target}.txt
	id := c.Param("id")
	if id != "" {
		cacheFile := fmt.Sprintf("export_%s.txt", target)
		if data, err := configStorage.LoadFile(id, cacheFile); err == nil {
			logger.Info("Serving from result cache", "id", id, "target", target)
			// TODO: 根据 target 设置适当的 Content-Type
			c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
			return
		}
	}

	// 2. 开启 Keep-Alive 协程 (可选，对于可能会超时的长连接很有用)
	// 通过 Flush 空格保持连接
	stopKeepAlive := make(chan bool)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 发送一个空格，防止负载均衡或代理超时
				// 注意：这会改变响应流，对于非流式响应（如 JSON）不适用
				// 但订阅内容通常是文本/YAML，前面的空格通常会被忽略或可以被 Parser 容忍
				// 更好的方案是注释，但不同格式注释不同
				c.Writer.Write([]byte(" "))
				c.Writer.Flush()
			case <-stopKeepAlive:
				return
			}
		}
	}()
	defer close(stopKeepAlive)

	// 创建合并器
	m := merger.NewMerger()

	// 处理订阅 URL
	if urlParam != "" {
		// 分割多个订阅（用 | 分隔）
		urls := strings.Split(urlParam, "|")
		
		// 过滤空 URL
		var validURLs []string
		for _, subURL := range urls {
			subURL = strings.TrimSpace(subURL)
			if subURL != "" {
				validURLs = append(validURLs, subURL)
			}
		}

		if len(validURLs) > 0 {
			// 并发获取所有订阅
			type fetchResult struct {
				nodes []*parser.Node
				err   error
				url   string
			}

			results := make(chan fetchResult, len(validURLs))
			
			// 启动并发获取
			for _, subURL := range validURLs {
				go func(url string) {
					// 尝试从存储加载内容缓存 (Tier 2: Content Cache)
					urlHash := hash.URLHash(url)
					if id != "" {
						cacheFile := fmt.Sprintf("cache_%s.json", urlHash)
						if data, err := configStorage.LoadFile(id, cacheFile); err == nil {
							var nodes []*parser.Node
							if json.Unmarshal(data, &nodes) == nil {
								logger.Debug("Loaded subscription from content cache", "hash", urlHash, "url", url)
								results <- fetchResult{nodes: nodes, url: url}
								return
							}
						}
					}

					f := fetcher.NewFetcher()
					content, err := f.Fetch(url)
					if err != nil {
						results <- fetchResult{err: err, url: url}
						return
					}

					nodes, err := parseSubscription(content)
					if err == nil && id != "" {
						// 异步保存到缓存
						nodeData, _ := json.Marshal(nodes)
						go configStorage.SaveFile(id, fmt.Sprintf("cache_%s.json", urlHash), nodeData)
					}
					results <- fetchResult{nodes: nodes, err: err, url: url}
				}(subURL)
			}

			// 收集结果
			var fetchErrors []string
			for i := 0; i < len(validURLs); i++ {
				result := <-results
				if result.err != nil {
					fetchErrors = append(fetchErrors, fmt.Sprintf("%s: %v", result.url, result.err))
					continue
				}
				m.Add(result.nodes)
			}

			// 如果所有订阅都失败了，返回错误
			if len(fetchErrors) == len(validURLs) {
				c.JSON(http.StatusBadGateway, gin.H{
					"error":   "failed to fetch all subscriptions",
					"details": fetchErrors,
				})
				return
			}

			// 如果部分失败，记录警告但继续处理
			if len(fetchErrors) > 0 {
				// 可以选择在响应头中添加警告信息
				c.Header("X-Fetch-Warnings", fmt.Sprintf("%d/%d subscriptions failed", len(fetchErrors), len(validURLs)))
			}
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

	// 解析 target 参数（兼容旧格式 surge&ver=X）
	target = strings.ToLower(target)

	// 获取版本参数（用于 Surge）
	version := 4 // 默认版本

	// 兼容旧格式：target=surge&ver=X
	if strings.Contains(target, "&ver=") {
		parts := strings.Split(target, "&")
		target = parts[0] // 提取 surge
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "ver=") {
				fmt.Sscanf(part, "ver=%d", &version)
			}
		}
	}

	// 设置版本
	if ver != "" {
		fmt.Sscanf(ver, "%d", &version)
	}

	switch target {
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

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     fmt.Sprintf("unsupported target format: %s", target),
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
	logger.Info("Conversion successful", "target", target, "nodes", len(nodes))
	
	// 异步保存结果缓存 (Tier 1)
	if id != "" {
		go configStorage.SaveFile(id, fmt.Sprintf("export_%s.txt", target), []byte(result))
	}

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
