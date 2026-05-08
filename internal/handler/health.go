package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/storage"
)

// Health 健康检查
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ChatCompletions OpenAI 兼容的聊天补全接口
func ChatCompletions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现路由转发逻辑
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "路由转发尚未实现",
		})
	}
}

// ListModels 返回可用模型列表 (OpenAI 兼容格式)
func ListModels(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ms := storage.NewModelStorage()
		models, err := ms.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		data := make([]gin.H, 0)
		for _, m := range models {
			if m.IsActive {
				data = append(data, gin.H{
					"id":       m.Name,
					"object":   "model",
					"owned_by": m.Provider,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"object": "list",
			"data":   data,
		})
	}
}
