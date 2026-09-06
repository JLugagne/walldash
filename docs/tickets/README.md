# Tickets — Overview Dashboard design & bug fixes

Generated 2026-09-04 from a code review of `frontend/src/components/OverviewsView.tsx`,
`overview/widgets/*`, `AutomationListWidget.tsx`, `AddWidgetModal.tsx` and the Go overview
stack, plus a headless render at 1024×768 (tablet landscape) of a 12×8 dashboard holding one
widget per Display Mode.

Measured facts driving the design tickets (1024×768, 12×8 grid, current header 81 px):

| Item | Measured |
| --- | --- |
| Cell size | 72 × 72 px |
| Number 1×1 body scrollHeight / clientHeight | 80 / 70 (clipped) |
| Bar 2×1 body scrollHeight / clientHeight | 111 / 70 (clipped) |
| Toggle 1×1 body scrollHeight / clientHeight | 103 / 70 (clipped) |
| Automation list 2×2 body height vs cell | 420 / 157 (overflows onto 5 rows) |
| Arc 2×2 | fits |

Dependency order:

```
T00 split OverviewsView
 ├─ T01 widget frame & compact layouts ── T02 automation list widget
 ├─ T03 header redesign
 ├─ T04 edit-mode touch & drag fixes
 └─ T05 error surfacing (modal + overview CRUD)
T06 backend transactional delete (independent)
```

| ID | Title | Kind | Status |
| --- | --- | --- | --- |
| [T00](T00-split-overviews-view.md) | Split OverviewsView into header, grid and view | refactor | done |
| [T01](T01-widget-frame-compact-layouts.md) | Shared WidgetFrame and compact layouts that fit MIN_SIZE cells | design/bug | done |
| [T02](T02-automation-list-widget.md) | AutomationListWidget confined to its cell, compact rows, labels | design/bug | done |
| [T03](T03-header-redesign.md) | Fixed-height tablet header consistent with EditorHeader | design | done |
| [T04](T04-edit-mode-touch-drag.md) | Edit Mode: touch drag, pointercancel, min-size resize, handles | bug | done |
| [T05](T05-error-surfacing.md) | Surface server errors in the widget modal and overview CRUD | bug | done |
| [T06](T06-backend-transactional-delete.md) | DeleteOverview in one transaction, stop swallowing errors | bug (Go) | done |

## Outcome (2026-09-05)

All seven tickets landed; verified at 1024×768 with the seeded dashboard (header 56 px, every
widget body `scrollHeight <= clientHeight`, automation list confined to its 2×2 cell, `cloudy`
rendered as `Cloudy`). Deviations recorded:

- T00: `OverviewsView.tsx` ends at ~430 lines, above the 300 target. The remainder is the CRUD
  handlers, `renderWidgetBody` and the three onboarding states, which the ticket kept in place.
- T04: instead of a per-cell edit/delete pill, Edit Mode uses tap-to-select (indigo ring), shows
  the 44 px resize handle only on the selected widget, and hosts Modifier / Supprimer in one
  floating bar at the bottom of the grid. Content of 1-row widgets stays readable while laying out.
- Toast: a single banner is owned by `OverviewsView`; `WidgetGrid` reports through `onMessage` and
  `onPersistLayout` resolves to a message (or `null`) instead of a boolean.
