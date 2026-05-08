package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/discovery"
	"github.com/unfound/llm-router/internal/handler"
	"github.com/unfound/llm-router/internal/storage"
)

//go:embed all:web
var webFS embed.FS

func main() {
	// 加载配置
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 确保数据目录存在
	dbDir := filepath.Dir(cfg.Storage.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 初始化数据库
	if err := storage.Init(cfg.Storage.DBPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 从配置文件初始化端点（仅首次）
	if err := storage.InitEndpointsFromConfig(cfg.Endpoints); err != nil {
		log.Fatalf("导入端点配置失败: %v", err)
	}

	// 从配置文件初始化模型路由（仅首次）
	if err := storage.InitModelsFromConfig(cfg.Models); err != nil {
		log.Fatalf("导入模型配置失败: %v", err)
	}

	// 启动时自动发现模型
	go func() {
		log.Println("开始自动发现模型...")
		discovery.SyncAll()
		log.Println("模型发现完成")
	}()

	// 路由服务 (转发 + 管理面板)
	go startRouteServer(cfg)

	// 管理 API 服务
	startAdminServer(cfg)
}

// startRouteServer 启动路由转发服务（同时托管前端）
func startRouteServer(cfg *config.Config) {
	r := gin.Default()

	// 健康检查
	r.GET("/health", handler.Health)

	// OpenAI 兼容接口
	r.POST("/v1/chat/completions", handler.ChatCompletions(cfg))
	r.GET("/v1/models", handler.ListModels(cfg))

	// 管理 API（挂在同一个端口下，方便前端访问）
	adminAPI := r.Group("/admin/api")
	{
		adminAPI.GET("/endpoints", handler.AdminListEndpoints(cfg))
		adminAPI.POST("/endpoints", handler.AdminCreateEndpoint(cfg))
		adminAPI.DELETE("/endpoints/:id", handler.AdminDeleteEndpoint(cfg))

		adminAPI.GET("/models", handler.AdminListModels(cfg))
		adminAPI.POST("/models", handler.AdminCreateModel(cfg))
		adminAPI.PUT("/models/:id", handler.AdminUpdateModel(cfg))
		adminAPI.DELETE("/models/:id", handler.AdminDeleteModel(cfg))
		adminAPI.PUT("/models/:id/toggle", handler.AdminToggleModel(cfg))
		adminAPI.POST("/models/sync", handler.AdminSyncModels(cfg))

		adminAPI.GET("/logs", handler.AdminListLogs(cfg))
		adminAPI.GET("/logs/:id", handler.AdminGetLog(cfg))

		adminAPI.GET("/stats/overview", handler.AdminStatsOverview(cfg))
		adminAPI.GET("/stats/models", handler.AdminStatsModels(cfg))
		adminAPI.GET("/stats/timeseries", handler.AdminStatsTimeSeries(cfg))

		adminAPI.GET("/sessions", handler.AdminSessions(cfg))
	}

	// 托管前端静态文件
	distFS, err := fs.Sub(webFS, "web")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))

		// 静态资源
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 尝试静态文件
			if f, err := distFS.(fs.ReadFileFS).ReadFile(path[1:]); err == nil {
				// 根据扩展名设置 Content-Type
				ext := filepath.Ext(path)
				contentType := "application/octet-stream"
				switch ext {
				case ".js":
					contentType = "application/javascript"
				case ".css":
					contentType = "text/css"
				case ".html":
					contentType = "text/html"
				case ".json":
					contentType = "application/json"
				case ".png":
					contentType = "image/png"
				case ".svg":
					contentType = "image/svg+xml"
				}
				c.Data(http.StatusOK, contentType, f)
				return
			}

			// SPA fallback: 所有非 API 路径都返回 index.html
			if len(path) < 5 || path[:5] != "/v1/" {
				if indexHTML, err := distFS.(fs.ReadFileFS).ReadFile("index.html"); err == nil {
					c.Data(http.StatusOK, "text/html", indexHTML)
					return
				}
			}

			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务启动: %s（路由 + 管理面板）", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// startAdminServer 启动独立管理 API 服务（可选）
func startAdminServer(cfg *config.Config) {
	r := gin.Default()

	// 管理 API
	r.GET("/admin/api/endpoints", handler.AdminListEndpoints(cfg))
	r.POST("/admin/api/endpoints", handler.AdminCreateEndpoint(cfg))
	r.DELETE("/admin/api/endpoints/:id", handler.AdminDeleteEndpoint(cfg))

	r.GET("/admin/api/models", handler.AdminListModels(cfg))
	r.POST("/admin/api/models", handler.AdminCreateModel(cfg))
	r.PUT("/admin/api/models/:id", handler.AdminUpdateModel(cfg))
	r.DELETE("/admin/api/models/:id", handler.AdminDeleteModel(cfg))
	r.PUT("/admin/api/models/:id/toggle", handler.AdminToggleModel(cfg))
	r.POST("/admin/api/models/sync", handler.AdminSyncModels(cfg))

	r.GET("/admin/api/logs", handler.AdminListLogs(cfg))
	r.GET("/admin/api/logs/:id", handler.AdminGetLog(cfg))

	r.GET("/admin/api/stats/overview", handler.AdminStatsOverview(cfg))
	r.GET("/admin/api/stats/models", handler.AdminStatsModels(cfg))
	r.GET("/admin/api/stats/timeseries", handler.AdminStatsTimeSeries(cfg))

	r.GET("/admin/api/sessions", handler.AdminSessions(cfg))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.AdminPort)
	log.Printf("管理 API 启动: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("管理 API 启动失败: %v", err)
	}
}
