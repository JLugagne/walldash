import { Home, Trees } from 'lucide-react'
import type { Level } from '../types'

interface LevelSelectorProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (levelId: string) => void
}

export function LevelSelector({ levels, activeLevelId, onSelectLevel }: LevelSelectorProps) {
  if (levels.length === 0) return null

  // Sort levels by order
  const sortedLevels = [...levels].sort((a, b) => a.order - b.order)

  return (
    <nav
      aria-label="Sélecteur de niveau"
      className="bg-slate-900/90 backdrop-blur-md px-2 py-1.5 rounded-2xl border border-slate-800/80 shadow-2xl shadow-black/50 flex items-center space-x-1.5 pointer-events-auto"
    >
      <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-400 px-2 select-none hidden sm:inline">
        Niveau
      </span>

      <div className="flex items-center space-x-1 overflow-x-auto max-w-[80vw] sm:max-w-none py-0.5">
        {sortedLevels.map((lvl) => {
          const isActive = lvl.id === activeLevelId
          return (
            <button
              key={lvl.id}
              type="button"
              onClick={() => onSelectLevel(lvl.id)}
              className={`min-h-[42px] px-3.5 py-2 rounded-xl text-xs font-semibold transition-all flex items-center space-x-2 select-none active:scale-95 ${
                isActive
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30 ring-1 ring-indigo-400/50'
                  : 'text-slate-300 hover:text-white hover:bg-slate-800/70 border border-transparent hover:border-slate-700/60'
              }`}
            >
              {lvl.is_outdoor ? (
                <Trees className={`w-4 h-4 ${isActive ? 'text-white' : 'text-emerald-400'}`} />
              ) : (
                <Home className={`w-4 h-4 ${isActive ? 'text-white' : 'text-indigo-400'}`} />
              )}
              <span className="whitespace-nowrap">{lvl.name}</span>
            </button>
          )
        })}
      </div>
    </nav>
  )
}
