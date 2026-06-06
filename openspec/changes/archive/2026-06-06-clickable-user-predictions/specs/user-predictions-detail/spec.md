# Delta for user-predictions-detail

> Base spec: `openspec/specs/user-predictions-detail/spec.md` (does not exist yet — this delta will create it on archive).

## ADDED Requirements

### Requirement: User predictions endpoint returns match info and points

The `GET /users/{id}/predictions` endpoint SHALL return a JSON object with the user's profile (id, username) and an array of predictions. Each prediction MUST include the match info (id, group_name, home_team, home_flag, away_team, away_flag, match_date, status, real home/away score) plus the user's predicted scores and the points earned for that match.

#### Scenario: User has predictions for all matches

- GIVEN a user with 72 predictions
- WHEN `GET /users/{id}/predictions` is called with that user's id
- THEN the response is `200 OK` with the user object and an array of 72 predictions
- AND each prediction has the match info, predicted scores, and points (0, 1, 2, or 3)

#### Scenario: User has no predictions

- GIVEN a user with zero predictions
- WHEN `GET /users/{id}/predictions` is called
- THEN the response is `200 OK` with the user object and an array of all 72 matches
- AND every match has `has_prediction: false`, `home_score_pred: 0`, `away_score_pred: 0`, `points: 0`

#### Scenario: User does not exist

- GIVEN no user with the given id
- WHEN `GET /users/{id}/predictions` is called
- THEN the response is `404 Not Found`

#### Scenario: Points are calculated correctly

- GIVEN a finished match where the real result is 2-1
- WHEN a user predicted 2-1 (exact)
- THEN the prediction's `points` field SHALL be `3`

- GIVEN a finished match where the real result is 2-1
- WHEN a user predicted 1-0 (correct winner, wrong score)
- THEN the prediction's `points` field SHALL be `2`

- GIVEN a finished match where the real result is 1-1
- WHEN a user predicted 0-0 (draw predicted, draw real, wrong score)
- THEN the prediction's `points` field SHALL be `1`

- GIVEN a finished match where the real result is 1-0
- WHEN a user predicted 2-1 (wrong winner)
- THEN the prediction's `points` field SHALL be `0`

- GIVEN a match with `status != "finished"`
- WHEN the response is built
- THEN the prediction's `points` field SHALL be `0` (or null)

### Requirement: Predictions are publicly visible to any authenticated user

The endpoint MUST be accessible to any authenticated user (logged in). The system MUST NOT restrict access to the prediction owner. Admin status is NOT required.

#### Scenario: Authenticated user A views user B's predictions

- GIVEN user A is logged in
- WHEN user A calls `GET /users/{B.id}/predictions`
- THEN the response is `200 OK` with user B's predictions

#### Scenario: Unauthenticated request is rejected

- GIVEN no authenticated session
- WHEN `GET /users/{id}/predictions` is called
- THEN the response is `401 Unauthorized` or `403 Forbidden`

### Requirement: Ranking rows are clickable links to user detail

In the ranking page, each user's username MUST be a clickable link that navigates to `/user/{user.id}`. The rest of the row MAY also be clickable for usability. The link SHALL NOT change the visual style of the ranking table.

#### Scenario: Click on username navigates to detail

- GIVEN the ranking page is rendered
- WHEN the user clicks on a username
- THEN the browser navigates to `/user/{user.id}`

#### Scenario: Cursor indicates clickability

- GIVEN the ranking page is rendered
- WHEN the user hovers over a username
- THEN the cursor SHALL be a pointer

### Requirement: User detail page groups predictions by group

The `/user/:id` page MUST display the target user's predictions grouped by `group_name` (A, B, C... L). Each group SHALL be presented as a tab or section. Each match within a group MUST display: teams with flags, real result, user's prediction, and points.

#### Scenario: Page shows 12 groups

- GIVEN a user with predictions across all 12 groups
- WHEN the detail page is rendered
- THEN 12 group tabs are visible (A through L)
- AND each tab shows the matches of that group with predictions and points

#### Scenario: Group with no matches is hidden or empty

- GIVEN a user with no matches in group G
- WHEN the detail page is rendered
- THEN group G tab either is hidden or shows an empty state

#### Scenario: Match with no prediction is shown with empty prediction

- GIVEN a user with no prediction for a specific match
- WHEN the detail page renders that match
- THEN the "prediction" column SHALL show "—" or "Sin pronóstico"

#### Scenario: Back button returns to ranking

- GIVEN the user is on `/user/{id}`
- WHEN the user clicks "Volver al ranking"
- THEN the browser navigates back to `/ranking`
