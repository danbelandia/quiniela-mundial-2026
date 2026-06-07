# User Qualifier Predictions Specification

## Purpose

Read-only public view of a user's 1°/2° group qualifier predictions. Surfaces the picks on `/user/:id/predictions` alongside the existing per-match score predictions, and — once a group is closed — shows the actual qualifiers with a per-position correctness hint.

## Requirements

### Requirement: Retrieve Qualifier Predictions By User

The system MUST expose `GET /users/{id}/qualifier-predictions` returning the array of `QualifierPrediction` rows for that user, ordered by `group_name` ascending.

#### Scenario: User with picks

- GIVEN a user who has saved qualifier predictions for groups A and B
- WHEN the endpoint is called
- THEN it returns 200 with a JSON array of length 2
- AND each element has `group_name`, `predicted_first`, `predicted_second`, plus the team flags for both picks

#### Scenario: User with no picks

- GIVEN a user with no rows in `group_qualifier_predictions`
- WHEN the endpoint is called
- THEN it returns 200 with `[]`

#### Scenario: Unknown user

- GIVEN a user id that does not exist
- WHEN the endpoint is called
- THEN it returns 404 with an error message

### Requirement: Enrich Each Entry With Team Flags

The response entries MUST include the flag emoji of both the predicted first and the predicted second teams so the frontend can render the picks without a second round trip.

#### Scenario: Flags populated

- GIVEN a pick `predicted_first = "México"` and `predicted_second = "Corea del Sur"`
- WHEN the response is returned
- THEN `predicted_first_flag` is `🇲🇽` and `predicted_second_flag` is `🇰🇷`

#### Scenario: Unknown team name

- GIVEN a pick referencing a team name not present in the teams table
- THEN the flag field is an empty string (not null, not error)

### Requirement: Include Actual Qualifiers When Group Is Closed

The system MUST include `actual_first` and `actual_second` in the response entry ONLY when the group is closed (all 6 of the group's matches have `status = "finished"`), computed by the same FIFA tie-breaker rules used by the scoring service.

#### Scenario: Group not yet closed

- GIVEN at least one match in the group is still `status = "scheduled"`
- WHEN the response is returned
- THEN `actual_first` and `actual_second` are empty strings

#### Scenario: Group closed

- GIVEN all 6 group matches have `status = "finished"`
- WHEN the response is returned
- THEN `actual_first` and `actual_second` hold the team names of the 1st and 2nd place teams (with their flags)
- AND a `points_earned` field holds 0, 3, or 6 based on the user's picks

### Requirement: Render Qualifier Card Per Group On UserPredictionsPage

The frontend MUST render a card for each group the user has a qualifier prediction for, displayed inside the corresponding group tab above the matches table.

#### Scenario: User has picks for the active group

- GIVEN the user has a qualifier prediction for the active group
- WHEN the user opens `/user/:id/predictions` and selects that group tab
- THEN a card is rendered showing 1° and 2° picks with flags

#### Scenario: Group still in progress

- GIVEN the qualifier card is rendered for a group that is not closed
- THEN the card shows the user's picks only (no actual qualifiers, no correctness hints)

#### Scenario: Group closed

- GIVEN the group is closed and the user has a pick
- THEN the card also shows the actual 1°/2° teams with flags
- AND a green check or red X marks each position based on correctness
- AND the earned points are shown

### Requirement: Graceful Empty State

The frontend MUST render a placeholder when the user has no qualifier prediction for the active group, and MUST NOT block the per-match score predictions table.

#### Scenario: Group with no qualifier pick

- GIVEN the user has match score predictions for the group but no qualifier prediction
- THEN the card area shows a muted "Sin pronóstico de clasificación" message
- AND the matches table below renders normally

## Out of Scope

- Writing or editing qualifier predictions from this page (still done from the home page).
- Showing other users' qualifier predictions on the public ranking page.
- Auth / authorization — the endpoint is public read-only, same as the existing `/users/{id}/predictions` endpoint.
