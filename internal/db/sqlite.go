package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"li-chat/pkg/logger"
)

func NewSQLite(path string) *sql.DB {
	logger.Info("Opening SQLite database", zap.String("path", path))
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Error("Failed to open database", zap.Error(err))
		logger.Warn("Database initialization failed, application may not function correctly")
		panic(err)
	}
	logger.Debug("Database driver initialized", zap.String("path", path))
	logger.Info("Database connection established successfully")

	logger.Debug("Starting schema creation process")
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
		logger.Error("Failed to create database schema", zap.Error(err))
		logger.Warn("Schema creation error - tables may be missing or corrupted")
		panic(err)
	}
	logger.Debug("All tables created or already exist")
	logger.Info("Database schema initialization completed successfully")

	return db
}
