# group-standings Specification

## Purpose

Real-time per-group standings (FIFA 3/1/0) for the world cup, exposed via API and embedded in HomePage below each group's matches. Lets users see the "race" of each team as the admin finishes matches.

## Requirements

### Requirement: Standings endpoint returns one table per group

The `GET /standings` endpoint MUST return a JSON object with 12 groups (A through L), each containing exactly 4 teams with their stats (position, team name, flag, played, won, drawn, lost, goals_for, goals_against, goal_difference, points). The endpoint MUST be accessible to any authenticated user.

#### Scenario: All 12 groups returned

- GIVEN any authenticated user
- WHEN the client calls `GET /standings`
- THEN the response is `200 OK` with an array of 12 group standings
- AND each group has exactly 4 team rows

#### Scenario: Team with no finished matches

- GIVEN a group where all 6 matches have `status != "finished"`
- WHEN the client calls `GET /standings`
- THEN each of the 4 teams in that group appears with `played: 0`, `won: 0`, `drawn: 0`, `lost: 0`, `goals_for: 0`, `goals_against: 0`, `goal_difference: 0`, `points: 0`
- AND the order is alphabetical by team name

### Requirement: Points follow FIFA 3/1/0 rules and only finished matches count

A finished match (status `finished`) MUST award 3 points to the winner, 1 point to each team in a draw, and 0 points to the loser. A match whose status is not `finished` MUST NOT contribute to any team's stats. Standings MUST be ordered by points descending, then goal difference descending, then goals for descending, then team name ascending.

#### Scenario: Win awards 3 points

- GIVEN a finished match where the real result is 2-0
- WHEN the standings are calculated
- THEN the home team has 3 points and `won: 1`
- AND the away team has 0 points and `lost: 1`

#### Scenario: Draw awards 1 point to each team

- GIVEN a finished match where the real result is 1-1
- WHEN the standings are calculated
- THEN both teams have 1 point and `drawn: 1`

#### Scenario: Scheduled match is ignored

- GIVEN a match with `status: "scheduled"` and any scores
- WHEN the standings are calculated
- THEN neither team receives any points and `played` is not incremented

#### Scenario: Tiebreakers order teams

- GIVEN three teams with 6 points each, team X with GD +3 / GF 5, team Y with GD +3 / GF 4, team Z with GD +1
- WHEN the standings are displayed
- THEN team X is first, team Y is second, team Z is third

#### Scenario: All tiebreakers exhausted, fall back to alphabetical

- GIVEN two teams with identical points, goal difference, and goals for
- WHEN the standings are displayed
- THEN the team whose name comes first alphabetically is ranked higher

### Requirement: HomePage embeds the standings for the currently displayed group

The HomePage MUST render a `<GroupStandings groupName={currentGroup} />` component below the existing matches table. When the user navigates between groups (Anterior/Siguiente), the standings component MUST show the table for the new current group. The component MUST refetch data on the window's `focus` event, throttled to one refetch per 10 seconds.

#### Scenario: Standings appear below the matches

- GIVEN the user is on HomePage
- WHEN the page renders group A
- THEN a standings table for group A is visible below the matches table

#### Scenario: Standings follow group navigation

- GIVEN the user is viewing group A
- WHEN the user clicks "Siguiente"
- THEN the standings table now shows group B's teams

#### Scenario: Refetch on focus after admin update

- GIVEN the admin just finished a match and updated the real result
- WHEN the user returns to the HomePage tab (window focus)
- THEN the standings reflect the updated result within 10 seconds
