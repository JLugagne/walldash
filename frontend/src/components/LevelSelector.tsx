import type { Level } from '../types'
import type { ReactNode } from 'react'

interface LevelSelectorProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (levelId: string) => void
  leadingAction?: ReactNode
}

export function LevelSelector({ levels, activeLevelId, onSelectLevel, leadingAction }: LevelSelectorProps) {
  if (levels.length === 0) return null

  const sortedLevels = [...levels].sort((a, b) => a.order - b.order)

  return (
    <div
      role="group"
      aria-label="Level selection"
      className="bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-xl p-1 flex items-center gap-1 pointer-events-auto"
    >
      {leadingAction}
      {sortedLevels.map((lvl) => {
        const isActive = lvl.id === activeLevelId
        return (
          <button
            key={lvl.id}
            type="button"
            onClick={() => onSelectLevel(lvl.id)}
            aria-pressed={isActive}
            aria-label={`Level ${lvl.name}`}
            className={`px-3.5 py-1.5 rounded-lg text-xs transition-all cursor-pointer whitespace-nowrap select-none ${
              isActive
                ? 'bg-[#6d76e8] text-white font-semibold'
                : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
            }`}
          >
            {lvl.name}
          </button>
        )
      })}
    </div>
  )
}
