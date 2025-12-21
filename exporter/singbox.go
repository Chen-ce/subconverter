package exporter

import (
	"encoding/json"
	"fmt"
	
	"github.com/Chen-ce/subconverter/parser"
)

// SingboxExporter Sing-box 配置导出器
type SingboxExporter struct{}

// NewSingboxExporter 创建 Sing-box 导出器
func NewSingboxExporter() *SingboxExporter {
	return &SingboxExporter{}
}

// Export 导出为 Sing-box 配置
func (e *SingboxExporter) Export(nodes []*parser.Node) (string, error) {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "google", "address": "8.8.8.8"},
				{"tag": "local", "address": "223.5.5.5", "detour": "direct"},
			},
		},
		"inbounds": []map[string]interface{}{
			{
				"type":   "mixed",
				"tag":    "mixed-in",
				"listen": "127.0.0.1",
				"listen_port": 7890,
			},
		},
		"outbounds": e.nodesToOutbounds(nodes),
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"geoip":    "cn",
					"outbound": "direct",
				},
			},
			"final": "proxy",
		},
	}
	
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sing-box config: %w", err)
	}
	
	return string(jsonBytes), nil
}

// nodesToOutbounds 将节点转换为 outbounds
func (e *SingboxExporter) nodesToOutbounds(nodes []*parser.Node) []map[string]interface{} {
	outbounds := []map[string]interface{}{}
	
	// 添加代理节点
	for _, node := range nodes {
		outbound, err := e.nodeToOutbound(node)
		if err != nil {
			continue
		}
		outbounds = append(outbounds, outbound)
	}
	
	// 添加 selector
	proxyTags := []string{}
	for _, node := range nodes {
		proxyTags = append(proxyTags, node.Name)
	}
	
	outbounds = append([]map[string]interface{}{
		{
			"type":     "selector",
			"tag":      "proxy",
			"outbounds": proxyTags,
		},
		{
			"type": "direct",
			"tag":  "direct",
		},
		{
			"type": "block",
			"tag":  "block",
		},
	}, outbounds...)
	
	return outbounds
}

// nodeToOutbound 将节点转换为 outbound
func (e *SingboxExporter) nodeToOutbound(node *parser.Node) (map[string]interface{}, error) {
	switch node.Type {
	case "vmess":
		return e.vmessToOutbound(node), nil
	case "vless":
		return e.vlessToOutbound(node), nil
	case "trojan":
		return e.trojanToOutbound(node), nil
	case "ss":
		return e.ssToOutbound(node), nil
	default:
		return nil, fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// vmessToOutbound VMess 转 Sing-box outbound
func (e *SingboxExporter) vmessToOutbound(node *parser.Node) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":   "vmess",
		"tag":    node.Name,
		"server": node.Server,
		"server_port": node.Port,
		"uuid":   node.UUID,
		"alter_id": node.AlterId,
	}
	
	if node.TLS {
		outbound["tls"] = map[string]interface{}{
			"enabled": true,
		}
		if node.SNI != "" {
			outbound["tls"].(map[string]interface{})["server_name"] = node.SNI
		}
	}
	
	if node.Network == "ws" {
		outbound["transport"] = map[string]interface{}{
			"type": "ws",
			"path": node.WSPath,
		}
		if len(node.WSHeaders) > 0 {
			outbound["transport"].(map[string]interface{})["headers"] = node.WSHeaders
		}
	}
	
	return outbound
}

// vlessToOutbound VLESS 转 Sing-box outbound
func (e *SingboxExporter) vlessToOutbound(node *parser.Node) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":   "vless",
		"tag":    node.Name,
		"server": node.Server,
		"server_port": node.Port,
		"uuid":   node.UUID,
	}
	
	if node.TLS {
		outbound["tls"] = map[string]interface{}{
			"enabled": true,
		}
		if node.SNI != "" {
			outbound["tls"].(map[string]interface{})["server_name"] = node.SNI
		}
	}
	
	if node.Flow != "" {
		outbound["flow"] = node.Flow
	}
	
	return outbound
}

// trojanToOutbound Trojan 转 Sing-box outbound
func (e *SingboxExporter) trojanToOutbound(node *parser.Node) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":     "trojan",
		"tag":      node.Name,
		"server":   node.Server,
		"server_port": node.Port,
		"password": node.Password,
	}
	
	if node.TLS {
		outbound["tls"] = map[string]interface{}{
			"enabled": true,
		}
		if node.SNI != "" {
			outbound["tls"].(map[string]interface{})["server_name"] = node.SNI
		}
		if node.SkipCertVerify {
			outbound["tls"].(map[string]interface{})["insecure"] = true
		}
	}
	
	return outbound
}

// ssToOutbound Shadowsocks 转 Sing-box outbound
func (e *SingboxExporter) ssToOutbound(node *parser.Node) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":     "shadowsocks",
		"tag":      node.Name,
		"server":   node.Server,
		"server_port": node.Port,
		"method":   node.Cipher,
		"password": node.Password,
	}
	
	return outbound
}

// ExportWithConfig 使用自定义配置导出
func (e *SingboxExporter) ExportWithConfig(config map[string]interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sing-box config: %w", err)
	}
	
	return string(jsonBytes), nil
}
