package domain

import (
	"strings"
	"time"
	"unicode"
)

type TopScorerCandidate struct {
	Name string `json:"name"`
	Team string `json:"team"`
	Flag string `json:"flag"`
}

type TopScorerPrediction struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	PredictedPlayer string    `json:"predicted_player"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func TopScorerPointsEarned(predicted, actual string) int {
	if actual == "" {
		return 0
	}
	if NormalizeName(predicted) == NormalizeName(actual) {
		return 6
	}
	return 0
}

func NormalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
