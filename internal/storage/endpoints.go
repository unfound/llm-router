package storage

import (
	"database/sql"
	"time"

	"github.com/unfound/llm-router/internal/config"
)

// EndpointEntry 端点记录
type EndpointEntry struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	APIBase   string `json:"api_base"`
	APIKey    string `json:"api_key"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// EndpointStorage 端点存储操作
type EndpointStorage struct {
	db *sql.DB
}

// NewEndpointStorage 创建端点存储实例
func NewEndpointStorage() *EndpointStorage {
	return &EndpointStorage{db: GetDB()}
}

// GetAll 获取所有端点
func (s *EndpointStorage) GetAll() ([]EndpointEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, name, api_base, api_key, is_active, created_at
		FROM endpoints ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []EndpointEntry
	for rows.Next() {
		var e EndpointEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.APIBase, &e.APIKey, &e.IsActive, &e.CreatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, e)
	}
	return endpoints, nil
}

// GetByName 按名称获取端点
func (s *EndpointStorage) GetByName(name string) (*EndpointEntry, error) {
	var e EndpointEntry
	err := s.db.QueryRow(`
		SELECT id, name, api_base, api_key, is_active, created_at
		FROM endpoints WHERE name = ?
	`, name).Scan(&e.ID, &e.Name, &e.APIBase, &e.APIKey, &e.IsActive, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetByID 按ID获取端点
func (s *EndpointStorage) GetByID(id int64) (*EndpointEntry, error) {
	var e EndpointEntry
	err := s.db.QueryRow(`
		SELECT id, name, api_base, api_key, is_active, created_at
		FROM endpoints WHERE id = ?
	`, id).Scan(&e.ID, &e.Name, &e.APIBase, &e.APIKey, &e.IsActive, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Create 创建端点
func (s *EndpointStorage) Create(e *EndpointEntry) error {
	result, err := s.db.Exec(`
		INSERT INTO endpoints (name, api_base, api_key, is_active)
		VALUES (?, ?, ?, ?)
	`, e.Name, e.APIBase, e.APIKey, e.IsActive)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

// Update 更新端点
func (s *EndpointStorage) Update(e *EndpointEntry) error {
	_, err := s.db.Exec(`
		UPDATE endpoints SET name=?, api_base=?, api_key=?, is_active=?, updated_at=?
		WHERE id=?
	`, e.Name, e.APIBase, e.APIKey, e.IsActive, time.Now(), e.ID)
	return err
}

// Delete 删除端点
func (s *EndpointStorage) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM endpoints WHERE id = ?`, id)
	return err
}

// InitEndpointsFromConfig 从配置文件初始化端点（仅首次）
func InitEndpointsFromConfig(endpoints []config.EndpointConfig) error {
	es := NewEndpointStorage()
	existing, err := es.GetAll()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	for _, ep := range endpoints {
		entry := &EndpointEntry{
			Name:     ep.Name,
			APIBase:  ep.APIBase,
			APIKey:   ep.APIKey,
			IsActive: true,
		}
		if err := es.Create(entry); err != nil {
			return err
		}
	}
	return nil
}
