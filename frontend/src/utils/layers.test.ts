import { describe, it, expect } from 'vitest'
import {
  DEFAULT_LAYERS,
  formatLayerLabel,
  filterPlacementsByLayer,
  resolveActiveLayer,
  getDefaultLayerNames,
  getLayerByName,
  inferLayerFromDevice,
} from './layers'
import type { DevicePlacement, Device, Layer } from '../types'

describe('layers utilities', () => {
  describe('DEFAULT_LAYERS', () => {
    it('contains controls and sensors by default', () => {
      expect(DEFAULT_LAYERS).toEqual([
        { name: 'controls', hide_gauges: false },
        { name: 'sensors', hide_gauges: false },
      ])
    })
  })

  describe('getDefaultLayerNames', () => {
    it('returns array of default layer names', () => {
      expect(getDefaultLayerNames()).toEqual(['controls', 'sensors'])
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

  describe('getLayerByName', () => {
    const layers: Layer[] = [
      { name: 'controls', hide_gauges: false },
      { name: 'sensors', hide_gauges: true },
    ]

    it('returns the layer if found', () => {
      expect(getLayerByName(layers, 'controls')).toEqual({ name: 'controls', hide_gauges: false })
      expect(getLayerByName(layers, 'sensors')).toEqual({ name: 'sensors', hide_gauges: true })
    })

    it('returns undefined if not found', () => {
      expect(getLayerByName(layers, 'unknown')).toBeUndefined()
    })
  })

  describe('resolveActiveLayer', () => {
    const availableLayers: Layer[] = [
      { name: 'controls', hide_gauges: false },
      { name: 'sensors', hide_gauges: false },
      { name: 'lights', hide_gauges: false },
    ]

    it('returns the current layer if it exists in available layers', () => {
      expect(resolveActiveLayer('sensors', availableLayers)).toBe('sensors')
    })

    it('falls back to the first available layer if current layer is not in available layers', () => {
      expect(resolveActiveLayer('unknown', [
        { name: 'security', hide_gauges: false },
        { name: 'cameras', hide_gauges: false },
      ])).toBe('security')
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

  describe('inferLayerFromDevice', () => {
    const deviceMap: Record<string, Device> = {
      'sensor.living_temp': {
        id: 'sensor.living_temp', name: 'Temp', domain: 'sensor', state: '21.5', attributes: {}, last_updated: '2026-01-01T00:00:00Z',
      },
      'climate.thermostat': {
        id: 'climate.thermostat', name: 'Thermostat', domain: 'climate', state: 'heat', attributes: {}, last_updated: '2026-01-01T00:00:00Z',
      },
      'light.ceiling': {
        id: 'light.ceiling', name: 'Ceiling Light', domain: 'light', state: 'off', attributes: {}, last_updated: '2026-01-01T00:00:00Z',
      },
    }

    it('returns "sensors" for sensor domain via deviceMap', () => {
      expect(inferLayerFromDevice('sensor.living_temp', deviceMap)).toBe('sensors')
    })

    it('returns "controls" for climate domain via deviceMap', () => {
      expect(inferLayerFromDevice('climate.thermostat', deviceMap)).toBe('controls')
    })

    it('returns "controls" for light domain via deviceMap', () => {
      expect(inferLayerFromDevice('light.ceiling', deviceMap)).toBe('controls')
    })

    it('falls back to entity_id parsing when device not in deviceMap', () => {
      expect(inferLayerFromDevice('sensor.bedroom_temp', {})).toBe('sensors')
      expect(inferLayerFromDevice('climate.basement', {})).toBe('controls')
      expect(inferLayerFromDevice('switch.garden', {})).toBe('controls')
    })

    it('returns null for invalid entity_id format', () => {
      expect(inferLayerFromDevice('invalid', {})).toBeNull()
      expect(inferLayerFromDevice('', {})).toBeNull()
    })

    it('uses deviceMap domain even when available (never falls back if device exists)', () => {
      // Device has domain 'sensor' even though entity_id parsing would suggest 'climate'
      const customMap: Record<string, Device> = {
        'climate.odd_name': {
          id: 'climate.odd_name', name: 'Weird', domain: 'sensor', state: '21', attributes: {}, last_updated: '2026-01-01T00:00:00Z',
        },
      }
      expect(inferLayerFromDevice('climate.odd_name', customMap)).toBe('sensors')
    })
  })
})