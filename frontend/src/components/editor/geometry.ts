import type { DevicePlacement, Point2D, WallOpening, WallSegment, Zone } from '../../types'

/** Plan units: 1 unit = 2.5 cm (see planImport.ts and the 3D SCALE constant). */
export const UNITS_PER_METER = 40
export const UNITS_PER_CM = UNITS_PER_METER / 100
export const MIN_OPENING_WIDTH = 12
export const MIN_WALL_LENGTH = 8

let idCounter = 0
export function generateId(prefix: string): string {
  idCounter += 1
  return `${prefix}-${Date.now()}-${idCounter}`
}

export function formatMeters(units: number): string {
  return `${(units / UNITS_PER_METER).toFixed(2)} m`
}

export function unitsToCm(units: number): number {
  return Math.round(units / UNITS_PER_CM)
}

export function cmToUnits(cm: number): number {
  return Math.round(cm * UNITS_PER_CM)
}

export function wallLength(wall: WallSegment): number {
  return Math.hypot(wall.x2 - wall.x1, wall.y2 - wall.y1)
}

export function wallAngleDeg(wall: WallSegment): number {
  return (Math.atan2(wall.y2 - wall.y1, wall.x2 - wall.x1) * 180) / Math.PI
}

export function snapToGrid(pt: Point2D, size: number): Point2D {
  return { x: Math.round(pt.x / size) * size, y: Math.round(pt.y / size) * size }
}

/** Constrains `to` onto the horizontal or vertical axis through `from`, whichever is closer. */
export function axisLock(from: Point2D, to: Point2D): Point2D {
  const dx = Math.abs(to.x - from.x)
  const dy = Math.abs(to.y - from.y)
  return dx >= dy ? { x: to.x, y: from.y } : { x: from.x, y: to.y }
}

export function samePoint(a: Point2D, b: Point2D, tolerance = 0.5): boolean {
  return Math.hypot(a.x - b.x, a.y - b.y) <= tolerance
}

export type WallEnd = 'start' | 'end'

export interface VertexRef {
  wallId: string
  end: WallEnd
}

export function wallEndpoint(wall: WallSegment, end: WallEnd): Point2D {
  return end === 'start' ? { x: wall.x1, y: wall.y1 } : { x: wall.x2, y: wall.y2 }
}

export function sameRef(a: VertexRef, b: VertexRef): boolean {
  return a.wallId === b.wallId && a.end === b.end
}

/** Every wall endpoint lying within `tolerance` of `pt`. */
export function verticesAt(walls: WallSegment[], pt: Point2D, tolerance: number): VertexRef[] {
  const refs: VertexRef[] = []
  for (const w of walls) {
    if (Math.hypot(w.x1 - pt.x, w.y1 - pt.y) <= tolerance) refs.push({ wallId: w.id, end: 'start' })
    if (Math.hypot(w.x2 - pt.x, w.y2 - pt.y) <= tolerance) refs.push({ wallId: w.id, end: 'end' })
  }
  return refs
}

/** Nearest wall endpoint within `tolerance`, ignoring refs for which `skip` returns true. */
export function nearestVertex(
  walls: WallSegment[],
  pt: Point2D,
  tolerance: number,
  skip?: (ref: VertexRef) => boolean
): Point2D | null {
  let best: Point2D | null = null
  let bestDist = tolerance
  for (const w of walls) {
    for (const end of ['start', 'end'] as WallEnd[]) {
      if (skip?.({ wallId: w.id, end })) continue
      const v = wallEndpoint(w, end)
      const d = Math.hypot(v.x - pt.x, v.y - pt.y)
      if (d <= bestDist) {
        bestDist = d
        best = v
      }
    }
  }
  return best
}

/** Vertices shared by at least two wall endpoints, used to draw joint markers. */
export function sharedJoints(walls: WallSegment[], tolerance: number): Point2D[] {
  const joints: Point2D[] = []
  const seen: Point2D[] = []
  for (const w of walls) {
    for (const end of ['start', 'end'] as WallEnd[]) {
      const v = wallEndpoint(w, end)
      if (seen.some((s) => samePoint(s, v, tolerance))) {
        if (!joints.some((j) => samePoint(j, v, tolerance))) joints.push(v)
      } else {
        seen.push(v)
      }
    }
  }
  return joints
}

export function clampOpening(op: WallOpening, length: number): WallOpening {
  const width = Math.min(Math.max(op.width, MIN_OPENING_WIDTH), Math.max(MIN_OPENING_WIDTH, length))
  const half = width / 2
  const offset = Math.min(Math.max(op.offset, half), Math.max(half, length - half))
  return { ...op, width: Math.round(width), offset: Math.round(offset) }
}

export function withClampedOpenings(wall: WallSegment): WallSegment {
  if (!wall.openings || wall.openings.length === 0) return wall
  const len = wallLength(wall)
  return { ...wall, openings: wall.openings.map((op) => clampOpening(op, len)) }
}

function groupRefs(refs: VertexRef[]): Map<string, VertexRef[]> {
  const byWall = new Map<string, VertexRef[]>()
  for (const r of refs) byWall.set(r.wallId, [...(byWall.get(r.wallId) ?? []), r])
  return byWall
}

/** Moves every referenced endpoint onto `to`. */
export function setVertices(walls: WallSegment[], refs: VertexRef[], to: Point2D): WallSegment[] {
  const byWall = groupRefs(refs)
  return walls.map((w) => {
    const rs = byWall.get(w.id)
    if (!rs) return w
    const next = { ...w }
    for (const r of rs) {
      if (r.end === 'start') {
        next.x1 = to.x
        next.y1 = to.y
      } else {
        next.x2 = to.x
        next.y2 = to.y
      }
    }
    return withClampedOpenings(next)
  })
}

/** Translates every referenced endpoint by (dx, dy). */
export function translateVertices(walls: WallSegment[], refs: VertexRef[], dx: number, dy: number): WallSegment[] {
  const byWall = groupRefs(refs)
  return walls.map((w) => {
    const rs = byWall.get(w.id)
    if (!rs) return w
    const next = { ...w }
    for (const r of rs) {
      if (r.end === 'start') {
        next.x1 += dx
        next.y1 += dy
      } else {
        next.x2 += dx
        next.y2 += dy
      }
    }
    return withClampedOpenings(next)
  })
}

/** Endpoints of `wall` plus every other wall endpoint touching them, so a wall drags its corners along. */
export function wallWithJoints(walls: WallSegment[], wall: WallSegment, tolerance: number): VertexRef[] {
  const refs: VertexRef[] = [
    { wallId: wall.id, end: 'start' },
    { wallId: wall.id, end: 'end' },
  ]
  const others = walls.filter((w) => w.id !== wall.id)
  for (const end of ['start', 'end'] as WallEnd[]) {
    for (const r of verticesAt(others, wallEndpoint(wall, end), tolerance)) {
      if (!refs.some((x) => sameRef(x, r))) refs.push(r)
    }
  }
  return refs
}

export function vertexWithJoints(walls: WallSegment[], ref: VertexRef, tolerance: number): VertexRef[] {
  const wall = walls.find((w) => w.id === ref.wallId)
  if (!wall) return [ref]
  const pt = wallEndpoint(wall, ref.end)
  const refs = [ref]
  for (const r of verticesAt(walls, pt, tolerance)) {
    if (!refs.some((x) => sameRef(x, r))) refs.push(r)
  }
  return refs
}

export interface WallProjection {
  wall: WallSegment
  offset: number
  angle: number
  point: Point2D
  distance: number
}

export function projectOnNearestWall(walls: WallSegment[], point: Point2D, maxDist: number): WallProjection | null {
  let best: WallProjection | null = null
  for (const w of walls) {
    const dx = w.x2 - w.x1
    const dy = w.y2 - w.y1
    const len = Math.hypot(dx, dy)
    if (len < 1) continue
    const t = ((point.x - w.x1) * dx + (point.y - w.y1) * dy) / (len * len)
    if (t < -0.05 || t > 1.05) continue
    const clampedT = Math.max(0, Math.min(1, t))
    const proj = { x: w.x1 + clampedT * dx, y: w.y1 + clampedT * dy }
    const distance = Math.hypot(point.x - proj.x, point.y - proj.y)
    if (distance <= maxDist && (!best || distance < best.distance)) {
      best = { wall: w, offset: clampedT * len, angle: Math.atan2(dy, dx), point: proj, distance }
    }
  }
  return best
}

/** Distance along the wall from its start point to the projection of `point`. */
export function offsetAlongWall(wall: WallSegment, point: Point2D): number {
  const dx = wall.x2 - wall.x1
  const dy = wall.y2 - wall.y1
  const len = Math.hypot(dx, dy)
  if (len < 1) return 0
  return (((point.x - wall.x1) * dx + (point.y - wall.y1) * dy) / (len * len)) * len
}

export function setWallLength(wall: WallSegment, length: number): WallSegment {
  const len = wallLength(wall)
  if (len < 1) return wall
  const scale = Math.max(MIN_WALL_LENGTH, length) / len
  return withClampedOpenings({
    ...wall,
    x2: Math.round(wall.x1 + (wall.x2 - wall.x1) * scale),
    y2: Math.round(wall.y1 + (wall.y2 - wall.y1) * scale),
  })
}

/** Splits a wall at its midpoint; openings stay on the half they belong to. */
export function splitWall(wall: WallSegment, secondId: string): [WallSegment, WallSegment] {
  const mid = { x: Math.round((wall.x1 + wall.x2) / 2), y: Math.round((wall.y1 + wall.y2) / 2) }
  const half = wallLength(wall) / 2
  const first: WallOpening[] = []
  const second: WallOpening[] = []
  for (const op of wall.openings ?? []) {
    if (op.offset <= half) first.push(op)
    else second.push({ ...op, offset: op.offset - half })
  }
  return [
    withClampedOpenings({ ...wall, x2: mid.x, y2: mid.y, openings: first }),
    withClampedOpenings({ ...wall, id: secondId, x1: mid.x, y1: mid.y, openings: second }),
  ]
}

export function translatePoints(points: Point2D[], dx: number, dy: number): Point2D[] {
  return points.map((p) => ({ x: p.x + dx, y: p.y + dy }))
}

export interface Bounds {
  minX: number
  minY: number
  maxX: number
  maxY: number
}

export function planBounds(walls: WallSegment[], zones: Zone[], placements: DevicePlacement[]): Bounds | null {
  let minX = Infinity
  let minY = Infinity
  let maxX = -Infinity
  let maxY = -Infinity
  const add = (x: number, y: number) => {
    minX = Math.min(minX, x)
    maxX = Math.max(maxX, x)
    minY = Math.min(minY, y)
    maxY = Math.max(maxY, y)
  }
  for (const w of walls) {
    add(w.x1, w.y1)
    add(w.x2, w.y2)
  }
  for (const z of zones) for (const p of z.points) add(p.x, p.y)
  for (const p of placements) add(p.x, p.y)
  if (!isFinite(minX)) return null
  return { minX, minY, maxX, maxY }
}
