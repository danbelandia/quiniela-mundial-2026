# top-scorer-prediction Specification

## Purpose

Allow each user to predict the top scorer of the World Cup 2026 group stage by picking one player from a curated static candidate list. The user earns 6 points if their pick matches the actual top scorer (case-insensitive exact-name match) when the group stage closes. Ties in the actual top-scorer count are resolved at the admin level: only the admin-set name counts as correct. The prediction closes at the same global deadline as the 1st/2nd qualifier predictions (`2026-06-11T18:00:00Z` = 1h before the World Cup opener).

## Requirements

### Requirement: Static candidate list of players

The system MUST expose a fixed list of candidate players that users can pick from. The list is loaded at server startup from `backend/data/top_scorer_candidates.json` and served to the frontend via the existing `GET /config` endpoint (extended) OR a dedicated `GET /top-scorer/candidates` endpoint. The list contains at least one entry per `{name, team, flag}` triple.

#### Scenario: Frontend fetches the candidate list

- GIVEN the server is running with the default candidate JSON
- WHEN `GET /top-scorer/candidates` (or `GET /config`) is called
- THEN the response includes a `top_scorer_candidates` array of at least 40 entries
- AND each entry has `name`, `team`, and `flag` fields

#### Scenario: Empty or missing candidate file is handled gracefully

- GIVEN the candidate JSON file is missing
- WHEN the server starts
- THEN the server logs a warning and the candidates endpoint returns an empty array
- AND the rest of the server still works (no crash)

### Requirement: User saves / updates their top-scorer prediction

Each authenticated user can save or update their own top-scorer prediction at any time before the global lock deadline. There is at most one prediction per user; saving again overwrites the previous value.

#### Scenario: User saves a pick for the first time

- GIVEN a logged-in user with no prior top-scorer prediction
- WHEN `PUT /top-scorer-prediction/me` is called with body `{"user_id": <id>, "predicted_player": "Kylian Mbappé"}`
- THEN the response is `200 OK` (or 201) and a row is written to `top_scorer_predictions`

#### Scenario: User changes their pick before the deadline

- GIVEN a user with an existing prediction for "Harry Kane"
- WHEN `PUT /top-scorer-prediction/me` is called with `{"user_id": <id>, "predicted_player": "Erling Haaland"}`
- THEN the existing row is updated to "Erling Haaland" (UPSERT semantics)
- AND the response is `200 OK`

#### Scenario: User picks a name not in the candidate list

- GIVEN the candidate list contains "Kylian Mbappé"
- WHEN a user PUTs `{"predicted_player": "Some Random Player"}`
- THEN the request is rejected with `400 Bad Request` and body `player not in candidate list`

### Requirement: Lock enforcement at the global deadline

The top-scorer prediction endpoint MUST reject updates with HTTP 403 when the current time is at or after the global `QUALIFIER_LOCK_AT` timestamp (reused, not a separate env var). The check MUST happen before any DB write.

#### Scenario: Update before the deadline is accepted

- GIVEN current time is 24 hours before `QUALIFIER_LOCK_AT`
- WHEN `PUT /top-scorer-prediction/me` is called
- THEN the response is `200 OK` and the row is written

#### Scenario: Update after the deadline is rejected

- GIVEN current time is 1 minute after `QUALIFIER_LOCK_AT`
- WHEN `PUT /top-scorer-prediction/me` is called
- THEN the response is `403 Forbidden` with body `match is locked`
- AND no row is written

#### Scenario: Update exactly at the deadline is rejected (boundary inclusive)

- GIVEN current time equals `QUALIFIER_LOCK_AT` exactly
- WHEN `PUT /top-scorer-prediction/me` is called
- THEN the response is `403 Forbidden` (boundary inclusive: `>=`)

### Requirement: Read endpoints

The system MUST expose:
- `GET /top-scorer-prediction/me?user_id=<id>` — returns the current user's pick (or 404 if none)
- `GET /users/{id}/top-scorer-prediction` — returns any user's pick (publicly visible to any logged-in user, like match predictions)
- `GET /top-scorer/candidates` — returns the static candidate list (see Requirement 1)

#### Scenario: User has a pick

- GIVEN user 1 has predicted "Kylian Mbappé"
- WHEN `GET /users/1/top-scorer-prediction` is called by any logged-in user
- THEN the response is `200 OK` with body `{"user_id": 1, "predicted_player": "Kylian Mbappé", "created_at": "...", "updated_at": "..."}`

#### Scenario: User has no pick

- GIVEN user 1 has never called PUT
- WHEN `GET /users/1/top-scorer-prediction` is called
- THEN the response is `404 Not Found`

#### Scenario: Reading another user's pick is allowed

- GIVEN user 2 (non-admin) is logged in
- AND user 1 has a pick
- WHEN user 2 calls `GET /users/1/top-scorer-prediction`
- THEN the response is `200 OK` (any authenticated user can read)

### Requirement: Admin sets the actual top scorer

The system MUST expose `POST /admin/top-scorer` that sets the actual top scorer of the group stage. The body is `{"player": "Kylian Mbappé"}`. The admin can submit any string; the system MUST warn (log) if the name is not in the candidate list but still accept the submission (so the admin can record a dark-horse winner). The actual top scorer is global (single value, not per user).

#### Scenario: Admin sets the actual top scorer

- GIVEN the group stage is closed
- WHEN `POST /admin/top-scorer` is called with `{"player": "Kylian Mbappé"}`
- THEN the actual is stored (in `app_config` or a single-row table)
- AND subsequent `GET /ranking` calls reflect the 6 points for matching users

#### Scenario: Admin sets a name not in the candidate list

- GIVEN the candidate list contains only the 50 known players
- WHEN admin calls `POST /admin/top-scorer` with `{"player": "Hansi Flick"}` (a coach, not a player — defensive check)
- THEN the system stores the value
- AND logs a warning `top scorer not in candidate list: Hansi Flick`
- AND the endpoint returns `200 OK` (admin override)

### Requirement: Scoring: 6 points for correct pick

Once the admin has set the actual top scorer, every user whose `predicted_player` matches the actual (case-insensitive, exact-name match) MUST receive 6 points in the `top_scorer_score` field of the `/ranking` response. All other users receive 0.

#### Scenario: User picked the actual top scorer

- GIVEN admin set actual = "Kylian Mbappé"
- AND user 1 predicted "Kylian Mbappé"
- AND user 2 predicted "Harry Kane"
- WHEN `GET /ranking` is called
- THEN the response includes `top_scorer_score: 6` for user 1
- AND `top_scorer_score: 0` for user 2

#### Scenario: Match is case-insensitive

- GIVEN admin set actual = "Kylian Mbappé" (capital M, accent)
- AND user 1 predicted "kylian mbappe" (lowercase, no accent)
- WHEN `GET /ranking` is called
- THEN user 1 receives 6 points (case-insensitive + accent-insensitive comparison)

#### Scenario: Admin has not set the actual yet

- GIVEN no `POST /admin/top-scorer` has been called
- WHEN `GET /ranking` is called
- THEN all users receive `top_scorer_score: 0`

#### Scenario: User has no prediction

- GIVEN user 1 has no row in `top_scorer_predictions`
- AND admin has set the actual
- WHEN `GET /ranking` is called
- THEN user 1 receives `top_scorer_score: 0` (not 6)

### Requirement: Frontend page `/goleador`

The frontend MUST expose a page at `/goleador` that shows the candidate list as a dropdown. The user picks one, saves it, and sees their current pick with a "Cambiar" button. The page MUST show the global lock state (locked after `qualifier_lock_at`).

#### Scenario: User picks a player

- GIVEN a logged-in user on `/goleador` with no prior pick
- WHEN the user selects "Kylian Mbappé" from the dropdown and clicks "Guardar"
- THEN the pick is saved and the UI shows "Tu pick: 🇫🇷 Kylian Mbappé — Francia"
- AND the "Guardar" button changes to "Cambiar"

#### Scenario: Page is locked

- GIVEN current time is after `qualifier_lock_at`
- WHEN the user navigates to `/goleador`
- THEN the dropdown is disabled
- AND a lock banner is shown: "El plazo para pronosticar al goleador cerró el 11/06 a las 14:00 CLT"

### Requirement: Show pick on rival user profile

The `UserPredictionsPage` MUST show the target user's top-scorer pick as a card, positioned ABOVE the qualifier prediction card and the per-match table.

#### Scenario: Viewing a rival's profile

- GIVEN a logged-in user navigates to `/user/<rival_id>/predictions`
- AND the rival has a top-scorer pick
- WHEN the page renders
- THEN a card titled "Goleador fase de grupos" appears at the top, showing the rival's pick

#### Scenario: Rival has no pick yet

- GIVEN a rival has no top-scorer pick
- WHEN viewing their profile
- THEN the card shows a placeholder "Aún no pronosticó"

### Requirement: Admin can set the actual top scorer

The `AdminPage` MUST include a new section titled "Goleador de la fase de grupos" with a dropdown (same candidate list as the user page) and a "Guardar resultado" button. Optionally, a free-text input is available for names not in the list (with a warning).

#### Scenario: Admin sets the actual

- GIVEN an admin is on the admin page
- WHEN they select "Kylian Mbappé" from the dropdown and click "Guardar"
- THEN the actual is stored
- AND the ranking recalculates the `top_scorer_score` for all users

### Requirement: Ranking page shows the new column

The `/ranking` endpoint MUST include a `top_scorer_score` field per user. The `RankingPage` MUST show this as a new column in the table, with the same styling as `qualifier_score`.

#### Scenario: Ranking table has the new column

- GIVEN `GET /ranking` returns the standard user objects
- WHEN the page renders
- THEN a new column "Goleador" appears between "Clasificación" and "Total"
- AND each row shows the user's `top_scorer_score` (0, 6, etc.)

### Requirement: Lock state is shared with the existing global deadline

The top-scorer prediction MUST close at the same instant as the 1st/2nd qualifier prediction. There is no separate env var or config; both use `QUALIFIER_LOCK_AT`.

#### Scenario: Same lock deadline

- GIVEN `QUALIFIER_LOCK_AT=2026-06-11T18:00:00Z`
- WHEN the server is running
- THEN both `PUT /top-scorer-prediction/me` and `PUT /groups/{group}/qualifier-predictions` reject updates at or after that instant
- AND both accept updates before that instant
