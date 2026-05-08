package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/handler"
	"github.com/unfound/llm-router/internal/storage"
)

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

	// 从配置文件导入模型（仅首次）
	if err := storage.InitFromConfig(cfg.Models); err != nil {
		log.Fatalf("导入模型配置失败: %v", err)
	}

	// 路由服务 (转发)
	go startRouteServer(cfg)

	// 管理服务
	startAdminServer(cfg)
}

// startRouteServer 启动路由转发服务
func startRouteServer(cfg *config.Config) {
	r := gin.Default()

	// 健康检查
	r.GET("/health", handler.Health)

	// OpenAI 兼容接口
	r.POST("/v1/chat/completions", handler.ChatCompletions(cfg))
	r.GET("/v1/models", handler.ListModels(cfg))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("路由服务启动: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("路由服务启动失败: %v", err)
	}
}

// startAdminServer 启动管理服务
func startAdminServer(cfg *config.Config) {
	r := gin.Default()

	// 模型管理 API
	r.GET("/admin/api/models", handler.AdminListModels(cfg))
	r.POST("/admin/api/models", handler.AdminCreateModel(cfg))
	r.PUT("/admin/api/models/:id", handler.AdminUpdateModel(cfg))
	r.DELETE("/admin/api/models/:id", handler.AdminDeleteModel(cfg))

	// 日志查询 API
	r.GET("/admin/api/logs", handler.AdminListLogs(cfg))

	// 统计 API
	r.GET("/admin/api/stats/overview", handler.AdminStatsOverview(cfg))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.AdminPort)
	log.Printf("管理服务启动: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("管理服务启动失败: %v", err)
	}
}
