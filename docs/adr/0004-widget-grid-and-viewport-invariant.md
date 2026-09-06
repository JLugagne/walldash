# 4. Integer-coordinate widget grid and no-overflow invariant

Date: 2026-09-04

## Status

Accepted

## Context

The dashboard is intended for a wall-mounted tablet: it must occupy exactly the window, without scrollbars, and no Widget must ever leave the viewport. The model from ADR 0003 only carried a per-Widget `Order int`, insufficient for free movement and resizing.

Two alternatives were ruled out. Normalized absolute positions (`x, y, w, h` as 0–1 floats) allow pixel-perfect placement, but require revalidating containment on every window resize and produce touch targets of arbitrary sizes. A CSS-driven responsive flow does not allow resizing a Widget.

## Decision

Anchor each Widget in a Widget Grid with integer coordinates (`col`, `row`, `col_span`, `row_span`). The grid dimensions are persisted on each Overview Dashboard, with a default of 12 × 8 and no edit screen in v1.

No-overflow becomes a model property, checked in the domain: `col + col_span <= cols` and `row + row_span <= rows`. Each Display Mode also declares a minimum size, applied at both creation and resize time.

Movement is done by pushing neighboring Widgets in the gesture direction, applied all-or-nothing: the full rearrangement is calculated before being applied, and if the result does not fit entirely within the grid, the drop is rejected and nothing moves.

Positions resulting from a push are written by a dedicated layout route, in a single transaction, separate from Widget content updates.

## Consequences

- The invariant "no Widget outside the viewport" cannot be lost by forgetting an interface validation: it is false or true in the domain.
- Since the grid is persisted per dashboard, changing the default does not silently rearrange already-composed dashboards.
- The all-or-nothing approach may refuse a gesture the user believed was feasible; this is the price of the invariant, and the refusal must therefore be signaled visually during hover and not only on drop.
- A gesture-directed push is not commutative: two trajectories arriving at the same cell can produce two different layouts. Accepted in favor of natural touch gesture feel.
- Cells are not square and their aspect ratio varies with tablet orientation; circular Display Modes must draw themselves in a centered square within their cell rather than filling it.