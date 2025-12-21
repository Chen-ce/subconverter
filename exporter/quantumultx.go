package exporter

import (
	"fmt"
	"strings"
	
	"github.com/Chen-ce/subconverter/parser"
)

// QuantumultXExporter Quantumult X 配置导出器
type QuantumultXExporter struct{}

// NewQuantumultXExporter 创建 Quantumult X 导出器
func NewQuantumultXExporter() *QuantumultXExporter {
	return &QuantumultXExporter{}
}

// Export 导出为 Quantumult X 配置
func (e *QuantumultXExporter) Export(nodes []*parser.Node) (string, error) {
	var sb strings.Builder
	
	// General 部分
	sb.WriteString("[general]\n")
	sb.WriteString("server_check_url=http://www.gstatic.com/generate_204\n")
	sb.WriteString("dns_exclusion_list=*.cmpassport.com, *.jegotrip.com.cn, *.icitymobile.mobi, id6.me\n")
	sb.WriteString("\n")
	
	// DNS 部分
	sb.WriteString("[dns]\n")
	sb.WriteString("server=223.5.5.5\n")
	sb.WriteString("server=119.29.29.29\n")
	sb.WriteString("server=8.8.8.8\n")
	sb.WriteString("\n")
	
	// Policy 部分
	sb.WriteString("[policy]\n")
	sb.WriteString("static=Proxy")
	for _, node := range nodes {
		sb.WriteString(", ")
		sb.WriteString(node.Name)
	}
	sb.WriteString("\n\n")
	
	// Server Local 部分
	sb.WriteString("[server_local]\n")
	for _, node := range nodes {
		proxyLine, err := e.nodeToQuantumultX(node)
		if err != nil {
			continue
		}
		sb.WriteString(proxyLine)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	
	// Filter Local 部分
	sb.WriteString("[filter_local]\n")
	sb.WriteString("host-suffix, local, direct\n")
	sb.WriteString("ip-cidr, 192.168.0.0/16, direct\n")
	sb.WriteString("ip-cidr, 10.0.0.0/8, direct\n")
	sb.WriteString("ip-cidr, 172.16.0.0/12, direct\n")
	sb.WriteString("ip-cidr, 127.0.0.0/8, direct\n")
	sb.WriteString("geoip, cn, direct\n")
	sb.WriteString("final, Proxy\n")
	
	return sb.String(), nil
}

// nodeToQuantumultX 将节点转换为 Quantumult X 格式
func (e *QuantumultXExporter) nodeToQuantumultX(node *parser.Node) (string, error) {
	switch node.Type {
	case "vmess":
		return e.vmessToQuantumultX(node), nil
	case "trojan":
		return e.trojanToQuantumultX(node), nil
	case "ss":
		return e.ssToQuantumultX(node), nil
	case "ssr":
		return e.ssrToQuantumultX(node), nil
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// vmessToQuantumultX VMess 转 Quantumult X
func (e *QuantumultXExporter) vmessToQuantumultX(node *parser.Node) string {
	// vmess=server:port, method=chacha20-poly1305, password=uuid, obfs=ws, obfs-uri=/path, obfs-host=host, tag=name
	parts := []string{
		fmt.Sprintf("vmess=%s:%d", node.Server, node.Port),
		fmt.Sprintf("method=%s", node.Cipher),
		fmt.Sprintf("password=%s", node.UUID),
	}
	
	if node.Network == "ws" {
		parts = append(parts, "obfs=ws")
		if node.WSPath != "" {
			parts = append(parts, fmt.Sprintf("obfs-uri=%s", node.WSPath))
		}
		if len(node.WSHeaders) > 0 {
			if host, ok := node.WSHeaders["Host"]; ok {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
			}
		}
	}
	
	if node.TLS {
		parts = append(parts, "tls-verification=true")
	}
	
	parts = append(parts, fmt.Sprintf("tag=%s", node.Name))
	
	return strings.Join(parts, ", ")
}

// trojanToQuantumultX Trojan 转 Quantumult X
func (e *QuantumultXExporter) trojanToQuantumultX(node *parser.Node) string {
	// trojan=server:port, password=pwd, over-tls=true, tls-verification=true, tag=name
	parts := []string{
		fmt.Sprintf("trojan=%s:%d", node.Server, node.Port),
		fmt.Sprintf("password=%s", node.Password),
		"over-tls=true",
	}
	
	if !node.SkipCertVerify {
		parts = append(parts, "tls-verification=true")
	}
	
	if node.SNI != "" {
		parts = append(parts, fmt.Sprintf("tls-host=%s", node.SNI))
	}
	
	parts = append(parts, fmt.Sprintf("tag=%s", node.Name))
	
	return strings.Join(parts, ", ")
}

// ssToQuantumultX Shadowsocks 转 Quantumult X
func (e *QuantumultXExporter) ssToQuantumultX(node *parser.Node) string {
	// shadowsocks=server:port, method=aes-256-gcm, password=pwd, tag=name
	parts := []string{
		fmt.Sprintf("shadowsocks=%s:%d", node.Server, node.Port),
		fmt.Sprintf("method=%s", node.Cipher),
		fmt.Sprintf("password=%s", node.Password),
	}
	
	if node.Plugin == "obfs" {
		if len(node.PluginOpts) > 0 {
			if mode, ok := node.PluginOpts["mode"].(string); ok {
				parts = append(parts, fmt.Sprintf("obfs=%s", mode))
			}
			if host, ok := node.PluginOpts["host"].(string); ok {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
			}
		}
	}
	
	parts = append(parts, fmt.Sprintf("tag=%s", node.Name))
	
	return strings.Join(parts, ", ")
}

// ssrToQuantumultX ShadowsocksR 转 Quantumult X
func (e *QuantumultXExporter) ssrToQuantumultX(node *parser.Node) string {
	// Quantumult X 不完全支持 SSR，尝试转换为 SS
	return e.ssToQuantumultX(node)
}
