# T03 — Fixed-height tablet header consistent with EditorHeader

Kind: design. Depends on: T00 (edits `overview/OverviewHeader.tsx`).

## Problem (verified)
The sub-header is `flex-wrap px-6 py-3`: 81 px tall in user mode and it grows when admin
controls appear (rename, delete, Mode Édition, Ajouter un Widget, Mode Admin Actif). Tabs are
32 px tall and icon buttons 30 px, below the 44 px touch target the rest of the app uses
(`LevelSelector` 40 px rows, `NavigationControls` 44 px). The 2D editor already has a fixed
`h-14` header (`components/editor/EditorHeader.tsx`); the 3D view uses floating 48 px FABs.
During initial load the whole view including the header is replaced by a spinner, so the view
mode menu is unreachable while `/api/overviews` is pending.

## Scope
File: `frontend/src/components/overview/OverviewHeader.tsx`. Mirror `EditorHeader`:

- `header.h-14 shrink-0 bg-slate-900/90 border-b border-slate-800 px-3 gap-3` and never wraps.
- Left: view mode menu slot, divider, tabs in a horizontally scrolling strip
  (`overflow-x-auto [scrollbar-width:none]`), each tab `h-10 px-4 rounded-xl`. Active tab
  indigo, others slate as today.
- Right, user mode: only the admin toggle (icon + short label, `h-10`).
- Right, admin mode: `Ajouter` (primary, `h-10`), `Disposition` toggle (`h-10`, cyan when active,
  same styling as today), and a `MoreHorizontal` overflow menu (`h-10 w-10`) containing
  `Nouvel Overview`, `Renommer`, `Supprimer` (red). Click-outside closes it (same pattern as
  `EditorHeader`/`ViewModeMenu`).
- Create and rename become a small popover/inline form anchored under the header
  (`absolute top-full`), not inline in the tab strip, so the header height never changes.
- Loading state: `OverviewsView` keeps rendering the header; the spinner is only in `main`.

## Acceptance
- Header height is exactly 56 px in user mode, admin mode and edit mode, with and without a
  rename in progress.
- All tappable controls are ≥ 40 px tall.
- View mode menu is visible while overviews are loading.
- `tsc` and `oxlint` clean.
