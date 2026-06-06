# Proposal: Login con username o email

## Intent

Hoy el login solo acepta `username`. Los usuarios que olvidaron su username quedan bloqueados. Permitir que el usuario decida con cuál de los dos identificadores (username o email) quiere iniciar sesión, manteniendo la contraseña como segundo factor. El backend detecta automáticamente qué tipo de credencial recibió.

## Scope

### In Scope
- Backend: aceptar `username` o `email` en el body de `POST /login`
- Backend: agregar `GetByEmail` al port `UserRepository` + implementación SQLite/libsql
- Backend: agregar índice único `idx_users_email` en la migración (idempotente con `IF NOT EXISTS`)
- Backend: agregar `UNIQUE` al constraint `email` en el schema (para DBs nuevas)
- Backend: ejecutar `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` en la DB de Turso existente
- Frontend: cambiar placeholder del input de "Usuario" a "Usuario o email"
- Frontend: lógica de envío: si el valor contiene `@`, enviar `{email, password}`, si no, `{username, password}`

### Out of Scope
- Recuperación de contraseña
- Login social (Google, etc.)
- 2FA
- Cambios al flujo de registro (sigue pidiendo ambos campos)
- Búsqueda case-insensitive (SQLite es case-sensitive por defecto; este cambio no toca eso)
- Normalización de email (lowercase, trim) — se puede agregar después

## Capabilities

### New Capabilities
- `user-login-by-email`: el sistema permite autenticar usuarios usando su email en lugar de username, manteniendo el resto del contrato de login.

### Modified Capabilities
- `user-auth` (si existe) o nuevo spec: el `POST /login` ahora acepta dos campos alternativos como identificador.

> Nota: si no existe `openspec/specs/user-auth/spec.md`, sdd-spec debe crearlo.

## Approach

1. **Schema** (`backend/internal/infrastructure/repository/sqlite_repository.go`): agregar `UNIQUE` a `email` en el `CREATE TABLE` (línea 39) y agregar `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` al final de `Migrate()`.
2. **Port** (`backend/internal/core/ports/repositories.go:5-12`): agregar método `GetByEmail(email string) (*domain.User, error)`.
3. **Repositorio** (`backend/internal/infrastructure/repository/sqlite_repository.go`): implementar `GetByEmail` espejo de `GetByUsername` (línea 248).
4. **Service** (`backend/internal/application/services/services.go:84`): el método `Login(identifier, password string)` detecta `@` en `identifier` y despacha a `GetByEmail` o `GetByUsername`.
5. **Handler**: leer `email` o `username` del body; pasar el identificador único al service.
6. **Frontend** (`frontend/src/pages/LoginPage.tsx:26`): cambiar placeholder a `"Usuario o email"`, y la lógica de envío detecta `@` para mandar `{email, password}` vs `{username, password}`.
7. **Turso manual**: correr `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` una vez antes del deploy (usuario ya confirmó que no hay duplicados).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modified | Schema + Migrate + nuevo `GetByEmail` |
| `backend/internal/core/ports/repositories.go` | Modified | Nuevo método `GetByEmail` en el port |
| `backend/internal/application/services/services.go` | Modified | `Login` acepta username o email |
| `backend/internal/infrastructure/handlers/handlers.go` | Modified | Lee `email` o `username` del body |
| `frontend/src/pages/LoginPage.tsx` | Modified | Placeholder + lógica de envío |
| Turso DB (producción) | Manual | Correr `CREATE UNIQUE INDEX` antes del deploy |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Email duplicado cause falla del `CREATE UNIQUE INDEX` | Low | Usuario confirmó 0 duplicados |
| Usuario con username que contiene `@` rompa la detección | Low | Username no permite `@` en el register; aún así, `GetByUsername` se llama solo si no hay `@` |
| Migración nueva choque con la DB existente | Low | `CREATE UNIQUE INDEX IF NOT EXISTS` es idempotente |
| Confusión UX si usuario no sabe si es username o email | Low | Placeholder lo aclara: "Usuario o email" |
| Contraseña incorrecta con email válido devuelva mismo error que "no existe" | Low | El service ya retorna `errors.New("invalid credentials")` para ambos casos (mantiene el comportamiento actual) |

## Rollback Plan

1. Revertir el commit: `git revert <commit-sha>`
2. Redeploy backend y frontend
3. La DB de Turso **no se toca** — el `CREATE UNIQUE INDEX` no se elimina (es harmless y permite rollback futuro más rápido). Si se quisiera limpiar: `DROP INDEX IF EXISTS idx_users_email;`

## Dependencies

- Usuario debe correr manualmente `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` en la consola de Turso **antes** del deploy (1 minuto, no rompe nada).
- Backend en Render: redeploy manual desde el dashboard o esperar auto-deploy si está configurado.
- Frontend en Render (Static Site): redeploy con `git push` o manual desde el dashboard.

## Success Criteria

- [ ] Login con `username` existente sigue funcionando (regresión)
- [ ] Login con `email` registrado funciona y devuelve el mismo user_id
- [ ] Login con credenciales inválidas devuelve 401 con mensaje genérico
- [ ] El schema de la DB de Turso tiene el índice `idx_users_email`
- [ ] El placeholder del frontend dice "Usuario o email"
- [ ] No hay nuevos tipos de error que rompan el flujo de login existente
- [ ] Test manual con `dangel` / `admin@quiniela.com` + password `281193` funciona con ambos campos
