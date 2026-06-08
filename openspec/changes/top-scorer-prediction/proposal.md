# Proposal: Top Scorer of the Group Stage Prediction

## Intent

Add a new pre-tournament prediction: each user picks one player from a curated list of ~50 candidates as their guess for **top scorer of the group stage**. The user gets 6 points if their pick matches the actual top scorer when the group stage closes. Ties (multiple players sharing the top goal count) all count as correct. Predictions close 1 hour before the World Cup opener (same global deadline as the 1st/2nd qualifier predictions). The admin sets the actual top scorer after the last group match.

## Scope

**In scope:**
- Backend: new `top_scorer_predictions` table, domain logic, service, handlers, lock enforcement, scoring
- Frontend: new page `/goleador` with a dropdown of candidates
- Frontend: new section in `UserPredictionsPage` (above the qualifiers) showing the rival's top-scorer pick
- Frontend: new section in `AdminPage` to set the actual top scorer
- Frontend: new column `top_scorer_score` in `RankingPage`
- A static JSON list of ~50 candidates loaded at server startup
- Migration: add new table; no changes to existing data

**Out of scope:**
- Group stage goal statistics tracking (we don't count goals per player; the admin simply enters the winner's name)
- Dynamic candidate management (admin cannot add/remove candidates — list is static)
- Per-user multiple top-scorer predictions (one per user, changeable until deadline)
- Notifications/emails when the user hasn't picked yet

## Capabilities

### New: top-scorer-prediction

- **CRUD**: each authenticated user can save / change their top-scorer pick (one per user)
- **Lock**: deadline = `QUALIFIER_LOCK_AT` (reused, same instant as 1st/2nd qualifier close: `2026-06-11T18:00:00Z`)
- **Visibility**: any logged-in user can read any other user's pick (same as match/qualifier predictions)
- **Candidates**: static list served by `GET /config` (extended) or a new `GET /top-scorer/candidates`
- **Scoring**: 6 points if `predicted_player == admin_set_top_scorer` (case-insensitive, exact match on the candidate's `name` field); 0 otherwise
- **Tie handling**: if the admin sets a player whose name is one of several tied top scorers (case "shared top count"), users who picked any of the tied names get 6 points (admin is responsible for listing all tied names; we match the single admin-set value, no multi-select in v1)
- **Admin**: `POST /admin/top-scorer` sets the actual top scorer (free-text name, validated against the candidate list with a warning if not present)
- **Ranking**: new column `top_scorer_score` in `/ranking`, sums to total alongside `score` and `qualifier_score`

## Approach

Reuse the existing hexagonal pattern from `group-qualifier-prediction`:

1. **DB**: new `top_scorer_predictions` table (one row per user) + a key-value `app_config` table (or extend `ConfigHandler`) for the admin-set actual top scorer
2. **Domain**: new `TopScorerCandidate`, `TopScorerPrediction`, `TopScorerResult` structs; new `TopScorerPointsEarned(predicted, actual string) int` function (mirror of `QualifiersPointsEarned`)
3. **Service**: new `TopScorerService` with `GetMine`, `Upsert`, `GetByUser`, `SetActualResult` (admin), `ScoreUser`
4. **Handlers**: 4 endpoints (mine, public read, admin set, candidates read) + extend `ConfigHandler` to include the candidate list
5. **Lock check**: reuse the same `qualifierLockAt` injected in `main.go` — pass it to the new service constructor
6. **Frontend**: new `<TopScorerPage>` page + new `<TopScorerPredictionCard>` component for the user-profile view + new section in `<AdminPage>` for setting the result + new column in `<RankingPage>`
7. **Candidates**: JSON file `backend/data/top_scorer_candidates.json` loaded at server startup (read once into memory, exposed via `/config`)

## Affected Areas

- `backend/internal/core/domain/` — new file `top_scorer.go`
- `backend/internal/core/ports/repositories.go` — extend interfaces
- `backend/internal/application/services/services.go` — new `TopScorerService`
- `backend/internal/infrastructure/handlers/handlers.go` — new `TopScorerHandler` + extend `ConfigHandler`
- `backend/cmd/api/main.go` — wire service, load candidates JSON, register routes
- `backend/internal/infrastructure/repository/sqlite_repository.go` — new table DDL + methods
- `backend/data/top_scorer_candidates.json` — new static data file
- `frontend/src/pages/GoleadorPage.tsx` — new page
- `frontend/src/components/TopScorerPredictionCard.tsx` — new component
- `frontend/src/pages/UserPredictionsPage.tsx` — add card above qualifiers
- `frontend/src/pages/AdminPage.tsx` — add section to set result
- `frontend/src/pages/RankingPage.tsx` — add `top_scorer_score` column
- `frontend/src/features/hooks.ts` — new `useTopScorer` hook
- `frontend/src/App.tsx` (or router) — register new route
- `frontend/src/components/NavBar.tsx` (or similar) — link to new page

## Risks

- **Lock deadline alignment**: reusing `QUALIFIER_LOCK_AT` keeps the deadline consistent but couples two features. Mitigation: a docstring on the new service and the proposal explicitly notes the dependency.
- **Candidate list churn**: if the user wants to add/remove players after deploy, they must edit the JSON and redeploy. Mitigation: documented in the file's purpose comment.
- **Admin sets an actual scorer not in the candidate list**: edge case. The system stores whatever the admin submits; users who picked a matching name (case-insensitive) get 6 points. If the admin submits "Mbappé" but a user picked "Kylian Mbappé" (full name), they don't match. Mitigation: v1 uses exact-string match; future iteration can add aliases.
- **Frontend dropdown size**: 50 items is fine for a native `<select>`. No need for a custom searchable combobox.
- **No tests**: project has 0 unit tests. We follow the established pattern (smoke-test only) per `config.yaml: testing.backend.runner: none` and `strict_tdd: false`.

## Rollback Plan

- Single branch `feature/top-scorer-prediction` not merged = trivial revert
- If merged: the new endpoints and DB table are additive. The frontend route can be removed; the DB table can stay without breaking the app (no other code reads it). `DELETE FROM top_scorer_predictions` to clear data.
- Reverting the merge via `git revert <merge-sha>` is safe.

## Success Criteria

- [ ] A user can navigate to `/goleador`, see the candidate list, pick one, and save it
- [ ] After `2026-06-11T18:00:00Z` (or with `QUALIFIER_LOCK_AT` overridden), the PUT returns 403
- [ ] The user's pick appears in their own profile and in any rival's profile view
- [ ] Admin can set the actual top scorer via the admin panel
- [ ] Once the admin sets it, users who picked that player show 6 points in `/ranking` (in the new `top_scorer_score` column)
- [ ] Users who picked a different player (or no one) show 0 in that column
- [ ] Smoke test: create 3 test users, each picks a different player, admin sets one of them as actual, verify only that user gets 6
- [ ] All existing functionality (match predictions, 1st/2nd qualifiers, auto-lock) still works (no regressions)
