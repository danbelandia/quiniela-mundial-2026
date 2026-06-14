# Proposal: Shorten Match Lock Window From 3h to 15min Before Kickoff

## Intent

After the first jornada of the 2026 World Cup, gameplay observation shows that 3 hours is too long: users keep editing predictions in the last hour as team news breaks (lineups, injuries), which makes the standings meaningless from minute 1 of the match. The change shortens the auto-lock window to 15 minutes before kickoff so that the "last call" period is short and meaningful, while still leaving room for legitimate late edits. The mechanism is unchanged: the window is a `LOCK_WINDOW_HOURS` env var, default stays at `3` for safety, production sets it to `0.25`.

## Scope

### In Scope
- Production configuration: set `LOCK_WINDOW_HOURS=0.25` in Render (already done by operator).
- Documentation: update `backend/.env.example` to `0.25` and update the `auto-lock-matches` spec to reflect the new operational default.
- Tests: add boundary cases for the 15-minute window in `backend/internal/core/domain/match_test.go`.

### Out of Scope
- Code changes to domain, services, repo, or handlers. The `IsEffectivelyLocked(now, window)` helper is generic over the window duration; no behavior change is required.
- Frontend changes. The UI already reads `match_lock_window_hours` dynamically from `GET /config`.
- Changing the default in the Go fallback. The default stays at `3h` for safety; production overrides via env var. This avoids accidental shortening on any future self-hosted deployment.
- Migration of the `predictions` table. Already-submitted predictions are not invalidated.
- Touching the `is_locked` manual override column or the `/admin/matches/lock*` endpoints.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `auto-lock-matches`: the operational lock window is now `15min` in production. The spec text, scenarios, and examples are updated to reflect this. The mechanism (env var, default fallback, UTC comparison, manual override) is unchanged.

## Approach

The change is configuration-first, docs-second, tests-third. The mechanism designed in the original `2026-06-06-auto-lock-matches` change already isolates the window in a single env var, so the production switch is one Render setting. The repo just needs to reflect the new operational reality in `.env.example`, the spec, and the test file.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/.env.example` | Modified | `LOCK_WINDOW_HOURS=3` → `LOCK_WINDOW_HOURS=0.25` |
| `openspec/specs/auto-lock-matches/spec.md` | Modified | Update purpose, scenarios, and `GET /config` example to 15min |
| `backend/internal/core/domain/match_test.go` | Modified | Add 15min boundary cases (14min, 15min exact, 16min) alongside the existing 3h cases |
| Render (production env) | Modified | Operator has set `LOCK_WINDOW_HOURS=0.25` (out of repo) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Users with slow connections miss the short window | Med | Window of 15min is a deliberate product decision; the manual admin override (`/admin/matches/lock`) is always available as escape hatch |
| Operator forgets to set the env var on a future redeploy | Low | `GET /config` exposes the live value; smoke test on deploy; default fallback is 3h which is the safe direction |
| Tests break due to env var behavior in CI | None | The test uses an explicit `window` variable, not the env var; the boundary cases are parametrized |
| Spec drift between code and docs | Low | This change syncs the spec; archive step moves the delta into the canonical spec |

## Rollback Plan

In Render: set `LOCK_WINDOW_HOURS=3` (or remove the var to use the 3h default) and redeploy. The behavior reverts to the previous 3h window. In the repo: revert the commit that updates `.env.example` and the spec, then re-archive the change. The Go code is untouched so no rollback is needed there.

## Success Criteria

- [ ] Production `GET /config` returns `"match_lock_window_hours": 0.25`
- [ ] A match whose `match_date` is 16 minutes in the future is reported with `is_locked = false`
- [ ] A match whose `match_date` is 14 minutes in the future is reported with `is_locked = true`
- [ ] `POST /predictions` for a match 14 minutes before kickoff returns `403` with body `match is locked`
- [ ] `go test ./...` passes (existing 3h cases + new 15min cases)
- [ ] `.env.example` reads `LOCK_WINDOW_HOURS=0.25`
