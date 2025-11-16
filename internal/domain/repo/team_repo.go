package repo

import "github.com/shmul/avito-task/internal/domain/entity"

type TeamRepository interface {
    Create(team *entity.Team) error
    GetByName(teamName string) (*entity.Team, error)
    Exists(teamName string) (bool, error)
}