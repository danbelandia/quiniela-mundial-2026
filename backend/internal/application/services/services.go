package services

import (
	"errors"
	"strings"
	"quiniela-backend/internal/core/domain"
	"quiniela-backend/internal/core/ports"
)

// MatchService
type MatchService struct {
	repo ports.MatchRepository
}

func NewMatchService(repo ports.MatchRepository) *MatchService {
	return &MatchService{repo: repo}
}

func (s *MatchService) ListMatches() ([]domain.Match, error) {
	return s.repo.GetAllMatches()
}

func (s *MatchService) UpdateResult(id int, home, away int) error {
	return s.repo.UpdateMatchResult(id, home, away)
}

func (s *MatchService) ToggleLock(id int, locked bool) error {
	return s.repo.ToggleLock(id, locked)
}

func (s *MatchService) LockAll(locked bool) error {
	return s.repo.LockAll(locked)
}

// PredictionService
type PredictionService struct {
	repo ports.PredictionRepository
}

func NewPredictionService(repo ports.PredictionRepository) *PredictionService {
	return &PredictionService{repo: repo}
}

func (s *PredictionService) PlacePrediction(p domain.Prediction) error {
	return s.repo.Save(p)
}

func (s *PredictionService) GetAllPredictions() ([]domain.Prediction, error) {
	return s.repo.GetAllPredictions()
}

// RankingService
type RankingService struct {
	predRepo ports.PredictionRepository
}

func NewRankingService(predRepo ports.PredictionRepository) *RankingService {
	return &RankingService{predRepo: predRepo}
}

func (s *RankingService) CalculatePoints(pred domain.Prediction, actualHome, actualAway int) int {
	return pred.PointsEarned(actualHome, actualAway)
}

// UserService
type UserService struct {
	repo     ports.UserRepository
	predRepo ports.PredictionRepository
}

func NewUserService(repo ports.UserRepository, predRepo ports.PredictionRepository) *UserService {
	return &UserService{repo: repo, predRepo: predRepo}
}

func (s *UserService) Login(identifier, password string) (*domain.User, error) {
	if identifier == "" {
		return nil, errors.New("identifier is required")
	}
	var u *domain.User
	var err error
	if strings.Contains(identifier, "@") {
		u, err = s.repo.GetByEmail(identifier)
	} else {
		u, err = s.repo.GetByUsername(identifier)
	}
	if err != nil {
		return nil, err
	}
	if u.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

func (s *UserService) Register(u domain.User) error {
	if u.Username == "" || u.Email == "" || u.Password == "" {
		return errors.New("missing required fields")
	}
	return s.repo.Create(u)
}

func (s *UserService) GetByID(id int) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) GetUserPredictions(userID int) ([]ports.PredictionWithMatch, error) {
	return s.predRepo.GetByUserIDWithMatch(userID)
}

func (s *UserService) DeleteUser(id int) error {
	return s.repo.Delete(id)
}

func (s *UserService) UpdateUser(u domain.User) error {
	return s.repo.Update(u)
}

func (s *UserService) GetAllUsers() ([]domain.User, error) {
	return s.repo.GetAll()
}
