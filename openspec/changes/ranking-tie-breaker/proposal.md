# Proposal: Ranking Tie-Breaker for Users

## Intent

The current ranking sorts users only by total points (`score + qualifier_score + top_scorer_score`). When two users tie on total, the order falls back to whatever `GetAllUsers()` returns (effectively user `id` ascending). This makes the ranking feel arbitrary at the top of the table. This change introduces a deterministic, merit-based tie-breaker that rewards precision: more exact-score predictions beat fewer, and more winner-correct predictions beat fewer. Final fallback is the user who registered first (lower `id`).

The change is invisible to non-tied users and to the UI's column layout — only the row order changes when totals are equal.

## Scope

### In Scope
- Backend: add `ExactScore` (count of 3-point predictions) and `WinnerScore` (count of 2-point predictions) to the `User` struct, computed in `GetRanking` handler.
- Frontend: extend the sort comparator in `RankingPage.tsx` with two new criteria after `total`, and a stable `id`-ascending fallback.
- New `openspec` capability `ranking-tie-breaker` documenting the rule and the criteria order.
- Tests: new domain test covering the sort criterion order (regression for future changes).

### Out of Scope
- Showing `exact_score` and `winner_score` as visible columns in the ranking table (separate feature).
- Changing the points awarded for each prediction type (still 3 / 2 / 0).
- Changing `qualifier_score` or `top_scorer_score` weights.
- Tie-breakers inside the FIFA group table (already implemented in `domain/qualifier.go`).
- A new `created_at` column on `users` (out of scope; `id` is a sufficient proxy since SQLite autoincrement reflects insertion order).

## Capabilities

### New Capabilities
- `ranking-tie-breaker`: defines the deterministic ordering used by `GET /ranking` when two or more users have the same total points. The order is: total (desc) → exact_score (desc) → winner_score (desc) → id (asc).

### Modified Capabilities
- None.

## Approach

Single source of truth in the backend: the handler that already computes `Score` for each user also counts how many of their predictions earned 3 points (`exact_score`) and how many earned 2 points (`winner_score`). These are returned as new JSON fields on the existing `User` struct. The frontend's existing sort is extended with two extra criteria; no new request, no N+1. The fallback `id` ordering is deterministic and stable across renders.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/core/domain/user.go` | Modified | Add `ExactScore` and `WinnerScore` fields |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | Compute the two counters in `GetRanking` loop |
| `frontend/src/pages/RankingPage.tsx` | Modified | Extend the sort comparator with the two new criteria |
| `backend/internal/core/domain/ranking_test.go` | New | Unit test for the tie-breaker order |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Backend computes counters on every `/ranking` call (already O(n*m)) | Low | Same loop that computes `Score`; no new DB queries; same complexity |
| New fields break a frontend consumer that expects a strict shape | Low | Only additive JSON fields; existing columns are untouched |
| Sort behavior surprises a user who was winning by id-order | Low | The new order rewards accuracy; communicated via the new spec capability |
| `id` as a proxy for "registered first" fails if IDs are reseeded | None | SQLite autoincrement is monotonic within a DB; production uses Turso with the same guarantee |

## Rollback Plan

Revert the merge commit. The `User` struct gets two new fields, the handler two new lines, the sort two new lines. None of this changes the DB schema, so the rollback is purely code.

## Success Criteria

- [ ] `GET /ranking` returns `exact_score` and `winner_score` for every user
- [ ] When two users have the same total, the one with more exact scores is listed first
- [ ] When total and exact scores are equal, the one with more winner scores is listed first
- [ ] When all three are equal, the user with the lower `id` is listed first
- [ ] `go test ./...` passes (existing 20 cases + new ranking tie-breaker cases)
- [ ] Ranking page renders unchanged: same 6 columns, same data
