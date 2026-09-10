import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChevronRight, LayoutDashboard, Plus, RefreshCw } from 'lucide-react'
import type { OverviewDashboard } from '../../types'
import { apiFetch, readApiError } from '../../api'
import {
  DEFAULT_BACKGROUND_BLUR,
  DEFAULT_BACKGROUND_DIM,
  DEFAULT_BACKGROUND_OPACITY,
} from '../overview/backgroundSamples'

export function DashboardsManager() {
  const navigate = useNavigate()
  const [overviews, setOverviews] = useState<OverviewDashboard[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setError(null)
    try {
      const res = await fetch('/api/overviews')
      const payload = await res.json()
      if (res.ok && payload?.status === 'success' && Array.isArray(payload.data)) {
        setOverviews(payload.data)
      } else {
        setError('Unable to load overviews')
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

  const createOverview = useCallback(async () => {
    setCreating(true)
    setError(null)
    try {
      const res = await apiFetch('/api/overviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'New overview',
          order: overviews.length,
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
  }, [overviews.length, load])

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <div className="shrink-0 flex items-center gap-3 px-4 py-3 border-b border-slate-800/80">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Dashboards</h1>
          <p className="text-[11.5px] text-slate-400">Overview Dashboards shown on the wall panel</p>
        </div>
        <div className="flex-1" />
        <button
          type="button"
          onClick={() => void createOverview()}
          disabled={creating}
          className="flex items-center gap-1.5 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all cursor-pointer"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>New overview</span>
        </button>
        <button
          type="button"
          onClick={() => navigate('/overviews')}
          className="flex items-center gap-1.5 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-2 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
        >
          <span>Open Overviews</span>
          <ChevronRight className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        {loading && overviews.length === 0 ? (
          <div className="flex items-center gap-2 text-slate-400 text-sm">
            <RefreshCw className="w-4 h-4 animate-spin" />
            <span>Loading dashboards…</span>
          </div>
        ) : error ? (
          <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-300">
            {error}
          </div>
        ) : overviews.length === 0 ? (
          <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
            <LayoutDashboard className="w-6 h-6 mx-auto text-slate-500 mb-2" />
            <p className="text-sm text-slate-300">No overview dashboard yet</p>
            <p className="text-[11.5px] text-slate-500 mt-1">
              Create one and add widgets from the Overviews view.
            </p>
          </div>
        ) : (
          <ul className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md divide-y divide-slate-800/80 overflow-hidden">
            {overviews.map((overview) => (
              <li key={overview.id} className="flex items-center gap-3 px-4 py-3">
                <span className="w-8 h-8 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
                  <LayoutDashboard className="w-4 h-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-slate-100 truncate">{overview.name}</p>
                  <p className="text-[11px] uppercase tracking-[0.08em] text-slate-500">
                    {overview.widgets.length} widget{overview.widgets.length === 1 ? '' : 's'}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => navigate('/overviews')}
                  className="shrink-0 flex items-center gap-1 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
                >
                  <span>Edit</span>
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
