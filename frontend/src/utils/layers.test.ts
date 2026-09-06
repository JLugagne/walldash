import { describe, it, expect } from 'vitest'
import {
  DEFAULT_LAYERS,
  formatLayerLabel,
  filterPlacementsByLayer,
  resolveActiveLayer,
} from './layers'
import type { DevicePlacement } from '../types'

describe('layers utilities', () => {
  describe('DEFAULT_LAYERS', () => {
    it('contains controls and sensors by default', () => {
      expect(DEFAULT_LAYERS).toEqual(['controls', 'sensors'])
    })
  })

  describe('formatLayerLabel', () => {
    it('capitalizes the first letter of a layer name', () => {
      expect(formatLayerLabel('controls')).toBe('Controls')
      expect(formatLayerLabel('sensors')).toBe('Sensors')
      expect(formatLayerLabel('security')).toBe('Security')
      expect(formatLayerLabel('hvac')).toBe('Hvac')
    })

    it('handles empty string and already capitalized strings', () => {
      expect(formatLayerLabel('')).toBe('')
      expect(formatLayerLabel('Controls')).toBe('Controls')
    })
  })

  describe('resolveActiveLayer', () => {
    it('returns the current layer if it exists in available layers', () => {
      expect(resolveActiveLayer('sensors', ['controls', 'sensors', 'lights'])).toBe('sensors')
    })

    it('falls back to the first available layer if current layer is not in available layers', () => {
      expect(resolveActiveLayer('unknown', ['security', 'cameras'])).toBe('security')
    })

    it('falls back to the first layer of DEFAULT_LAYERS if available layers is empty or undefined', () => {
      expect(resolveActiveLayer('unknown', [])).toBe('controls')
      expect(resolveActiveLayer('unknown', undefined)).toBe('controls')
      expect(resolveActiveLayer(null, undefined)).toBe('controls')
    })

    it('keeps active layer if it matches one of the default layers when available layers is undefined', () => {
      expect(resolveActiveLayer('sensors', undefined)).toBe('sensors')
      expect(resolveActiveLayer('sensors', [])).toBe('sensors')
    })
  })

  describe('filterPlacementsByLayer', () => {
    const placements: DevicePlacement[] = [
      { id: '1', level_id: 'lvl1', device_id: 'light.living', x: 10, y: 10, layer: 'controls' },
      { id: '2', level_id: 'lvl1', device_id: 'sensor.temp', x: 20, y: 20, layer: 'sensors' },
      { id: '3', level_id: 'lvl1', device_id: 'light.kitchen', x: 30, y: 30 }, // unassigned layer -> defaults to 'controls'
      { id: '4', level_id: 'lvl1', device_id: 'camera.front', x: 40, y: 40, layer: 'security' },
    ]

    it('filters placements for controls layer, including unassigned placements', () => {
      const result = filterPlacementsByLayer(placements, 'controls')
      expect(result.map((p) => p.id)).toEqual(['1', '3'])
    })

    it('filters placements for sensors layer', () => {
      const result = filterPlacementsByLayer(placements, 'sensors')
      expect(result.map((p) => p.id)).toEqual(['2'])
    })

    it('filters placements for custom security layer', () => {
      const result = filterPlacementsByLayer(placements, 'security')
      expect(result.map((p) => p.id)).toEqual(['4'])
    })

    it('returns empty array when no placement matches active layer', () => {
      const result = filterPlacementsByLayer(placements, 'hvac')
      expect(result).toEqual([])
    })

    it('handles empty placements array', () => {
      const result = filterPlacementsByLayer([], 'controls')
      expect(result).toEqual([])
    })
  })
})
