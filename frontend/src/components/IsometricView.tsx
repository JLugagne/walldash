import { useState, useEffect, useCallback, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Canvas } from '@react-three/fiber'
import { Layers, Edit3, Maximize2, Minimize2 } from 'lucide-react'
import type { Level, Plan } from '../types'
import { IsometricScene } from './IsometricScene'
import { LevelSelector } from './LevelSelector'
import { LayerSelector } from './LayerSelector'
import { ViewModeMenu } from './ViewModeMenu'
import { useRealtimeDevices } from '../hooks/useRealtimeDevices'
import { DEFAULT_LAYERS, resolveActiveLayer } from '../utils/layers'

interface IsometricViewProps {
  level: Level | null
  levels: Level[]
  onSelectLevel: (levelId: string) => void
}

export function IsometricView({
  level,
  levels,
  onSelectLevel,
}: IsometricViewProps) {
  const navigate = useNavigate()
  const [plan, setPlan] = useState<Plan | null>(null)
  const [loading, setLoading] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [viewAngle, setViewAngle] = useState(() => {
    try {
      const stored = localStorage.getItem('ha_dash_view_angle')
      return stored !== null ? parseFloat(stored) : 0.6
    } catch {
      return 0.6
    }
  })

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    setViewAngle((prev) => {
      const next = Math.max(0, Math.min(1, prev + e.deltaY * 0.0003))
      try {
        localStorage.setItem('ha_dash_view_angle', String(next))
      } catch { /* ignore storage errors */ }
      return next
    })
  }, [])

  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(() => {})
    } else {
      document.exitFullscreen().catch(() => {})
    }
  }, [])

  useEffect(() => {
    const handler = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', handler)
    return () => document.removeEventListener('fullscreenchange', handler)
  }, [])

  // Display layer state for 3D equipment filtering (persisted across views)
  const [selectedLayer, setSelectedLayer] = useState<string | null>(() => {
    try {
      return typeof window !== 'undefined' ? localStorage.getItem('ha_dash_active_layer') : null
    } catch {
      return null
    }
  })
  const availableLayers = useMemo(() => 
    level?.layers && level.layers.length > 0 ? level.layers : DEFAULT_LAYERS, [level?.layers])
  const activeLayer = resolveActiveLayer(selectedLayer, availableLayers)

  const handleSelectLayer = (layer: string) => {
    setSelectedLayer(layer)
    try {
      localStorage.setItem('ha_dash_active_layer', layer)
    } catch {
      // ignore storage errors
    }
  }

  // Real-time device placements & WebSocket action synchronization
  const { deviceMap, placements, pendingDevices, toggleDevice } = useRealtimeDevices(level?.id || null)

  // Fetch plan whenever active level changes
  useEffect(() => {
    if (!level) {
      setPlan(null)
      return
    }

    let ignore = false
    setLoading(true)

    fetch(`/api/levels/${level.id}/plan`)
      .then(async (res) => {
        if (!res.ok) {
          // Empty / missing plan
          return { status: 'success', data: { level_id: level.id, walls: [], zones: [] } }
        }
        return res.json()
      })
      .then((payload) => {
        if (!ignore && payload?.status === 'success' && payload?.data) {
          setPlan({
            level_id: level.id,
            walls: payload.data.walls || [],
            zones: payload.data.zones || [],
          })
        }
      })
      .catch((err) => {
        if (!ignore) {
          console.error('Failed to load level plan:', err)
          setPlan({ level_id: level.id, walls: [], zones: [] })
        }
      })
      .finally(() => {
        if (!ignore) {
          setLoading(false)
        }
      })

    return () => {
      ignore = true
    }
  }, [level])

  const hasWallsOrZones = plan && (plan.walls.length > 0 || plan.zones.length > 0)

  return (
    <section className="relative flex-1 w-full h-full min-h-[420px] min-h-0 bg-[#0b0f19] overflow-hidden select-none" onWheel={handleWheel}>
      {/* 3D WebGL Canvas */}
      <div className="absolute inset-0 w-full h-full">
        <Canvas
          shadows
          dpr={[1, 2]}
          camera={{ position: [0, 22, 32], fov: 42, near: 0.5, far: 500 }}
          className="w-full h-full"
        >
          <color attach="background" args={['#0b0f19']} />
          <IsometricScene
            plan={plan}
            viewAngle={viewAngle}
            placements={placements}
            deviceMap={deviceMap}
            pendingDevices={pendingDevices}
            onToggleDevice={toggleDevice}
            activeLayer={activeLayer}
            layers={availableLayers}
          />
        </Canvas>
      </div>

      {/* Top-Center Floating Level & Layer Selectors */}
      <div className="absolute top-6 left-1/2 -translate-x-1/2 z-20 pointer-events-none flex flex-col items-center space-y-3">
        <LevelSelector
          levels={levels}
          activeLevelId={level?.id || null}
          onSelectLevel={onSelectLevel}
        />

        {level && availableLayers.length > 1 && (
          <LayerSelector
            layers={level.layers}
            activeLayer={activeLayer}
            onSelectLayer={handleSelectLayer}
          />
        )}
      </div>

      {/* Bottom-Right Floating Controls */}
      <div className="absolute bottom-6 right-6 z-20 pointer-events-none flex flex-col items-end space-y-3">
        {/* Fullscreen + View Mode side by side */}
        <div className="pointer-events-auto flex items-center space-x-3">
          <button
            type="button"
            onClick={toggleFullscreen}
            title={isFullscreen ? 'Exit full screen' : 'Full screen'}
            aria-label={isFullscreen ? 'Exit full screen' : 'Full screen'}
            className="w-12 h-12 rounded-2xl bg-slate-900/90 backdrop-blur-md border border-slate-800/80 shadow-2xl shadow-black/50 flex items-center justify-center text-indigo-400 hover:text-white hover:bg-slate-800/90 active:scale-95 transition-all cursor-pointer"
          >
            {isFullscreen ? (
              <Minimize2 className="w-5 h-5" />
            ) : (
              <Maximize2 className="w-5 h-5" />
            )}
          </button>
          <ViewModeMenu direction="up" />
        </div>
      </div>

      {/* Empty State: No levels registered */}
      {levels.length === 0 && (
        <div className="absolute inset-0 z-10 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-sm pointer-events-auto">
          <div className="max-w-md w-full bg-slate-900/95 border border-slate-800 rounded-2xl p-6 shadow-2xl text-center space-y-4">
            <div className="w-12 h-12 rounded-xl bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 flex items-center justify-center mx-auto">
              <Layers className="w-6 h-6" />
            </div>
            <div className="space-y-1">
              <h3 className="text-base font-semibold text-white">No levels configured</h3>
              <p className="text-xs text-slate-400 leading-relaxed">
                Create your floors and outdoor spaces in the editor to view and navigate in 3D.
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate('/admin')}
              className="w-full bg-indigo-600 hover:bg-indigo-500 active:scale-98 text-white px-4 py-2.5 rounded-xl text-xs font-semibold shadow-lg shadow-indigo-500/30 transition-all flex items-center justify-center space-x-2"
            >
              <Edit3 className="w-4 h-4" />
              <span>Open 2D Editor (Admin)</span>
            </button>
          </div>
        </div>
      )}

      {/* Empty State: Level has no walls and no zones */}
      {levels.length > 0 && level && !loading && !hasWallsOrZones && (
        <div className="absolute top-20 left-1/2 -translate-x-1/2 z-10 pointer-events-auto">
          <div className="bg-slate-900/95 backdrop-blur-md border border-slate-800 rounded-2xl p-4 shadow-2xl flex items-center space-x-4 max-w-lg">
            <div className="w-10 h-10 rounded-xl bg-amber-500/20 border border-amber-500/30 text-amber-400 flex items-center justify-center shrink-0">
              <Edit3 className="w-5 h-5" />
            </div>
            <div className="space-y-0.5">
              <h4 className="text-xs font-semibold text-white">
                Empty plan for &laquo; {level.name} &raquo;
              </h4>
              <p className="text-[11px] text-slate-400">
                This level has no walls or zones yet.
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate('/admin')}
              className="shrink-0 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white px-3 py-1.5 rounded-lg text-xs font-semibold shadow-md transition-all flex items-center space-x-1.5"
            >
              <Edit3 className="w-3.5 h-3.5" />
              <span>Draw</span>
            </button>
          </div>
        </div>
      )}
    </section>
  )
}
