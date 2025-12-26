package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// JSONStorage JSON 文件存储实现
type JSONStorage struct {
	dataDir string
}

// NewJSONStorage 创建 JSON 存储
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	return &JSONStorage{
		dataDir: dataDir,
	}, nil
}

// GenerateID 生成随机 ID
func (s *JSONStorage) GenerateID() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Save 保存配置
func (s *JSONStorage) Save(cfg *SubscriptionConfig) error {
	cfg.UpdatedAt = time.Now()
	cfg.Version++

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return s.SaveFile(cfg.ID, "core.json", data)
}

// SaveFile 保存辅助文件
func (s *JSONStorage) SaveFile(id, filename string, data []byte) error {
	dirPath := filepath.Join(s.dataDir, id)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create id directory: %w", err)
	}

	filePath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load 加载配置
func (s *JSONStorage) Load(id string) (*SubscriptionConfig, error) {
	data, err := s.LoadFile(id, "core.json")
	if err != nil {
		return nil, err
	}

	var cfg SubscriptionConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// LoadFile 加载辅助文件
func (s *JSONStorage) LoadFile(id, filename string) ([]byte, error) {
	filePath := filepath.Join(s.dataDir, id, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s/%s", id, filename)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Delete 删除配置 (删除整个 ID 目录)
func (s *JSONStorage) Delete(id string) error {
	dirPath := filepath.Join(s.dataDir, id)

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to delete config directory: %w", err)
	}

	return nil
}

// ClearCache 清除该 ID 下的所有缓存文件 (保留 core.json)
func (s *JSONStorage) ClearCache(id string) error {
	dirPath := filepath.Join(s.dataDir, id)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "core.json" {
			os.Remove(filepath.Join(dirPath, entry.Name()))
		}
	}

	return nil
}
// List 列出所有配置
func (s *JSONStorage) List() ([]*SubscriptionConfig, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}
	
	var configs []*SubscriptionConfig
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		cfg, err := s.Load(id)
		if err != nil {
			continue
		}

		configs = append(configs, cfg)
	}
	
	return configs, nil
}

// Exists 检查配置是否存在
func (s *JSONStorage) Exists(id string) bool {
	dirPath := filepath.Join(s.dataDir, id)
	_, err := os.Stat(filepath.Join(dirPath, "core.json"))
	return err == nil
}

// ConfigStorage 为向后兼容保留的别名
type ConfigStorage = JSONStorage

// NewConfigStorage 为向后兼容保留的工厂函数
func NewConfigStorage(dataDir string) (*ConfigStorage, error) {
	return NewJSONStorage(dataDir)
}
