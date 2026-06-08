# Design: Top Scorer of the Group Stage Prediction

## Technical Approach

Mirror the existing `group-qualifier-prediction` pattern in full: a new `top_scorer_predictions` table (one row per user, UPSERT on save), a new `TopScorerService` in the application layer, four HTTP handlers, and a single-row `app_config` table to store the admin-set actual top scorer. The static candidate list is loaded once at server startup from `backend/data/top_scorer_candidates.json` into a Go slice held in memory; the same slice is served via the extended `/config` endpoint (no new endpoint needed). The lock check reuses the existing `qualifierLockAt` injected into the new service at construction — no new env var, no new wiring.

## Architecture Decisions

| Decision | Choice | Alternatives | Why |
|---|---|---|---|
| Candidate list source | Static JSON file loaded at startup | DB table seeded at startup / hardcoded Go slice / fetched from external API | User provided the list explicitly; JSON is the natural format; no migration needed; in-memory slice is fast and read-only |
| Service exposes candidates | Extend `ConfigHandler` to include `top_scorer_candidates` array | New `/top-scorer/candidates` endpoint | Single source of truth (config), one fewer route, already cached client-side |
| Lock deadline | Reuse `qualifierLockAt` (same as 1st/2nd qualifier) | New `TOP_SCORER_LOCK_AT` env var | Both predictions close at the same logical moment (1h before opener); user confirmed reuse; one less config knob |
| One prediction per user | UNIQUE(user_id) on `top_scorer_predictions` | Multiple rows per user (history) | User picks "one player"; no history needed |
| UPSERT semantics | `INSERT OR REPLACE` (or `ON CONFLICT DO UPDATE`) | DELETE + INSERT | Preserves `created_at`, updates `updated_at`; matches `predictions` table pattern |
| Actual top scorer storage | Single-row `app_config` table (key-value) | New `top_scorer_result` table | Only one value exists; key-value is the simplest representation; reusable for future single-value config (e.g., next-year winner) |
| Admin input validation | Warn on names not in candidate list, still accept | Hard reject | Defensive: admin must be able to record a dark-horse winner (no one in the list) |
| Name match (predicted vs actual) | Case-insensitive + accent-insensitive exact match | Fuzzy match / Levenshtein | v1: simple and predictable. If user wrote "Mbappé" (with accent) and admin wrote "Mbappe" (without), it should match. If user wrote "Kylian Mbappé" and admin wrote "Mbappé", it should NOT match (admin should use the candidate's exact `name` field) |
| Frontend dropdown | Native `<select>` | Custom searchable combobox | 50 items is small; native is accessible and trivial |
| Card position on rival profile | Above qualifiers | Below matches / in a separate tab | User requirement: "arriba con los qualifiers" |
| Ranking column position | Between "Clasificación" and "Total" | At the end / before "Clasificación" | Symmetry: the 3 non-match score types (qualifier, top scorer, total) cluster together |
| Test infra | Smoke test only, no unit tests | Add Go test | Project has 0 tests; `config.yaml: testing.backend.runner: none` and `strict_tdd: false`; matches established pattern |

## Data Flow

```
[Server startup]
  main.go reads data/top_scorer_candidates.json
  → loads into configHandler.Candidates []TopScorerCandidate
  → exposes via GET /config.top_scorer_candidates

[User saves pick]
  Client                         main.go                       service / handler
    │                                │                                  │
    │ PUT /top-scorer-prediction/me  │                                  │
    │ {user_id, predicted_player}    │                                  │
    ├───────────────────────────────►│ handler.TopScorer.Upsert ────────►│
    │                                │                                  │
    │                                │ if !time.Now().UTC().Before(      │
    │                                │     s.qualifierLockAt) {         │
    │                                │   return ErrMatchLocked           │
    │                                │ }                                │
    │                                │                                  │
    │                                │ if !playerInCandidates(player) {  │
    │                                │   return ErrInvalidPlayer         │
    │                                │ }                                │
    │                                │                                  │
    │                                │ repo.UpsertTopScorerPrediction()  │
    │                                │                                  │
    │ 200 OK / 403 / 400             │                                  │
    │◄───────────────────────────────┤◄─────────────────────────────────┤

[Admin sets actual]
  Client                         main.go                       service / handler
    │                                │                                  │
    │ POST /admin/top-scorer         │                                  │
    │ {player: "Kylian Mbappé"}      │                                  │
    ├───────────────────────────────►│ handler.TopScorerAdmin.SetActual ►│
    │                                │                                  │
    │                                │ if !playerInCandidates(player) {  │
    │                                │   log.Printf("warn: top scorer    │
    │                                │     not in candidate list: %s",   │
    │                                │     player)                       │
    │                                │ }                                │
    │                                │                                  │
    │                                │ repo.SetAppConfig("top_scorer",   │
    │                                │                   player)        │
    │                                │                                  │
    │ 200 OK                         │                                  │
    │◄───────────────────────────────┤◄─────────────────────────────────┤

[GET /ranking]
  Client                         main.go                       service / handler
    │                                │                                  │
    │ GET /ranking                   │                                  │
    ├───────────────────────────────►│ RankingHandler.Get ──────────────►│
    │                                │                                  │
    │                                │ for each user:                   │
    │                                │   score = sum(predictions)        │
    │                                │   qualifier_score = ScoreUser(    │
    │                                │                     user.ID)      │
    │                                │   actual := repo.GetAppConfig(    │
    │                                │                "top_scorer")      │
    │                                │   predicted := repo.Get(          │
    │                                │        user.ID)                   │
    │                                │   if actual != "" &&              │
    │                                │      normalize(predicted) ==      │
    │                                │      normalize(actual) {          │
    │                                │     top_scorer_score = 6          │
    │                                │   } else {                       │
    │                                │     top_scorer_score = 0          │
    │                                │   }                              │
    │                                │                                  │
    │ {users: [...,                  │                                  │
    │   top_scorer_score: 6, ...]}  │                                  │
    │◄───────────────────────────────┤◄─────────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/data/top_scorer_candidates.json` | Create | 50 entries: `{name, team, flag}` |
| `backend/internal/core/domain/top_scorer.go` | Create | `TopScorerCandidate`, `TopScorerPrediction`, `TopScorerResult` structs + `TopScorerPointsEarned(predicted, actual string) int` + `NormalizeName(s string) string` (lowercase, strip accents) |
| `backend/internal/core/ports/repositories.go` | Modify | Extend `UserRepository` with `UpsertTopScorerPrediction`, `GetTopScorerPredictionByUserID`, `GetAllTopScorerPredictions`; new `AppConfigRepository` interface or extend existing with `GetAppConfig`, `SetAppConfig` |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modify | Add DDL for `top_scorer_predictions` (UNIQUE on user_id) and `app_config` (key TEXT PK, value TEXT); add the 5 new repo methods; load candidates JSON in `NewSQLiteRepository` (or in a dedicated `LoadCandidates` method called by `main.go`) |
| `backend/internal/application/services/services.go` | Modify | New `TopScorerService` with `qualifierLockAt` injected at construction; methods: `GetMine`, `Upsert`, `GetByUserID`, `SetActual`, `ScoreUser`; new sentinel `ErrInvalidPlayer` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | New `TopScorerHandler` with `UpsertMine`, `GetMine`, `GetByUserID`; new `TopScorerAdminHandler` with `SetActual`; extend `ConfigHandler` to include `TopScorerCandidates` and serve them in `Get` |
| `backend/cmd/api/main.go` | Modify | Load candidates JSON at startup; instantiate new repos + service; register 4 new routes (`GET /top-scorer-prediction/me`, `PUT /top-scorer-prediction/me`, `GET /users/{id}/top-scorer-prediction`, `POST /admin/top-scorer`); extend ranking handler to compute `top_scorer_score` |
| `frontend/src/features/hooks.ts` | Modify | New `useTopScorerCandidates`, `useMyTopScorerPrediction`, `useUpsertTopScorerPrediction`, `useUserTopScorerPrediction` |
| `frontend/src/pages/GoleadorPage.tsx` | Create | New page at route `/goleador`; dropdown + save + lock banner |
| `frontend/src/components/TopScorerPredictionCard.tsx` | Create | Card showing a user's pick (used on GoleadorPage and UserPredictionsPage) |
| `frontend/src/pages/UserPredictionsPage.tsx` | Modify | Add `<TopScorerPredictionCard>` above the qualifier card |
| `frontend/src/pages/AdminPage.tsx` | Modify | New section "Goleador de la fase de grupos" with dropdown + optional free-text input + save button |
| `frontend/src/pages/RankingPage.tsx` | Modify | Add `top_scorer_score` column between "Clasificación" and "Total" |
| `frontend/src/App.tsx` (router) | Modify | Register `/goleador` route + nav link |
| `openspec/specs/top-scorer-prediction/spec.md` | Create (on archive) | Synced from `openspec/changes/top-scorer-prediction/specs/top-scorer-prediction/spec.md` |

## Interfaces / Contracts

```go
// backend/internal/core/domain/top_scorer.go
type TopScorerCandidate struct {
    Name string `json:"name"`
    Team string `json:"team"`
    Flag string `json:"flag"`
}

type TopScorerPrediction struct {
    ID              int       `json:"id"`
    UserID          int       `json:"user_id"`
    PredictedPlayer string    `json:"predicted_player"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// Returns 6 if predicted == actual (case + accent insensitive), 0 otherwise.
func TopScorerPointsEarned(predicted, actual string) int {
    if actual == "" {
        return 0
    }
    return BoolToPoints(NormalizeName(predicted) == NormalizeName(actual))
}

func NormalizeName(s string) string {
    // Lowercase + Unicode NFD + strip combining marks (accents)
    s = strings.ToLower(strings.TrimSpace(s))
    t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
    out, _, _ := transform.String(t, s)
    return out
}

// backend/internal/application/services/services.go
type TopScorerService struct {
    userRepo        ports.UserRepository
    cfgRepo         ports.AppConfigRepository
    qualifierLockAt time.Time
    candidates      []domain.TopScorerCandidate
}

func NewTopScorerService(userRepo ports.UserRepository, cfgRepo ports.AppConfigRepository, lockAt time.Time, candidates []domain.TopScorerCandidate) *TopScorerService

var ErrInvalidPlayer = errors.New("player not in candidate list")

// HTTP endpoints
//   GET    /top-scorer-prediction/me?user_id=<id>
//   PUT    /top-scorer-prediction/me  body: {user_id, predicted_player}
//   GET    /users/{id}/top-scorer-prediction
//   POST   /admin/top-scorer          body: {player: "..."}
//   GET    /config                    response adds top_scorer_candidates: [...]
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| E2E manual (smoke) | Full flow: load candidates, save pick, lock rejects, admin sets actual, ranking reflects | After deploy, with `QUALIFIER_LOCK_AT` overridden to the past, verify 403; restore to future, save 3 users with different picks, set one as actual, verify only that user gets 6 in `/ranking` |
| E2E manual | Rivals see each other's picks | Login as user A, save a pick, login as user B, navigate to user A's profile, verify the card shows user A's pick |
| E2E manual | Admin can set a name not in the list | On admin page, type a custom name, click save, verify warning logged + ranking recomputed |
| E2E manual | UI lock banner | With `QUALIFIER_LOCK_AT` in the past, navigate to `/goleador`, verify dropdown disabled and lock banner shown |

No unit tests added — matches the project's established pattern (`config.yaml: testing.backend.runner: none`, `strict_tdd: false`, the first test `match_test.go` was added for the auto-lock feature as a special case).

## Migration / Rollout

No migration. Two new tables (`top_scorer_predictions` and `app_config`) are created on startup via the existing `CREATE TABLE IF NOT EXISTS` pattern. Existing data is untouched.

Rollout: standard `feature/top-scorer-prediction` → `master` merge → Render auto-deploys. The candidates JSON ships in the binary's working dir; on Render, it's part of the deploy artifact. No env var changes (reuses `QUALIFIER_LOCK_AT`).

## Open Questions

None blocking. The case-insensitive + accent-insensitive match is the only design choice that could go either way (exact match vs fuzzy). We chose exact normalized match to keep scoring predictable for the user (their pick name must match the admin's recorded name).
