package domain

import (
	"sort"
	"testing"
)

// TestRankingTieBreaker exercises the comparator order that the frontend
// applies on top of the GET /ranking response:
//  1. total (score + qualifier_score + top_scorer_score) descending
//  2. exact_score descending
//  3. winner_score descending
//  4. id ascending
//
// The comparator itself lives in the frontend (RankingPage.tsx); this test
// mirrors the logic in Go so the contract is locked in even if the sort
// is later moved server-side.
func TestRankingTieBreaker(t *testing.T) {
	total := func(u User) int {
		return u.Score + u.QualifierScore + u.TopScorerScore
	}
	less := func(a, b User) bool {
		ta, tb := total(a), total(b)
		if ta != tb {
			return ta > tb
		}
		if a.ExactScore != b.ExactScore {
			return a.ExactScore > b.ExactScore
		}
		if a.WinnerScore != b.WinnerScore {
			return a.WinnerScore > b.WinnerScore
		}
		return a.ID < b.ID
	}

	cases := []struct {
		name    string
		users   []User
		wantIDs []int
	}{
		{
			name: "higher total wins regardless of exact counts",
			users: []User{
				{ID: 1, Score: 60, ExactScore: 0, WinnerScore: 30},
				{ID: 2, Score: 90, ExactScore: 0, WinnerScore: 45},
			},
			wantIDs: []int{2, 1},
		},
		{
			name: "tie on total — more exact scores wins",
			users: []User{
				{ID: 1, Score: 50, ExactScore: 4, WinnerScore: 6},
				{ID: 2, Score: 50, ExactScore: 6, WinnerScore: 2},
			},
			wantIDs: []int{2, 1},
		},
		{
			name: "tie on total and exact — more winner scores wins",
			users: []User{
				{ID: 1, Score: 50, ExactScore: 4, WinnerScore: 5},
				{ID: 2, Score: 50, ExactScore: 4, WinnerScore: 7},
			},
			wantIDs: []int{2, 1},
		},
		{
			name: "full tie — lower id wins",
			users: []User{
				{ID: 7, Score: 50, ExactScore: 4, WinnerScore: 5},
				{ID: 3, Score: 50, ExactScore: 4, WinnerScore: 5},
			},
			wantIDs: []int{3, 7},
		},
		{
			name: "combined — total, then exact, then winner, then id",
			users: []User{
				{ID: 1, Score: 30, QualifierScore: 0, TopScorerScore: 0, ExactScore: 4, WinnerScore: 9},
				{ID: 2, Score: 30, QualifierScore: 0, TopScorerScore: 0, ExactScore: 5, WinnerScore: 7},
				{ID: 3, Score: 30, QualifierScore: 0, TopScorerScore: 0, ExactScore: 4, WinnerScore: 9},
				{ID: 4, Score: 25, QualifierScore: 0, TopScorerScore: 0, ExactScore: 8, WinnerScore: 0},
			},
			wantIDs: []int{2, 1, 3, 4},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := make([]User, len(c.users))
			copy(got, c.users)
			sort.SliceStable(got, func(i, j int) bool { return less(got[i], got[j]) })
			for i, want := range c.wantIDs {
				if got[i].ID != want {
					t.Errorf("position %d: got id %d, want %d (full order: %v)", i, got[i].ID, want, idsOf(got))
				}
			}
		})
	}
}

func idsOf(users []User) []int {
	out := make([]int, len(users))
	for i, u := range users {
		out[i] = u.ID
	}
	return out
}
