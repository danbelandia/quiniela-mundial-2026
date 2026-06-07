# Group Qualifier Prediction Specification

## Purpose

Predict which two teams finish **1st and 2nd** in each group. Correct predictions earn 3 points each (6 max per group). Points are awarded **only** when all 6 group matches are `status = "finished"`. Predictions lock in lockstep with match predictions.

## Requirements

### Requirement: Save Qualifier Prediction

The system SHALL allow an authenticated user to save or update their 1st/2nd prediction for a group.

#### Scenario: Valid prediction

- GIVEN the user is authenticated and the group has 4 teams
- WHEN the user submits two distinct teams from the group
- THEN the prediction is persisted and returned

#### Scenario: Validation errors

- GIVEN the user submits a prediction
- WHEN both positions reference the same team OR a team not in the group
- THEN the response is 400 and no row is persisted

#### Scenario: Update existing prediction

- GIVEN the user already has a prediction for the group
- WHEN the user submits a new prediction
- THEN the existing row is updated (one row per user+group)

### Requirement: Retrieve My Prediction

The system SHALL allow an authenticated user to retrieve their saved prediction for a group.

#### Scenario: Existing prediction

- GIVEN a saved prediction exists
- WHEN the user requests it
- THEN the response is 200 with the predicted teams

#### Scenario: No prediction

- GIVEN no saved prediction exists
- WHEN the user requests it
- THEN the response is 200 with an empty payload (NOT 404)

### Requirement: Synchronized Lock

The qualifier inputs SHALL be disabled when any match in the group has `is_locked = true`. The frontend MUST re-evaluate on group filter change.

#### Scenario: Lock state matches matches

- GIVEN at least one match in the group is `is_locked = true`
- WHEN the qualifier section renders
- THEN both inputs are `disabled`
- AND when the user switches to a group with no locked matches, both inputs become enabled with the saved prediction as the current value

### Requirement: FIFA Tie-Breaker Scoring

The system SHALL determine 1st and 2nd using FIFA tie-breakers in order: points, goal difference, goals scored, head-to-head points, head-to-head goal difference, head-to-head goals scored. If still tied, the system SHALL use a deterministic sortear (hash of match IDs).

#### Scenario: Tie broken by head-to-head

- GIVEN two teams have equal points, GD, and GF
- WHEN ranking is computed
- THEN more points in the head-to-head match ranks higher

#### Scenario: Persistent tie falls to sortear

- GIVEN all FIFA tie-breakers are exhausted
- WHEN ranking is computed
- THEN a deterministic hash of match IDs determines order (stable)

### Requirement: Points Awarded Only When Group Closed

The system SHALL NOT award qualifier points until all 6 matches of the group have `status = "finished"`.

#### Scenario: Group partially finished

- GIVEN 5 of 6 matches are `finished`
- WHEN ranking is computed
- THEN the user's qualifier points for that group are 0

#### Scenario: Group fully finished

- GIVEN all 6 matches are `finished`
- WHEN ranking is computed
- THEN the user's qualifier points are 3 per correct position

### Requirement: Ranking Integration

The ranking endpoint SHALL include each user's qualifier points and total (match-prediction + qualifier).

#### Scenario: User with correct prediction in closed group

- GIVEN a correct prediction and the group is closed
- WHEN ranking is fetched
- THEN `qualifier_points = 3` (or 6 if both correct) is included
- AND `total_points` reflects the sum

#### Scenario: User without prediction

- GIVEN the user has no qualifier prediction
- WHEN ranking is fetched
- THEN `qualifier_points = 0` and `total_points` equals match-prediction points only
