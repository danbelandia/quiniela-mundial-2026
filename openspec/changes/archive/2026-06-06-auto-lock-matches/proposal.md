# Proposal: Auto-Lock Matches 3h Before Kickoff

## Intent

Today, match locking is 100% manual: an admin must call `POST /admin/matches/lock` before each game. With 72 matches in 30 days this is unworkable. Worse, the server **does not validate `is_locked`** when a prediction is submitted — the frontend disables inputs, but anyone with `curl` can bypass it (handlers.go:43-62). This change adds automatic time-based locking and closes the server-side validation gap.

## Scope

### In Scope
- Auto-lock each match exactly 3 hours before its `match_date` (UTC).
- Server-side validation on `POST /predictions` and qualifier `Upsert` returning 403 when locked.
- Configurable window via `LOCK_WINDOW_HOURS` env var (default 3) for testing and future tuning.
- Per-match evaluation (group or jornada level locking is NOT in scope).
- Preserve the manual admin override (`/admin/matches/lock*`).

### Out of Scope
- Render cron job that writes lock state to DB (deferred — Opción A from exploration).
- UI changes beyond the existing `is_locked` flag the frontend already reads.
- Timezone display in the frontend (kickoff time remains UTC, frontend already shows CLT).
- Notifications/emails to users when their prediction window closes.
- Unlocking after auto-lock (admin can still force-set, but next read recomputes).

## Capabilities

### New Capabilities
- `auto-lock-matches`: defines the effective lock state computation, server-side enforcement, and admin override semantics.

### Modified Capabilities
- None. `user-predictions-detail` only describes read views and is unaffected at the spec level.

## Approach

Hybrid effective-lock model. Add `func (m Match) IsEffectivelyLocked(now time.Time) bool` in `domain/match.go` returning `m.IsLocked || now.Add(window).After(m.Date)`. Override `IsLocked` in repo `Scan` blocks (sqlite_repository.go:161, 175, 253) with the computed value. Add 403 checks in `CreatePrediction` (handlers.go:43-62) and `QualifierPredictionHandler.Upsert` (handlers.go:398-451). The admin endpoints stay untouched. A small table-driven test in `domain/match_test.go` covers the helper (the project has zero tests today, this is the seed test).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/core/domain/match.go` | Modified | New `IsEffectivelyLocked` helper |
| `backend/internal/core/domain/match_test.go` | New | First test file in project |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modified | Override `IsLocked` after Scan in 3 read paths |
| `backend/internal/application/services/services.go` | Modified | Read `LOCK_WINDOW_HOURS`; pass to helpers (or inject via constructor) |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | 403 on locked match in `CreatePrediction` and qualifier `Upsert` |
| `backend/cmd/api/main.go` | Modified | Build `lockWindow` from env, inject into services |
| `backend/.env.example` | New | Document `LOCK_WINDOW_HOURS` |
| `openspec/specs/auto-lock-matches/spec.md` | New | Synced on archive |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Stale `is_locked` in DB overrides computed value the wrong way | Low | Effective lock ORs both — DB flag forces lock, never opens it after date |
| TZ bug: helper uses local time instead of UTC | Med | `time.Now().UTC()` everywhere; tests assert UTC |
| Server-side check forgotten in some submit path | Med | Grep for all `INSERT INTO predictions` and `INSERT INTO group_qualifier_predictions`; add spec scenario |
| Frontend keeps stale `is_locked: false` in component state | Low | Frontend already calls `GET /matches` on mount; refetch on tab focus is a nice-to-have, not blocker |
| Test infra doesn't exist; setting up Go test runner | Low | `go test` is built-in, no setup; just need `_test.go` file |

## Rollback Plan

Revert the merge commit on `master`. The DB schema and `is_locked` column are unchanged, so the rollback is purely code: remove the helper, undo the Scan overrides, drop the 403 checks. Admin manual lock continues to work as before. No data migration needed.

## Success Criteria

- [ ] Unit test for `IsEffectivelyLocked` passes (5 cases: manual lock, before window, inside window, after kickoff, exact boundary)
- [ ] `go test ./...` returns 0
- [ ] `POST /predictions` for a match whose `match_date - now < 3h` returns 403 with a clear message
- [ ] Qualifier `Upsert` for a group with a match inside the window returns 403
- [ ] Admin `POST /admin/matches/lock {id: X, locked: true}` still works and is reflected immediately
- [ ] With `LOCK_WINDOW_HOURS=0.01` set, a match becomes locked within ~36 seconds in local dev
- [ ] Smoke test in production post-deploy: existing user can still predict a match >3h away
