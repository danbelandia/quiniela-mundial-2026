# auto-lock-matches Specification (Delta)

## MODIFIED Requirements

### Requirement: Effective lock state computation

The `Match` domain entity SHALL expose `IsEffectivelyLocked(now time.Time) bool` returning `true` when the manual `is_locked` flag is set OR when `now.Add(LOCK_WINDOW_HOURS)` is greater than or equal to `match_date`. The comparison MUST use UTC. The operational value of `LOCK_WINDOW_HOURS` in production is `0.25` (15 minutes).

#### Scenario: Manual lock takes precedence

- GIVEN a match with `is_locked = true` and `match_date` 48h in the future
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true`

#### Scenario: Within the lock window (15 min)

- GIVEN a match with `is_locked = false` and `match_date = now + 10min`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true` (15min window - 10min to kickoff = inside window)

#### Scenario: Outside the lock window (15 min)

- GIVEN a match with `is_locked = false` and `match_date = now + 1h`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `false`

#### Scenario: Exactly at the 15-minute boundary

- GIVEN a match with `is_locked = false` and `match_date = now + 15min` (exact)
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true` (boundary inclusive: `>=`)

#### Scenario: Match already kicked off

- GIVEN a match with `is_locked = false` and `match_date = now - 1m`
- WHEN `IsEffectivelyLocked(now)` is called
- THEN it returns `true`

### Requirement: GET /matches returns effective lock state

The `GET /matches` endpoint MUST return each match with an `is_locked` field equal to `IsEffectivelyLocked(time.Now().UTC())` at the moment of the response. The raw DB column is read into the struct, then overwritten by the computed value before serialization.

#### Scenario: Read of in-window match (15 min) returns locked

- GIVEN a match stored with `is_locked = false` and `match_date = now + 5min`
- WHEN the API serves `GET /matches`
- THEN the response includes `"is_locked": true` for that match

#### Scenario: Read of out-of-window match (15 min) returns unlocked

- GIVEN a match stored with `is_locked = false` and `match_date = now + 2h`
- WHEN the API serves `GET /matches`
- THEN the response includes `"is_locked": false`

### Requirement: Server-side enforcement on prediction submission

`POST /predictions` MUST reject with HTTP 403 when the target match is effectively locked. The check MUST happen before any DB write, regardless of what the client sends in the body or the `is_locked` value the client saw in a previous response.

#### Scenario: Submission to a locked match is rejected

- GIVEN a match whose `match_date` is 5 minutes from now (inside the 15-minute window)
- WHEN `POST /predictions` is called with `match_id` of that match
- THEN the response is `403 Forbidden` with body `match is locked`
- AND no row is written to `predictions`

#### Scenario: Submission to an open match is accepted

- GIVEN a match whose `match_date` is 1 hour from now (outside the 15-minute window)
- WHEN `POST /predictions` is called with `match_id` of that match and valid body
- THEN the response is `201 Created`
- AND a row is written to `predictions`

#### Scenario: Submission to a non-existent match is rejected

- GIVEN no match with the given `match_id`
- WHEN `POST /predictions` is called
- THEN the response is `404 Not Found`

### Requirement: Configurable lock window via environment variable

The system MUST read the lock window from the `LOCK_WINDOW_HOURS` environment variable. If unset, invalid, or non-positive, the default SHALL be `3` hours (safe fallback). The value is read once at server start. The operational value in production is `0.25` (15 minutes).

#### Scenario: Default window is 3 hours

- GIVEN `LOCK_WINDOW_HOURS` is unset
- WHEN the API server starts
- THEN the effective lock window is `3h`

#### Scenario: Production window is 15 minutes

- GIVEN `LOCK_WINDOW_HOURS=0.25`
- WHEN the API server starts
- THEN the effective lock window is `15m`
- AND a match whose `match_date` is 14 minutes in the future is reported as locked
- AND a match whose `match_date` is 16 minutes in the future is reported as unlocked

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

- GIVEN group A with match id=1 at `now + 5min` and match id=6 at `now + 48h`
- WHEN the API serves `GET /matches`
- THEN id=1 has `"is_locked": true` and id=6 has `"is_locked": false`

### Requirement: Config endpoint exposes global settings

The API MUST expose a `GET /config` endpoint that returns the global deadline timestamp and the match lock window. The frontend uses this to align its UI state with the server.

#### Scenario: /config returns the production settings

- GIVEN the API is running with default `QUALIFIER_LOCK_AT=2026-06-11T18:00:00Z` and `LOCK_WINDOW_HOURS=0.25`
- WHEN `GET /config` is called
- THEN the response is `200 OK` with body `{"qualifier_lock_at": "2026-06-11T18:00:00Z", "match_lock_window_hours": 0.25}`

## REMOVED Scenarios

The following scenarios from the original spec described the **3-hour** window. They are kept as historical reference in the archive but the active spec describes the **15-minute** operational window above. The mechanism is unchanged; only the operational value moved.

- (Original) "Within the lock window" with `match_date = now + 2h` — superseded by the 15-minute equivalent above.
- (Original) "Outside the lock window" with `match_date = now + 4h` — superseded.
- (Original) "Exactly at the window boundary" with `match_date = now + 3h` — superseded.
- (Original) "Read of in-window match returns locked" with `match_date = now + 1h` — superseded.
- (Original) "Read of out-of-window match returns unlocked" with `match_date = now + 12h` — superseded.
- (Original) "Submission to a locked match is rejected" with `match_date` 1h from now — superseded.
- (Original) "Submission to an open match is accepted" with `match_date` 5h from now — superseded.
- (Original) "Matches in the same group lock independently" with `now + 1h` and `now + 48h` — superseded by the 15-minute version above.
- (Original) "/config returns the global settings" example with `match_lock_window_hours: 3` — superseded.
