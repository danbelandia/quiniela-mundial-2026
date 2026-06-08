# Tasks: Top Scorer of the Group Stage Prediction

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~700 (15 files: 7 backend, 7 frontend, 1 new) |
| 400-line budget risk | High — split into 2 PRs is recommended |
| Chained PRs recommended | Yes |
| Suggested split | (1) backend + DB + spec; (2) frontend page + integrations + smoke |
| Delivery strategy | chained-pr |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: backend-first, frontend-second
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend: domain + repo + service + handlers + routes + smoke | PR 1 | `feature/top-scorer-prediction` → master (backend) |
| 2 | Frontend: page + components + integrations + smoke | PR 2 | Same branch (no merge between) or split if review focus requires |

For simplicity in this initial implementation, we keep a single branch with backend + frontend in a logical commit order. If a chained-PR is required by the user, split at the integration point (after backend smoke, before frontend work).

## Phase 1: Data — candidates JSON + DB schema

- [ ] 1.1 Create `backend/data/top_scorer_candidates.json` with the 50 candidates (already provided by user; saved)
- [ ] 1.2 In `sqlite_repository.go`, add DDL for `top_scorer_predictions` table:
  ```sql
  CREATE TABLE IF NOT EXISTS top_scorer_predictions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    predicted_player TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
  )
  ```
- [ ] 1.3 Add DDL for `app_config` table:
  ```sql
  CREATE TABLE IF NOT EXISTS app_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME NOT NULL
  )
  ```
- [ ] 1.4 Run `go build ./...` — must compile

## Phase 2: Domain + ports

- [ ] 2.1 Create `backend/internal/core/domain/top_scorer.go` with:
  - `TopScorerCandidate` struct (`Name`, `Team`, `Flag`)
  - `TopScorerPrediction` struct (`ID`, `UserID`, `PredictedPlayer`, `CreatedAt`, `UpdatedAt`)
  - `TopScorerPointsEarned(predicted, actual string) int` — returns 6 if normalized match, 0 otherwise (or 0 if actual == "")
  - `NormalizeName(s string) string` — lowercase + Unicode NFD + strip combining marks (accents)
- [ ] 2.2 In `backend/internal/core/ports/repositories.go`, extend the user repository interface (or add a new `TopScorerRepository` interface) with:
  - `UpsertTopScorerPrediction(userID int, player string) error`
  - `GetTopScorerPredictionByUserID(userID int) (*domain.TopScorerPrediction, error)`
  - `GetAllTopScorerPredictions() ([]domain.TopScorerPrediction, error)`
- [ ] 2.3 Add new `AppConfigRepository` interface (or extend an existing one) with:
  - `GetAppConfig(key string) (string, error)` — returns "" if not set
  - `SetAppConfig(key, value string) error`
- [ ] 2.4 Run `go build ./...` — must compile

## Phase 3: Repository implementations

- [ ] 3.1 In `sqlite_repository.go`, implement `UpsertTopScorerPrediction` with `INSERT ... ON CONFLICT(user_id) DO UPDATE SET predicted_player=excluded.predicted_player, updated_at=excluded.updated_at`
- [ ] 3.2 Implement `GetTopScorerPredictionByUserID` and `GetAllTopScorerPredictions`
- [ ] 3.3 Implement `GetAppConfig` and `SetAppConfig` against the `app_config` table
- [ ] 3.4 Run `go build ./...` — must compile

## Phase 4: Service

- [ ] 4.1 In `services.go`, add `TopScorerService` struct with fields: `userRepo`, `cfgRepo`, `qualifierLockAt time.Time`, `candidates []domain.TopScorerCandidate`
- [ ] 4.2 Add `NewTopScorerService(userRepo, cfgRepo, lockAt, candidates)` constructor
- [ ] 4.3 Implement `GetMine(userID int) (*domain.TopScorerPrediction, error)` — returns the user's prediction or a "not found" sentinel (handler maps to 404)
- [ ] 4.4 Implement `Upsert(userID int, player string) error` — checks `qualifierLockAt`, checks `player` is in `candidates` (returns `ErrInvalidPlayer` if not), then writes
- [ ] 4.5 Implement `GetByUserID(userID int) (*domain.TopScorerPrediction, error)`
- [ ] 4.6 Implement `SetActual(player string) error` — admin write; warn-log if `player` not in candidates but still accept
- [ ] 4.7 Implement `ScoreUser(userID int) (int, error)` — reads actual from config, reads user's prediction, calls `TopScorerPointsEarned`
- [ ] 4.8 Add `var ErrInvalidPlayer = errors.New("player not in candidate list")` sentinel
- [ ] 4.9 Run `go build ./...` — must compile

## Phase 5: Handlers + routes

- [ ] 5.1 In `handlers.go`, add `TopScorerHandler` struct with fields: `Service *services.TopScorerService`
- [ ] 5.2 Implement `TopScorerHandler.UpsertMine(w, r)` — reads body `{user_id, predicted_player}`, calls `Service.Upsert`, maps errors:
  - `ErrMatchLocked` → 403 "match is locked"
  - `ErrInvalidPlayer` → 400 "player not in candidate list"
  - default → 500
  - success → 200/201
- [ ] 5.3 Implement `TopScorerHandler.GetMine(w, r)` — reads `user_id` from query, calls `Service.GetMine`, returns 200 with the prediction or 404
- [ ] 5.4 Implement `TopScorerHandler.GetByUserID(w, r)` — extracts `{id}` from path, returns 200/404
- [ ] 5.5 Add `TopScorerAdminHandler` with `SetActual(w, r)` — reads body `{player}`, calls `Service.SetActual`, returns 200
- [ ] 5.6 Extend `ConfigHandler` to include `TopScorerCandidates []domain.TopScorerCandidate`; serialize in `Get`
- [ ] 5.7 Extend `RankingHandler.Get` to compute `top_scorer_score` per user (calls `TopScorerService.ScoreUser` per user) and include it in the response
- [ ] 5.8 Run `go build ./...` — must compile

## Phase 6: Wiring (main.go)

- [ ] 6.1 In `main.go`, add a helper `loadCandidates() []domain.TopScorerCandidate` that reads `data/top_scorer_candidates.json`; on error, log a warning and return an empty slice (do not crash)
- [ ] 6.2 Call `loadCandidates()` at startup
- [ ] 6.3 Construct `topScorerService` and pass to:
  - `configHandler.TopScorerCandidates = candidates`
  - `rankingHandler.TopScorerService = topScorerService`
  - new `topScorerHandler` and `topScorerAdminHandler`
- [ ] 6.4 Register routes:
  - `GET /top-scorer-prediction/me`
  - `PUT /top-scorer-prediction/me`
  - `GET /users/{id}/top-scorer-prediction`
  - `POST /admin/top-scorer`
  - (and their `OPTIONS` CORS preflights)
- [ ] 6.5 Run `go build ./...` — must compile
- [ ] 6.6 Run `go run ./cmd/api` and verify:
  - `[GIN-debug] Listening on :8080`
  - log: `loaded 50 top scorer candidates`
  - `GET /config` now includes `top_scorer_candidates` array
  - `GET /top-scorer-prediction/me?user_id=1` → 404 (no pick yet)
  - `PUT /top-scorer-prediction/me` with valid body → 200
  - `PUT /top-scorer-prediction/me` with `"Some Random Player"` → 400
  - `GET /top-scorer-prediction/me?user_id=1` → 200 with the saved pick

## Phase 7: Frontend hooks + config

- [ ] 7.1 In `frontend/src/features/hooks.ts`, add:
  - `useTopScorerCandidates()` — reads `config.top_scorer_candidates` (extend the existing `useConfig` if simpler)
  - `useMyTopScorerPrediction(userId)` — fetches `GET /top-scorer-prediction/me?user_id=<id>`
  - `useUpsertTopScorerPrediction()` — returns a function that PUTs the new pick
  - `useUserTopScorerPrediction(userId)` — fetches `GET /users/{id}/top-scorer-prediction`

## Phase 8: Frontend page

- [ ] 8.1 Create `frontend/src/pages/GoleadorPage.tsx`:
  - Loads candidates from `useConfig` (or `useTopScorerCandidates`)
  - Loads my current pick from `useMyTopScorerPrediction`
  - Renders a `<select>` with all candidates (grouped by team? or flat with team as subtitle?)
  - "Guardar" / "Cambiar" button
  - Lock banner: if `Date.now() >= new Date(config.qualifier_lock_at).getTime()`, show locked state with disabled controls
- [ ] 8.2 Create `frontend/src/components/TopScorerPredictionCard.tsx`:
  - Props: `predictedPlayer: string | null`, `actualPlayer: string | null` (optional, for showing after admin sets), `points: number` (optional)
  - Renders: "Goleador fase de grupos" card with the player name + flag + team, or "Aún no pronosticó" if null
  - If `actualPlayer` is set and matches, shows a green checkmark + points
- [ ] 8.3 Register `/goleador` route in the router (likely `frontend/src/App.tsx`)
- [ ] 8.4 Add a link to the new page in the nav bar (`NavBar.tsx` or wherever the existing links are)
- [ ] 8.5 Run `npm run build` — must compile (Vite does inline TS typecheck)

## Phase 9: Frontend integrations

- [ ] 9.1 In `UserPredictionsPage.tsx`, add `<TopScorerPredictionCard>` ABOVE the `<QualifierPredictionCard>`
- [ ] 9.2 In `RankingPage.tsx`, add a new `<th>Goleador</th>` column between "Clasificación" and "Total", and a new `<td>{u.top_scorer_score || 0}</td>` cell
- [ ] 9.3 In `AdminPage.tsx`, add a new section "Goleador de la fase de grupos":
  - Dropdown of candidates (default)
  - Optional free-text input (escape hatch for dark horses)
  - "Guardar resultado" button that POSTs to `/admin/top-scorer`
  - Show a warning if the free-text name is not in the list (but the API also warns server-side)
- [ ] 9.4 Run `npm run build` — must compile

## Phase 10: Smoke test

- [ ] 10.1 Start the API with default `QUALIFIER_LOCK_AT` in the future
- [ ] 10.2 Login as dangel, save pick "Kylian Mbappé" via `PUT /top-scorer-prediction/me` → expect 200
- [ ] 10.3 Read back via `GET /top-scorer-prediction/me?user_id=1` → expect 200 with the pick
- [ ] 10.4 Try to save an invalid name "Random" → expect 400
- [ ] 10.5 Login as test (user 2), save pick "Harry Kane" → expect 200
- [ ] 10.6 Login as vacio (user 3), do NOT save any pick
- [ ] 10.7 Set the actual via `POST /admin/top-scorer {player: "Kylian Mbappé"}` → expect 200
- [ ] 10.8 `GET /ranking` → verify dangel has `top_scorer_score: 6`, test has 0, vacio has 0
- [ ] 10.9 Test case-insensitive: change dangel's pick to "KYLIAN MBAPPÉ" (uppercase), re-check — dangel still has 6 (case-insensitive match). To do this in the smoke test, we'd restart the API with `QUALIFIER_LOCK_AT` in the future (default), re-PUT, then verify
- [ ] 10.10 Override `QUALIFIER_LOCK_AT=2026-06-06T18:00:00Z` (1h in the past), restart, `PUT /top-scorer-prediction/me` → expect 403 "match is locked"
- [ ] 10.11 Restore `QUALIFIER_LOCK_AT` to default, restart, verify lock reopens
- [ ] 10.12 Test the rival profile view: `GET /users/1/top-scorer-prediction` as user 2 → expect 200 with dangel's pick
- [ ] 10.13 Admin sets actual to a name not in the list: `POST /admin/top-scorer {player: "Hansi Flick"}` → expect 200 + warning log; verify all users have `top_scorer_score: 0`
- [ ] 10.14 Clean up: delete test picks (or leave them, depending on test data policy)

## Phase 11: Commit + (NO push, NO merge)

- [ ] 11.1 Commit on `feature/top-scorer-prediction` (single PR or split per chained-PR decision):
  - `feat(domain): add TopScorerCandidate, TopScorerPrediction, scoring + NormalizeName`
  - `feat(repo): add top_scorer_predictions and app_config tables, repo methods`
  - `feat(service): add TopScorerService with lock + candidate validation`
  - `feat(handler): add TopScorerHandler + admin handler + extend ConfigHandler for candidates`
  - `feat(ranking): compute top_scorer_score per user in /ranking`
  - `chore(main): wire TopScorerService, load candidates JSON, register routes`
  - `feat(frontend): add /goleador page with dropdown + lock banner`
  - `feat(frontend): add TopScorerPredictionCard, integrate to UserPredictionsPage + AdminPage + RankingPage`
  - `chore(openspec): add top-scorer-prediction change artifacts`
- [ ] 11.2 Stop. Do NOT push. Do NOT merge. Report results to user and wait for authorization.

## Notes

- All file paths are relative to repo root.
- Times are UTC everywhere; `NormalizeName` is pure and can be unit-tested if desired (not added in v1 to keep parity with the project).
- The candidates JSON ships in the binary's working dir; on Render, the file is part of the deploy artifact. No path env var needed.
- Lock reuses `QUALIFIER_LOCK_AT`; if the user later wants a separate top-scorer deadline, a follow-up change can introduce `TOP_SCORER_LOCK_AT` with minimal churn.
- No DB migration needed — the new tables use `CREATE TABLE IF NOT EXISTS` and have no foreign data to backfill.
- The `app_config` table is generic and can store future single-value config (e.g., "tournament_year": "2026").
- The candidate list is provided as-is by the user. If duplicates or unwanted entries exist, the user can edit the JSON and redeploy.
