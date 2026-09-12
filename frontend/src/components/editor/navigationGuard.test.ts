import { describe, it, expect } from 'vitest'
import { shouldBlockPlanExit } from './navigationGuard'

describe('shouldBlockPlanExit', () => {
  it('blocks leaving the plan editor when there are unsaved changes', () => {
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/floor/lvl-1', true)).toBe(true)
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/dashboards', true)).toBe(true)
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/', true)).toBe(true)
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/setup/settings', true)).toBe(true)
  })

  it('never blocks a saved plan', () => {
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/floor/lvl-1', false)).toBe(false)
  })

  it('does not block switching levels inside the editor', () => {
    expect(shouldBlockPlanExit('/setup/plans/lvl-1', '/setup/plans/lvl-2', true)).toBe(false)
  })

  it('does not block when already outside the editor', () => {
    expect(shouldBlockPlanExit('/dashboards', '/setup/plans/lvl-1', true)).toBe(false)
    expect(shouldBlockPlanExit('/', '/floor/lvl-1', true)).toBe(false)
  })
})
