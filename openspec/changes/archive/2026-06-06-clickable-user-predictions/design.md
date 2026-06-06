# Design: Predicciones de usuario clickeables desde el ranking

## Technical Approach

Habilitar una vista detalle de las predicciones de un usuario, accedida haciendo click en cualquier fila del ranking. La vista detalle es una página dedicada `/user/:id` con tabs por grupo (A-L), alimentada por un nuevo endpoint `GET /users/{id}/predictions` que devuelve predicciones con la información del match embebida y los puntos ya calculados.

## Architecture Decisions

### Decision: Mover la lógica de puntos a `domain.Prediction.PointsEarned`

**Choice**: Convertir `CalculatePoints` (actualmente en `RankingService:61`) en un método del dominio `domain.Prediction.PointsEarned(actualHome, actualAway int) int`. `RankingService.CalculatePoints` queda como wrapper deprecado o se elimina y se actualiza el único caller (handlers.go:109).
**Alternatives considered**:
- Duplicar la fórmula en el nuevo `PredictionService.GetUserPredictions`.
- Inyectar `RankingService` en `PredictionService` para llamar `rankingService.CalculatePoints`.
**Rationale**: Domain-driven — la predicción sabe calcular sus puntos. Evita acoplar servicios entre sí y duplicar lógica. La fórmula vive en un solo lugar.

### Decision: Endpoint devuelve predicciones con match embebido (no dos requests)

**Choice**: El endpoint devuelve una sola response con `user` + `predictions[]`, donde cada `prediction` incluye los campos del match (group, teams, flags, date, status, real scores). El cálculo de puntos se hace en el service.
**Alternatives considered**:
- Frontend hace 2 requests: `GET /users/{id}/predictions` + `GET /matches` y los mergea client-side.
- Backend devuelve predicciones crudas y el frontend calcula puntos (replicando la fórmula).
**Rationale**: Una sola query SQL con JOIN, sin N+1, contrato más claro. El cálculo de puntos se queda en el backend (única fuente de verdad, evita que el frontend quede desincronizado si cambian las reglas).

### Decision: Páginas con tabs por grupo, no tabla única

**Choice**: `UserPredictionsPage` agrupa las predicciones por `group_name` y muestra tabs. Solo se renderiza el grupo activo a la vez.
**Alternatives considered**:
- Una sola tabla larga con los 72 partidos y secciones por grupo.
- Acordeón expandible por grupo.
**Rationale**: 72 filas es mucho para escanear. Tabs reducen carga cognitiva. En mobile, las tabs se vuelven un selector nativo (`<select>`) si el ancho es chico.

### Decision: Backend NO chequea ownership

**Choice**: El handler `GetUserPredictions` no valida que el `user_id` del path coincida con el del JWT (no hay JWT, solo el estado del cliente). Cualquier usuario logueado puede ver las predicciones de cualquier otro.
**Alternatives considered**:
- Validar `userID` contra la sesión, devolver 403 si no coincide.
**Rationale**: Decisión confirmada por el usuario: predicciones públicas para todos los logueados. El frontend ya protege la ruta con `ProtectedRoute` (login required), no por usuario.

## Data Flow

```
User clicks username "dangel" in ranking
        │
        ▼
RankingPage <td> wrapped in <Link to="/user/1">
        │
        ▼
React Router matches /user/:id → UserPredictionsPage
        │
        ▼
useEffect: apiClient.get('/users/1/predictions')
        │
        ▼
GET /users/{id}/predictions
        │
        ▼
UserHandler.GetUserPredictions
        │  validates user exists (userService.GetByID) → 404 if not
        ▼
PredictionService.GetUserPredictions(userID)
        │  calls repo, calculates points per prediction
        ▼
SQLiteRepository.GetByUserIDWithMatch (JOIN a matches)
        │
        ▼
DB → rows → for each row: build PredictionWithMatch
        │  calculate points with pred.PointsEarned(match.HomeScore, match.AwayScore)
        ▼
Return JSON: { user: {...}, predictions: [{..., points: N}, ...] }
        │
        ▼
Frontend: useState, group by group_name, render tabs
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/domain/match.go` (or new prediction.go) | Modify | + método `PointsEarned` en `Prediction` |
| `backend/internal/application/services/services.go` | Modify | `RankingService.CalculatePoints` ahora delega a `p.PointsEarned` (refactor mínimo) + nuevo método `PredictionService.GetUserPredictions` |
| `backend/internal/core/ports/repositories.go` | Modify | + struct `PredictionWithMatch` + método `GetByUserIDWithMatch` en port |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modify | + impl `GetByUserIDWithMatch` con JOIN |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | + `GetUserPredictions` handler |
| `backend/cmd/api/main.go` | Modify | + ruta `GET /users/{id}/predictions` |
| `frontend/src/pages/UserPredictionsPage.tsx` | New | Página nueva con tabs por grupo |
| `frontend/src/app/App.tsx` | Modify | + ruta `/user/:id` dentro de `ProtectedRoute` |
| `frontend/src/pages/RankingPage.tsx` | Modify | Username envuelto en `<Link to={'/user/' + u.id}>` |

## Interfaces / Contracts

```go
// domain/prediction.go (new) or match.go (modify)
type Prediction struct {
    ID, UserID, MatchID, HomeScore, AwayScore int
}
func (p *Prediction) PointsEarned(actualHome, actualAway int) int {
    if p.HomeScore == actualHome && p.AwayScore == actualAway { return 3 }
    if p.HomeScore == p.AwayScore && actualHome == actualAway { return 1 }
    if (p.HomeScore > p.AwayScore) == (actualHome > actualAway) { return 2 }
    return 0
}

// ports/repositories.go
type PredictionWithMatch struct {
    Prediction
    GroupName  string  `json:"group_name"`
    HomeTeam   string  `json:"home_team"`
    HomeFlag   string  `json:"home_flag"`
    AwayTeam   string  `json:"away_team"`
    AwayFlag   string  `json:"away_flag"`
    MatchDate  string  `json:"match_date"`
    Status     string  `json:"status"`
    RealHome   int     `json:"home_score_real"`
    RealAway   int     `json:"away_score_real"`
    Points     int     `json:"points"`
}

type PredictionRepository interface {
    Save(prediction domain.Prediction) error
    GetByMatchID(matchID int) ([]domain.Prediction, error)
    GetAllPredictions() ([]domain.Prediction, error)
    GetByUserIDWithMatch(userID int) ([]domain.PredictionWithMatch, error)
}

// JSON response
{
  "user": { "id": 1, "username": "dangel" },
  "predictions": [
    {
      "match_id": 5, "group_name": "A",
      "home_team": "Argentina", "home_flag": "🇦🇷",
      "away_team": "Brasil", "away_flag": "🇧🇷",
      "match_date": "2026-06-15T20:00:00Z", "status": "finished",
      "home_score_pred": 2, "away_score_pred": 1,
      "home_score_real": 2, "away_score_real": 1,
      "points": 3
    }
  ]
}
```

```tsx
// RankingPage.tsx (diff)
- <td className="...">{u.username}</td>
+ <td className="...">
+   <Link to={`/user/${u.id}`} className="hover:underline cursor-pointer">
+     {u.username}
+   </Link>
+ </td>

// App.tsx (diff)
+ <Route path="/user/:id" element={<UserPredictionsPage />} />
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | — | N/A (no test runner, `strict_tdd: false`) |
| Integration | Endpoint + JOIN | Manual con `curl` (8 casos) |
| E2E | Click en ranking → detail page | Manual en browser (3 casos) |

**Test plan manual**:

Backend (con `body1.json` o equivalente):
1. `GET /users/1/predictions` con user dangel → 200 + JSON con predicciones
2. `GET /users/9999/predictions` (id inexistente) → 404
3. `GET /users/1/predictions` con user sin predicciones → 200 + array vacío
4. Verificar `points` field: predicción exacta → 3, ganador correcto → 2, empate → 1, otro → 0
5. Verificar `points` para match `status != "finished"` → 0 (o null)

Frontend:
6. Login → ir a `/ranking` → click en username "dangel" → URL cambia a `/user/1`
7. En `/user/1`: 12 tabs visibles, tabla del grupo A con partidos
8. Click en tab "B" → muestra partidos del grupo B
9. Volver al ranking → URL `/ranking`

## Migration / Rollout

No requiere migración de DB. El schema no cambia.

**Orden**:
1. Branch `feature/clickable-user-predictions` (ya creada)
2. Implementar (fase apply)
3. Local test con backend en `:8080` y frontend en `:5173`
4. Commit
5. **NO pushear** (usuario quiere trabajar en local)
6. Cuando confirmen, merge a `master` → Render auto-deploys

## Open Questions

Ninguna. La decisión de privacidad y la opción B están confirmadas.
