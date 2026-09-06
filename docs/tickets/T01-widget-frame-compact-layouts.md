# T01 — Shared WidgetFrame and compact layouts that fit MIN_SIZE cells

Kind: design + bug. Depends on: T00. Blocks: T02.

## Problem (verified at 1024×768, 12×8 grid)
A 1-row cell is 72 px tall. Every widget spends 36 px on a bordered header row
(`px-3 py-2 border-b`), so:

- Number 1×1: value `text-3xl` is clipped, unit disappears, label truncates to 3 letters
  (`Te…`). Raw state `21.3456` is shown unformatted. A non-numeric sensor state (`weather.*`,
  `cloudy`) renders as `—` forever because only numbers are accepted.
- Bar 2×1: the bar and the min/max labels are below the fold (scrollHeight 111 vs 70).
- Toggle 1×1: half the power button is clipped (103 vs 70).
- Arc 2×2 fits; keep its geometry (ADR 0004: circular modes draw in a centred square).
- Five components duplicate the same chrome (rounded-2xl card, header, StaleBadge slot).

## Scope
Files: `frontend/src/components/overview/widgets/*`, new `overview/widgets/WidgetFrame.tsx`,
new `overview/format.ts`. Do not touch `OverviewsView.tsx` beyond passing new props.

1. `WidgetFrame` — single card chrome used by every widget: `w-full h-full` rounded card,
   `overflow-hidden`, a **caption** row (label `text-[11px] font-semibold text-slate-400
   truncate`, optional icon, StaleBadge) that costs at most 18 px and has no border, then a
   body that is `flex-1 min-h-0`. Accept `dense` (1-row widgets) to shrink paddings to `px-2 py-1`.
2. `format.ts` — `formatSensorValue(state: string | number | null, unit?: string)` returning
   `{ text, unit }`: numbers via `Intl.NumberFormat('fr-FR', { maximumFractionDigits: 1 })`,
   non-numeric strings passed through (first letter upper-cased), null → `—`. Unit-test-free is
   acceptable (no frontend test runner) but keep it pure.
3. `NumberWidget` — caption + one centred value line with `clamp()` font size
   (`text-[clamp(1.1rem,3.2cqh,2.2rem)]` via container queries, or a `ResizeObserver`-free
   CSS-only approach). Show non-numeric states (value prop becomes `string | number | null`).
4. `BarWidget` — caption row holds label on the left and value+unit on the right; body is the
   bar only (height 8 px) with min/max as `text-[9px]` inside the bar track ends or omitted when
   `dense`. Must fit 72 px total.
5. `ToggleWidget` — the button fills the body (`inset-0`), power icon `w-6 h-6` in 1×1, label in
   the caption. Keep `pending` spinner and `aria-pressed`. Touch target = whole cell.
6. `ArcWidget` — adopt `WidgetFrame`; keep SVG square-centred; use `formatSensorValue`.
7. Grid container (`WidgetGrid.tsx`): `gridTemplateColumns/Rows` must use `minmax(0, 1fr)` so a
   tall child can never stretch a track.

Tablet defaults to consider: rounded-xl on 1-row widgets (2xl looks heavy at 72 px), consistent
accent colour per Display Mode (emerald number, cyan bar/arc, indigo toggle) applied to icon and
fill only.

## Acceptance
- At 1024×768 with the seeded dashboard (see `docs/tickets/README.md`), every widget at its
  MIN_SIZE has `body.scrollHeight <= body.clientHeight` (no clipping) and the value, unit and
  label are all visible on Number 1×1, Bar 2×1, Toggle 1×1.
- `weather.maison` state `cloudy` shows `Cloudy`, `21.3456 °C` shows `21,3 °C`.
- No duplicated chrome: each widget file is only its body.
- `tsc` and `oxlint` clean for the touched files.
