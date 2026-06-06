# Proposal: Predicción 1° y 2° por grupo

## Intent

Complemento a la predicción de marcadores: pronosticar **1° y 2° de cada grupo** (12 × 6 pts máx = +72 pts al total).

## Scope

### In Scope
- Tabla `group_qualifier_predictions` (sin migración).
- 2 endpoints: upsert + GET mi predicción por grupo.
- UI "Predicción 1° y 2°" bajo la grilla en `HomePage`, dos `<select>` filtrados por grupo activo.
- Lock sincronizado: partido bloqueado → `<select>` disabled.
- Scoring FIFA (Pts > DG > GF > H2H > DG H2H > GF H2H > fair play > sorteo).

### Out of Scope
- Histórico. Fair play real. Eliminatorias. Notificaciones.

## Capabilities

### New Capabilities
- `group-qualifier-prediction`: predicción, validación, lock, scoring FIFA, ranking.

### Modified Capabilities
- Ninguna. Ranking y predicción de partidos no tienen spec propia; este cambio es autocontenido.

## Approach

**Backend**: tabla con `UNIQUE(user_id, group_name)`. `domain.CalculateQualifiers` (FIFA tie-breakers puros en Go) + `QualifiersPointsEarned` (3 por acierto). `QualifierPredictionService` con Upsert/Get/Score. 2 handlers + `PUT/GET /groups/{group}/qualifier-predictions`.

**Frontend**: hook `useQualifierPrediction`; `QualifierPredictionSection` con 2 `<select>` disabled si algún match `is_locked`; autosave con debounce + `AbortController`. Ranking añade `qualifier_points` on-the-fly.

## Affected Areas

| Area | Impact |
|------|--------|
| `backend/internal/core/domain/qualifier.go` | New |
| `backend/internal/core/ports/repositories.go` | Modified (3 métodos) |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modified (CREATE TABLE + 3 métodos) |
| `backend/internal/application/services/services.go` | Modified (`QualifierPredictionService` + ranking) |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified (2 handlers + routes) |
| `backend/cmd/api/main.go` | Modified (wire) |
| `frontend/src/features/hooks.ts` | Modified (`useQualifierPrediction`) |
| `frontend/src/components/QualifierPredictionSection.tsx` | New |
| `frontend/src/pages/HomePage.tsx` | Modified (render + lock sync) |

## Risks

| Risk | Lik | Mitigation |
|------|-----|------------|
| Tie-breakers FIFA en empates raros | Med | Tests manuales |
| Lock check lee 6 matches/request | Low | Despreciable |
| Autosave + unmount | Low | `AbortController` |

## Rollback Plan

1. Revert PR → Render redeploy sin rutas.
2. `DROP TABLE group_qualifier_predictions;` en Turso.

## Dependencies

- `domain.CalculateStandings` (base para FIFA tie-breakers).
- `Match.IsLocked` (lock sincronizado).

## Success Criteria

- [ ] Guardar/editar predicción desde `HomePage`.
- [ ] Partido bloqueado → `<select>` disabled.
- [ ] Validación: mismo equipo 1°/2° o fuera del grupo → 400.
- [ ] Grupo cerrado (6 partidos `finished`) → ranking suma 3 pts por acierto.
- [ ] Tie-breakers FIFA: Pts > DG > GF > H2H > DG H2H > GF H2H.
- [ ] Empate persistente → sortear (hash determinístico).
- [ ] Sin regresiones en scoring de partidos.
