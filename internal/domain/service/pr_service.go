package service

import (
	"fmt"
	"math/rand"
	"time"
	"github.com/shmul/avito-task/internal/domain/entity"
	"github.com/shmul/avito-task/internal/domain/repo"
)

type PRService struct {
	prRepo   repo.PRRepository
	userRepo repo.UserRepository
	teamRepo repo.TeamRepository
	config   *PRServiceConfig
	rng      *rand.Rand
}

//для тестов
// func (s *PRService) GetDB() *sql.DB {
// 	if repo, ok := s.teamRepo.(interface{ GetDB() *sql.DB }); ok {
// 		return repo.GetDB()
// 	}
// 	return nil
// }
//

type PRServiceConfig struct {
	ReviewerCount int
	RandomSeed    int64
}

type ReassignResult struct {
	PR         *entity.PullRequest
	ReplacedBy string
}

func NewPRService(prRepo repo.PRRepository, userRepo repo.UserRepository, teamRepo repo.TeamRepository, config *PRServiceConfig) *PRService {
	var seed int64
	if config.RandomSeed == 0 {
		seed = time.Now().UnixNano()
	} else {
		seed = config.RandomSeed
	}

	src := rand.NewSource(seed)
	rng := rand.New(src)

	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		config:   config,
		rng:      rng,
	}
}

func (s *PRService) CreatePR(prID, prName, authorID string) (*entity.PullRequest, error) {
	exists, err := s.prRepo.Exists(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pr existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("pr already exists: %s", prID)
	}

	author, err := s.userRepo.GetByID(authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %s", authorID)
	}

	teamUsers, err := s.userRepo.GetActiveUsersByTeam(author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team users: %w", err)
	}

	var candidates []*entity.User
	for _, user := range teamUsers {
		if user.UserID != authorID {
			candidates = append(candidates, user)
		}
	}

	reviewers := s.selectRandomReviewers(candidates, s.config.ReviewerCount)

	pr := &entity.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.Create(pr); err != nil {
		return nil, fmt.Errorf("failed to create pr: %w", err)
	}

	return pr, nil
}

func (s *PRService) MergePR(prID string) (*entity.PullRequest, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, fmt.Errorf("pr not found: %s", prID)
	}

	if pr.Status == entity.StatusMerged {
		return pr, nil
	}

	now := time.Now()
	pr.Status = entity.StatusMerged
	pr.MergedAt = &now

	if err := s.prRepo.Update(pr); err != nil {
		return nil, fmt.Errorf("failed to merge pr: %w", err)
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(prID, oldReviewerID string) (*ReassignResult, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, fmt.Errorf("pr not found: %s", prID)
	}

	if pr.Status == entity.StatusMerged {
		return nil, fmt.Errorf("cannot reassign on merged pr")
	}

	if !s.contains(pr.AssignedReviewers, oldReviewerID) {
		return nil, fmt.Errorf("reviewer is not assigned to this pr")
	}

	old, err := s.userRepo.GetByID(oldReviewerID)
	if err != nil {
		return nil, fmt.Errorf("reviewer not found: %s", oldReviewerID)
	}

	teamUsers, err := s.userRepo.GetActiveUsersByTeam(old.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team users: %w", err)
	}

	var candidates []*entity.User
	for _, user := range teamUsers {
		if user.UserID != pr.AuthorID &&
			!s.contains(pr.AssignedReviewers, user.UserID) &&
			user.UserID != oldReviewerID {
			candidates = append(candidates, user)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no active replacement candidate in team")
	}

	new := candidates[s.rng.Intn(len(candidates))]

	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			pr.AssignedReviewers[i] = new.UserID
			break
		}
	}

	if err := s.prRepo.Update(pr); err != nil {
		return nil, fmt.Errorf("failed to update pr: %w", err)
	}

	result := &ReassignResult{
		PR:         pr,
		ReplacedBy: new.UserID,
	}

	return result, nil
}

func (s *PRService) GetPRsByReviewer(userID string) ([]*entity.PullRequest, error) {
	prs, err := s.prRepo.GetByReviewer(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	return prs, nil
}

func (s *PRService) selectRandomReviewers(candidates []*entity.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	shuffled := make([]*entity.User, len(candidates))
	copy(shuffled, candidates)
	s.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	count := min(len(shuffled), maxCount)
	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}

func (s *PRService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
