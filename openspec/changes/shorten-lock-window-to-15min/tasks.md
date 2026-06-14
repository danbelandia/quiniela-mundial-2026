# Tasks: Shorten Match Lock Window From 3h to 15min

## 1. Production configuration

- [x] 1.1 Set `LOCK_WINDOW_HOURS=0.25` in Render (done by operator before this change)
- [x] 1.2 Trigger redeploy on Render (done by operator)
- [ ] 1.3 Smoke test: `curl https://<api>/config` returns `"match_lock_window_hours": 0.25`
- [ ] 1.4 Smoke test: a match 16 minutes from now appears with `is_locked: false` in `GET /matches`
- [ ] 1.5 Smoke test: a match 14 minutes from now appears with `is_locked: true` in `GET /matches`

## 2. Repository documentation

- [x] 2.1 Update `backend/.env.example`: `LOCK_WINDOW_HOURS=3` → `LOCK_WINDOW_HOURS=0.25`
- [x] 2.2 Create this change folder `openspec/changes/shorten-lock-window-to-15min/`
- [x] 2.3 Write `proposal.md` describing intent, scope, and risks
- [x] 2.4 Write `specs/auto-lock-matches/spec.md` delta with 15-minute scenarios
- [x] 2.5 Write this `tasks.md`

## 3. Tests

- [x] 3.1 In `backend/internal/core/domain/match_test.go`, keep the existing 3-hour cases (regression coverage for the historical default)
- [x] 3.2 Add 15-minute boundary cases: 14min locked, 15min exact locked, 16min unlocked
- [ ] 3.3 Run `go test ./backend/...` (or `go test ./...` from backend/) and confirm all cases pass

## 4. Code

No code changes. The `IsEffectivelyLocked` helper, the repository overrides, and the service-side enforcement are all generic over the window duration. The change is configuration + docs + tests only.

## 5. Archive

- [ ] 5.1 After verification, archive this change by syncing the delta into `openspec/specs/auto-lock-matches/spec.md` (replace the 3-hour scenarios with the 15-minute ones) and moving this folder to `openspec/changes/archive/2026-06-XX-shorten-lock-window-to-15min/`
