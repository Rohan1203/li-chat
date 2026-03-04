package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"li-chat/pkg/logger"
)

type Repository struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewRepository() (*Repository, error) {
	logger.Info("Initializing Supabase repository")

	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY must be set")
	}

	return &Repository{
		baseURL: url + "/rest/v1",
		apiKey:  key,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (r *Repository) request(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		r.baseURL+endpoint,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", r.apiKey)
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	return r.client.Do(req)
}

func handleError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("supabase error: %s", string(body))
}

func (r *Repository) GetOrCreateUser(username string) (int64, error) {
	logger.Info("Getting or creating user", zap.String("username", username))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1️⃣ Try fetch existing user
	endpoint := fmt.Sprintf("/users?username=eq.%s&select=id", username)

	resp, err := r.request(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return 0, err
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return 0, err
	}

	if len(users) > 0 {
		id := int64(users[0]["id"].(float64))
		logger.Info("User already exists", zap.Int64("id", id))
		return id, nil
	}

	// 2️⃣ Create new user
	body := map[string]string{
		"username": username,
	}

	resp, err = r.request(ctx, "POST", "/users", body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return 0, err
	}

	var created []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return 0, err
	}

	id := int64(created[0]["id"].(float64))
	logger.Info("New user created", zap.Int64("id", id))

	return id, nil
}

func (r *Repository) SaveMessage(userID int64, content string) error {
	logger.Info("Saving message", zap.Int64("user_id", userID))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"user_id": userID,
		"content": content,
	}

	resp, err := r.request(ctx, "POST", "/messages", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handleError(resp)
}

func (r *Repository) GetMessages(limit int) ([]interface{}, error) {
	logger.Info("Fetching messages", zap.Int("limit", limit))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf(
		"/messages?select=content,created_at,users(username)&order=created_at.asc&limit=%d",
		limit,
	)

	resp, err := r.request(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return nil, err
	}

	var result []interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	return result, err
}