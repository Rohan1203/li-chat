package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"li-chat/internal/auth"
	"li-chat/internal/db"
	"li-chat/internal/model"
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
		auth.SendJSONResponse(w, http.StatusBadRequest, auth.ErrorResponse("method not allowed"))
		return
	}

	var c credentials
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusBadRequest, auth.ErrorResponse("invalid request body"))
		return
	}

	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("server error"))
		return
	}

	err = h.repo.CreateUser(c.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			auth.SendJSONResponse(w, http.StatusConflict, auth.ErrorResponse("user with this username already exists"))
			return
		}
		auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("failed to create user"))
		return
	}

	response := auth.SuccessResponse(model.SuccessResponseStruct{
		Username: c.Username,
		Password: "*****",
		Message:  "user created",
	})

	auth.SendJSONResponse(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		auth.SendJSONResponse(w, http.StatusBadRequest, auth.ErrorResponse("method not allowed"))
		return
	}

	var c credentials
	// json.NewDecoder(r.Body).Decode(&c)
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusBadRequest, auth.ErrorResponse("invalid request body"))
		return
	}

	userID, hash, err := h.repo.GetUserForLogin(c.Username)
	if err != nil || auth.CheckPassword(hash, c.Password) != nil {
		auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("invalid credentials"))
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(userID, c.Username)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("server error"))
		return
	}

	// jwt refresh token
	refreshToken, err := auth.GenerateRefreshToken(userID)
    if err != nil {
        auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("Failed to generate refresh token"))
        return
    }

	response := auth.SuccessResponse(model.SuccessResponseStruct{
		Username: c.Username,
		Message:  "login success",
		Token:    token,
		RefreshToken: refreshToken,
	})

	auth.SendJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		auth.SendJSONResponse(w, http.StatusMethodNotAllowed, auth.ErrorResponse("method not allowed"))
		return
	}

	// JWT is stateless, logout is client-side (remove token from localStorage)
	auth.SendJSONResponse(w, http.StatusOK, auth.SuccessResponse(model.SuccessResponseStruct{Message: "logout success"}))
}

func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("unauthorized; token missing?"))
		return
	}

	// Bearer <token>
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("unauthorized; not valid format?"))
		return
	}

	claims, err := auth.ValidateToken(parts[1])
	if err != nil {
		auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("unauthorized; invalid token"))
		return
	}

	response := auth.SuccessResponse(model.SuccessResponseStruct{
		Username: claims.Username,
		Message:  "here you are!",
	})

	auth.SendJSONResponse(w, http.StatusOK, response)

}

// Refresh token interface
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		auth.SendJSONResponse(w, http.StatusMethodNotAllowed, auth.ErrorResponse("method not allowed"))
		return
	}

	var req model.RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusBadRequest, auth.ErrorResponse("invalid request body"))
		return
	}

	// 1. Validate the refresh token
	refreshClaims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		// If refresh token is invalid or expired, force re-login
		auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("invalid or expired refresh token"))
		return
	}

	// 2. (Optional but recommended) Check if the refresh token is still valid in your database
	//    This allows for refresh token revocation.
	// isValid, err := h.repo.IsRefreshTokenValid(refreshClaims.UserID, req.RefreshToken)
	// if err != nil || !isValid {
	//     auth.SendJSONResponse(w, http.StatusUnauthorized, auth.ErrorResponse("Invalid or revoked refresh token. Please log in again."))
	//     return
	// }

	// 3. Get user details needed for the new access token (e.g., username)
	//    You might need to fetch this from your database using refreshClaims.UserID
	username, err := h.repo.GetUserByID(refreshClaims.UserID) // Placeholder
	if err != nil {
		auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("user not found for refresh token."))
		return
	}

	// 4. Generate a new access token
	newAccessToken, err := auth.GenerateToken(refreshClaims.UserID, username)
	if err != nil {
		auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("failed to generate new access token."))
		return
	}

	// 5. Optionally, generate a new refresh token and invalidate the old one
	//    This is called "refresh token rotation" and improves security.
	// newRefreshToken, err := auth.GenerateRefreshToken(refreshClaims.UserID)
	// if err != nil {
	//     auth.SendJSONResponse(w, http.StatusInternalServerError, auth.ErrorResponse("Failed to generate new refresh token."))
	//     return
	// }
	// h.repo.InvalidateRefreshToken(req.RefreshToken) // Invalidate old
	// h.repo.SaveRefreshToken(refreshClaims.UserID, newRefreshToken) // Save new

	auth.SendJSONResponse(w, http.StatusOK, model.RefreshTokenSuccessResponse{
		AccessToken: newAccessToken,
		// If you implemented rotation: RefreshToken: newRefreshToken,
		Message:     "new access token generated",
	})
}

