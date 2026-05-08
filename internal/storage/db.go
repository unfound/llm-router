package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init 初始化 SQLite 数据库
func Init(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return err
	}

	// 创建表
	if err := createTables(); err != nil {
		return err
	}

	log.Printf("数据库初始化完成: %s", dbPath)
	return nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return db
}

// createTables 创建数据库表
func createTables() error {
	// 端点表
	endpointsTable := `
	CREATE TABLE IF NOT EXISTS endpoints (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL UNIQUE,
		api_base    TEXT NOT NULL,
		api_key     TEXT NOT NULL,
		is_active   BOOLEAN DEFAULT 1,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 模型表（引用端点）
	modelsTable := `
	CREATE TABLE IF NOT EXISTS models (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL UNIQUE,
		endpoint_id INTEGER NOT NULL,
		model_id    TEXT NOT NULL,
		discovered  BOOLEAN DEFAULT 0,
		is_active   BOOLEAN DEFAULT 1,
		max_retries INTEGER DEFAULT 2,
		fallback    TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (endpoint_id) REFERENCES endpoints(id)
	);`

	// 请求日志表
	logsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id      TEXT,
		model_name      TEXT NOT NULL,
		alias_name      TEXT,
		request_tokens  INTEGER,
		response_tokens INTEGER,
		total_tokens    INTEGER,
		latency_ms      INTEGER,
		status          TEXT NOT NULL,
		error_message   TEXT,
		request_summary TEXT,
		response_summary TEXT,
		request_body    TEXT,
		response_body   TEXT,
		created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	for _, stmt := range []string{endpointsTable, modelsTable, logsTable} {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	// 创建索引
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_models_endpoint ON models(endpoint_id);`,
		`CREATE INDEX IF NOT EXISTS idx_logs_session ON logs(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_logs_model ON logs(model_name);`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created ON logs(created_at);`,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
