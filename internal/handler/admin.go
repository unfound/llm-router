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

// AdminToggleModel 管理接口 - 切换模型启用状态
func AdminToggleModel(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
			return
		}
		ms := storage.NewModelStorage()
		if err := ms.ToggleActive(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "状态切换成功",
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

		successRate := 0.0
		if totalRequests > 0 {
			successRate = float64(successCount) / float64(totalRequests) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"total_requests": totalRequests,
			"success_count":  successCount,
			"fail_count":     failCount,
			"success_rate":   successRate,
			"avg_latency_ms": avgLatency,
			"total_tokens":   totalTokens,
		})
	}
}

// AdminStatsModels 管理接口 - 按模型维度统计
func AdminStatsModels(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ls := storage.NewLogStorage()
		stats, err := ls.GetModelStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"models": stats,
		})
	}
}

// AdminStatsTimeSeries 管理接口 - 时间序列数据
func AdminStatsTimeSeries(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		hours := 24
		if h, err := strconv.Atoi(c.DefaultQuery("hours", "24")); err == nil {
			hours = h
		}

		ls := storage.NewLogStorage()
		points, err := ls.GetTimeSeries(hours)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"timeseries": points,
		})
	}
}

// AdminSessions 管理接口 - 获取会话列表
func AdminSessions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ls := storage.NewLogStorage()
		sessions, err := ls.GetSessions()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"sessions": sessions,
		})
	}
}
