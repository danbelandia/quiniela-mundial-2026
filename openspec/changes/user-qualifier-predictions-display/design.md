# Design: Display User Qualifier Predictions (1°/2° per group)

## Technical Approach

Add a new read-only endpoint `GET /users/{id}/qualifier-predictions` that returns the user's 1°/2° picks per group, enriched with team flags and (when the group is closed) the actual qualifiers + earned points. Render the result as a card inside the active group tab of `UserPredictionsPage`, above the matches table. All FIFA tie-breaker logic stays on the server (single source of truth, no rule duplication on the frontend).

## Architecture Decisions

### Decision: Response DTO lives in `ports/`, not in `domain/`

**Choice**: Define `QualifierPredictionView` in `backend/internal/core/ports/repositories.go` next to the repository interface. The `domain.QualifierPrediction` stays as the persistence shape.
**Alternatives considered**: Reusing `domain.QualifierPrediction` and adding optional fields; creating a handler-local DTO.
**Rationale**: `QualifierPredictionView` is a public read contract (flags, actuals, points). Putting it in `ports/` keeps it discoverable next to the interface that produces it and separate from the pure domain entity.

### Decision: Service, not handler, does the enrichment

**Choice**: Add `QualifierPredictionService.GetViewsForUser(userID)` that does flag lookup + group-closed detection + actual-qualifier computation.
**Alternatives considered**: Enriching in the handler.
**Rationale**: The handler stays thin and serializable; business rules (when to compute actuals, how to map team name → flag) belong in the service layer. Mirrors the pattern used by `RankingHandler` which delegates to `RankingService` + `QualifierService`.

### Decision: User existence check via injected `UserService`

**Choice**: `QualifierPredictionHandler` gains a `UserService` field for the 404 check.
**Alternatives considered**: Adding `UserRepository` to the qualifier service.
**Rationale**: `UserService` already exists, is already constructed in `main.go`, and the same handler can reuse it without a new dependency.

### Decision: All matches fetched once per request

**Choice**: `GetViewsForUser` calls `matchRepo.GetAllMatches()` once, builds `map[group]map[team]flag` + per-group match list, then enriches.
**Alternatives considered**: One SQL query per qualifier prediction row.
**Rationale**: 12 groups × ≤6 matches = ≤72 rows. One round trip beats N+1; no caching needed for a public endpoint with light traffic.

### Decision: Frontend gets actual qualifiers from the server, not from client-side recomputation

**Choice**: The `QualifierPredictionView` includes `actual_first`/`actual_first_flag`/`actual_second`/`actual_second_flag`/`points_earned`/`group_closed` fields.
**Alternatives considered**: Calling the public `/standings` endpoint from the frontend and ranking in JS.
**Rationale**: Same source of truth, no risk of frontend rule drift, no second request, no JS dependency on `domain.qualifier`.

## Data Flow

```
Client GET /users/1/predictions   ──┐
Client GET /users/1/qualifier-     ─┤  UserPredictionsPage
        predictions                │  ├─ apiClient.get(predictions) → match table
                                   │  └─ useQualifierPredictionsByUser(id) → card
                                   │
Server  QualifierPredictionHandler.GetByUser
            │
            ├─ UserService.GetByID(id)        → 404 if missing
            ├─ QualifierPredictionService.GetViewsForUser(id)
            │     ├─ MatchRepo.GetAllMatches()            (1 query)
            │     ├─ QualifierPredictionRepository.GetAllQualifierPredictionsByUser(id)
            │     └─ for each pick:
            │           lookup flags (matches map)
            │           if all 6 group matches finished:
            │                domain.CalculateQualifiers(matches, group)
            │                domain.QualifiersPointsEarned(...)
            └─ json.Encode(views)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/ports/repositories.go` | Modify | Add `QualifierPredictionView` struct + `GetViewsForUser` to `QualifierPredictionService` consumer |
| `backend/internal/application/services/services.go` | Modify | Implement `QualifierPredictionService.GetViewsForUser` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | New `QualifierPredictionHandler.GetByUser` + add `UserService` field |
| `backend/cmd/api/main.go` | Modify | Wire `UserService` into `qualifierHandler` + new route `GET /users/{id}/qualifier-predictions` |
| `frontend/src/features/hooks.ts` | Modify | New `useQualifierPredictionsByUser(userId)` hook |
| `frontend/src/components/QualifierPredictionCard.tsx` | Create | Read-only card showing 1°/2° picks + actuals + correctness hints |
| `frontend/src/pages/UserPredictionsPage.tsx` | Modify | Fetch + render the card above the matches table inside the active group tab |

## Interfaces / Contracts

```go
// ports/repositories.go
type QualifierPredictionView struct {
    GroupName          string `json:"group_name"`
    PredictedFirst     string `json:"predicted_first"`
    PredictedFirstFlag string `json:"predicted_first_flag"`
    PredictedSecond    string `json:"predicted_second"`
    PredictedSecondFlag string `json:"predicted_second_flag"`
    ActualFirst        string `json:"actual_first"`
    ActualFirstFlag    string `json:"actual_first_flag"`
    ActualSecond       string `json:"actual_second"`
    ActualSecondFlag   string `json:"actual_second_flag"`
    PointsEarned       int    `json:"points_earned"`
    GroupClosed        bool   `json:"group_closed"`
}
```

```ts
// features/hooks.ts
export function useQualifierPredictionsByUser(userId: number): {
  views: QualifierPredictionView[];
  loading: boolean;
  error: string | null;
};
```

```ts
interface QualifierPredictionView {
  group_name: string;
  predicted_first: string;
  predicted_first_flag: string;
  predicted_second: string;
  predicted_second_flag: string;
  actual_first: string;
  actual_first_flag: string;
  actual_second: string;
  actual_second_flag: string;
  points_earned: number;
  group_closed: boolean;
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Manual / smoke | `GET /users/1/qualifier-predictions` returns enriched array | `curl.exe` after build; assert `predicted_first_flag`, `group_closed` |
| Manual / UI | `/user/1/predictions` renders the card for group A | Open in browser, screenshot, verify flags + actuals + checkmarks |
| Regression | `go build ./...` | Catch duplicate JSON tag regressions like commit `2b7ec86` |
| Regression | `npx tsc --noEmit` (or build) | Catch TS type mismatches in the new hook / component |

## Migration / Rollout

No migration required. The endpoint is additive; the frontend change is additive; no data shape changes to existing endpoints.

## Open Questions

None.
