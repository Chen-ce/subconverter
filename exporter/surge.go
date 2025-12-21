package exporter

import (
	"fmt"
	"strings"
	
	"github.com/Chen-ce/subconverter/parser"
)

// SurgeExporter Surge 配置导出器
type SurgeExporter struct {
	version int // 2, 3, or 4
}

// NewSurgeExporter 创建 Surge 导出器
func NewSurgeExporter(version int) *SurgeExporter {
	if version < 2 || version > 4 {
		version = 4 // 默认使用 Surge 4
	}
	return &SurgeExporter{version: version}
}

// Export 导出为 Surge 配置
func (e *SurgeExporter) Export(nodes []*parser.Node) (string, error) {
	var sb strings.Builder
	
	// General 部分
	sb.WriteString("[General]\n")
	sb.WriteString("loglevel = notify\n")
	sb.WriteString("dns-server = 223.5.5.5, 119.29.29.29, system\n")
	sb.WriteString("skip-proxy = 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, localhost, *.local\n")
	sb.WriteString("ipv6 = false\n")
	sb.WriteString("\n")
	
	// Proxy 部分
	sb.WriteString("[Proxy]\n")
	for _, node := range nodes {
		proxyLine, err := e.nodeToSurgeProxy(node)
		if err != nil {
			continue
		}
		sb.WriteString(proxyLine)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	
	// Proxy Group 部分
	sb.WriteString("[Proxy Group]\n")
	sb.WriteString("Proxy = select")
	for _, node := range nodes {
		sb.WriteString(", ")
		sb.WriteString(node.Name)
	}
	sb.WriteString("\n")
	sb.WriteString("Auto = url-test")
	for _, node := range nodes {
		sb.WriteString(", ")
		sb.WriteString(node.Name)
	}
	sb.WriteString(", url = http://www.gstatic.com/generate_204, interval = 300\n")
	sb.WriteString("\n")
	
	// Rule 部分
	sb.WriteString("[Rule]\n")
	sb.WriteString("# LAN\n")
	sb.WriteString("DOMAIN-SUFFIX,local,DIRECT\n")
	sb.WriteString("IP-CIDR,192.168.0.0/16,DIRECT\n")
	sb.WriteString("IP-CIDR,10.0.0.0/8,DIRECT\n")
	sb.WriteString("IP-CIDR,172.16.0.0/12,DIRECT\n")
	sb.WriteString("IP-CIDR,127.0.0.0/8,DIRECT\n")
	sb.WriteString("\n")
	sb.WriteString("# China\n")
	sb.WriteString("GEOIP,CN,DIRECT\n")
	sb.WriteString("\n")
	sb.WriteString("# Final\n")
	sb.WriteString("FINAL,Proxy\n")
	
	return sb.String(), nil
}

// nodeToSurgeProxy 将节点转换为 Surge 代理行
func (e *SurgeExporter) nodeToSurgeProxy(node *parser.Node) (string, error) {
	switch node.Type {
	case "vmess":
		return e.vmessToSurge(node)
	case "vless":
		return e.vlessToSurge(node)
	case "trojan":
		return e.trojanToSurge(node)
	case "ss":
		return e.ssToSurge(node)
	case "ssr":
		return e.ssrToSurge(node)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// vmessToSurge VMess 转 Surge
func (e *SurgeExporter) vmessToSurge(node *parser.Node) (string, error) {
	// Surge 格式: name = vmess, server, port, username=uuid, ...
	parts := []string{
		node.Name + " = vmess",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		fmt.Sprintf("username=%s", node.UUID),
	}
	
	if node.TLS {
		parts = append(parts, "tls=true")
		if node.SNI != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", node.SNI))
		}
	}
	
	if node.Network != "" && node.Network != "tcp" {
		parts = append(parts, fmt.Sprintf("ws=true"))
		if node.WSPath != "" {
			parts = append(parts, fmt.Sprintf("ws-path=%s", node.WSPath))
		}
		if len(node.WSHeaders) > 0 {
			if host, ok := node.WSHeaders["Host"]; ok {
				parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", host))
			}
		}
	}
	
	return strings.Join(parts, ", "), nil
}

// vlessToSurge VLESS 转 Surge (Surge 4+ 支持)
func (e *SurgeExporter) vlessToSurge(node *parser.Node) (string, error) {
	if e.version < 4 {
		return "", fmt.Errorf("vless requires Surge 4+")
	}
	
	parts := []string{
		node.Name + " = vless",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		fmt.Sprintf("username=%s", node.UUID),
	}
	
	if node.TLS {
		parts = append(parts, "tls=true")
		if node.SNI != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", node.SNI))
		}
	}
	
	return strings.Join(parts, ", "), nil
}

// trojanToSurge Trojan 转 Surge
func (e *SurgeExporter) trojanToSurge(node *parser.Node) (string, error) {
	parts := []string{
		node.Name + " = trojan",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		fmt.Sprintf("password=%s", node.Password),
	}
	
	if node.SNI != "" {
		parts = append(parts, fmt.Sprintf("sni=%s", node.SNI))
	}
	
	if node.SkipCertVerify {
		parts = append(parts, "skip-cert-verify=true")
	}
	
	return strings.Join(parts, ", "), nil
}

// ssToSurge Shadowsocks 转 Surge
func (e *SurgeExporter) ssToSurge(node *parser.Node) (string, error) {
	parts := []string{
		node.Name + " = ss",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		fmt.Sprintf("encrypt-method=%s", node.Cipher),
		fmt.Sprintf("password=%s", node.Password),
	}
	
	if node.Plugin != "" {
		parts = append(parts, fmt.Sprintf("obfs=%s", node.Plugin))
		if len(node.PluginOpts) > 0 {
			if host, ok := node.PluginOpts["host"].(string); ok {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
			}
		}
	}
	
	return strings.Join(parts, ", "), nil
}

// ssrToSurge ShadowsocksR 转 Surge (部分支持)
func (e *SurgeExporter) ssrToSurge(node *parser.Node) (string, error) {
	// Surge 不完全支持 SSR，尝试转换为 SS
	return e.ssToSurge(node)
}

// ExportWithConfig 使用自定义配置导出
func (e *SurgeExporter) ExportWithConfig(config map[string]interface{}) (string, error) {
	// TODO: 实现自定义配置导出
	return "", fmt.Errorf("custom config not implemented for Surge yet")
}
