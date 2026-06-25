# Tasks: Ranking Tie-Breaker for Users

## 1. Backend domain

- [x] 1.1 In `backend/internal/core/domain/user.go`, add `ExactScore int` and `WinnerScore int` fields with `json:"exact_score"` and `json:"winner_score"`
- [x] 1.2 Confirm no other consumer of the `User` struct breaks (grep for `domain.User` struct literal usages)

## 2. Backend handler

- [x] 2.1 In `backend/internal/infrastructure/handlers/handlers.go::GetRanking`, declare `exactScore` and `winnerScore` counters
- [x] 2.2 Inside the existing prediction loop, increment `exactScore` when `pts == 3` and `winnerScore` when `pts == 2`
- [x] 2.3 Assign `users[i].ExactScore = exactScore` and `users[i].WinnerScore = winnerScore` after the loop

## 3. Frontend sort

- [x] 3.1 In `frontend/src/pages/RankingPage.tsx`, replace the single-criterion `b.total - a.total` sort with a multi-criteria comparator
- [x] 3.2 Tie-breaker order: `total` → `exact_score` → `winner_score` → `id`

## 4. Tests

- [x] 4.1 New `backend/internal/core/domain/ranking_test.go` covering the sort criteria
  - [x] 4.1.1 Higher total wins
  - [x] 4.1.2 Equal total — more exact scores wins
  - [x] 4.1.3 Equal total and exact — more winner scores wins
  - [x] 4.1.4 Full tie — lower id wins
- [ ] 4.2 Run `go test ./...` and confirm all pass

## 5. Spec

- [x] 5.1 Create this change folder `openspec/changes/ranking-tie-breaker/`
- [x] 5.2 Write `proposal.md` (Intent, Scope, Approach, Risks, Rollback, Success)
- [x] 5.3 Write `specs/ranking-tie-breaker/spec.md` with requirements and scenarios
- [x] 5.4 Write this `tasks.md`

## 6. Local validation

- [ ] 6.1 Run backend locally with the change
- [ ] 6.2 `curl http://localhost:8080/ranking` and confirm the new `exact_score` and `winner_score` fields appear
- [ ] 6.3 Build the frontend and load the ranking page; confirm 6 columns, same UI, sort order changes only for ties
- [ ] 6.4 Simulate a tie scenario with two users having the same total but different exact scores (use admin endpoint to set results on finished matches)

## 7. Archive

- [ ] 7.1 After all tests pass and local validation is done, archive the change (move folder to `archive/2026-06-XX-ranking-tie-breaker/`)
- [ ] 7.2 Sync the delta into a new canonical `openspec/specs/ranking-tie-breaker/spec.md`
