package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/storage"
)

// AdminListModels 管理接口 - 获取模型列表
func AdminListModels(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ms := storage.NewModelStorage()
		models, err := ms.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"models": models,
		})
	}
}

// AdminCreateModel 管理接口 - 创建模型
func AdminCreateModel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var model config.ModelConfig
		if err := c.ShouldBindJSON(&model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ms := storage.NewModelStorage()
		if err := ms.Create(&model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "模型创建成功",
			"model":   model,
		})
	}
}

// AdminUpdateModel 管理接口 - 更新模型
func AdminUpdateModel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
			return
		}
		var model config.ModelConfig
		if err := c.ShouldBindJSON(&model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		model.ID = id
		ms := storage.NewModelStorage()
		if err := ms.Update(&model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "模型更新成功",
			"model":   model,
		})
	}
}

// AdminDeleteModel 管理接口 - 删除模型
func AdminDeleteModel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
			return
		}
		ms := storage.NewModelStorage()
		if err := ms.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "模型删除成功",
			"id":      id,
		})
	}
}

// AdminListLogs 管理接口 - 获取日志列表
func AdminListLogs(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Query("session_id")
		modelName := c.Query("model_name")
		status := c.Query("status")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		ls := storage.NewLogStorage()
		logs, total, err := ls.List(sessionID, modelName, status, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"logs":  logs,
			"total": total,
		})
	}
}

// AdminStatsOverview 管理接口 - 统计概览
func AdminStatsOverview(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ls := storage.NewLogStorage()
		totalRequests, successCount, failCount, avgLatency, totalTokens, err := ls.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total_requests": totalRequests,
			"success_count":  successCount,
			"fail_count":     failCount,
			"avg_latency_ms": avgLatency,
			"total_tokens":   totalTokens,
		})
	}
}
