import React, { useEffect, useRef, useState } from 'react'
import {
  Check,
  ChevronDown,
  Edit2,
  Image,
  LayoutGrid,
  Plus,
  SlidersHorizontal,
  Trash2,
  X,
} from 'lucide-react'
import type { OverviewDashboard } from '../../types'

export interface OverviewHeaderProps {
  overviews: OverviewDashboard[]
  activeOverviewId: string | null
  isAdmin: boolean
  isEditMode: boolean
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
  onToggleBackground: () => void
  onToggleAdmin: () => void
}

type HeaderForm = 'create' | 'rename'

/**
 * OverviewHeader renders the row-2 `● {name} · LIVE` menu button and its dropdown. It owns the
 * create/rename popover with its local state and never calls the API itself: every mutation goes
 * through its callback props. It deliberately renders nothing in the app top bar so the overview
 * controls never occupy the first chrome row.
 */
export const OverviewHeader: React.FC<OverviewHeaderProps> = ({
  overviews,
  activeOverviewId,
  isAdmin,
  isEditMode,
  onSelectOverview,
  onCreateOverview,
  onRenameOverview,
  onDeleteOverview,
  onToggleEditMode,
  onAddWidget,
  onToggleBackground,
  onToggleAdmin,
}) => {
  const [menuOpen, setMenuOpen] = useState(false)
  const [openForm, setOpenForm] = useState<HeaderForm | null>(null)
  const [formValue, setFormValue] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const rootRef = useRef<HTMLDivElement>(null)

  const activeOverview = overviews.find((o) => o.id === activeOverviewId) || null

  useEffect(() => {
    if (!menuOpen && !openForm) return
    const dismiss = () => {
      setMenuOpen(false)
      setOpenForm(null)
      setFormValue('')
      setFormError(null)
    }
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) dismiss()
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') dismiss()
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [menuOpen, openForm])

  const closeForm = () => {
    setOpenForm(null)
    setFormValue('')
    setFormError(null)
  }

  const closeAll = () => {
    setMenuOpen(false)
    closeForm()
  }

  const startCreate = () => {
    setMenuOpen(false)
    setFormValue('')
    setFormError(null)
    setOpenForm('create')
  }

  const startRename = () => {
    if (!activeOverview) return
    setMenuOpen(false)
    setFormValue(activeOverview.name)
    setFormError(null)
    setOpenForm('rename')
  }

  const handleSelectOverview = (id: string) => {
    onSelectOverview(id)
    setMenuOpen(false)
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

  const itemClass =
    'h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-slate-800 text-left cursor-pointer transition-colors'
  const activeItemClass =
    'h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-white bg-slate-800/80 text-left cursor-pointer transition-colors'

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        onClick={() => {
          if (menuOpen) closeForm()
          setMenuOpen((v) => !v)
        }}
        title="Overview menu"
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider text-slate-400 hover:text-white transition-colors cursor-pointer"
      >
        <span className="h-1.5 w-1.5 rounded-full bg-[#6d76e8]" />
        <span className="text-white">{activeOverview?.name ?? 'Overview'}</span>
        <span>· Live</span>
        <ChevronDown className={`w-3 h-3 transition-transform ${menuOpen ? 'rotate-180' : ''}`} />
      </button>

      {menuOpen && (
        <div
          role="menu"
          className="absolute right-0 top-full mt-2 z-40 w-56 bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-2xl shadow-xl p-1.5 flex flex-col"
        >
          <div className="max-h-56 overflow-y-auto flex flex-col">
            {overviews.map((ov) => {
              const active = ov.id === activeOverviewId
              return (
                <button
                  key={ov.id}
                  type="button"
                  role="menuitemradio"
                  aria-checked={active}
                  onClick={() => handleSelectOverview(ov.id)}
                  className={active ? activeItemClass : itemClass}
                >
                  <Check className={`w-4 h-4 shrink-0 ${active ? 'text-[#6d76e8]' : 'opacity-0'}`} />
                  <span className="truncate">{ov.name}</span>
                </button>
              )
            })}
          </div>

          <div className="my-1 h-px bg-slate-800/80" />

          <button type="button" role="menuitem" onClick={startCreate} className={itemClass}>
            <Plus className="w-4 h-4" />
            <span>New overview</span>
          </button>

          {isAdmin && activeOverview && (
            <>
              <button type="button" role="menuitem" onClick={startRename} className={itemClass}>
                <Edit2 className="w-4 h-4" />
                <span>Rename</span>
              </button>
              <button
                type="button"
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false)
                  onDeleteOverview(activeOverview.id)
                }}
                className="h-10 flex items-center gap-2 px-3 rounded-lg text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 text-left cursor-pointer transition-colors"
              >
                <Trash2 className="w-4 h-4" />
                <span>Delete</span>
              </button>
            </>
          )}

          <div className="my-1 h-px bg-slate-800/80" />

          {isAdmin && activeOverview && (
            <>
              <button
                type="button"
                role="menuitem"
                aria-pressed={isEditMode}
                onClick={() => {
                  closeAll()
                  onToggleEditMode()
                }}
                className={isEditMode ? activeItemClass : itemClass}
              >
                <LayoutGrid className="w-4 h-4" />
                <span>Edit layout</span>
                {isEditMode && <Check className="w-4 h-4 ml-auto text-[#6d76e8]" />}
              </button>
              <button
                type="button"
                role="menuitem"
                onClick={() => {
                  closeAll()
                  onAddWidget()
                }}
                className={itemClass}
              >
                <Plus className="w-4 h-4" />
                <span>Add widget</span>
              </button>
            </>
          )}

          <button
            type="button"
            role="menuitem"
            onClick={() => {
              closeAll()
              onToggleBackground()
            }}
            className={itemClass}
          >
            <Image className="w-4 h-4" />
            <span>Change background</span>
          </button>

          <button
            type="button"
            role="menuitem"
            aria-pressed={isAdmin}
            onClick={() => {
              closeAll()
              onToggleAdmin()
            }}
            className={isAdmin ? activeItemClass : itemClass}
          >
            <SlidersHorizontal className="w-4 h-4" />
            <span>Admin mode</span>
            {isAdmin && <Check className="w-4 h-4 ml-auto text-[#6d76e8]" />}
          </button>
        </div>
      )}

      {isAdmin && openForm && (
        <form
          onSubmit={handleSubmitForm}
          className="absolute right-0 top-full mt-2 z-40 w-80 bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-2xl shadow-xl p-3 flex flex-col gap-2"
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
    </div>
  )
}
