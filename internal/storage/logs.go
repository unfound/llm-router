package storage

import (
	"database/sql"
	"time"

	"github.com/unfound/llm-router/internal/config"
)

// LogEntry 日志条目
type LogEntry struct {
	ID              int64  `json:"id"`
	SessionID       string `json:"session_id"`
	ModelName       string `json:"model_name"`
	AliasName       string `json:"alias_name"`
	RequestTokens   int    `json:"request_tokens"`
	ResponseTokens  int    `json:"response_tokens"`
	TotalTokens     int    `json:"total_tokens"`
	LatencyMs       int    `json:"latency_ms"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message"`
	RequestSummary  string `json:"request_summary"`
	ResponseSummary string `json:"response_summary"`
	RequestBody     string `json:"request_body,omitempty"`
	ResponseBody    string `json:"response_body,omitempty"`
	CreatedAt       string `json:"created_at"`
}

// LogStorage 日志存储操作
type LogStorage struct {
	db *sql.DB
}

// NewLogStorage 创建日志存储实例
func NewLogStorage() *LogStorage {
	return &LogStorage{db: GetDB()}
}

// Create 创建日志
func (s *LogStorage) Create(entry *LogEntry) error {
	result, err := s.db.Exec(`
		INSERT INTO logs (session_id, model_name, alias_name, request_tokens, response_tokens, total_tokens, latency_ms, status, error_message, request_summary, response_summary, request_body, response_body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.SessionID, entry.ModelName, entry.AliasName, entry.RequestTokens, entry.ResponseTokens, entry.TotalTokens, entry.LatencyMs, entry.Status, entry.ErrorMessage, entry.RequestSummary, entry.ResponseSummary, entry.RequestBody, entry.ResponseBody)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = id
	return nil
}

// List 查询日志列表
func (s *LogStorage) List(sessionID, modelName, status string, limit, offset int) ([]LogEntry, int, error) {
	query := `SELECT id, session_id, model_name, alias_name, request_tokens, response_tokens, total_tokens, latency_ms, status, error_message, request_summary, response_summary, created_at FROM logs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM logs WHERE 1=1`
	args := []interface{}{}

	if sessionID != "" {
		query += ` AND session_id = ?`
		countQuery += ` AND session_id = ?`
		args = append(args, sessionID)
	}
	if modelName != "" {
		query += ` AND model_name = ?`
		countQuery += ` AND model_name = ?`
		args = append(args, modelName)
	}
	if status != "" {
		query += ` AND status = ?`
		countQuery += ` AND status = ?`
		args = append(args, status)
	}

	// 获取总数
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页查询
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.SessionID, &entry.ModelName, &entry.AliasName, &entry.RequestTokens, &entry.ResponseTokens, &entry.TotalTokens, &entry.LatencyMs, &entry.Status, &entry.ErrorMessage, &entry.RequestSummary, &entry.ResponseSummary, &entry.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, entry)
	}
	return logs, total, nil
}

// GetByID 按ID获取日志详情
func (s *LogStorage) GetByID(id int64) (*LogEntry, error) {
	var entry LogEntry
	err := s.db.QueryRow(`
		SELECT id, session_id, model_name, alias_name, request_tokens, response_tokens, total_tokens, latency_ms, status, error_message, request_summary, response_summary, request_body, response_body, created_at
		FROM logs WHERE id = ?
	`, id).Scan(&entry.ID, &entry.SessionID, &entry.ModelName, &entry.AliasName, &entry.RequestTokens, &entry.ResponseTokens, &entry.TotalTokens, &entry.LatencyMs, &entry.Status, &entry.ErrorMessage, &entry.RequestSummary, &entry.ResponseSummary, &entry.RequestBody, &entry.ResponseBody, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// UpdateResponse 更新日志响应信息
func (s *LogStorage) UpdateResponse(id int64, responseTokens, totalTokens, latencyMs int, status, errorMessage string) error {
	_, err := s.db.Exec(`
		UPDATE logs SET response_tokens=?, total_tokens=?, latency_ms=?, status=?, error_message=? WHERE id=?
	`, responseTokens, totalTokens, latencyMs, status, errorMessage, id)
	return err
}

// GetStats 获取统计信息
func (s *LogStorage) GetStats() (totalRequests int, successCount int, failCount int, avgLatency float64, totalTokens int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM logs`).Scan(&totalRequests)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE status = 'success'`).Scan(&successCount)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE status = 'failed'`).Scan(&failCount)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COALESCE(AVG(latency_ms), 0) FROM logs WHERE status = 'success'`).Scan(&avgLatency)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COALESCE(SUM(total_tokens), 0) FROM logs`).Scan(&totalTokens)
	return
}

// InitFromConfig 从配置文件初始化模型数据
func InitFromConfig(models []config.ModelConfig) error {
	ms := NewModelStorage()
	existing, err := ms.GetAll()
	if err != nil {
		return err
	}

	// 如果数据库已有数据，跳过
	if len(existing) > 0 {
		return nil
	}

	// 从配置文件导入
	for _, m := range models {
		if err := ms.Create(&m); err != nil {
			return err
		}
	}
	return nil
}

// 日志创建时间字段
var _ = time.Now
