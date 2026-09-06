import { useEffect, useState, useCallback } from 'react'
import type { Level } from './types'
import { PlanEditor2D } from './components/PlanEditor2D'
import { IsometricView } from './components/IsometricView'
import { OverviewsView } from './components/OverviewsView'
import { ViewModeMenu, type ViewMode } from './components/ViewModeMenu'

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
  const [levels, setLevels] = useState<Level[]>([])
  const [activeLevelId, setActiveLevelId] = useState<string | null>(null)

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
    fetchLevels()
  }, [fetchLevels])

  const activeLevel = levels.find((l) => l.id === activeLevelId) || null

  return (
    <div className="h-screen w-screen bg-slate-950 text-slate-100 flex flex-col font-sans overflow-hidden">
      {mode === 'overviews' ? (
        <OverviewsView
          initialIsAdmin={false}
          viewModeMenu={<ViewModeMenu mode={mode} onSelect={setMode} />}
        />
      ) : mode === 'admin' ? (
        <main className="flex-1 flex flex-col min-h-0 overflow-hidden">
          <PlanEditor2D
            level={activeLevel}
            levels={levels}
            onSelectLevel={setActiveLevelId}
            onRefreshLevels={fetchLevels}
            viewModeMenu={<ViewModeMenu mode={mode} onSelect={setMode} />}
          />
        </main>
      ) : (
        <IsometricView
          level={activeLevel}
          levels={levels}
          onSelectLevel={setActiveLevelId}
          onSwitchToAdmin={() => setMode('admin')}
          viewMode={mode}
          onSelectViewMode={setMode}
        />
      )}
    </div>
  )
}

export default App
