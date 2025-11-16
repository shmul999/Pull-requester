package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/shmul/avito-task/internal/domain/service"
    "github.com/shmul/avito-task/internal/infrastructure/http/dto"
)

type PRHandler struct {
    prService *service.PRService
}

func NewPRHandler(prService *service.PRService) *PRHandler {
    return &PRHandler{
        prService: prService,
    }
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.CreatePRRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    pr, err := h.prService.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
    if err != nil {
        switch err.Error() {
        case fmt.Sprintf("PR already exists: %s", req.PullRequestID):
            sendError(w, "PR id already exists", "PR_EXISTS", http.StatusConflict)
        case fmt.Sprintf("author not found: %s", req.AuthorID):
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
        default:
            sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(dto.PRResponse{PR: pr})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.MergePRRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    pr, err := h.prService.MergePR(req.PullRequestID)
    if err != nil {
        if err.Error() == fmt.Sprintf("PR not found: %s", req.PullRequestID) {
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
            return
        }
        sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.PRResponse{PR: pr})
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.ReassignReviewerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    result, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
    if err != nil {
        switch err.Error() {
        case "cannot reassign on merged PR":
            sendError(w, "cannot reassign on merged PR", "PR_MERGED", http.StatusConflict)
        case "reviewer is not assigned to this PR":
            sendError(w, "reviewer is not assigned to this PR", "NOT_ASSIGNED", http.StatusConflict)
        case "no active replacement candidate in team":
            sendError(w, "no active replacement candidate in team", "NO_CANDIDATE", http.StatusConflict)
        case fmt.Sprintf("PR not found: %s", req.PullRequestID):
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
        case fmt.Sprintf("reviewer not found: %s", req.OldUserID):
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
        default:
            sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.ReassignResponse{
        PR:         result.PR,
        ReplacedBy: result.ReplacedBy,
    })
}