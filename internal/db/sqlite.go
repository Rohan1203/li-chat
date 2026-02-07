package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"li-chat/pkg/logger"
)

func NewSQLite(path string) *sql.DB {
	logger.Info("[DB::INIT] Opening SQLite database at path: %s", path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Error("[DB::INIT] Failed to open database: %v", err)
		logger.Warn("[DB::INIT] Database initialization failed, application may not function correctly")
		panic(err)
	}
	logger.Debug("[DB::INIT] Database driver initialized for path: %s", path)
	logger.Info("[DB::INIT] Database connection established successfully")

	logger.Debug("[DB::SCHEMA] Starting schema creation process...")
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	if _, err := db.Exec(schema); err != nil {
		logger.Error("[DB::SCHEMA] Failed to create database schema: %v", err)
		logger.Warn("[DB::SCHEMA] Schema creation error - tables may be missing or corrupted")
		panic(err)
	}
	logger.Debug("[DB::SCHEMA] All tables created or already exist")
	logger.Info("[DB::SCHEMA] Database schema initialization completed successfully")

	return db
}
