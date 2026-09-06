import type { DevicePlacement, Device, Layer } from '../types'

export const DEFAULT_LAYERS: Layer[] = [
  { name: 'controls', hide_gauges: false },
  { name: 'sensors', hide_gauges: false },
]

export function formatLayerLabel(layer: string): string {
  if (!layer) return ''
  return layer.charAt(0).toUpperCase() + layer.slice(1)
}

export function getDefaultLayerNames(): string[] {
  return DEFAULT_LAYERS.map((l) => l.name)
}

export function resolveActiveLayer(
  currentActiveLayer: string | null | undefined,
  availableLayers?: Layer[]
): string {
  const layers = availableLayers && availableLayers.length > 0 ? availableLayers : DEFAULT_LAYERS
  const layerNames = layers.map((l) => l.name)
  if (currentActiveLayer && layerNames.includes(currentActiveLayer)) {
    return currentActiveLayer
  }
  return layerNames[0] || 'controls'
}

export function getLayerByName(layers: Layer[], name: string): Layer | undefined {
  return layers.find((l) => l.name === name)
}

export function filterPlacementsByLayer(
  placements: DevicePlacement[],
  activeLayer: string
): DevicePlacement[] {
  return placements.filter((p) => (p.layer || 'controls') === activeLayer)
}

/**
 * Infers which display layer a Home Assistant entity belongs to.
 * Uses deviceMap for the most accurate domain, falling back to parsing
 * entity_id (format: domain.entity_name) when the device is not yet loaded.
 */
export function inferLayerFromDevice(
  entityId: string,
  deviceMap: Record<string, Device>
): string | null {
  const device = deviceMap[entityId]
  const domain = device?.domain || (entityId.includes('.') ? entityId.split('.')[0] : null)
  if (!domain) return null
  if (domain === 'sensor') return 'sensors'
  if (domain === 'climate') return 'controls'
  return 'controls'
}