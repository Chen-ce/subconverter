package parser

// ProxyType 代理协议类型
type ProxyType string

const (
	ProxyTypeVMess          ProxyType = "vmess"
	ProxyTypeVLESS          ProxyType = "vless"
	ProxyTypeTrojan         ProxyType = "trojan"
	ProxyTypeShadowsocks    ProxyType = "ss"
	ProxyTypeShadowsocksR   ProxyType = "ssr"
	ProxyTypeHysteria       ProxyType = "hysteria"
	ProxyTypeHysteria2      ProxyType = "hysteria2"
)

// Node 代理节点通用结构
type Node struct {
	Type     ProxyType         `json:"type"`
	Name     string            `json:"name"`
	Server   string            `json:"server"`
	Port     int               `json:"port"`
	
	// 通用字段
	UUID     string            `json:"uuid,omitempty"`      // VMess/VLESS
	Password string            `json:"password,omitempty"`  // Trojan/SS/SSR
	
	// VMess/VLESS 特有
	AlterId  int               `json:"alterId,omitempty"`
	Cipher   string            `json:"cipher,omitempty"`
	
	// 传输层配置
	Network  string            `json:"network,omitempty"`   // tcp, ws, grpc, http, quic
	TLS      bool              `json:"tls,omitempty"`
	SNI      string            `json:"sni,omitempty"`
	
	// WebSocket
	WSPath   string            `json:"ws-path,omitempty"`
	WSHeaders map[string]string `json:"ws-headers,omitempty"`
	
	// gRPC
	GRPCServiceName string     `json:"grpc-service-name,omitempty"`
	
	// HTTP/2
	HTTPPath []string          `json:"http-path,omitempty"`
	HTTPHost []string          `json:"http-host,omitempty"`
	
	// Shadowsocks
	Method   string            `json:"method,omitempty"`
	Plugin   string            `json:"plugin,omitempty"`
	PluginOpts map[string]interface{} `json:"plugin-opts,omitempty"`
	
	// ShadowsocksR
	Protocol      string       `json:"protocol,omitempty"`
	ProtocolParam string       `json:"protocol-param,omitempty"`
	Obfs          string       `json:"obfs,omitempty"`
	ObfsParam     string       `json:"obfs-param,omitempty"`
	
	// VLESS
	Flow     string            `json:"flow,omitempty"`
	
	// 其他
	UDP      bool              `json:"udp,omitempty"`
	SkipCertVerify bool       `json:"skip-cert-verify,omitempty"`
}

// Parser 订阅解析器接口
type Parser interface {
	Parse(content string) ([]*Node, error)
}
