import { DEFAULT_LAYERS, formatLayerLabel } from '../utils/layers'

export interface LayerSelectorProps {
  layers?: string[]
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
      className="bg-slate-900/80 backdrop-blur-md border border-slate-800 rounded-full p-1 shadow-xl flex items-center space-x-1 pointer-events-auto"
    >
      {displayLayers.map((layer) => {
        const isActive = layer === activeLayer
        const label = formatLayerLabel(layer)

        return (
          <button
            key={layer}
            type="button"
            onClick={() => onSelectLayer(layer)}
            aria-pressed={isActive}
            aria-label={`Layer ${label}`}
            className={`px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap select-none ${
              isActive
                ? 'bg-indigo-600 text-white shadow-md font-semibold'
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
