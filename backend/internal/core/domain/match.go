package domain

import "time"

type Match struct {
	ID        int       `json:"id"`
	HomeTeam  string    `json:"home_team"`
	AwayTeam  string    `json:"away_team"`
	HomeScore int       `json:"home_score"`
	AwayScore int       `json:"away_score"`
	Date      time.Time `json:"date"`
	Status    string    `json:"status"`
	Group     string    `json:"group"`
	HomeFlag  string    `json:"home_flag"`
	AwayFlag  string    `json:"away_flag"`
	IsLocked  bool      `json:"is_locked"`
}

func (m Match) IsEffectivelyLocked(now time.Time, window time.Duration) bool {
	if m.IsLocked {
		return true
	}
	return !now.Add(window).Before(m.Date)
}