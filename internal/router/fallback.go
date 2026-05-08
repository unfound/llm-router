package router

import (
	"log"
	"time"

	"github.com/unfound/llm-router/internal/config"
)

// FallbackChain 降级链
type FallbackChain struct {
	Primary  *config.ModelConfig
	Fallback *config.ModelConfig
}

// BuildChain 构建降级链
func BuildChain(alias string) (*FallbackChain, error) {
	primary, err := ResolveModel(alias)
	if err != nil {
		return nil, err
	}

	var fallback *config.ModelConfig
	if primary.Fallback != "" {
		fallback, _ = ResolveModel(primary.Fallback)
	}

	return &FallbackChain{
		Primary:  primary,
		Fallback: fallback,
	}, nil
}

// NextModel 获取下一个降级模型
func (c *FallbackChain) NextModel(current *config.ModelConfig) *config.ModelConfig {
	if c.Fallback != nil && current.Name == c.Primary.Name {
		return c.Fallback
	}
	return nil
}

// RetryWithBackoff 带指数退避的重试
func RetryWithBackoff(maxRetries int, fn func() error) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxRetries {
			wait := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s...
			log.Printf("请求失败，%v 后重试 (%d/%d): %v", wait, i+1, maxRetries, err)
			time.Sleep(wait)
		}
	}
	return err
}
