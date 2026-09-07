import { describe, it, expect } from 'vitest'
import type { Zone, Device } from '../types'
import {
  formatZoneMetrics,
  getZoneTemperature,
  getZoneCeilingText,
  calculateCeilingColor,
  hasConfiguredSensors,
  getZoneHUDMetrics,
  calculateDiscDiameter,
  computeZoneGaugesHidden,
  isCeilingHiddenForActiveLayer,
} from './ceilingDisplay'

describe('ceilingDisplay utils', () => {
  const mockDevices: Record<string, Device> = {
    'sensor.salon_temp': {
      id: 'sensor.salon_temp',
      name: 'Living Room Temperature',
      domain: 'sensor',
      state: '21.4',
      attributes: { unit_of_measurement: '°C' },
      last_updated: '2026-09-05T12:00:00Z',
    },
    'sensor.salon_humidity': {
      id: 'sensor.salon_humidity',
      name: 'Living Room Humidity',
      domain: 'sensor',
      state: '48',
      attributes: { unit_of_measurement: '%' },
      last_updated: '2026-09-05T12:00:00Z',
    },
    'climate.thermostat': {
      id: 'climate.thermostat',
      name: 'Thermostat',
      domain: 'climate',
      state: 'heat',
      attributes: { current_temperature: 19.5, temperature: 21 },
      last_updated: '2026-09-05T12:00:00Z',
    },
    'sensor.unavailable_temp': {
      id: 'sensor.unavailable_temp',
      name: 'Unknown Temp',
      domain: 'sensor',
      state: 'unavailable',
      attributes: {},
      last_updated: '2026-09-05T12:00:00Z',
    },
  }

  describe('hasConfiguredSensors', () => {
    it('returns false when neither temp_sensor nor humidity_sensor is set', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Kitchen',
        color: '#ff0000',
        points: [],
      }
      expect(hasConfiguredSensors(zone)).toBe(false)
    })

    it('returns true when temp_sensor is configured', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Living Room',
        color: '#ff0000',
        points: [],
        temp_sensor: 'sensor.salon_temp',
      }
      expect(hasConfiguredSensors(zone)).toBe(true)
    })

    it('returns true when humidity_sensor is configured', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Bathroom',
        color: '#ff0000',
        points: [],
        humidity_sensor: 'sensor.salon_humidity',
      }
      expect(hasConfiguredSensors(zone)).toBe(true)
    })
  })

  describe('formatZoneMetrics', () => {
    it('returns null when no sensors configured for the zone', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Hallway',
        color: '#ffffff',
        points: [],
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBeNull()
    })

    it('formats temperature alone when only temp_sensor is configured', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Living Room',
        color: '#ffffff',
        points: [],
        temp_sensor: 'sensor.salon_temp',
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBe('21.4°C')
    })

    it('formats humidity alone when only humidity_sensor is configured', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Living Room',
        color: '#ffffff',
        points: [],
        humidity_sensor: 'sensor.salon_humidity',
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBe('48%')
    })

    it('formats combined metrics "21.4°C · 48%" when both are present', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Living Room',
        color: '#ffffff',
        points: [],
        temp_sensor: 'sensor.salon_temp',
        humidity_sensor: 'sensor.salon_humidity',
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBe('21.4°C · 48%')
    })

    it('extracts temperature from climate current_temperature', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Bedroom',
        color: '#ffffff',
        points: [],
        temp_sensor: 'climate.thermostat',
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBe('19.5°C')
    })

    it('returns placeholder when configured sensor is unavailable', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Garage',
        color: '#ffffff',
        points: [],
        temp_sensor: 'sensor.unavailable_temp',
      }
      expect(formatZoneMetrics(zone, mockDevices)).toBe('--')
    })
  })

  describe('getZoneTemperature', () => {
    it('returns null when no temp_sensor is configured', () => {
      const zone: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [] }
      expect(getZoneTemperature(zone, mockDevices)).toBeNull()
    })

    it('returns numeric temperature from standard sensor', () => {
      const zone: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [], temp_sensor: 'sensor.salon_temp' }
      expect(getZoneTemperature(zone, mockDevices)).toBe(21.4)
    })

    it('returns numeric temperature from climate entity', () => {
      const zone: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [], temp_sensor: 'climate.thermostat' }
      expect(getZoneTemperature(zone, mockDevices)).toBe(19.5)
    })

    it('returns null when sensor state is non-numeric or missing', () => {
      const zone: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [], temp_sensor: 'sensor.unavailable_temp' }
      expect(getZoneTemperature(zone, mockDevices)).toBeNull()

      const zoneMissing: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [], temp_sensor: 'sensor.does_not_exist' }
      expect(getZoneTemperature(zoneMissing, mockDevices)).toBeNull()
    })
  })

  describe('getZoneCeilingText', () => {
    it('returns uppercase zone name when no sensors are configured', () => {
      const zone: Zone = { id: 'z1', name: 'Tea Room', color: '#fff', points: [] }
      expect(getZoneCeilingText(zone, mockDevices)).toBe('TEA ROOM')
    })

    it('returns formatted metrics and hides zone name when sensors are configured', () => {
      const zone: Zone = {
        id: 'z1',
        name: 'Living Room',
        color: '#fff',
        points: [],
        temp_sensor: 'sensor.salon_temp',
        humidity_sensor: 'sensor.salon_humidity',
      }
      expect(getZoneCeilingText(zone, mockDevices)).toBe('21.4°C · 48%')
      expect(getZoneCeilingText(zone, mockDevices)).not.toContain('Salon')
    })
  })

  describe('calculateCeilingColor & Zone Alert Pulse triggers', () => {
    it('returns neutral color #f1f5f9 and isAlert false when no thresholds are defined', () => {
      const result = calculateCeilingColor(21.4)
      expect(result.color.toLowerCase()).toBe('#f1f5f9')
      expect(result.isAlert).toBe(false)
    })

    it('returns neutral color and isAlert false when thresholds are explicitly null or undefined', () => {
      const result1 = calculateCeilingColor(21.4, null, null)
      expect(result1.color.toLowerCase()).toBe('#f1f5f9')
      expect(result1.isAlert).toBe(false)

      const result2 = calculateCeilingColor(21.4, undefined, null)
      expect(result2.color.toLowerCase()).toBe('#f1f5f9')
      expect(result2.isAlert).toBe(false)

      const result3 = calculateCeilingColor(21.4, null, undefined)
      expect(result3.color.toLowerCase()).toBe('#f1f5f9')
      expect(result3.isAlert).toBe(false)
    })

    it('returns neutral color #f1f5f9 and isAlert false when temperature is null', () => {
      const result = calculateCeilingColor(null, 18, 24)
      expect(result.color.toLowerCase()).toBe('#f1f5f9')
      expect(result.isAlert).toBe(false)
    })

    it('returns neutral color and isAlert false when within thresholds [18, 24]', () => {
      const result = calculateCeilingColor(21.0, 18, 24)
      expect(result.color.toLowerCase()).toBe('#f1f5f9')
      expect(result.isAlert).toBe(false)
    })

    it('triggers cold alert and interpolates towards cyan neon when temp < temp_min', () => {
      // At temp_min - 3 (maximum cold saturation)
      const fullCold = calculateCeilingColor(15.0, 18.0, 24.0)
      expect(fullCold.isAlert).toBe(true)
      expect(fullCold.alertType).toBe('cold')
      expect(fullCold.color.toLowerCase()).toBe('#38bdf8')

      // At temp_min - 1 (partial progression)
      const partialCold = calculateCeilingColor(17.0, 18.0, 24.0)
      expect(partialCold.isAlert).toBe(true)
      expect(partialCold.alertType).toBe('cold')
      expect(partialCold.color).not.toBe('#f1f5f9')
      expect(partialCold.color).not.toBe('#38bdf8')
    })

    it('triggers hot alert and interpolates towards red neon when temp > temp_max', () => {
      // At temp_max + 3 (maximum heat saturation)
      const fullHot = calculateCeilingColor(27.0, 18.0, 24.0)
      expect(fullHot.isAlert).toBe(true)
      expect(fullHot.alertType).toBe('hot')
      expect(fullHot.color.toLowerCase()).toBe('#ef4444')

      // At temp_max + 1 (partial progression)
      const partialHot = calculateCeilingColor(25.0, 18.0, 24.0)
      expect(partialHot.isAlert).toBe(true)
      expect(partialHot.alertType).toBe('hot')
      expect(partialHot.color).not.toBe('#f1f5f9')
      expect(partialHot.color).not.toBe('#ef4444')
    })
  })

  describe('getZoneHUDMetrics', () => {
    it('returns minimal metrics when no sensors configured on zone', () => {
      const zone: Zone = {
        id: 'z-empty',
        name: 'Hallway',
        color: '#ff0000',
        points: [],
      }
      const metrics = getZoneHUDMetrics(zone, mockDevices)
      expect(metrics.zoneName).toBe('HALLWAY')
      expect(metrics.hasSensors).toBe(false)
      expect(metrics.tempValue).toBeNull()
      expect(metrics.humidityValue).toBeNull()
      expect(metrics.humidityPct).toBeNull()
      expect(metrics.statusText).toBe('')
      expect(metrics.isAlert).toBe(false)
    })

    it('returns comfort state when temperature is in normal range', () => {
      const zone: Zone = {
        id: 'z-living',
        name: 'Living Room',
        color: '#3b82f6',
        points: [],
        temp_sensor: 'sensor.salon_temp',
        humidity_sensor: 'sensor.salon_humidity',
        temp_min: 19,
        temp_max: 23,
      }
      const metrics = getZoneHUDMetrics(zone, mockDevices)
      expect(metrics.zoneName).toBe('LIVING ROOM')
      expect(metrics.hasSensors).toBe(true)
      expect(metrics.tempValue).toBe('21.4')
      expect(metrics.humidityValue).toBe('48% RH')
      expect(metrics.humidityPct).toBe(48)
      expect(metrics.statusText).toBe('COMFORT')
      expect(metrics.accentColor).toBe('#2dd4bf')
      expect(metrics.isAlert).toBe(false)
    })

    it('returns cold alert state when temp < temp_min', () => {
      const zone: Zone = {
        id: 'z-cold',
        name: 'Office',
        color: '#3b82f6',
        points: [],
        temp_sensor: 'sensor.salon_temp', // 21.4
        temp_min: 22.5, // 21.4 < 22.5
        temp_max: 26,
      }
      const metrics = getZoneHUDMetrics(zone, mockDevices)
      expect(metrics.statusText).toBe('TOO COOL')
      expect(metrics.accentColor).toBe('#38bdf8')
      expect(metrics.isAlert).toBe(true)
      expect(metrics.alertType).toBe('cold')
    })

    it('returns hot alert state when temp > temp_max', () => {
      const zone: Zone = {
        id: 'z-hot',
        name: 'Kitchen',
        color: '#3b82f6',
        points: [],
        temp_sensor: 'sensor.salon_temp', // 21.4
        temp_min: 18,
        temp_max: 20, // 21.4 > 20
      }
      const metrics = getZoneHUDMetrics(zone, mockDevices)
      expect(metrics.statusText).toBe('TOO WARM')
      expect(metrics.accentColor).toBe('#f87171')
      expect(metrics.isAlert).toBe(true)
      expect(metrics.alertType).toBe('hot')
    })

    it('progressively interpolates humidity gauge color based on thresholds', () => {
      // In comfort range (48% between 40% and 60%): comfort teal
      const comfortZone: Zone = {
        id: 'z-comf',
        name: 'Bedroom',
        color: '#3b82f6',
        points: [],
        humidity_sensor: 'sensor.salon_humidity', // 48%
        humidity_min: 40,
        humidity_max: 60,
      }
      const comfMetrics = getZoneHUDMetrics(comfortZone, mockDevices)
      expect(comfMetrics.humidityColor).toBe('#2dd4bf')

      // Partial dry transition (48% with min = 53%, delta = 5%, half transition)
      const partialDryZone: Zone = {
        id: 'z-partial-dry',
        name: 'Bedroom',
        color: '#3b82f6',
        points: [],
        humidity_sensor: 'sensor.salon_humidity', // 48%
        humidity_min: 53,
      }
      const partialDryMetrics = getZoneHUDMetrics(partialDryZone, mockDevices)
      expect(partialDryMetrics.humidityColor).not.toBe('#2dd4bf')
      expect(partialDryMetrics.humidityColor).not.toBe('#38bdf8')

      // Fully saturated dry (48% with min = 60%, delta = 12% >= 10% transition window)
      const fullDryZone: Zone = {
        id: 'z-full-dry',
        name: 'Bedroom',
        color: '#3b82f6',
        points: [],
        humidity_sensor: 'sensor.salon_humidity', // 48%
        humidity_min: 60,
      }
      const fullDryMetrics = getZoneHUDMetrics(fullDryZone, mockDevices)
      expect(fullDryMetrics.humidityColor).toBe('#38bdf8')

      // Partial humid transition (48% with max = 43%, delta = 5%, half transition)
      const partialHumidZone: Zone = {
        id: 'z-partial-humid',
        name: 'Bedroom',
        color: '#3b82f6',
        points: [],
        humidity_sensor: 'sensor.salon_humidity', // 48%
        humidity_max: 43,
      }
      const partialHumidMetrics = getZoneHUDMetrics(partialHumidZone, mockDevices)
      expect(partialHumidMetrics.humidityColor).not.toBe('#2dd4bf')
      expect(partialHumidMetrics.humidityColor).not.toBe('#ef4444')

      // Fully saturated humid (48% with max = 35%, delta = 13% >= 10% transition window)
      const fullHumidZone: Zone = {
        id: 'z-full-humid',
        name: 'Bedroom',
        color: '#3b82f6',
        points: [],
        humidity_sensor: 'sensor.salon_humidity', // 48%
        humidity_max: 35,
      }
      const fullHumidMetrics = getZoneHUDMetrics(fullHumidZone, mockDevices)
      expect(fullHumidMetrics.humidityColor).toBe('#ef4444')
    })
  })

  describe('calculateDiscDiameter', () => {
    it('returns default 3.0m when zone points are empty', () => {
      const zone: Zone = { id: 'z1', name: 'Z', color: '#fff', points: [] }
      expect(calculateDiscDiameter(zone)).toBe(3.0)
    })

    it('adapts and downscales for smaller rooms so it does not overflow walls', () => {
      // 100 x 100 pixels = 2m x 2m in world coords (SCALE = 0.02)
      const smallZone: Zone = {
        id: 'z-small',
        name: 'Small Room',
        color: '#fff',
        points: [
          { x: 0, y: 0 },
          { x: 100, y: 0 },
          { x: 100, y: 100 },
          { x: 0, y: 100 },
        ],
      }
      const diameter = calculateDiscDiameter(smallZone)
      // 2m * 1.0 = 2.0m
      expect(diameter).toBeCloseTo(2.0, 1)
      expect(diameter).toBeLessThan(3.0)
    })

    it('caps at 3.0m maximum for large rooms', () => {
      // 500 x 400 pixels = 10m x 8m in world coords
      const largeZone: Zone = {
        id: 'z-large',
        name: 'Large Room',
        color: '#fff',
        points: [
          { x: 0, y: 0 },
          { x: 500, y: 0 },
          { x: 500, y: 400 },
          { x: 0, y: 400 },
        ],
      }
      expect(calculateDiscDiameter(largeZone)).toBe(3.0)
    })
  })

  describe('computeZoneGaugesHidden', () => {
    // Zones don't carry their own layer. A zone's ceiling badge follows
    // the SAME layer mechanism as devices: the layer of the DevicePlacement
    // for the zone's temp_sensor/humidity_sensor entity. A sensor with no
    // placement defaults to "controls", exactly like an unplaced device.

    it('hides zone when its temp_sensor is placed on a layer with hide_gauges=true', () => {
      const zones: Zone[] = [{
        id: 'z1', name: 'Living', color: '#fff', points: [],
        temp_sensor: 'sensor.living_temp',
      }]
      const placementLayerMap = { 'sensor.living_temp': 'climate' }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { climate: true, controls: false })
      expect(result['z1']).toBe(true)
    })

    it('shows zone when its temp_sensor is placed on a layer with hide_gauges=false', () => {
      const zones: Zone[] = [{
        id: 'z2', name: 'Bedroom', color: '#fff', points: [],
        temp_sensor: 'sensor.living_temp',
      }]
      const placementLayerMap = { 'sensor.living_temp': 'climate' }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { climate: false, controls: true })
      expect(result['z2']).toBe(false)
    })

    it('treats a sensor with no placement as being on the "controls" layer', () => {
      const zones: Zone[] = [{
        id: 'z3', name: 'Kitchen', color: '#fff', points: [],
        temp_sensor: 'sensor.kitchen_temp',
      }]
      const result = computeZoneGaugesHidden(zones, {}, { controls: true, sensors: false })
      expect(result['z3']).toBe(true)
    })

    it('does not hide when the sensor placement layer is not the one with hide_gauges=true', () => {
      const zones: Zone[] = [{
        id: 'z4', name: 'Office', color: '#fff', points: [],
        temp_sensor: 'climate.thermo',
      }]
      const placementLayerMap = { 'climate.thermo': 'sensors' }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { sensors: false, controls: true })
      expect(result['z4']).toBe(false)
    })

    it('hides zone when ONLY humidity_sensor is placed on a hidden layer', () => {
      const zones: Zone[] = [{
        id: 'z5', name: 'Bathroom', color: '#fff', points: [],
        humidity_sensor: 'sensor.bedroom_hum',
      }]
      const placementLayerMap = { 'sensor.bedroom_hum': 'climate' }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { climate: true })
      expect(result['z5']).toBe(true)
    })

    it('hides zone when EITHER sensor is on a hidden layer (temp hidden, humidity visible)', () => {
      const zones: Zone[] = [{
        id: 'z8', name: 'Hall', color: '#fff', points: [],
        temp_sensor: 'sensor.living_temp',
        humidity_sensor: 'sensor.bedroom_hum',
      }]
      const placementLayerMap = {
        'sensor.living_temp': 'climate',
        'sensor.bedroom_hum': 'sensors',
      }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { climate: true, sensors: false })
      expect(result['z8']).toBe(true)
    })

    it('does not include zones without any configured sensor in the result', () => {
      const zones: Zone[] = [{
        id: 'z6', name: 'Closet', color: '#fff', points: [],
      }]
      const result = computeZoneGaugesHidden(zones, {}, { controls: true })
      expect(result['z6']).toBeUndefined()
    })

    it('handles a combination of hidden and visible zones across different layers', () => {
      const zones: Zone[] = [
        { id: 'visible', name: 'Room A', color: '#fff', points: [], temp_sensor: 'sensor.room_a' },
        { id: 'hidden', name: 'Room B', color: '#fff', points: [], temp_sensor: 'sensor.room_b' },
        { id: 'nosensor', name: 'Room C', color: '#fff', points: [] },
      ]
      const placementLayerMap = {
        'sensor.room_a': 'controls',
        'sensor.room_b': 'climate',
      }
      const result = computeZoneGaugesHidden(zones, placementLayerMap, { climate: true, controls: false })
      expect(result['visible']).toBe(false)
      expect(result['hidden']).toBe(true)
      expect(result['nosensor']).toBeUndefined()
    })

    it('is driven purely by the placed layer, not by which layer is "default" — toggling a second custom layer only affects sensors placed on it', () => {
      const zones: Zone[] = [
        { id: 'onDefault', name: 'Room A', color: '#fff', points: [], temp_sensor: 'sensor.room_a' }, // unplaced -> controls
        { id: 'onCustom', name: 'Room B', color: '#fff', points: [], temp_sensor: 'sensor.room_b' },
      ]
      const placementLayerMap = { 'sensor.room_b': 'custom' }

      // hide_gauges only on the custom layer: only the zone placed on it hides
      const result1 = computeZoneGaugesHidden(zones, placementLayerMap, { controls: false, custom: true })
      expect(result1['onDefault']).toBe(false)
      expect(result1['onCustom']).toBe(true)

      // hide_gauges only on controls: only the unplaced (default) zone hides
      const result2 = computeZoneGaugesHidden(zones, placementLayerMap, { controls: true, custom: false })
      expect(result2['onDefault']).toBe(true)
      expect(result2['onCustom']).toBe(false)
    })
  })

  describe('isCeilingHiddenForActiveLayer', () => {
    it('hides ceiling when the active layer has hide_gauges=true', () => {
      expect(isCeilingHiddenForActiveLayer('sensors', { controls: false, sensors: true })).toBe(true)
    })

    it('shows ceiling when the active layer has hide_gauges=false', () => {
      expect(isCeilingHiddenForActiveLayer('controls', { controls: false, sensors: true })).toBe(false)
    })

    it('is attached to each layer: toggling one layer does not affect another', () => {
      const map = { controls: false, custom: true }
      expect(isCeilingHiddenForActiveLayer('custom', map)).toBe(true)
      expect(isCeilingHiddenForActiveLayer('controls', map)).toBe(false)
    })

    it('shows ceiling when activeLayer is undefined or unknown', () => {
      expect(isCeilingHiddenForActiveLayer(undefined, { controls: true })).toBe(false)
      expect(isCeilingHiddenForActiveLayer('unknown', { controls: true })).toBe(false)
    })

    it('hides ceiling for unplaced sensors when viewing a hidden layer (no placement lookup)', () => {
      // Regression: previously an unplaced sensor defaulted to "controls",
      // so hiding a non-default layer never hid anything and the default
      // layer controlled everything. Active-layer logic fixes this.
      expect(isCeilingHiddenForActiveLayer('sensors', { controls: false, sensors: true })).toBe(true)
      expect(isCeilingHiddenForActiveLayer('controls', { controls: false, sensors: true })).toBe(false)
    })
  })
})
