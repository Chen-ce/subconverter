package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/Chen-ce/subconverter/config"
)

// Auth API 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		
		// 如果认证未启用，直接通过
		if !cfg.Auth.Enabled {
			c.Next()
			return
		}
		
		// 从 Header 获取 token
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		
		// 如果 Header 中没有，尝试从 URL 参数获取
		if token == "" {
			token = c.Query("token")
		}
		
		// 验证 token
		if !cfg.IsValidAPIKey(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
				"message": "invalid or missing API key",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// HeaderOnlyAuth 严格仅限 Header 认证的中间件
func HeaderOnlyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		if !cfg.Auth.Enabled {
			c.Next()
			return
		}

		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" || !cfg.IsValidAPIKey(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "this endpoint requires a valid API key in the Authorization header",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractTokenFromHeader 从 Authorization header 提取 token
func extractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	
	// 支持 "Bearer <token>" 格式
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return strings.TrimSpace(parts[1])
	}
	
	// 也支持直接传递 token
	return strings.TrimSpace(authHeader)
}
