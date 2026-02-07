package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"li-chat/internal/auth"
	"li-chat/internal/db"
)

type AuthHandler struct {
	repo *db.Repository
}

func NewAuthHandler(repo *db.Repository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var c credentials
	json.NewDecoder(r.Body).Decode(&c)

	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	err = h.repo.CreateUser(c.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "user exists", http.StatusConflict)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var c credentials
	json.NewDecoder(r.Body).Decode(&c)

	userID, hash, err := h.repo.GetUserForLogin(c.Username)
	if err != nil || auth.CheckPassword(hash, c.Password) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(userID, c.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JWT is stateless, logout is client-side (remove token from localStorage)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Bearer <token>
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(parts[1])
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"username": claims.Username})
}
