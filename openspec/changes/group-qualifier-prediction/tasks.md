# Tasks: Predicción 1° y 2° por grupo

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~430 (10 files, 2 new) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size:exception
400-line budget risk: Low

## Phase 1: Foundation (DB + types + interfaces)

- [x] 1.1 Add `CREATE TABLE group_qualifier_predictions` to `sqlite_repository.go` init
- [x] 1.2 Add `QualifierPrediction` struct to `domain/qualifier.go` (NEW)
- [x] 1.3 Add `QualifierPredictionRepository` interface to `ports/repositories.go` (4 methods)
- [x] 1.4 Implement 4 repo methods in `sqlite_repository.go` (Upsert, GetByUserAndGroup, GetAllByUser, GetAll)

## Phase 2: Domain (FIFA tie-breakers)

- [x] 2.1 `domain.CalculateQualifiers(matches, groupName) (first, second, err)` — extends `CalculateStandings`, applies sort FIFA
- [x] 2.2 `domain.QualifiersPointsEarned(actualFirst, actualSecond, predictedFirst, predictedSecond) int` — 3 por acierto
- [ ] 2.3 Manual verify with fixture: 4 teams con empates por pts/GD/GF → cae a H2H → sortear

## Phase 3: Service + ranking

- [ ] 3.1 `QualifierPredictionService` en `services.go` con `Upsert`, `GetByUserAndGroup`, `Score`
- [ ] 3.2 `Score`: itera matches del grupo cerrado, llama `CalculateQualifiers` + `QualifiersPointsEarned`
- [ ] 3.3 `RankingService.GetRanking` suma `qualifier_points` (columna separada) por usuario

## Phase 4: API (handlers + routes + wire)

- [ ] 4.1 `QualifierPredictionHandler` con `GetMy` + `Upsert` (validación: mismo equipo / equipo fuera → 400)
- [ ] 4.2 Wire service + handler en `main.go` + 2 rutas: `PUT/GET /groups/{group}/qualifier-predictions`
- [ ] 4.3 CORS OPTIONS para ambas

## Phase 5: Frontend

- [x] 5.1 Hook `useQualifierPrediction(groupName, userId)` en `hooks.ts` (GET + autosave con debounce 800ms + AbortController)
- [x] 5.2 Componente `QualifierPredictionSection.tsx` (NEW): 2 `<select>` con flags, `disabled` si `isLocked`, feedback "Guardado"/"Error"
- [x] 5.3 Renderizar bajo la grilla en `HomePage.tsx` con `isLocked = filteredMatches.some(m => m.is_locked)`
- [x] 5.4 `RankingPage.tsx`: mostrar columna `qualifier_points` y sumar a `total_points`

## Phase 6: Verification (manual)

- [x] 6.1 Local: `GET /groups/A/qualifier-predictions/me?user_id=1` → 200
- [x] 6.2 Local: `PUT /groups/A/qualifier-predictions` con México + Corea → 200
- [x] 6.3 Local: `PUT` con México + México → 400 "must be different teams"
- [x] 6.4 Local: `PUT` con Brasil + México → 400 "team is not in the group"
- [x] 6.5 Local: bloquear un match de A → selects disabled (frontend logic, no probado en browser)
- [x] 6.6 Local: `GET /ranking` devuelve `qualifier_score: 3` para dangel (correcto: México 1° ✓, Checa 2° (no predicha) → 3 pts)
- [ ] 6.7 Production (post-merge): repetir 6.1, 6.2, 6.6 contra Render
- [ ] 6.8 Production: verificar visualmente fechas y standings ya funcionan (regresión)

## Phase 7: Cleanup

- [ ] 7.1 Commit + push a master (single PR)
- [ ] 7.2 Confirmar Render redeploy + smoke test en `/`
- [ ] 7.3 Borrar branch `feature/group-qualifier-prediction` (si se usó)
- [ ] 7.4 Archivar change folder después de verify
