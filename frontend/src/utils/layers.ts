import type { DevicePlacement } from '../types'

export const DEFAULT_LAYERS: string[] = ['controls', 'sensors']

export function formatLayerLabel(layer: string): string {
  if (!layer) return ''
  return layer.charAt(0).toUpperCase() + layer.slice(1)
}

export function resolveActiveLayer(
  currentActiveLayer: string | null | undefined,
  availableLayers?: string[]
): string {
  const layers = availableLayers && availableLayers.length > 0 ? availableLayers : DEFAULT_LAYERS
  if (currentActiveLayer && layers.includes(currentActiveLayer)) {
    return currentActiveLayer
  }
  return layers[0] || 'controls'
}

export function filterPlacementsByLayer(
  placements: DevicePlacement[],
  activeLayer: string
): DevicePlacement[] {
  return placements.filter((p) => (p.layer || 'controls') === activeLayer)
}
