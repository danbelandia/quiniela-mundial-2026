# Tasks: Auto-Lock Matches 3h Before Kickoff

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~170 (8 files) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full feature | PR 1 | Single branch `feature/auto-lock-matches` → master after smoke test |

## Phase 1: Foundation (domain + test)

- [ ] 1.1 Add `IsEffectivelyLocked(now time.Time, window time.Duration) bool` method to `domain.Match` in `backend/internal/core/domain/match.go`
- [ ] 1.2 Create `backend/internal/core/domain/match_test.go` with 5 table-driven cases (manual override, in-window, out-of-window, boundary `>=`, post-kickoff)
- [ ] 1.3 Run `go test ./internal/core/domain/...` — must be green (this is the project's first test)

## Phase 2: Repository (override at read time)

- [ ] 2.1 Add `lockWindow time.Duration` and `clock func() time.Time` fields to `SQLiteRepository` struct in `backend/internal/infrastructure/repository/sqlite_repository.go`
- [ ] 2.2 Change `NewSQLiteRepository` signature to accept `lockWindow time.Duration`; init `clock = time.Now` if nil
- [ ] 2.3 After Scan in `GetAllMatches` (line ~161), override each match's `IsLocked` with `m.IsEffectivelyLocked(r.clock().UTC(), r.lockWindow)`
- [ ] 2.4 After Scan in `GetMatchByID` (line ~175), same override
- [ ] 2.5 After Scan in `GetByUserIDWithMatch` (line ~253), same override (consistency, even though frontend doesn't read it there)
- [ ] 2.6 Run `go build ./...` — must compile

## Phase 3: Service + Handler (server-side enforcement)

- [ ] 3.1 Add `var ErrMatchLocked = errors.New("match is locked")` to `backend/internal/application/services/services.go`
- [ ] 3.2 Add `matchRepo ports.MatchRepository` field to `PredictionService` struct; update `NewPredictionService` signature
- [ ] 3.3 In `PredictionService.PlacePrediction`, fetch the match by ID and return `ErrMatchLocked` if `match.IsEffectivelyLocked(time.Now().UTC(), window)` — note: window is repo-owned, expose via method or pass-through. **Decision**: add `IsLocked(ctx, matchID)` method to `MatchService` that returns the effective state
- [ ] 3.4 In `QualifierPredictionService.Upsert`, fetch all matches for the group; return `ErrMatchLocked` if any is effectively locked (compute via the same helper using a fixed window, e.g. read from `QualifierService` or accept as a ctor param)
- [ ] 3.5 In `handlers.go` `CreatePrediction`, map `errors.Is(err, services.ErrMatchLocked)` → `http.Error(w, "match is locked", http.StatusForbidden)`
- [ ] 3.6 In `handlers.go` `QualifierPredictionHandler.Upsert`, same mapping
- [ ] 3.7 Run `go build ./...` — must compile

## Phase 4: Wiring (main + config)

- [ ] 4.1 In `backend/cmd/api/main.go`, add `buildLockWindow()` helper that reads `LOCK_WINDOW_HOURS`, parses via `strconv.ParseFloat`, returns `time.Duration`; default 3h, warning log on invalid
- [ ] 4.2 Pass `lockWindow` to `NewSQLiteRepository(driver, conn, lockWindow)`
- [ ] 4.3 Pass the same `lockWindow` (or expose via repo) to `MatchService` and `QualifierPredictionService` constructors — pick the cleanest path: either add to repo's exported accessors, or pass `lockWindow` directly to services
- [ ] 4.4 Update `main.go` service constructors for `PredictionService`, `MatchService`, `QualifierPredictionService` to receive the window
- [ ] 4.5 Create `backend/.env.example` with `LOCK_WINDOW_HOURS=3` and a comment
- [ ] 4.6 Run `go build ./...` — must compile

## Phase 5: Local verification (no push, no prod)

- [ ] 5.1 `go test ./...` — all green
- [ ] 5.2 Build + start API with `LOCK_WINDOW_HOURS=0.01` (~36s window)
- [ ] 5.3 Use `sqlite3` to bump one match's `match_date` to `datetime('now', '+30 seconds')` — pick id=1 (MEX-RSA)
- [ ] 5.4 `curl GET /matches` — confirm `is_locked: true` within 40s
- [ ] 5.5 `curl POST /predictions` with that match's id — expect `403 match is locked`
- [ ] 5.6 Pick a match 5h out, `curl POST /predictions` — expect `201 Created`
- [ ] 5.7 `curl POST /admin/matches/lock {id: 1, locked: true}` — confirm `GET /matches` shows it locked immediately (manual override works)
- [ ] 5.8 Pick a group with one match in the window, `curl PUT /groups/A/qualifier-predictions` — expect `403 group is locked`
- [ ] 5.9 Pick a fully-open group, `curl PUT /groups/...` — expect `200 OK`
- [ ] 5.10 Commit on `feature/auto-lock-matches` (single commit or split: `feat(domain):`, `feat(repo):`, `feat(service):`, `chore(config):`)
- [ ] 5.11 Stop. Do NOT push. Do NOT merge. Report results to user.

## Notes

- All file paths are relative to repo root.
- Times are UTC everywhere; the helper is pure (takes `now` as param) so tests don't need a clock mock.
- The `IsEffectivelyLocked` helper is in `domain`, not `ports`, because it's behavior of the entity, not infrastructure.
- The `lockWindow` is duplicated in the repo and the services for clarity; alternative is to expose `repo.LockWindow()` and read from there. **Decision during apply**: pick whichever is cleaner after seeing the actual call sites.
