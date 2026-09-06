# T06 — DeleteOverview in one transaction, stop swallowing errors

Kind: bug (Go). Independent.

## Problem
`sqlite.overviewRepo.DeleteOverview` runs `DELETE FROM widgets …` with `_, _ =` then deletes the
dashboard, outside any transaction, and `App.DeleteOverview` calls the repository directly instead
of going through `uow.Do` like `CreateWidget` and `UpdateLayout`. A failure between the two
statements leaves a dashboard with no widgets or orphan widgets, and a failed widget delete is
never reported.

## Scope (TDD, per house rules: failing test first)
1. `internal/dashboard/app/overviews_test.go`: test that `DeleteOverview` performs
   `Widgets.DeleteWidgetsByDashboardID` **and** `Overviews.DeleteOverview` inside a single
   `uow.Do` call (use the existing `uowtest` fake; assert both repos were invoked from within the
   unit of work and that a widget-delete error aborts before the dashboard delete).
2. Make it pass: `App.DeleteOverview` uses `a.uow.Do`, deletes widgets via
   `repos.Widgets.DeleteWidgetsByDashboardID`, then `repos.Overviews.DeleteOverview`.
3. `sqlite.overviewRepo.DeleteOverview` no longer deletes widgets itself and no longer ignores
   errors; update the repository contract test in `overviewstest/contract.go` if it asserted the
   cascade at repository level (the cascade now belongs to the app layer).

House rules: `errors.Join`, no comments inside function bodies, doc comments on exported symbols
only when they add information, tests in `_test` package.

## Acceptance
- `go test ./...` green, new test fails before the fix and passes after.
- `go vet ./...` clean.
