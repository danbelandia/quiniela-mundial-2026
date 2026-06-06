# Design: Auto-Lock Matches 3h Before Kickoff

## Technical Approach

Time-based effective lock computed in two places for two different consumers: at **read time in the repo** (so the frontend transparently sees the right `is_locked`), and at **decision time in the service** (so `POST /predictions` and qualifier `Upsert` are TOCTOU-safe). The repo overrides `IsLocked` after every `Scan`. The service re-evaluates with the current clock at the moment of the write. A single `domain.Match.IsEffectivelyLocked(now)` helper holds the rule; the lock window is a `time.Duration` injected into the repo at construction.

## Architecture Decisions

| Decision | Choice | Alternatives | Why |
|---|---|---|---|
| Compute effective lock in repo AND service | Both: repo at read, service at decision | Repo only / service only / handler | Repo gives frontend correct data; service guarantees safety regardless of time elapsed between read and write |
| Inject `now` into repo | `clock func() time.Time` field, default `time.Now` | Global / package var / interface | Mockable in future tests; no global state; no extra interface needed yet |
| Inject `lockWindow` | Constructor param on repo | Package var / env-read inside repo | Matches existing pattern (`driverName`, `connStr` already ctor-injected) |
| Service-level error | `var ErrMatchLocked = errors.New("match is locked")` | bool return / string sentinel | Idiomatic Go; `errors.Is` in handler |
| Qualifier group lock rule | Reject if **any** match in group is effectively locked | First match in window / all matches locked | Mirrors existing frontend `groupIsLocked = filteredMatches.some(m => m.is_locked)` (HomePage.tsx:41-44) |
| Frontend changes | None | Various | Frontend already reads `is_locked` and disables inputs |
| Tests | Only `domain/match_test.go` (5 cases) | Full integration | Project has 0 tests; start with the highest-value unit |
| Migration | None | Schema change | `is_locked` column already exists |
| Admin override | Preserved | Remove endpoints | User explicit requirement: keep manual lock for edge cases |
| Env var fallback | Invalid value → log warning + default 3h | Hard fail | Matches the spec's "graceful degradation" scenario |

## Data Flow

```
[Client]                          [main.go]                       [repo]                          [service / handler]
   │                                   │                              │                                       │
   │ GET /matches                      │                              │                                       │
   ├──────────────────────────────────►│ ctor: lockWindow=3h          │                                       │
   │                                   ├─► NewSQLiteRepository        │                                       │
   │                                   │                              │                                       │
   │                                   │   GetAllMatches()            │                                       │
   │                                   ├─────────────────────────────►│                                       │
   │                                   │                              │ SELECT id, ..., is_locked FROM matches   │
   │                                   │                              │ for each row:                          │
   │                                   │                              │   m.IsLocked = m.IsEffectivelyLocked(   │
   │                                   │                              │     clock().UTC(), lockWindow)          │
   │                                   │   JSON marshal               │                                       │
   │◄──────────────────────────────────┤◄────────────────────────────┤                                       │
   {"is_locked": true/false}            │                              │                                       │
   │                                   │                              │                                       │
   │ POST /predictions                 │                              │                                       │
   │ { match_id: 1, ... }              │                              │                                       │
   ├──────────────────────────────────►├─ handler.CreatePrediction    │                                       │
   │                                   │                              │                                       │
   │                                   ├─ service.PlacePrediction ──────────────────────────────────────────────►│
   │                                   │                              │                                       │
   │                                   │                              │   match, _ := matchRepo.GetByID(1)       │
   │                                   │                              │                                       │
   │                                   │                              │   if match.IsEffectivelyLocked(         │
   │                                   │                              │       time.Now().UTC(),                 │
   │                                   │                              │       lockWindow)                      │
   │                                   │                              │     return ErrMatchLocked               │
   │                                   │                              │                                       │
   │                                   │                              │   repo.Save(p)                         │
   │                                   │                              │                                       │
   │   handler maps ErrMatchLocked ──────────────────────────────────────────────────────────────────────►│
   │◄──────────────────────────────────┤                              │                                       │
   201 Created  /  403 Forbidden        │                              │                                       │
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/domain/match.go` | Modify | Add `IsEffectivelyLocked(now time.Time) bool` method; reads `lockWindow` from a package-level `var LockWindow = 3*time.Hour` set at init, OR receives it as a 2nd param. **Decision**: 2nd param to keep it pure and testable |
| `backend/internal/core/domain/match_test.go` | Create | First test file. Table-driven, 5 cases. Uses fixed `time.Time` for `now` and `matchDate` |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modify | Add `lockWindow time.Duration` + `clock func() time.Time` fields; constructor `NewSQLiteRepository(driver, conn, lockWindow)`; init `clock = time.Now` if nil; override `IsLocked` in 3 Scan loops (lines ~161, ~175, ~253) right after Scan |
| `backend/internal/application/services/services.go` | Modify | Add `var ErrMatchLocked = errors.New("match is locked")`; `PredictionService` needs `matchRepo ports.MatchRepository` (add to struct + ctor); `PlacePrediction` fetches match, checks lock, then saves; `QualifierPredictionService.Upsert` fetches all group matches, checks if any locked |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | `PredictionHandler.CreatePrediction` maps `ErrMatchLocked` → 403; `QualifierPredictionHandler.Upsert` does the same |
| `backend/cmd/api/main.go` | Modify | `buildAppConfig()` reads `LOCK_WINDOW_HOURS` (default 3.0, warning on invalid); pass `lockWindow` to `NewSQLiteRepository` |
| `backend/.env.example` | Create | Document `LOCK_WINDOW_HOURS=3` |
| `openspec/specs/auto-lock-matches/spec.md` | Create (on archive) | Synced from `openspec/changes/auto-lock-matches/specs/auto-lock-matches/spec.md` |

## Interfaces / Contracts

```go
// backend/internal/core/domain/match.go
func (m Match) IsEffectivelyLocked(now time.Time, window time.Duration) bool {
    if m.IsLocked {
        return true
    }
    return !now.Add(window).Before(m.Date)  // window reached or passed
}

// backend/internal/application/services/services.go
var ErrMatchLocked = errors.New("match is locked")

// backend/internal/infrastructure/repository/sqlite_repository.go
type SQLiteRepository struct {
    DB         *sql.DB
    lockWindow time.Duration
    clock      func() time.Time
}

func NewSQLiteRepository(driverName, connStr string, lockWindow time.Duration) (*SQLiteRepository, error)
```

The helper's contract: `IsLocked` is a manual override and always wins. The window check is `now + window >= matchDate` (boundary inclusive, matching the spec's "exactly at the boundary → locked" scenario). Times are UTC.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `domain.Match.IsEffectivelyLocked` — 5 cases: manual override, in-window, out-of-window, boundary, post-kickoff | `go test ./internal/core/domain/...`. Table-driven. ~30 lines. No mocks |
| E2E manual | Full server-side enforcement | After deploy: with `LOCK_WINDOW_HOURS=0.01`, manipulate one match's date to `now + 30s`, hit `POST /predictions` via curl, expect 403 |
| E2E manual | Admin override still works | With a match 24h out, call `POST /admin/matches/lock {id, locked: true}`, expect next `GET /matches` to show `is_locked: true` |

No integration tests yet — the project has 0 tests and no test infra. The unit test is the seed; full integration coverage is a follow-up scope.

## Migration / Rollout

No migration. The `is_locked` column already exists in `matches` (sqlite_repository.go:54). The change is purely additive code.

Rollout: standard `master` merge → Render auto-deploys. Feature flag not needed — the lock window is configurable via env, and the default of 3h matches the spec. To disable in an emergency, set `LOCK_WINDOW_HOURS=` to empty (falls back to 3h, but the manual `is_locked` column can also be flipped off via admin). Worst case: revert the merge.

## Open Questions

None blocking. The qualifier group rule (any-match-locked → group-locked) is the defensible default that matches existing frontend semantics. If we later need "lock when first match of group is in window", it's a one-line change in `QualifierPredictionService.Upsert`.
