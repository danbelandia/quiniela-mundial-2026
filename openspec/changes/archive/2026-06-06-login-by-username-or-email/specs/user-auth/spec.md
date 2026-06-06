# Delta for user-auth

> Base spec: `openspec/specs/user-auth/spec.md` (does not exist yet — this delta will create it on archive).

## ADDED Requirements

### Requirement: Login accepts username or email as identifier

The `POST /login` endpoint SHALL accept exactly one of these body fields as the user identifier: `username` (string) or `email` (string). The endpoint MUST also accept a `password` (string) field. The system MUST authenticate the user against whichever identifier is provided and return the user's session info on success.

#### Scenario: Login with valid username

- GIVEN a registered user with `username="dangel"` and `email="admin@quiniela.com"`
- WHEN the client posts `{ "username": "dangel", "password": "281193" }` to `/login`
- THEN the response is `200 OK` with the user object (including `id` and `is_admin`)

#### Scenario: Login with valid email

- GIVEN a registered user with `username="dangel"` and `email="admin@quiniela.com"`
- WHEN the client posts `{ "email": "admin@quiniela.com", "password": "281193" }` to `/login`
- THEN the response is `200 OK` with the same user object as the username login

#### Scenario: Login with wrong password

- GIVEN a registered user
- WHEN the client posts either identifier with an incorrect `password`
- THEN the response is `401 Unauthorized` with a generic "invalid credentials" message (no leak of whether the user exists)

#### Scenario: Login with non-existent identifier

- GIVEN no registered user with the given identifier
- WHEN the client posts either identifier with any password
- THEN the response is `401 Unauthorized` with a generic "invalid credentials" message

#### Scenario: Login with empty body

- GIVEN no request body
- WHEN the client posts to `/login` with an empty body or missing both `username` and `email`
- THEN the response is `400 Bad Request` with a descriptive error

#### Scenario: Login with both username and email

- GIVEN the client sends both fields
- WHEN the client posts `{ "username": "a", "email": "b", "password": "x" }` to `/login`
- THEN the response is `400 Bad Request` (ambiguous request — exactly one identifier MUST be provided)

### Requirement: Email is unique in storage

The `users` table MUST enforce uniqueness on the `email` column. Two users SHALL NOT share the same email. The system MUST return a `409 Conflict` error with a descriptive message on registration attempt with an already-used email.

#### Scenario: Register with existing email

- GIVEN a user already registered with `email="foo@bar.com"`
- WHEN a new registration is submitted with `email="foo@bar.com"` and a different `username`
- THEN the response is `409 Conflict` with a message indicating the email is already registered

#### Scenario: Idempotent migration on existing database

- GIVEN a database table `users` was created before this change (without `email` UNIQUE)
- WHEN the backend starts and runs `Migrate()`
- THEN the migration MUST add the `idx_users_email` unique index without error
- AND MUST NOT drop or rewrite existing user data

### Requirement: Login form accepts username or email

The login page SHALL display a single identifier input with placeholder text indicating both options. The frontend MUST send `email` (not `username`) when the input value contains `@`, and `username` otherwise.

#### Scenario: User types an email

- GIVEN the user is on the login page
- WHEN the user types `admin@quiniela.com` in the identifier field and a password
- THEN the request body sent to `/login` contains `email` (not `username`)

#### Scenario: User types a username

- GIVEN the user is on the login page
- WHEN the user types `dangel` in the identifier field and a password
- THEN the request body sent to `/login` contains `username` (not `email`)

#### Scenario: Placeholder text is clear

- GIVEN the user is on the login page
- WHEN the login form is rendered
- THEN the identifier input's placeholder SHALL be `"Usuario o email"` (or equivalent clear bilingual/localized text)
