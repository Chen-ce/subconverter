package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/Chen-ce/subconverter/pkg/logger"
	"github.com/Chen-ce/subconverter/storage"
	"github.com/gin-gonic/gin"
)

var configStorage storage.Storage

// InitConfigStorage 初始化配置存储
func InitConfigStorage(cfgType, path, repoID, token string) error {
	var err error
	configStorage, err = storage.NewStorage(cfgType, path, repoID, token)
	return err
}

// CreateConfig 创建配置
func CreateConfig(c *gin.Context) {
	var req struct {
		URLs    []string `json:"urls" binding:"required"`
		Nodes   []string `json:"nodes"`
		Target  string   `json:"target" binding:"required"`
		Config  string   `json:"config"`
		Include string   `json:"include"`
		Exclude string   `json:"exclude"`
		Ver     int      `json:"ver"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 生成 ID
	id, err := configStorage.GenerateID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate ID"})
		return
	}
	
	// 创建配置
	cfg := &storage.SubscriptionConfig{
		ID:        id,
		URLs:      req.URLs,
		Nodes:     req.Nodes,
		Target:    req.Target,
		Config:    req.Config,
		Include:   req.Include,
		Exclude:   req.Exclude,
		Ver:       req.Ver,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}
	
	// 保存
	if err := configStorage.Save(cfg); err != nil {
		logger.Error("Failed to create config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config: " + err.Error()})
		return
	}
	
	// 返回结果
	c.JSON(http.StatusCreated, gin.H{
		"id":  id,
		"url": c.Request.Host + "/sub/" + id,
	})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	id := c.Param("id")
	
	cfg, err := configStorage.Load(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	
	c.JSON(http.StatusOK, cfg)
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	id := c.Param("id")
	
	// 检查配置是否存在
	cfg, err := configStorage.Load(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	
	// 解析请求
	var req struct {
		URLs    []string `json:"urls"`
		Nodes   []string `json:"nodes"`
		Target  string   `json:"target"`
		Config  string   `json:"config"`
		Include string   `json:"include"`
		Exclude string   `json:"exclude"`
		Ver     int      `json:"ver"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 更新字段
	if len(req.URLs) > 0 {
		cfg.URLs = req.URLs
	}
	cfg.Nodes = req.Nodes // 允许清空
	if req.Target != "" {
		cfg.Target = req.Target
	}
	cfg.Config = req.Config
	cfg.Include = req.Include
	cfg.Exclude = req.Exclude
	cfg.Ver = req.Ver
	
	// 保存
	if err := configStorage.Save(cfg); err != nil {
		logger.Error("Failed to update config", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, cfg)
}

// DeleteConfig 删除配置
func DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	
	if err := configStorage.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "config deleted"})
}

// ListConfigs 列出所有配置
func ListConfigs(c *gin.Context) {
	configs, err := configStorage.List()
	if err != nil {
		logger.Error("Failed to list configs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list configs: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// ConvertByID 通过 ID 转换订阅
func ConvertByID(c *gin.Context) {
	id := c.Param("id")
	
	// 加载配置
	cfg, err := configStorage.Load(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	
	// 转换参数
	var urlParam string
	if len(cfg.URLs) > 0 {
		urlParam = strings.Join(cfg.URLs, "|")
	}

	ver := ""
	if cfg.Ver > 0 {
		ver = fmt.Sprintf("%d", cfg.Ver)
	}

	// 执行转换逻辑（直接传递参数，绕过 Gin 的 Query 缓存）
	executeConversion(c, cfg.Target, urlParam, cfg.Nodes, cfg.Config, cfg.Include, cfg.Exclude, ver)
}
