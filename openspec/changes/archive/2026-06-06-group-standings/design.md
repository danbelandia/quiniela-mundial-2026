# Design: Tabla de posiciones por grupo (embebida en HomePage)

## Technical Approach

Función pura `domain.CalculateStandings` que itera los 72 matches, agrupa por grupo y equipo, calcula stats (PJ, G, E, P, GF, GC, Dif, Pts) solo para matches `finished`, y ordena con tiebreakers FIFA. Nuevo `StandingsService` + `StandingsHandler.GetAll` + route `GET /standings`. Frontend: nuevo hook `useStandings()` en `features/hooks.ts` con refetch al `focus` (throttled), nuevo componente `<GroupStandings>` que se renderiza debajo de la `<table>` en `HomePage.tsx`.

Sigue el patrón hexagonal existente: `domain.X.PointsEarned` → `services.RankingService` → `handlers.RankingHandler` → `mux.HandleFunc("GET /ranking", ...)` (mirror 1-a-1).

## Architecture Decisions

### Decision: Backend calcula standings, frontend consume

| Option | Tradeoff | Decision |
|---|---|---|
| Backend endpoint (DDD) | Reusable, single source of truth, testable | ✅ |
| Frontend calculation | Menos código, 0 cambios backend | ❌ duplica reglas de fútbol |
| Híbrido (raw + cálculo FE) | Backend solo I/O, FE cálculo | ❌ sin valor agregado |

### Decision: Un solo endpoint que devuelve los 12 grupos

| Option | Tradeoff | Decision |
|---|---|---|
| `GET /standings` (12 grupos) | 1 request, 432 celdas, cacheo trivial | ✅ |
| `GET /standings/{group}` (per-group) | 12 requests al navegar todos los grupos | ❌ |
| `GET /standings?group=X` con query | 2 endpoints, más params | ❌ overengineering |

### Decision: Función pura en `domain/standings.go`, NO método en `Match`

| Option | Tradeoff | Decision |
|---|---|---|
| `domain.CalculateStandings([]Match) []GroupStanding` | Función pura aislada, fácil de testear | ✅ |
| `(m Match).Stats() TeamStanding` | Acopla stats al match individual | ❌ rompe invariantes |
| Nuevo `Service.CalculateStandings` (en services) | Lógica de dominio en capa incorrecta | ❌ viola hexagonal |

### Decision: Refetch con throttling en frontend

| Option | Tradeoff | Decision |
|---|---|---|
| Listener de `focus` con throttle 10s | Simple, sin polling | ✅ |
| Polling cada 30s | Siempre fresco, gasta batería | ❌ |
| SSE/WebSocket | Tiempo real, overkill | ❌ fuera de scope |
| Botón manual | UX pobre | ❌ |

## Data Flow

```
Admin PUT /admin/matches/{id}    [cambia status a "finished", setea scores]
        │
        ▼
SQLite/Turso (tabla matches)     [datos actualizados]
        │
        ▼
GET /standings
        │
        ▼
StandingsHandler.GetAll
        │
        ▼
StandingsService.CalculateAll
        │
        ▼
domain.CalculateStandings(matches)  [función pura, on-demand]
        │
        ▼
JSON { groups: [{name, teams: [{position, name, flag, played, won, drawn, lost,
                                  goals_for, goals_against, goal_difference, points}]}] }
        │
        ▼
Frontend: useStandings() → useState → <GroupStandings groupName="A" />
        │
        ▼
Usuario ve la tabla actualizada
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/domain/standings.go` | Create | `TeamStanding` y `GroupStanding` structs + `CalculateStandings([]Match) []GroupStanding` pura |
| `backend/internal/application/services/services.go` | Modify | Nuevo `StandingsService` con `NewStandingsService(repo)` + `CalculateAll()` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | Nuevo `StandingsHandler.Service` + método `GetAll(w, r)` con `SetCORS` + `json.NewEncoder` |
| `backend/cmd/api/main.go` | Modify | `mux.HandleFunc("GET /standings", standingsHandler.GetAll)` + `OPTIONS /standings` |
| `frontend/src/features/hooks.ts` | Modify | Nuevo `useStandings()` con `useEffect` + `useState` + `addEventListener('focus', ...)` (throttled) |
| `frontend/src/pages/HomePage.tsx` | Modify | Importar `useStandings` + `<GroupStandings group={currentGroup} />` debajo de la `<table>` |
| `frontend/src/components/GroupStandings.tsx` | Create | Componente puro: renderiza la tabla del grupo actual con `min-w` responsive |

## Interfaces / Contracts

```go
// domain/standings.go
type TeamStanding struct {
    Position      int    `json:"position"`
    Team          string `json:"team"`
    Flag          string `json:"flag"`
    Played        int    `json:"played"`
    Won           int    `json:"won"`
    Drawn         int    `json:"drawn"`
    Lost          int    `json:"lost"`
    GoalsFor      int    `json:"goals_for"`
    GoalsAgainst  int    `json:"goals_against"`
    GoalDiff      int    `json:"goal_difference"`
    Points        int    `json:"points"`
}

type GroupStanding struct {
    Group string         `json:"group"`
    Teams []TeamStanding `json:"teams"`
}

func CalculateStandings(matches []Match) []GroupStanding
```

```typescript
// frontend/src/features/hooks.ts
interface TeamStanding { position, team, flag, played, won, drawn, lost,
                         goals_for, goals_against, goal_difference, points }
interface GroupStanding { group, teams: TeamStanding[] }

function useStandings(): { standings: GroupStanding[]; loading: boolean; refresh: () => void }
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Backend (manual curl) | 200 con 12 grupos, cálculo correcto de puntos, tiebreakers, partidos `scheduled` ignorados | `curl GET /standings` + comparación manual con `SELECT` |
| Frontend (manual browser) | Render debajo de la `<table>`, switch de grupo, refetch al focus | Tests E2E manuales (no hay framework) |
| Función pura | Lógica de cálculo (puede extraerse a tabla de verdad) | En MVP: casos manuales via curl |

## Migration / Rollout

No migration required. Cálculo on-demand sobre datos existentes.

## Open Questions

- [ ] ¿El admin debe ver también las standings en `AdminPage`? (de momento no — fuera de scope).
- [ ] ¿Cacheo en backend con TTL? (decidido: NO en MVP, 72 matches es despreciable).
