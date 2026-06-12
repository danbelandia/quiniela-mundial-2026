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
	sign := func(a, b int) int {
		switch {
		case a > b:
			return 1
		case a < b:
			return -1
		default:
			return 0
		}
	}
	if sign(p.HomeScore, p.AwayScore) == sign(actualHome, actualAway) {
		return 2
	}
	return 0
}