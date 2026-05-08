package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Models  []ModelConfig `yaml:"models"`
	Headers HeaderConfig  `yaml:"headers"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	AdminPort int    `yaml:"admin_port"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DBPath         string `yaml:"db_path"`
	LogFullContent bool   `yaml:"log_full_content"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	ID         int64  `json:"id" yaml:"-"`
	Name       string `json:"name" yaml:"name"`
	Provider   string `json:"provider" yaml:"provider"`
	ModelID    string `json:"model_id" yaml:"model_id"`
	APIBase    string `json:"api_base" yaml:"api_base"`
	APIKey     string `json:"api_key" yaml:"api_key"`
	IsActive   bool   `json:"is_active" yaml:"is_active"`
	MaxRetries int    `json:"max_retries" yaml:"max_retries"`
	Fallback   string `json:"fallback" yaml:"fallback"`
}

// HeaderConfig 请求头配置
type HeaderConfig struct {
	ModelOverride string `yaml:"model_override"`
	SessionID     string `yaml:"session_id"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.AdminPort == 0 {
		cfg.Server.AdminPort = 8081
	}
	if cfg.Storage.DBPath == "" {
		cfg.Storage.DBPath = "./data/llm-router.db"
	}
	if cfg.Headers.ModelOverride == "" {
		cfg.Headers.ModelOverride = "X-Msf-Model"
	}
	if cfg.Headers.SessionID == "" {
		cfg.Headers.SessionID = "X-Msf-Session-Id"
	}

	return cfg, nil
}
