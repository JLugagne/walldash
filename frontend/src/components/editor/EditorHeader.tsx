import { useEffect, useRef, useState, type ReactNode } from 'react'
import {
  Check,
  Eraser,
  Home,
  Loader2,
  MoreHorizontal,
  PanelRight,
  Redo2,
  RotateCcw,
  Save,
  Settings2,
  Trees,
  Undo2,
  Upload,
} from 'lucide-react'
import type { Level } from '../../types'

export type SaveState = 'idle' | 'saving' | 'saved' | 'error'

interface EditorHeaderProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (id: string) => void
  levelsManager: ReactNode
  levelsOpen: boolean
  onToggleLevels: (open: boolean) => void
  viewModeMenu?: ReactNode
  canUndo: boolean
  canRedo: boolean
  onUndo: () => void
  onRedo: () => void
  onImport: () => void
  onReset: () => void
  onClear: () => void
  onSave: () => void
  isDirty: boolean
  saveState: SaveState
  panelOpen: boolean
  onTogglePanel: () => void
  disabled: boolean
}

export function EditorHeader({
  levels,
  activeLevelId,
  onSelectLevel,
  levelsManager,
  levelsOpen,
  onToggleLevels,
  viewModeMenu,
  canUndo,
  canRedo,
  onUndo,
  onRedo,
  onImport,
  onReset,
  onClear,
  onSave,
  isDirty,
  saveState,
  panelOpen,
  onTogglePanel,
  disabled,
}: EditorHeaderProps) {
  const [moreOpen, setMoreOpen] = useState(false)
  const moreRef = useRef<HTMLDivElement>(null)
  const levelsRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!moreOpen && !levelsOpen) return
    const onDown = (e: MouseEvent) => {
      const target = e.target as Node
      if (moreOpen && moreRef.current && !moreRef.current.contains(target)) setMoreOpen(false)
      if (levelsOpen && levelsRef.current && !levelsRef.current.contains(target)) onToggleLevels(false)
    }
    document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [moreOpen, levelsOpen, onToggleLevels])

  const sortedLevels = [...levels].sort((a, b) => a.order - b.order)

  const saveLabel =
    saveState === 'saving'
      ? 'Saving…'
      : saveState === 'saved' && !isDirty
        ? 'Saved'
        : isDirty
          ? 'Save'
          : 'Up to date'

  return (
    <header className="relative z-30 h-14 shrink-0 bg-slate-900/90 border-b border-slate-800 flex items-center px-3 gap-3 select-none">
      {viewModeMenu && <div className="shrink-0 -ml-1">{viewModeMenu}</div>}

      <div className="h-6 w-px bg-slate-800 shrink-0" />

      <div ref={levelsRef} className="relative flex items-center gap-1 min-w-0 flex-1">
        <div className="flex items-center gap-1 overflow-x-auto min-w-0 py-1 [scrollbar-width:none]">
          {sortedLevels.length === 0 && (
            <span className="text-xs text-slate-500 px-2 whitespace-nowrap">No levels</span>
          )}
          {sortedLevels.map((lvl) => {
            const active = lvl.id === activeLevelId
            const Icon = lvl.is_outdoor ? Trees : Home
            return (
              <button
                key={lvl.id}
                type="button"
                onClick={() => onSelectLevel(lvl.id)}
                className={`h-8 px-3 rounded-lg text-xs font-semibold flex items-center gap-1.5 whitespace-nowrap transition-all cursor-pointer ${
                  active
                    ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/25'
                    : 'text-slate-300 hover:text-white hover:bg-slate-800'
                }`}
              >
                <Icon className={`w-3.5 h-3.5 ${active ? 'text-white' : lvl.is_outdoor ? 'text-emerald-400' : 'text-indigo-400'}`} />
                <span className="truncate max-w-[10rem]">{lvl.name}</span>
              </button>
            )
          })}
        </div>
        <button
          type="button"
          onClick={() => onToggleLevels(!levelsOpen)}
          title="Manage levels"
          aria-expanded={levelsOpen}
          className={`h-8 w-8 shrink-0 rounded-lg flex items-center justify-center transition-all cursor-pointer ${
            levelsOpen ? 'bg-slate-700 text-white' : 'text-slate-400 hover:text-white hover:bg-slate-800'
          }`}
        >
          <Settings2 className="w-4 h-4" />
        </button>

        {levelsOpen && (
          <div className="absolute top-full left-0 mt-2 z-40 w-[22rem] max-h-[70vh] bg-slate-900 border border-slate-700/80 rounded-2xl shadow-2xl shadow-black/60 overflow-hidden flex flex-col">
            {levelsManager}
          </div>
        )}
      </div>

      <div className="flex items-center gap-1 shrink-0">
        <IconButton onClick={onUndo} disabled={disabled || !canUndo} title="Undo (Ctrl+Z)">
          <Undo2 className="w-4 h-4" />
        </IconButton>
        <IconButton onClick={onRedo} disabled={disabled || !canRedo} title="Redo (Ctrl+Y)">
          <Redo2 className="w-4 h-4" />
        </IconButton>

        <div className="h-6 w-px bg-slate-800 mx-1" />

        <button
          type="button"
          onClick={onImport}
          disabled={disabled}
          title="Import a plan from AI-generated JSON"
          className="h-8 px-2.5 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1.5 transition-all cursor-pointer"
        >
          <Upload className="w-3.5 h-3.5" />
          <span className="hidden lg:inline">Import</span>
        </button>

        <div ref={moreRef} className="relative">
          <IconButton onClick={() => setMoreOpen((v) => !v)} disabled={disabled} title="More actions">
            <MoreHorizontal className="w-4 h-4" />
          </IconButton>
          {moreOpen && (
            <div className="absolute right-0 top-full mt-2 z-40 w-56 bg-slate-900 border border-slate-700/80 rounded-xl shadow-2xl shadow-black/60 p-1.5 flex flex-col">
              <button
                type="button"
                onClick={() => {
                  setMoreOpen(false)
                  onReset()
                }}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-slate-800 text-left cursor-pointer"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>Revert to saved version</span>
              </button>
              <button
                type="button"
                onClick={() => {
                  setMoreOpen(false)
                  onClear()
                }}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 text-left cursor-pointer"
              >
                <Eraser className="w-3.5 h-3.5" />
                <span>Clear walls and zones</span>
              </button>
            </div>
          )}
        </div>

        <button
          type="button"
          onClick={onSave}
          disabled={disabled || saveState === 'saving' || (!isDirty && saveState !== 'error')}
          title="Save plan (Ctrl+S)"
          className={`h-8 px-3.5 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer disabled:cursor-not-allowed ml-1 ${
            isDirty
              ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-md shadow-emerald-600/25'
              : saveState === 'error'
                ? 'bg-rose-600 hover:bg-rose-500 text-white'
                : 'bg-slate-800 text-slate-400'
          }`}
        >
          {saveState === 'saving' ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
          ) : isDirty ? (
            <Save className="w-3.5 h-3.5" />
          ) : (
            <Check className="w-3.5 h-3.5" />
          )}
          <span>{saveLabel}</span>
          {isDirty && <span className="w-1.5 h-1.5 rounded-full bg-white/90 animate-pulse" />}
        </button>

        <div className="h-6 w-px bg-slate-800 mx-1" />

        <IconButton onClick={onTogglePanel} title={panelOpen ? 'Hide panel' : 'Show panel'} active={panelOpen}>
          <PanelRight className="w-4 h-4" />
        </IconButton>
      </div>
    </header>
  )
}

function IconButton({
  children,
  onClick,
  disabled,
  title,
  active,
}: {
  children: ReactNode
  onClick: () => void
  disabled?: boolean
  title: string
  active?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={`h-8 w-8 rounded-lg flex items-center justify-center transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed ${
        active ? 'bg-slate-700 text-white' : 'text-slate-300 hover:text-white hover:bg-slate-800'
      }`}
    >
      {children}
    </button>
  )
}
