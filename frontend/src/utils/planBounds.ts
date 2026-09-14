import type { DevicePlacement, Plan } from '../types'
import { toWorldX, toWorldZ } from '../components/wallGeometry'

export interface PlanBounds {
  minX: number
  maxX: number
  minZ: number
  maxZ: number
}

export const DEFAULT_PLAN_BOUNDS: PlanBounds = { minX: -10, maxX: 10, minZ: -7, maxZ: 7 }

/** Padding around the fitted content so outer wall faces stay in frame. */
const WALL_PADDING = 0.6

/**
 * Bounding box of a floor plan in 3D world space, used to fit the camera.
 * Covers walls, zones and device placements.
 */
export function computePlanBounds(plan: Plan | null, placements: DevicePlacement[] = []): PlanBounds {
  let minX = Infinity
  let maxX = -Infinity
  let minZ = Infinity
  let maxZ = -Infinity
  for (const w of plan?.walls || []) {
    minX = Math.min(minX, toWorldX(w.x1), toWorldX(w.x2))
    maxX = Math.max(maxX, toWorldX(w.x1), toWorldX(w.x2))
    minZ = Math.min(minZ, toWorldZ(w.y1), toWorldZ(w.y2))
    maxZ = Math.max(maxZ, toWorldZ(w.y1), toWorldZ(w.y2))
  }
  for (const z of plan?.zones || []) {
    for (const p of z.points) {
      minX = Math.min(minX, toWorldX(p.x))
      maxX = Math.max(maxX, toWorldX(p.x))
      minZ = Math.min(minZ, toWorldZ(p.y))
      maxZ = Math.max(maxZ, toWorldZ(p.y))
    }
  }
  for (const p of placements) {
    const wx = toWorldX(p.x)
    const wz = toWorldZ(p.y)
    minX = Math.min(minX, wx)
    maxX = Math.max(maxX, wx)
    minZ = Math.min(minZ, wz)
    maxZ = Math.max(maxZ, wz)
  }
  if (!isFinite(minX) || !isFinite(maxX) || !isFinite(minZ) || !isFinite(maxZ)) {
    return { ...DEFAULT_PLAN_BOUNDS }
  }

  // Expand by 0.6 units to include outer wall faces and cylinder corners
  return {
    minX: minX - WALL_PADDING,
    maxX: maxX + WALL_PADDING,
    minZ: minZ - WALL_PADDING,
    maxZ: maxZ + WALL_PADDING,
  }
}
