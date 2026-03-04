package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (r *Repository) CreateUser(username, passwordHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"username":      username,
		"password_hash": passwordHash,
	}

	resp, err := r.request(ctx, "POST", "/users", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleError(resp)
}

func (r *Repository) GetUserForLogin(username string) (int64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf(
		"/users?username=eq.%s&select=id,password_hash",
		username,
	)

	resp, err := r.request(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return 0, "", err
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return 0, "", err
	}

	if len(users) == 0 {
		return 0, "", fmt.Errorf("user not found")
	}

	id := int64(users[0]["id"].(float64))
	hash := users[0]["password_hash"].(string)

	return id, hash, nil
}

func (r *Repository) GetUserByID(userID int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf(
		"/users?id=eq.%d&select=username",
		userID,
	)

	resp, err := r.request(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return "", err
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", fmt.Errorf("user not found")
	}

	return users[0]["username"].(string), nil
}