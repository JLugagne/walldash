import { DEFAULT_LAYERS, formatLayerLabel } from '../utils/layers'
import type { Layer } from '../types'

export interface LayerSelectorProps {
  layers?: Layer[]
  activeLayer: string
  onSelectLayer: (layer: string) => void
}

export function LayerSelector({
  layers,
  activeLayer,
  onSelectLayer,
}: LayerSelectorProps) {
  const displayLayers = layers && layers.length > 0 ? layers : DEFAULT_LAYERS

  return (
    <div
      role="group"
      aria-label="Display Layers"
      className="rounded-full bg-slate-900/70 backdrop-blur-md border border-slate-800/80 p-1 flex items-center space-x-1 pointer-events-auto"
    >
      {displayLayers.map((layer) => {
        const isActive = layer.name === activeLayer
        const label = formatLayerLabel(layer.name)

        return (
          <button
            key={layer.name}
            type="button"
            onClick={() => onSelectLayer(layer.name)}
            aria-pressed={isActive}
            aria-label={`Layer ${label}`}
            className={`px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap select-none ${
              isActive
                ? 'bg-[#6d76e8] text-white font-semibold'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
            }`}
          >
            {label}
          </button>
        )
      })}
    </div>
  )
}