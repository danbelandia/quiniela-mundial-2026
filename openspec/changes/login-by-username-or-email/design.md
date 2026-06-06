# Design: Login con username o email

## Technical Approach

Habilitar dos formas de identificar al usuario en `POST /login`: `username` o `email`. El handler acepta ambos campos en el body, valida que venga exactamente uno, y delega al service. El service detecta si el identificador es email (contiene `@`) y despacha al método de repo correspondiente. El frontend decide qué campo mandar según el contenido del input. La unicidad de email se enforce a nivel DB con un índice único idempotente.

## Architecture Decisions

### Decision: Detección de tipo de credencial en el service (no en el handler)

**Choice**: El handler pasa el identificador tal cual al service; el service decide si llamar `GetByUsername` o `GetByEmail` según si contiene `@`.
**Alternatives considered**:
- Detectar en el handler y llamar dos métodos diferentes del service (`LoginByUsername` / `LoginByEmail`).
- Detectar en el frontend y mandar `{kind, value, password}` con un discriminador.
**Rationale**: El handler queda thin (parsea JSON, llama service, mapea error→HTTP). El service encapsula la regla de negocio. El frontend no necesita conocer el "kind" — solo envía el campo que corresponda.

### Decision: `email` validado por índice UNIQUE, no por constraint en el `CREATE TABLE`

**Choice**: Agregar `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);` al final de `Migrate()`, además de `UNIQUE` en el `CREATE TABLE`.
**Alternatives considered**:
- Solo agregar `UNIQUE` en el `CREATE TABLE` (no aplica a DBs existentes porque `IF NOT EXISTS` no migra).
- Hacer un ALTER TABLE separado en una función de migración versionada.
**Rationale**: SQLite no soporta `ALTER TABLE ADD CONSTRAINT`. El índice único tiene el mismo efecto de unicidad. `IF NOT EXISTS` lo hace idempotente — corre cada vez que el backend arranca y no rompe DBs nuevas ni existentes. La DB de Turso se actualiza con un `CREATE UNIQUE INDEX` manual (que ya está confirmado por el usuario que no tiene duplicados).

### Decision: Validación de "exactamente un identificador" en el handler

**Choice**: Si el body trae `username` Y `email` ambos con valor → 400. Si trae ambos vacíos → 400. Si trae exactamente uno → procede.
**Rationale**: Evita ambigüedad. Si mandan ambos, el service no sabe cuál priorizar.

### Decision: Frontend decide qué campo mandar (no discriminador en el body)

**Choice**: El LoginPage evalúa `value.includes('@')` y manda `{email, password}` o `{username, password}`.
**Alternatives considered**:
- Siempre mandar `{username: value, password}` y que el backend detecte.
- Un toggle radio button "Usuario" / "Email".
**Rationale**: Coherente con REST (el cliente declara qué tipo de credencial manda). El backend igual valida por `@` como segunda capa de defensa.

## Data Flow

```
User types "admin@foo.com" + pwd
        │
        ▼
LoginPage (frontend)
        │  includes('@') === true
        ▼
POST /login  body: { email, password }
        │
        ▼
UserHandler.Login  (handlers.go)
        │  parses JSON, validates exactly one identifier
        ▼
UserService.Login(identifier, password)
        │  identifier contains '@' → GetByEmail
        ▼
SQLiteRepository.GetByEmail(email)
        │
        ▼
DB  →  user  →  return 200 + user JSON
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/core/ports/repositories.go` | Modify | Agregar `GetByEmail(email string) (*domain.User, error)` al port |
| `backend/internal/infrastructure/repository/sqlite_repository.go` | Modify | (1) `email TEXT NOT NULL UNIQUE` en schema. (2) `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email` en Migrate. (3) Implementar `GetByEmail` |
| `backend/internal/application/services/services.go` | Modify | `Login` detecta `@`, despacha a `GetByEmail` o `GetByUsername` |
| `backend/internal/infrastructure/handlers/handlers.go` | Modify | `Login` lee `username` + `email` del body, valida exactamente uno |
| `frontend/src/pages/LoginPage.tsx` | Modify | Placeholder + lógica de envío según `@` |

## Interfaces / Contracts

```go
// ports/repositories.go
type UserRepository interface {
    // ... existing methods
    GetByEmail(email string) (*domain.User, error)
}

// services.go
func (s *UserService) Login(identifier, password string) (*domain.User, error) {
    if identifier == "" {
        return nil, errors.New("identifier is required")
    }
    var u *domain.User
    var err error
    if strings.Contains(identifier, "@") {
        u, err = s.repo.GetByEmail(identifier)
    } else {
        u, err = s.repo.GetByUsername(identifier)
    }
    if err != nil { return nil, err }
    if u.Password != password {
        return nil, errors.New("invalid credentials")
    }
    return u, nil
}
```

```typescript
// LoginPage.tsx
const handleLogin = async () => {
  const isEmail = identifier.includes('@');
  const body = isEmail
    ? { email: identifier, password }
    : { username: identifier, password };
  const user = await apiClient.post('/login', body);
  // ...
};
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | — | N/A (no test runner configured, `strict_tdd: false`) |
| Integration | Login flow | Manual via `curl` o el frontend |
| E2E | Login UI con user, con email, con password incorrecta | Manual en navegador |

**Manual test plan** (a ejecutar en dev y contra Turso en prod):

1. `POST /login {"username":"dangel","password":"281193"}` → 200 con user
2. `POST /login {"email":"admin@quiniela.com","password":"281193"}` → 200 con mismo user
3. `POST /login {"username":"dangel","password":"wrong"}` → 401
4. `POST /login {"email":"admin@quiniela.com","password":"wrong"}` → 401
5. `POST /login {"email":"nobody@foo.com","password":"x"}` → 401
6. `POST /login {"username":"x","email":"y","password":"z"}` → 400
7. `POST /login {}` → 400
8. Frontend: tipear `dangel` + pwd → loguea
9. Frontend: tipear `admin@quiniela.com` + pwd → loguea

## Migration / Rollout

1. **Pre-deploy (manual, ~1 min)**: en la consola de Turso correr `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);`. Usuario ya confirmó 0 duplicados.
2. **Deploy backend**: push a `master` → Render auto-deploy (o manual desde dashboard).
3. **Deploy frontend**: push a `master` → Render Static Site auto-rebuild.
4. **Verificar**: correr test plan manual #1-#7 contra prod, y #8-#9 en el navegador.

No hay flag feature. El cambio es retrocompatible: el viejo `{username, password}` sigue funcionando.

## Open Questions

- Ninguna. La viabilidad ya fue confirmada y los criterios de éxito están claros.
