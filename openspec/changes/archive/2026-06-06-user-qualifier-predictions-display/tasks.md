# Tasks: Display User Qualifier Predictions (1°/2° per group)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~180 (1 new component, 2 modified services, 2 modified handlers/routes, 1 new hook) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size:exception (not needed) |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend endpoint + frontend card + page wiring | PR 1 | base: `feature/user-qualifier-predictions-display`; all tests + smoke included |

## Phase 1: Backend DTO + Service

- [ ] 1.1 Add `QualifierPredictionView` struct in `backend/internal/core/ports/repositories.go` (12 fields per design)
- [ ] 1.2 Add `QualifierPredictionService.GetViewsForUser(userID int) ([]QualifierPredictionView, error)` in `backend/internal/application/services/services.go`
- [ ] 1.3 Service: fetch all matches once, build `map[group]map[team]flag`
- [ ] 1.4 Service: fetch user's qualifier predictions via `repo.GetAllQualifierPredictionsByUser`
- [ ] 1.5 Service: for each pick, populate flags from the map; if all 6 group matches `status="finished"`, call `domain.CalculateQualifiers` + `QualifiersPointsEarned`

## Phase 2: Backend Handler + Route

- [ ] 2.1 Add `UserService *services.UserService` field to `QualifierPredictionHandler` in `backend/internal/infrastructure/handlers/handlers.go`
- [ ] 2.2 Add `GetByUser` method: parse `{id}`, 400 on parse error, 404 if `UserService.GetByID` fails, otherwise call service + `json.Encode`
- [ ] 2.3 Wire `UserService` into `qualifierHandler` in `backend/cmd/api/main.go`
- [ ] 2.4 Add route `GET /users/{id}/qualifier-predictions` + `OPTIONS` CORS preflight in `main.go`

## Phase 3: Frontend Hook

- [ ] 3.1 Add `useQualifierPredictionsByUser(userId)` in `frontend/src/features/hooks.ts` (GET, no autosave, AbortController cleanup, no `useEffect` deps on `userId` array)

## Phase 4: Frontend Card Component

- [ ] 4.1 Create `frontend/src/components/QualifierPredictionCard.tsx`
- [ ] 4.2 Render 1° and 2° picks with flags (always)
- [ ] 4.3 If `view.group_closed`, render actual 1°/2° with flags + green check / red X per position + `points_earned` badge
- [ ] 4.4 If `predicted_first` is empty, show muted "Sin pronóstico de clasificación" placeholder

## Phase 5: Frontend Page Wiring

- [ ] 5.1 In `frontend/src/pages/UserPredictionsPage.tsx`: import hook + card
- [ ] 5.2 Call `useQualifierPredictionsByUser(Number(id))`; build a `Map<group, view>` keyed by `group_name`
- [ ] 5.3 Render `<QualifierPredictionCard>` inside the active group tab, above the matches table
- [ ] 5.4 Keep the existing "Sin pronósticos" guard for empty `availableGroups` intact

## Phase 6: Verify

- [ ] 6.1 `cd backend && go build ./...` — no JSON tag regressions
- [ ] 6.2 `curl.exe http://localhost:8080/users/1/qualifier-predictions` — assert array shape, flags populated, `group_closed: true` for group A in local DB
- [ ] 6.3 `curl.exe http://localhost:8080/users/9999/qualifier-predictions` — assert 404
- [ ] 6.4 Open `/user/1/predictions` in browser, switch to group A tab — assert card renders with flags + actuals + checkmarks + `3` points badge
- [ ] 6.5 Switch to group B tab — assert card shows user picks only, no actuals, no checkmarks
- [ ] 6.6 Open `/user/3/predictions` (vacio) — assert "Sin pronóstico de clasificación" placeholder renders, no crash

## Phase 7: Push

- [ ] 7.1 `git add` all changes (backend + frontend + change folder)
- [ ] 7.2 Commit: `feat(quiniela): display user qualifier predictions on profile page` (conventional, no AI attribution)
- [ ] 7.3 `git push origin feature/user-qualifier-predictions-display` (NOT master)
- [ ] 7.4 Tell user the branch is ready to review / merge
