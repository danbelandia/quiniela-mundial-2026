# Tasks: Predicciones de usuario clickeables desde el ranking

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~200 (8 modified, 1 new) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend endpoint + frontend detail page + clickable ranking | PR 1 | Single PR, ~200 lines, local-only first |

## Phase 1: Backend — Foundation (domain + port)

- [x] 1.1 Add `PointsEarned(actualHome, actualAway int) int` method to `domain.Prediction` in `backend/internal/core/domain/match.go` (or new `prediction.go` file)
- [x] 1.2 Add `PredictionWithMatch` struct in `backend/internal/core/ports/repositories.go` (embeds `Prediction` + match fields + `Points` field)
- [x] 1.3 Add `GetByUserIDWithMatch(userID int) ([]PredictionWithMatch, error)` method to `PredictionRepository` interface in same file

## Phase 2: Backend — Core Implementation

- [x] 2.1 Implement `GetByUserIDWithMatch` in `SQLiteRepository` (`sqlite_repository.go`) with `LEFT JOIN matches ON predictions.match_id = matches.id`, ordered by `match_date`
- [x] 2.2 Refactor `RankingService.CalculatePoints` (`services.go:61`) to delegate to `pred.PointsEarned(actualHome, actualAway)` (keep the method signature for backwards compat)
- [x] 2.3 Add `GetUserPredictions(userID int) ([]PredictionWithMatch, error)` method to `PredictionService` in same file — calls repo, iterates and sets `Points` using `PointsEarned`
- [x] 2.4 Add `GetUserPredictions(w, r)` handler in `handlers.go` — validates user exists via `userService.GetByID`, returns 404 if not, calls predictionService, returns JSON `{user, predictions}`
- [x] 2.5 Add route `mux.HandleFunc("GET /users/{id}/predictions", userHandler.GetUserPredictions)` in `main.go`

## Phase 3: Frontend — Page + Routing + Ranking

- [x] 3.1 Create `frontend/src/pages/UserPredictionsPage.tsx` with `useParams<{id: string}>`, `useEffect` to fetch, `useState` for predictions, `useState` for active group tab
- [x] 3.2 Render 12 group buttons (A-L) at top, show "Volver al ranking" link to `/ranking`
- [x] 3.3 Render table per group: `Match (Local | Visitante) | Real | Pronóstico | Puntos` with empty prediction showing "—"
- [x] 3.4 Group predictions by `group_name` client-side, filter by active tab
- [x] 3.5 Add `import { Link }` to `RankingPage.tsx` and wrap the username `<td>` content in `<Link to={'/user/' + u.id} className="hover:underline cursor-pointer">`
- [x] 3.6 Add route `<Route path="/user/:id" element={<UserPredictionsPage />} />` inside the `ProtectedRoute` in `App.tsx`

## Phase 4: Verification (manual)

- [x] 4.1 Start backend (`go run ./cmd/api`) and frontend (`npm run dev`) in `feature/clickable-user-predictions` branch
- [x] 4.2 Backend test 1: `curl GET /users/1/predictions` with `dangel` → 200 + JSON with predictions
- [x] 4.3 Backend test 2: `curl GET /users/9999/predictions` → 404
- [x] 4.4 Backend test 3: `curl GET /users/{noPredictionsUser}/predictions` → 200 + 72 matches with `has_prediction: false` (LEFT JOIN, spec adjusted)
- [x] 4.5 Backend test 4: verify `points` field in response: 3 for exact (user test), 2 for correct winner (dangel), 0 for other (multiple)
- [x] 4.6 Frontend test 1: login → `/ranking` → click "dangel" → URL changes to `/user/1` and detail page loads
- [x] 4.7 Frontend test 2: click tab "B" → only group B matches visible
- [x] 4.8 Frontend test 3: "Volver al ranking" → URL `/ranking`
- [x] 4.9 Mobile check: 12 group tabs don't break layout, table scrollable

## Phase 5: Commit

- [ ] 5.1 Commit with conventional message: `feat(ranking): add clickable user predictions detail page`
- [ ] 5.2 Stay on `feature/clickable-user-predictions` branch (DO NOT push — user is working locally)
- [ ] 5.3 When ready: `git checkout master && git merge feature/clickable-user-predictions && git push origin master`
- [ ] 5.4 After verify in prod, archive this change (sdd-archive phase)
