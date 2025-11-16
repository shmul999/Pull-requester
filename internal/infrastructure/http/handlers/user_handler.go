package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/shmul/avito-task/internal/domain/service"
    "github.com/shmul/avito-task/internal/infrastructure/http/dto"
)

type UserHandler struct {
    userService *service.UserService
    prService   *service.PRService
}

func NewUserHandler(userService *service.UserService, prService *service.PRService) *UserHandler {
    return &UserHandler{
        userService: userService,
        prService:   prService,
    }
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.SetUserActiveRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    user, err := h.userService.SetUserActive(req.UserID, req.IsActive)
    if err != nil {
        if err.Error() == fmt.Sprintf("user not found: %s", req.UserID) {
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
            return
        }
        sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.UserResponse{User: user})
}

func (h *UserHandler) GetUserReview(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        sendError(w, "user_id is required", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    prs, err := h.prService.GetPRsByReviewer(userID)
    if err != nil {
        sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        return
    }

    response := dto.UserPRsResponse{
        UserID:       userID,
        PullRequests: prs,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}