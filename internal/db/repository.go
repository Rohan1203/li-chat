package db

import (
	"database/sql"
	"fmt"
	"time"

	"li-chat/pkg/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(path string) (*Repository, error) {
	logger.Info("[REPO::INIT] Initializing database repository with sqlite3 at path: %s", path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Error("[REPO::INIT] Failed to open SQLite database: %v", err)
		logger.Warn("[REPO::INIT] Repository initialization failed, database operations will not be available")
		return nil, err
	}
	logger.Debug("[REPO::INIT] SQLite connection established")

	logger.Debug("[REPO::SCHEMA] Creating database schema...")
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
		logger.Error("[REPO::SCHEMA] Failed to create schema: %v", err)
		logger.Warn("[REPO::SCHEMA] Schema creation error, database tables may not exist")
		return nil, err
	}
	logger.Debug("[REPO::SCHEMA] All tables created or already exist")
	logger.Info("[REPO::INIT] Repository initialized successfully")

	return &Repository{db: db}, nil
}

func (r *Repository) GetOrCreateUser(username string) (int64, error) {
	logger.Info("[DB::USER] Getting or creating user: %s", username)
	logger.Debug("[DB::USER] Querying database for existing user: %s", username)

	var id int64
	err := r.db.QueryRow(
		"SELECT id FROM users WHERE username = ?",
		username,
	).Scan(&id)

	if err == sql.ErrNoRows {
		logger.Debug("[DB::USER] User %s not found in database", username)
		logger.Debug("[DB::USER] Creating new user record for: %s", username)

		res, err := r.db.Exec(
			"INSERT INTO users(username) VALUES(?)",
			username,
		)
		if err != nil {
			logger.Error("[DB::USER] Failed to create user record for %s: %v", username, err)
			logger.Warn("[DB::USER] User creation failed - possible duplicate username or database error")
			return 0, err
		}

		userID, err := res.LastInsertId()
		if err != nil {
			logger.Error("[DB::USER] Failed to get last insert ID for user %s: %v", username, err)
			logger.Warn("[DB::USER] Could not retrieve user ID after insertion")
			return 0, err
		}

		logger.Debug("[DB::USER] New user inserted with ID: %d", userID)
		logger.Info("[DB::USER] New user created successfully: username=%s, userId=%d", username, userID)
		return userID, nil
	}

	if err != nil {
		logger.Error("[DB::USER] Failed to query user %s from database: %v", username, err)
		logger.Warn("[DB::USER] Database query error occurred while retrieving user")
		return 0, err
	}

	logger.Debug("[DB::USER] Existing user found: username=%s, userId=%d", username, id)
	logger.Info("[DB::USER] User already exists in database: username=%s, userId=%d", username, id)
	return id, err
}

func (r *Repository) SaveMessage(userID int64, content string) error {
	logger.Info("[DB::MSG] Saving new message for user ID: %d", userID)
	logger.Debug("[DB::MSG] Message details - UserID: %d, Content length: %d bytes", userID, len(content))

	if content == "" {
		logger.Warn("[DB::MSG] Empty message content provided for user ID %d", userID)
	}

	logger.Debug("[DB::MSG] Executing INSERT query for message...")
	_, err := r.db.Exec(
		"INSERT INTO messages(user_id, content) VALUES(?, ?)",
		userID,
		content,
	)
	if err != nil {
		logger.Error("[DB::MSG] Failed to save message for user ID %d: %v", userID, err)
		logger.Warn("[DB::MSG] Message insertion failed - database may be unavailable or corrupted")
		return err
	}

	logger.Debug("[DB::MSG] Message record inserted successfully into database")
	logger.Info("[DB::MSG] Message saved successfully for user ID %d (size: %d bytes)", userID, len(content))
	return nil
}

func generateSessionID() string {
	// Simple session ID generation (in production, use crypto/rand with UUID)
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
