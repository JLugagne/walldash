import { Home } from 'lucide-react'
import { createPortal } from 'react-dom'
import type { Layer, Level } from '../types'
import { LevelSelector } from './LevelSelector'
import { LayerSelector } from './LayerSelector'
import { useTopBarSlot } from './TopBarSlot'

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

/**
 * ViewTopSelectors renders the floor selector and the optional layer selector. The floors live in
 * the app top bar (via the top-bar slot) so they cost no vertical canvas space; the layers stay a
 * floating top-center canvas overlay. When rendered outside the TopBarSlotProvider (for example in
 * unit tests) both fall back to the stacked top-center overlay.
 */
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
  const slot = useTopBarSlot()

  const levelSelector = (
    <LevelSelector
      levels={levels}
      activeLevelId={overviewActive ? null : activeLevelId}
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
  )

  const layerSelector =
    showLayers && activeLayer && onSelectLayer ? (
      <LayerSelector layers={layers} activeLayer={activeLayer} onSelectLayer={onSelectLayer} />
    ) : null

  if (slot) {
    return (
      <>
        {createPortal(
          <div className="mx-auto flex h-14 min-w-0 max-w-full items-center gap-2 overflow-x-auto [scrollbar-width:none]">
            {levelSelector}
          </div>,
          slot,
        )}
        {layerSelector && (
          <div className="absolute top-6 left-1/2 z-20 flex -translate-x-1/2 items-center pointer-events-none">
            {layerSelector}
          </div>
        )}
      </>
    )
  }

  return (
    <div className="absolute top-6 left-1/2 z-20 flex -translate-x-1/2 flex-col items-center gap-2 pointer-events-none">
      {levelSelector}
      {layerSelector}
    </div>
  )
}
