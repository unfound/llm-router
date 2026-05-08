package storage

import (
	"database/sql"
	"time"

	"github.com/unfound/llm-router/internal/config"
)

// ModelEntry 模型记录
type ModelEntry struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	EndpointID int64  `json:"endpoint_id"`
	ModelID    string `json:"model_id"`
	Discovered bool   `json:"discovered"`
	IsActive   bool   `json:"is_active"`
	MaxRetries int    `json:"max_retries"`
	Fallback   string `json:"fallback"`
	CreatedAt  string `json:"created_at"`
}

// ModelStorage 模型存储操作
type ModelStorage struct {
	db *sql.DB
}

// NewModelStorage 创建模型存储实例
func NewModelStorage() *ModelStorage {
	return &ModelStorage{db: GetDB()}
}

// GetAll 获取所有模型（关联端点信息）
func (s *ModelStorage) GetAll() ([]ModelWithEndpoint, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.name, m.endpoint_id, m.model_id, m.discovered, m.is_active, m.max_retries, m.fallback,
		       e.name, e.api_base, e.api_key
		FROM models m
		JOIN endpoints e ON m.endpoint_id = e.id
		ORDER BY m.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []ModelWithEndpoint
	for rows.Next() {
		var m ModelWithEndpoint
		if err := rows.Scan(&m.ID, &m.Name, &m.EndpointID, &m.ModelID, &m.Discovered, &m.IsActive, &m.MaxRetries, &m.Fallback,
			&m.EndpointName, &m.APIBase, &m.APIKey); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// ModelWithEndpoint 带端点信息的模型
type ModelWithEndpoint struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	EndpointID   int64  `json:"endpoint_id"`
	ModelID      string `json:"model_id"`
	Discovered   bool   `json:"discovered"`
	IsActive     bool   `json:"is_active"`
	MaxRetries   int    `json:"max_retries"`
	Fallback     string `json:"fallback"`
	EndpointName string `json:"endpoint_name"`
	APIBase      string `json:"api_base"`
	APIKey       string `json:"api_key"`
}

// GetByName 按别名获取模型（含端点信息）
func (s *ModelStorage) GetByName(name string) (*ModelWithEndpoint, error) {
	var m ModelWithEndpoint
	err := s.db.QueryRow(`
		SELECT m.id, m.name, m.endpoint_id, m.model_id, m.discovered, m.is_active, m.max_retries, m.fallback,
		       e.name, e.api_base, e.api_key
		FROM models m
		JOIN endpoints e ON m.endpoint_id = e.id
		WHERE m.name = ? AND m.is_active = 1
	`, name).Scan(&m.ID, &m.Name, &m.EndpointID, &m.ModelID, &m.Discovered, &m.IsActive, &m.MaxRetries, &m.Fallback,
		&m.EndpointName, &m.APIBase, &m.APIKey)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByID 按ID获取模型
func (s *ModelStorage) GetByID(id int64) (*ModelWithEndpoint, error) {
	var m ModelWithEndpoint
	err := s.db.QueryRow(`
		SELECT m.id, m.name, m.endpoint_id, m.model_id, m.discovered, m.is_active, m.max_retries, m.fallback,
		       e.name, e.api_base, e.api_key
		FROM models m
		JOIN endpoints e ON m.endpoint_id = e.id
		WHERE m.id = ?
	`, id).Scan(&m.ID, &m.Name, &m.EndpointID, &m.ModelID, &m.Discovered, &m.IsActive, &m.MaxRetries, &m.Fallback,
		&m.EndpointName, &m.APIBase, &m.APIKey)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create 创建模型
func (s *ModelStorage) Create(m *ModelEntry) error {
	result, err := s.db.Exec(`
		INSERT INTO models (name, endpoint_id, model_id, discovered, is_active, max_retries, fallback)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, m.Name, m.EndpointID, m.ModelID, m.Discovered, m.IsActive, m.MaxRetries, m.Fallback)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = id
	return nil
}

// Update 更新模型
func (s *ModelStorage) Update(m *ModelEntry) error {
	_, err := s.db.Exec(`
		UPDATE models SET name=?, endpoint_id=?, model_id=?, discovered=?, is_active=?, max_retries=?, fallback=?, updated_at=?
		WHERE id=?
	`, m.Name, m.EndpointID, m.ModelID, m.Discovered, m.IsActive, m.MaxRetries, m.Fallback, time.Now(), m.ID)
	return err
}

// Delete 删除模型
func (s *ModelStorage) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM models WHERE id = ?`, id)
	return err
}

// ToggleActive 切换模型启用状态
func (s *ModelStorage) ToggleActive(id int64) error {
	_, err := s.db.Exec(`UPDATE models SET is_active = NOT is_active, updated_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

// ExistsByEndpointAndModel 检查模型是否已存在（按端点+model_id去重）
func (s *ModelStorage) ExistsByEndpointAndModel(endpointID int64, modelID string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM models WHERE endpoint_id = ? AND model_id = ?`, endpointID, modelID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// InitModelsFromConfig 从配置文件初始化模型路由（仅首次）
func InitModelsFromConfig(models []config.ModelConfig) error {
	ms := NewModelStorage()
	es := NewEndpointStorage()

	existing, err := ms.GetAll()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	for _, m := range models {
		// 查找端点 ID
		ep, err := es.GetByName(m.Endpoint)
		if err != nil {
			continue
		}

		entry := &ModelEntry{
			Name:       m.Name,
			EndpointID: ep.ID,
			ModelID:    m.ModelID,
			Discovered: false,
			IsActive:   m.IsActive,
			MaxRetries: m.MaxRetries,
			Fallback:   m.Fallback,
		}
		if err := ms.Create(entry); err != nil {
			return err
		}
	}
	return nil
}
