package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/Chen-ce/subconverter/parser"
	"gopkg.in/yaml.v3"
)

// Template 规则模板
type Template struct {
	Port               int                      `yaml:"port"`
	SocksPort          int                      `yaml:"socks-port"`
	AllowLan           bool                     `yaml:"allow-lan"`
	Mode               string                   `yaml:"mode"`
	LogLevel           string                   `yaml:"log-level"`
	ExternalController string                   `yaml:"external-controller"`
	DNS                map[string]interface{}   `yaml:"dns,omitempty"`
	ProxyGroups        []map[string]interface{} `yaml:"proxy-groups"`
	Rules              []string                 `yaml:"rules"`
}

// Manager 模板管理器
type Manager struct {
	templatesDir string
	templates    map[string]*Template
}

// NewManager 创建模板管理器
func NewManager(templatesDir string) *Manager {
	return &Manager{
		templatesDir: templatesDir,
		templates:    make(map[string]*Template),
	}
}

// Load 加载模板
func (m *Manager) Load(name string) (*Template, error) {
	// 检查缓存
	if tmpl, ok := m.templates[name]; ok {
		return tmpl, nil
	}
	
	// 读取模板文件
	path := filepath.Join(m.templatesDir, "rules", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}
	
	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	// 缓存模板
	m.templates[name] = &tmpl
	return &tmpl, nil
}

// List 列出所有可用模板
func (m *Manager) List() ([]string, error) {
	rulesDir := filepath.Join(m.templatesDir, "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}
	
	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			name := strings.TrimSuffix(entry.Name(), ".yaml")
			templates = append(templates, name)
		}
	}
	
	return templates, nil
}

// ApplyToNodes 应用模板到节点列表
func (m *Manager) ApplyToNodes(templateName string, nodes []*parser.Node) (map[string]interface{}, error) {
	tmpl, err := m.Load(templateName)
	if err != nil {
		return nil, err
	}
	
	// 构建代理列表
	var proxies []map[string]interface{}
	var proxyNames []string
	
	for _, node := range nodes {
		proxy := nodeToClashProxy(node)
		proxies = append(proxies, proxy)
		proxyNames = append(proxyNames, node.Name)
	}
	
	// 处理代理组
	processedGroups := processProxyGroups(tmpl.ProxyGroups, proxyNames, nodes)
	
	// 构建完整配置
	config := map[string]interface{}{
		"port":                tmpl.Port,
		"socks-port":          tmpl.SocksPort,
		"allow-lan":           tmpl.AllowLan,
		"mode":                tmpl.Mode,
		"log-level":           tmpl.LogLevel,
		"external-controller": tmpl.ExternalController,
		"proxies":             proxies,
		"proxy-groups":        processedGroups,
		"rules":               tmpl.Rules,
	}
	
	if tmpl.DNS != nil {
		config["dns"] = tmpl.DNS
	}
	
	return config, nil
}

// processProxyGroups 处理代理组模板
func processProxyGroups(groups []map[string]interface{}, proxyNames []string, nodes []*parser.Node) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, group := range groups {
		newGroup := make(map[string]interface{})
		for k, v := range group {
			newGroup[k] = v
		}
		
		// 处理 proxies 字段
		if proxies, ok := group["proxies"].([]interface{}); ok {
			var newProxies []string
			for _, p := range proxies {
				if str, ok := p.(string); ok {
					if strings.Contains(str, "{{proxies}}") {
						// 替换为所有节点名称
						newProxies = append(newProxies, proxyNames...)
					} else if strings.Contains(str, "{{proxies|filter:") {
						// 过滤节点
						filter := extractFilter(str)
						filtered := filterNodes(nodes, filter)
						newProxies = append(newProxies, filtered...)
					} else {
						newProxies = append(newProxies, str)
					}
				}
			}
			
			// 如果过滤后没有节点，跳过这个代理组
			if len(newProxies) == 0 {
				continue
			}
			
			newGroup["proxies"] = newProxies
		}
		
		result = append(result, newGroup)
	}
	
	return result
}

// extractFilter 提取过滤条件
func extractFilter(str string) []string {
	// {{proxies|filter:香港|HK|Hong Kong}} -> [香港, HK, Hong Kong]
	start := strings.Index(str, "filter:")
	if start == -1 {
		return nil
	}
	
	start += 7 // len("filter:")
	end := strings.Index(str[start:], "}}")
	if end == -1 {
		return nil
	}
	
	filterStr := str[start : start+end]
	return strings.Split(filterStr, "|")
}

// filterNodes 根据关键词过滤节点
func filterNodes(nodes []*parser.Node, keywords []string) []string {
	var result []string
	
	for _, node := range nodes {
		for _, keyword := range keywords {
			if strings.Contains(node.Name, keyword) {
				result = append(result, node.Name)
				break
			}
		}
	}
	
	return result
}

// nodeToClashProxy 将节点转换为 Clash 代理格式（简化版）
func nodeToClashProxy(node *parser.Node) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":   node.Name,
		"server": node.Server,
		"port":   node.Port,
	}
	
	switch node.Type {
	case parser.ProxyTypeVMess:
		proxy["type"] = "vmess"
		proxy["uuid"] = node.UUID
		proxy["alterId"] = node.AlterId
		proxy["cipher"] = "auto"
		if node.Cipher != "" {
			proxy["cipher"] = node.Cipher
		}
		if node.Network != "" {
			proxy["network"] = node.Network
		}
		if node.TLS {
			proxy["tls"] = true
			if node.SNI != "" {
				proxy["servername"] = node.SNI
			}
			if node.SkipCertVerify {
				proxy["skip-cert-verify"] = true
			}
		}
		if node.Network == "ws" {
			wsOpts := map[string]interface{}{}
			if node.WSPath != "" {
				wsOpts["path"] = node.WSPath
			}
			if len(node.WSHeaders) > 0 {
				wsOpts["headers"] = node.WSHeaders
			}
			if len(wsOpts) > 0 {
				proxy["ws-opts"] = wsOpts
			}
		}
		if node.Network == "grpc" {
			grpcOpts := map[string]interface{}{}
			if node.GRPCServiceName != "" {
				grpcOpts["grpc-service-name"] = node.GRPCServiceName
			}
			if len(grpcOpts) > 0 {
				proxy["grpc-opts"] = grpcOpts
			}
		}
		
	case parser.ProxyTypeVLESS:
		proxy["type"] = "vless"
		proxy["uuid"] = node.UUID
		if node.Flow != "" {
			proxy["flow"] = node.Flow
		}
		if node.Network != "" {
			proxy["network"] = node.Network
		}
		if node.TLS {
			proxy["tls"] = true
			if node.SNI != "" {
				proxy["servername"] = node.SNI
			}
			if node.SkipCertVerify {
				proxy["skip-cert-verify"] = true
			}
		}
		if node.Network == "ws" {
			wsOpts := map[string]interface{}{}
			if node.WSPath != "" {
				wsOpts["path"] = node.WSPath
			}
			if len(node.WSHeaders) > 0 {
				wsOpts["headers"] = node.WSHeaders
			}
			if len(wsOpts) > 0 {
				proxy["ws-opts"] = wsOpts
			}
		}
		if node.Network == "grpc" {
			grpcOpts := map[string]interface{}{}
			if node.GRPCServiceName != "" {
				grpcOpts["grpc-service-name"] = node.GRPCServiceName
			}
			if len(grpcOpts) > 0 {
				proxy["grpc-opts"] = grpcOpts
			}
		}
		
	case parser.ProxyTypeTrojan:
		proxy["type"] = "trojan"
		proxy["password"] = node.Password
		if node.SNI != "" {
			proxy["sni"] = node.SNI
		}
		if node.SkipCertVerify {
			proxy["skip-cert-verify"] = true
		}
		if node.Network != "" && node.Network != "tcp" {
			proxy["network"] = node.Network
		}
		if node.Network == "ws" {
			wsOpts := map[string]interface{}{}
			if node.WSPath != "" {
				wsOpts["path"] = node.WSPath
			}
			if len(node.WSHeaders) > 0 {
				wsOpts["headers"] = node.WSHeaders
			}
			if len(wsOpts) > 0 {
				proxy["ws-opts"] = wsOpts
			}
		}
		if node.Network == "grpc" {
			grpcOpts := map[string]interface{}{}
			if node.GRPCServiceName != "" {
				grpcOpts["grpc-service-name"] = node.GRPCServiceName
			}
			if len(grpcOpts) > 0 {
				proxy["grpc-opts"] = grpcOpts
			}
		}
		
	case parser.ProxyTypeShadowsocks:
		proxy["type"] = "ss"
		method := node.Method
		if method == "" {
			method = node.Cipher
		}
		proxy["cipher"] = method
		proxy["password"] = node.Password
		if node.UDP {
			proxy["udp"] = true
		}
		if node.Plugin != "" {
			proxy["plugin"] = node.Plugin
			if len(node.PluginOpts) > 0 {
				proxy["plugin-opts"] = node.PluginOpts
			}
		}
		
	case parser.ProxyTypeShadowsocksR:
		proxy["type"] = "ssr"
		method := node.Method
		if method == "" {
			method = node.Cipher
		}
		proxy["cipher"] = method
		proxy["password"] = node.Password
		proxy["protocol"] = node.Protocol
		proxy["obfs"] = node.Obfs
		if node.ProtocolParam != "" {
			proxy["protocol-param"] = node.ProtocolParam
		}
		if node.ObfsParam != "" {
			proxy["obfs-param"] = node.ObfsParam
		}
	}
	
	return proxy
}
