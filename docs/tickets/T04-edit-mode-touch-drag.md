# T04 — Edit Mode: touch drag, pointercancel, min-size resize, handles

Kind: bug. Depends on: T00 (edits `overview/WidgetGrid.tsx` and `overview/grid.ts`).

## Problems (code review)
1. **Touch drag is cancelled by the browser.** The move overlay and resize handle have no
   `touch-action: none`; on a tablet the browser claims the gesture for panning, fires
   `pointercancel` and never `pointerup`. `handleUp` is only bound to `pointerup`, so `dragInfo`
   stays set and the grid is stuck in a drag until the next pointerup anywhere.
2. **A plain tap shows a refusal toast.** On `pointerup` with no movement `previewRectsRef` is
   null, so the code falls into the `else` branch and shows "Déplacement refusé…".
3. **Resize ignores the Display Mode minimum.** `handleMove` clamps spans to ≥ 1; the server
   then rejects with 4xx and the user sees "Le serveur a refusé cette disposition." ADR 0004
   says the minimum is applied at resize time.
4. **Handles are not touch-sized.** Resize handle is 16 × 16 px; edit/delete buttons are ~26 px
   and overlap the widget caption (verified in the edit-mode screenshot).
5. Toast timers are never cleared; two toasts in 3 s race each other.
6. Releasing the pointer outside the window leaves the drag active (no `blur` handling).

## Scope
- Add `touch-none select-none` to the move overlay and the resize handle.
- Listen to `pointercancel` and `window blur` as **cancel** (snap back, no toast).
- On `pointerup` with no preview: do nothing (no toast, no persist).
- Resize: clamp `colSpan/rowSpan` to `MIN_SIZE[widget.config.display]` before `reflow`; export
  a small `clampToMinSize` helper from `grid.ts`.
- Resize handle: visible 24 × 24 px rounded square, hit area 44 × 44 px (padding or `::before`).
- Edit actions: a single pill at the **bottom-left** of the cell (`h-9`, two `w-9` icon buttons)
  so the caption stays readable; keep them `z-20` above the move overlay. Selected/dragged widget
  gets a `ring-2 ring-indigo-400`; refused keeps `ring-red-500`.
- Toast: keep one `useRef` timer, clear on replace and on unmount.

## Acceptance
- Manual: in Playwright with `hasTouch: true`, a drag of 2 cells via `page.touchscreen` (or
  pointer events with `pointerType: 'touch'`) persists a new layout; a tap without movement
  produces no toast and no request.
- Shrinking an `arc` widget below 2×2 is impossible in the preview; no 4xx is emitted.
- `tsc` and `oxlint` clean.

## Decision (2026-09-05)
The per-cell pill was replaced after the first render check showed it covering the whole body of
1-row widgets (72 px cells). Final design: tap selects a widget (indigo ring, `z-20`), the resize
handle (24 px visual, 44 px hit area) overhangs the selected widget's bottom-right corner into the
grid gap, and Modifier / Supprimer / Désélectionner live in a floating `h-12` bar centred at the
bottom of the grid (top when the selection touches the last row). The drag preview and refusal
state are written to refs synchronously from the pointer handlers so a fast drop never persists a
stale preview.
