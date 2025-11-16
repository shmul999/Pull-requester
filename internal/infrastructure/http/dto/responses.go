package dto

import "github.com/shmul/avito-task/internal/domain/entity"

type ErrorResponse struct {
    Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type TeamResponse struct {
    Team *entity.Team `json:"team"`
}

type UserResponse struct {
    User *entity.User `json:"user"`
}

type PRResponse struct {
    PR *entity.PullRequest `json:"pr"`
}

type ReassignResponse struct {
    PR         *entity.PullRequest `json:"pr"`
    ReplacedBy string              `json:"replaced_by"`
}

type UserPRsResponse struct {
    UserID        string                 `json:"user_id"`
    PullRequests  []*entity.PullRequest `json:"pull_requests"`
}