package config

import (
	"fmt"
	"os"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Auth      AuthConfig      `yaml:"auth"`
	Storage   StorageConfig   `yaml:"storage"`
	Templates TemplatesConfig `yaml:"templates"`
	Clash     ClashConfig     `yaml:"clash"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	WebPath  string `yaml:"web_path"` // Web 界面路径前缀，默认为空（根路径）
	LogLevel string `yaml:"log_level"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled bool     `yaml:"enabled"`
	APIKeys []string `yaml:"api_keys"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type   string `yaml:"type"`    // "json" or "hf"
	Path   string `yaml:"path"`    // 本地数据目录
	RepoID string `yaml:"repo_id"` // HF Dataset ID (e.g., username/repo)
	Token  string `yaml:"token"`   // HF Access Token
}

// TemplatesConfig 模板配置
type TemplatesConfig struct {
	Dir     string `yaml:"dir"`
	Default string `yaml:"default"`
}

// ClashConfig Clash 配置
type ClashConfig struct {
	DefaultRules string `yaml:"default_rules"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 展开环境变量
	content := os.ExpandEnv(string(data))
	
	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// 处理 API keys 中的环境变量
	for i, key := range cfg.Auth.APIKeys {
		cfg.Auth.APIKeys[i] = os.ExpandEnv(key)
	}
	
	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Set 设置全局配置
func Set(cfg *Config) {
	globalConfig = cfg
}


// IsValidAPIKey 验证 API 密钥
func (c *Config) IsValidAPIKey(key string) bool {
	if !c.Auth.Enabled {
		return true
	}
	
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	
	for _, validKey := range c.Auth.APIKeys {
		if validKey != "" && validKey == key {
			return true
		}
	}
	
	return false
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "default-api-key-please-change"
	}
	
	return &Config{
		Server: ServerConfig{
			Port:     8080,
			Host:     "0.0.0.0",
			WebPath:  "", // 默认根路径
			LogLevel: "INFO",
		},
		Auth: AuthConfig{
			Enabled: true,
			APIKeys: []string{apiKey},
		},
		Templates: TemplatesConfig{
			Dir:     "./templates",
			Default: "default",
		},
		Clash: ClashConfig{
			DefaultRules: "default",
		},
	}
}
