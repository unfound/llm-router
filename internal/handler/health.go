package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/router"
	"github.com/unfound/llm-router/internal/storage"
)

// Health 健康检查
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ChatCompletions OpenAI 兼容的聊天补全接口
func ChatCompletions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
			return
		}

		// 解析请求体获取 model 字段
		var reqBody struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
			return
		}

		if reqBody.Model == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 model 字段"})
			return
		}

		// 别名解析
		chain, err := router.BuildChain(reqBody.Model)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "模型不存在或未启用",
				"model": reqBody.Model,
			})
			return
		}

		// 提取请求摘要
		requestSummary := storage.ExtractSummary(body, 200)

		// 尝试主模型 + 降级
		current := chain.Primary
		var resp *router.ProxyResponse

		for current != nil {
			log.Printf("转发请求: %s -> %s/%s", reqBody.Model, current.EndpointName, current.ModelID)

			// 创建日志记录（初始状态）
			logEntry := &storage.LogEntry{
				ModelName:      current.ModelID,
				AliasName:      reqBody.Model,
				RequestSummary: requestSummary,
				Status:         "pending",
				CreatedAt:      startTime.Format("2006-01-02 15:04:05"),
			}

			// 完整请求内容（根据配置决定是否记录）
			if cfg.Storage.LogFullContent {
				logEntry.RequestBody = string(body)
			}

			// 带重试的转发
			err := router.RetryWithBackoff(current.MaxRetries, func() error {
				req := &router.ProxyRequest{
					Model:     current,
					Body:      body,
					IsStream:  reqBody.Stream,
					StartTime: startTime,
				}
				resp = router.Forward(c.Writer, req)
				if resp.Error != nil {
					return resp.Error
				}
				if resp.StatusCode >= 400 {
					return &HTTPError{StatusCode: resp.StatusCode}
				}
				return nil
			})

			if err == nil {
				// 成功
				logEntry.ResponseTokens = resp.ResponseTokens
				logEntry.TotalTokens = resp.TotalTokens
				logEntry.LatencyMs = int(resp.LatencyMs)
				logEntry.Status = "success"

				// 提取响应摘要
				if len(resp.Body) > 0 {
					logEntry.ResponseSummary = storage.ExtractSummary(resp.Body, 200)
					if cfg.Storage.LogFullContent {
						logEntry.ResponseBody = string(resp.Body)
					}
				}

				ls := storage.NewLogStorage()
				ls.Create(logEntry)
				return
			}

			// 失败，记录日志
			logEntry.Status = "failed"
			logEntry.ErrorMessage = err.Error()
			logEntry.LatencyMs = int(time.Since(startTime).Milliseconds())
			ls := storage.NewLogStorage()
			ls.Create(logEntry)

			// 尝试降级
			next := chain.NextModel(current)
			if next != nil {
				log.Printf("降级: %s -> %s/%s", current.Name, next.EndpointName, next.ModelID)
				current = next
			} else {
				break
			}
		}

		// 全部失败
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "所有模型均不可用",
			"details": "已尝试所有降级路径",
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
					"owned_by": m.EndpointName,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"object": "list",
			"data":   data,
		})
	}
}

// HTTPError HTTP 错误
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return "HTTP 错误: " + http.StatusText(e.StatusCode)
}

// AdminListLogs 管理接口 - 获取日志列表
func AdminListLogs(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := &storage.LogFilter{
			SessionID: c.Query("session_id"),
			ModelName: c.Query("model_name"),
			Status:    c.Query("status"),
			Limit:     20,
			Offset:    0,
		}

		if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
			filter.Limit = l
		}
		if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
			filter.Offset = o
		}

		ls := storage.NewLogStorage()
		logs, total, err := ls.List(filter)
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

// AdminGetLog 管理接口 - 获取日志详情
func AdminGetLog(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日志ID"})
			return
		}

		ls := storage.NewLogStorage()
		entry, err := ls.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "日志不存在"})
			return
		}
		c.JSON(http.StatusOK, entry)
	}
}
