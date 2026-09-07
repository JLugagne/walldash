import { describe, it, expect } from 'vitest'
import { resolveParamLevelId } from './levelSync'
import type { Level } from '../types'

const levels: Level[] = [
  { id: 'a', name: 'A', order: 1, is_outdoor: false, created_at: '', updated_at: '' },
  { id: 'b', name: 'B', order: 2, is_outdoor: false, created_at: '', updated_at: '' },
]

describe('resolveParamLevelId', () => {
  it('returns null when there is no param', () => {
    expect(resolveParamLevelId(undefined, levels, 'a')).toBeNull()
  })

  it('returns null when the param matches the active level', () => {
    expect(resolveParamLevelId('a', levels, 'a')).toBeNull()
  })

  it('returns the param when it selects another known level', () => {
    expect(resolveParamLevelId('b', levels, 'a')).toBe('b')
  })

  it('returns null for a stale param (deleted or unknown level)', () => {
    expect(resolveParamLevelId('ghost', levels, 'a')).toBeNull()
    expect(resolveParamLevelId('ghost', levels, null)).toBeNull()
  })

  it('returns null when no levels are loaded yet', () => {
    expect(resolveParamLevelId('a', [], null)).toBeNull()
  })
})
