package storage

import (
	"database/sql"
	"time"

	"github.com/unfound/llm-router/internal/config"
)

// ModelStorage 模型存储操作
type ModelStorage struct {
	db *sql.DB
}

// NewModelStorage 创建模型存储实例
func NewModelStorage() *ModelStorage {
	return &ModelStorage{db: GetDB()}
}

// GetAll 获取所有模型
func (s *ModelStorage) GetAll() ([]config.ModelConfig, error) {
	rows, err := s.db.Query(`
		SELECT id, name, provider, model_id, api_base, api_key, is_active, max_retries, fallback
		FROM models ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []config.ModelConfig
	for rows.Next() {
		var m config.ModelConfig
		if err := rows.Scan(&m.ID, &m.Name, &m.Provider, &m.ModelID, &m.APIBase, &m.APIKey, &m.IsActive, &m.MaxRetries, &m.Fallback); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// GetByName 按名称获取模型
func (s *ModelStorage) GetByName(name string) (*config.ModelConfig, error) {
	var m config.ModelConfig
	err := s.db.QueryRow(`
		SELECT id, name, provider, model_id, api_base, api_key, is_active, max_retries, fallback
		FROM models WHERE name = ? AND is_active = 1
	`, name).Scan(&m.ID, &m.Name, &m.Provider, &m.ModelID, &m.APIBase, &m.APIKey, &m.IsActive, &m.MaxRetries, &m.Fallback)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByID 按ID获取模型
func (s *ModelStorage) GetByID(id int64) (*config.ModelConfig, error) {
	var m config.ModelConfig
	err := s.db.QueryRow(`
		SELECT id, name, provider, model_id, api_base, api_key, is_active, max_retries, fallback
		FROM models WHERE id = ?
	`, id).Scan(&m.ID, &m.Name, &m.Provider, &m.ModelID, &m.APIBase, &m.APIKey, &m.IsActive, &m.MaxRetries, &m.Fallback)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create 创建模型
func (s *ModelStorage) Create(m *config.ModelConfig) error {
	result, err := s.db.Exec(`
		INSERT INTO models (name, provider, model_id, api_base, api_key, is_active, max_retries, fallback)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, m.Name, m.Provider, m.ModelID, m.APIBase, m.APIKey, m.IsActive, m.MaxRetries, m.Fallback)
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
func (s *ModelStorage) Update(m *config.ModelConfig) error {
	_, err := s.db.Exec(`
		UPDATE models SET name=?, provider=?, model_id=?, api_base=?, api_key=?, is_active=?, max_retries=?, fallback=?, updated_at=?
		WHERE id=?
	`, m.Name, m.Provider, m.ModelID, m.APIBase, m.APIKey, m.IsActive, m.MaxRetries, m.Fallback, time.Now(), m.ID)
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
