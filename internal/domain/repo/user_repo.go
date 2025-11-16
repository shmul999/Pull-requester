package repo

import "github.com/shmul/avito-task/internal/domain/entity"

type UserRepository interface {
    CreateOrUpdate(user *entity.User) error
    GetByID(userID string) (*entity.User, error)
    SetActive(userID string, isActive bool) (*entity.User, error)
    GetActiveUsersByTeam(teamName string) ([]*entity.User, error)
    GetByTeam(teamName string) ([]*entity.User, error)
    Exists(userID string) (bool, error)
}