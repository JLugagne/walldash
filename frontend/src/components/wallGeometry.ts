import * as THREE from 'three'
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js'
import polygonClipping, { type MultiPolygon, type Polygon, type Ring } from 'polygon-clipping'
import type { WallOpening, WallSegment } from '../types'

export const WALL_HEIGHT = 2.4
export const DOOR_HEIGHT = 2.1
export const WINDOW_SILL_HEIGHT = 0.9
export const WINDOW_TOP_HEIGHT = 2.1

const SVG_CENTER_X = 500
const SVG_CENTER_Y = 350
export const SCALE = 0.04
const MIN_THICKNESS = 0.25
const JOINT_SNAP = 12 * SCALE

export function toWorldX(x: number): number {
  return (x - SVG_CENTER_X) * SCALE
}

export function toWorldZ(y: number): number {
  return (y - SVG_CENTER_Y) * SCALE
}

export function wallThickness3D(wall: WallSegment): number {
  return Math.max((wall.thickness || 12) * SCALE, MIN_THICKNESS)
}

export interface OpeningPlacement {
  key: string
  type: WallOpening['type']
  x: number
  z: number
  rotY: number
  width: number
  thickness: number
  flipSide?: boolean
  flipHinge?: boolean
  hideDoor?: boolean
}

export interface WallGeometryResult {
  sides: THREE.BufferGeometry | null
  caps: THREE.BufferGeometry | null
  openings: OpeningPlacement[]
}

interface Band {
  y0: number
  y1: number
  excludes: (type: WallOpening['type']) => boolean
}

const BANDS: Band[] = [
  { y0: 0, y1: WINDOW_SILL_HEIGHT, excludes: (t) => t === 'door' },
  { y0: WINDOW_SILL_HEIGHT, y1: DOOR_HEIGHT, excludes: () => true },
  { y0: DOOR_HEIGHT, y1: WALL_HEIGHT, excludes: () => false },
]

interface WallFrame {
  wall: WallSegment
  x1: number
  z1: number
  ux: number
  uz: number
  nx: number
  nz: number
  length: number
  halfThickness: number
  extStart: number
  extEnd: number
  openings: { start: number; end: number; opening: WallOpening }[]
}

function buildFrames(walls: WallSegment[]): WallFrame[] {
  const raw = walls
    .map((wall) => {
      const x1 = toWorldX(wall.x1)
      const z1 = toWorldZ(wall.y1)
      const x2 = toWorldX(wall.x2)
      const z2 = toWorldZ(wall.y2)
      const length = Math.hypot(x2 - x1, z2 - z1)
      return { wall, x1, z1, x2, z2, length, halfThickness: wallThickness3D(wall) / 2 }
    })
    .filter((f) => f.length > 0.001)

  const jointExtension = (x: number, z: number, self: number): number => {
    let ext = 0
    for (let i = 0; i < raw.length; i++) {
      const o = raw[i]
      if (i === self) continue
      if (Math.hypot(o.x1 - x, o.z1 - z) <= JOINT_SNAP || Math.hypot(o.x2 - x, o.z2 - z) <= JOINT_SNAP) {
        ext = Math.max(ext, o.halfThickness)
      }
    }
    return ext
  }

  return raw.map((f, idx) => {
    const ux = (f.x2 - f.x1) / f.length
    const uz = (f.z2 - f.z1) / f.length
    const openings: WallFrame['openings'] = []
    for (const op of f.wall.openings || []) {
      const center = op.offset * SCALE
      const w = op.width * SCALE
      const start = Math.max(0, center - w / 2)
      const end = Math.min(f.length, center + w / 2)
      if (end - start > 0.05) openings.push({ start, end, opening: op })
    }
    openings.sort((a, b) => a.start - b.start)
    return {
      wall: f.wall,
      x1: f.x1,
      z1: f.z1,
      ux,
      uz,
      nx: -uz,
      nz: ux,
      length: f.length,
      halfThickness: f.halfThickness,
      extStart: jointExtension(f.x1, f.z1, idx),
      extEnd: jointExtension(f.x2, f.z2, idx),
      openings,
    }
  })
}

function solidIntervals(frame: WallFrame, band: Band): { a: number; b: number }[] {
  const result: { a: number; b: number }[] = []
  let cursor = -frame.extStart
  for (const item of frame.openings) {
    if (!band.excludes(item.opening.type)) continue
    if (item.start - cursor > 0.01) result.push({ a: cursor, b: item.start })
    cursor = Math.max(cursor, item.end)
  }
  const end = frame.length + frame.extEnd
  if (end - cursor > 0.01) result.push({ a: cursor, b: end })
  return result
}

function intervalRect(frame: WallFrame, a: number, b: number): Polygon {
  const { x1, z1, ux, uz, nx, nz, halfThickness: h } = frame
  const ring: Ring = [
    [x1 + ux * a + nx * h, z1 + uz * a + nz * h],
    [x1 + ux * b + nx * h, z1 + uz * b + nz * h],
    [x1 + ux * b - nx * h, z1 + uz * b - nz * h],
    [x1 + ux * a - nx * h, z1 + uz * a - nz * h],
  ]
  return [ring]
}

function bandFootprint(frames: WallFrame[], band: Band): MultiPolygon {
  const rects: Polygon[] = []
  for (const frame of frames) {
    for (const { a, b } of solidIntervals(frame, band)) rects.push(intervalRect(frame, a, b))
  }
  if (rects.length === 0) return []
  try {
    return polygonClipping.union(rects[0], ...rects.slice(1))
  } catch {
    return rects
  }
}

function ringArea(ring: Ring): number {
  let area = 0
  for (let i = 0; i < ring.length; i++) {
    const p = ring[i]
    const q = ring[(i + 1) % ring.length]
    area += p[0] * q[1] - q[0] * p[1]
  }
  return area / 2
}

function normalizeRing(ring: Ring): Ring {
  if (ring.length < 2) return ring
  const first = ring[0]
  const last = ring[ring.length - 1]
  if (first[0] === last[0] && first[1] === last[1]) return ring.slice(0, -1)
  return ring
}

function appendSideQuads(
  target: { positions: number[]; normals: number[]; uvs: number[] },
  ring: Ring,
  isOuter: boolean,
  y0: number,
  y1: number
) {
  const pts = normalizeRing(ring)
  if (pts.length < 3) return
  const ccw = ringArea(pts) > 0
  const outwardSign = isOuter === ccw ? 1 : -1
  for (let i = 0; i < pts.length; i++) {
    const p = pts[i]
    const q = pts[(i + 1) % pts.length]
    const dx = q[0] - p[0]
    const dz = q[1] - p[1]
    const len = Math.hypot(dx, dz)
    if (len < 1e-6) continue
    const nx = (outwardSign * dz) / len
    const nz = (-outwardSign * dx) / len
    const a = [p[0], y0, p[1]]
    const b = [q[0], y0, q[1]]
    const c = [q[0], y1, q[1]]
    const d = [p[0], y1, p[1]]
    const ex = b[0] - a[0]
    const ez = b[2] - a[2]
    const crossX = -ez * (y1 - y0)
    const crossZ = ex * (y1 - y0)
    const frontFacing = crossX * nx + crossZ * nz > 0
    const tri1 = frontFacing ? [a, b, c] : [a, c, b]
    const tri2 = frontFacing ? [a, c, d] : [a, d, c]
    const uv = (v: number[]) => [v === a || v === d ? 0 : len, v[1] - y0]
    for (const v of [...tri1, ...tri2]) {
      target.positions.push(v[0], v[1], v[2])
      target.normals.push(nx, 0, nz)
      target.uvs.push(...uv(v))
    }
  }
}

function multiPolygonToShapes(mp: MultiPolygon): THREE.Shape[] {
  const shapes: THREE.Shape[] = []
  for (const poly of mp) {
    if (poly.length === 0) continue
    const outer = normalizeRing(poly[0])
    if (outer.length < 3) continue
    const shape = new THREE.Shape()
    shape.moveTo(outer[0][0], -outer[0][1])
    for (let i = 1; i < outer.length; i++) shape.lineTo(outer[i][0], -outer[i][1])
    shape.closePath()
    for (let h = 1; h < poly.length; h++) {
      const hole = normalizeRing(poly[h])
      if (hole.length < 3) continue
      const path = new THREE.Path()
      path.moveTo(hole[0][0], -hole[0][1])
      for (let i = 1; i < hole.length; i++) path.lineTo(hole[i][0], -hole[i][1])
      path.closePath()
      shape.holes.push(path)
    }
    shapes.push(shape)
  }
  return shapes
}

function capGeometry(mp: MultiPolygon, y: number): THREE.BufferGeometry | null {
  const shapes = multiPolygonToShapes(mp)
  if (shapes.length === 0) return null
  const geometry = new THREE.ShapeGeometry(shapes)
  geometry.rotateX(-Math.PI / 2)
  geometry.translate(0, y, 0)
  return geometry
}

function difference(a: MultiPolygon, b: MultiPolygon): MultiPolygon {
  if (a.length === 0) return []
  if (b.length === 0) return a
  try {
    return polygonClipping.difference(a, b)
  } catch {
    return a
  }
}

export function buildWallGeometry(walls: WallSegment[]): WallGeometryResult {
  const frames = buildFrames(walls)
  const openings: OpeningPlacement[] = []
  for (const frame of frames) {
    for (const item of frame.openings) {
      const mid = (item.start + item.end) / 2
      openings.push({
        key: `${frame.wall.id}-${item.opening.id}`,
        type: item.opening.type,
        x: frame.x1 + frame.ux * mid,
        z: frame.z1 + frame.uz * mid,
        rotY: -Math.atan2(frame.uz, frame.ux),
        width: item.end - item.start,
        thickness: frame.halfThickness * 2,
        flipSide: item.opening.flip_side,
        flipHinge: item.opening.flip_hinge,
        hideDoor: item.opening.hide_door,
      })
    }
  }

  if (frames.length === 0) return { sides: null, caps: null, openings }

  const footprints = BANDS.map((band) => bandFootprint(frames, band))
  const sideBuffers = { positions: [] as number[], normals: [] as number[], uvs: [] as number[] }
  const capParts: THREE.BufferGeometry[] = []

  for (let i = 0; i < BANDS.length; i++) {
    const band = BANDS[i]
    const footprint = footprints[i]
    for (const poly of footprint) {
      poly.forEach((ring, ringIdx) => appendSideQuads(sideBuffers, ring, ringIdx === 0, band.y0, band.y1))
    }
    const above = i + 1 < footprints.length ? footprints[i + 1] : []
    const cap = capGeometry(difference(footprint, above), band.y1)
    if (cap) capParts.push(cap)
  }

  let sides: THREE.BufferGeometry | null = null
  if (sideBuffers.positions.length > 0) {
    sides = new THREE.BufferGeometry()
    sides.setAttribute('position', new THREE.Float32BufferAttribute(sideBuffers.positions, 3))
    sides.setAttribute('normal', new THREE.Float32BufferAttribute(sideBuffers.normals, 3))
    sides.setAttribute('uv', new THREE.Float32BufferAttribute(sideBuffers.uvs, 2))
  }

  let caps: THREE.BufferGeometry | null = null
  if (capParts.length === 1) caps = capParts[0]
  else if (capParts.length > 1) {
    caps = mergeGeometries(capParts, false)
    capParts.forEach((g) => g.dispose())
  }

  return { sides, caps, openings }
}

export interface WallSnap {
  x: number
  z: number
  normalX: number
  normalZ: number
  rotY: number
  distance: number
}

// Projects a floor point onto the nearest wall face; the normal points from the wall toward the original point
export function snapToNearestWall(walls: WallSegment[], x: number, z: number): WallSnap | null {
  let best: WallSnap | null = null
  for (const wall of walls) {
    const x1 = toWorldX(wall.x1)
    const z1 = toWorldZ(wall.y1)
    const x2 = toWorldX(wall.x2)
    const z2 = toWorldZ(wall.y2)
    const dx = x2 - x1
    const dz = z2 - z1
    const lenSq = dx * dx + dz * dz
    if (lenSq < 1e-6) continue
    const t = Math.max(0, Math.min(1, ((x - x1) * dx + (z - z1) * dz) / lenSq))
    const px = x1 + t * dx
    const pz = z1 + t * dz
    const distance = Math.hypot(x - px, z - pz)
    if (best && distance >= best.distance) continue
    const len = Math.sqrt(lenSq)
    let nx = -dz / len
    let nz = dx / len
    if (nx * (x - px) + nz * (z - pz) < 0) {
      nx = -nx
      nz = -nz
    }
    const half = wallThickness3D(wall) / 2
    best = {
      x: px + nx * half,
      z: pz + nz * half,
      normalX: nx,
      normalZ: nz,
      rotY: -Math.atan2(dz, dx),
      distance,
    }
  }
  return best
}
