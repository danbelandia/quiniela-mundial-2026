# ranking-tie-breaker Specification

## Purpose

Define a deterministic, merit-based ordering for the public ranking when two or more users are tied on total points. The criteria reward prediction accuracy: more exact-score predictions outrank fewer, and more winner-correct predictions outrank fewer. The final fallback is the user who registered first (lower `id`).

## Requirements

### Requirement: Ranking comparator with tie-breaker

The `GET /ranking` response MUST be ordered by the following criteria, applied in sequence until a difference is found:

1. `total` (sum of `score + qualifier_score + top_scorer_score`) — descending
2. `exact_score` (count of predictions that earned 3 points) — descending
3. `winner_score` (count of predictions that earned 2 points) — descending
4. `id` — ascending (earlier registered user wins)

The comparator MUST be stable: rows whose criteria are all equal keep their relative order from the backend response. The visible UI MUST NOT change (no new columns) — only the row order changes when ties exist.

#### Scenario: Higher total wins

- GIVEN user A with `total = 60` and user B with `total = 55`
- WHEN `GET /ranking` is called
- THEN user A is listed before user B

#### Scenario: Tie on total — more exact scores wins

- GIVEN user A with `total = 50, exact_score = 6, winner_score = 4`
- AND user B with `total = 50, exact_score = 4, winner_score = 6`
- WHEN `GET /ranking` is called
- THEN user A is listed before user B (more exact scores)

#### Scenario: Tie on total and exact — more winner scores wins

- GIVEN user A with `total = 50, exact_score = 4, winner_score = 6`
- AND user B with `total = 50, exact_score = 4, winner_score = 5`
- WHEN `GET /ranking` is called
- THEN user A is listed before user B (more winner scores)

#### Scenario: Full tie — lower id wins

- GIVEN user A with `id = 1, total = 50, exact_score = 4, winner_score = 5`
- AND user B with `id = 2, total = 50, exact_score = 4, winner_score = 5`
- WHEN `GET /ranking` is called
- THEN user A is listed before user B (registered first)

#### Scenario: No tie — comparator only affects equal totals

- GIVEN user A with `total = 30` and user B with `total = 20`
- AND user A has fewer exact scores than user B
- WHEN `GET /ranking` is called
- THEN user A is still listed before user B (total dominates)

### Requirement: User struct exposes tie-breaker counters

The `User` domain entity MUST expose `ExactScore int` and `WinnerScore int` fields, both JSON-serialized as `exact_score` and `winner_score` respectively. Both MUST be `0` for a user with no finished predictions.

#### Scenario: New user has zero counters

- GIVEN a newly registered user with no predictions
- WHEN `GET /ranking` is called
- THEN that user has `exact_score = 0` and `winner_score = 0`

#### Scenario: User with all exact predictions

- GIVEN a user whose only finished predictions are 5 exact scores
- WHEN `GET /ranking` is called
- THEN `exact_score = 5` and `winner_score = 0`

### Requirement: Counters are computed in the ranking handler, not stored

The counters MUST be computed on each `GET /ranking` call by iterating the user's predictions, not persisted to the database. Adding a finished match result MUST update the counters on the next call without any DB migration.

#### Scenario: New finished match updates counters

- GIVEN a user with `score = 6` (from 2 exact predictions) before a match
- AND the user had predicted the next match exactly
- WHEN the admin sets the next match result and `GET /ranking` is called
- THEN that user's `score = 9` and `exact_score = 3`
