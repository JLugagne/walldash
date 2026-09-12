import React, { useRef, useState } from 'react'
import { Check, Edit2, Grid2x2, Image, LayoutGrid, Plus, Trash2, X } from 'lucide-react'
import type { Dashboard } from '../../types'

/** Bounds offered by the grid-size editor; the server still validates every change. */
export const GRID_MIN = 2
export const GRID_MAX = 24

export interface DashboardHeaderProps {
  dashboard: Dashboard | null
  isEditMode: boolean
  /**
   * Resolves to whether the server accepted the rename: the rename form only closes on `true`.
   * May throw to report a server-side rejection; the form then stays open and shows the thrown
   * message.
   */
  onRename: (name: string) => Promise<boolean>
  /**
   * Resizes this dashboard's Widget Grid. The dimensions are per dashboard. May throw to report a
   * server-side rejection (for example a shrink that would push a widget outside the grid).
   */
  onResize: (cols: number, rows: number) => Promise<boolean>
  onDelete: () => void
  onAddWidget: () => void
  onToggleEditMode: () => void
  onToggleBackground: () => void
}

/**
 * DashboardHeader is the edit-mode toolbar: it shows the dashboard name and exposes the
 * rename, delete, add-widget, edit-layout, grid-size and change-background actions. It never calls
 * the API itself: every mutation goes through its callback props.
 */
export const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  dashboard,
  isEditMode,
  onRename,
  onResize,
  onDelete,
  onAddWidget,
  onToggleEditMode,
  onToggleBackground,
}) => {
  const [renaming, setRenaming] = useState(false)
  const [formValue, setFormValue] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [gridOpen, setGridOpen] = useState(false)
  const [colsValue, setColsValue] = useState('10')
  const [rowsValue, setRowsValue] = useState('6')
  const [gridSubmitting, setGridSubmitting] = useState(false)
  const [gridError, setGridError] = useState<string | null>(null)
  const rootRef = useRef<HTMLDivElement>(null)

  const startRename = () => {
    setFormValue(dashboard?.name ?? '')
    setFormError(null)
    setRenaming(true)
  }

  const closeRename = () => {
    setRenaming(false)
    setFormValue('')
    setFormError(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const name = formValue.trim()
    if (!name || submitting) return
    setSubmitting(true)
    setFormError(null)
    try {
      const ok = await onRename(name)
      if (ok) closeRename()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  const startGridEdit = () => {
    setColsValue(String(dashboard?.cols ?? 12))
    setRowsValue(String(dashboard?.rows ?? 8))
    setGridError(null)
    setGridOpen(true)
  }

  const handleGridSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const cols = Number.parseInt(colsValue, 10)
    const rows = Number.parseInt(rowsValue, 10)
    const inRange = (n: number) => Number.isInteger(n) && n >= GRID_MIN && n <= GRID_MAX
    if (!inRange(cols) || !inRange(rows) || gridSubmitting) return
    setGridSubmitting(true)
    setGridError(null)
    try {
      const ok = await onResize(cols, rows)
      if (ok) setGridOpen(false)
    } catch (err) {
      setGridError(err instanceof Error ? err.message : String(err))
    } finally {
      setGridSubmitting(false)
    }
  }

  const buttonClass =
    'h-9 flex items-center gap-1.5 px-3 rounded-lg border border-slate-800/80 bg-slate-900/70 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer'

  return (
    <div ref={rootRef} className="relative flex items-center gap-2">
      <span className="flex h-9 items-center gap-2 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 text-xs font-semibold text-white">
        <span className="h-1.5 w-1.5 rounded-full bg-[#6d76e8]" />
        <span className="max-w-[14rem] truncate">{dashboard?.name ?? 'Dashboard'}</span>
      </span>

      <button type="button" onClick={startRename} disabled={!dashboard} className={buttonClass}>
        <Edit2 className="w-3.5 h-3.5" />
        <span>Rename</span>
      </button>

      <button
        type="button"
        onClick={onDelete}
        disabled={!dashboard}
        className="h-9 flex items-center gap-1.5 px-3 rounded-lg border border-rose-500/30 bg-rose-500/10 text-xs font-semibold text-rose-300 hover:text-rose-200 hover:bg-rose-500/20 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
      >
        <Trash2 className="w-3.5 h-3.5" />
        <span>Delete</span>
      </button>

      <button type="button" onClick={onAddWidget} disabled={!dashboard} className={buttonClass}>
        <Plus className="w-3.5 h-3.5" />
        <span>Add widget</span>
      </button>

      <button type="button" onClick={startGridEdit} disabled={!dashboard} className={buttonClass}>
        <Grid2x2 className="w-3.5 h-3.5" />
        <span>Grid</span>
      </button>

      <button
        type="button"
        onClick={onToggleEditMode}
        aria-pressed={isEditMode}
        disabled={!dashboard}
        className={
          isEditMode
            ? 'h-9 flex items-center gap-1.5 px-3 rounded-lg border border-[#6d76e8] bg-[#6d76e8] text-xs font-semibold text-white transition-colors cursor-pointer'
            : buttonClass
        }
      >
        <LayoutGrid className="w-3.5 h-3.5" />
        <span>Edit layout</span>
      </button>

      <button type="button" onClick={onToggleBackground} disabled={!dashboard} className={buttonClass}>
        <Image className="w-3.5 h-3.5" />
        <span>Change background</span>
      </button>

      {renaming && (
        <form
          onSubmit={handleSubmit}
          className="absolute left-0 top-full mt-2 z-40 w-80 bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-2xl shadow-xl p-3 flex flex-col gap-2"
        >
          <label htmlFor="dashboard-name-input" className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">
            Rename "{dashboard?.name ?? ''}"
          </label>
          <div className="flex items-center gap-2">
            <input
              id="dashboard-name-input"
              type="text"
              value={formValue}
              onChange={(e) => {
                setFormValue(e.target.value)
                setFormError(null)
              }}
              onKeyDown={(e) => {
                if (e.key === 'Escape') closeRename()
              }}
              placeholder="Dashboard name..."
              autoFocus
              disabled={submitting}
              className="h-10 flex-1 min-w-0 bg-slate-800 border border-[#6d76e8] rounded-lg px-3 text-sm text-white placeholder-slate-500 focus:outline-none disabled:opacity-60"
            />
            <button
              type="submit"
              disabled={submitting || !formValue.trim()}
              title="Confirm"
              className="h-10 w-10 shrink-0 rounded-lg flex items-center justify-center bg-[#6d76e8] hover:bg-[#7b83ea] text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all cursor-pointer"
            >
              <Check className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={closeRename}
              title="Cancel"
              className="h-10 w-10 shrink-0 rounded-lg flex items-center justify-center bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white transition-all cursor-pointer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
          {formError && <p className="text-[11px] font-medium text-red-400">{formError}</p>}
        </form>
      )}

      {gridOpen && (
        <form
          onSubmit={handleGridSubmit}
          className="absolute left-0 top-full mt-2 z-40 w-72 bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-2xl shadow-xl p-3 flex flex-col gap-2"
        >
          <span className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">
            Grid size for "{dashboard?.name ?? ''}"
          </span>
          <div className="flex items-end gap-2">
            <label htmlFor="dashboard-cols-input" className="flex-1 min-w-0 flex flex-col gap-1">
              <span className="text-[11px] text-slate-400">Columns</span>
              <input
                id="dashboard-cols-input"
                type="number"
                min={GRID_MIN}
                max={GRID_MAX}
                value={colsValue}
                onChange={(e) => {
                  setColsValue(e.target.value)
                  setGridError(null)
                }}
                disabled={gridSubmitting}
                className="h-10 w-full bg-slate-800 border border-[#6d76e8] rounded-lg px-3 text-sm text-white focus:outline-none disabled:opacity-60"
              />
            </label>
            <label htmlFor="dashboard-rows-input" className="flex-1 min-w-0 flex flex-col gap-1">
              <span className="text-[11px] text-slate-400">Rows</span>
              <input
                id="dashboard-rows-input"
                type="number"
                min={GRID_MIN}
                max={GRID_MAX}
                value={rowsValue}
                onChange={(e) => {
                  setRowsValue(e.target.value)
                  setGridError(null)
                }}
                disabled={gridSubmitting}
                className="h-10 w-full bg-slate-800 border border-[#6d76e8] rounded-lg px-3 text-sm text-white focus:outline-none disabled:opacity-60"
              />
            </label>
          </div>
          <div className="flex items-center justify-between gap-2">
            <span className="text-[10px] leading-tight text-slate-500">
              Only this dashboard. Existing widgets must still fit.
            </span>
            <div className="flex items-center gap-1.5 shrink-0">
              <button
                type="submit"
                disabled={gridSubmitting}
                className="h-9 px-3 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] text-xs font-semibold text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all cursor-pointer"
              >
                {gridSubmitting ? 'Saving…' : 'Apply'}
              </button>
              <button
                type="button"
                onClick={() => setGridOpen(false)}
                disabled={gridSubmitting}
                className="h-9 px-3 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-semibold text-slate-300 disabled:opacity-40 transition-all cursor-pointer"
              >
                Cancel
              </button>
            </div>
          </div>
          {gridError && <p className="text-[11px] font-medium text-red-400">{gridError}</p>}
        </form>
      )}
    </div>
  )
}
