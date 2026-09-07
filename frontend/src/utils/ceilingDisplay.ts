import type { Zone, Device } from '../types'
import { inferLayerFromDevice } from './layers'

export const CEILING_NEUTRAL_COLOR = '#f1f5f9'
export const CEILING_COLD_COLOR = '#38bdf8'
export const CEILING_HOT_COLOR = '#ef4444'

export interface CeilingAlertState {
  color: string
  isAlert: boolean
  alertType?: 'cold' | 'hot'
}

/**
 * Parses a hex color string into [r, g, b] (0-255).
 */
function hexToRgb(hex: string): [number, number, number] {
  const clean = hex.replace('#', '')
  const intVal = parseInt(clean, 16)
  return [(intVal >> 16) & 255, (intVal >> 8) & 255, intVal & 255]
}

/**
 * Formats RGB components to standard #rrggbb string.
 */
function rgbToHex(r: number, g: number, b: number): string {
  const toHex = (n: number) => Math.round(n).toString(16).padStart(2, '0')
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/**
 * Linearly interpolates between two hex colors by factor t (0 to 1).
 */
export function interpolateHexColor(fromHex: string, toHex: string, t: number): string {
  const clampedT = Math.min(1, Math.max(0, t))
  const [r1, g1, b1] = hexToRgb(fromHex)
  const [r2, g2, b2] = hexToRgb(toHex)

  return rgbToHex(
    r1 + (r2 - r1) * clampedT,
    g1 + (g2 - g1) * clampedT,
    b1 + (b2 - b1) * clampedT
  )
}

/**
 * Checks if a zone has any sensor (temperature or humidity) configured.
 */
export function hasConfiguredSensors(zone: Zone): boolean {
  return Boolean(
    (zone.temp_sensor && zone.temp_sensor.trim() !== '') ||
    (zone.humidity_sensor && zone.humidity_sensor.trim() !== '')
  )
}

/**
 * Extracts a numeric temperature value from a device or climate entity.
 */
export function getZoneTemperature(zone: Zone, deviceMap: Record<string, Device>): number | null {
  if (!zone.temp_sensor) return null

  const device = deviceMap[zone.temp_sensor]
  if (!device) return null

  if (device.domain === 'climate') {
    const cur = device.attributes?.current_temperature
    if (typeof cur === 'number' && !isNaN(cur)) return cur
    const target = device.attributes?.temperature
    if (typeof target === 'number' && !isNaN(target)) return target
  }

  const parsed = parseFloat(device.state)
  return isNaN(parsed) ? null : parsed
}

/**
 * Extracts a numeric humidity value from a device.
 */
export function getZoneHumidity(zone: Zone, deviceMap: Record<string, Device>): number | null {
  if (!zone.humidity_sensor) return null

  const device = deviceMap[zone.humidity_sensor]
  if (!device) return null

  const parsed = parseFloat(device.state)
  return isNaN(parsed) ? null : parsed
}

/**
 * Formats zone sensor metrics (e.g. "21.4°C · 48%") based on current HA state.
 * Returns null if no sensors are configured for the zone.
 */
export function formatZoneMetrics(zone: Zone, deviceMap: Record<string, Device>): string | null {
  if (!hasConfiguredSensors(zone)) return null

  const parts: string[] = []

  if (zone.temp_sensor) {
    const dev = deviceMap[zone.temp_sensor]
    const temp = getZoneTemperature(zone, deviceMap)
    if (temp !== null) {
      const unit = dev?.attributes?.unit_of_measurement || '°C'
      parts.push(`${temp}${unit}`)
    }
  }

  if (zone.humidity_sensor) {
    const dev = deviceMap[zone.humidity_sensor]
    const hum = getZoneHumidity(zone, deviceMap)
    if (hum !== null) {
      const unit = dev?.attributes?.unit_of_measurement || '%'
      parts.push(`${hum}${unit}`)
    }
  }

  if (parts.length > 0) {
    return parts.join(' · ')
  }

  return '--'
}

/**
 * Computes the text to be displayed on the 3D ceiling for a zone.
 * - If sensors are configured: displays formatted sensor metrics (e.g. "21.4°C · 48%").
 * - If no sensors are configured: displays the uppercase zone name (e.g. "SALON").
 */
export function getZoneCeilingText(zone: Zone, deviceMap: Record<string, Device>): string {
  if (hasConfiguredSensors(zone)) {
    return formatZoneMetrics(zone, deviceMap) || '--'
  }
  return (zone.name || '').toUpperCase()
}

/**
 * Calculates progressive chromatic interpolation and alert state.
 * - Normal range: neutral `#f1f5f9` with isAlert = false.
 * - Under temp_min: progressive interpolation towards neon cyan `#38bdf8` (delta / 3°C) with isAlert = true.
 * - Over temp_max: progressive interpolation towards neon red `#ef4444` (delta / 3°C) with isAlert = true.
 */
export function calculateCeilingColor(
  temperature: number | null,
  tempMin?: number | null,
  tempMax?: number | null
): CeilingAlertState {
  if (temperature === null || isNaN(temperature)) {
    return { color: CEILING_NEUTRAL_COLOR, isAlert: false }
  }

  const DELTA_RANGE = 3.0 // Degrees for full saturation

  if (tempMin != null && !isNaN(tempMin) && temperature < tempMin) {
    const delta = tempMin - temperature
    const t = Math.min(1, Math.max(0, delta / DELTA_RANGE))
    const color = interpolateHexColor(CEILING_NEUTRAL_COLOR, CEILING_COLD_COLOR, t)
    return {
      color,
      isAlert: true,
      alertType: 'cold',
    }
  }

  if (tempMax != null && !isNaN(tempMax) && temperature > tempMax) {
    const delta = temperature - tempMax
    const t = Math.min(1, Math.max(0, delta / DELTA_RANGE))
    const color = interpolateHexColor(CEILING_NEUTRAL_COLOR, CEILING_HOT_COLOR, t)
    return {
      color,
      isAlert: true,
      alertType: 'hot',
    }
  }

  return { color: CEILING_NEUTRAL_COLOR, isAlert: false }
}

export interface ZoneHUDMetrics {
  zoneName: string
  hasSensors: boolean
  tempValue: string | null
  humidityValue: string | null
  humidityPct: number | null
  humidityColor: string
  statusText: string
  accentColor: string
  isAlert: boolean
  alertType?: 'cold' | 'hot'
}

export const HUD_COLOR_COMFORT = '#2dd4bf'
export const HUD_COLOR_COLD = '#38bdf8'
export const HUD_COLOR_HOT = '#f87171'
export const HUD_COLOR_MUTED = '#94a3b8'

export const GAUGE_COLOR_DRY = '#38bdf8'     // Cyan neon for dry / below min
export const GAUGE_COLOR_COMFORT = '#2dd4bf' // Teal / comfort
export const GAUGE_COLOR_HUMID = '#ef4444'   // Coral/red for humid / above max

/**
 * Calculates progressive chromatic interpolation for the humidity gauge arc.
 * Transitions smoothly (interpolated) rather than abruptly:
 * - Around and below humMin: progressively transitions from comfort (#2dd4bf) to dry cyan (#38bdf8).
 *   Full cyan saturation is reached 10% below min (or at min if margin is small).
 * - Around and above humMax: progressively transitions from comfort (#2dd4bf) to humid red/coral (#ef4444).
 *   Full saturation is reached 10% above max.
 * - In comfort zone: remains comfortable teal.
 * If thresholds are not defined, defaults to smooth transitions around 30% and 60%.
 */
export function calculateHumidityGaugeColor(
  humidityPct: number | null,
  humMin?: number | null,
  humMax?: number | null
): string {
  if (humidityPct === null || isNaN(humidityPct)) {
    return GAUGE_COLOR_COMFORT
  }

  const min = humMin != null && !isNaN(humMin) ? humMin : 30
  const max = humMax != null && !isNaN(humMax) ? humMax : 60
  const TRANSITION_MARGIN = 10 // 10% progressive gradient window

  if (humidityPct < min) {
    const delta = min - humidityPct
    const t = Math.min(1, Math.max(0, delta / TRANSITION_MARGIN))
    return interpolateHexColor(GAUGE_COLOR_COMFORT, GAUGE_COLOR_DRY, t)
  }

  if (humidityPct > max) {
    const delta = humidityPct - max
    const t = Math.min(1, Math.max(0, delta / TRANSITION_MARGIN))
    return interpolateHexColor(GAUGE_COLOR_COMFORT, GAUGE_COLOR_HUMID, t)
  }

  return GAUGE_COLOR_COMFORT
}

/**
 * Computes structured HUD metrics for circular 3D ceiling display.
 */
export function getZoneHUDMetrics(
  zone: Zone,
  deviceMap: Record<string, Device>
): ZoneHUDMetrics {
  const zoneName = (zone.name || '').toUpperCase()
  const hasSensors = hasConfiguredSensors(zone)

  if (!hasSensors) {
    return {
      zoneName,
      hasSensors: false,
      tempValue: null,
      humidityValue: null,
      humidityPct: null,
      humidityColor: GAUGE_COLOR_COMFORT,
      statusText: '',
      accentColor: HUD_COLOR_MUTED,
      isAlert: false,
    }
  }

  const temp = getZoneTemperature(zone, deviceMap)
  const hum = getZoneHumidity(zone, deviceMap)

  let statusText = 'COMFORT'
  let accentColor = HUD_COLOR_COMFORT
  let isAlert = false
  let alertType: 'cold' | 'hot' | undefined

  if (temp !== null) {
    if (zone.temp_min != null && !isNaN(zone.temp_min) && temp < zone.temp_min) {
      statusText = 'TOO COOL'
      accentColor = HUD_COLOR_COLD
      isAlert = true
      alertType = 'cold'
    } else if (zone.temp_max != null && !isNaN(zone.temp_max) && temp > zone.temp_max) {
      statusText = 'TOO WARM'
      accentColor = HUD_COLOR_HOT
      isAlert = true
      alertType = 'hot'
    }
  }

  const tempValue = temp !== null ? `${Math.round(temp * 10) / 10}` : null
  const humidityValue = hum !== null ? `${Math.round(hum)}% RH` : null
  const humidityPct = hum !== null ? Math.round(hum) : null
  const humidityColor = calculateHumidityGaugeColor(humidityPct, zone.humidity_min, zone.humidity_max)

  return {
    zoneName,
    hasSensors: true,
    tempValue,
    humidityValue,
    humidityPct,
    humidityColor,
    statusText,
    accentColor,
    isAlert,
    alertType,
  }
}

/**
 * Computes the optimal diameter for the circular HUD disc based on room dimensions.
 * Bounded between 1.1m and 1.8m so it never overflows room partitions.
 */
export function calculateDiscDiameter(zone: Zone): number {
  const DEFAULT_MAX_DIAMETER = 3.0
  const MIN_DIAMETER = 1.1
  const SCALE = 0.02 // 2D plan units to world meters

  if (!zone.points || zone.points.length === 0) {
    return DEFAULT_MAX_DIAMETER
  }

  let minX = Infinity
  let maxX = -Infinity
  let minY = Infinity
  let maxY = -Infinity

  for (const pt of zone.points) {
    if (pt.x < minX) minX = pt.x
    if (pt.x > maxX) maxX = pt.x
    if (pt.y < minY) minY = pt.y
    if (pt.y > maxY) maxY = pt.y
  }

  const widthWorld = (maxX - minX) * SCALE
  const depthWorld = (maxY - minY) * SCALE
  const minRoomDim = Math.min(widthWorld, depthWorld)

  if (minRoomDim <= 0 || isNaN(minRoomDim)) {
    return DEFAULT_MAX_DIAMETER
  }

  return Math.min(DEFAULT_MAX_DIAMETER, Math.max(MIN_DIAMETER, minRoomDim * 1.0))
}

/**
 * Determines which zones should have their ceiling HUD gauges hidden.
 * A zone is hidden when ALL of its configured sensors belong to layers
 * where hide_gauges is true.
 * Zones without any configured sensors are never in the result (they
 * are filtered out at the render site via hasConfiguredSensors).
 */
export function computeZoneGaugesHidden(
  zones: Zone[],
  placementLayerMap: Record<string, string>,
  layerHideGaugesMap: Record<string, boolean>,
  deviceMap: Record<string, Device>
): Record<string, boolean> {
  const hidden: Record<string, boolean> = {}
  for (const zone of zones) {
    if (!zone.temp_sensor && !zone.humidity_sensor) continue
    let hiddenForZone = zone.temp_sensor ? isSensorOnHiddenLayer(zone.temp_sensor, placementLayerMap, layerHideGaugesMap, deviceMap) : false
    if (!hiddenForZone && zone.humidity_sensor) {
      hiddenForZone = isSensorOnHiddenLayer(zone.humidity_sensor, placementLayerMap, layerHideGaugesMap, deviceMap)
    }
    hidden[zone.id] = hiddenForZone
  }
  return hidden
}

function isSensorOnHiddenLayer(
  entityId: string,
  placementLayerMap: Record<string, string>,
  layerHideGaugesMap: Record<string, boolean>,
  deviceMap: Record<string, Device>
): boolean {
  const layer = placementLayerMap[entityId] || inferLayerFromDevice(entityId, deviceMap)
  return layer !== null && layerHideGaugesMap[layer] === true
}

