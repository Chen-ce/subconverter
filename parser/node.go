package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParseNodeURI 解析单个节点 URI
func ParseNodeURI(uri string) (*Node, error) {
	uri = strings.TrimSpace(uri)
	
	if strings.HasPrefix(uri, "vmess://") {
		return parseVMessURI(uri)
	} else if strings.HasPrefix(uri, "vless://") {
		return parseVLESSURI(uri)
	} else if strings.HasPrefix(uri, "trojan://") {
		return parseTrojanURI(uri)
	} else if strings.HasPrefix(uri, "ss://") {
		return parseShadowsocksURI(uri)
	} else if strings.HasPrefix(uri, "ssr://") {
		return parseShadowsocksRURI(uri)
	}
	
	return nil, fmt.Errorf("unsupported URI scheme: %s", uri)
}

// parseVMessURI 解析 VMess URI (vmess://<base64>)
func parseVMessURI(uri string) (*Node, error) {
	// 移除 vmess:// 前缀
	encoded := strings.TrimPrefix(uri, "vmess://")
	
	// Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// 尝试 RawStdEncoding
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess URI: %w", err)
		}
	}
	
	// 解析 JSON
	var vmess struct {
		V    string `json:"v"`
		PS   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		Aid  string `json:"aid"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		SNI  string `json:"sni"`
	}
	
	if err := json.Unmarshal(decoded, &vmess); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}
	
	port, _ := strconv.Atoi(vmess.Port)
	alterId, _ := strconv.Atoi(vmess.Aid)
	
	node := &Node{
		Type:    ProxyTypeVMess,
		Name:    vmess.PS,
		Server:  vmess.Add,
		Port:    port,
		UUID:    vmess.ID,
		AlterId: alterId,
		Network: vmess.Net,
		TLS:     vmess.TLS == "tls",
		SNI:     vmess.SNI,
	}
	
	// WebSocket 配置
	if vmess.Net == "ws" {
		node.WSPath = vmess.Path
		if vmess.Host != "" {
			node.WSHeaders = map[string]string{"Host": vmess.Host}
		}
	}
	
	// gRPC 配置
	if vmess.Net == "grpc" {
		node.GRPCServiceName = vmess.Path
	}
	
	return node, nil
}

// parseVLESSURI 解析 VLESS URI
func parseVLESSURI(uri string) (*Node, error) {
	// vless://uuid@server:port?params#name
	uri = strings.TrimPrefix(uri, "vless://")
	
	// 分离名称
	parts := strings.SplitN(uri, "#", 2)
	name := ""
	if len(parts) == 2 {
		name, _ = url.QueryUnescape(parts[1])
		uri = parts[0]
	}
	
	// 分离参数
	parts = strings.SplitN(uri, "?", 2)
	params := url.Values{}
	if len(parts) == 2 {
		params, _ = url.ParseQuery(parts[1])
		uri = parts[0]
	}
	
	// 解析 uuid@server:port
	parts = strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid vless URI format")
	}
	
	uuid := parts[0]
	serverPort := strings.SplitN(parts[1], ":", 2)
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}
	
	port, _ := strconv.Atoi(serverPort[1])
	
	node := &Node{
		Type:   ProxyTypeVLESS,
		Name:   name,
		Server: serverPort[0],
		Port:   port,
		UUID:   uuid,
		Network: params.Get("type"),
		TLS:    params.Get("security") == "tls",
		SNI:    params.Get("sni"),
		Flow:   params.Get("flow"),
	}
	
	// WebSocket 配置
	if node.Network == "ws" {
		node.WSPath = params.Get("path")
		if host := params.Get("host"); host != "" {
			node.WSHeaders = map[string]string{"Host": host}
		}
	}
	
	// gRPC 配置
	if node.Network == "grpc" {
		node.GRPCServiceName = params.Get("serviceName")
	}
	
	return node, nil
}

// parseTrojanURI 解析 Trojan URI
func parseTrojanURI(uri string) (*Node, error) {
	// trojan://password@server:port?params#name
	uri = strings.TrimPrefix(uri, "trojan://")
	
	// 分离名称
	parts := strings.SplitN(uri, "#", 2)
	name := ""
	if len(parts) == 2 {
		name, _ = url.QueryUnescape(parts[1])
		uri = parts[0]
	}
	
	// 分离参数
	parts = strings.SplitN(uri, "?", 2)
	params := url.Values{}
	if len(parts) == 2 {
		params, _ = url.ParseQuery(parts[1])
		uri = parts[0]
	}
	
	// 解析 password@server:port
	parts = strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid trojan URI format")
	}
	
	password, _ := url.QueryUnescape(parts[0])
	serverPort := strings.SplitN(parts[1], ":", 2)
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}
	
	port, _ := strconv.Atoi(serverPort[1])
	
	node := &Node{
		Type:     ProxyTypeTrojan,
		Name:     name,
		Server:   serverPort[0],
		Port:     port,
		Password: password,
		Network:  params.Get("type"),
		TLS:      true, // Trojan 总是使用 TLS
		SNI:      params.Get("sni"),
		SkipCertVerify: params.Get("allowInsecure") == "1",
	}
	
	// WebSocket 配置
	if node.Network == "ws" {
		node.WSPath = params.Get("path")
		if host := params.Get("host"); host != "" {
			node.WSHeaders = map[string]string{"Host": host}
		}
	}
	
	// gRPC 配置
	if node.Network == "grpc" {
		node.GRPCServiceName = params.Get("serviceName")
	}
	
	return node, nil
}

// parseShadowsocksURI 解析 Shadowsocks URI
func parseShadowsocksURI(uri string) (*Node, error) {
	// ss://base64(method:password)@server:port#name
	uri = strings.TrimPrefix(uri, "ss://")
	
	// 分离名称
	parts := strings.SplitN(uri, "#", 2)
	name := ""
	if len(parts) == 2 {
		name, _ = url.QueryUnescape(parts[1])
		uri = parts[0]
	}
	
	// 分离 server:port
	parts = strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss URI format")
	}
	
	// 解码 method:password
	decoded, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode ss credentials: %w", err)
		}
	}
	
	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, fmt.Errorf("invalid ss credentials format")
	}
	
	serverPort := strings.SplitN(parts[1], ":", 2)
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}
	
	port, _ := strconv.Atoi(serverPort[1])
	
	return &Node{
		Type:     ProxyTypeShadowsocks,
		Name:     name,
		Server:   serverPort[0],
		Port:     port,
		Method:   credentials[0],
		Password: credentials[1],
	}, nil
}

// parseShadowsocksRURI 解析 ShadowsocksR URI
func parseShadowsocksRURI(uri string) (*Node, error) {
	// ssr://base64(server:port:protocol:method:obfs:password_base64/?params)
	encoded := strings.TrimPrefix(uri, "ssr://")
	
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ssr URI: %w", err)
		}
	}
	
	content := string(decoded)
	
	// 分离参数
	parts := strings.SplitN(content, "/?", 2)
	params := url.Values{}
	if len(parts) == 2 {
		params, _ = url.ParseQuery(parts[1])
		content = parts[0]
	}
	
	// 解析主要部分
	fields := strings.Split(content, ":")
	if len(fields) < 6 {
		return nil, fmt.Errorf("invalid ssr URI format")
	}
	
	port, _ := strconv.Atoi(fields[1])
	
	// 解码密码
	passwordDecoded, _ := base64.StdEncoding.DecodeString(fields[5])
	if len(passwordDecoded) == 0 {
		passwordDecoded, _ = base64.RawStdEncoding.DecodeString(fields[5])
	}
	
	// 解码备注
	remarksDecoded, _ := base64.StdEncoding.DecodeString(params.Get("remarks"))
	if len(remarksDecoded) == 0 {
		remarksDecoded, _ = base64.RawStdEncoding.DecodeString(params.Get("remarks"))
	}
	
	return &Node{
		Type:          ProxyTypeShadowsocksR,
		Name:          string(remarksDecoded),
		Server:        fields[0],
		Port:          port,
		Protocol:      fields[2],
		Method:        fields[3],
		Obfs:          fields[4],
		Password:      string(passwordDecoded),
		ProtocolParam: params.Get("protoparam"),
		ObfsParam:     params.Get("obfsparam"),
	}, nil
}
