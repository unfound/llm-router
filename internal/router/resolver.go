package router

import (
	"errors"

	"github.com/unfound/llm-router/internal/config"
	"github.com/unfound/llm-router/internal/storage"
)

var (
	ErrModelNotFound = errors.New("模型不存在或未启用")
)

// ResolveModel 别名解析 - 将模型别名解析为真实配置
func ResolveModel(alias string) (*config.ModelConfig, error) {
	ms := storage.NewModelStorage()
	m, err := ms.GetByName(alias)
	if err != nil {
		return nil, ErrModelNotFound
	}
	return m, nil
}
