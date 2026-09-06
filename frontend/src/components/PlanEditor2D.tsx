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
} from 'lucide-react'
import type { Level, Plan, WallSegment, Zone, Point2D } from '../types'

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

let idCounter = 0
function generateId(prefix: string): string {
  idCounter += 1
  return `${prefix}-${Date.now()}-${idCounter}`
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
    type: 'wall' | 'zone'
    id: string
  } | null>(null)

  const svgRef = useRef<SVGSVGElement | null>(null)

  // Load plan when level changes
  useEffect(() => {
    if (!level) return
    let ignore = false

    fetch(`/api/levels/${level.id}/plan`)
      .then(async (res) => {
        if (!res.ok) {
          // If 404 or missing, initialize blank plan
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
        if (!ignore) {
          setLoading(false)
        }
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
  const getSvgCoordinates = (e: React.MouseEvent<SVGSVGElement> | React.TouchEvent<SVGSVGElement>): Point2D | null => {
    if (!svgRef.current) return null
    const svg = svgRef.current
    const ctm = svg.getScreenCTM()
    if (!ctm) return null

    let clientX: number
    let clientY: number

    if ('touches' in e && e.touches.length > 0) {
      clientX = e.touches[0].clientX
      clientY = e.touches[0].clientY
    } else if ('clientX' in e) {
      clientX = e.clientX
      clientY = e.clientY
    } else {
      return null
    }

    const inverse = ctm.inverse()
    return {
      x: inverse.a * clientX + inverse.c * clientY + inverse.e,
      y: inverse.b * clientX + inverse.d * clientY + inverse.f,
    }
  }

  // Handle pointer movements
  const handlePointerMove = (e: React.MouseEvent<SVGSVGElement> | React.TouchEvent<SVGSVGElement>) => {
    const raw = getSvgCoordinates(e)
    if (!raw) return
    const snapped = snap(raw)
    setCursorPos(snapped)
  }

  // Handle SVG canvas clicks / taps
  const handleSvgClick = (e: React.MouseEvent<SVGSVGElement>) => {
    const raw = getSvgCoordinates(e)
    if (!raw) return
    const point = snap(raw)

    if (tool === 'wall') {
      if (!wallStart) {
        // Start first point of the wall
        setWallStart(point)
      } else {
        // Only create wall if it has non-zero length
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
      // Check if clicking near starting point to auto-close polygon
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
      // Clicking on empty space deselects
      if (e.target === svgRef.current || (e.target as HTMLElement).tagName === 'svg') {
        setSelectedElement(null)
      }
    }
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

  // Delete selected item
  const handleDeleteSelected = () => {
    if (!selectedElement) return
    pushHistory(plan)
    if (selectedElement.type === 'wall') {
      setPlan((prev) => ({
        ...prev,
        walls: prev.walls.filter((w) => w.id !== selectedElement.id),
      }))
    } else {
      setPlan((prev) => ({
        ...prev,
        zones: prev.zones.filter((z) => z.id !== selectedElement.id),
      }))
    }
    setSelectedElement(null)
  }

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

  // Save plan to backend
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
          Sélectionnez un niveau dans la barre latérale ou créez-en un nouveau pour commencer à dessiner le plan 2D.
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
        {tool === 'wall' && (
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

        {tool === 'zone' && (
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

            <span className="text-slate-400">
              {currentZonePoints.length} points placés
            </span>

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

        {tool === 'select' && (
          <div className="flex items-center space-x-4">
            <span className="font-semibold text-indigo-400">Mode Sélection / Gomme:</span>
            {selectedElement ? (
              <div className="flex items-center space-x-2">
                <span className="text-white">
                  Élément sélectionné: {selectedElement.type === 'wall' ? 'Segment de Mur' : 'Zone'}
                </span>
                <button
                  type="button"
                  onClick={handleDeleteSelected}
                  className="bg-rose-600 hover:bg-rose-500 text-white px-2.5 py-0.5 rounded text-xs flex items-center space-x-1"
                >
                  <Trash2 className="w-3 h-3" />
                  <span>Supprimer</span>
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedElement(null)}
                  className="text-slate-400 underline text-xs"
                >
                  Désélectionner
                </button>
              </div>
            ) : (
              <span className="text-slate-400">
                Cliquez sur un mur ou une zone pour le sélectionner et le supprimer.
              </span>
            )}
          </div>
        )}

        {/* Summary info right */}
        <div className="ml-auto text-[11px] text-slate-400 flex items-center space-x-3">
          <span>
            {plan.walls.length} murs • {plan.zones.length} zones
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
        </div>
      )}

      {/* SVG Interactive Canvas */}
      <div className="flex-1 relative overflow-auto bg-slate-950 flex items-center justify-center p-4">
        {loading ? (
          <div className="text-slate-400 text-sm">Chargement du plan...</div>
        ) : (
          <svg
            ref={svgRef}
            viewBox="0 0 1000 700"
            className="w-full h-full max-w-[1000px] max-h-[700px] bg-slate-900/90 rounded-xl border border-slate-800 shadow-2xl shadow-black/60 cursor-crosshair touch-none"
            onMouseMove={handlePointerMove}
            onTouchMove={handlePointerMove}
            onClick={handleSvgClick}
          >
            <defs>{gridPattern}</defs>

            {/* Grid layer */}
            <rect width="1000" height="700" fill="url(#editorGrid)" />

            {/* Render Saved Zones (Polygons) */}
            {plan.zones.map((zone) => {
              const isSelected = selectedElement?.type === 'zone' && selectedElement.id === zone.id
              const pointsStr = zone.points.map((p) => `${p.x},${p.y}`).join(' ')

              // Compute centroid for label
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
                {/* Rubber band line from last point to cursor */}
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
                {/* Zone vertices dots */}
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
                  {/* Invisible fat stroke for easy clicking / selecting */}
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
                  {/* Visual Wall */}
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
                  {/* Wall Caps */}
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
    </div>
  )
}
