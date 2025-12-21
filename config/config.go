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
	Templates TemplatesConfig `yaml:"templates"`
	Clash     ClashConfig     `yaml:"clash"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled bool     `yaml:"enabled"`
	APIKeys []string `yaml:"api_keys"`
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
			Port: 8080,
			Host: "0.0.0.0",
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
