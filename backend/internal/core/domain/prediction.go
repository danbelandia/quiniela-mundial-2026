package domain

type Prediction struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	MatchID   int `json:"match_id"`
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}