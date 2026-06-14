package domain

import (
	"testing"
	"time"
)

func TestMatch_IsEffectivelyLocked(t *testing.T) {
	kickoff := time.Date(2026, 6, 11, 19, 0, 0, 0, time.UTC)

	t.Run("3h default window", func(t *testing.T) {
		window := 3 * time.Hour
		cases := []struct {
			name   string
			locked bool
			now    time.Time
			want   bool
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
	})

	t.Run("15min production window", func(t *testing.T) {
		window := 15 * time.Minute
		cases := []struct {
			name   string
			locked bool
			now    time.Time
			want   bool
		}{
			{
				name:   "manual lock overrides 15min window",
				locked: true,
				now:    kickoff.Add(-24 * time.Hour),
				want:   true,
			},
			{
				name:   "16min before kickoff is unlocked",
				locked: false,
				now:    kickoff.Add(-16 * time.Minute),
				want:   false,
			},
			{
				name:   "14min before kickoff is locked",
				locked: false,
				now:    kickoff.Add(-14 * time.Minute),
				want:   true,
			},
			{
				name:   "exactly at 15min boundary is locked",
				locked: false,
				now:    kickoff.Add(-15 * time.Minute),
				want:   true,
			},
			{
				name:   "10min before kickoff is locked",
				locked: false,
				now:    kickoff.Add(-10 * time.Minute),
				want:   true,
			},
			{
				name:   "1 hour before kickoff is unlocked",
				locked: false,
				now:    kickoff.Add(-1 * time.Hour),
				want:   false,
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
	})
}
