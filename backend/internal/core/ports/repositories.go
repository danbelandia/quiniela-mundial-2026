package ports

import "quiniela-backend/internal/core/domain"

type UserRepository interface {
	Create(user domain.User) error
	GetByID(id int) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetAll() ([]domain.User, error)
	Delete(id int) error
	Update(u domain.User) error
}


type MatchRepository interface {
	GetAllMatches() ([]domain.Match, error)
	GetMatchByID(id int) (*domain.Match, error)
	UpdateMatchResult(id int, homeScore, awayScore int) error
	ToggleLock(id int, locked bool) error
	LockAll(locked bool) error
}

type PredictionRepository interface {
	Save(prediction domain.Prediction) error
	GetByMatchID(matchID int) ([]domain.Prediction, error)
	GetAllPredictions() ([]domain.Prediction, error)
}