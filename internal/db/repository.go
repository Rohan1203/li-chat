package db

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"li-chat/pkg/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(path string) (*Repository, error) {
	logger.Info("Initializing database repository", zap.String("path", path))
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Error("Failed to open SQLite database", zap.Error(err))
		logger.Warn("Repository initialization failed, database operations will not be available")
		return nil, err
	}
	logger.Debug("SQLite connection established")

	logger.Debug("Creating database schema")
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		logger.Error("Failed to create schema", zap.Error(err))
		logger.Warn("Schema creation error, database tables may not exist")
		return nil, err
	}
	logger.Debug("All tables created or already exist")
	logger.Info("Repository initialized successfully")

	return &Repository{db: db}, nil
}

func (r *Repository) GetOrCreateUser(username string) (int64, error) {
	logger.Info("Getting or creating user", zap.String("username", username))
	logger.Debug("Querying database for existing user", zap.String("username", username))

	var id int64
	err := r.db.QueryRow(
		"SELECT id FROM users WHERE username = ?",
		username,
	).Scan(&id)

	if err == sql.ErrNoRows {
		logger.Debug("User not found in database", zap.String("username", username))
		logger.Debug("Creating new user record", zap.String("username", username))

		res, err := r.db.Exec(
			"INSERT INTO users(username) VALUES(?)",
			username,
		)
		if err != nil {
			logger.Error("Failed to create user record", zap.String("username", username), zap.Error(err))
			logger.Warn("User creation failed - possible duplicate username or database error")
			return 0, err
		}

		userID, err := res.LastInsertId()
		if err != nil {
			logger.Error("Failed to get last insert ID for user", zap.String("username", username), zap.Error(err))
			logger.Warn("Could not retrieve user ID after insertion")
			return 0, err
		}

		logger.Debug("New user inserted", zap.String("username", username), zap.Int64("user_id", userID))
		logger.Info("New user created successfully", zap.String("username", username), zap.Int64("user_id", userID))
		return userID, nil
	}

	if err != nil {
		logger.Error("Failed to query user from database", zap.String("username", username), zap.Error(err))
		logger.Warn("Database query error occurred while retrieving user")
		return 0, err
	}

	logger.Debug("Existing user found", zap.String("username", username), zap.Int64("user_id", id))
	logger.Info("User already exists in database", zap.String("username", username), zap.Int64("user_id", id))
	return id, err
}

func (r *Repository) SaveMessage(userID int64, content string) error {
	logger.Info("Saving new message", zap.Int64("user_id", userID))
	logger.Debug("Message details", zap.Int64("user_id", userID), zap.Int("content_length", len(content)))

	if content == "" {
		logger.Warn("Empty message content provided", zap.Int64("user_id", userID))
	}

	logger.Debug("Executing INSERT query for message")
	_, err := r.db.Exec(
		"INSERT INTO messages(user_id, content) VALUES(?, ?)",
		userID,
		content,
	)
	if err != nil {
		logger.Error("Failed to save message", zap.Int64("user_id", userID), zap.Error(err))
		logger.Warn("Message insertion failed - database may be unavailable or corrupted")
		return err
	}

	logger.Debug("Message record inserted successfully into database")
	logger.Info("Message saved successfully", zap.Int64("user_id", userID), zap.Int("content_size", len(content)))
	return nil
}

func generateSessionID() string {
	// Simple session ID generation (in production, use crypto/rand with UUID)
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
