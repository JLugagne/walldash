import { useState, useEffect, useRef, useMemo, useCallback } from 'react'
import {
  Save,
  Undo,
  RotateCcw,
  Trash2,
  Grid,
  Square,
  Minus,
  Pointer,
  Check,
  AlertCircle,
  Maximize2,
  Palette,
  Lightbulb,
  Power,
  Thermometer,
  Flame,
  Music,
  Cpu,
  Search,
  X,
  GripVertical,
  Plus,
} from 'lucide-react'
import type { Level, Plan, WallSegment, Zone, Point2D, Device, DevicePlacement, SavePlacementRequest } from '../types'

interface PlanEditor2DProps {
  level: Level | null
}

type ToolMode = 'select' | 'wall' | 'zone'

const COLOR_PRESETS = [
  { name: 'Bleu Salon', value: '#3b82f6' },
  { name: 'Vert Jardin', value: '#10b981' },
  { name: 'Ambre Cuisine', value: '#f59e0b' },
  { name: 'Violet Chambre', value: '#8b5cf6' },
  { name: 'Rose Salle de bain', value: '#ec4899' },
  { name: 'Gris Couloir', value: '#64748b' },
  { name: 'Émeraude Terrasse', value: '#059669' },
]

const DOMAIN_CATEGORIES = [
  { key: 'all', label: 'Tous' },
  { key: 'light', label: 'Lumières' },
  { key: 'switch', label: 'Prises' },
  { key: 'sensor', label: 'Capteurs' },
  { key: 'climate', label: 'Chauffage' },
  { key: 'media_player', label: 'Médias' },
]

let idCounter = 0
function generateId(prefix: string): string {
  idCounter += 1
  return `${prefix}-${Date.now()}-${idCounter}`
}

function getDomainColor(domain: string) {
  switch (domain) {
    case 'light':
      return {
        bg: 'bg-amber-500/10',
        text: 'text-amber-400',
        border: 'border-amber-500/30',
        fill: '#f59e0b',
        ring: '#fbbf24',
      }
    case 'switch':
      return {
        bg: 'bg-cyan-500/10',
        text: 'text-cyan-400',
        border: 'border-cyan-500/30',
        fill: '#06b6d4',
        ring: '#22d3ee',
      }
    case 'sensor':
      return {
        bg: 'bg-emerald-500/10',
        text: 'text-emerald-400',
        border: 'border-emerald-500/30',
        fill: '#10b981',
        ring: '#34d399',
      }
    case 'climate':
      return {
        bg: 'bg-rose-500/10',
        text: 'text-rose-400',
        border: 'border-rose-500/30',
        fill: '#f43f5e',
        ring: '#fb7185',
      }
    case 'media_player':
      return {
        bg: 'bg-purple-500/10',
        text: 'text-purple-400',
        border: 'border-purple-500/30',
        fill: '#a855f7',
        ring: '#c084fc',
      }
    default:
      return {
        bg: 'bg-indigo-500/10',
        text: 'text-indigo-400',
        border: 'border-indigo-500/30',
        fill: '#6366f1',
        ring: '#818cf8',
      }
  }
}

function getDomainIcon(domain: string, className = 'w-4 h-4') {
  switch (domain) {
    case 'light':
      return <Lightbulb className={className} />
    case 'switch':
      return <Power className={className} />
    case 'sensor':
      return <Thermometer className={className} />
    case 'climate':
      return <Flame className={className} />
    case 'media_player':
      return <Music className={className} />
    default:
      return <Cpu className={className} />
  }
}

export function PlanEditor2D({ level }: PlanEditor2DProps) {
  const [plan, setPlan] = useState<Plan>({
    level_id: level?.id || '',
    walls: [],
    zones: [],
  })
  const [history, setHistory] = useState<Plan[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [savedNotification, setSavedNotification] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Devices & placements state
  const [devices, setDevices] = useState<Device[]>([])
  const [placements, setPlacements] = useState<DevicePlacement[]>([])
  const [loadingDevices, setLoadingDevices] = useState(false)
  const [showDevicePalette, setShowDevicePalette] = useState(true)
  const [deviceCategory, setDeviceCategory] = useState<string>('all')
  const [deviceSearch, setDeviceSearch] = useState<string>('')
  const [deviceToPlace, setDeviceToPlace] = useState<Device | null>(null)

  // Dragging placed devices on SVG canvas
  const [draggingPlacementId, setDraggingPlacementId] = useState<string | null>(null)
  const [dragOffset, setDragOffset] = useState<Point2D>({ x: 0, y: 0 })
  const [isMovedDuringDrag, setIsMovedDuringDrag] = useState(false)

  // Editor configuration
  const [tool, setTool] = useState<ToolMode>('wall')
  const [snapGrid, setSnapGrid] = useState(true)
  const [gridSize, setGridSize] = useState(20)
  const [wallThickness, setWallThickness] = useState(12)

  // Zone creation state
  const [zoneName, setZoneName] = useState('Salon')
  const [zoneColor, setZoneColor] = useState('#3b82f6')
  const [currentZonePoints, setCurrentZonePoints] = useState<Point2D[]>([])

  // Wall creation state
  const [wallStart, setWallStart] = useState<Point2D | null>(null)
  const [cursorPos, setCursorPos] = useState<Point2D | null>(null)

  // Selection state
  const [selectedElement, setSelectedElement] = useState<{
    type: 'wall' | 'zone' | 'device'
    id: string
  } | null>(null)

  const svgRef = useRef<SVGSVGElement | null>(null)

  // Fetch Home Assistant devices
  const fetchDevices = useCallback(async () => {
    setLoadingDevices(true)
    try {
      const res = await fetch('/api/devices')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setDevices(payload.data)
        }
      }
    } catch (err) {
      console.error('Failed to load devices:', err)
    } finally {
      setLoadingDevices(false)
    }
  }, [])

  useEffect(() => {
    fetchDevices()
  }, [fetchDevices])

  // Fetch plan and device placements when level changes
  useEffect(() => {
    if (!level) return
    let ignore = false
    setLoading(true)
    setSelectedElement(null)
    setDeviceToPlace(null)
    setWallStart(null)
    setCurrentZonePoints([])

    // Load plan
    fetch(`/api/levels/${level.id}/plan`)
      .then(async (res) => {
        if (!res.ok) {
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
          console.error('Failed to load plan:', err)
          setError('Erreur lors du chargement du plan')
        }
      })
      .finally(() => {
        if (!ignore) setLoading(false)
      })

    // Load device placements
    fetch(`/api/levels/${level.id}/placements`)
      .then((res) => res.json())
      .then((payload) => {
        if (!ignore && payload?.status === 'success' && Array.isArray(payload.data)) {
          setPlacements(payload.data)
        }
      })
      .catch((err) => {
        if (!ignore) console.error('Failed to load placements:', err)
      })

    return () => {
      ignore = true
    }
  }, [level])

  // Push history snapshot before mutating
  const pushHistory = useCallback((currentPlan: Plan) => {
    setHistory((prev) => [...prev.slice(-20), JSON.parse(JSON.stringify(currentPlan))])
  }, [])

  // Coordinate snapping helper
  const snap = useCallback(
    (coord: Point2D): Point2D => {
      if (!snapGrid) return coord
      return {
        x: Math.round(coord.x / gridSize) * gridSize,
        y: Math.round(coord.y / gridSize) * gridSize,
      }
    },
    [snapGrid, gridSize]
  )

  // Convert client mouse/touch event to SVG coordinate space
  const getSvgCoordinates = useCallback(
    (e: { clientX?: number; clientY?: number; touches?: React.TouchList }): Point2D | null => {
      if (!svgRef.current) return null
      const svg = svgRef.current
      const ctm = svg.getScreenCTM()
      if (!ctm) return null

      let clientX: number | undefined
      let clientY: number | undefined

      if (e.touches && e.touches.length > 0) {
        clientX = e.touches[0].clientX
        clientY = e.touches[0].clientY
      } else if (typeof e.clientX === 'number' && typeof e.clientY === 'number') {
        clientX = e.clientX
        clientY = e.clientY
      }

      if (typeof clientX !== 'number' || typeof clientY !== 'number') {
        return null
      }

      const inverse = ctm.inverse()
      return {
        x: inverse.a * clientX + inverse.c * clientY + inverse.e,
        y: inverse.b * clientX + inverse.d * clientY + inverse.f,
      }
    },
    []
  )

  // Save device placement to backend
  const handleSavePlacement = useCallback(
    async (req: SavePlacementRequest) => {
      if (!level) return
      setError(null)
      try {
        const res = await fetch(`/api/levels/${level.id}/placements`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(req),
        })
        const payload = await res.json()
        if (res.ok && payload.status === 'success') {
          const saved: DevicePlacement = payload.data
          setPlacements((prev) => {
            const idx = prev.findIndex((p) => p.id === saved.id)
            if (idx >= 0) {
              const copy = [...prev]
              copy[idx] = saved
              return copy
            }
            return [...prev, saved]
          })
          return saved
        } else {
          setError(payload.error?.message || 'Erreur lors du placement de l’appareil')
        }
      } catch {
        setError('Impossible de joindre le serveur')
      }
    },
    [level]
  )

  // Delete device placement from backend
  const handleDeletePlacement = useCallback(
    async (placementId: string) => {
      if (!level) return
      setError(null)
      try {
        const res = await fetch(`/api/levels/${level.id}/placements/${placementId}`, {
          method: 'DELETE',
        })
        const payload = await res.json()
        if (res.ok && payload.status === 'success') {
          setPlacements((prev) => prev.filter((p) => p.id !== placementId))
          setSelectedElement((prev) => (prev?.type === 'device' && prev.id === placementId ? null : prev))
        } else {
          setError(payload.error?.message || 'Erreur lors de la suppression du placement')
        }
      } catch {
        setError('Impossible de joindre le serveur')
      }
    },
    [level]
  )

  // Handle pointer movements on SVG
  const handlePointerMove = (e: React.MouseEvent<SVGSVGElement> | React.TouchEvent<SVGSVGElement>) => {
    const raw = getSvgCoordinates(e)
    if (!raw) return
    const snapped = snap(raw)
    setCursorPos(snapped)

    // Handle dragging an existing device placement
    if (draggingPlacementId) {
      setIsMovedDuringDrag(true)
      setPlacements((prev) =>
        prev.map((p) => {
          if (p.id === draggingPlacementId) {
            return {
              ...p,
              x: Math.max(20, Math.min(980, snapped.x - dragOffset.x)),
              y: Math.max(20, Math.min(680, snapped.y - dragOffset.y)),
            }
          }
          return p
        })
      )
    }
  }

  // Handle pointer up (finish dragging a placed device)
  const handlePointerUp = () => {
    if (draggingPlacementId) {
      const moved = placements.find((p) => p.id === draggingPlacementId)
      if (moved && isMovedDuringDrag) {
        handleSavePlacement({
          id: moved.id,
          device_id: moved.device_id,
          x: moved.x,
          y: moved.y,
          icon: moved.icon,
          custom_name: moved.custom_name,
        })
      }
      setDraggingPlacementId(null)
      setIsMovedDuringDrag(false)
    }
  }

  // Handle SVG canvas clicks
  const handleSvgClick = (e: React.MouseEvent<SVGSVGElement>) => {
    if (isMovedDuringDrag) {
      return
    }

    const raw = getSvgCoordinates(e)
    if (!raw) return
    const point = snap(raw)

    // If a device is chosen for placement via click-to-place
    if (deviceToPlace) {
      handleSavePlacement({
        device_id: deviceToPlace.id,
        x: point.x,
        y: point.y,
        custom_name: deviceToPlace.name,
        icon: deviceToPlace.domain,
      })
      setDeviceToPlace(null)
      return
    }

    if (tool === 'wall') {
      if (!wallStart) {
        setWallStart(point)
      } else {
        if (wallStart.x !== point.x || wallStart.y !== point.y) {
          pushHistory(plan)
          const newWall: WallSegment = {
            id: generateId('wall'),
            x1: wallStart.x,
            y1: wallStart.y,
            x2: point.x,
            y2: point.y,
            thickness: wallThickness,
          }
          setPlan((prev) => ({
            ...prev,
            walls: [...prev.walls, newWall],
          }))
        }
        setWallStart(null)
      }
    } else if (tool === 'zone') {
      if (currentZonePoints.length >= 3) {
        const startPoint = currentZonePoints[0]
        const dist = Math.hypot(point.x - startPoint.x, point.y - startPoint.y)
        if (dist <= gridSize * 1.2) {
          completeCurrentZone()
          return
        }
      }
      setCurrentZonePoints((prev) => [...prev, point])
    } else if (tool === 'select') {
      if (e.target === svgRef.current || (e.target as HTMLElement).tagName === 'svg') {
        setSelectedElement(null)
      }
    }
  }

  // HTML5 Drag and Drop handlers for dropping devices onto SVG
  const handleDragOver = (e: React.DragEvent<SVGSVGElement>) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'copy'
    const raw = getSvgCoordinates(e)
    if (raw) {
      setCursorPos(snap(raw))
    }
  }

  const handleDrop = (e: React.DragEvent<SVGSVGElement>) => {
    e.preventDefault()
    const deviceId = e.dataTransfer.getData('text/plain')
    if (!deviceId || !level) return

    const dev = devices.find((d) => d.id === deviceId)
    if (!dev) return

    const raw = getSvgCoordinates(e)
    if (!raw) return
    const point = snap(raw)

    handleSavePlacement({
      device_id: dev.id,
      x: point.x,
      y: point.y,
      custom_name: dev.name,
      icon: dev.domain,
    })
  }

  const completeCurrentZone = () => {
    if (currentZonePoints.length < 3) {
      alert('Une zone doit contenir au moins 3 points pour former un polygone fermé.')
      return
    }
    pushHistory(plan)
    const newZone: Zone = {
      id: generateId('zone'),
      name: zoneName.trim() || 'Zone',
      color: zoneColor,
      points: currentZonePoints,
    }
    setPlan((prev) => ({
      ...prev,
      zones: [...prev.zones, newZone],
    }))
    setCurrentZonePoints([])
  }

  // Delete selected item (wall, zone, or device)
  const handleDeleteSelected = useCallback(() => {
    if (!selectedElement) return
    if (selectedElement.type === 'device') {
      handleDeletePlacement(selectedElement.id)
      return
    }

    pushHistory(plan)
    if (selectedElement.type === 'wall') {
      setPlan((prev) => ({
        ...prev,
        walls: prev.walls.filter((w) => w.id !== selectedElement.id),
      }))
    } else if (selectedElement.type === 'zone') {
      setPlan((prev) => ({
        ...prev,
        zones: prev.zones.filter((z) => z.id !== selectedElement.id),
      }))
    }
    setSelectedElement(null)
  }, [selectedElement, handleDeletePlacement, plan, pushHistory])

  // Global keyboard shortcuts (Delete / Esc)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const activeTag = document.activeElement?.tagName.toLowerCase()
      if (activeTag === 'input' || activeTag === 'textarea' || activeTag === 'select') {
        return
      }

      if (e.key === 'Delete' || e.key === 'Backspace') {
        if (selectedElement) {
          e.preventDefault()
          handleDeleteSelected()
        }
      } else if (e.key === 'Escape') {
        setDeviceToPlace(null)
        setWallStart(null)
        setCurrentZonePoints([])
        setSelectedElement(null)
        setDraggingPlacementId(null)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedElement, handleDeleteSelected])

  // Undo last modification
  const handleUndo = () => {
    if (history.length === 0) return
    const last = history[history.length - 1]
    setHistory((prev) => prev.slice(0, -1))
    setPlan(last)
    setSelectedElement(null)
    setCurrentZonePoints([])
    setWallStart(null)
  }

  // Reset to original plan
  const handleReset = () => {
    if (!confirm('Réinitialiser le plan aux dernières données sauvegardées ?')) return
    if (!level) return
    setHistory([])
    setSelectedElement(null)
    setCurrentZonePoints([])
    setWallStart(null)

    fetch(`/api/levels/${level.id}/plan`)
      .then((res) => res.json())
      .then((payload) => {
        if (payload?.status === 'success' && payload?.data) {
          setPlan({
            level_id: level.id,
            walls: payload.data.walls || [],
            zones: payload.data.zones || [],
          })
        }
      })
  }

  // Clear all walls and zones
  const handleClear = () => {
    if (!confirm('Supprimer tous les murs et toutes les zones de ce niveau ?')) return
    pushHistory(plan)
    setPlan((prev) => ({ ...prev, walls: [], zones: [] }))
    setSelectedElement(null)
    setCurrentZonePoints([])
    setWallStart(null)
  }

  // Save plan walls/zones to backend
  const handleSavePlan = async () => {
    if (!level) return
    setSaving(true)
    setError(null)
    try {
      const res = await fetch(`/api/levels/${level.id}/plan`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          walls: plan.walls,
          zones: plan.zones,
        }),
      })
      const result = await res.json()
      if (res.ok && result.status === 'success') {
        setSavedNotification(true)
        setTimeout(() => setSavedNotification(false), 3000)
      } else {
        setError(result.error?.message || 'Erreur lors de la sauvegarde du plan')
      }
    } catch {
      setError('Impossible de joindre le serveur')
    } finally {
      setSaving(false)
    }
  }

  // Filter devices in palette
  const filteredDevices = useMemo(() => {
    return devices.filter((d) => {
      const matchesCategory = deviceCategory === 'all' || d.domain === deviceCategory
      const matchesSearch =
        !deviceSearch ||
        d.name.toLowerCase().includes(deviceSearch.toLowerCase()) ||
        d.id.toLowerCase().includes(deviceSearch.toLowerCase())
      return matchesCategory && matchesSearch
    })
  }, [devices, deviceCategory, deviceSearch])

  // Count placed devices on this level
  const placedDeviceIds = useMemo(() => {
    return new Set(placements.map((p) => p.device_id))
  }, [placements])

  // SVG grid pattern calculation
  const gridPattern = useMemo(() => {
    return (
      <pattern
        id="editorGrid"
        width={gridSize}
        height={gridSize}
        patternUnits="userSpaceOnUse"
      >
        <path
          d={`M ${gridSize} 0 L 0 0 0 ${gridSize}`}
          fill="none"
          stroke="#334155"
          strokeWidth="0.5"
          strokeDasharray={snapGrid ? 'none' : '2 2'}
        />
        <circle cx={0} cy={0} r="1" fill="#475569" />
      </pattern>
    )
  }, [gridSize, snapGrid])

  if (!level) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-500 bg-slate-950">
        <Maximize2 className="w-12 h-12 text-slate-600 mb-3" />
        <p className="text-sm font-semibold text-slate-400">Aucun niveau sélectionné</p>
        <p className="text-xs text-slate-500 max-w-sm mt-1">
          Sélectionnez un niveau dans la barre latérale ou créez-en un nouveau pour commencer à dessiner le plan 2D et placer vos appareils.
        </p>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col h-full bg-slate-950 overflow-hidden select-none">
      {/* Top Toolbar */}
      <div className="bg-slate-900/90 border-b border-slate-800 p-3 flex flex-wrap items-center justify-between gap-3 text-xs">
        {/* Tool selector */}
        <div className="flex items-center space-x-1 bg-slate-800/80 p-1 rounded-lg border border-slate-700">
          <button
            type="button"
            onClick={() => {
              setTool('wall')
              setWallStart(null)
              setCurrentZonePoints([])
              setDeviceToPlace(null)
            }}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md font-medium transition-all ${
              tool === 'wall'
                ? 'bg-indigo-600 text-white shadow-sm shadow-indigo-500/30'
                : 'text-slate-300 hover:text-white hover:bg-slate-700/60'
            }`}
          >
            <Minus className="w-4 h-4 stroke-[3]" />
            <span>Mur</span>
          </button>

          <button
            type="button"
            onClick={() => {
              setTool('zone')
              setWallStart(null)
              setCurrentZonePoints([])
              setDeviceToPlace(null)
            }}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md font-medium transition-all ${
              tool === 'zone'
                ? 'bg-indigo-600 text-white shadow-sm shadow-indigo-500/30'
                : 'text-slate-300 hover:text-white hover:bg-slate-700/60'
            }`}
          >
            <Square className="w-4 h-4" />
            <span>Zone</span>
          </button>

          <button
            type="button"
            onClick={() => {
              setTool('select')
              setWallStart(null)
              setCurrentZonePoints([])
              setDeviceToPlace(null)
            }}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md font-medium transition-all ${
              tool === 'select'
                ? 'bg-indigo-600 text-white shadow-sm shadow-indigo-500/30'
                : 'text-slate-300 hover:text-white hover:bg-slate-700/60'
            }`}
          >
            <Pointer className="w-4 h-4" />
            <span>Sélection / Gomme</span>
          </button>
        </div>

        {/* Snap to grid controls */}
        <div className="flex items-center space-x-2">
          <button
            type="button"
            onClick={() => setSnapGrid(!snapGrid)}
            className={`flex items-center space-x-1.5 px-2.5 py-1.5 rounded-md border text-xs transition-colors ${
              snapGrid
                ? 'bg-emerald-500/10 border-emerald-500/40 text-emerald-400'
                : 'bg-slate-800 border-slate-700 text-slate-400 hover:text-slate-200'
            }`}
            title="Activer/Désactiver l'aimantation sur la grille"
          >
            <Grid className="w-3.5 h-3.5" />
            <span>Grille: {snapGrid ? 'Aimantée' : 'Libre'}</span>
          </button>

          <select
            value={gridSize}
            onChange={(e) => setGridSize(Number(e.target.value))}
            className="bg-slate-800 border border-slate-700 text-slate-200 rounded-md px-2 py-1.5 text-xs focus:outline-none"
          >
            <option value="10">Pas: 10px</option>
            <option value="20">Pas: 20px</option>
            <option value="40">Pas: 40px</option>
          </select>
        </div>

        {/* HA Devices Drawer Toggle Button */}
        <div className="flex items-center space-x-2">
          <button
            type="button"
            onClick={() => setShowDevicePalette(!showDevicePalette)}
            className={`flex items-center space-x-2 px-3 py-1.5 rounded-md border text-xs font-medium transition-all ${
              showDevicePalette
                ? 'bg-indigo-600 text-white border-indigo-500 shadow-sm shadow-indigo-500/30'
                : 'bg-slate-800 border-slate-700 text-slate-300 hover:text-white hover:bg-slate-700/60'
            }`}
          >
            <Cpu className="w-4 h-4" />
            <span>Appareils HA</span>
            <span className="bg-black/30 text-slate-200 px-1.5 py-0.5 rounded-full text-[10px] font-mono">
              {placements.length}/{devices.length}
            </span>
          </button>
        </div>

        {/* Plan actions */}
        <div className="flex items-center space-x-2">
          <button
            type="button"
            disabled={history.length === 0}
            onClick={handleUndo}
            className="flex items-center space-x-1 bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-300 px-2.5 py-1.5 rounded-md border border-slate-700"
            title="Annuler dernière action"
          >
            <Undo className="w-3.5 h-3.5" />
            <span>Annuler</span>
          </button>

          <button
            type="button"
            onClick={handleReset}
            className="flex items-center space-x-1 bg-slate-800 hover:bg-slate-700 text-slate-300 px-2.5 py-1.5 rounded-md border border-slate-700"
            title="Réinitialiser le plan"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Reset</span>
          </button>

          <button
            type="button"
            onClick={handleClear}
            className="flex items-center space-x-1 bg-slate-800 hover:bg-rose-900/40 text-rose-400 px-2.5 py-1.5 rounded-md border border-slate-700 hover:border-rose-700"
            title="Effacer tout"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>Effacer</span>
          </button>

          <button
            type="button"
            onClick={handleSavePlan}
            disabled={saving}
            className="flex items-center space-x-1.5 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white font-medium px-4 py-1.5 rounded-md shadow-md shadow-emerald-600/20 transition-all ml-1"
          >
            {savedNotification ? (
              <>
                <Check className="w-4 h-4 text-white" />
                <span>Sauvegardé !</span>
              </>
            ) : (
              <>
                <Save className="w-4 h-4" />
                <span>{saving ? 'Enregistrement...' : 'Sauvegarder le Plan'}</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Tool Context Sub-Bar */}
      <div className="bg-slate-900/60 border-b border-slate-800 px-4 py-2 flex flex-wrap items-center justify-between text-xs text-slate-300 gap-2">
        {deviceToPlace && (
          <div className="flex items-center space-x-3 bg-indigo-950/90 border border-indigo-500/50 px-3 py-1 rounded-md">
            <span className="text-indigo-300 font-medium">
              🎯 Cliquez sur le plan pour placer : <strong className="text-white">{deviceToPlace.name}</strong>
            </span>
            <button
              type="button"
              onClick={() => setDeviceToPlace(null)}
              className="text-rose-400 hover:text-rose-300 underline text-xs"
            >
              Annuler (Esc)
            </button>
          </div>
        )}

        {!deviceToPlace && tool === 'wall' && (
          <div className="flex items-center space-x-4">
            <span className="font-semibold text-indigo-400">Outil Mur:</span>
            <span>
              {wallStart
                ? `Point A fixé (${wallStart.x}, ${wallStart.y}) • Cliquez pour fixer le point B`
                : 'Cliquez sur la grille pour placer le premier point'}
            </span>
            <div className="flex items-center space-x-2">
              <span className="text-slate-400">Épaisseur:</span>
              <select
                value={wallThickness}
                onChange={(e) => setWallThickness(Number(e.target.value))}
                className="bg-slate-800 border border-slate-700 rounded px-2 py-0.5 text-xs text-white"
              >
                <option value="8">Fin (8px)</option>
                <option value="12">Standard (12px)</option>
                <option value="16">Porteur (16px)</option>
                <option value="20">Épais (20px)</option>
              </select>
            </div>
            {wallStart && (
              <button
                type="button"
                onClick={() => setWallStart(null)}
                className="text-rose-400 underline ml-2"
              >
                Annuler segment
              </button>
            )}
          </div>
        )}

        {!deviceToPlace && tool === 'zone' && (
          <div className="flex items-center space-x-3 flex-wrap gap-y-1">
            <span className="font-semibold text-indigo-400">Outil Zone:</span>
            <div className="flex items-center space-x-1.5">
              <span>Nom:</span>
              <input
                type="text"
                value={zoneName}
                onChange={(e) => setZoneName(e.target.value)}
                className="bg-slate-800 border border-slate-700 rounded px-2 py-0.5 text-white w-28 text-xs"
              />
            </div>

            <div className="flex items-center space-x-1">
              <Palette className="w-3.5 h-3.5 text-slate-400" />
              {COLOR_PRESETS.map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => {
                    setZoneColor(preset.value)
                    setZoneName(preset.name.split(' ')[1] || zoneName)
                  }}
                  className={`w-4 h-4 rounded-full border transition-transform ${
                    zoneColor === preset.value
                      ? 'scale-125 border-white ring-1 ring-white/50'
                      : 'border-transparent hover:scale-110'
                  }`}
                  style={{ backgroundColor: preset.value }}
                  title={preset.name}
                />
              ))}
              <input
                type="color"
                value={zoneColor}
                onChange={(e) => setZoneColor(e.target.value)}
                className="w-5 h-5 bg-transparent cursor-pointer rounded overflow-hidden"
                title="Couleur personnalisée"
              />
            </div>

            <span className="text-slate-400">{currentZonePoints.length} points placés</span>

            {currentZonePoints.length >= 3 && (
              <button
                type="button"
                onClick={completeCurrentZone}
                className="bg-indigo-600 hover:bg-indigo-500 text-white px-2.5 py-0.5 rounded text-xs flex items-center space-x-1"
              >
                <Check className="w-3 h-3" />
                <span>Boucler la Zone</span>
              </button>
            )}

            {currentZonePoints.length > 0 && (
              <button
                type="button"
                onClick={() => setCurrentZonePoints([])}
                className="text-rose-400 underline"
              >
                Abandonner zone
              </button>
            )}
          </div>
        )}

        {!deviceToPlace && tool === 'select' && (
          <div className="flex items-center space-x-4">
            <span className="font-semibold text-indigo-400">Mode Sélection / Gomme:</span>
            {selectedElement ? (
              <div className="flex items-center space-x-2">
                {selectedElement.type === 'device' ? (
                  <>
                    <span className="text-white">
                      Appareil:{' '}
                      {(() => {
                        const p = placements.find((item) => item.id === selectedElement.id)
                        if (!p) return ''
                        const d = devices.find((dev) => dev.id === p.device_id)
                        return `${p.custom_name || d?.name || p.device_id}`
                      })()}
                    </span>
                    <button
                      type="button"
                      onClick={handleDeleteSelected}
                      className="bg-rose-600 hover:bg-rose-500 text-white px-2.5 py-0.5 rounded text-xs flex items-center space-x-1"
                    >
                      <Trash2 className="w-3 h-3" />
                      <span>Retirer du plan</span>
                    </button>
                  </>
                ) : (
                  <>
                    <span className="text-white">
                      Élément: {selectedElement.type === 'wall' ? 'Segment de Mur' : 'Zone'}
                    </span>
                    <button
                      type="button"
                      onClick={handleDeleteSelected}
                      className="bg-rose-600 hover:bg-rose-500 text-white px-2.5 py-0.5 rounded text-xs flex items-center space-x-1"
                    >
                      <Trash2 className="w-3 h-3" />
                      <span>Supprimer</span>
                    </button>
                  </>
                )}
                <button
                  type="button"
                  onClick={() => setSelectedElement(null)}
                  className="text-slate-400 underline text-xs ml-2"
                >
                  Désélectionner
                </button>
              </div>
            ) : (
              <span className="text-slate-400">
                Cliquez sur un mur, une zone ou un appareil pour le sélectionner et le supprimer.
              </span>
            )}
          </div>
        )}

        {/* Summary info right */}
        <div className="ml-auto text-[11px] text-slate-400 flex items-center space-x-3">
          <span>
            {plan.walls.length} murs • {plan.zones.length} zones • {placements.length} appareils
          </span>
          {cursorPos && (
            <span className="font-mono text-slate-500">
              X: {Math.round(cursorPos.x)} Y: {Math.round(cursorPos.y)}
            </span>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-rose-500/10 border-b border-rose-500/20 px-4 py-1.5 text-xs text-rose-400 flex items-center space-x-2">
          <AlertCircle className="w-4 h-4" />
          <span>{error}</span>
          <button type="button" onClick={() => setError(null)} className="ml-auto text-rose-400 hover:text-white">
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Main Workspace Layout (Canvas + Drawer) */}
      <div className="flex-1 flex flex-row overflow-hidden relative">
        {/* SVG Interactive Canvas */}
        <div className="flex-1 relative overflow-auto bg-slate-950 flex items-center justify-center p-4">
          {loading ? (
            <div className="text-slate-400 text-sm">Chargement du plan...</div>
          ) : (
            <svg
              ref={svgRef}
              viewBox="0 0 1000 700"
              className="w-full h-full max-w-[1000px] max-h-[700px] bg-slate-900/90 rounded-xl border border-slate-800 shadow-2xl shadow-black/60 cursor-crosshair touch-none select-none"
              onMouseMove={handlePointerMove}
              onTouchMove={handlePointerMove}
              onMouseUp={handlePointerUp}
              onTouchEnd={handlePointerUp}
              onClick={handleSvgClick}
              onDragOver={handleDragOver}
              onDrop={handleDrop}
            >
              <defs>{gridPattern}</defs>

              {/* Grid layer */}
              <rect width="1000" height="700" fill="url(#editorGrid)" />

              {/* Render Saved Zones (Polygons) */}
              {plan.zones.map((zone) => {
                const isSelected = selectedElement?.type === 'zone' && selectedElement.id === zone.id
                const pointsStr = zone.points.map((p) => `${p.x},${p.y}`).join(' ')

                const cx = zone.points.reduce((acc, p) => acc + p.x, 0) / zone.points.length
                const cy = zone.points.reduce((acc, p) => acc + p.y, 0) / zone.points.length

                return (
                  <g key={zone.id} className="cursor-pointer">
                    <polygon
                      points={pointsStr}
                      fill={zone.color}
                      fillOpacity={isSelected ? 0.45 : 0.25}
                      stroke={isSelected ? '#ffffff' : zone.color}
                      strokeWidth={isSelected ? 3 : 1.5}
                      strokeDasharray={isSelected ? '4 2' : 'none'}
                      onClick={(e) => {
                        if (tool === 'select') {
                          e.stopPropagation()
                          setSelectedElement({ type: 'zone', id: zone.id })
                        }
                      }}
                    />
                    <text
                      x={cx}
                      y={cy}
                      textAnchor="middle"
                      dominantBaseline="middle"
                      fill="#ffffff"
                      fontSize="12"
                      fontWeight="600"
                      className="pointer-events-none select-none drop-shadow"
                    >
                      {zone.name}
                    </text>
                  </g>
                )
              })}

              {/* Render In-Progress Zone Drawing */}
              {currentZonePoints.length > 0 && (
                <g className="pointer-events-none">
                  {currentZonePoints.length > 1 && (
                    <polyline
                      points={currentZonePoints.map((p) => `${p.x},${p.y}`).join(' ')}
                      fill="none"
                      stroke={zoneColor}
                      strokeWidth="2"
                      strokeDasharray="4 4"
                    />
                  )}
                  {cursorPos && (
                    <line
                      x1={currentZonePoints[currentZonePoints.length - 1].x}
                      y1={currentZonePoints[currentZonePoints.length - 1].y}
                      x2={cursorPos.x}
                      y2={cursorPos.y}
                      stroke={zoneColor}
                      strokeWidth="2"
                      strokeDasharray="2 2"
                    />
                  )}
                  {currentZonePoints.map((pt, idx) => (
                    <circle
                      key={idx}
                      cx={pt.x}
                      cy={pt.y}
                      r={idx === 0 ? 6 : 4}
                      fill={idx === 0 ? '#38bdf8' : zoneColor}
                      stroke="#ffffff"
                      strokeWidth="1.5"
                    />
                  ))}
                </g>
              )}

              {/* Render Saved Walls */}
              {plan.walls.map((wall) => {
                const isSelected = selectedElement?.type === 'wall' && selectedElement.id === wall.id
                return (
                  <g key={wall.id} className="cursor-pointer">
                    <line
                      x1={wall.x1}
                      y1={wall.y1}
                      x2={wall.x2}
                      y2={wall.y2}
                      stroke="transparent"
                      strokeWidth={Math.max(wall.thickness + 10, 20)}
                      strokeLinecap="round"
                      onClick={(e) => {
                        if (tool === 'select') {
                          e.stopPropagation()
                          setSelectedElement({ type: 'wall', id: wall.id })
                        }
                      }}
                    />
                    <line
                      x1={wall.x1}
                      y1={wall.y1}
                      x2={wall.x2}
                      y2={wall.y2}
                      stroke={isSelected ? '#6366f1' : '#94a3b8'}
                      strokeWidth={wall.thickness}
                      strokeLinecap="square"
                      className="transition-all"
                    />
                    <circle cx={wall.x1} cy={wall.y1} r={wall.thickness / 2} fill={isSelected ? '#818cf8' : '#cbd5e1'} />
                    <circle cx={wall.x2} cy={wall.y2} r={wall.thickness / 2} fill={isSelected ? '#818cf8' : '#cbd5e1'} />
                  </g>
                )
              })}

              {/* Render In-Progress Wall Drawing */}
              {wallStart && cursorPos && (
                <g className="pointer-events-none">
                  <line
                    x1={wallStart.x}
                    y1={wallStart.y}
                    x2={cursorPos.x}
                    y2={cursorPos.y}
                    stroke="#818cf8"
                    strokeWidth={wallThickness}
                    strokeLinecap="square"
                    opacity="0.8"
                  />
                  <circle cx={wallStart.x} cy={wallStart.y} r={wallThickness / 2} fill="#818cf8" />
                  <circle cx={cursorPos.x} cy={cursorPos.y} r={wallThickness / 2} fill="#818cf8" />
                </g>
              )}

              {/* Render Placed Home Assistant Devices */}
              {placements.map((p) => {
                const isSelected = selectedElement?.type === 'device' && selectedElement.id === p.id
                const dev = devices.find((d) => d.id === p.device_id)
                const domain = dev?.domain || p.icon || 'light'
                const colors = getDomainColor(domain)
                const displayName = p.custom_name || dev?.name || p.device_id
                const stateLabel = dev?.state || ''

                return (
                  <g
                    key={p.id}
                    className="cursor-move select-none group"
                    onMouseDown={(e) => {
                      e.stopPropagation()
                      const raw = getSvgCoordinates(e)
                      if (raw) {
                        setDraggingPlacementId(p.id)
                        setDragOffset({ x: raw.x - p.x, y: raw.y - p.y })
                        setIsMovedDuringDrag(false)
                      }
                      setSelectedElement({ type: 'device', id: p.id })
                    }}
                    onClick={(e) => {
                      e.stopPropagation()
                      setSelectedElement({ type: 'device', id: p.id })
                    }}
                  >
                    {/* Pulsing selection halo */}
                    {isSelected && (
                      <circle
                        cx={p.x}
                        cy={p.y}
                        r="24"
                        fill="none"
                        stroke={colors.ring}
                        strokeWidth="2"
                        strokeDasharray="4 3"
                        className="animate-pulse"
                      />
                    )}

                    {/* Circular badge background */}
                    <circle
                      cx={p.x}
                      cy={p.y}
                      r="16"
                      fill="#0f172a"
                      stroke={isSelected ? colors.ring : colors.fill}
                      strokeWidth={isSelected ? '2.5' : '1.5'}
                      className="transition-transform group-hover:scale-110"
                    />

                    {/* Domain icon SVG path */}
                    <g transform={`translate(${p.x - 7}, ${p.y - 7})`} className="pointer-events-none">
                      {domain === 'light' && (
                        <path
                          d="M9 18h6m-4 4h2M12 2a7 7 0 0 0-7 7c0 2.5 1.5 4.5 3 6h8c1.5-1.5 3-3.5 3-6a7 7 0 0 0-7-7z"
                          fill="none"
                          stroke={colors.fill}
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          transform="scale(0.58)"
                        />
                      )}
                      {domain === 'switch' && (
                        <path
                          d="M12 2v10m-7.5-6a9 9 0 1 0 15 0"
                          fill="none"
                          stroke={colors.fill}
                          strokeWidth="2.2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          transform="scale(0.58)"
                        />
                      )}
                      {domain === 'sensor' && (
                        <path
                          d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z"
                          fill="none"
                          stroke={colors.fill}
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          transform="scale(0.58)"
                        />
                      )}
                      {domain === 'climate' && (
                        <path
                          d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"
                          fill="none"
                          stroke={colors.fill}
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          transform="scale(0.58)"
                        />
                      )}
                      {domain === 'media_player' && (
                        <path
                          d="M9 18V5l12-2v13M9 18a3 3 0 1 1-6 0 3 3 0 0 1 6 0zm12-2a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"
                          fill="none"
                          stroke={colors.fill}
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          transform="scale(0.58)"
                        />
                      )}
                    </g>

                    {/* State dot indicator */}
                    <circle
                      cx={p.x + 11}
                      cy={p.y - 11}
                      r="3.5"
                      fill={stateLabel === 'on' || (stateLabel !== 'off' && stateLabel !== '' && stateLabel !== 'idle') ? '#22c55e' : '#64748b'}
                      stroke="#0f172a"
                      strokeWidth="1.5"
                    />

                    {/* Device name label */}
                    <text
                      x={p.x}
                      y={p.y + 26}
                      textAnchor="middle"
                      fill="#e2e8f0"
                      fontSize="10"
                      fontWeight="600"
                      className="pointer-events-none select-none drop-shadow"
                    >
                      {displayName}
                    </text>

                    {/* State pill if sensor or value */}
                    {stateLabel && (
                      <text
                        x={p.x}
                        y={p.y + 36}
                        textAnchor="middle"
                        fill="#94a3b8"
                        fontSize="8.5"
                        className="pointer-events-none select-none font-mono"
                      >
                        {stateLabel}
                      </text>
                    )}
                  </g>
                )
              })}

              {/* Cursor Grid Snapping Indicator */}
              {cursorPos && (
                <g className="pointer-events-none">
                  <circle
                    cx={cursorPos.x}
                    cy={cursorPos.y}
                    r="3.5"
                    fill="#ffffff"
                    stroke="#6366f1"
                    strokeWidth="1.5"
                  />
                </g>
              )}
            </svg>
          )}
        </div>

        {/* Collapsible HA Devices Drawer on the Right */}
        {showDevicePalette && (
          <div className="w-80 bg-slate-900 border-l border-slate-800 flex flex-col h-full z-10 shadow-xl select-none">
            {/* Drawer Header */}
            <div className="p-3 border-b border-slate-800 flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Cpu className="w-4 h-4 text-indigo-400" />
                <h3 className="text-xs font-bold text-white uppercase tracking-wider">
                  Appareils Home Assistant
                </h3>
              </div>
              <button
                type="button"
                onClick={() => setShowDevicePalette(false)}
                className="text-slate-400 hover:text-white p-1 rounded hover:bg-slate-800"
                title="Masquer la palette"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Category tabs */}
            <div className="p-2 border-b border-slate-800 bg-slate-950/40">
              <div className="flex flex-wrap gap-1">
                {DOMAIN_CATEGORIES.map((cat) => (
                  <button
                    key={cat.key}
                    type="button"
                    onClick={() => setDeviceCategory(cat.key)}
                    className={`px-2 py-1 rounded text-[11px] font-medium transition-colors ${
                      deviceCategory === cat.key
                        ? 'bg-indigo-600 text-white'
                        : 'bg-slate-800 text-slate-400 hover:text-slate-200 hover:bg-slate-700'
                    }`}
                  >
                    {cat.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Search Input */}
            <div className="p-2 border-b border-slate-800">
              <div className="relative">
                <Search className="w-3.5 h-3.5 absolute left-2.5 top-2 text-slate-500" />
                <input
                  type="text"
                  placeholder="Rechercher un appareil..."
                  value={deviceSearch}
                  onChange={(e) => setDeviceSearch(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 text-slate-200 text-xs rounded-md pl-8 pr-2 py-1.5 focus:outline-none focus:border-indigo-500"
                />
                {deviceSearch && (
                  <button
                    type="button"
                    onClick={() => setDeviceSearch('')}
                    className="absolute right-2 top-2 text-slate-400 hover:text-white"
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </div>
            </div>

            {/* Device list */}
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
              {loadingDevices ? (
                <div className="p-4 text-center text-xs text-slate-500">
                  Chargement des appareils...
                </div>
              ) : filteredDevices.length === 0 ? (
                <div className="p-6 text-center text-xs text-slate-500">
                  Aucun appareil trouvé
                </div>
              ) : (
                filteredDevices.map((dev) => {
                  const colors = getDomainColor(dev.domain)
                  const isPlaced = placedDeviceIds.has(dev.id)
                  const isCurrentlyPlacing = deviceToPlace?.id === dev.id

                  return (
                    <div
                      key={dev.id}
                      draggable
                      onDragStart={(e) => {
                        e.dataTransfer.setData('text/plain', dev.id)
                        e.dataTransfer.effectAllowed = 'copy'
                      }}
                      className={`p-2.5 rounded-lg border transition-all cursor-grab active:cursor-grabbing bg-slate-850 hover:bg-slate-800 ${
                        isCurrentlyPlacing
                          ? 'border-indigo-500 ring-1 ring-indigo-500 bg-indigo-950/30'
                          : 'border-slate-800 hover:border-slate-700'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-start space-x-2 min-w-0">
                          <div className={`p-1.5 rounded-md border ${colors.bg} ${colors.border} ${colors.text} shrink-0`}>
                            {getDomainIcon(dev.domain, 'w-3.5 h-3.5')}
                          </div>
                          <div className="min-w-0">
                            <h4 className="text-xs font-semibold text-slate-200 truncate" title={dev.name}>
                              {dev.name}
                            </h4>
                            <p className="text-[10px] text-slate-500 font-mono truncate" title={dev.id}>
                              {dev.id}
                            </p>
                          </div>
                        </div>

                        {/* State badge */}
                        <span
                          className={`text-[10px] px-1.5 py-0.5 rounded font-mono shrink-0 ${
                            dev.state === 'on' || (dev.state !== 'off' && dev.state !== '' && dev.state !== 'idle')
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-slate-800 text-slate-400'
                          }`}
                        >
                          {dev.state || 'off'}
                        </span>
                      </div>

                      {/* Actions & Placement badge */}
                      <div className="mt-2 pt-2 border-t border-slate-800/60 flex items-center justify-between text-[11px]">
                        <span className="flex items-center text-slate-500 text-[10px]">
                          <GripVertical className="w-3 h-3 mr-0.5" />
                          Glisser sur le plan
                        </span>

                        <div className="flex items-center space-x-1.5">
                          {isPlaced && (
                            <span className="text-[10px] text-emerald-400 bg-emerald-500/10 px-1.5 py-0.5 rounded border border-emerald-500/30 flex items-center space-x-0.5">
                              <Check className="w-2.5 h-2.5" />
                              <span>Placé</span>
                            </span>
                          )}

                          <button
                            type="button"
                            onClick={() => {
                              if (isCurrentlyPlacing) {
                                setDeviceToPlace(null)
                              } else {
                                setDeviceToPlace(dev)
                              }
                            }}
                            className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors flex items-center space-x-1 ${
                              isCurrentlyPlacing
                                ? 'bg-rose-600 text-white hover:bg-rose-500'
                                : 'bg-indigo-600/80 hover:bg-indigo-600 text-white'
                            }`}
                          >
                            {isCurrentlyPlacing ? (
                              <span>Annuler</span>
                            ) : (
                              <>
                                <Plus className="w-3 h-3" />
                                <span>Placer</span>
                              </>
                            )}
                          </button>
                        </div>
                      </div>
                    </div>
                  )
                })
              )}
            </div>

            {/* Drawer Footer Guide */}
            <div className="p-3 border-t border-slate-800 bg-slate-950/60 text-[11px] text-slate-500">
              <p>
                💡 <strong>Glissez-déposez</strong> un appareil sur le plan, ou cliquez sur <strong>Placer</strong> puis sur la zone désirée.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
