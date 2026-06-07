# Design: Predicción 1° y 2° por grupo

## Technical Approach

Extender `domain.CalculateStandings` con `CalculateQualifiers(matches, groupName) (first, second, err)` que aplica tie-breakers FIFA. Nueva tabla `group_qualifier_predictions` con `UNIQUE(user_id, group_name)`. Servicio `QualifierPredictionService` con `Upsert`, `GetByUserAndGroup`, `Score`. `RankingService` suma `qualifier_points` on-the-fly. Frontend consume 2 endpoints y renderiza 2 `<select>` debajo de la grilla, deshabilitados si algún match del grupo activo está `is_locked`.

## Architecture Decisions

### Decision: FIFA tie-breakers en dominio puro

**Choice**: `domain.CalculateQualifiers` recibe `[]Match` y devuelve `(first, second string)`. Reusa `CalculateStandings` y aplica sort FIFA.
**Alternatives**: SQL window functions; frontend calcula; tabla cacheada.
**Rationale**: Dominio puro (testable sin DB), evita N+1, sin migración. Sort O(N²) por grupo (N=4) — despreciable.

### Decision: Auth implícita por user_id en body

**Choice**: Frontend manda `user_id` en el body. Backend no valida token.
**Alternatives**: JWT middleware; cookie session.
**Rationale**: Ya así funcionan `/predictions` y `/users/{id}/predictions`. Mantener consistencia.

### Decision: Sortear determinístico por hash de match IDs

**Choice**: Si los tie-breakers FIFA fallan, orden = `sha256(match_ids_sorted + team_name)[:8]` como entero.
**Alternatives**: Math/rand; persistir sortear; pedir al admin.
**Rationale**: Determinístico, reproducible, sin estado extra.

### Decision: Lock check 100% frontend

**Choice**: `QualifierPredictionSection` recibe `isLocked: boolean` = `filteredMatches.some(m => m.is_locked)`.
**Alternatives**: Endpoint `/groups/{group}/is-locked`; SSE.
**Rationale**: `filteredMatches` ya está en memoria, no justifica round-trip extra.

### Decision: Score en ranking bajo demanda

**Choice**: `RankingService.GetRanking()` itera usuarios, suma `qualifier_points` (solo grupos cerrados) a `total_points`.
**Alternatives**: Cache por partido finalizado; job batch.
**Rationale**: ~3600 ops en el peor caso. Despreciable. Sin cache hasta evidenciar problema.

## Data Flow

```
HomePage (currentGroup)
   │
   ├─── <QualifierPredictionSection group={currentGroup} isLocked={...} userId={user.id} />
   │         │
   │         ├── GET /groups/{group}/qualifier-predictions/me?user_id=N
   │         │      └── QualifierPredictionService.GetByUserAndGroup → row
   │         │
   │         └── PUT /groups/{group}/qualifier-predictions  body={user_id, predicted_first, predicted_second}
   │                └── QualifierPredictionService.Upsert → upsert row
   │
   └─── <RankingPage /> (refresh)
              │
              └── GET /ranking
                     └── RankingService.GetRanking
                            ├── para cada user: match points (existente)
                            └── para cada user, cada grupo cerrado:
                                  ├── domain.CalculateQualifiers(matches, group) → (first, second)
                                  └── QualifierPredictionService.Score(user, group, first, second) → 3/0/3
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/domain/qualifier.go` | Create | `CalculateQualifiers` (FIFA tie-breakers) + `QualifiersPointsEarned` |
| `backend/internal/core/ports/repositories.go` | Modify | Add `QualifierPredictionRepository` interface + 4 métodos |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modify | CREATE TABLE + 4 métodos (Upsert/GetByUserAndGroup/GetAllForRanking/GetAllUsers) |
| `backend/internal/application/services/services.go` | Modify | `QualifierPredictionService` + ranking agrega `qualifier_points` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | `QualifierPredictionHandler` con 2 métodos + 2 rutas |
| `backend/cmd/api/main.go` | Modify | Wire `qualifierRepo`, service, handler, 2 routes |
| `frontend/src/features/hooks.ts` | Modify | `useQualifierPrediction(groupName, userId)` con autosave debounced |
| `frontend/src/components/QualifierPredictionSection.tsx` | Create | UI con 2 `<select>` + flag emoji + disabled state + feedback |
| `frontend/src/pages/HomePage.tsx` | Modify | Render `<QualifierPredictionSection>` debajo de la grilla + cálculo de `isLocked` |
| `frontend/src/pages/RankingPage.tsx` | Modify | Mostrar `qualifier_points` columna + sumar en `total_points` |

## Interfaces / Contracts

**Go (dominio)**:
```go
// domain.CalculateQualifiers: pura, []Match → (first, second string, err)
type QualifierPrediction struct {
    ID             int
    UserID         int
    GroupName      string
    PredictedFirst string
    PredictedSecond string
    CreatedAt      time.Time
}

func CalculateQualifiers(matches []Match, groupName string) (string, string, error)
func QualifiersPointsEarned(actualFirst, actualSecond, predictedFirst, predictedSecond string) int
```

**REST**:
- `PUT /groups/{group}/qualifier-predictions` body: `{user_id, predicted_first, predicted_second}` → 200 `{id, ...}` o 400
- `GET /groups/{group}/qualifier-predictions/me?user_id=N` → 200 `{predicted_first, predicted_second}` o `{...null}` o 401

**Frontend (TypeScript)**:
```ts
interface QualifierPrediction {
  predicted_first: string;
  predicted_second: string;
}
// Hook retorna { prediction, save(first, second), loading, error }
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | `CalculateQualifiers` con fixtures de empates | Go test nativo (no `_test.go` aún, se puede añadir) |
| Integration | `PUT` con datos válidos + 400 con inválidos | curl manual contra local + Render después |
| E2E (browser) | Save → reload → prediction persiste. Lock match → selects disabled | Playwright manual: F5 + check |
| Regression | `GET /ranking` con y sin qualifier predictions devuelve totales correctos | curl: comparar antes/después |

No hay test framework instalado (`strict_tdd: false`). Plan: testing manual con un grupo hipotético cerrado + inspección visual.

## Migration / Rollout

No migration required. `CREATE TABLE group_qualifier_predictions` se ejecuta en el init del repo. La tabla se crea vacía — sin backfill, sin migración de datos existentes. Render redeploya y la crea automáticamente.

## Open Questions

- ¿Mostrar puntos de qualifier en la columna "Total" del ranking, o en una columna separada `qualifier_points`? **Decisión**: columna separada (transparencia).
- ¿Qué pasa si Render redeploya antes de que el frontend tenga el JS actualizado? 404 en `/groups/{group}/qualifier-predictions`. **Mitigación**: deploy atómico (backend + frontend en el mismo push). Acceptable.
- ¿Auto-save con debounce puede causar guardado parcial si el usuario navega rápido? **Mitigación**: `AbortController` cleanup en el hook, server-side idempotente.
