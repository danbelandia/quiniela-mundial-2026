# Proposal: Predicciones de usuario clickeables desde el ranking

## Intent

Hoy el ranking muestra solo el puntaje de cada usuario. Para gamificar y transparentar la quiniela, cada fila del ranking debe permitir ver el detalle: qué pronósticos hizo ese usuario en cada partido, agrupados por grupo (A, B, C... L), con el resultado real y los puntos obtenidos. La predicción es **pública para todos los usuarios logueados** (decisión confirmada por el usuario).

## Scope

### In Scope
- Backend: nuevo endpoint `GET /users/{id}/predictions` que devuelve predicciones de un usuario con info del match embebida y puntos calculados
- Backend: nuevo método en `PredictionRepository` (port + impl SQLite) con `JOIN` a `matches`
- Frontend: nueva página `UserPredictionsPage` con tabs por grupo (A-L) y tabla de partidos del grupo con pronóstico + resultado + puntos
- Frontend: cada fila del ranking se vuelve clickeable (link al username) navegando a `/user/{id}`
- Frontend: ruta nueva `/user/{id}` protegida por `ProtectedRoute`

### Out of Scope
- Perfil público completo del usuario (bio, avatar, stats históricos)
- Compartir predicciones en redes sociales
- Edición de predicciones desde la vista detalle (la edición sigue en HomePage)
- Comparar dos usuarios side-by-side
- Permisos especiales (todos los usuarios logueados ven todo, sin roles)

## Capabilities

### New Capabilities
- `user-predictions-detail`: el sistema permite a un usuario logueado ver el detalle de predicciones de cualquier otro usuario, agrupadas por grupo, mostrando pronóstico + resultado real + puntos.

### Modified Capabilities
- `user-auth`: ninguno (no cambia login).
- `user-ranking`: el username de cada fila pasa a ser un link, no a un `<td>` plano. (Esto es solo UI, sin cambio de requisitos — no necesita delta spec.)

## Approach

1. **Port** (`ports/repositories.go:24-27`): agregar `GetByUserID(userID int) ([]PredictionWithMatch, error)` con un struct `PredictionWithMatch` que embebe `Prediction` + campos del match.
2. **Repo** (`sqlite_repository.go`): implementar con `JOIN` a `matches` + `LEFT JOIN` a sí mismo si hace falta. Calcular puntos con `CalculatePoints` del `RankingService` (o replicar la fórmula simple).
3. **Service** (`services.go`): método `GetUserPredictions(userID int)` que valida que el user existe, llama al repo, calcula puntos.
4. **Handler** (`handlers.go`): nuevo `GetUserPredictions(w, r)` con `r.PathValue("id")`.
5. **Routing** (`main.go`): `mux.HandleFunc("GET /users/{id}/predictions", userHandler.GetUserPredictions)`.
6. **Frontend nuevo** (`UserPredictionsPage.tsx`): fetch al endpoint, agrupa por `group_name`, muestra tabs, tabla con columnas `Match | Resultado Real | Pronóstico | Puntos`. Botón "Volver al ranking".
7. **App.tsx**: + ruta `/user/:id` dentro de `ProtectedRoute`.
8. **RankingPage.tsx**: el `<td>{u.username}</td>` se envuelve en `<Link to={'/user/' + u.id}>` (o se reemplaza por un `<td>` con `onClick` + `useNavigate`).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/core/ports/repositories.go` | Modified | + `GetByUserID` + struct |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modified | + impl con JOIN |
| `backend/internal/application/services/services.go` | Modified | + `GetUserPredictions` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | + handler |
| `backend/cmd/api/main.go` | Modified | + ruta |
| `frontend/src/pages/UserPredictionsPage.tsx` | New | Página nueva |
| `frontend/src/app/App.tsx` | Modified | + ruta `/user/:id` |
| `frontend/src/pages/RankingPage.tsx` | Modified | Username → link |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Performance con 72 predicciones JOIN | Low | Una sola query, ~72 rows, sin problema |
| Lógica de puntos duplicada entre ranking y detail | Low | Reusar `CalculatePoints` del `RankingService` |
| Usuario con pocas predicciones muestra UI vacía | Low | Mostrar mensaje "Aún no hizo pronósticos" |
| Mobile: 12 tabs de grupos se ven apretados | Low | Scroll horizontal o select nativo en mobile |
| Cambios en `RankingPage` rompen estilos | Low | Mantener la misma estructura de `<tr>`, solo cambiar el contenido del `<td>` |

## Rollback Plan

1. `git revert <commit-sha>` desde `feature/clickable-user-predictions` o `git reset --hard master` en la rama
2. Redeploy → frontend vuelve a mostrar ranking sin links
3. Endpoint nuevo queda disponible (no rompe nada si no se llama)

## Dependencies

- Backend en Render: redeploy después del merge a master
- Frontend en Render: redeploy después del merge a master
- **No requiere migración de DB**

## Success Criteria

- [ ] Click en un username del ranking navega a `/user/{id}`
- [ ] La página muestra los 12 grupos como tabs/selector
- [ ] Cada partido muestra: equipos con flags, resultado real, pronóstico del usuario, puntos (3, 2, 1, 0)
- [ ] Partidos sin pronóstico del usuario se ven vacíos en la columna "Pronóstico"
- [ ] Predicciones de cualquier usuario son visibles para cualquier usuario logueado (no hay chequeo de ownership)
- [ ] Botón "Volver al ranking" funciona
- [ ] Mobile: la página es usable (tabs no rompen layout)
- [ ] Endpoint `GET /users/{id}/predictions` responde en <500ms para un usuario con 72 predicciones
