package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/shmul/avito-task/internal/infrastructure/http/dto"
)

func sendError(w http.ResponseWriter, message, code string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(dto.ErrorResponse{
        Error: dto.ErrorDetails{
            Code:    code,
            Message: message,
        },
    })
}