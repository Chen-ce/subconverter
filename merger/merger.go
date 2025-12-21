package merger

import (
	"fmt"
	"github.com/Chen-ce/subconverter/parser"
)

// Merger 节点合并器
type Merger struct {
	nodes []*parser.Node
}

// NewMerger 创建节点合并器
func NewMerger() *Merger {
	return &Merger{
		nodes: make([]*parser.Node, 0),
	}
}

// Add 添加节点列表
func (m *Merger) Add(nodes []*parser.Node) {
	m.nodes = append(m.nodes, nodes...)
}

// Merge 合并并去重节点
func (m *Merger) Merge() []*parser.Node {
	if len(m.nodes) == 0 {
		return []*parser.Node{}
	}
	
	// 使用 map 进行去重，key 为 server:port
	seen := make(map[string]bool)
	var result []*parser.Node
	
	for _, node := range m.nodes {
		key := fmt.Sprintf("%s:%d", node.Server, node.Port)
		
		if !seen[key] {
			seen[key] = true
			result = append(result, node)
		}
	}
	
	return result
}

// Filter 过滤节点（可选功能）
func (m *Merger) Filter(filterFunc func(*parser.Node) bool) {
	var filtered []*parser.Node
	for _, node := range m.nodes {
		if filterFunc(node) {
			filtered = append(filtered, node)
		}
	}
	m.nodes = filtered
}

// Rename 重命名节点（可选功能）
func (m *Merger) Rename(renameFunc func(*parser.Node) string) {
	for _, node := range m.nodes {
		node.Name = renameFunc(node)
	}
}
