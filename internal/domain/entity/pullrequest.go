package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}
