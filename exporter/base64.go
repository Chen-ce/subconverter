package exporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	
	"github.com/Chen-ce/subconverter/parser"
)

// Base64Exporter Base64 格式导出器
type Base64Exporter struct{}

// NewBase64Exporter 创建 Base64 导出器
func NewBase64Exporter() *Base64Exporter {
	return &Base64Exporter{}
}

// Export 导出为 Base64 订阅格式
func (e *Base64Exporter) Export(nodes []*parser.Node) (string, error) {
	if len(nodes) == 0 {
		return "", fmt.Errorf("no nodes to export")
	}
	
	var lines []string
	for _, node := range nodes {
		uri, err := nodeToURI(node)
		if err != nil {
			fmt.Printf("Warning: failed to convert node to URI: %v\n", err)
			continue
		}
		lines = append(lines, uri)
	}
	
	if len(lines) == 0 {
		return "", fmt.Errorf("no valid nodes to export")
	}
	
	// 合并所有 URI
	content := strings.Join(lines, "\n")
	
	// Base64 编码
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	
	return encoded, nil
}

// nodeToURI 将节点转换为 URI 格式
func nodeToURI(node *parser.Node) (string, error) {
	switch node.Type {
	case parser.ProxyTypeVMess:
		return vmessToURI(node)
	case parser.ProxyTypeVLESS:
		return vlessToURI(node)
	case parser.ProxyTypeTrojan:
		return trojanToURI(node)
	case parser.ProxyTypeShadowsocks:
		return shadowsocksToURI(node)
	case parser.ProxyTypeShadowsocksR:
		return shadowsocksRToURI(node)
	default:
		return "", fmt.Errorf("unsupported proxy type: %s", node.Type)
	}
}

// vmessToURI 转换 VMess 节点为 URI
func vmessToURI(node *parser.Node) (string, error) {
	vmess := map[string]interface{}{
		"v":    "2",
		"ps":   node.Name,
		"add":  node.Server,
		"port": fmt.Sprintf("%d", node.Port),
		"id":   node.UUID,
		"aid":  fmt.Sprintf("%d", node.AlterId),
		"net":  node.Network,
		"type": "none",
		"tls":  "",
	}
	
	if node.TLS {
		vmess["tls"] = "tls"
		if node.SNI != "" {
			vmess["sni"] = node.SNI
		}
	}
	
	// WebSocket 配置
	if node.Network == "ws" {
		vmess["path"] = node.WSPath
		if host, ok := node.WSHeaders["Host"]; ok {
			vmess["host"] = host
		}
	}
	
	// gRPC 配置
	if node.Network == "grpc" {
		vmess["path"] = node.GRPCServiceName
	}
	
	jsonBytes, err := json.Marshal(vmess)
	if err != nil {
		return "", err
	}
	
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded, nil
}

// vlessToURI 转换 VLESS 节点为 URI
func vlessToURI(node *parser.Node) (string, error) {
	// vless://uuid@server:port?params#name
	params := url.Values{}
	
	if node.Network != "" {
		params.Set("type", node.Network)
	}
	
	if node.TLS {
		params.Set("security", "tls")
		if node.SNI != "" {
			params.Set("sni", node.SNI)
		}
	}
	
	if node.Flow != "" {
		params.Set("flow", node.Flow)
	}
	
	// WebSocket 配置
	if node.Network == "ws" {
		if node.WSPath != "" {
			params.Set("path", node.WSPath)
		}
		if host, ok := node.WSHeaders["Host"]; ok {
			params.Set("host", host)
		}
	}
	
	// gRPC 配置
	if node.Network == "grpc" {
		if node.GRPCServiceName != "" {
			params.Set("serviceName", node.GRPCServiceName)
		}
	}
	
	uri := fmt.Sprintf("vless://%s@%s:%d", node.UUID, node.Server, node.Port)
	
	if len(params) > 0 {
		uri += "?" + params.Encode()
	}
	
	if node.Name != "" {
		uri += "#" + url.QueryEscape(node.Name)
	}
	
	return uri, nil
}

// trojanToURI 转换 Trojan 节点为 URI
func trojanToURI(node *parser.Node) (string, error) {
	// trojan://password@server:port?params#name
	params := url.Values{}
	
	if node.Network != "" && node.Network != "tcp" {
		params.Set("type", node.Network)
	}
	
	if node.SNI != "" {
		params.Set("sni", node.SNI)
	}
	
	if node.SkipCertVerify {
		params.Set("allowInsecure", "1")
	}
	
	// WebSocket 配置
	if node.Network == "ws" {
		if node.WSPath != "" {
			params.Set("path", node.WSPath)
		}
		if host, ok := node.WSHeaders["Host"]; ok {
			params.Set("host", host)
		}
	}
	
	// gRPC 配置
	if node.Network == "grpc" {
		if node.GRPCServiceName != "" {
			params.Set("serviceName", node.GRPCServiceName)
		}
	}
	
	uri := fmt.Sprintf("trojan://%s@%s:%d", url.QueryEscape(node.Password), node.Server, node.Port)
	
	if len(params) > 0 {
		uri += "?" + params.Encode()
	}
	
	if node.Name != "" {
		uri += "#" + url.QueryEscape(node.Name)
	}
	
	return uri, nil
}

// shadowsocksToURI 转换 Shadowsocks 节点为 URI
func shadowsocksToURI(node *parser.Node) (string, error) {
	// ss://base64(method:password)@server:port#name
	credentials := fmt.Sprintf("%s:%s", node.Method, node.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	
	uri := fmt.Sprintf("ss://%s@%s:%d", encoded, node.Server, node.Port)
	
	if node.Name != "" {
		uri += "#" + url.QueryEscape(node.Name)
	}
	
	return uri, nil
}

// shadowsocksRToURI 转换 ShadowsocksR 节点为 URI
func shadowsocksRToURI(node *parser.Node) (string, error) {
	// ssr://base64(server:port:protocol:method:obfs:password_base64/?params)
	passwordEncoded := base64.StdEncoding.EncodeToString([]byte(node.Password))
	
	content := fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		node.Server,
		node.Port,
		node.Protocol,
		node.Method,
		node.Obfs,
		passwordEncoded,
	)
	
	params := url.Values{}
	if node.Name != "" {
		remarksEncoded := base64.StdEncoding.EncodeToString([]byte(node.Name))
		params.Set("remarks", remarksEncoded)
	}
	if node.ProtocolParam != "" {
		params.Set("protoparam", node.ProtocolParam)
	}
	if node.ObfsParam != "" {
		params.Set("obfsparam", node.ObfsParam)
	}
	
	if len(params) > 0 {
		content += "/?" + params.Encode()
	}
	
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	return "ssr://" + encoded, nil
}
