package model

import "time"

type User struct {
	ID       int64
	Username string
}

// Response structs for consistency
type SuccessResponseStruct struct {
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

type ErrorResponseStruct struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenSuccessResponse struct {
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}
