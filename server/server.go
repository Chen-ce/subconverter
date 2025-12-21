package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	
	s.setupRoutes()
	
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件（Web 界面）
	s.engine.Static("/", "./web")
	
	// 健康检查（无需认证）
	s.engine.GET("/health", handler.Health)
	
	// API 路由（需要认证）
	api := s.engine.Group("/api")
	api.Use(middleware.Auth())
	{
		api.GET("/sub", handler.Convert)
		api.GET("/templates", handler.Templates)
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
