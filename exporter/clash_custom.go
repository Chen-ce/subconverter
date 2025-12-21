package exporter

import (
	"fmt"
	
	"gopkg.in/yaml.v3"
)

// ExportWithConfig 使用自定义配置导出
func (e *ClashExporter) ExportWithConfig(config map[string]interface{}) (string, error) {
	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal clash config: %w", err)
	}
	
	return string(yamlBytes), nil
}
