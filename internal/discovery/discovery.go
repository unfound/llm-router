package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/unfound/llm-router/internal/storage"
)

// DiscoverModels 从端点自动发现可用模型
func DiscoverModels(endpoint *storage.EndpointEntry) ([]string, error) {
	url := strings.TrimRight(endpoint.APIBase, "/") + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+endpoint.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求端点失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("端点返回 %d: %s", resp.StatusCode, string(body))
	}

	// 解析 OpenAI 格式的模型列表
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

// SyncModels 同步端点发现的模型到数据库
func SyncModels(endpoint *storage.EndpointEntry) (added int, skipped int, err error) {
	models, err := DiscoverModels(endpoint)
	if err != nil {
		return 0, 0, err
	}

	ms := storage.NewModelStorage()
	for _, modelID := range models {
		exists, err := ms.ExistsByEndpointAndModel(endpoint.ID, modelID)
		if err != nil {
			continue
		}
		if exists {
			skipped++
			continue
		}

		// 自动发现的模型默认不启用，需要用户在配置文件中显式配置才会启用
		entry := &storage.ModelEntry{
			Name:       modelID,
			EndpointID: endpoint.ID,
			ModelID:    modelID,
			Discovered: true,
			IsActive:   false,
			MaxRetries: 2,
		}
		if err := ms.Create(entry); err != nil {
			log.Printf("保存发现模型失败 %s/%s: %v", endpoint.Name, modelID, err)
			continue
		}
		added++
		log.Printf("发现新模型: %s/%s", endpoint.Name, modelID)
	}
	return added, skipped, nil
}

// SyncAll 同步所有活跃端点的模型
func SyncAll() {
	es := storage.NewEndpointStorage()
	endpoints, err := es.GetAll()
	if err != nil {
		log.Printf("获取端点列表失败: %v", err)
		return
	}

	for _, ep := range endpoints {
		if !ep.IsActive {
			continue
		}
		added, skipped, err := SyncModels(&ep)
		if err != nil {
			log.Printf("端点 %s 模型发现失败: %v", ep.Name, err)
			continue
		}
		if added > 0 || skipped > 0 {
			log.Printf("端点 %s: 发现 %d 个新模型, %d 个已存在", ep.Name, added, skipped)
		}
	}
}
