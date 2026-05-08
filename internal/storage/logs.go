package storage

import (
	"database/sql"
	"strings"
	"time"
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

// LogFilter 日志查询过滤条件
type LogFilter struct {
	SessionID string
	ModelName string
	Status    string
	Limit     int
	Offset    int
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
func (s *LogStorage) List(filter *LogFilter) ([]LogEntry, int, error) {
	query := `SELECT id, session_id, model_name, alias_name, request_tokens, response_tokens, total_tokens, latency_ms, status, error_message, request_summary, response_summary, created_at FROM logs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM logs WHERE 1=1`
	args := []interface{}{}

	if filter.SessionID != "" {
		query += ` AND session_id = ?`
		countQuery += ` AND session_id = ?`
		args = append(args, filter.SessionID)
	}
	if filter.ModelName != "" {
		query += ` AND model_name = ?`
		countQuery += ` AND model_name = ?`
		args = append(args, filter.ModelName)
	}
	if filter.Status != "" {
		query += ` AND status = ?`
		countQuery += ` AND status = ?`
		args = append(args, filter.Status)
	}

	// 获取总数
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页查询
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

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

// GetByID 按ID获取日志详情（包含完整内容）
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
func (s *LogStorage) UpdateResponse(id int64, responseTokens, totalTokens, latencyMs int, status, errorMessage, responseSummary, responseBody string) error {
	_, err := s.db.Exec(`
		UPDATE logs SET response_tokens=?, total_tokens=?, latency_ms=?, status=?, error_message=?, response_summary=?, response_body=? WHERE id=?
	`, responseTokens, totalTokens, latencyMs, status, errorMessage, responseSummary, responseBody, id)
	return err
}

// GetStats 获取全局统计
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

// GetModelStats 按模型维度统计
func (s *LogStorage) GetModelStats() ([]ModelStat, error) {
	rows, err := s.db.Query(`
		SELECT 
			COALESCE(alias_name, model_name) as display_name,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			COALESCE(AVG(CASE WHEN status = 'success' THEN latency_ms END), 0) as avg_latency,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM logs 
		GROUP BY COALESCE(alias_name, model_name)
		ORDER BY total DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ModelStat
	for rows.Next() {
		var stat ModelStat
		if err := rows.Scan(&stat.Name, &stat.Total, &stat.Success, &stat.Failed, &stat.AvgLatency, &stat.TotalTokens); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// ModelStat 模型统计
type ModelStat struct {
	Name        string  `json:"name"`
	Total       int     `json:"total"`
	Success     int     `json:"success"`
	Failed      int     `json:"failed"`
	AvgLatency  float64 `json:"avg_latency"`
	TotalTokens int     `json:"total_tokens"`
}

// GetTimeSeries 获取时间序列数据（按小时聚合）
func (s *LogStorage) GetTimeSeries(hours int) ([]TimeSeriesPoint, error) {
	// 计算起始时间
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour).Format("2006-01-02 15:04:05")

	rows, err := s.db.Query(`
		SELECT 
			strftime('%Y-%m-%d %H:00', created_at) as hour,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
			COALESCE(SUM(total_tokens), 0) as tokens
		FROM logs 
		WHERE created_at >= ?
		GROUP BY hour
		ORDER BY hour
	`, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Hour, &p.Total, &p.Success, &p.Tokens); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Hour    string `json:"hour"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Tokens  int    `json:"tokens"`
}

// GetSessions 获取会话列表
func (s *LogStorage) GetSessions() ([]SessionInfo, error) {
	rows, err := s.db.Query(`
		SELECT 
			session_id,
			COUNT(*) as request_count,
			MIN(created_at) as first_request,
			MAX(created_at) as last_request
		FROM logs 
		WHERE session_id != '' AND session_id IS NOT NULL
		GROUP BY session_id
		ORDER BY last_request DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var sess SessionInfo
		if err := rows.Scan(&sess.SessionID, &sess.RequestCount, &sess.FirstRequest, &sess.LastRequest); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID    string `json:"session_id"`
	RequestCount int    `json:"request_count"`
	FirstRequest string `json:"first_request"`
	LastRequest  string `json:"last_request"`
}

// ExtractSummary 从请求体提取摘要
func ExtractSummary(body []byte, maxLen int) string {
	if len(body) == 0 {
		return ""
	}
	s := string(body)
	if len(s) > maxLen {
		s = s[:maxLen] + "..."
	}
	// 尝试提取 messages 数组中的最后一条消息内容
	if idx := strings.LastIndex(s, `"content"`); idx > 0 {
		start := idx + len(`"content":`)
		if start < len(s) {
			end := strings.Index(s[start:], `"`)
			if end > 0 {
				return s[start : start+end]
			}
		}
	}
	return s
}
