import type { DisplayMode } from '../../types'

/** An axis-aligned area of the Widget Grid, in whole cells, half-open on the far edge. */
export interface Rect {
  id: string
  col: number
  row: number
  colSpan: number
  rowSpan: number
}

/** The fixed shape of a Widget Grid. Never derived from the viewport: cells stretch instead. */
export interface GridSize {
  cols: number
  rows: number
}

/**
 * MIN_SIZE is the smallest legal span of a Widget per Display Mode.
 * It mirrors the Go matrix in internal/dashboard/domain/overview.go, which stays the single
 * source of truth: the server rejects anything smaller, this table only keeps the UI honest.
 */
export const MIN_SIZE: Record<DisplayMode, { cols: number; rows: number }> = {
  number: { cols: 1, rows: 1 },
  arc: { cols: 2, rows: 2 },
  bar: { cols: 2, rows: 1 },
  toggle: { cols: 1, rows: 1 },
  list: { cols: 2, rows: 2 },
  weather: { cols: 2, rows: 2 },
}

/**
 * clampToMinSize raises a span to the MIN_SIZE floor of a Display Mode so that a resize preview
 * never proposes an area the server would reject with a 4xx. It knows nothing about the grid
 * bounds: apply it last, and let reflow refuse a minimum that no longer fits.
 */
export function clampToMinSize(
  display: DisplayMode,
  colSpan: number,
  rowSpan: number,
): { colSpan: number; rowSpan: number } {
  const min = MIN_SIZE[display]
  return { colSpan: Math.max(colSpan, min.cols), rowSpan: Math.max(rowSpan, min.rows) }
}

/** rectsOverlap reports whether two areas share at least one cell. Touching edges do not overlap. */
export function rectsOverlap(a: Rect, b: Rect): boolean {
  return (
    a.col < b.col + b.colSpan &&
    b.col < a.col + a.colSpan &&
    a.row < b.row + b.rowSpan &&
    b.row < a.row + a.rowSpan
  )
}

/**
 * fitsInGrid enforces the viewport invariant col + colSpan <= cols and row + rowSpan <= rows,
 * plus non-negative origin and strictly positive spans. A grid never scrolls, so an area that
 * does not fit can never be shown.
 */
export function fitsInGrid(r: Rect, g: GridSize): boolean {
  return (
    r.colSpan > 0 &&
    r.rowSpan > 0 &&
    r.col >= 0 &&
    r.row >= 0 &&
    r.col + r.colSpan <= g.cols &&
    r.row + r.rowSpan <= g.rows
  )
}

/**
 * findFreeArea returns the origin of the first area of the requested size that fits in the grid
 * and touches none of the occupied areas, scanning row-major from the top-left. It returns null
 * when the grid has no room, which is what disables the Add button. Inputs are never mutated.
 */
export function findFreeArea(
  occupied: Rect[],
  g: GridSize,
  cols: number,
  rows: number,
): { col: number; row: number } | null {
  if (cols <= 0 || rows <= 0) return null
  for (let row = 0; row + rows <= g.rows; row++) {
    for (let col = 0; col + cols <= g.cols; col++) {
      const candidate: Rect = { id: '', col, row, colSpan: cols, rowSpan: rows }
      if (!occupied.some((r) => rectsOverlap(candidate, r))) return { col, row }
    }
  }
  return null
}

export type ReflowResult =
  | { ok: true; rects: Rect[] }
  | { ok: false; reason: 'out-of-bounds' | 'cycle' | 'overlap' }

type Axis = 'x' | 'y'

function dominantAxis(dx: number, dy: number): Axis {
  return Math.abs(dx) > Math.abs(dy) ? 'x' : 'y'
}

function pushSign(delta: number): 1 | -1 {
  return delta < 0 ? -1 : 1
}

function shiftClear(target: Rect, pusher: Rect, axis: Axis, sign: 1 | -1): void {
  if (axis === 'x') {
    target.col = sign > 0 ? pusher.col + pusher.colSpan : pusher.col - target.colSpan
    return
  }
  target.row = sign > 0 ? pusher.row + pusher.rowSpan : pusher.row - target.rowSpan
}

function firstCollision(work: Rect[], pusher: Rect, from: number): number {
  for (let i = from; i < work.length; i++) {
    const other = work[i]
    if (other.id !== pusher.id && rectsOverlap(pusher, other)) return i
  }
  return -1
}

/**
 * reflow moves one Widget to a target area and pushes the Widgets it now covers out of the way,
 * all at once. It is pure: the input rects are never mutated and, on success, a brand new array
 * of brand new rects is returned in the input order.
 *
 * Semantics, per ADR 0004: neighbours are pushed along the dominant axis of the vector from the
 * moved rect's current origin to the target origin, measured at drop time, and each push cascades
 * to whatever it in turn covers. A tie between the two axes, and a pure resize whose vector is
 * zero, both resolve to the vertical axis with a downward push.
 *
 * It is all-or-nothing. Failure is total and no partial layout is ever observable:
 *  - 'out-of-bounds' when the result, or an unknown movedId, cannot sit inside the grid,
 *  - 'cycle' when the cascade comes back to a rect it has already displaced, which also
 *    guarantees termination since every rect is displaced at most once,
 *  - 'overlap' when the settled result still has two rects sharing a cell.
 * The caller keeps its previous layout untouched on any failure.
 */
export function reflow(
  rects: Rect[],
  g: GridSize,
  movedId: string,
  target: { col: number; row: number; colSpan: number; rowSpan: number },
): ReflowResult {
  const originIndex = rects.findIndex((r) => r.id === movedId)
  if (originIndex < 0) return { ok: false, reason: 'out-of-bounds' }

  const origin = rects[originIndex]
  const dx = target.col - origin.col
  const dy = target.row - origin.row
  const axis = dominantAxis(dx, dy)
  const sign = pushSign(axis === 'x' ? dx : dy)

  const work: Rect[] = rects.map((r) => ({ ...r }))
  const moved = work[originIndex]
  moved.col = target.col
  moved.row = target.row
  moved.colSpan = target.colSpan
  moved.rowSpan = target.rowSpan

  const displaced = new Set<string>([movedId])
  const queue: Rect[] = [moved]

  while (queue.length > 0) {
    const pusher = queue.shift()
    if (!pusher) break
    let scanFrom = 0
    for (;;) {
      const hit = firstCollision(work, pusher, scanFrom)
      if (hit < 0) break
      const other = work[hit]
      if (displaced.has(other.id)) return { ok: false, reason: 'cycle' }
      displaced.add(other.id)
      shiftClear(other, pusher, axis, sign)
      queue.push(other)
      scanFrom = hit + 1
    }
  }

  for (const r of work) {
    if (!fitsInGrid(r, g)) return { ok: false, reason: 'out-of-bounds' }
  }
  for (let i = 0; i < work.length; i++) {
    for (let j = i + 1; j < work.length; j++) {
      if (rectsOverlap(work[i], work[j])) return { ok: false, reason: 'overlap' }
    }
  }
  return { ok: true, rects: work }
}
