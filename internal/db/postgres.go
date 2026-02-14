package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"li-chat/pkg/logger"
)

func NewPostgres(connStr string) *sql.DB {
	logger.Info("Opening PostgreSQL database")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database", zap.Error(err))
		logger.Warn("Database initialization failed, application may not function correctly")
		panic(err)
	}

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		panic(err)
	}
	logger.Info("Database connection established successfully")

	logger.Debug("Starting schema creation process")
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		user_id INTEGER,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
