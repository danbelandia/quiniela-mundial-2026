# Tasks: Standings por grupo embebidos en HomePage

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~240 (4 modified, 2 new) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend + frontend embed | PR 1 | Single PR, total <300 lines |

## Phase 1: Backend — Foundation (domain)

- [x] 1.1 Create `backend/internal/core/domain/standings.go` with `TeamStanding` and `GroupStanding` structs
- [x] 1.2 Add pure `CalculateStandings(matches []Match) []GroupStanding`: group by `Match.Group`, skip `status != "finished"`, compute PJ/G/E/P/GF/GC/Dif/Pts
- [x] 1.3 In same function, sort teams per group: Pts desc, GD desc, GF desc, name asc; assign `Position` 1..4
- [x] 1.4 Edge case: 0 finished matches → all 4 teams with zeros, alphabetical, position 1..4

## Phase 2: Backend — Service + Handler + Route

- [x] 2.1 In `services.go`, add `StandingsService` + `NewStandingsService(repo)` with `matchRepo ports.MatchRepository`
- [x] 2.2 Add `StandingsService.CalculateAll() ([]GroupStanding, error)` calling `matchRepo.GetAllMatches()` + `domain.CalculateStandings`
- [x] 2.3 In `handlers.go`, add `StandingsHandler` with `GetAll(w, r)` method (mirror `MatchHandler.GetMatches` pattern with `SetCORS`)
- [x] 2.4 In `main.go`, wire `standingsService` + `standingsHandler`, add `GET /standings` and `OPTIONS /standings` routes

## Phase 3: Frontend — Hook + Component + Wiring

- [x] 3.1 In `features/hooks.ts`, add `useStandings()` hook with `useState<GroupStanding[]>` + initial `useEffect` fetch
- [x] 3.2 In same hook, add throttled `focus` listener on `window` (skip if <10s since last fetch); cleanup on unmount
- [x] 3.3 Create `frontend/src/components/GroupStandings.tsx` accepting `groupName` and `standings` props
- [x] 3.4 In same component, filter `standings` by `groupName`, render responsive table (10 columns, `overflow-x-auto`, `min-w-[600px]` per `RankingPage` pattern)
- [x] 3.5 In same component, handle empty/loading state
- [x] 3.6 In `HomePage.tsx`, import hook + component, call hook, render `<GroupStandings group={currentGroup} standings={standings} />` after the `<table>`

## Phase 4: Verification (manual)

- [x] 4.1 Start backend + frontend on `feature/group-standings` branch
- [x] 4.2 `curl GET /standings` → 200 + 12 groups
- [x] 4.3 Verify a group with finished matches awards correct points (3/1/0) (México 9 pts, Rep.Checa 6, Sudafrica 3, Corea 0)
- [x] 4.4 Verify scheduled matches are ignored
- [x] 4.5 Verify tiebreakers order (Pts > GD > GF > alphabetical)
- [x] 4.6 Open HomePage → standings visible below matches for current group
- [x] 4.7 Click "Siguiente" → standings update to new group
- [x] 4.8 Admin updates a match result in another tab, return to HomePage → standings reflect change
- [x] 4.9 Mobile: standings table scrollable horizontally

## Phase 5: Commit & Archive

- [x] 5.1 Commit: `feat(homepage): embed group standings table`
- [ ] 5.2 Stay on `feature/group-standings` branch (no push, local)
- [ ] 5.3 When ready: merge to master and push
- [ ] 5.4 After prod verification, archive via sdd-archive
