package db

import (
	"context"
	"time"
)

func (r *Repository) CreateUser(username, passwordHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx,
		"INSERT INTO users(username, password_hash) VALUES ($1, $2)",
		username, passwordHash,
	)
	return err
}

func (r *Repository) GetUserForLogin(username string) (int64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int64
	var hash string

	err := r.pool.QueryRow(ctx,
		"SELECT id, password_hash FROM users WHERE username = $1",
		username,
	).Scan(&id, &hash)

	return id, hash, err
}

func (r *Repository) GetUserByID(userID int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var username string

	err := r.pool.QueryRow(ctx,
		"SELECT username FROM users WHERE id = $1",
		userID,
	).Scan(&username)

	return username, err
}
