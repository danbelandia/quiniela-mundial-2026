# Exploration: group-standings

> Phase: explore
> Created: 2026-06-06
> Decision: continue to propose

## Current State

- `HomePage.tsx` muestra **un solo grupo a la vez** (botones "Anterior"/"Siguiente" ciclan por grupos A–L). El grupo activo se elige con `useState('A')`.
- `useMatches()` (hook en `features/hooks/`) trae **todos los 72 matches** en una sola request. Cada match ya tiene `group`, `home_team`, `away_team`, `home_score`, `away_score`, `status`, `home_flag`, `away_flag`.
- `Match.HomeScore` / `Match.AwayScore` = resultado real (no predicción). `Match.Status` ∈ {`"scheduled"`, `"finished"`, ...}.
- Existe `domain.Prediction.PointsEarned(actualHome, actualAway int) int` como fuente de verdad del cálculo 3/1/0 para predicciones de usuarios. **No hay** un equivalente para standings.
- `useMatches` no se vuelve a ejecutar al volver a la pestaña: solo se re-monta si cambia la ruta.
- No existe endpoint ni componente que muestre standings por grupo.

## Affected Areas

- `backend/internal/core/domain/` — crear `standings.go` con `GroupStanding`, `TeamStanding`, y función pura `CalculateStandings(matches []Match) []GroupStanding`.
- `backend/internal/core/ports/repositories.go` — agregar `GetAllMatchesForStandings(ctx) ([]domain.Match, error)` (o reusar el existente `GetAllMatches`).
- `backend/internal/application/services/services.go` — nuevo `StandingsService.CalculateAll() ([]domain.GroupStanding, error)` que delega al dominio.
- `backend/internal/infrastructure/handlers/handlers.go` — nuevo `StandingsHandler.GetAll(w, r)`.
- `backend/cmd/api/main.go` — nuevo route `mux.HandleFunc("GET /standings", standingsHandler.GetAll)`.
- `frontend/src/pages/HomePage.tsx` — agregar `<GroupStandings group={currentGroup} />` debajo de la `<table>` actual.
- `frontend/src/features/hooks/` — opcional: nuevo `useStandings()` con refetch al `focus` event.
- `frontend/src/shared/api/apiClient.ts` — sin cambios (usa `apiClient.get`).
- `openspec/specs/` — nueva spec `group-standings/spec.md` tras archivar este change.

## Approaches

### 1. Frontend-only calculation
- **Pros**: 0 cambios al backend. Hook `useMatches` ya tiene todos los matches. Cálculo en cliente: trivial con 72 matches. Más rápido de implementar.
- **Cons**: lógica de fútbol (3/1/0, tiebreakers) duplicada en JS. Rompe el principio de "single source of truth" que ya respetamos con `PointsEarned` en Go. No reusable para otros clientes.
- **Effort**: Low

### 2. Backend endpoint + frontend consumer (DDD-consistent)
- **Pros**: `domain.StandingsCalculator` paralelo a `domain.Prediction.PointsEarned`. Single source of truth. Reusable (mobile, API pública, tests). Encaja limpio en hexagonal: nuevo puerto + servicio + handler.
- **Cons**: más archivos modificados (5 backend + 1 frontend).
- **Effort**: Low–Medium

### 3. Backend endpoint + frontend calculation (híbrido)
- Backend expone `GET /standings/raw` con todos los matches finalizados.
- Frontend calcula y ordena.
- **Pros**: backend hace solo I/O, frontend hace cálculo (lógica de UI).
- **Cons**: duplicación parcial. La función pura de cálculo de fútbol se queda en frontend.
- **Effort**: Low–Medium

## Recommendation

**Approach 2 (backend endpoint + frontend consumer).** Razones:

1. **Consistencia con DDD existente**: ya extrajimos `PointsEarned` como función pura de dominio. Las reglas de fútbol 3/1/0 + tiebreakers (Pts > GD > GF) merecen el mismo tratamiento.
2. **El cálculo no es trivial de testear en JS** si queremos head-to-head, fair play, etc. En Go es un `if/else` simple, sin tipos raros.
3. **El dominio es el dueño de las reglas del mundial**, no la UI. Si mañana hay mobile o API pública, ya está cubierto.
4. **Esfuerzo adicional despreciable**: ~1 archivo nuevo + 2 archivos modificados en backend. El frontend queda igual de simple (1 fetch + 1 componente).

## Risks

- **Bajo**: 72 matches es data chica. Cálculo on-demand server-side < 5ms.
- **Bajo**: 12 grupos × 4 equipos × 9 columnas = 432 celdas, no es problema de render.
- **Medio**: refetch al focus event. Si el admin actualiza resultados y el usuario está en otra pestaña, al volver a HomePage se refetchea. Si está en HomePage mismo (caso raro porque el admin usa AdminPage), el refetch al focus es suficiente.
- **Bajo**: tiebreakers FIFA completos (head-to-head, fair play, sorteo) no se implementan en MVP. Usamos Pts > GD > GF > alfabético. Documentado en spec.

## Out of scope (para este change)

- Snapshots de standings en el tiempo (evolución "este equipo subió 2 posiciones") → ver Variante B de la evaluación previa.
- Tabla de posiciones general (de los 12 grupos) → solo la del grupo actual mostrado.
- Predicción del usuario mezclada con standings → no. Standings es solo del resultado real.
- Página nueva `/standings` con los 12 grupos a la vez → no. Es embebida en HomePage.

## Ready for Proposal

Yes. Decisiones de diseño ya tomadas con el usuario:

- **Columnas**: `# | Selección | PJ | G | E | P | GF | GC | Dif | Pts` (completa FIFA).
- **Visibilidad**: siempre, con 0s si no hay finalizados.
- **Refetch**: al `focus` event de la ventana (no polling, no botón manual).
- **Tiebreakers MVP**: Pts > GD > GF > nombre alfabético.
- **Change name**: `group-standings`.
- **Branch**: `feature/group-standings` (trabajo local, push solo con confirmación).
- **Persistencia**: en `openspec/` (config actual).
