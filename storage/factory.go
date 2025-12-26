package storage

import (
	"fmt"
	"strings"
)

// NewStorage 根据配置创建存储实例
func NewStorage(cfgType, path, repoID, token string) (Storage, error) {
	switch strings.ToLower(cfgType) {
	case "hf", "huggingface":
		return NewHFDatasetStorage(repoID, token)
	case "json", "file", "":
		// 默认使用 JSON
		return NewJSONStorage(path)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s (supported: json, hf)", cfgType)
	}
}
