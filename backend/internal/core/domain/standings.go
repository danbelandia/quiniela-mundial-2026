package domain

import "sort"

type TeamStanding struct {
	Position     int    `json:"position"`
	Team         string `json:"team"`
	Flag         string `json:"flag"`
	Played       int    `json:"played"`
	Won          int    `json:"won"`
	Drawn        int    `json:"drawn"`
	Lost         int    `json:"lost"`
	GoalsFor     int    `json:"goals_for"`
	GoalsAgainst int    `json:"goals_against"`
	GoalDiff     int    `json:"goal_difference"`
	Points       int    `json:"points"`
}

type GroupStanding struct {
	Group string         `json:"group"`
	Teams []TeamStanding `json:"teams"`
}

func CalculateStandings(matches []Match) []GroupStanding {
	groupSet := make(map[string]bool)
	for _, m := range matches {
		groupSet[m.Group] = true
	}

	groupNames := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	result := make([]GroupStanding, 0, len(groupNames))

	for _, groupName := range groupNames {
		teamSet := make(map[string]*TeamStanding)
		for _, m := range matches {
			if m.Group != groupName {
				continue
			}
			if _, ok := teamSet[m.HomeTeam]; !ok {
				teamSet[m.HomeTeam] = &TeamStanding{Team: m.HomeTeam, Flag: m.HomeFlag}
			}
			if _, ok := teamSet[m.AwayTeam]; !ok {
				teamSet[m.AwayTeam] = &TeamStanding{Team: m.AwayTeam, Flag: m.AwayFlag}
			}
		}

		for _, m := range matches {
			if m.Group != groupName {
				continue
			}
			if m.Status != "finished" {
				continue
			}
			home := teamSet[m.HomeTeam]
			away := teamSet[m.AwayTeam]
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

		teams := make([]TeamStanding, 0, len(teamSet))
		for _, t := range teamSet {
			t.GoalDiff = t.GoalsFor - t.GoalsAgainst
			teams = append(teams, *t)
		}

		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Points != teams[j].Points {
				return teams[i].Points > teams[j].Points
			}
			if teams[i].GoalDiff != teams[j].GoalDiff {
				return teams[i].GoalDiff > teams[j].GoalDiff
			}
			if teams[i].GoalsFor != teams[j].GoalsFor {
				return teams[i].GoalsFor > teams[j].GoalsFor
			}
			return teams[i].Team < teams[j].Team
		})

		for i := range teams {
			teams[i].Position = i + 1
		}

		result = append(result, GroupStanding{Group: groupName, Teams: teams})
	}

	return result
}
