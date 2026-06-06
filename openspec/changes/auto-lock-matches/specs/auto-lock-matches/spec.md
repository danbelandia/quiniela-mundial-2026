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

### Requirement: Server-side enforcement on qualifier prediction submission

`PUT /groups/{group}/qualifier-predictions` MUST reject with HTTP 403 when any match in the group is effectively locked. The check MUST happen before any DB write.

#### Scenario: Qualifier submission to a group with one locked match is rejected

- GIVEN group A with match id=1 effectively locked (kickoff in 1h) and other 5 matches unlocked
- WHEN `PUT /groups/A/qualifier-predictions` is called with valid body
- THEN the response is `403 Forbidden` with body `group is locked`
- AND no row is written to `group_qualifier_predictions`

#### Scenario: Qualifier submission to a fully open group is accepted

- GIVEN group A with all 6 matches more than 3h away
- WHEN `PUT /groups/A/qualifier-predictions` is called with valid body
- THEN the response is `200 OK` (or 201) and the row is written

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
