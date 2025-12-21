package cmd

import (
	"fmt"
	"os"
	"strings"
	
	"github.com/Chen-ce/subconverter/exporter"
	"github.com/Chen-ce/subconverter/fetcher"
	"github.com/Chen-ce/subconverter/merger"
	"github.com/Chen-ce/subconverter/parser"
	"github.com/Chen-ce/subconverter/templates"
	"github.com/spf13/cobra"
)

var (
	subscriptions []string
	nodeURIs      []string
	outputFormat  string
	outputFile    string
	includeFilter string
	excludeFilter string
	configName    string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple subscriptions and nodes",
	Long: `Merge multiple proxy subscriptions and individual nodes into one,
then export to the specified format (base64 or clash).`,
	Example: `  # Merge subscriptions and export to Clash format
  subconverter merge --sub https://example.com/sub1 --sub https://example.com/sub2 --output clash --file output.yaml
  
  # Merge with individual nodes
  subconverter merge --sub https://example.com/sub1 --node "vmess://..." --output base64 --file output.txt
  
  # Filter nodes by keyword
  subconverter merge --sub https://example.com/sub1 --include "香港|HK" --exclude "x2|x3" --output clash`,
	RunE: runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	
	mergeCmd.Flags().StringArrayVarP(&subscriptions, "sub", "s", []string{}, "Subscription URLs (can be specified multiple times)")
	mergeCmd.Flags().StringArrayVarP(&nodeURIs, "node", "n", []string{}, "Individual node URIs (can be specified multiple times)")
	mergeCmd.Flags().StringVarP(&outputFormat, "output", "o", "base64", "Output format: base64 or clash")
	mergeCmd.Flags().StringVarP(&outputFile, "file", "f", "", "Output file path (optional, prints to stdout if not specified)")
	mergeCmd.Flags().StringVar(&includeFilter, "include", "", "Include nodes matching keywords (separated by |)")
	mergeCmd.Flags().StringVar(&excludeFilter, "exclude", "", "Exclude nodes matching keywords (separated by |)")
	mergeCmd.Flags().StringVarP(&configName, "config", "c", "default", "Rule config name for Clash output")
}

func runMerge(cmd *cobra.Command, args []string) error {
	if len(subscriptions) == 0 && len(nodeURIs) == 0 {
		return fmt.Errorf("at least one subscription or node must be specified")
	}
	
	// 创建合并器
	m := merger.NewMerger()
	
	// 获取并解析订阅
	if len(subscriptions) > 0 {
		f := fetcher.NewFetcher()
		
		for _, subURL := range subscriptions {
			fmt.Printf("Fetching subscription: %s\n", subURL)
			
			content, err := f.Fetch(subURL)
			if err != nil {
				fmt.Printf("Warning: failed to fetch subscription %s: %v\n", subURL, err)
				continue
			}
			
			// 尝试解析为不同格式
			nodes, err := parseSubscription(content)
			if err != nil {
				fmt.Printf("Warning: failed to parse subscription %s: %v\n", subURL, err)
				continue
			}
			
			fmt.Printf("Parsed %d nodes from %s\n", len(nodes), subURL)
			m.Add(nodes)
		}
	}
	
	// 解析单独的节点
	if len(nodeURIs) > 0 {
		for _, uri := range nodeURIs {
			node, err := parser.ParseNodeURI(uri)
			if err != nil {
				fmt.Printf("Warning: failed to parse node URI: %v\n", err)
				continue
			}
			
			fmt.Printf("Parsed node: %s\n", node.Name)
			m.Add([]*parser.Node{node})
		}
	}
	
	// 合并并去重
	nodes := m.Merge()
	fmt.Printf("\nTotal nodes after merging: %d\n", len(nodes))
	
	// 应用过滤
	if includeFilter != "" || excludeFilter != "" {
		nodes = filterNodes(nodes, includeFilter, excludeFilter)
		fmt.Printf("Nodes after filtering: %d\n", len(nodes))
	}
	
	if len(nodes) == 0 {
		return fmt.Errorf("no valid nodes to export")
	}
	
	// 导出
	var result string
	var err error
	
	switch strings.ToLower(outputFormat) {
	case "base64":
		exp := exporter.NewBase64Exporter()
		result, err = exp.Export(nodes)
	case "clash":
		// 使用模板
		tmplMgr := templates.NewManager("./templates")
		clashConfig, err := tmplMgr.ApplyToNodes(configName, nodes)
		if err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}
		
		exp := exporter.NewClashExporter()
		result, err = exp.ExportWithConfig(clashConfig)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: base64, clash)", outputFormat)
	}
	
	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}
	
	// 输出结果
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("\nExported to: %s\n", outputFile)
	} else {
		fmt.Println("\n" + result)
	}
	
	return nil
}

// parseSubscription 尝试解析订阅内容
func parseSubscription(content string) ([]*parser.Node, error) {
	// 先尝试 Base64
	base64Parser := parser.NewBase64Parser()
	nodes, err := base64Parser.Parse(content)
	if err == nil {
		return nodes, nil
	}
	
	// 再尝试 Clash
	clashParser := parser.NewClashParser()
	nodes, err = clashParser.Parse(content)
	if err == nil {
		return nodes, nil
	}
	
	return nil, fmt.Errorf("failed to parse subscription as base64 or clash format")
}

// filterNodes 过滤节点
func filterNodes(nodes []*parser.Node, include, exclude string) []*parser.Node {
	var result []*parser.Node
	
	includeKeywords := strings.Split(include, "|")
	excludeKeywords := strings.Split(exclude, "|")
	
	for _, node := range nodes {
		// 检查排除条件
		if exclude != "" {
			excluded := false
			for _, keyword := range excludeKeywords {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" && strings.Contains(node.Name, keyword) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		// 检查包含条件
		if include != "" {
			included := false
			for _, keyword := range includeKeywords {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" && strings.Contains(node.Name, keyword) {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		result = append(result, node)
	}
	
	return result
}
