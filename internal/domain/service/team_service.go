package service

import (
    "fmt"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/repo"
)

type TeamService struct {
    teamRepo repo.TeamRepository
    userRepo repo.UserRepository
}

func NewTeamService(teamRepo repo.TeamRepository, userRepo repo.UserRepository) *TeamService {
    return &TeamService{
        teamRepo: teamRepo,
        userRepo: userRepo,
    }
}

func (s *TeamService) CreateTeam(team *entity.Team) error {
    exists, err := s.teamRepo.Exists(team.TeamName)
    if err != nil {
        return fmt.Errorf("failed to check team existence: %w", err)
    }
    if exists {
        return fmt.Errorf("team already exists: %s", team.TeamName)
    }

    if err := s.teamRepo.Create(team); err != nil {
        return fmt.Errorf("failed to create team: %w", err)
    }

    return nil
}

func (s *TeamService) GetTeam(teamName string) (*entity.Team, error) {
    team, err := s.teamRepo.GetByName(teamName)
    if err != nil {
        return nil, fmt.Errorf("failed to get team: %w", err)
    }
    return team, nil
}