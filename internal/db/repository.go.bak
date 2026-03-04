package db

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"li-chat/pkg/logger"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(connStr string) (*Repository, error) {
	logger.Info("Initializing database repository")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse config to support both IPv4 and IPv6 with smart fallback
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Error("Failed to parse connection string", zap.Error(err))
		return nil, err
	}

	// Use custom dialer that tries IPv4 first, then IPv6 (better for cloud platforms)
	config.ConnConfig.DialFunc = func(ctx context.Context, network string, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		// Try tcp4 first (IPv4 only) - better compatibility with cloud platforms
		conn, err := dialer.DialContext(ctx, "tcp4", addr)
		if err == nil {
			return conn, nil
		}
		// Fall back to tcp (both IPv4 and IPv6)
		return dialer.DialContext(ctx, "tcp", addr)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("Failed to create connection pool", zap.Error(err))
		logger.Warn("Repository initialization failed, database operations will not be available")
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		pool.Close()
		return nil, err
	}
	logger.Debug("PostgreSQL connection pool established")

	logger.Debug("Creating database schema")
	_, err = pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		user_id INTEGER,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		logger.Error("Failed to create schema", zap.Error(err))
		logger.Warn("Schema creation error, database tables may not exist")
		pool.Close()
		return nil, err
	}
	logger.Debug("All tables created or already exist")
	logger.Info("Repository initialized successfully")

	return &Repository{pool: pool}, nil
}

func (r *Repository) GetOrCreateUser(username string) (int64, error) {
	logger.Info("Getting or creating user", zap.String("username", username))
	logger.Debug("Querying database for existing user", zap.String("username", username))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int64
	err := r.pool.QueryRow(ctx,
		"SELECT id FROM users WHERE username = $1",
		username,
	).Scan(&id)

	if err == pgx.ErrNoRows {
		logger.Debug("User not found in database", zap.String("username", username))
		logger.Debug("Creating new user record", zap.String("username", username))

		var userID int64
		err := r.pool.QueryRow(ctx,
			"INSERT INTO users(username) VALUES($1) RETURNING id",
			username,
		).Scan(&userID)
		if err != nil {
			logger.Error("Failed to create user record", zap.String("username", username), zap.Error(err))
			logger.Warn("User creation failed - possible duplicate username or database error")
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Debug("Executing INSERT query for message")
	_, err := r.pool.Exec(ctx,
		"INSERT INTO messages(user_id, content) VALUES($1, $2)",
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

func (r *Repository) GetMessages(limit int) ([]interface{}, error) {
	logger.Info("[DB::MSG] Fetching message history with limit: %d", zap.Int("limit", limit))
	logger.Debug("[DB::MSG] Executing SELECT query for recent messages...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		ORDER BY m.created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		logger.Error("[DB::MSG] Failed to fetch messages: %v", zap.Error(err))
		logger.Warn("[DB::MSG] Message retrieval failed - database query error")
		return nil, err
	}
	defer rows.Close()

	var messages []interface{}
	for rows.Next() {
		var username, content, createdAt string

		err := rows.Scan(&username, &content, &createdAt)
		if err != nil {
			logger.Error("[DB::MSG] Failed to scan message row: %v", zap.Error(err))
			continue
		}

		messages = append(messages, map[string]string{
			"username":   username,
			"content":    content,
			"created_at": createdAt,
		})
	}

	if err = rows.Err(); err != nil {
		logger.Error("[DB::MSG] Error iterating message rows: %v", zap.Error(err))
		return nil, err
	}

	logger.Debug("[DB::MSG] Retrieved %d messages from database", zap.Int("messages", len(messages)))
	logger.Info("[DB::MSG] Message history loaded successfully (count: %d)",  zap.Int("messages", len(messages)))
	return messages, nil
}

func generateSessionID() string {
	// Simple session ID generation (in production, use crypto/rand with UUID)
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
