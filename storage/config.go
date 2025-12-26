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
	// 更新时间戳
	cfg.UpdatedAt = time.Now()
	cfg.Version++
	
	// 序列化为 JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// 写入文件
	filePath := filepath.Join(s.dataDir, cfg.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Load 加载配置
func (s *JSONStorage) Load(id string) (*SubscriptionConfig, error) {
	filePath := filepath.Join(s.dataDir, id+".json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var cfg SubscriptionConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &cfg, nil
}

// Delete 删除配置
func (s *JSONStorage) Delete(id string) error {
	filePath := filepath.Join(s.dataDir, id+".json")
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found: %s", id)
		}
		return fmt.Errorf("failed to delete config file: %w", err)
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
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		
		id := entry.Name()[:len(entry.Name())-5] // 去调 .json
		cfg, err := s.Load(id)
		if err != nil {
			continue // 跳过损坏的文件
		}
		
		configs = append(configs, cfg)
	}
	
	return configs, nil
}

// Exists 检查配置是否存在
func (s *JSONStorage) Exists(id string) bool {
	filePath := filepath.Join(s.dataDir, id+".json")
	_, err := os.Stat(filePath)
	return err == nil
}

// ConfigStorage 为向后兼容保留的别名
type ConfigStorage = JSONStorage

// NewConfigStorage 为向后兼容保留的工厂函数
func NewConfigStorage(dataDir string) (*ConfigStorage, error) {
	return NewJSONStorage(dataDir)
}
