# Proposal: Tabla de posiciones por grupo (embebida en HomePage)

## Intent

Los usuarios no pueden ver cómo van los grupos del mundial. Solo ven partidos aislados y sus propias predicciones. Necesitamos una **tabla de posiciones en vivo** (3/1/0 con tiebreakers FIFA) embebida debajo del bloque de partidos en HomePage, para visualizar la "carrera" de cada selección.

## Scope

### In Scope
- Endpoint `GET /standings` con las 12 tablas calculadas desde resultados reales.
- Cálculo on-demand server-side. Sin nueva tabla en DB.
- Tiebreakers MVP: Pts → GD → GF → alfabético.
- Componente `<GroupStandings groupName="X" />` debajo de la `<table>` de HomePage.
- Refetch al `focus` (con throttle >10s).
- Spec nueva `group-standings` con 2–3 requirements y escenarios.

### Out of Scope
- Snapshots / evolución temporal, página `/standings` standalone, mezclar con predicciones, tiebreakers FIFA completos, stats avanzadas.

## Capabilities

### New Capabilities
- `group-standings`: cálculo y exposición de tablas de posiciones por grupo basadas en resultados reales.

### Modified Capabilities
- Ninguna.

## Approach

Backend hexagonal: `domain.CalculateStandings([]Match) []GroupStanding` (paralela a `PointsEarned`), nuevo `StandingsService` + handler + route. Reutiliza `GetAllMatches`. Frontend: hook `useStandings()` con refetch al `focus`, componente Tailwind responsive.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/core/domain/standings.go` | New | `GroupStanding`, `TeamStanding` structs + `CalculateStandings` pura |
| `backend/internal/application/services/services.go` | Modified | Nuevo `StandingsService` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | Nuevo `StandingsHandler.GetAll` |
| `backend/cmd/api/main.go` | Modified | `mux.HandleFunc("GET /standings", ...)` |
| `frontend/src/pages/HomePage.tsx` | Modified | `<GroupStandings group={currentGroup} />` debajo de la `<table>` |
| `frontend/src/features/hooks/useStandings.ts` | New | Hook con refetch al `focus` |
| `openspec/specs/group-standings/spec.md` | New | Spec al archivar el change |

## Risks

| Risk | Mitigation |
|------|------------|
| Refetch al focus se dispara en interacciones no relacionadas | Throttle >10s desde último fetch |
| Cálculo inconsistente con reglas FIFA futuras | Función pura aislada; spec documenta tiebreakers MVP |
| Tabla rompe layout mobile | Misma técnica que `RankingPage` (`overflow-x-auto` + `min-w`) |

## Rollback Plan

Revertir el commit. No hay migración de DB. Feature puramente aditiva: el endpoint nuevo no se llama desde código existente, y el componente se renderiza solo en HomePage. Eliminar componente + endpoint = vuelta al estado anterior.

## Success Criteria

- [ ] `GET /standings` devuelve 12 grupos con 4 equipos ordenados por puntos.
- [ ] Tras `PUT /admin/matches/{id}` que finaliza un partido, al volver al tab de HomePage la tabla refleja el nuevo resultado.
- [ ] Test backend: México 2–0 finalizado → México +3 pts; rival 0.
- [ ] Test frontend: ranking en orden esperado tras varios finalizados.
- [ ] Spec archivada tras merge.
- [ ] Sin regresión.
