package service

import (
    "fmt"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/repo"
)

type UserService struct {
    userRepo repo.UserRepository
    teamRepo repo.TeamRepository
}

func NewUserService(userRepo repo.UserRepository, teamRepo repo.TeamRepository) *UserService {
    return &UserService{
        userRepo: userRepo,
        teamRepo: teamRepo,
    }
}

func (s *UserService) SetUserActive(userID string, isActive bool) (*entity.User, error) {
    exists, err := s.userRepo.Exists(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to check user existence: %w", err)
    }
    if !exists {
        return nil, fmt.Errorf("user not found: %s", userID)
    }

    user, err := s.userRepo.SetActive(userID, isActive)
    if err != nil {
        return nil, fmt.Errorf("failed to set user active: %w", err)
    }

    return user, nil
}