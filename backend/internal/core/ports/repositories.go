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
	GetByUserIDWithMatch(userID int) ([]PredictionWithMatch, error)
}

type PredictionWithMatch struct {
	GroupName  string `json:"group_name"`
	HomeTeam   string `json:"home_team"`
	HomeFlag   string `json:"home_flag"`
	AwayTeam   string `json:"away_team"`
	AwayFlag   string `json:"away_flag"`
	MatchDate  string `json:"match_date"`
	Status     string `json:"status"`
	RealHome   int    `json:"home_score_real"`
	RealAway   int    `json:"away_score_real"`
	Points     int    `json:"points"`
	HasPrediction bool   `json:"has_prediction"`
	PredictionID   int    `json:"prediction_id,omitempty"`
	PredHome       int    `json:"home_score_pred"`
	PredAway       int    `json:"away_score_pred"`
}

type QualifierPredictionRepository interface {
	UpsertQualifierPrediction(prediction domain.QualifierPrediction) error
	GetQualifierPredictionByUserAndGroup(userID int, groupName string) (*domain.QualifierPrediction, error)
	GetAllQualifierPredictionsByUser(userID int) ([]domain.QualifierPrediction, error)
	GetAllQualifierPredictions() ([]domain.QualifierPrediction, error)
}

type QualifierPredictionView struct {
	GroupName          string `json:"group_name"`
	PredictedFirst     string `json:"predicted_first"`
	PredictedFirstFlag string `json:"predicted_first_flag"`
	PredictedSecond    string `json:"predicted_second"`
	PredictedSecondFlag string `json:"predicted_second_flag"`
	ActualFirst        string `json:"actual_first"`
	ActualFirstFlag    string `json:"actual_first_flag"`
	ActualSecond       string `json:"actual_second"`
	ActualSecondFlag   string `json:"actual_second_flag"`
	PointsEarned       int    `json:"points_earned"`
	GroupClosed        bool   `json:"group_closed"`
}

type TopScorerRepository interface {
	UpsertTopScorerPrediction(userID int, player string) error
	GetTopScorerPredictionByUserID(userID int) (*domain.TopScorerPrediction, error)
	GetAllTopScorerPredictions() ([]domain.TopScorerPrediction, error)
}

type AppConfigRepository interface {
	GetAppConfig(key string) (string, error)
	SetAppConfig(key, value string) error
}