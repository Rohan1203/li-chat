package auth

import (
	"encoding/json"
	"li-chat/internal/model"
	"net/http"
	"time"
)

// JSON response
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(data)
}

// return interface
func ErrorResponse(message string) *model.ErrorResponseStruct {
	return &model.ErrorResponseStruct{
		Error:     message,
		Timestamp: time.Now(),
	}
}

func SuccessResponse(data model.SuccessResponseStruct) *model.SuccessResponseStruct {
	response := model.SuccessResponseStruct{
		Message:   data.Message,
		Timestamp: time.Now(),
	}
	if data.Token != "" {
		response.Token = data.Token
	}

	if data.RefreshToken != "" {
		response.RefreshToken = data.RefreshToken
	}
	if data.Username != "" {
		response.Username = data.Username
	}
	if data.Password != "" {
		response.Password = data.Password
	}

	return &response
}
