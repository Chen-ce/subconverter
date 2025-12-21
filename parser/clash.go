package parser

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// ClashParser Clash 配置解析器
type ClashParser struct{}

// NewClashParser 创建 Clash 解析器
func NewClashParser() *ClashParser {
	return &ClashParser{}
}

// ClashConfig Clash 配置结构
type ClashConfig struct {
	Proxies []map[string]interface{} `yaml:"proxies"`
}

// Parse 解析 Clash YAML 配置
func (p *ClashParser) Parse(content string) ([]*Node, error) {
	var config ClashConfig
	
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse clash config: %w", err)
	}
	
	if len(config.Proxies) == 0 {
		return nil, fmt.Errorf("no proxies found in clash config")
	}
	
	var nodes []*Node
	for _, proxy := range config.Proxies {
		node, err := parseClashProxy(proxy)
		if err != nil {
			fmt.Printf("Warning: failed to parse clash proxy: %v\n", err)
			continue
		}
		nodes = append(nodes, node)
	}
	
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid nodes parsed from clash config")
	}
	
	return nodes, nil
}

// parseClashProxy 解析单个 Clash 代理配置
func parseClashProxy(proxy map[string]interface{}) (*Node, error) {
	proxyType, ok := proxy["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing proxy type")
	}
	
	name, _ := proxy["name"].(string)
	server, _ := proxy["server"].(string)
	port, _ := proxy["port"].(int)
	
	node := &Node{
		Name:   name,
		Server: server,
		Port:   port,
	}
	
	switch proxyType {
	case "vmess":
		node.Type = ProxyTypeVMess
		node.UUID, _ = proxy["uuid"].(string)
		node.AlterId, _ = proxy["alterId"].(int)
		node.Cipher, _ = proxy["cipher"].(string)
		
		// 传输层配置
		node.Network, _ = proxy["network"].(string)
		node.TLS, _ = proxy["tls"].(bool)
		node.SNI, _ = proxy["servername"].(string)
		node.SkipCertVerify, _ = proxy["skip-cert-verify"].(bool)
		
		// WebSocket
		if wsOpts, ok := proxy["ws-opts"].(map[string]interface{}); ok {
			node.WSPath, _ = wsOpts["path"].(string)
			if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
				node.WSHeaders = make(map[string]string)
				for k, v := range headers {
					if strVal, ok := v.(string); ok {
						node.WSHeaders[k] = strVal
					}
				}
			}
		}
		
		// gRPC
		if grpcOpts, ok := proxy["grpc-opts"].(map[string]interface{}); ok {
			node.GRPCServiceName, _ = grpcOpts["grpc-service-name"].(string)
		}
		
	case "vless":
		node.Type = ProxyTypeVLESS
		node.UUID, _ = proxy["uuid"].(string)
		node.Flow, _ = proxy["flow"].(string)
		
		// 传输层配置
		node.Network, _ = proxy["network"].(string)
		node.TLS, _ = proxy["tls"].(bool)
		node.SNI, _ = proxy["servername"].(string)
		node.SkipCertVerify, _ = proxy["skip-cert-verify"].(bool)
		
		// WebSocket
		if wsOpts, ok := proxy["ws-opts"].(map[string]interface{}); ok {
			node.WSPath, _ = wsOpts["path"].(string)
			if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
				node.WSHeaders = make(map[string]string)
				for k, v := range headers {
					if strVal, ok := v.(string); ok {
						node.WSHeaders[k] = strVal
					}
				}
			}
		}
		
		// gRPC
		if grpcOpts, ok := proxy["grpc-opts"].(map[string]interface{}); ok {
			node.GRPCServiceName, _ = grpcOpts["grpc-service-name"].(string)
		}
		
	case "trojan":
		node.Type = ProxyTypeTrojan
		node.Password, _ = proxy["password"].(string)
		node.TLS = true
		
		node.Network, _ = proxy["network"].(string)
		node.SNI, _ = proxy["sni"].(string)
		node.SkipCertVerify, _ = proxy["skip-cert-verify"].(bool)
		
		// WebSocket
		if wsOpts, ok := proxy["ws-opts"].(map[string]interface{}); ok {
			node.WSPath, _ = wsOpts["path"].(string)
			if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
				node.WSHeaders = make(map[string]string)
				for k, v := range headers {
					if strVal, ok := v.(string); ok {
						node.WSHeaders[k] = strVal
					}
				}
			}
		}
		
		// gRPC
		if grpcOpts, ok := proxy["grpc-opts"].(map[string]interface{}); ok {
			node.GRPCServiceName, _ = grpcOpts["grpc-service-name"].(string)
		}
		
	case "ss":
		node.Type = ProxyTypeShadowsocks
		node.Password, _ = proxy["password"].(string)
		node.Cipher, _ = proxy["cipher"].(string)
		node.Method = node.Cipher // SS 中 cipher 和 method 是同一个概念
		
		node.UDP, _ = proxy["udp"].(bool)
		
		// 插件
		if plugin, ok := proxy["plugin"].(string); ok {
			node.Plugin = plugin
			if pluginOpts, ok := proxy["plugin-opts"].(map[string]interface{}); ok {
				node.PluginOpts = pluginOpts
			}
		}
		
	case "ssr":
		node.Type = ProxyTypeShadowsocksR
		node.Password, _ = proxy["password"].(string)
		node.Cipher, _ = proxy["cipher"].(string)
		node.Method = node.Cipher
		node.Protocol, _ = proxy["protocol"].(string)
		node.ProtocolParam, _ = proxy["protocol-param"].(string)
		node.Obfs, _ = proxy["obfs"].(string)
		node.ObfsParam, _ = proxy["obfs-param"].(string)
		
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxyType)
	}
	
	return node, nil
}
