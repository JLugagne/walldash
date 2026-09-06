import { Grid3x3, Layers, Magnet } from 'lucide-react'
import { TOOLS, type ToolMode } from './constants'

interface ToolRailProps {
  tool: ToolMode
  onSelectTool: (tool: ToolMode) => void
  snapGrid: boolean
  onToggleSnap: () => void
  gridSize: number
  onCycleGrid: () => void
  layersOpen: boolean
  onToggleLayers: () => void
}

export function ToolRail({ tool, onSelectTool, snapGrid, onToggleSnap, gridSize, onCycleGrid, layersOpen, onToggleLayers }: ToolRailProps) {
  return (
    <aside className="w-14 shrink-0 bg-slate-900/80 border-r border-slate-800 flex flex-col items-center py-2 gap-1 select-none">
      <button
        type="button"
        onClick={onToggleLayers}
        title={`Layers (L)`}
        aria-pressed={layersOpen}
        className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all cursor-pointer ${
          layersOpen ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30' : 'text-slate-400 hover:text-white hover:bg-slate-800'
        }`}
      >
        <Layers className="w-[18px] h-[18px]" />
      </button>

      <div className="w-8 h-px bg-slate-800 my-0.5" />

      {TOOLS.map((def) => {
        const Icon = def.icon
        const active = tool === def.key
        return (
          <button
            key={def.key}
            type="button"
            onClick={() => onSelectTool(def.key)}
            title={`${def.label} (${def.shortcut})`}
            aria-label={def.label}
            aria-pressed={active}
            className={`relative w-10 h-10 rounded-xl flex items-center justify-center transition-all cursor-pointer ${
              active
                ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30'
                : 'text-slate-400 hover:text-white hover:bg-slate-800'
            }`}
          >
            <Icon className="w-[18px] h-[18px]" />
            <span
              className={`absolute bottom-0.5 right-1 text-[8px] font-mono leading-none ${
                active ? 'text-indigo-200' : 'text-slate-600'
              }`}
            >
              {def.shortcut}
            </span>
          </button>
        )
      })}

      <div className="w-8 h-px bg-slate-800 my-1" />

      <button
        type="button"
        onClick={onToggleSnap}
        title={`Snap to grid: ${snapGrid ? 'enabled' : 'disabled'} (G)`}
        aria-pressed={snapGrid}
        className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all cursor-pointer ${
          snapGrid ? 'text-emerald-400 bg-emerald-500/10' : 'text-slate-500 hover:text-white hover:bg-slate-800'
        }`}
      >
        <Magnet className="w-[18px] h-[18px]" />
      </button>

      <button
        type="button"
        onClick={onCycleGrid}
        title="Grid step"
        className="w-10 h-10 rounded-xl flex flex-col items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition-all cursor-pointer"
      >
        <Grid3x3 className="w-4 h-4" />
        <span className="text-[9px] font-mono leading-none mt-0.5">{gridSize}</span>
      </button>
    </aside>
  )
}
