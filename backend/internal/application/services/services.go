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

// StandingsService
type StandingsService struct {
	matchRepo ports.MatchRepository
}

func NewStandingsService(matchRepo ports.MatchRepository) *StandingsService {
	return &StandingsService{matchRepo: matchRepo}
}

func (s *StandingsService) CalculateAll() ([]domain.GroupStanding, error) {
	matches, err := s.matchRepo.GetAllMatches()
	if err != nil {
		return nil, err
	}
	return domain.CalculateStandings(matches), nil
}

// QualifierPredictionService
type QualifierPredictionService struct {
	repo      ports.QualifierPredictionRepository
	matchRepo ports.MatchRepository
}

func NewQualifierPredictionService(repo ports.QualifierPredictionRepository, matchRepo ports.MatchRepository) *QualifierPredictionService {
	return &QualifierPredictionService{repo: repo, matchRepo: matchRepo}
}

func (s *QualifierPredictionService) Upsert(p domain.QualifierPrediction) error {
	return s.repo.UpsertQualifierPrediction(p)
}

func (s *QualifierPredictionService) GetByUserAndGroup(userID int, groupName string) (*domain.QualifierPrediction, error) {
	return s.repo.GetQualifierPredictionByUserAndGroup(userID, groupName)
}

func (s *QualifierPredictionService) ScoreUser(userID int) (int, error) {
	preds, err := s.repo.GetAllQualifierPredictionsByUser(userID)
	if err != nil {
		return 0, err
	}
	if len(preds) == 0 {
		return 0, nil
	}

	matches, err := s.matchRepo.GetAllMatches()
	if err != nil {
		return 0, err
	}

	total := 0
	groupsSeen := make(map[string]bool)
	for _, p := range preds {
		if groupsSeen[p.GroupName] {
			continue
		}
		groupsSeen[p.GroupName] = true

		groupMatches := make([]domain.Match, 0)
		for _, m := range matches {
			if m.Group == p.GroupName {
				groupMatches = append(groupMatches, m)
			}
		}
		if len(groupMatches) == 0 {
			continue
		}
		allFinished := true
		for _, m := range groupMatches {
			if m.Status != "finished" {
				allFinished = false
				break
			}
		}
		if !allFinished {
			continue
		}

		first, second, err := domain.CalculateQualifiers(groupMatches, p.GroupName)
		if err != nil {
			continue
		}
		total += domain.QualifiersPointsEarned(first, second, p.PredictedFirst, p.PredictedSecond)
	}
	return total, nil
}
