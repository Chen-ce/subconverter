package parser

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Base64Parser Base64 订阅解析器
type Base64Parser struct{}

// NewBase64Parser 创建 Base64 解析器
func NewBase64Parser() *Base64Parser {
	return &Base64Parser{}
}

// Parse 解析 Base64 编码的订阅内容
func (p *Base64Parser) Parse(content string) ([]*Node, error) {
	// 去除空白字符
	content = strings.TrimSpace(content)
	
	// Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		// 尝试 RawStdEncoding
		decoded, err = base64.RawStdEncoding.DecodeString(content)
		if err != nil {
			// 尝试 URLEncoding
			decoded, err = base64.URLEncoding.DecodeString(content)
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 content: %w", err)
			}
		}
	}
	
	// 按行分割
	lines := strings.Split(string(decoded), "\n")
	
	var nodes []*Node
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 解析每一行的节点 URI
		node, err := ParseNodeURI(line)
		if err != nil {
			// 跳过无法解析的行，继续处理其他行
			fmt.Printf("Warning: failed to parse node URI: %s, error: %v\n", line, err)
			continue
		}
		
		nodes = append(nodes, node)
	}
	
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid nodes found in subscription")
	}
	
	return nodes, nil
}
