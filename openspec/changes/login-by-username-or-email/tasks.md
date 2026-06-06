# Tasks: Login con username o email

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~40 (5 files modified, 0 new) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend port + repo + service + handler + frontend | PR 1 | Single PR — total < 50 lines |

## Phase 1: Backend — Foundation

- [x] 1.1 Add `GetByEmail(email string) (*domain.User, error)` to `UserRepository` port in `backend/internal/core/ports/repositories.go`
- [x] 1.2 Add `UNIQUE` to `email` column in `CREATE TABLE users` schema in `backend/internal/infrastructure/repository/sqlite_repository.go` (line 39)
- [x] 1.3 Add `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` to the end of `Migrate()` query in the same file
- [x] 1.4 Implement `GetByEmail` in `SQLiteRepository` (mirror of `GetByUsername` at line 248)

## Phase 2: Backend — Logic

- [x] 2.1 Update `UserService.Login` in `backend/internal/application/services/services.go:84` to detect `@` in identifier and dispatch to `GetByEmail` or `GetByUsername`
- [x] 2.2 Update `UserHandler.Login` in `backend/internal/infrastructure/handlers/handlers.go:124` to parse both `username` and `email` from body and validate exactly one is provided
- [x] 2.3 Map missing/ambiguous identifier to `400 Bad Request` with descriptive message in same handler

## Phase 3: Frontend

- [x] 3.1 Rename `username` state to `identifier` in `frontend/src/pages/LoginPage.tsx` for clarity
- [x] 3.2 Update placeholder text to `"Usuario o email"` in same file
- [x] 3.3 Update `handleLogin` to dispatch by `@` presence: send `{email, password}` or `{username, password}`

## Phase 4: Pre-deploy & Verification

- [x] 4.1 Run `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` in Turso console (user, manual)
- [x] 4.2 Local: start backend, run test plan #1-#7 with `curl` against `/login`
- [x] 4.3 Local: test frontend in browser, run test plan #8-#9 (login with `dangel` and with `admin@quiniela.com`)
- [ ] 4.4 Deploy: push to `master` → verify Render auto-deploys both services
- [ ] 4.5 Prod: re-run test plan #1-#2 against production URL

## Phase 5: Commit & Push

- [ ] 5.1 Commit with conventional message: `feat(auth): accept username or email for login`
- [ ] 5.2 Push to `master` (only after user confirms local is OK)
- [ ] 5.3 After prod verified, archive the change (sdd-archive phase)
