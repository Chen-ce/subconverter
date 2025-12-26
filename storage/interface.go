package storage

import (
	"time"
)

// SubscriptionConfig 订阅配置
type SubscriptionConfig struct {
	ID        string    `json:"id"`
	URLs      []string  `json:"urls"`
	Nodes     []string  `json:"nodes,omitempty"`
	Target    string    `json:"target"`
	Config    string    `json:"config,omitempty"`
	Include   string    `json:"include,omitempty"`
	Exclude   string    `json:"exclude,omitempty"`
	Ver       int       `json:"ver,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

// Storage 存储接口
type Storage interface {
	// GenerateID 生成随机 ID
	GenerateID() (string, error)
	
	// Save 保存配置
	Save(cfg *SubscriptionConfig) error
	
	// Load 加载配置
	Load(id string) (*SubscriptionConfig, error)
	
	// Delete 删除配置
	Delete(id string) error
	
	// List 列出所有配置
	List() ([]*SubscriptionConfig, error)
	
	// Exists 检查配置是否存在
	Exists(id string) bool
}
