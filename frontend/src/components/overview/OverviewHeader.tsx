import React, { useEffect, useRef, useState } from 'react'
import { Check, Edit2, LayoutGrid, MoreHorizontal, Plus, SlidersHorizontal, Trash2, X } from 'lucide-react'
import type { OverviewDashboard } from '../../types'

export interface OverviewHeaderProps {
  overviews: OverviewDashboard[]
  activeOverviewId: string | null
  isAdmin: boolean
  isEditMode: boolean
  viewModeMenu?: React.ReactNode
  onSelectOverview: (id: string) => void
  /**
   * Resolves to whether the server accepted the new overview: the create form only closes on
   * `true`. May throw to report a server-side rejection; the form then stays open and shows the
   * thrown message.
   */
  onCreateOverview: (name: string) => Promise<boolean>
  /**
   * Resolves to whether the server accepted the rename: the rename form only closes on `true`.
   * May throw to report a server-side rejection; the form then stays open and shows the thrown
   * message.
   */
  onRenameOverview: (name: string) => Promise<boolean>
  onDeleteOverview: (id: string) => void
  onToggleEditMode: () => void
  onAddWidget: () => void
  onToggleAdmin: () => void
}

type HeaderForm = 'create' | 'rename'

/**
 * OverviewHeader is the fixed 56 px (`h-14`) sub-header of the Overview view, styled like
 * `EditorHeader`. It owns the tab strip, the view-mode menu slot, the admin overflow menu
 * and the create/rename popover with its local state; the popover is anchored under the
 * header so the header height never changes. It never calls the API itself: every mutation
 * goes through its callback props.
 */
export const OverviewHeader: React.FC<OverviewHeaderProps> = ({
  overviews,
  activeOverviewId,
  isAdmin,
  isEditMode,
  viewModeMenu,
  onSelectOverview,
  onCreateOverview,
  onRenameOverview,
  onDeleteOverview,
  onToggleEditMode,
  onAddWidget,
  onToggleAdmin,
}) => {
  const [moreOpen, setMoreOpen] = useState(false)
  const [openForm, setOpenForm] = useState<HeaderForm | null>(null)
  const [formValue, setFormValue] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const moreRef = useRef<HTMLDivElement>(null)

  const activeOverview = overviews.find((o) => o.id === activeOverviewId) || null

  useEffect(() => {
    if (!moreOpen) return
    const onDown = (e: MouseEvent) => {
      if (moreRef.current && !moreRef.current.contains(e.target as Node)) setMoreOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [moreOpen])

  const closeForm = () => {
    setOpenForm(null)
    setFormValue('')
    setFormError(null)
  }

  const handleToggleAdmin = () => {
    setMoreOpen(false)
    closeForm()
    onToggleAdmin()
  }

  const startCreate = () => {
    setMoreOpen(false)
    setFormValue('')
    setFormError(null)
    setOpenForm('create')
  }

  const startRename = () => {
    if (!activeOverview) return
    setMoreOpen(false)
    setFormValue(activeOverview.name)
    setFormError(null)
    setOpenForm('rename')
  }

  const handleSelectOverview = (id: string) => {
    onSelectOverview(id)
    if (openForm === 'rename') closeForm()
  }

  const handleSubmitForm = async (e: React.FormEvent) => {
    e.preventDefault()
    const name = formValue.trim()
    if (!name || !openForm || submitting) return
    setSubmitting(true)
    setFormError(null)
    try {
      const ok = openForm === 'create' ? await onCreateOverview(name) : await onRenameOverview(name)
      if (ok) closeForm()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <header className="relative z-30 h-14 shrink-0 bg-slate-900/90 border-b border-slate-800 flex items-center px-3 gap-3 select-none">
      {viewModeMenu && (
        <>
          <div className="shrink-0 -ml-1">{viewModeMenu}</div>
          <div className="h-6 w-px bg-slate-800 shrink-0" />
        </>
      )}

      <div className="flex items-center gap-1.5 overflow-x-auto min-w-0 flex-1 [scrollbar-width:none]">
        {overviews.map((ov) => {
          const active = ov.id === activeOverviewId
          return (
            <button
              key={ov.id}
              type="button"
              onClick={() => handleSelectOverview(ov.id)}
              aria-current={active ? 'page' : undefined}
              className={`h-10 px-4 rounded-xl text-xs font-bold whitespace-nowrap shrink-0 transition-all active:scale-95 cursor-pointer ${
                active
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30'
                  : 'bg-slate-800/60 text-slate-300 hover:bg-slate-800 hover:text-white'
              }`}
            >
              <span className="block truncate max-w-[12rem]">{ov.name}</span>
            </button>
          )
        })}
      </div>

      <div className="flex items-center gap-2 shrink-0">
        {isAdmin && activeOverview && (
          <button
            type="button"
            onClick={onAddWidget}
            title="Add widget"
            className="h-10 px-3.5 rounded-xl text-xs font-bold flex items-center gap-1.5 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white shadow-lg shadow-indigo-600/30 transition-all cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            <span>Add</span>
          </button>
        )}

        {isAdmin && activeOverview && (
          <button
            type="button"
            onClick={onToggleEditMode}
            aria-pressed={isEditMode}
            title={isEditMode ? 'Finish layout' : 'Edit layout'}
            className={`h-10 px-3 rounded-xl text-xs font-semibold flex items-center gap-1.5 border transition-all cursor-pointer ${
              isEditMode
                ? 'bg-cyan-500/20 text-cyan-300 border-cyan-500/40 shadow-sm shadow-cyan-500/20'
                : 'bg-slate-800/60 text-slate-400 border-slate-700 hover:text-slate-200'
            }`}
          >
            <LayoutGrid className="w-4 h-4" />
            <span>Layout</span>
          </button>
        )}

        {isAdmin && (
          <div ref={moreRef} className="relative">
            <button
              type="button"
              onClick={() => {
                closeForm()
                setMoreOpen((v) => !v)
              }}
              title="More actions"
              aria-haspopup="menu"
              aria-expanded={moreOpen}
              className={`h-10 w-10 rounded-xl flex items-center justify-center transition-all cursor-pointer ${
                moreOpen ? 'bg-slate-700 text-white' : 'text-slate-300 hover:text-white hover:bg-slate-800'
              }`}
            >
              <MoreHorizontal className="w-5 h-5" />
            </button>
            {moreOpen && (
              <div
                role="menu"
                className="absolute right-0 top-full mt-2 z-40 w-56 bg-slate-900 border border-slate-700/80 rounded-xl shadow-2xl shadow-black/60 p-1.5 flex flex-col"
              >
                <button
                  type="button"
                  role="menuitem"
                  onClick={startCreate}
                  className="h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-slate-800 text-left cursor-pointer"
                >
                  <Plus className="w-4 h-4" />
                  <span>New Overview</span>
                </button>
                {activeOverview && (
                  <button
                    type="button"
                    role="menuitem"
                    onClick={startRename}
                    className="h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-slate-800 text-left cursor-pointer"
                  >
                    <Edit2 className="w-4 h-4" />
                    <span>Rename</span>
                  </button>
                )}
                {activeOverview && (
                  <button
                    type="button"
                    role="menuitem"
                    onClick={() => {
                      setMoreOpen(false)
                      onDeleteOverview(activeOverview.id)
                    }}
                    className="h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 text-left cursor-pointer"
                  >
                    <Trash2 className="w-4 h-4" />
                    <span>Delete</span>
                  </button>
                )}
              </div>
            )}
          </div>
        )}

        <button
          type="button"
          onClick={handleToggleAdmin}
          aria-pressed={isAdmin}
          title={isAdmin ? 'Exit admin mode' : 'Enter admin mode'}
          className={`h-10 px-3 rounded-xl text-xs font-semibold flex items-center gap-1.5 border transition-all cursor-pointer ${
            isAdmin
              ? 'bg-amber-500/20 text-amber-300 border-amber-500/40 shadow-sm shadow-amber-500/20'
              : 'bg-slate-800/60 text-slate-400 border-slate-700 hover:text-slate-200'
          }`}
        >
          <SlidersHorizontal className="w-4 h-4" />
          <span>{isAdmin ? 'Admin' : 'User'}</span>
        </button>
      </div>

      {isAdmin && openForm && (
        <form
          onSubmit={handleSubmitForm}
          className="absolute right-3 top-full mt-2 z-40 w-80 bg-slate-900 border border-slate-700/80 rounded-xl shadow-2xl shadow-black/60 p-3 flex flex-col gap-2"
        >
          <label htmlFor="overview-name-input" className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">
            {openForm === 'create' ? 'New Overview' : `Rename "${activeOverview?.name ?? ''}"`}
          </label>
          <div className="flex items-center gap-2">
            <input
              id="overview-name-input"
              type="text"
              value={formValue}
              onChange={(e) => {
                setFormValue(e.target.value)
                setFormError(null)
              }}
              onKeyDown={(e) => {
                if (e.key === 'Escape') closeForm()
              }}
              placeholder="Overview name..."
              autoFocus
              disabled={submitting}
              className="h-10 flex-1 min-w-0 bg-slate-800 border border-indigo-500 rounded-lg px-3 text-sm text-white placeholder-slate-500 focus:outline-none disabled:opacity-60"
            />
            <button
              type="submit"
              disabled={submitting || !formValue.trim()}
              title="Confirm"
              className="h-10 w-10 shrink-0 rounded-lg flex items-center justify-center bg-indigo-600 hover:bg-indigo-500 text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all cursor-pointer"
            >
              <Check className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={closeForm}
              title="Cancel"
              className="h-10 w-10 shrink-0 rounded-lg flex items-center justify-center bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white transition-all cursor-pointer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
          {formError && <p className="text-[11px] font-medium text-red-400">{formError}</p>}
        </form>
      )}
    </header>
  )
}
