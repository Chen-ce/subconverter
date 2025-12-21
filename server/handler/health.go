package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// Health 健康检查处理器
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "subconverter is running",
	})
}
