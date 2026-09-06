import type { Level } from '../types'

interface LevelSelectorProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (levelId: string) => void
}

export function LevelSelector({ levels, activeLevelId, onSelectLevel }: LevelSelectorProps) {
  if (levels.length === 0) return null

  const sortedLevels = [...levels].sort((a, b) => a.order - b.order)

  return (
    <div
      role="group"
      aria-label="Level selection"
      className="bg-slate-900/80 backdrop-blur-md border border-slate-800 rounded-full p-1 shadow-xl flex items-center space-x-1 pointer-events-auto"
    >
      {sortedLevels.map((lvl) => {
        const isActive = lvl.id === activeLevelId
        return (
          <button
            key={lvl.id}
            type="button"
            onClick={() => onSelectLevel(lvl.id)}
            aria-pressed={isActive}
            aria-label={`Level ${lvl.name}`}
            className={`px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap select-none ${
              isActive
                ? 'bg-indigo-600 text-white shadow-md font-semibold'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
            }`}
          >
            {lvl.name}
          </button>
        )
      })}
    </div>
  )
}
