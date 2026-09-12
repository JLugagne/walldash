import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Check, ChevronRight, Edit2, LayoutDashboard, Plus, RefreshCw, Trash2, X } from 'lucide-react'
import type { Dashboard } from '../../types'
import { apiFetch, readApiError } from '../../api'
import {
  DEFAULT_BACKGROUND_BLUR,
  DEFAULT_BACKGROUND_DIM,
  DEFAULT_BACKGROUND_OPACITY,
} from '../dashboard/backgroundSamples'

export function DashboardsManager() {
  const navigate = useNavigate()
  const [dashboards, setDashboards] = useState<Dashboard[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [busyId, setBusyId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    try {
      const res = await fetch('/api/dashboards')
      const payload = await res.json()
      if (res.ok && payload?.status === 'success' && Array.isArray(payload.data)) {
        setDashboards(payload.data)
      } else {
        setError('Unable to load dashboards')
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const createDashboard = useCallback(async () => {
    setCreating(true)
    setError(null)
    try {
      const res = await apiFetch('/api/dashboards', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'New dashboard',
          order: dashboards.length,
          background_opacity: DEFAULT_BACKGROUND_OPACITY,
          background_blur: DEFAULT_BACKGROUND_BLUR,
          background_dim: DEFAULT_BACKGROUND_DIM,
        }),
      })
      if (res.ok) {
        await load()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setCreating(false)
    }
  }, [dashboards.length, load])

  const startRename = (dashboard: Dashboard) => {
    setRenamingId(dashboard.id)
    setRenameValue(dashboard.name)
    setError(null)
  }

  const cancelRename = () => {
    setRenamingId(null)
    setRenameValue('')
  }

  const confirmRename = async (dashboard: Dashboard) => {
    const name = renameValue.trim()
    if (!name) return
    setBusyId(dashboard.id)
    setError(null)
    try {
      const res = await apiFetch(`/api/dashboards/${dashboard.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name,
          order: dashboard.order,
          cols: dashboard.cols,
          rows: dashboard.rows,
          background_image: dashboard.background_image,
          background_opacity: dashboard.background_opacity,
          background_blur: dashboard.background_blur,
          background_dim: dashboard.background_dim,
        }),
      })
      if (res.ok) {
        cancelRename()
        await load()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const deleteDashboard = async (dashboard: Dashboard) => {
    if (!window.confirm(`Delete dashboard "${dashboard.name}"?`)) return
    setBusyId(dashboard.id)
    setError(null)
    try {
      const res = await apiFetch(`/api/dashboards/${dashboard.id}`, { method: 'DELETE' })
      if (res.ok) {
        await load()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const rowButtonClass =
    'shrink-0 flex items-center gap-1 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed'

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <div className="shrink-0 flex items-center gap-3 px-4 py-3 border-b border-slate-800/80">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Dashboards</h1>
          <p className="text-[11.5px] text-slate-400">Dashboards shown on the wall panel</p>
        </div>
        <div className="flex-1" />
        <button
          type="button"
          onClick={() => void createDashboard()}
          disabled={creating}
          className="flex items-center gap-1.5 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all cursor-pointer"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>New dashboard</span>
        </button>
        <button
          type="button"
          onClick={() => navigate('/dashboards')}
          className="flex items-center gap-1.5 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-2 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
        >
          <span>Open dashboards</span>
          <ChevronRight className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        {loading && dashboards.length === 0 ? (
          <div className="flex items-center gap-2 text-slate-400 text-sm">
            <RefreshCw className="w-4 h-4 animate-spin" />
            <span>Loading dashboards…</span>
          </div>
        ) : error ? (
          <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-300">
            {error}
          </div>
        ) : dashboards.length === 0 ? (
          <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
            <LayoutDashboard className="w-6 h-6 mx-auto text-slate-500 mb-2" />
            <p className="text-sm text-slate-300">No dashboard yet</p>
            <p className="text-[11.5px] text-slate-500 mt-1">
              Create one and add widgets from the dashboard editor.
            </p>
          </div>
        ) : (
          <ul className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md divide-y divide-slate-800/80 overflow-hidden">
            {dashboards.map((dashboard) => {
              const isRenaming = renamingId === dashboard.id
              const isBusy = busyId === dashboard.id
              return (
                <li key={dashboard.id} className="flex items-center gap-3 px-4 py-3">
                  <span className="w-8 h-8 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
                    <LayoutDashboard className="w-4 h-4" />
                  </span>
                  <div className="min-w-0 flex-1">
                    {isRenaming ? (
                      <input
                        type="text"
                        value={renameValue}
                        autoFocus
                        disabled={isBusy}
                        onChange={(e) => setRenameValue(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') void confirmRename(dashboard)
                          if (e.key === 'Escape') cancelRename()
                        }}
                        className="h-8 w-full max-w-xs rounded-lg border border-[#6d76e8] bg-slate-800 px-2 text-sm text-white focus:outline-none disabled:opacity-60"
                      />
                    ) : (
                      <>
                        <p className="text-sm text-slate-100 truncate">{dashboard.name}</p>
                        <p className="text-[11px] uppercase tracking-[0.08em] text-slate-500">
                          {dashboard.widgets.length} widget{dashboard.widgets.length === 1 ? '' : 's'}
                        </p>
                      </>
                    )}
                  </div>

                  {isRenaming ? (
                    <>
                      <button
                        type="button"
                        disabled={isBusy || !renameValue.trim()}
                        onClick={() => void confirmRename(dashboard)}
                        className={rowButtonClass}
                      >
                        <Check className="w-3.5 h-3.5" />
                        <span>Save</span>
                      </button>
                      <button type="button" disabled={isBusy} onClick={cancelRename} className={rowButtonClass}>
                        <X className="w-3.5 h-3.5" />
                        <span>Cancel</span>
                      </button>
                    </>
                  ) : (
                    <>
                      <button
                        type="button"
                        onClick={() => navigate(`/setup/dashboards/${dashboard.id}`)}
                        className={rowButtonClass}
                      >
                        <span>Edit</span>
                        <ChevronRight className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        disabled={isBusy}
                        onClick={() => startRename(dashboard)}
                        title="Rename dashboard"
                        className={rowButtonClass}
                      >
                        <Edit2 className="w-3.5 h-3.5" />
                        <span>Rename</span>
                      </button>
                      <button
                        type="button"
                        disabled={isBusy}
                        onClick={() => void deleteDashboard(dashboard)}
                        title="Delete dashboard"
                        className="shrink-0 flex items-center gap-1 rounded-lg border border-rose-500/30 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-300 hover:text-rose-200 hover:bg-rose-500/20 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                        <span>Delete</span>
                      </button>
                    </>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </div>
    </div>
  )
}
