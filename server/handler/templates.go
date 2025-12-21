package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/Chen-ce/subconverter/config"
	"github.com/Chen-ce/subconverter/templates"
)

// Templates 模板列表处理器
func Templates(c *gin.Context) {
	cfg := config.Get()
	tmplMgr := templates.NewManager(cfg.Templates.Dir)
	
	list, err := tmplMgr.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list templates",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"templates": list,
		"default": cfg.Clash.DefaultRules,
	})
}
