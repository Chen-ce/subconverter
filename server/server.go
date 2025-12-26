package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/Chen-ce/subconverter/config"
	"github.com/Chen-ce/subconverter/pkg/logger"
	"github.com/Chen-ce/subconverter/server/handler"
	"github.com/Chen-ce/subconverter/server/middleware"
)

// Server HTTP 服务器
type Server struct {
	engine *gin.Engine
	server *http.Server
	config *config.Config
}

// New 创建服务器
func New(cfg *config.Config) *Server {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)
	
	// 初始化日志
	logger.Init(cfg.Server.LogLevel)
	logger.Info("Logger initialized", "level", cfg.Server.LogLevel)
	
	engine := gin.New()
	engine.Use(gin.Logger()) // 启用请求日志，方便排查 API 错误
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS())
	
	s := &Server{
		engine: engine,
		config: cfg,
	}
	
	// 初始化配置存储
	storageType := cfg.Storage.Type
	storagePath := cfg.Storage.Path
	if storagePath == "" {
		storagePath = "./data/configs"
	}
	hfRepoID := cfg.Storage.RepoID
	hfToken := cfg.Storage.Token
	
	if err := handler.InitConfigStorage(storageType, storagePath, hfRepoID, hfToken); err != nil {
		logger.Error("Failed to initialize config storage", "error", err)
	} else {
		if storageType == "hf" {
			logger.Info("Storage initialized", "type", storageType, "repo", hfRepoID)
		} else {
			logger.Info("Storage initialized", "type", storageType, "path", storagePath)
		}
	}
	
	s.setupRoutes()
	
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查（无需认证）
	s.engine.GET("/health", handler.Health)
	
	// 获取 Web 路径配置
	webPath := s.config.Server.WebPath
	if webPath != "" && !strings.HasPrefix(webPath, "/") {
		webPath = "/" + webPath
	}
	
	// API 路由（需要认证）
	api := s.engine.Group("/api")
	api.Use(middleware.Auth())
	{
		// 订阅转换
		api.GET("/sub", handler.Convert)

		// 编程调用接口 (专门供外部项目使用，强制要求 Header 认证)
		api.POST("/convert", middleware.HeaderOnlyAuth(), handler.PureConvert)
		api.GET("/templates", middleware.HeaderOnlyAuth(), handler.Templates)
		
		// 配置管理
		api.POST("/config", handler.CreateConfig)
		api.GET("/config/:id", handler.GetConfig)
		api.PUT("/config/:id", handler.UpdateConfig)
		api.DELETE("/config/:id", handler.DeleteConfig)
		api.GET("/configs", handler.ListConfigs)
	}
	
	// 短链接订阅（需要认证）
	s.engine.GET("/sub/:id", middleware.Auth(), handler.ConvertByID)
	
	// 兼容旧路径 /sub（无 ID）- 放在 API 路由之外以保持兼容性
	s.engine.GET("/sub", middleware.Auth(), handler.Convert)
	
	// 静态文件（Web 界面）
	if webPath == "" {
		// 根路径模式：直接在根路径提供 Web 界面
		// 先注册具体的静态文件
		s.engine.StaticFile("/index.html", "./web/index.html")
		s.engine.StaticFile("/configs.html", "./web/configs.html")
		s.engine.StaticFile("/health.html", "./web/health.html")
		s.engine.StaticFile("/app.js", "./web/app.js")
		s.engine.StaticFile("/configs.js", "./web/configs.js")
		s.engine.StaticFile("/toast.js", "./web/toast.js")
		s.engine.StaticFile("/style.css", "./web/style.css")
		
		// 根路径重定向到 index.html
		s.engine.GET("/", func(c *gin.Context) {
			c.File("./web/index.html")
		})
	} else {
		// 自定义路径模式：在指定路径下提供 Web 界面
		s.engine.Static(webPath, "./web")
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	
	// 启动服务器
	go func() {
		logger.Info("Starting server", "addr", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")
	
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	
	logger.Info("Server stopped")
	return nil
}
