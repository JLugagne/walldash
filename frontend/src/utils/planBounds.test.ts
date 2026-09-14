import { describe, it, expect } from 'vitest'
import type { DevicePlacement, Plan } from '../types'
import { computePlanBounds, DEFAULT_PLAN_BOUNDS } from './planBounds'

const emptyPlan: Plan = { level_id: 'l1', walls: [], zones: [] }

function placement(id: string, x: number, y: number): DevicePlacement {
  return { id, level_id: 'l1', device_id: `light.${id}`, x, y, layer: 'controls' }
}

describe('computePlanBounds', () => {
  it('returns the default bounds when there is no plan and no placements', () => {
    expect(computePlanBounds(null, [])).toEqual(DEFAULT_PLAN_BOUNDS)
    expect(computePlanBounds(emptyPlan, [])).toEqual(DEFAULT_PLAN_BOUNDS)
  })

  it('includes device placements even when the plan has no walls or zones', () => {
    // SVG (1000, 700) -> world (20, 14): outside the default bounds, so the
    // camera must widen to include the device instead of leaving it off-frame.
    const bounds = computePlanBounds(emptyPlan, [placement('p1', 1000, 700)])
    expect(bounds.maxX).toBeGreaterThan(DEFAULT_PLAN_BOUNDS.maxX)
    expect(bounds.maxZ).toBeGreaterThan(DEFAULT_PLAN_BOUNDS.maxZ)
  })

  it('still covers placements outside the walls bounding box', () => {
    const plan: Plan = {
      level_id: 'l1',
      walls: [{ id: 'w1', x1: 450, y1: 300, x2: 550, y2: 300, thickness: 12, openings: [] }],
      zones: [],
    }
    const bounds = computePlanBounds(plan, [placement('p1', 1000, 700)])
    expect(bounds.maxX).toBeGreaterThan(10)
    expect(bounds.maxZ).toBeGreaterThan(7)
  })

  it('pads the fitted content so outer wall faces stay in frame', () => {
    const plan: Plan = {
      level_id: 'l1',
      walls: [{ id: 'w1', x1: 500, y1: 350, x2: 600, y2: 350, thickness: 12, openings: [] }],
      zones: [],
    }
    // Wall x maps to world [0, 4]: bounds must extend 0.6 beyond on each side.
    const bounds = computePlanBounds(plan, [])
    expect(bounds.minX).toBeCloseTo(-0.6)
    expect(bounds.maxX).toBeCloseTo(4.6)
  })
})
