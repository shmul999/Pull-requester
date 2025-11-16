package repo

import "github.com/shmul/avito-task/internal/domain/entity"

type PRRepository interface {
    Create(pr *entity.PullRequest) error
    GetByID(prID string) (*entity.PullRequest, error)
    Update(pr *entity.PullRequest) error
    GetByReviewer(userID string) ([]*entity.PullRequest, error)
    Exists(prID string) (bool, error)
}