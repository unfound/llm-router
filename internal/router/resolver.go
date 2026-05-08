package router

import (
	"errors"

	"github.com/unfound/llm-router/internal/storage"
)

var (
	ErrModelNotFound = errors.New("模型不存在或未启用")
)

// ResolveModel 别名解析 - 将模型别名解析为完整配置（含端点信息）
func ResolveModel(alias string) (*storage.ModelWithEndpoint, error) {
	ms := storage.NewModelStorage()
	m, err := ms.GetByName(alias)
	if err != nil {
		return nil, ErrModelNotFound
	}
	return m, nil
}
