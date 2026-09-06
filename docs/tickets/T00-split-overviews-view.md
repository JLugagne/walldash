# T00 — Split OverviewsView into header, grid and view

Kind: refactor (no behavior change). Blocks: T01, T03, T04, T05.

## Why
`OverviewsView.tsx` is 818 lines and mixes data fetching, overview CRUD, the header bar, the
drag/resize state machine and the widget renderer. T03, T04 and T05 all need to edit it; splitting
first lets them run in parallel without conflicts.

## Scope
Create, under `frontend/src/components/overview/`:

- `OverviewHeader.tsx` — the whole sub-header (view mode menu slot, tabs, create/rename forms,
  admin actions, admin toggle). Props only: `overviews`, `activeOverviewId`, `isAdmin`,
  `isEditMode`, callbacks. No fetch calls inside.
- `WidgetGrid.tsx` — the grid container, edit-mode guides, per-widget cell wrapper, the drag/resize
  gesture (`DragInfo`, `startDrag`, the pointer effect, preview rects, refusal ring), the layout
  toast. Receives `overview`, `isEditMode`, `renderWidgetBody`, `onPersistLayout`, `onEditWidget`,
  `onDeleteWidget`.
- `useOverviewData.ts` (hook) — `overviews`, `automations`, `loading`, `fetchOverviews`,
  `fetchAutomations`, the 5 s automations poll.

`OverviewsView.tsx` keeps: state wiring, CRUD handlers calling `apiFetch`, `renderWidgetBody`, the
`AddWidgetModal`. Target under 300 lines.

Keep `widgetsToRects`, `isStaleDevice`, `parseNumericState` where they are used; move pure
helpers to `overview/grid.ts` or a new `overview/format.ts` if shared.

## Acceptance
- `npx tsc -b --force` and `npx oxlint src/` produce no new errors or warnings.
- Rendering and behavior are pixel-identical (same class names, same DOM order).
- No French in code or comments; no comments inside function bodies.
