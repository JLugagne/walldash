import { describe, expect, it } from 'vitest'
import type { Device, DevicePlacement, Level } from '../types'
import { configForLevel, filterLightPlacements, floorElevation, orderedLevels } from './houseOverview'

const placement = (id: string, deviceId: string, renderDomain?: string): DevicePlacement => ({
  id,
  level_id: 'level-1',
  device_id: deviceId,
  x: 0,
  y: 0,
  render_domain: renderDomain,
})

describe('house overview configuration', () => {
  it('uses visible floors by default and preserves explicit offsets', () => {
    expect(configForLevel({ floor: { x: 12, y: -4, visible: false } }, 'floor')).toEqual({ x: 12, y: -4, visible: false })
    expect(configForLevel({}, 'new-floor')).toEqual({ x: 0, y: 0, visible: true })
  })

  it('orders floors from their configured order', () => {
    const levels = [
      { id: 'upper', order: 2 } as Level,
      { id: 'ground', order: 0 } as Level,
      { id: 'middle', order: 1 } as Level,
    ]
    expect(orderedLevels(levels).map((level) => level.id)).toEqual(['ground', 'middle', 'upper'])
  })

  it('assigns a distinct elevation from the sorted floor index', () => {
    expect(floorElevation(0)).toBe(0)
    expect(floorElevation(1)).toBe(3.2)
    expect(floorElevation(3)).toBeCloseTo(9.6)
  })

  it('keeps lights from every layer and excludes sensors and actuators', () => {
    const devices: Record<string, Device> = {
      'light.ceiling': { id: 'light.ceiling', name: 'Ceiling', domain: 'light', state: 'on', attributes: {}, last_updated: '' },
      'sensor.room': { id: 'sensor.room', name: 'Room', domain: 'sensor', state: '20', attributes: {}, last_updated: '' },
    }
    const placements = [
      placement('light-placement', 'light.ceiling', 'light'),
      placement('sensor-placement', 'sensor.room', 'sensor'),
      placement('fallback-light', 'light.garden'),
    ]
    expect(filterLightPlacements(placements, devices).map((item) => item.id)).toEqual(['light-placement', 'fallback-light'])
  })
})
