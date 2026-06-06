package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"time"
)

type QualifierPrediction struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	GroupName       string    `json:"group_name"`
	PredictedFirst  string    `json:"predicted_first"`
	PredictedSecond string    `json:"predicted_second"`
	CreatedAt       time.Time `json:"created_at"`
}

func CalculateQualifiers(matches []Match, groupName string) (string, string, error) {
	allStandings := CalculateStandings(matches)
	var groupStanding *GroupStanding
	for i := range allStandings {
		if allStandings[i].Group == groupName {
			groupStanding = &allStandings[i]
			break
		}
	}
	if groupStanding == nil || len(groupStanding.Teams) < 2 {
		return "", "", errors.New("group not found or has fewer than 2 teams")
	}

	teams := append([]TeamStanding{}, groupStanding.Teams...)
	final := applyFIFATieBreakers(teams, matches, groupName)
	return final[0].Team, final[1].Team, nil
}

func applyFIFATieBreakers(teams []TeamStanding, matches []Match, groupName string) []TeamStanding {
	result := make([]TeamStanding, 0, len(teams))
	i := 0
	for i < len(teams) {
		j := i + 1
		for j < len(teams) &&
			teams[j].Points == teams[i].Points &&
			teams[j].GoalDiff == teams[i].GoalDiff &&
			teams[j].GoalsFor == teams[i].GoalsFor {
			j++
		}
		subset := teams[i:j]
		if len(subset) > 1 {
			subset = applyH2HTieBreaker(subset, matches, groupName)
		}
		result = append(result, subset...)
		i = j
	}
	return result
}

func applyH2HTieBreaker(tied []TeamStanding, matches []Match, groupName string) []TeamStanding {
	teamNames := make(map[string]bool)
	for _, t := range tied {
		teamNames[t.Team] = true
	}

	h2h := make(map[string]*TeamStanding)
	for _, t := range tied {
		h2h[t.Team] = &TeamStanding{Team: t.Team, Flag: t.Flag}
	}

	for _, m := range matches {
		if m.Group != groupName || m.Status != "finished" {
			continue
		}
		if !teamNames[m.HomeTeam] || !teamNames[m.AwayTeam] {
			continue
		}
		home := h2h[m.HomeTeam]
		away := h2h[m.AwayTeam]
		home.Played++
		away.Played++
		home.GoalsFor += m.HomeScore
		home.GoalsAgainst += m.AwayScore
		away.GoalsFor += m.AwayScore
		away.GoalsAgainst += m.HomeScore

		if m.HomeScore > m.AwayScore {
			home.Won++
			home.Points += 3
			away.Lost++
		} else if m.HomeScore < m.AwayScore {
			away.Won++
			away.Points += 3
			home.Lost++
		} else {
			home.Drawn++
			away.Drawn++
			home.Points++
			away.Points++
		}
	}

	for _, t := range h2h {
		t.GoalDiff = t.GoalsFor - t.GoalsAgainst
	}

	result := make([]TeamStanding, 0, len(h2h))
	for _, t := range h2h {
		result = append(result, *t)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Points != result[j].Points {
			return result[i].Points > result[j].Points
		}
		if result[i].GoalDiff != result[j].GoalDiff {
			return result[i].GoalDiff > result[j].GoalDiff
		}
		if result[i].GoalsFor != result[j].GoalsFor {
			return result[i].GoalsFor > result[j].GoalsFor
		}
		return sortearTiebreak(result[i].Team, result[j].Team, groupName)
	})

	return result
}

func sortearTiebreak(a, b, groupName string) bool {
	return sortearHash(a, groupName) < sortearHash(b, groupName)
}

func sortearHash(team, groupName string) uint64 {
	h := sha256.Sum256([]byte(groupName + ":" + team))
	return binary.BigEndian.Uint64(h[:8])
}

func QualifiersPointsEarned(actualFirst, actualSecond, predictedFirst, predictedSecond string) int {
	points := 0
	if predictedFirst == actualFirst {
		points += 3
	}
	if predictedSecond == actualSecond {
		points += 3
	}
	return points
}
