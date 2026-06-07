# Proposal: Display User Qualifier Predictions (1°/2° per group)

## Intent

The `/user/:id/predictions` page shows per-match score predictions but NOT the 1°/2° qualifier picks per group. Users already save those picks on the home page (see `group-qualifier-prediction` change), but there is no public view of them. We need to surface the picks — and, when the group is closed, show the actual qualifiers with a visual hint of which positions were guessed correctly.

## Scope

### In Scope
- Backend: new endpoint `GET /users/{id}/qualifier-predictions` returning the array of `QualifierPrediction` for that user (with team flags).
- Frontend: on `UserPredictionsPage`, render a compact card per group showing the user's 1°/2° picks.
- If the group is closed: also render the actual qualifiers and a green check / red X per correct position. 3 pts each, max 6.
- New `useQualifierPredictionsByUser(userId)` hook in `features/hooks.ts` (mirror the existing `useQualifierPrediction` pattern).
- 404 for nonexistent user; 200 with `[]` for a user with no qualifier predictions yet.

### Out of Scope
- Changing the save/upsert flow or scoring logic.
- Showing qualifier predictions on the public ranking page (already shows `qualifier_score`).
- Revealing the actual qualifiers while the group is still in progress.
- Locking the user's own view (this page is read-only — the home page remains the edit surface).

## Capabilities

### New Capabilities
- `user-qualifier-predictions`: read-only view of a user's group qualifier predictions, including visual feedback against the actual qualifiers once the group is closed.

### Modified Capabilities
- None (no requirement of any existing spec changes; this is additive).

## Approach

**Backend**: add a new handler method on `QualifierPredictionHandler` called `GetByUser` (reads `{id}` from path, 404 if user not found, calls existing `repo.GetAllQualifierPredictionsByUser` + enriches with team flags via a small SQL join or a map lookup against the teams table). Wire `GET /users/{id}/qualifier-predictions` in `main.go`.

**Frontend**: add `useQualifierPredictionsByUser(userId)` hook (GET, no autosave — read-only). On `UserPredictionsPage`, group qualifier predictions by `group_name`, render a card above the matches table inside the active-group tab. The card shows: 1° pick with flag, 2° pick with flag. If the user has no prediction for that group, show "Sin pronóstico de clasificación". If the group is closed (derived from `predictions[].status === 'finished'` for all 6 matches), compute the actual qualifiers client-side with the same FIFA rules (mirror the backend `CalculateQualifiers` to keep the public view consistent — or, simpler, ship the actual qualifiers from the server in the same response).

**Decision (deferred to design)**: ship actual qualifiers from the server (single round trip, single source of truth, no rule duplication in the frontend).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | New `GetByUser` method on `QualifierPredictionHandler` |
| `backend/cmd/api/main.go` | Modified | New route `GET /users/{id}/qualifier-predictions` |
| `frontend/src/pages/UserPredictionsPage.tsx` | Modified | Fetch + render qualifier card per group |
| `frontend/src/features/hooks.ts` | Modified | New `useQualifierPredictionsByUser` hook |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Duplicate JSON tag regression (see commit `2b7ec86`) | Low | Read full struct blocks before editing; verify response keys after build. |
| User has 0 predictions → empty array breaks UI | Low | Render "Sin pronóstico" placeholder |
| Group has 0 finished matches but we expose actuals | Low | Server only includes `actual_first`/`actual_second` when the group is closed; frontend only shows the comparison block when both are present. |
| Endpoint leaks data across users | None | Public read of qualifier picks is intentional (same as match predictions are public on this page). |

## Rollback Plan

Revert the single commit on the feature branch. No schema changes, no data migration. Endpoint is additive — the rest of the app is unaffected.

## Dependencies

- `group-qualifier-prediction` change (already on master as commit `56c2036`): provides `QualifierPrediction` domain, repository, and the `group_qualifier_predictions` table.
- `UserPredictionsPage` already loads `/users/{id}/predictions`; we add a parallel fetch.

## Success Criteria

- [ ] `GET /users/{id}/qualifier-predictions` returns the array (with team flags) for valid users, 404 for missing users, `[]` for users with no predictions.
- [ ] `/user/:id/predictions` renders a 1°/2° card for every group the user picked.
- [ ] When the group is closed, the card also shows the actual 1°/2° and a green check / red X per correct position.
- [ ] `go build ./...` passes with no JSON tag regressions.
- [ ] Smoke test: hit the new endpoint, confirm shape, confirm frontend renders for at least dangel (who has a group A pick in the local DB).
