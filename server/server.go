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
	
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS())
	
	s := &Server{
		engine: engine,
		config: cfg,
	}
	
	// 初始化配置存储
	dataDir := "./data/configs"
	if err := handler.InitConfigStorage(dataDir); err != nil {
		fmt.Printf("Warning: failed to initialize config storage: %v\n", err)
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
		api.GET("/templates", handler.Templates)
		
		// 配置管理
		api.POST("/config", handler.CreateConfig)
		api.GET("/config/:id", handler.GetConfig)
		api.PUT("/config/:id", handler.UpdateConfig)
		api.DELETE("/config/:id", handler.DeleteConfig)
		api.GET("/configs", handler.ListConfigs)
	}
	
	// 短链接订阅（需要认证）
	s.engine.GET("/sub/:id", middleware.Auth(), handler.ConvertByID)
	
	// 兼容旧路径 /sub（无 ID）
	s.engine.GET("/sub", middleware.Auth(), handler.Convert)
	
	// 静态文件（Web 界面）
	if webPath == "" {
		// 根路径模式：直接在根路径提供 Web 界面
		s.engine.StaticFile("/", "./web/index.html")
		s.engine.StaticFile("/configs.html", "./web/configs.html")
		s.engine.StaticFile("/app.js", "./web/app.js")
		s.engine.StaticFile("/configs.js", "./web/configs.js")
		s.engine.StaticFile("/style.css", "./web/style.css")
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
		fmt.Printf("Starting server on %s\n", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("\nShutting down server...")
	
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	
	fmt.Println("Server stopped")
	return nil
}
