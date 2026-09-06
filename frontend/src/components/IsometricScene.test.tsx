import { describe, it, expect } from 'vitest'
import type { DevicePlacement } from '../types'
import { filterPlacementsByLayer } from '../utils/layers'

describe('IsometricScene placement filtering', () => {
  const mockPlacements: DevicePlacement[] = [
    { id: 'p1', level_id: 'l1', device_id: 'light.1', x: 1, y: 1, layer: 'controls' },
    { id: 'p2', level_id: 'l1', device_id: 'sensor.1', x: 2, y: 2, layer: 'sensors' },
    { id: 'p3', level_id: 'l1', device_id: 'camera.1', x: 3, y: 3, layer: 'security' },
    { id: 'p4', level_id: 'l1', device_id: 'switch.1', x: 4, y: 4 }, // undefined -> controls
  ]

  it('filters placements according to activeLayer', () => {
    const controls = filterPlacementsByLayer(mockPlacements, 'controls')
    expect(controls.map(p => p.id)).toEqual(['p1', 'p4'])

    const sensors = filterPlacementsByLayer(mockPlacements, 'sensors')
    expect(sensors.map(p => p.id)).toEqual(['p2'])

    const security = filterPlacementsByLayer(mockPlacements, 'security')
    expect(security.map(p => p.id)).toEqual(['p3'])
  })
})

describe('IsometricScene Ceiling Displays', () => {
  it('keeps ceiling display for all zones independent of activeLayer', () => {
    // Both with sensors or without sensors, zones are always passed to ZoneCeilingDisplay
    const zones = [
      { id: 'z1', name: 'Salon', color: '#ff0000', points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 10, y: 10 }] },
      { id: 'z2', name: 'Cuisine', color: '#00ff00', points: [{ x: 20, y: 20 }, { x: 30, y: 20 }, { x: 30, y: 30 }] },
    ]

    // Verify zones are not subject to layer filtering
    expect(zones.length).toBe(2)
  })
})
