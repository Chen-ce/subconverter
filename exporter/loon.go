package exporter

import (
	"fmt"
	"strings"
	
	"github.com/Chen-ce/subconverter/parser"
)

// LoonExporter Loon 配置导出器
type LoonExporter struct{}

// NewLoonExporter 创建 Loon 导出器
func NewLoonExporter() *LoonExporter {
	return &LoonExporter{}
}

// Export 导出为 Loon 配置
func (e *LoonExporter) Export(nodes []*parser.Node) (string, error) {
	var sb strings.Builder
	
	// General 部分
	sb.WriteString("[General]\n")
	sb.WriteString("skip-proxy = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,localhost,*.local\n")
	sb.WriteString("bypass-tun = 10.0.0.0/8,127.0.0.0/8,169.254.0.0/16,192.168.0.0/16,172.16.0.0/12\n")
	sb.WriteString("dns-server = system,223.5.5.5,119.29.29.29\n")
	sb.WriteString("allow-udp-proxy = true\n")
	sb.WriteString("proxy-test-url = http://www.gstatic.com/generate_204\n")
	sb.WriteString("\n")
	
	// Proxy 部分
	sb.WriteString("[Proxy]\n")
	for _, node := range nodes {
		proxyLine, err := e.nodeToLoon(node)
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
		sb.WriteString(",")
		sb.WriteString(node.Name)
	}
	sb.WriteString("\n")
	sb.WriteString("Auto = url-test")
	for _, node := range nodes {
		sb.WriteString(",")
		sb.WriteString(node.Name)
	}
	sb.WriteString(",url = http://www.gstatic.com/generate_204,interval = 600\n")
	sb.WriteString("\n")
	
	// Rule 部分
	sb.WriteString("[Rule]\n")
	sb.WriteString("DOMAIN-SUFFIX,local,DIRECT\n")
	sb.WriteString("IP-CIDR,192.168.0.0/16,DIRECT\n")
	sb.WriteString("IP-CIDR,10.0.0.0/8,DIRECT\n")
	sb.WriteString("IP-CIDR,172.16.0.0/12,DIRECT\n")
	sb.WriteString("IP-CIDR,127.0.0.0/8,DIRECT\n")
	sb.WriteString("GEOIP,CN,DIRECT\n")
	sb.WriteString("FINAL,Proxy\n")
	
	return sb.String(), nil
}

// nodeToLoon 将节点转换为 Loon 格式
func (e *LoonExporter) nodeToLoon(node *parser.Node) (string, error) {
	switch node.Type {
	case "vmess":
		return e.vmessToLoon(node), nil
	case "trojan":
		return e.trojanToLoon(node), nil
	case "ss":
		return e.ssToLoon(node), nil
	case "ssr":
		return e.ssrToLoon(node), nil
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

// vmessToLoon VMess 转 Loon
func (e *LoonExporter) vmessToLoon(node *parser.Node) string {
	// name = vmess,server,port,method,"uuid",transport:ws,path:/,host:host,over-tls:true
	parts := []string{
		node.Name + " = vmess",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		node.Cipher,
		fmt.Sprintf("\"%s\"", node.UUID),
	}
	
	if node.Network == "ws" {
		parts = append(parts, "transport:ws")
		if node.WSPath != "" {
			parts = append(parts, fmt.Sprintf("path:%s", node.WSPath))
		}
		if len(node.WSHeaders) > 0 {
			if host, ok := node.WSHeaders["Host"]; ok {
				parts = append(parts, fmt.Sprintf("host:%s", host))
			}
		}
	}
	
	if node.TLS {
		parts = append(parts, "over-tls:true")
		if node.SNI != "" {
			parts = append(parts, fmt.Sprintf("tls-name:%s", node.SNI))
		}
	}
	
	return strings.Join(parts, ",")
}

// trojanToLoon Trojan 转 Loon
func (e *LoonExporter) trojanToLoon(node *parser.Node) string {
	// name = trojan,server,port,password,skip-cert-verify=false
	parts := []string{
		node.Name + " = trojan",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		node.Password,
	}
	
	if node.SkipCertVerify {
		parts = append(parts, "skip-cert-verify=true")
	} else {
		parts = append(parts, "skip-cert-verify=false")
	}
	
	if node.SNI != "" {
		parts = append(parts, fmt.Sprintf("tls-name=%s", node.SNI))
	}
	
	return strings.Join(parts, ",")
}

// ssToLoon Shadowsocks 转 Loon
func (e *LoonExporter) ssToLoon(node *parser.Node) string {
	// name = ss,server,port,method,"password"
	parts := []string{
		node.Name + " = ss",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		node.Cipher,
		fmt.Sprintf("\"%s\"", node.Password),
	}
	
	return strings.Join(parts, ",")
}

// ssrToLoon ShadowsocksR 转 Loon
func (e *LoonExporter) ssrToLoon(node *parser.Node) string {
	// Loon 支持 SSR
	// name = ssr,server,port,method,"password",protocol,obfs
	parts := []string{
		node.Name + " = ssr",
		node.Server,
		fmt.Sprintf("%d", node.Port),
		node.Cipher,
		fmt.Sprintf("\"%s\"", node.Password),
	}
	
	if node.Protocol != "" {
		parts = append(parts, node.Protocol)
	}
	
	if node.Obfs != "" {
		parts = append(parts, node.Obfs)
	}
	
	return strings.Join(parts, ",")
}
