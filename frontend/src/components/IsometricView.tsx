import { useState, useEffect, useRef, useCallback } from 'react'
import { Canvas } from '@react-three/fiber'
import { Box, Layers, Edit3, Loader2, Wifi } from 'lucide-react'
import type { Level, Plan } from '../types'
import { IsometricScene } from './IsometricScene'
import { LevelSelector } from './LevelSelector'
import { NavigationControls } from './NavigationControls'
import { useRealtimeDevices } from '../hooks/useRealtimeDevices'

interface IsometricViewProps {
  level: Level | null
  levels: Level[]
  onSelectLevel: (levelId: string) => void
  onSwitchToAdmin: () => void
}

const DEFAULT_ZOOM = 35
const MIN_ZOOM = 15
const MAX_ZOOM = 90

export function IsometricView({
  level,
  levels,
  onSelectLevel,
  onSwitchToAdmin,
}: IsometricViewProps) {
  const [plan, setPlan] = useState<Plan | null>(null)
  const [loading, setLoading] = useState(false)
  const [zoom, setZoom] = useState<number>(DEFAULT_ZOOM)
  const [pan, setPan] = useState<{ x: number; z: number }>({ x: 0, z: 0 })

  // Real-time device placements & WebSocket action synchronization
  const {
    deviceMap,
    placements,
    connected: wsConnected,
    toggleDevice,
  } = useRealtimeDevices(level?.id || null)

  const dragStartRef = useRef<{
    clientX: number
    clientY: number
    startPan: { x: number; z: number }
  } | null>(null)
  const pinchRef = useRef<{ initialDistance: number; initialZoom: number } | null>(null)

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

  // Recenter pan & zoom
  const handleRecenter = useCallback(() => {
    setPan({ x: 0, z: 0 })
    setZoom(DEFAULT_ZOOM)
  }, [])

  // Zoom handlers
  const handleZoomIn = useCallback(() => {
    setZoom((prev) => Math.min(MAX_ZOOM, Math.round(prev * 1.25)))
  }, [])

  const handleZoomOut = useCallback(() => {
    setZoom((prev) => Math.max(MIN_ZOOM, Math.round(prev / 1.25)))
  }, [])

  // Touch handlers for 2-finger pinch-to-zoom
  const handleTouchStart = (e: React.TouchEvent<HTMLDivElement>) => {
    if (e.touches.length === 2) {
      const t1 = e.touches[0]
      const t2 = e.touches[1]
      const dist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY)
      pinchRef.current = { initialDistance: dist, initialZoom: zoom }
      dragStartRef.current = null
    }
  }

  const handleTouchMove = (e: React.TouchEvent<HTMLDivElement>) => {
    if (e.touches.length === 2 && pinchRef.current) {
      const t1 = e.touches[0]
      const t2 = e.touches[1]
      const dist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY)
      if (pinchRef.current.initialDistance > 0) {
        const ratio = dist / pinchRef.current.initialDistance
        const newZoom = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round(pinchRef.current.initialZoom * ratio)))
        setZoom(newZoom)
      }
    }
  }

  const handleTouchEnd = (e: React.TouchEvent<HTMLDivElement>) => {
    if (e.touches.length < 2) {
      pinchRef.current = null
    }
  }

  // Pointer handlers for 1-finger / mouse pan
  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (pinchRef.current) return
    if (e.button !== 0) return // Only primary button

    dragStartRef.current = {
      clientX: e.clientX,
      clientY: e.clientY,
      startPan: { ...pan },
    }
    try {
      e.currentTarget.setPointerCapture(e.pointerId)
    } catch {
      // Ignore if pointer capture fails
    }
  }

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragStartRef.current || pinchRef.current) return

    const deltaX = e.clientX - dragStartRef.current.clientX
    const deltaY = e.clientY - dragStartRef.current.clientY

    // Project screen pan onto 3D isometric ground plane
    const kX = (deltaX / zoom) * 0.70710678
    const kY = (deltaY / zoom) * 1.22474487

    setPan({
      x: dragStartRef.current.startPan.x + (kX + kY),
      z: dragStartRef.current.startPan.z + (-kX + kY),
    })
  }

  const handlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    dragStartRef.current = null
    try {
      if (e.currentTarget.hasPointerCapture(e.pointerId)) {
        e.currentTarget.releasePointerCapture(e.pointerId)
      }
    } catch {
      // Ignore
    }
  }

  // Mouse wheel zoom
  const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
    e.preventDefault()
    const factor = e.deltaY > 0 ? 0.9 : 1.1
    setZoom((prev) => Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round(prev * factor))))
  }

  const hasWallsOrZones = plan && (plan.walls.length > 0 || plan.zones.length > 0)

  return (
    <section
      className="relative flex-1 w-full h-full min-h-[420px] min-h-0 bg-slate-950 overflow-hidden select-none touch-none cursor-grab active:cursor-grabbing"
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerCancel={handlePointerUp}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onTouchCancel={handleTouchEnd}
      onWheel={handleWheel}
    >
      {/* 3D WebGL Canvas */}
      <div className="absolute inset-0 w-full h-full">
        <Canvas
          orthographic
          camera={{ position: [30, 30, 30], zoom: zoom, near: -100, far: 300 }}
          className="w-full h-full"
        >
          <IsometricScene
            plan={plan}
            zoom={zoom}
            pan={pan}
            placements={placements}
            deviceMap={deviceMap}
            onToggleDevice={toggleDevice}
          />
        </Canvas>
      </div>

      {/* Top Floating Bar: Level Info & Floating Level Selector */}
      <div className="absolute top-4 left-4 right-4 z-20 flex flex-wrap items-center justify-between gap-3 pointer-events-none">
        {/* Left Badge: Camera & Orientation info + WebSocket Live Status */}
        <div className="bg-slate-900/90 backdrop-blur-md px-3.5 py-2 rounded-2xl border border-slate-800/80 text-xs shadow-2xl shadow-black/50 space-y-1 pointer-events-auto">
          <div className="flex items-center space-x-2 text-slate-200 font-semibold">
            <Box className="w-4 h-4 text-indigo-400" />
            <span>Vue Isométrique Fixe</span>
            {loading && <Loader2 className="w-3.5 h-3.5 text-indigo-400 animate-spin ml-1" />}
          </div>
          <div className="flex items-center space-x-3 text-[11px] text-slate-400">
            <span>{level ? `${level.name} • Façade en bas` : 'Orientation façade avant'}</span>
            <span className="text-slate-600">•</span>
            <div className="flex items-center space-x-1.5">
              <Wifi
                className={`w-3 h-3 ${
                  wsConnected ? 'text-emerald-400' : 'text-amber-400'
                }`}
              />
              <span
                className={`w-2 h-2 rounded-full ${
                  wsConnected
                    ? 'bg-emerald-400 shadow-sm shadow-emerald-400/50 animate-pulse'
                    : 'bg-amber-400'
                }`}
              />
              <span className={wsConnected ? 'text-emerald-400 font-medium' : 'text-amber-400 font-medium'}>
                {wsConnected ? 'WebSocket Live' : 'Polling'}
              </span>
            </div>
          </div>
        </div>

        {/* Center / Right: Floating Tactile Level Selector */}
        <div className="pointer-events-auto">
          <LevelSelector
            levels={levels}
            activeLevelId={level?.id || null}
            onSelectLevel={onSelectLevel}
          />
        </div>
      </div>

      {/* Bottom-Right: Tactile Navigation HUD (Recenter, Zoom In/Out) */}
      <div className="absolute bottom-4 right-4 z-20 pointer-events-none">
        <NavigationControls
          zoom={zoom}
          defaultZoom={DEFAULT_ZOOM}
          onZoomIn={handleZoomIn}
          onZoomOut={handleZoomOut}
          onRecenter={handleRecenter}
        />
      </div>

      {/* Bottom-Left: Legend Indicators */}
      <div className="absolute bottom-4 left-4 z-20 pointer-events-none flex flex-wrap gap-2">
        <div className="bg-slate-900/90 backdrop-blur-md border border-slate-800/80 px-3 py-1.5 rounded-xl text-xs text-slate-300 flex items-center space-x-2 shadow-xl shadow-black/40">
          <span className="w-2.5 h-2.5 rounded-full bg-amber-400 shadow-sm shadow-amber-400/50 animate-pulse"></span>
          <span>Light Halo: Éclairage ALLUMÉ</span>
        </div>
        <div className="bg-slate-900/90 backdrop-blur-md border border-slate-800/80 px-3 py-1.5 rounded-xl text-xs text-slate-300 flex items-center space-x-2 shadow-xl shadow-black/40">
          <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 shadow-sm shadow-emerald-400/50"></span>
          <span>Sensors: Valeur Permanente</span>
        </div>
        <div className="bg-slate-900/90 backdrop-blur-md border border-slate-800/80 px-3 py-1.5 rounded-xl text-xs text-slate-300 flex items-center space-x-2 shadow-xl shadow-black/40">
          <span className="w-2.5 h-2.5 rounded-full bg-cyan-400 shadow-sm shadow-cyan-400/50"></span>
          <span>Actionneurs: Tap pour Commuter</span>
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
              <h3 className="text-base font-semibold text-white">Aucun niveau configuré</h3>
              <p className="text-xs text-slate-400 leading-relaxed">
                Créez vos étages et espaces extérieurs dans l'éditeur pour afficher et naviguer dans la vue 3D isométrique.
              </p>
            </div>
            <button
              type="button"
              onClick={onSwitchToAdmin}
              className="w-full bg-indigo-600 hover:bg-indigo-500 active:scale-98 text-white px-4 py-2.5 rounded-xl text-xs font-semibold shadow-lg shadow-indigo-500/30 transition-all flex items-center justify-center space-x-2"
            >
              <Edit3 className="w-4 h-4" />
              <span>Ouvrir l'Éditeur 2D (Admin)</span>
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
                Plan vide pour &laquo; {level.name} &raquo;
              </h4>
              <p className="text-[11px] text-slate-400">
                Ce niveau ne contient pas encore de murs ni de zones.
              </p>
            </div>
            <button
              type="button"
              onClick={onSwitchToAdmin}
              className="shrink-0 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white px-3 py-1.5 rounded-lg text-xs font-semibold shadow-md transition-all flex items-center space-x-1.5"
            >
              <Edit3 className="w-3.5 h-3.5" />
              <span>Dessiner</span>
            </button>
          </div>
        </div>
      )}
    </section>
  )
}
