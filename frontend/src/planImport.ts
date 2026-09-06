import type { Plan, WallSegment, WallOpening, Zone, Point2D } from './types'

// The 2D editor stores coordinates in plan units where 1 unit = 2.5cm
// (see the wall length readout in PlanEditor2D.tsx: meters = units * 0.025).
const UNITS_PER_METER = 40
const UNITS_PER_CM = UNITS_PER_METER / 100

const ZONE_COLOR_PALETTE = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#64748b', '#059669']

let importIdCounter = 0
function nextId(prefix: string): string {
  importIdCounter += 1
  return `${prefix}-import-${Date.now()}-${importIdCounter}`
}

function isFiniteNumber(v: unknown): v is number {
  return typeof v === 'number' && Number.isFinite(v)
}

function metersToUnits(m: number): number {
  return Math.round(m * UNITS_PER_METER)
}

function cmToUnits(cm: number): number {
  return Math.round(cm * UNITS_PER_CM)
}

export class PlanImportError extends Error {
  issues: string[]

  constructor(issues: string[]) {
    super(issues.join('\n'))
    this.name = 'PlanImportError'
    this.issues = issues
  }
}

interface RawPoint {
  x?: unknown
  y?: unknown
}

interface RawOpening {
  type?: unknown
  offset_m?: unknown
  width_cm?: unknown
  flip_side?: unknown
  flip_hinge?: unknown
  hide_door?: unknown
}

interface RawWall {
  start?: RawPoint
  end?: RawPoint
  thickness_cm?: unknown
  openings?: unknown
}

interface RawZone {
  name?: unknown
  color?: unknown
  points?: unknown
}

interface RawFloorPlan {
  walls?: unknown
  zones?: unknown
}

function convertOpening(raw: RawOpening, wallLenUnits: number, path: string, issues: string[]): WallOpening | null {
  if (raw.type !== 'door' && raw.type !== 'window') {
    issues.push(`${path}: "type" must be "door" or "window"`)
    return null
  }
  if (!isFiniteNumber(raw.offset_m)) {
    issues.push(`${path}: "offset_m" must be a number`)
    return null
  }

  const defaultWidthCm = raw.type === 'door' ? 90 : 120
  const widthCm = isFiniteNumber(raw.width_cm) && raw.width_cm > 0 ? raw.width_cm : defaultWidthCm
  const width = cmToUnits(widthCm)

  let offset = metersToUnits(raw.offset_m)
  if (wallLenUnits > 0) {
    const minOffset = width / 2
    const maxOffset = Math.max(minOffset, wallLenUnits - width / 2)
    offset = Math.max(minOffset, Math.min(maxOffset, offset))
  }

  return {
    id: nextId(raw.type),
    type: raw.type,
    offset: Math.round(offset),
    width: Math.round(width),
    flip_side: Boolean(raw.flip_side),
    flip_hinge: Boolean(raw.flip_hinge),
    hide_door: raw.type === 'door' && Boolean(raw.hide_door),
  }
}

function convertWall(raw: RawWall, index: number, issues: string[]): WallSegment | null {
  const path = `walls[${index}]`
  const sx = raw.start?.x
  const sy = raw.start?.y
  const ex = raw.end?.x
  const ey = raw.end?.y

  if (!isFiniteNumber(sx) || !isFiniteNumber(sy) || !isFiniteNumber(ex) || !isFiniteNumber(ey)) {
    issues.push(`${path}: "start" and "end" must both have numeric x/y coordinates (in meters)`)
    return null
  }
  if (sx === ex && sy === ey) {
    issues.push(`${path}: "start" and "end" must be different points`)
    return null
  }

  const thicknessCm = isFiniteNumber(raw.thickness_cm) && raw.thickness_cm > 0 ? raw.thickness_cm : 20
  const x1 = metersToUnits(sx)
  const y1 = metersToUnits(sy)
  const x2 = metersToUnits(ex)
  const y2 = metersToUnits(ey)
  const wallLenUnits = Math.hypot(x2 - x1, y2 - y1)

  const openings: WallOpening[] = []
  if (raw.openings !== undefined) {
    if (!Array.isArray(raw.openings)) {
      issues.push(`${path}.openings: must be an array when present`)
    } else {
      raw.openings.forEach((o, j) => {
        const opening = convertOpening(o as RawOpening, wallLenUnits, `${path}.openings[${j}]`, issues)
        if (opening) openings.push(opening)
      })
    }
  }

  return {
    id: nextId('wall'),
    x1,
    y1,
    x2,
    y2,
    thickness: cmToUnits(thicknessCm),
    openings,
  }
}

function convertZone(raw: RawZone, index: number, issues: string[]): Zone | null {
  const path = `zones[${index}]`
  if (!Array.isArray(raw.points) || raw.points.length < 3) {
    issues.push(`${path}: "points" must be an array of at least 3 {x, y} points (in meters)`)
    return null
  }

  const points: Point2D[] = []
  let allValid = true
  raw.points.forEach((p: RawPoint, k: number) => {
    if (!isFiniteNumber(p?.x) || !isFiniteNumber(p?.y)) {
      issues.push(`${path}.points[${k}]: must have numeric x/y`)
      allValid = false
      return
    }
    points.push({ x: metersToUnits(p.x), y: metersToUnits(p.y) })
  })
  if (!allValid) return null

  const name = typeof raw.name === 'string' && raw.name.trim() ? raw.name.trim() : `Zone ${index + 1}`
  const color =
    typeof raw.color === 'string' && raw.color.trim() ? raw.color.trim() : ZONE_COLOR_PALETTE[index % ZONE_COLOR_PALETTE.length]

  return { id: nextId('zone'), name, color, points }
}

/**
 * parseImportedPlan converts the AI-generated floor plan JSON (coordinates in
 * meters/centimeters) into the app's internal Plan representation (coordinates
 * in plan units, 1 unit = 2.5cm). Throws PlanImportError listing every problem
 * found in the input rather than stopping at the first one.
 */
export function parseImportedPlan(raw: string, levelId: string): Plan {
  let data: RawFloorPlan
  try {
    data = JSON.parse(raw)
  } catch (err) {
    throw new PlanImportError([`Invalid JSON: ${(err as Error).message}`])
  }

  if (!data || typeof data !== 'object' || Array.isArray(data)) {
    throw new PlanImportError(['The JSON root must be an object.'])
  }

  const issues: string[] = []
  if (!Array.isArray(data.walls)) {
    issues.push('Missing or invalid "walls" array.')
  }
  if (data.zones !== undefined && !Array.isArray(data.zones)) {
    issues.push('"zones" must be an array when present.')
  }
  if (issues.length > 0) throw new PlanImportError(issues)

  const walls: WallSegment[] = []
  ;(data.walls as RawWall[]).forEach((w, i) => {
    if (!w || typeof w !== 'object') {
      issues.push(`walls[${i}]: must be an object`)
      return
    }
    const wall = convertWall(w, i, issues)
    if (wall) walls.push(wall)
  })

  const zones: Zone[] = []
  const rawZones = (data.zones as RawZone[] | undefined) || []
  rawZones.forEach((z, i) => {
    if (!z || typeof z !== 'object') {
      issues.push(`zones[${i}]: must be an object`)
      return
    }
    const zone = convertZone(z, i, issues)
    if (zone) zones.push(zone)
  })

  if (issues.length > 0) throw new PlanImportError(issues)
  if (walls.length === 0) throw new PlanImportError(['The plan must contain at least one wall.'])

  return { level_id: levelId, walls, zones }
}

export const PLAN_IMPORT_PROMPT = `You are an expert architectural draftsperson. I will attach a photo (or scan) of a single floor/level plan of a house. Study it carefully and reply with ONLY a single JSON object (no markdown code fences, no comments, no explanations before or after) describing that floor plan, following EXACTLY this schema:

{
  "version": 1,
  "unit": "m",
  "walls": [
    {
      "start": { "x": 0, "y": 0 },
      "end": { "x": 5.2, "y": 0 },
      "thickness_cm": 20,
      "openings": [
        {
          "type": "door",
          "offset_m": 1.1,
          "width_cm": 90,
          "flip_side": false,
          "flip_hinge": false,
          "hide_door": false
        }
      ]
    }
  ],
  "zones": [
    {
      "name": "Living Room",
      "color": "#3b82f6",
      "points": [
        { "x": 0, "y": 0 },
        { "x": 5.2, "y": 0 },
        { "x": 5.2, "y": 4.3 },
        { "x": 0, "y": 4.3 }
      ]
    }
  ]
}

Rules:
- This JSON represents ONE floor/level. If the photo shows several floors, process only one floor now; I will send the others separately.
- Use a single, consistent coordinate system for the whole plan, in meters, X increasing to the right and Y increasing downward (image coordinates), with an arbitrary origin (e.g. the top-left corner of the outer footprint).
- "walls": one entry per straight wall segment.
  - "start"/"end": the two endpoints of the segment, in meters.
  - Walls that meet at a corner, T-junction, or intersection MUST share the exact same coordinates at that point so they connect cleanly. Split a wall into several segments at every corner or thickness change.
  - "thickness_cm": wall thickness in centimeters. Typical values: 7-10cm for a light partition, 15-20cm for a standard interior wall, 25-40cm for an exterior or load-bearing wall. Estimate from the drawing if not labeled.
- "openings" (optional array on a wall): the doors and windows carried by that specific wall segment.
  - "type": "door" or "window".
  - "offset_m": distance in meters, measured along the wall from its "start" point, to the CENTER of the opening.
  - "width_cm": opening width in centimeters. Typical door widths: 70-90cm (>= 150cm implies a double door). Typical window widths: 60-180cm.
  - "flip_side" (doors only, optional, default false): which side of the wall the door swings open into.
  - "flip_hinge" (doors only, optional, default false): which end of the opening the hinge is on.
  - If the swing direction is not visible on the plan, omit "flip_side"/"flip_hinge" or set them to false.
  - "hide_door" (doors only, optional, default false): true when the opening is a plain passage with no door leaf (an archway or a cased opening drawn without a swing arc).
- "zones" (optional but recommended): one entry per room/area, as a closed polygon.
  - "name": the room label written on the plan, translated to English if needed (e.g. "Kitchen", "Bedroom 1", "Bathroom", "Hallway", "Garage", "Terrace"). Make up a short descriptive name if the room is unlabeled.
  - "color": a distinct hex color per room, of your choosing, so rooms are visually distinguishable.
  - "points": the polygon vertices, in meters, listed in order around the room, tracing the INTERIOR boundary of the room (the inner face of its walls). At least 3 points.
- Derive real-world dimensions from any scale bar, printed dimensions, or standard reference sizes (e.g. an 80cm door) visible on the plan. If no scale is given, make your best realistic estimate from typical room and furniture proportions.
- Do not invent walls, rooms, doors, or windows that are not visible or clearly implied on the plan.
- Output strictly valid JSON: double-quoted keys and strings, no trailing commas, no NaN/Infinity, no comments.`
