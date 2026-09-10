import { Home } from 'lucide-react'
import type { Layer, Level } from '../types'
import { LevelSelector } from './LevelSelector'
import { LayerSelector } from './LayerSelector'

interface ViewTopSelectorsProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (levelId: string) => void
  overviewActive: boolean
  onToggleOverview: () => void
  showLayers?: boolean
  layers?: Layer[]
  activeLayer?: string
  onSelectLayer?: (layer: string) => void
}

export function ViewTopSelectors({
  levels,
  activeLevelId,
  onSelectLevel,
  overviewActive,
  onToggleOverview,
  showLayers = false,
  layers,
  activeLayer,
  onSelectLayer,
}: ViewTopSelectorsProps) {
  return (
    <div className="absolute top-6 left-1/2 z-20 flex -translate-x-1/2 flex-col items-center gap-2 pointer-events-none">
      <LevelSelector
        levels={levels}
        activeLevelId={activeLevelId}
        onSelectLevel={onSelectLevel}
        leadingAction={
          <button
            type="button"
            onClick={onToggleOverview}
            title={overviewActive ? 'Back to floor view' : 'House overview'}
            aria-label={overviewActive ? 'Back to floor view' : 'House overview'}
            aria-pressed={overviewActive}
            className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg transition-all active:scale-95 ${
              overviewActive
                ? 'bg-[#6d76e8] text-white'
                : 'text-slate-400 hover:bg-slate-800/60 hover:text-white'
            }`}
          >
            <Home className="h-4 w-4" />
          </button>
        }
      />
      {showLayers && activeLayer && onSelectLayer && (
        <LayerSelector layers={layers} activeLayer={activeLayer} onSelectLayer={onSelectLayer} />
      )}
    </div>
  )
}
