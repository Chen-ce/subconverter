package exporter

import (
	"fmt"
	"gopkg.in/yaml.v3"
	
	"github.com/Chen-ce/subconverter/parser"
)

// ClashExporter Clash 格式导出器
type ClashExporter struct {
	template string
}

// NewClashExporter 创建 Clash 导出器
func NewClashExporter() *ClashExporter {
	return &ClashExporter{
		template: defaultClashTemplate,
	}
}

// SetTemplate 设置自定义模板
func (e *ClashExporter) SetTemplate(template string) {
	e.template = template
}

// Export 导出为 Clash YAML 格式
func (e *ClashExporter) Export(nodes []*parser.Node) (string, error) {
	if len(nodes) == 0 {
		return "", fmt.Errorf("no nodes to export")
	}
	
	// 转换节点为 Clash 代理格式
	var proxies []map[string]interface{}
	var proxyNames []string
	
	for _, node := range nodes {
		proxy, err := nodeToClashProxy(node)
		if err != nil {
			fmt.Printf("Warning: failed to convert node to clash proxy: %v\n", err)
			continue
		}
		proxies = append(proxies, proxy)
		proxyNames = append(proxyNames, node.Name)
	}
	
	if len(proxies) == 0 {
		return "", fmt.Errorf("no valid nodes to export")
	}
	
	// 构建完整的 Clash 配置
	config := map[string]interface{}{
		"port":               7890,
		"socks-port":         7891,
		"allow-lan":          false,
		"mode":               "rule",
		"log-level":          "info",
		"external-controller": "127.0.0.1:9090",
		"proxies":            proxies,
		"proxy-groups": []map[string]interface{}{
			{
				"name":    "PROXY",
				"type":    "select",
				"proxies": append([]string{"auto"}, proxyNames...),
			},
			{
				"name":    "auto",
				"type":    "url-test",
				"proxies": proxyNames,
				"url":     "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
		},
		"rules": []string{
			"DOMAIN-SUFFIX,google.com,PROXY",
			"DOMAIN-KEYWORD,google,PROXY",
			"DOMAIN,google.com,PROXY",
			"DOMAIN-SUFFIX,ad.com,REJECT",
			"GEOIP,CN,DIRECT",
			"MATCH,PROXY",
		},
	}
	
	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal clash config: %w", err)
	}
	
	return string(yamlBytes), nil
}

// nodeToClashProxy 将节点转换为 Clash 代理格式
func nodeToClashProxy(node *parser.Node) (map[string]interface{}, error) {
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
		
		// WebSocket 配置
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
		
		// gRPC 配置
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
		
		// WebSocket 配置
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
		
		// gRPC 配置
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
		
		if node.Network != "" && node.Network != "tcp" {
			proxy["network"] = node.Network
		}
		
		if node.SNI != "" {
			proxy["sni"] = node.SNI
		}
		
		if node.SkipCertVerify {
			proxy["skip-cert-verify"] = true
		}
		
		// WebSocket 配置
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
		
		// gRPC 配置
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
		proxy["cipher"] = node.Method
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
		proxy["cipher"] = node.Method
		proxy["password"] = node.Password
		proxy["protocol"] = node.Protocol
		proxy["obfs"] = node.Obfs
		
		if node.ProtocolParam != "" {
			proxy["protocol-param"] = node.ProtocolParam
		}
		if node.ObfsParam != "" {
			proxy["obfs-param"] = node.ObfsParam
		}
		
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", node.Type)
	}
	
	return proxy, nil
}

// defaultClashTemplate 默认 Clash 配置模板
const defaultClashTemplate = `
port: 7890
socks-port: 7891
allow-lan: false
mode: rule
log-level: info
external-controller: 127.0.0.1:9090
`
