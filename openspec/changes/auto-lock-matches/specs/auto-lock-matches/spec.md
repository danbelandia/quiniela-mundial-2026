# auto-lock-matches Specification

## Purpose

Define automatic, time-based locking of match predictions 3 hours before kickoff, plus server-side enforcement that closes today's frontend-bypassable gap. The lock is per-match, computed at read time from `match_date` and an optional manual admin override.

## Requirements

### Requirement: Effective lock state computation

The `Match` domain entity SHALL expose `IsEffectivelyLocked(now time.Time) bool` returning `true` when the manual `is_locked` flag is set OR when `now.Add(LOCK_WINDOW_HOURS)` is greater than or equal to `match_date`. The comparison MUST use UTC.

#### Scenario: Manual lock takes precedence

- GIVEN a match with `is_locked = true` and `match_date` 48h in the future
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true`

#### Scenario: Within the lock window

- GIVEN a match with `is_locked = false` and `match_date = now + 2h`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true` (3h - 2h = inside window)

#### Scenario: Outside the lock window

- GIVEN a match with `is_locked = false` and `match_date = now + 4h`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `false`

#### Scenario: Exactly at the window boundary

- GIVEN a match with `is_locked = false` and `match_date = now + 3h` (exact)
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true` (boundary inclusive: `>=`)

#### Scenario: Match already kicked off

- GIVEN a match with `is_locked = false` and `match_date = now - 1m`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true`

### Requirement: GET /matches returns effective lock state

The `GET /matches` endpoint MUST return each match with an `is_locked` field equal to `IsEffectivelyLocked(time.Now().UTC())` at the moment of the response. The raw DB column is read into the struct, then overwritten by the computed value before serialization.

#### Scenario: Read of in-window match returns locked

- GIVEN a match stored with `is_locked = false` and `match_date = now + 1h`
- WHEN the API serves `GET /matches`
- THEN the response includes `"is_locked": true` for that match

#### Scenario: Read of out-of-window match returns unlocked

- GIVEN a match stored with `is_locked = false` and `match_date = now + 12h`
- WHEN the API serves `GET /matches`
- THEN the response includes `"is_locked": false`

### Requirement: Server-side enforcement on prediction submission

`POST /predictions` MUST reject with HTTP 403 when the target match is effectively locked. The check MUST happen before any DB write, regardless of what the client sends in the body or the `is_locked` value the client saw in a previous response.

#### Scenario: Submission to a locked match is rejected

- GIVEN a match whose `match_date` is 1h from now
- WHEN `POST /predictions` is called with `match_id` of that match
- THEN the response is `403 Forbidden` with body `match is locked`
- AND no row is written to `predictions`

#### Scenario: Submission to an open match is accepted

- GIVEN a match whose `match_date` is 5h from now
- WHEN `POST /predictions` is called with `match_id` of that match and valid body
- THEN the response is `201 Created`
- AND a row is written to `predictions`

#### Scenario: Submission to a non-existent match is rejected

- GIVEN no match with the given `match_id`
- WHEN `POST /predictions` is called
- THEN the response is `404 Not Found`

### Requirement: Global qualifier lock deadline

Qualifier predictions (1st and 2nd place of each group) MUST close at a single global deadline: `2026-06-11T18:00:00Z` (06-11 14:00 CLT, 1 hour before the World Cup opener MEX-RSA). The deadline applies uniformly to all 12 groups regardless of when each group's first match is played. The deadline is configurable via the `QUALIFIER_LOCK_AT` environment variable. If unset, invalid, or unparseable, the default SHALL be `2026-06-11T18:00:00Z`.

#### Scenario: Before the global deadline all qualifiers are open

- GIVEN current time is `2026-06-10T20:00:00Z` (22 hours before deadline)
- WHEN `PUT /groups/L/qualifier-predictions` is called with valid body
- THEN the response is `200 OK` and the row is written
- AND the same holds for groups A through K

#### Scenario: After the global deadline all qualifiers are closed

- GIVEN current time is `2026-06-12T10:00:00Z` (16 hours after deadline)
- WHEN `PUT /groups/A/qualifier-predictions` is called with valid body
- THEN the response is `403 Forbidden` with body `match is locked`
- AND no row is written to `group_qualifier_predictions`

#### Scenario: Deadline is exact and inclusive

- GIVEN current time is `2026-06-11T18:00:00Z` (deadline exact)
- WHEN `PUT /groups/B/qualifier-predictions` is called
- THEN the response is `403 Forbidden` (boundary inclusive: `>=`)

#### Scenario: Custom deadline for testing

- GIVEN `QUALIFIER_LOCK_AT=2026-06-06T19:00:00Z` and current time is `2026-06-06T18:00:00Z`
- WHEN `PUT /groups/C/qualifier-predictions` is called
- THEN the response is `200 OK` (deadline is 1h in the future)
- AND restarting with `QUALIFIER_LOCK_AT=2026-06-06T17:00:00Z` makes the same call return `403`

#### Scenario: Invalid value falls back to default

- GIVEN `QUALIFIER_LOCK_AT=not-a-date`
- WHEN the API server starts
- THEN the effective deadline is `2026-06-11T18:00:00Z` and a warning is logged

### Requirement: Server-side enforcement on qualifier prediction submission

`PUT /groups/{group}/qualifier-predictions` MUST reject with HTTP 403 when the current time is at or after the global qualifier lock deadline. The check MUST happen before any DB write, regardless of the state of any individual match in the group.

#### Scenario: Qualifier submission before the deadline is accepted

- GIVEN current time is 24 hours before the global deadline
- WHEN `PUT /groups/A/qualifier-predictions` is called with valid body
- THEN the response is `200 OK` (or 201) and the row is written
- AND the same holds even if some matches in group A are already kicked off (the qualifier deadline is independent of match-by-match locks)

#### Scenario: Qualifier submission after the deadline is rejected

- GIVEN current time is 1 minute after the global deadline
- WHEN `PUT /groups/B/qualifier-predictions` is called
- THEN the response is `403 Forbidden` with body `match is locked`
- AND no row is written to `group_qualifier_predictions`

#### Scenario: Qualifier lock is independent of per-match locks

- GIVEN current time is 48 hours before the global deadline
- AND group A contains a match that is already locked (kickoff in 30 minutes)
- WHEN `PUT /groups/A/qualifier-predictions` is called
- THEN the response is `200 OK` (the per-match lock on one match does not block the group's qualifier)

### Requirement: Manual admin lock override

The endpoints `POST /admin/matches/lock` and `POST /admin/matches/lock-all` MUST continue to set the `is_locked` column in the DB. The effective lock MUST honor the manual flag (a manually locked match is locked regardless of time). The endpoints MUST NOT be removed in this change.

#### Scenario: Admin force-locks a match 24h before kickoff

- GIVEN a match with `is_locked = false` and `match_date = now + 24h`
- WHEN `POST /admin/matches/lock` is called with `{"id": X, "locked": true}`
- THEN the DB `is_locked` is set to `1`
- AND the next `GET /matches` response shows `"is_locked": true` for that match

### Requirement: Configurable lock window via environment variable

The system MUST read the lock window from the `LOCK_WINDOW_HOURS` environment variable. If unset, invalid, or non-positive, the default SHALL be `3` hours. The value is read once at server start.

#### Scenario: Default window is 3 hours

- GIVEN `LOCK_WINDOW_HOURS` is unset
- WHEN the API server starts
- THEN the effective lock window is `3h`

#### Scenario: Custom window for testing

- GIVEN `LOCK_WINDOW_HOURS=0.01` (~36 seconds)
- WHEN the API server starts
- THEN a match whose `match_date` is 30 seconds in the future is reported as locked

#### Scenario: Invalid value falls back to default

- GIVEN `LOCK_WINDOW_HOURS=banana`
- WHEN the API server starts
- THEN the effective lock window is `3h` and a warning is logged

### Requirement: Per-match evaluation

The lock state of one match MUST NOT depend on the state of any other match. Each match is evaluated independently against `now + LOCK_WINDOW_HOURS` and its own `is_locked` column.

#### Scenario: Matches in the same group lock independently

- GIVEN group A with match id=1 at `now + 1h` and match id=6 at `now + 48h`
- WHEN the API serves `GET /matches`
- THEN id=1 has `"is_locked": true` and id=6 has `"is_locked": false`

### Requirement: Config endpoint exposes global settings

The API MUST expose a `GET /config` endpoint that returns the global deadline timestamp and the match lock window. The frontend uses this to align its UI state with the server.

#### Scenario: /config returns the global settings

- GIVEN the API is running with default `QUALIFIER_LOCK_AT=2026-06-11T18:00:00Z` and `LOCK_WINDOW_HOURS=3`
- WHEN `GET /config` is called
- THEN the response is `200 OK` with body `{"qualifier_lock_at": "2026-06-11T18:00:00Z", "match_lock_window_hours": 3}`
