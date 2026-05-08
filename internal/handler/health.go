package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

		// 尝试主模型 + 降级
		current := chain.Primary
		var resp *router.ProxyResponse

		for current != nil {
			log.Printf("转发请求: %s -> %s/%s", reqBody.Model, current.Provider, current.ModelID)

			// 创建日志记录（初始状态）
			logEntry := &storage.LogEntry{
				ModelName:  current.ModelID,
				AliasName:  reqBody.Model,
				Status:     "pending",
				CreatedAt:  startTime.Format("2006-01-02 15:04:05"),
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
				log.Printf("降级: %s -> %s/%s", current.Name, next.Provider, next.ModelID)
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

// HTTPError HTTP 错误
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return "HTTP 错误: " + http.StatusText(e.StatusCode)
}
