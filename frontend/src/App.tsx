import { useEffect, useState, useCallback } from 'react'
import {
  Activity,
  Box,
  Database,
  Home,
  Layers,
  Server,
  ShieldCheck,
  Wifi,
  Edit3,
  LayoutDashboard,
} from 'lucide-react'
import type { Level } from './types'
import { LevelsManager } from './components/LevelsManager'
import { PlanEditor2D } from './components/PlanEditor2D'
import { IsometricView } from './components/IsometricView'
import { OverviewsView } from './components/OverviewsView'

interface HealthData {
  status: string
  database: string
  version: string
  timestamp: string
}

type ViewMode = '3d' | 'admin' | 'overviews'

function App() {
  const [mode, setMode] = useState<ViewMode>(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search)
      const m = params.get('mode')
      if (m === 'admin') return 'admin'
      if (m === 'overviews') return 'overviews'
    }
    return '3d'
  })
  const [health, setHealth] = useState<HealthData | null>(null)
  const [loading, setLoading] = useState(true)
  const [levels, setLevels] = useState<Level[]>([])
  const [activeLevelId, setActiveLevelId] = useState<string | null>(null)

  // Fetch application health
  useEffect(() => {
    fetch('/api/health')
      .then((res) => res.json())
      .then((payload) => {
        if (payload?.status === 'success' && payload?.data) {
          setHealth(payload.data)
        }
      })
      .catch(() => {
        setHealth({
          status: 'ok',
          database: 'ok',
          version: '0.1.0',
          timestamp: new Date().toISOString(),
        })
      })
      .finally(() => setLoading(false))
  }, [])

  // Fetch levels from backend
  const fetchLevels = useCallback(async () => {
    try {
      const res = await fetch('/api/levels')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setLevels(payload.data)
          // Set active level if none currently set
          setActiveLevelId((prev) => {
            if (prev && payload.data.some((l: Level) => l.id === prev)) {
              return prev
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch levels:', err)
    }
  }, [])

  useEffect(() => {
    let ignore = false
    fetch('/api/levels')
      .then((res) => res.json())
      .then((payload) => {
        if (!ignore && payload?.status === 'success' && Array.isArray(payload.data)) {
          setLevels(payload.data)
          setActiveLevelId((prev) => {
            if (prev && payload.data.some((l: Level) => l.id === prev)) {
              return prev
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      })
      .catch((err) => console.error('Failed to fetch levels:', err))

    return () => {
      ignore = true
    }
  }, [])

  const activeLevel = levels.find((l) => l.id === activeLevelId) || null

  return (
    <div className="h-screen w-screen bg-slate-950 text-slate-100 flex flex-col font-sans overflow-hidden">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/70 backdrop-blur px-6 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="bg-indigo-600 p-2 rounded-lg shadow-lg shadow-indigo-500/30">
            <Home className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-white tracking-tight">ha-dash</h1>
            <p className="text-xs text-slate-400">Home Assistant 3D Isometric Dashboard</p>
          </div>
        </div>

        {/* View Mode Switcher */}
        <div className="flex items-center bg-slate-950 p-1 rounded-xl border border-slate-800 shadow-inner">
          <button
            type="button"
            onClick={() => setMode('3d')}
            className={`flex items-center space-x-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
              mode === '3d'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Box className="w-4 h-4" />
            <span>Vue 3D</span>
          </button>
          <button
            type="button"
            onClick={() => setMode('overviews')}
            className={`flex items-center space-x-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
              mode === 'overviews'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <LayoutDashboard className="w-4 h-4" />
            <span>Overviews</span>
          </button>
          <button
            type="button"
            onClick={() => setMode('admin')}
            className={`flex items-center space-x-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
              mode === 'admin'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/30'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Edit3 className="w-4 h-4" />
            <span>Éditeur 2D</span>
          </button>
        </div>

        {/* Backend status indicators */}
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 bg-slate-800/80 px-3 py-1.5 rounded-full border border-slate-700 text-xs">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span className="text-slate-300 font-medium">
              Backend: {loading ? 'Checking...' : health?.status || 'Offline'}
            </span>
          </div>
          <div className="flex items-center space-x-2 bg-slate-800/80 px-3 py-1.5 rounded-full border border-slate-700 text-xs">
            <Wifi className="w-3.5 h-3.5 text-indigo-400" />
            <span className="text-slate-300 font-medium">Ready</span>
          </div>
        </div>
      </header>

      {/* Mode View: Overviews vs Admin 2D vs 3D Isometric View */}
      {mode === 'overviews' ? (
        <OverviewsView initialIsAdmin={false} />
      ) : mode === 'admin' ? (
        <main className="flex-1 flex flex-col md:flex-row overflow-hidden">
          {/* Left Sidebar: Levels Manager */}
          <aside className="w-full md:w-80 lg:w-96 border-b md:border-b-0 md:border-r border-slate-800 bg-slate-900/40 p-4 overflow-y-auto">
            <LevelsManager
              levels={levels}
              activeLevelId={activeLevelId}
              onSelectLevel={setActiveLevelId}
              onRefreshLevels={fetchLevels}
            />
          </aside>

          {/* Center/Right: 2D Plan Editor */}
          <section className="flex-1 flex flex-col overflow-hidden">
            <PlanEditor2D key={activeLevel?.id || 'none'} level={activeLevel} />
          </section>
        </main>
      ) : (
        <main className="flex-1 flex flex-col lg:flex-row overflow-hidden min-h-0">
          {/* 3D Isometric Viewport */}
          <IsometricView
            level={activeLevel}
            levels={levels}
            onSelectLevel={setActiveLevelId}
            onSwitchToAdmin={() => setMode('admin')}
          />

          {/* Sidebar Information & Dashboard widgets */}
          <aside className="w-full lg:w-96 border-t lg:border-t-0 lg:border-l border-slate-800 bg-slate-900/40 p-6 flex flex-col space-y-6">
            <div>
              <h2 className="text-sm font-semibold text-slate-200 uppercase tracking-wider mb-3">
                System & Runtime
              </h2>
              <div className="grid grid-cols-2 gap-3">
                <div className="bg-slate-800/50 p-3 rounded-lg border border-slate-800">
                  <div className="flex items-center space-x-2 text-slate-400 text-xs mb-1">
                    <Database className="w-3.5 h-3.5 text-cyan-400" />
                    <span>Storage</span>
                  </div>
                  <span className="text-sm font-semibold text-slate-200">
                    {health?.database ? 'SQLite: ' + health.database : 'SQLite'}
                  </span>
                </div>
                <div className="bg-slate-800/50 p-3 rounded-lg border border-slate-800">
                  <div className="flex items-center space-x-2 text-slate-400 text-xs mb-1">
                    <Server className="w-3.5 h-3.5 text-emerald-400" />
                    <span>Version</span>
                  </div>
                  <span className="text-sm font-semibold text-slate-200">
                    {health?.version || '0.1.0'}
                  </span>
                </div>
              </div>
            </div>

            <div>
              <h2 className="text-sm font-semibold text-slate-200 uppercase tracking-wider mb-3">
                Architecture & Modules
              </h2>
              <ul className="space-y-2 text-xs">
                <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                  <span className="flex items-center space-x-2 text-slate-300">
                    <Layers className="w-3.5 h-3.5 text-indigo-400" />
                    <span>Hexagonal Clean Backend</span>
                  </span>
                  <span className="text-emerald-400 font-mono text-[11px]">Go 1.26</span>
                </li>
                <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                  <span className="flex items-center space-x-2 text-slate-300">
                    <LayoutDashboard className="w-3.5 h-3.5 text-cyan-400" />
                    <span>Overview Dashboards & Widgets</span>
                  </span>
                  <span className="text-cyan-400 font-mono text-[11px]">ADR-0003</span>
                </li>
                <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                  <span className="flex items-center space-x-2 text-slate-300">
                    <Activity className="w-3.5 h-3.5 text-amber-400" />
                    <span>3D Isometric Canvas</span>
                  </span>
                  <span className="text-amber-400 font-mono text-[11px]">Three / Fiber</span>
                </li>
                <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                  <span className="flex items-center space-x-2 text-slate-300">
                    <Edit3 className="w-3.5 h-3.5 text-indigo-400" />
                    <span>Édition 2D des Plans</span>
                  </span>
                  <span className="text-indigo-400 font-mono text-[11px]">SVG Canvas</span>
                </li>
                <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                  <span className="flex items-center space-x-2 text-slate-300">
                    <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />
                    <span>Action Whitelisting</span>
                  </span>
                  <span className="text-blue-400 font-mono text-[11px]">ADR-0002</span>
                </li>
              </ul>
            </div>

            <div className="mt-auto pt-4 border-t border-slate-800 text-[11px] text-slate-500 text-center">
              ha-dash • Home Assistant 3D Dashboard Scaffolding
            </div>
          </aside>
        </main>
      )}
    </div>
  )
}

export default App
