# T02 — AutomationListWidget confined to its cell, compact rows, labels

Kind: design + bug. Depends on: T01 (uses `WidgetFrame`).

## Problem (verified)
`AutomationListWidget` was written for a free-flowing page: root has no `h-full`, list has
`max-h-[480px]`, header is `px-5 py-4`, rows are `p-3.5` with a 120 px wide trigger button.
In a 2×2 cell (157 × 157 px) it renders 420 px tall and paints over the five rows beneath it;
names and times are unreadable and the buttons overflow the card horizontally.

Other defects:
- It renders its own delete button in admin mode; every other widget is deleted only through
  Edit Mode. Inconsistent and it sits under the drag overlay anyway.
- `widget.title` is used raw: an empty title yields an empty header (other widgets fall back).
- Per-entity `config.labels` (ADR 0005) are ignored; only `auto.name` is shown.
- The "show all when no ids" fallback is dead code (domain requires ≥ 1 entity); keep it only
  for legacy rows but do not rely on it.

## Scope
File: `frontend/src/components/AutomationListWidget.tsx` (may move to
`overview/widgets/AutomationListWidget.tsx`; update the import in `OverviewsView.tsx`).

1. Wrap in `WidgetFrame` (caption: title fallback → `Automatisations`, count as secondary text,
   StaleBadge not applicable).
2. Body: `flex-1 min-h-0 overflow-y-auto` list, rows `h-10` single line: status dot
   (emerald pulse when `current > 0`), label (`labels[id] ?? auto.name`, truncate), relative time
   `text-[10px]`, and an icon-only trigger button `w-9 h-9 rounded-lg` (Play / spinner / Check).
   Remove the 120 px text button and the `min-h-[44px]` row button.
3. Drop the `isAdmin`/`onDeleteWidget` props and the header trash icon.
4. Keep `handleTrigger`, the 2 s "triggered" feedback and `onTriggerSuccess`.

## Acceptance
- At 1024×768 the 2×2 list body height equals the cell height (157 px) and scrolls internally
  with 3 automations; nothing is painted outside the card.
- Labels from `config.labels` are shown when present.
- `tsc` and `oxlint` clean.
