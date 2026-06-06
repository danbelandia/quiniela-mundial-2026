package domain

import (
	"testing"
	"time"
)

func TestMatch_IsEffectivelyLocked(t *testing.T) {
	kickoff := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC)
	window := 3 * time.Hour

	cases := []struct {
		name    string
		locked  bool
		now     time.Time
		want    bool
	}{
		{
			name:   "manual lock overrides window",
			locked: true,
			now:    kickoff.Add(-24 * time.Hour),
			want:   true,
		},
		{
			name:   "3h+1min before kickoff is unlocked",
			locked: false,
			now:    kickoff.Add(-(3*time.Hour + time.Minute)),
			want:   false,
		},
		{
			name:   "3h-1min before kickoff is locked",
			locked: false,
			now:    kickoff.Add(-(3*time.Hour - time.Minute)),
			want:   true,
		},
		{
			name:   "exactly at window boundary is locked",
			locked: false,
			now:    kickoff.Add(-3 * time.Hour),
			want:   true,
		},
		{
			name:   "1 minute after kickoff is locked",
			locked: false,
			now:    kickoff.Add(time.Minute),
			want:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := Match{IsLocked: c.locked, Date: kickoff}
			if got := m.IsEffectivelyLocked(c.now, window); got != c.want {
				t.Errorf("IsEffectivelyLocked() = %v, want %v", got, c.want)
			}
		})
	}
}
