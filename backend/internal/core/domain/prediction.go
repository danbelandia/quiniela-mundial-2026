package domain

type Prediction struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	MatchID   int `json:"match_id"`
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

func (p *Prediction) PointsEarned(actualHome, actualAway int) int {
	if p.HomeScore == actualHome && p.AwayScore == actualAway {
		return 3
	}
	if p.HomeScore == p.AwayScore && actualHome == actualAway {
		return 2
	}
	predHomeWin := p.HomeScore > p.AwayScore
	actualHomeWin := actualHome > actualAway
	if predHomeWin == actualHomeWin {
		return 2
	}
	return 0
}