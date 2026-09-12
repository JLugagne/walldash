import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { AlertCircle, Cpu, Layers, Loader2, Maximize2, SlidersHorizontal, X, ZoomIn, ZoomOut } from 'lucide-react'
import type { Device, DevicePlacement, Level, Plan, Point2D, SavePlacementRequest, WallOpening, WallSegment, Zone } from '../types'
import { apiFetch } from '../api'
import { ImportPlanModal } from './ImportPlanModal'
import { OnboardingWizard } from './OnboardingWizard'
import { LevelsManager } from './LevelsManager'
import { EditorHeader, type SaveState } from './editor/EditorHeader'
import { ToolRail } from './editor/ToolRail'
import { Inspector, type ToolSettings } from './editor/Inspector'
import { DevicePalette } from './editor/DevicePalette'
import { DimensionLabel, WallLayer } from './editor/canvas/WallLayer'
import { ZoneDraft, ZoneLayer } from './editor/canvas/ZoneLayer'
import { DeviceLayer } from './editor/canvas/DeviceLayer'
import { LayerPanel } from './editor/LayerPanel'
import { AlignmentLayer, FloorAlignmentPanel } from './editor/FloorAlignment'
import { CANVAS, GRID_SIZES, TOOLS, type ToolMode } from './editor/constants'
import * as geo from './editor/geometry'
import type { EditorSelection, OpeningDragMode } from './editor/types'
import { configForLevel, DEFAULT_FLOOR_CONFIG, orderedLevels, type HouseOverviewConfig } from '../utils/houseOverview'

interface PlanEditor2DProps {
  level: Level | null
  levels: Level[]
  onSelectLevel: (id: string) => void
  onRefreshLevels: () => Promise<void>
  alignMode?: boolean
  onToggleAlignMode?: () => void
  overviewPlans?: Record<string, Plan>
  overviewConfig?: HouseOverviewConfig
  onOverviewConfigChange?: (config: HouseOverviewConfig) => void
}

interface ViewBox {
  x: number
  y: number
  w: number
  h: number
}

const DEFAULT_VIEW: ViewBox = { x: 0, y: 0, w: 1000, h: 700 }
const JOINT_TOLERANCE = 4
const VERTEX_SNAP_PX = 10
const OPENING_HOVER_PX = 22
const DRAG_THRESHOLD_PX = 3
const HISTORY_LIMIT = 50
const MIN_VIEW_W = 80
const MAX_VIEW_W = 30000

type DragState =
  | { kind: 'pan'; startClient: Point2D; startView: ViewBox; upx: number }
  | { kind: 'press'; startClient: Point2D; startView: ViewBox; upx: number; pan: boolean; moved: boolean }
  | {
      kind: 'wall-move'
      wallId: string
      refs: geo.VertexRef[]
      selfRefs: geo.VertexRef[]
      startPlan: Plan
      startPt: Point2D
      startClient: Point2D
      moved: boolean
    }
  | { kind: 'vertex'; refs: geo.VertexRef[]; selfRef: geo.VertexRef; startPlan: Plan; startClient: Point2D; moved: boolean }
  | { kind: 'opening'; wallId: string; openingId: string; mode: OpeningDragMode; startPlan: Plan; startClient: Point2D; moved: boolean }
  | { kind: 'zone-move'; zoneId: string; startPlan: Plan; startPt: Point2D; startClient: Point2D; moved: boolean }
  | { kind: 'zone-label'; zoneId: string; startPlan: Plan; startClient: Point2D; moved: boolean }
  | { kind: 'zone-vertex'; zoneId: string; index: number; startPlan: Plan; startClient: Point2D; moved: boolean }
  | { kind: 'device'; placementId: string; startPlacements: DevicePlacement[]; grabOffset: Point2D; startClient: Point2D; moved: boolean }
  | { kind: 'floor'; levelId: string; startConfig: HouseOverviewConfig; startPt: Point2D; startClient: Point2D; moved: boolean }

function emptyPlan(levelId: string): Plan {
  return { level_id: levelId, walls: [], zones: [] }
}

function serializePlan(plan: Plan): string {
  return JSON.stringify({ walls: plan.walls, zones: plan.zones })
}

function clonePlan(plan: Plan): Plan {
  return JSON.parse(JSON.stringify(plan))
}

export function PlanEditor2D({ level, levels, onSelectLevel, onRefreshLevels, alignMode = false, onToggleAlignMode, overviewPlans = {}, overviewConfig = {}, onOverviewConfigChange }: PlanEditor2DProps) {
  const [plan, setPlan] = useState<Plan>(() => emptyPlan(level?.id ?? ''))
  const [savedJson, setSavedJson] = useState(() => serializePlan(emptyPlan('')))
  const [history, setHistory] = useState<Plan[]>([])
  const [redoStack, setRedoStack] = useState<Plan[]>([])
  const lastHistoryRef = useRef<{ key: string; time: number } | null>(null)

  const [loading, setLoading] = useState(false)
  const [saveState, setSaveState] = useState<SaveState>('idle')
  const [error, setError] = useState<string | null>(null)
  const [showImport, setShowImport] = useState(false)
  const [showWizard, setShowWizard] = useState(false)
  const [levelsOpen, setLevelsOpen] = useState(false)
  const [layersOpen, setLayersOpen] = useState(true)

  const [devices, setDevices] = useState<Device[]>([])
  const [placements, setPlacements] = useState<DevicePlacement[]>([])
  const [loadingDevices, setLoadingDevices] = useState(false)
  const [deviceToPlace, setDeviceToPlace] = useState<Device | null>(null)
  const [activeLayer, setActiveLayer] = useState<string>('controls')

  const [tool, setToolState] = useState<ToolMode>('select')
  const [alignSelectedId, setAlignSelectedId] = useState<string | null>(null)
  const [snapGrid, setSnapGrid] = useState(true)
  const [gridSize, setGridSize] = useState(20)
  const [settings, setSettings] = useState<ToolSettings>({
    wallThickness: 8,
    doorWidth: 36,
    windowWidth: 48,
    doorAsPassage: false,
    zoneName: 'Living room',
    zoneColor: '#3b82f6',
  })

  const [wallStart, setWallStart] = useState<Point2D | null>(null)
  const chainOriginRef = useRef<Point2D | null>(null)
  const [zoneDraft, setZoneDraft] = useState<Point2D[]>([])

  const [selection, setSelection] = useState<EditorSelection | null>(null)
  const [hoverWallId, setHoverWallId] = useState<string | null>(null)
  const [hoverProjection, setHoverProjection] = useState<geo.WallProjection | null>(null)
  const [cursorRaw, setCursorRaw] = useState<Point2D | null>(null)
  const [cursorSnap, setCursorSnap] = useState<{ point: Point2D; vertex: boolean } | null>(null)

  const [panelOpen, setPanelOpen] = useState(true)
  const [panelTab, setPanelTab] = useState<'inspector' | 'devices'>('inspector')

  const svgRef = useRef<SVGSVGElement | null>(null)
  const [viewBox, setViewBox] = useState<ViewBox>(DEFAULT_VIEW)
  const [svgSize, setSvgSize] = useState({ width: 1000, height: 700 })
  const upx = viewBox.w / svgSize.width

  const dragRef = useRef<DragState | null>(null)
  const [dragKind, setDragKind] = useState<DragState['kind'] | null>(null)
  const spaceRef = useRef(false)
  const [spaceDown, setSpaceDown] = useState(false)
  const pinchRef = useRef<{ dist: number; center: Point2D; view: ViewBox } | null>(null)

  const isDirty = useMemo(() => serializePlan(plan) !== savedJson, [plan, savedJson])
  const placedDeviceIds = useMemo(() => new Set(placements.map((p) => p.device_id)), [placements])
  const visiblePlacements = useMemo(
    () => placements.filter((p) => (p.layer || 'controls') === activeLayer),
    [placements, activeLayer]
  )

  const fetchDevices = useCallback(async () => {
    setLoadingDevices(true)
    try {
      const res = await fetch('/api/devices')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) setDevices(payload.data)
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

  const refreshPlacements = useCallback(async () => {
    if (!level?.id) return
    try {
      const res = await fetch(`/api/levels/${level.id}/placements`)
      const data = await res.json()
      if (data?.data) setPlacements(data.data)
    } catch {
      // ignore
    }
  }, [level?.id])

  useEffect(() => {
    const el = svgRef.current
    if (!el) return
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const width = Math.max(1, entry.contentRect.width)
        const height = Math.max(1, entry.contentRect.height)
        setSvgSize((prev) =>
          prev.width === width && prev.height === height ? prev : { width, height },
        )
      }
    })
    ro.observe(el)
    return () => ro.disconnect()
  }, [])

  // Lock the viewBox aspect ratio to the rendered SVG (before paint) so the canvas and its grid
  // can never stretch: grid cells stay square whatever the window or side-panel geometry.
  useLayoutEffect(() => {
    setViewBox((prev) => {
      const targetH = (prev.w * svgSize.height) / svgSize.width
      return Math.abs(targetH - prev.h) < 0.5 ? prev : { ...prev, h: targetH }
    })
  }, [svgSize.width, svgSize.height, viewBox.w])

  const fitToBounds = useCallback(
    (walls: WallSegment[], zones: Zone[], plcs: DevicePlacement[]) => {
      const b = geo.planBounds(walls, zones, plcs)
      const aspect = svgSize.width / svgSize.height
      if (!b) {
        setViewBox({ x: 0, y: 0, w: DEFAULT_VIEW.w, h: DEFAULT_VIEW.w / aspect })
        return
      }
      const padding = 80
      const spanW = Math.max(b.maxX - b.minX + padding * 2, 300)
      const spanH = Math.max(b.maxY - b.minY + padding * 2, 200)
      let w = spanW
      let h = spanH
      if (spanW / spanH > aspect) h = spanW / aspect
      else w = spanH * aspect
      setViewBox({ x: (b.minX + b.maxX) / 2 - w / 2, y: (b.minY + b.maxY) / 2 - h / 2, w, h })
    },
    [svgSize.width, svgSize.height]
  )

  const fitRef = useRef(fitToBounds)
  useEffect(() => {
    fitRef.current = fitToBounds
  }, [fitToBounds])

  const prevAlignRef = useRef(false)
  const alignFittedRef = useRef(false)
  const alignmentWalls = useCallback(() => {
    const combined: WallSegment[] = []
    for (const lvl of orderedLevels(levels)) {
      const floor = configForLevel(overviewConfig, lvl.id)
      if (!floor.visible) continue
      for (const wall of overviewPlans[lvl.id]?.walls || []) {
        combined.push({ ...wall, x1: wall.x1 + floor.x, y1: wall.y1 + floor.y, x2: wall.x2 + floor.x, y2: wall.y2 + floor.y })
      }
    }
    return combined
  }, [levels, overviewConfig, overviewPlans])

  useEffect(() => {
    if (alignMode && !prevAlignRef.current) alignFittedRef.current = false
    prevAlignRef.current = alignMode
    if (!alignMode) return
    setAlignSelectedId((prev) => prev ?? level?.id ?? levels[0]?.id ?? null)
    if (alignFittedRef.current) return
    const combined = alignmentWalls()
    if (combined.length === 0 && levels.length > 0) return
    alignFittedRef.current = true
    fitToBounds(combined, [], [])
  }, [alignMode, alignmentWalls, fitToBounds, level?.id, levels.length])

  const resetAlignment = useCallback(() => {
    onOverviewConfigChange?.(Object.fromEntries(levels.map((lvl) => [lvl.id, { ...DEFAULT_FLOOR_CONFIG }])))
  }, [levels, onOverviewConfigChange])

  const resetEditing = useCallback(() => {
    setSelection(null)
    setDeviceToPlace(null)
    setWallStart(null)
    chainOriginRef.current = null
    setZoneDraft([])
    setHoverProjection(null)
    setHoverWallId(null)
    dragRef.current = null
    setDragKind(null)
  }, [])

  const levelId = level?.id ?? null
  useEffect(() => {
    resetEditing()
    setHistory([])
    setRedoStack([])
    setError(null)
    setSaveState('idle')
    if (!levelId) {
      const empty = emptyPlan('')
      setPlan(empty)
      setSavedJson(serializePlan(empty))
      setPlacements([])
      return
    }
    let ignore = false
    setLoading(true)
    Promise.all([
      fetch(`/api/levels/${levelId}/plan`)
        .then(async (res) => (res.ok ? res.json() : { status: 'success', data: { walls: [], zones: [] } }))
        .catch(() => null),
      fetch(`/api/levels/${levelId}/placements`)
        .then((res) => res.json())
        .catch(() => null),
    ])
      .then(([planPayload, plcPayload]) => {
        if (ignore) return
        const loaded: Plan = {
          level_id: levelId,
          walls: planPayload?.data?.walls || [],
          zones: planPayload?.data?.zones || [],
        }
        const plcs: DevicePlacement[] = plcPayload?.data || []
        setPlan(loaded)
        setSavedJson(serializePlan(loaded))
        setPlacements(plcs)
        fitRef.current(loaded.walls, loaded.zones, plcs)
      })
      .finally(() => {
        if (!ignore) setLoading(false)
      })
    return () => {
      ignore = true
    }
  }, [levelId, resetEditing])

  // Keep activeLayer in sync with the level's available layers.
  useEffect(() => {
    if (!level) return
    const available = level.layers && level.layers.length > 0 ? level.layers.map((l) => l.name) : ['controls', 'sensors']
    setActiveLayer((prev) => (available.includes(prev) ? prev : available[0]))
  }, [level])

  // Clear device selection when switching layers (device becomes invisible on canvas).
  useEffect(() => {
    setSelection((prev) => (prev?.type === 'device' ? null : prev))
  }, [activeLayer])

  useEffect(() => {
    if (!isDirty) return
    const onBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener('beforeunload', onBeforeUnload)
    return () => window.removeEventListener('beforeunload', onBeforeUnload)
  }, [isDirty])

  const selectElement = useCallback((next: EditorSelection | null) => {
    setSelection(next)
    if (next) setPanelTab('inspector')
  }, [])

  const pushHistory = useCallback((snapshot: Plan, coalesceKey?: string) => {
    const now = Date.now()
    const last = lastHistoryRef.current
    if (coalesceKey && last && last.key === coalesceKey && now - last.time < 1200) {
      lastHistoryRef.current = { key: coalesceKey, time: now }
      return
    }
    lastHistoryRef.current = { key: coalesceKey ?? '', time: now }
    setHistory((prev) => [...prev.slice(-(HISTORY_LIMIT - 1)), clonePlan(snapshot)])
    setRedoStack([])
  }, [])

  const mutatePlan = useCallback(
    (updater: (current: Plan) => Plan, coalesceKey?: string) => {
      pushHistory(plan, coalesceKey)
      setPlan(updater(plan))
    },
    [plan, pushHistory]
  )

  const undo = useCallback(() => {
    if (history.length === 0) return
    const previous = history[history.length - 1]
    setHistory((prev) => prev.slice(0, -1))
    setRedoStack((prev) => [...prev.slice(-(HISTORY_LIMIT - 1)), clonePlan(plan)])
    setPlan(previous)
    lastHistoryRef.current = null
    setSelection(null)
    setZoneDraft([])
    setWallStart(null)
  }, [history, plan])

  const redo = useCallback(() => {
    if (redoStack.length === 0) return
    const next = redoStack[redoStack.length - 1]
    setRedoStack((prev) => prev.slice(0, -1))
    setHistory((prev) => [...prev.slice(-(HISTORY_LIMIT - 1)), clonePlan(plan)])
    setPlan(next)
    lastHistoryRef.current = null
    setSelection(null)
  }, [redoStack, plan])

  const setTool = useCallback(
    (next: ToolMode) => {
      setToolState(next)
      setWallStart(null)
      chainOriginRef.current = null
      setZoneDraft([])
      setDeviceToPlace(null)
      setHoverProjection(null)
      if (next !== 'select') setSelection(null)
    },
    []
  )

  const toSvgPoint = useCallback((clientX: number, clientY: number): Point2D | null => {
    const svg = svgRef.current
    if (!svg) return null
    const ctm = svg.getScreenCTM()
    if (!ctm) return null
    const inv = ctm.inverse()
    return { x: inv.a * clientX + inv.c * clientY + inv.e, y: inv.b * clientX + inv.d * clientY + inv.f }
  }, [])

  const snapPoint = useCallback(
    (
      raw: Point2D,
      opts: { walls?: WallSegment[]; skip?: (ref: geo.VertexRef) => boolean; from?: Point2D | null; axis?: boolean; noVertex?: boolean } = {}
    ): { point: Point2D; vertex: boolean } => {
      let pt = raw
      if (opts.axis && opts.from) pt = geo.axisLock(opts.from, pt)
      if (!opts.noVertex) {
        const v = geo.nearestVertex(opts.walls ?? plan.walls, pt, VERTEX_SNAP_PX * upx, opts.skip)
        if (v) return { point: v, vertex: true }
      }
      if (snapGrid) pt = geo.snapToGrid(pt, gridSize)
      return { point: { x: Math.round(pt.x), y: Math.round(pt.y) }, vertex: false }
    },
    [plan.walls, snapGrid, gridSize, upx]
  )

  const savePlacement = useCallback(
    async (req: SavePlacementRequest): Promise<DevicePlacement | null> => {
      if (!level) return null
      setError(null)
      try {
        const res = await apiFetch(`/api/levels/${level.id}/placements`, {
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
        }
        setError(payload.error?.message || 'Error while placing the device')
      } catch {
        setError('Unable to reach the server')
      }
      return null
    },
    [level]
  )

  const deletePlacement = useCallback(
    async (placementId: string) => {
      if (!level) return
      setError(null)
      try {
        const res = await apiFetch(`/api/levels/${level.id}/placements/${placementId}`, { method: 'DELETE' })
        const payload = await res.json()
        if (res.ok && payload.status === 'success') {
          setPlacements((prev) => prev.filter((p) => p.id !== placementId))
          setSelection((prev) => (prev?.type === 'device' && prev.id === placementId ? null : prev))
        } else {
          setError(payload.error?.message || 'Error while deleting the placement')
        }
      } catch {
        setError('Impossible de joindre le serveur')
      }
    },
    [level]
  )

  const placementRequest = (p: DevicePlacement, patch: Partial<DevicePlacement> = {}): SavePlacementRequest => {
    const merged = { ...p, ...patch }
    return {
      id: merged.id,
      device_id: merged.device_id,
      x: merged.x,
      y: merged.y,
      icon: merged.icon,
      custom_name: merged.custom_name,
      render_domain: merged.render_domain,
      layer: merged.layer,
    }
  }

  const deleteSelection = useCallback(() => {
    if (!selection) return
    if (selection.type === 'device') {
      deletePlacement(selection.id)
      return
    }
    mutatePlan((p) => {
      if (selection.type === 'wall') return { ...p, walls: p.walls.filter((w) => w.id !== selection.id) }
      if (selection.type === 'zone') return { ...p, zones: p.zones.filter((z) => z.id !== selection.id) }
      return {
        ...p,
        walls: p.walls.map((w) =>
          w.id === selection.wallId ? { ...w, openings: (w.openings ?? []).filter((o) => o.id !== selection.id) } : w
        ),
      }
    })
    setSelection(null)
  }, [selection, deletePlacement, mutatePlan])

  const completeZone = useCallback(() => {
    if (zoneDraft.length < 3) return
    const zone: Zone = {
      id: geo.generateId('zone'),
      name: settings.zoneName.trim() || 'Zone',
      color: settings.zoneColor,
      points: zoneDraft,
    }
    mutatePlan((p) => ({ ...p, zones: [...p.zones, zone] }))
    setZoneDraft([])
  }, [zoneDraft, settings.zoneName, settings.zoneColor, mutatePlan])

  const endWallChain = useCallback(() => {
    setWallStart(null)
    chainOriginRef.current = null
  }, [])

  const savePlan = useCallback(async () => {
    if (!level || saveState === 'saving') return
    setSaveState('saving')
    setError(null)
    try {
      const res = await apiFetch(`/api/levels/${level.id}/plan`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ walls: plan.walls, zones: plan.zones }),
      })
      const result = await res.json()
      if (res.ok && result.status === 'success') {
        setSavedJson(serializePlan(plan))
        setSaveState('saved')
        setTimeout(() => setSaveState((s) => (s === 'saved' ? 'idle' : s)), 2500)
      } else {
        setSaveState('error')
        setError(result.error?.message || 'Error while saving the plan')
      }
    } catch {
      setSaveState('error')
      setError('Impossible de joindre le serveur')
    }
  }, [level, plan, saveState])

  const zoomBy = useCallback((factor: number, focus?: Point2D) => {
    setViewBox((prev) => {
      const w = prev.w / factor
      if (w < MIN_VIEW_W || w > MAX_VIEW_W) return prev
      const h = prev.h / factor
      const fx = focus?.x ?? prev.x + prev.w / 2
      const fy = focus?.y ?? prev.y + prev.h / 2
      return { x: fx - ((fx - prev.x) * w) / prev.w, y: fy - ((fy - prev.y) * h) / prev.h, w, h }
    })
  }, [])

  useEffect(() => {
    const svg = svgRef.current
    if (!svg) return
    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      const focus = toSvgPoint(e.clientX, e.clientY)
      zoomBy(e.deltaY < 0 ? 1.15 : 1 / 1.15, focus ?? undefined)
    }
    svg.addEventListener('wheel', onWheel, { passive: false })
    return () => svg.removeEventListener('wheel', onWheel)
  }, [toSvgPoint, zoomBy])

  const keyHandlerRef = useRef<(e: KeyboardEvent) => void>(() => {})
  const handleKeyDown = (e: KeyboardEvent) => {
    const target = e.target as HTMLElement | null
    const tag = target?.tagName?.toLowerCase()
    const typing = tag === 'input' || tag === 'textarea' || tag === 'select' || target?.isContentEditable
    const mod = e.ctrlKey || e.metaKey

    if (mod && (e.key === 'z' || e.key === 'Z') && !e.shiftKey) {
      if (typing) return
      e.preventDefault()
      undo()
      return
    }
    if (mod && (e.key === 'y' || e.key === 'Y' || ((e.key === 'z' || e.key === 'Z') && e.shiftKey))) {
      if (typing) return
      e.preventDefault()
      redo()
      return
    }
    if (mod && (e.key === 's' || e.key === 'S')) {
      e.preventDefault()
      if (isDirty) savePlan()
      return
    }
    if (typing) return

    if (alignMode) {
      if (e.key === 'Escape') onToggleAlignMode?.()
      return
    }

    if (e.code === 'Space') {
      e.preventDefault()
      if (!spaceRef.current) {
        spaceRef.current = true
        setSpaceDown(true)
      }
      return
    }
    if (e.key === 'Escape') {
      if (deviceToPlace) setDeviceToPlace(null)
      else if (wallStart) endWallChain()
      else if (zoneDraft.length > 0) setZoneDraft([])
      else if (selection) setSelection(null)
      dragRef.current = null
      setDragKind(null)
      return
    }
    if (e.key === 'Enter') {
      if (tool === 'zone' && zoneDraft.length >= 3) completeZone()
      else if (tool === 'wall' && wallStart) endWallChain()
      return
    }
    if (e.key === 'Delete' || e.key === 'Backspace') {
      if (selection) {
        e.preventDefault()
        deleteSelection()
      }
      return
    }
    if (mod || e.altKey) return
    // Digit keys 1-9 switch layers (like Photoshop)
    const digit = parseInt(e.key, 10)
    if (digit >= 1 && digit <= 9) {
      const available = level?.layers && level.layers.length > 0 ? level.layers.map((l) => l.name) : ['controls', 'sensors']
      const idx = digit - 1
      if (idx < available.length) {
        setActiveLayer(available[idx])
      }
      return
    }
    if (e.key === 'l' || e.key === 'L') {
      setLayersOpen((v) => !v)
      return
    }
    const key = e.key.toUpperCase()
    const toolDef = TOOLS.find((t) => t.shortcut === key)
    if (toolDef) {
      setTool(toolDef.key)
      return
    }
    if (key === 'G') setSnapGrid((v) => !v)
    if (e.key === '+' || e.key === '=') zoomBy(1.25)
    if (e.key === '-') zoomBy(1 / 1.25)
  }

  useEffect(() => {
    keyHandlerRef.current = handleKeyDown
  })

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => keyHandlerRef.current(e)
    const onKeyUp = (e: KeyboardEvent) => {
      if (e.code === 'Space') {
        spaceRef.current = false
        setSpaceDown(false)
      }
    }
    window.addEventListener('keydown', onKeyDown)
    window.addEventListener('keyup', onKeyUp)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('keyup', onKeyUp)
    }
  }, [])

  const beginDrag = (state: DragState, e: React.PointerEvent) => {
    dragRef.current = state
    setDragKind(state.kind)
    try {
      svgRef.current?.setPointerCapture(e.pointerId)
    } catch {
      /* capture is best effort */
    }
  }

  const clientPoint = (e: React.PointerEvent): Point2D => ({ x: e.clientX, y: e.clientY })

  const handleFloorPointerDown = (levelId: string, e: React.PointerEvent<SVGGElement>) => {
    if (e.button !== 0) return
    e.stopPropagation()
    setAlignSelectedId(levelId)
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!raw) return
    beginDrag({ kind: 'floor', levelId, startConfig: overviewConfig, startPt: raw, startClient: clientPoint(e), moved: false }, e)
  }

  const handleBackgroundPointerDown = (e: React.PointerEvent<SVGSVGElement>) => {
    if (e.button === 1 || e.button === 2 || tool === 'pan' || spaceRef.current) {
      e.preventDefault()
      beginDrag({ kind: 'pan', startClient: clientPoint(e), startView: viewBox, upx }, e)
      return
    }
    if (e.button !== 0) return
    beginDrag(
      { kind: 'press', startClient: clientPoint(e), startView: viewBox, upx, pan: tool === 'select' && !deviceToPlace, moved: false },
      e
    )
  }

  const handleWallPointerDown = (wall: WallSegment, e: React.PointerEvent) => {
    if (e.button !== 0 || tool !== 'select') return
    if (spaceRef.current) {
      beginDrag({ kind: 'pan', startClient: clientPoint(e), startView: viewBox, upx }, e)
      return
    }
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!raw) return
    selectElement({ type: 'wall', id: wall.id })
    beginDrag(
      {
        kind: 'wall-move',
        wallId: wall.id,
        refs: geo.wallWithJoints(plan.walls, wall, JOINT_TOLERANCE),
        selfRefs: [
          { wallId: wall.id, end: 'start' },
          { wallId: wall.id, end: 'end' },
        ],
        startPlan: plan,
        startPt: raw,
        startClient: clientPoint(e),
        moved: false,
      },
      e
    )
  }

  const handleVertexPointerDown = (ref: geo.VertexRef, e: React.PointerEvent) => {
    if (e.button !== 0) return
    beginDrag(
      {
        kind: 'vertex',
        refs: geo.vertexWithJoints(plan.walls, ref, JOINT_TOLERANCE),
        selfRef: ref,
        startPlan: plan,
        startClient: clientPoint(e),
        moved: false,
      },
      e
    )
  }

  const handleOpeningPointerDown = (wall: WallSegment, opening: WallOpening, mode: OpeningDragMode, e: React.PointerEvent) => {
    if (e.button !== 0 || tool !== 'select') return
    selectElement({ type: 'opening', id: opening.id, wallId: wall.id })
    beginDrag({ kind: 'opening', wallId: wall.id, openingId: opening.id, mode, startPlan: plan, startClient: clientPoint(e), moved: false }, e)
  }

  const handleZonePointerDown = (zone: Zone, e: React.PointerEvent) => {
    if (e.button !== 0 || tool !== 'select') return
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!raw) return
    selectElement({ type: 'zone', id: zone.id })
    beginDrag({ kind: 'zone-move', zoneId: zone.id, startPlan: plan, startPt: raw, startClient: clientPoint(e), moved: false }, e)
  }

  const handleZoneVertexPointerDown = (zone: Zone, index: number, e: React.PointerEvent) => {
    if (e.button !== 0) return
    beginDrag({ kind: 'zone-vertex', zoneId: zone.id, index, startPlan: plan, startClient: clientPoint(e), moved: false }, e)
  }

  const handleZoneLabelPointerDown = (zone: Zone, e: React.PointerEvent) => {
    if (e.button !== 0 || tool !== 'select') return
    selectElement({ type: 'zone', id: zone.id })
    beginDrag({ kind: 'zone-label', zoneId: zone.id, startPlan: plan, startClient: clientPoint(e), moved: false }, e)
  }

  const handlePlacementPointerDown = (placement: DevicePlacement, e: React.PointerEvent) => {
    if (e.button !== 0 || deviceToPlace) return
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!raw) return
    selectElement({ type: 'device', id: placement.id })
    beginDrag(
      {
        kind: 'device',
        placementId: placement.id,
        startPlacements: placements,
        grabOffset: { x: raw.x - placement.x, y: raw.y - placement.y },
        startClient: clientPoint(e),
        moved: false,
      },
      e
    )
  }

  const panFrom = (start: ViewBox, startUpx: number, dxPx: number, dyPx: number) => {
    setViewBox({ ...start, x: start.x - dxPx * startUpx, y: start.y - dyPx * startUpx })
  }

  const handlePointerMove = (e: React.PointerEvent<SVGSVGElement>) => {
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!raw) return
    const drag = dragRef.current
    if (!drag) {
      updateHover(raw, e.shiftKey)
      return
    }

    const dxPx = e.clientX - drag.startClient.x
    const dyPx = e.clientY - drag.startClient.y
    if (drag.kind === 'pan') {
      panFrom(drag.startView, drag.upx, dxPx, dyPx)
      return
    }
    if (!drag.moved) {
      if (Math.hypot(dxPx, dyPx) < DRAG_THRESHOLD_PX) return
      drag.moved = true
      if ('startPlan' in drag) pushHistory(drag.startPlan)
    }

    switch (drag.kind) {
      case 'floor': {
        const floor = configForLevel(drag.startConfig, drag.levelId)
        onOverviewConfigChange?.({
          ...drag.startConfig,
          [drag.levelId]: {
            ...floor,
            x: Math.round(floor.x + raw.x - drag.startPt.x),
            y: Math.round(floor.y + raw.y - drag.startPt.y),
          },
        })
        return
      }
      case 'press':
        if (drag.pan) panFrom(drag.startView, drag.upx, dxPx, dyPx)
        return
      case 'wall-move': {
        const wall = drag.startPlan.walls.find((w) => w.id === drag.wallId)
        if (!wall) return
        let dx = raw.x - drag.startPt.x
        let dy = raw.y - drag.startPt.y
        if (snapGrid) {
          const a = geo.snapToGrid({ x: wall.x1 + dx, y: wall.y1 + dy }, gridSize)
          dx = a.x - wall.x1
          dy = a.y - wall.y1
        }
        const refs = e.altKey ? drag.selfRefs : drag.refs
        setPlan({ ...drag.startPlan, walls: geo.translateVertices(drag.startPlan.walls, refs, Math.round(dx), Math.round(dy)) })
        return
      }
      case 'vertex': {
        const refs = e.altKey ? [drag.selfRef] : drag.refs
        const { point } = snapPoint(raw, { walls: drag.startPlan.walls, skip: (r) => refs.some((x) => geo.sameRef(x, r)) })
        setPlan({ ...drag.startPlan, walls: geo.setVertices(drag.startPlan.walls, refs, point) })
        return
      }
      case 'opening': {
        const wall = drag.startPlan.walls.find((w) => w.id === drag.wallId)
        if (!wall) return
        const len = geo.wallLength(wall)
        const dist = geo.offsetAlongWall(wall, raw)
        setPlan({
          ...drag.startPlan,
          walls: drag.startPlan.walls.map((w) => {
            if (w.id !== drag.wallId) return w
            return {
              ...w,
              openings: (w.openings ?? []).map((op) => {
                if (op.id !== drag.openingId) return op
                if (drag.mode === 'move') return geo.clampOpening({ ...op, offset: dist }, len)
                const startEdge = op.offset - op.width / 2
                const endEdge = op.offset + op.width / 2
                if (drag.mode === 'resize-end') {
                  const newEnd = Math.max(startEdge + geo.MIN_OPENING_WIDTH, Math.min(len, dist))
                  return geo.clampOpening({ ...op, width: newEnd - startEdge, offset: (startEdge + newEnd) / 2 }, len)
                }
                const newStart = Math.min(endEdge - geo.MIN_OPENING_WIDTH, Math.max(0, dist))
                return geo.clampOpening({ ...op, width: endEdge - newStart, offset: (newStart + endEdge) / 2 }, len)
              }),
            }
          }),
        })
        return
      }
      case 'zone-move': {
        const zone = drag.startPlan.zones.find((z) => z.id === drag.zoneId)
        if (!zone || zone.points.length === 0) return
        let dx = raw.x - drag.startPt.x
        let dy = raw.y - drag.startPt.y
        if (snapGrid) {
          const a = geo.snapToGrid({ x: zone.points[0].x + dx, y: zone.points[0].y + dy }, gridSize)
          dx = a.x - zone.points[0].x
          dy = a.y - zone.points[0].y
        }
        setPlan({
          ...drag.startPlan,
          zones: drag.startPlan.zones.map((z) => (z.id === drag.zoneId ? { ...z, points: geo.translatePoints(z.points, Math.round(dx), Math.round(dy)) } : z)),
        })
        return
      }
      case 'zone-vertex': {
        const { point } = snapPoint(raw, { walls: drag.startPlan.walls })
        setPlan({
          ...drag.startPlan,
          zones: drag.startPlan.zones.map((z) =>
            z.id === drag.zoneId ? { ...z, points: z.points.map((p, i) => (i === drag.index ? point : p)) } : z
          ),
        })
        return
      }
      case 'zone-label': {
        setPlan({
          ...drag.startPlan,
          zones: drag.startPlan.zones.map((z) => (z.id === drag.zoneId ? { ...z, label_position: { x: Math.round(raw.x), y: Math.round(raw.y) } } : z)),
        })
        return
      }
      case 'device': {
        const { point } = snapPoint({ x: raw.x - drag.grabOffset.x, y: raw.y - drag.grabOffset.y }, { noVertex: true })
        setPlacements(drag.startPlacements.map((p) => (p.id === drag.placementId ? { ...p, x: point.x, y: point.y } : p)))
        return
      }
    }
  }

  const updateHover = (raw: Point2D, shift: boolean) => {
    setCursorRaw(raw)
    if (deviceToPlace) {
      setCursorSnap({ ...snapPoint(raw, { noVertex: true }) })
      return
    }
    switch (tool) {
      case 'wall':
        setCursorSnap(snapPoint(raw, { from: wallStart, axis: shift && !!wallStart }))
        break
      case 'zone':
        setCursorSnap(snapPoint(raw))
        break
      case 'door':
      case 'window':
        setHoverProjection(geo.projectOnNearestWall(plan.walls, raw, OPENING_HOVER_PX * upx))
        break
      default:
        if (cursorSnap) setCursorSnap(null)
    }
  }

  const handlePointerUp = (e: React.PointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current
    if (!drag) return
    dragRef.current = null
    setDragKind(null)
    try {
      svgRef.current?.releasePointerCapture(e.pointerId)
    } catch {
      /* already released */
    }
    if (drag.kind === 'pan') return
    if (drag.kind === 'press') {
      if (!drag.moved) {
        const raw = toSvgPoint(e.clientX, e.clientY)
        if (raw) handleTap(raw, e.shiftKey)
      }
      return
    }
    if (drag.kind === 'device' && drag.moved) {
      const moved = placements.find((p) => p.id === drag.placementId)
      if (moved) savePlacement(placementRequest(moved))
    }
  }

  const handlePointerCancel = () => {
    dragRef.current = null
    setDragKind(null)
  }

  const handleTap = (raw: Point2D, shift: boolean) => {
    if (deviceToPlace) {
      const { point } = snapPoint(raw, { noVertex: true })
      savePlacement({ device_id: deviceToPlace.id, x: point.x, y: point.y, custom_name: deviceToPlace.name, icon: deviceToPlace.domain, layer: activeLayer })
      setDeviceToPlace(null)
      return
    }
    switch (tool) {
      case 'select':
        setSelection(null)
        return
      case 'wall':
        placeWallPoint(raw, shift)
        return
      case 'zone':
        addZonePoint(raw)
        return
      case 'door':
      case 'window':
        placeOpening(raw)
        return
      default:
        return
    }
  }

  const placeWallPoint = (raw: Point2D, shift: boolean) => {
    const { point } = snapPoint(raw, { from: wallStart, axis: shift && !!wallStart })
    if (!wallStart) {
      setWallStart(point)
      chainOriginRef.current = point
      return
    }
    if (geo.samePoint(point, wallStart)) {
      endWallChain()
      return
    }
    if (Math.hypot(point.x - wallStart.x, point.y - wallStart.y) < geo.MIN_WALL_LENGTH) return
    const wall: WallSegment = {
      id: geo.generateId('wall'),
      x1: wallStart.x,
      y1: wallStart.y,
      x2: point.x,
      y2: point.y,
      thickness: settings.wallThickness,
      openings: [],
    }
    mutatePlan((p) => ({ ...p, walls: [...p.walls, wall] }))
    const origin = chainOriginRef.current
    if (origin && geo.samePoint(point, origin)) endWallChain()
    else setWallStart(point)
  }

  const addZonePoint = (raw: Point2D) => {
    const { point } = snapPoint(raw)
    if (zoneDraft.length >= 3 && geo.samePoint(point, zoneDraft[0], Math.max(gridSize * 0.6, 8 * upx))) {
      completeZone()
      return
    }
    if (zoneDraft.length > 0 && geo.samePoint(point, zoneDraft[zoneDraft.length - 1])) return
    setZoneDraft((prev) => [...prev, point])
  }

  const placeOpening = (raw: Point2D) => {
    const proj = hoverProjection ?? geo.projectOnNearestWall(plan.walls, raw, OPENING_HOVER_PX * upx)
    if (!proj) return
    const isDoor = tool === 'door'
    const opening = geo.clampOpening(
      {
        id: geo.generateId(tool),
        type: isDoor ? 'door' : 'window',
        offset: proj.offset,
        width: isDoor ? settings.doorWidth : settings.windowWidth,
        hide_door: isDoor && settings.doorAsPassage,
      },
      geo.wallLength(proj.wall)
    )
    mutatePlan((p) => ({
      ...p,
      walls: p.walls.map((w) => (w.id === proj.wall.id ? { ...w, openings: [...(w.openings ?? []), opening] } : w)),
    }))
  }

  const handleDoubleClick = () => {
    if (tool === 'wall') endWallChain()
    else if (tool === 'zone' && zoneDraft.length >= 3) completeZone()
  }

  const handleTouchStart = (e: React.TouchEvent<SVGSVGElement>) => {
    if (e.touches.length !== 2) return
    dragRef.current = null
    setDragKind(null)
    const [t1, t2] = [e.touches[0], e.touches[1]]
    pinchRef.current = {
      dist: Math.hypot(t2.clientX - t1.clientX, t2.clientY - t1.clientY),
      center: { x: (t1.clientX + t2.clientX) / 2, y: (t1.clientY + t2.clientY) / 2 },
      view: viewBox,
    }
  }

  const handleTouchMove = (e: React.TouchEvent<SVGSVGElement>) => {
    const pinch = pinchRef.current
    if (e.touches.length !== 2 || !pinch) return
    e.preventDefault()
    const [t1, t2] = [e.touches[0], e.touches[1]]
    const dist = Math.hypot(t2.clientX - t1.clientX, t2.clientY - t1.clientY)
    const center = { x: (t1.clientX + t2.clientX) / 2, y: (t1.clientY + t2.clientY) / 2 }
    const ratio = dist / Math.max(pinch.dist, 1)
    const startUpx = pinch.view.w / svgSize.width
    const w = pinch.view.w / ratio
    if (w < MIN_VIEW_W || w > MAX_VIEW_W) return
    const h = pinch.view.h / ratio
    const focus = toSvgPoint(pinch.center.x, pinch.center.y)
    if (!focus) return
    setViewBox({
      x: focus.x - ((focus.x - pinch.view.x) * w) / pinch.view.w - (center.x - pinch.center.x) * startUpx,
      y: focus.y - ((focus.y - pinch.view.y) * h) / pinch.view.h - (center.y - pinch.center.y) * startUpx,
      w,
      h,
    })
  }

  const handleTouchEnd = () => {
    pinchRef.current = null
  }

  const handleDrop = (e: React.DragEvent<SVGSVGElement>) => {
    e.preventDefault()
    const deviceId = e.dataTransfer.getData('text/plain')
    const dev = devices.find((d) => d.id === deviceId)
    const raw = toSvgPoint(e.clientX, e.clientY)
    if (!dev || !raw || !level) return
    const { point } = snapPoint(raw, { noVertex: true })
    savePlacement({ device_id: dev.id, x: point.x, y: point.y, custom_name: dev.name, icon: dev.domain, layer: activeLayer })
  }

  const handleSelectLevel = (id: string) => {
    if (id === level?.id) return
    if (isDirty && !confirm('Unsaved modifications. Switching levels will discard them. Continue?')) return
    onSelectLevel(id)
  }

  const handleReset = () => {
    if (!isDirty || !confirm('Revert to the last saved version of the plan?')) return
    const saved = JSON.parse(savedJson) as { walls: WallSegment[]; zones: Zone[] }
    pushHistory(plan)
    setPlan({ level_id: plan.level_id, walls: saved.walls, zones: saved.zones })
    resetEditing()
  }

  const handleClear = () => {
    if (plan.walls.length === 0 && plan.zones.length === 0) return
    if (!confirm('Delete all walls and all zones from this level?')) return
    mutatePlan((p) => ({ ...p, walls: [], zones: [] }))
    resetEditing()
  }

  const handleImport = (imported: Plan) => {
    mutatePlan(() => imported)
    resetEditing()
    fitToBounds(imported.walls, imported.zones, placements)
  }

  const handleExport = async () => {
    try {
      const res = await fetch('/api/export')
      if (!res.ok) throw new Error('Export failed')
      const payload = await res.json()
      const blob = new Blob([JSON.stringify(payload.data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'walldash-export.json'
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      // silently ignore
    }
  }

  const updateWall = (id: string, patch: Partial<WallSegment>) =>
    mutatePlan((p) => ({ ...p, walls: p.walls.map((w) => (w.id === id ? geo.withClampedOpenings({ ...w, ...patch }) : w)) }), `wall:${id}`)

  const setWallLengthById = (id: string, length: number) =>
    mutatePlan((p) => ({ ...p, walls: p.walls.map((w) => (w.id === id ? geo.setWallLength(w, length) : w)) }), `wall-len:${id}`)

  const splitWallById = (id: string) => {
    const wall = plan.walls.find((w) => w.id === id)
    if (!wall) return
    const [a, b] = geo.splitWall(wall, geo.generateId('wall'))
    mutatePlan((p) => ({ ...p, walls: p.walls.flatMap((w) => (w.id === id ? [a, b] : [w])) }))
  }

  const updateOpening = (wallId: string, openingId: string, patch: Partial<WallOpening>) =>
    mutatePlan(
      (p) => ({
        ...p,
        walls: p.walls.map((w) => {
          if (w.id !== wallId) return w
          const len = geo.wallLength(w)
          return { ...w, openings: (w.openings ?? []).map((o) => (o.id === openingId ? geo.clampOpening({ ...o, ...patch }, len) : o)) }
        }),
      }),
      `opening:${openingId}:${Object.keys(patch).join(',')}`
    )

  const updateZone = (id: string, patch: Partial<Zone>) =>
    mutatePlan((p) => ({ ...p, zones: p.zones.map((z) => (z.id === id ? { ...z, ...patch } : z)) }), `zone:${id}:${Object.keys(patch).join(',')}`)

  const updatePlacement = (placement: DevicePlacement, patch: Partial<Pick<DevicePlacement, 'custom_name' | 'render_domain' | 'layer'>>) => {
    savePlacement(placementRequest(placement, patch))
  }

  const cycleGrid = () => setGridSize((g) => GRID_SIZES[(GRID_SIZES.indexOf(g) + 1) % GRID_SIZES.length] ?? 20)

  const displayGrid = useMemo(() => {
    let step = gridSize
    while (step / upx < 14) step *= 2
    return step
  }, [gridSize, upx])

  const wallDraftPoint = tool === 'wall' && wallStart && cursorSnap ? cursorSnap.point : null
  const wallDraftLength = wallDraftPoint && wallStart ? Math.hypot(wallDraftPoint.x - wallStart.x, wallDraftPoint.y - wallStart.y) : 0

  const cursorClass =
    dragKind === 'pan' || (dragKind === 'press' && tool === 'select')
      ? 'cursor-grabbing'
      : tool === 'pan' || spaceDown
        ? 'cursor-grab'
        : deviceToPlace
          ? 'cursor-copy'
          : tool === 'select'
            ? 'cursor-default'
            : 'cursor-crosshair'

  const statusText = (() => {
    if (alignMode) return 'Drag a floor to align it · Drag the background to pan · Scroll to zoom'
    if (!level) return 'Select or create a level to get started.'
    if (deviceToPlace) return `Click to place "${deviceToPlace.name}" on layer "${activeLayer}" · Escape to cancel`
    switch (tool) {
      case 'wall':
        return wallStart
          ? `${geo.formatMeters(wallDraftLength)} · Click to place corner · Shift = orthogonal · Escape to finish`
          : 'Click to place the wall start point'
      case 'zone':
        return zoneDraft.length === 0
          ? 'Click to place the first zone vertex'
          : `${zoneDraft.length} vertex${zoneDraft.length > 1 ? 'es' : ''} · Click first point or Enter to close`
      case 'door':
      case 'window':
        return hoverProjection ? `Click to place the ${tool === 'door' ? 'door' : 'window'}` : 'Hover over a wall to place the opening'
      case 'pan':
        return 'Drag to move the view · Scroll wheel to zoom'
      default:
        if (selection) {
          const label = { wall: 'Wall', zone: 'Zone', device: 'Device', opening: 'Opening' }[selection.type]
          return `${label} selected · Drag to move · Delete to remove`
        }
        return 'Click an element to select it · Drag the background to move the view'
    }
  })()

  const interactive = tool === 'select' && !deviceToPlace

  return (
    <div className="flex-1 flex flex-col h-full min-h-0 bg-slate-950 overflow-hidden">
      <EditorHeader
        levels={levels}
        activeLevelId={level?.id ?? null}
        onSelectLevel={handleSelectLevel}
        levelsManager={
          <LevelsManager
            levels={levels}
            activeLevelId={level?.id ?? null}
            onSelectLevel={(id) => {
              handleSelectLevel(id)
              setLevelsOpen(false)
            }}
            onRefreshLevels={async () => {
              await onRefreshLevels()
              if (level?.id) {
                try {
                  const res = await fetch(`/api/levels/${level.id}/placements`)
                  const data = await res.json()
                  if (data?.data) setPlacements(data.data)
                } catch {
                  // ignore
                }
              }
            }}
            onOpenWizard={() => setShowWizard(true)}
          />
        }
        levelsOpen={levelsOpen}
        onToggleLevels={setLevelsOpen}
        canUndo={history.length > 0}
        canRedo={redoStack.length > 0}
        onUndo={undo}
        onRedo={redo}
        onImport={() => setShowImport(true)}
          onExport={handleExport}
        onReset={handleReset}
        onClear={handleClear}
        onSave={savePlan}
        isDirty={isDirty}
        saveState={saveState}
        panelOpen={panelOpen}
        onTogglePanel={() => setPanelOpen((v) => !v)}
        disabled={!level}
        alignmentActive={alignMode}
        onToggleAlignment={onToggleAlignMode}
      />

      {error && (
        <div className="bg-rose-500/10 border-b border-rose-500/20 px-4 py-1.5 text-xs text-rose-300 flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span className="flex-1">{error}</span>
          <button type="button" onClick={() => setError(null)} className="text-rose-300 hover:text-white cursor-pointer">
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      <div className="flex-1 flex min-h-0">
        {!alignMode && (
          <div className="flex">
            <ToolRail tool={tool} onSelectTool={setTool} snapGrid={snapGrid} onToggleSnap={() => setSnapGrid((v) => !v)} gridSize={gridSize} onCycleGrid={cycleGrid} layersOpen={layersOpen} onToggleLayers={() => setLayersOpen((v) => !v)} />
            {layersOpen && (
              <div className="w-52 shrink-0 bg-slate-900 border-r border-slate-800 flex flex-col min-h-0">
                <LayerPanel
                  level={level}
                  placements={placements}
                  activeLayer={activeLayer}
                  onSelectLayer={setActiveLayer}
                  onRefreshLevels={onRefreshLevels}
                  onRefreshPlacements={refreshPlacements}
                />
              </div>
            )}
          </div>
        )}

        <div className="flex-1 relative min-w-0 min-h-0 overflow-hidden" style={{ backgroundColor: CANVAS.background }}>
          <svg
            ref={svgRef}
            viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.w} ${viewBox.h}`}
            preserveAspectRatio="none"
            className={`absolute inset-0 w-full h-full touch-none select-none ${cursorClass}`}
            onPointerDown={handleBackgroundPointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
            onPointerCancel={handlePointerCancel}
            onPointerLeave={() => {
              if (!dragRef.current) {
                setCursorRaw(null)
                setCursorSnap(null)
                setHoverProjection(null)
              }
            }}
            onDoubleClick={handleDoubleClick}
            onContextMenu={(e) => e.preventDefault()}
            onTouchStart={handleTouchStart}
            onTouchMove={handleTouchMove}
            onTouchEnd={handleTouchEnd}
            onTouchCancel={handleTouchEnd}
            onDragOver={(e) => {
              e.preventDefault()
              e.dataTransfer.dropEffect = 'copy'
            }}
            onDrop={handleDrop}
          >
            <defs>
              <pattern id="editorGrid" width={displayGrid} height={displayGrid} patternUnits="userSpaceOnUse">
                <rect width={displayGrid} height={displayGrid} fill={CANVAS.background} />
                <path d={`M ${displayGrid} 0 L 0 0 0 ${displayGrid}`} fill="none" stroke={CANVAS.gridLine} strokeWidth={0.8 * upx} />
                <circle cx={0} cy={0} r={1.2 * upx} fill={CANVAS.gridDot} />
              </pattern>
              <filter id="editorGlow" x="-30%" y="-30%" width="160%" height="160%">
                <feDropShadow dx="0" dy="0" stdDeviation={3 * upx} floodColor={CANVAS.accent} floodOpacity="0.55" />
              </filter>
            </defs>

            <rect x={viewBox.x} y={viewBox.y} width={viewBox.w} height={viewBox.h} fill="url(#editorGrid)" />
            <line x1={viewBox.x} y1={0} x2={viewBox.x + viewBox.w} y2={0} stroke={CANVAS.axis} strokeWidth={upx} strokeDasharray={`${4 * upx} ${4 * upx}`} opacity={0.5} />
            <line x1={0} y1={viewBox.y} x2={0} y2={viewBox.y + viewBox.h} stroke={CANVAS.axis} strokeWidth={upx} strokeDasharray={`${4 * upx} ${4 * upx}`} opacity={0.5} />

            {alignMode && (
              <AlignmentLayer
                levels={levels}
                plans={overviewPlans}
                config={overviewConfig}
                activeLevelId={alignSelectedId ?? level?.id ?? null}
                upx={upx}
                onFloorPointerDown={handleFloorPointerDown}
              />
            )}

            {!alignMode && (<>
            <ZoneLayer
              zones={plan.zones}
              selection={selection}
              interactive={interactive}
              upx={upx}
              onZonePointerDown={handleZonePointerDown}
              onZoneLabelPointerDown={handleZoneLabelPointerDown}
              onZoneVertexPointerDown={handleZoneVertexPointerDown}
            />
            {tool === 'zone' && <ZoneDraft points={zoneDraft} cursor={cursorSnap?.point ?? null} color={settings.zoneColor} upx={upx} />}

            <WallLayer
              walls={plan.walls}
              selection={selection}
              hoverWallId={hoverWallId}
              interactive={interactive}
              showJoints={tool === 'select' || tool === 'wall'}
              jointTolerance={JOINT_TOLERANCE}
              upx={upx}
              onWallPointerDown={handleWallPointerDown}
              onVertexPointerDown={handleVertexPointerDown}
              onOpeningPointerDown={handleOpeningPointerDown}
              onWallHover={setHoverWallId}
            />

            {hoverProjection && (tool === 'door' || tool === 'window') && (
              <OpeningPreview projection={hoverProjection} width={tool === 'door' ? settings.doorWidth : settings.windowWidth} upx={upx} />
            )}

            {wallStart && wallDraftPoint && (
              <g className="pointer-events-none">
                <line
                  x1={wallStart.x}
                  y1={wallStart.y}
                  x2={wallDraftPoint.x}
                  y2={wallDraftPoint.y}
                  stroke={CANVAS.accentSoft}
                  strokeWidth={settings.wallThickness}
                  strokeLinecap="square"
                  opacity={0.75}
                />
                <circle cx={wallStart.x} cy={wallStart.y} r={4 * upx} fill={CANVAS.accent} stroke="#fff" strokeWidth={1.5 * upx} />
                {wallDraftLength > 0 && (
                  <DimensionLabel
                    x={(wallStart.x + wallDraftPoint.x) / 2}
                    y={(wallStart.y + wallDraftPoint.y) / 2 - (settings.wallThickness / 2 + 16 * upx)}
                    text={geo.formatMeters(wallDraftLength)}
                    upx={upx}
                  />
                )}
              </g>
            )}

            <DeviceLayer
              placements={visiblePlacements}
              devices={devices}
              selection={selection}
              interactive={!deviceToPlace}
              upx={upx}
              onPlacementPointerDown={handlePlacementPointerDown}
            />

            {cursorSnap && (tool === 'wall' || tool === 'zone' || deviceToPlace) && (
              <g className="pointer-events-none">
                {cursorSnap.vertex && (
                  <circle cx={cursorSnap.point.x} cy={cursorSnap.point.y} r={9 * upx} fill="none" stroke={CANVAS.preview} strokeWidth={1.5 * upx} />
                )}
                <circle
                  cx={cursorSnap.point.x}
                  cy={cursorSnap.point.y}
                  r={3.5 * upx}
                  fill={cursorSnap.vertex ? CANVAS.preview : '#ffffff'}
                  stroke={CANVAS.accent}
                  strokeWidth={1.5 * upx}
                />
              </g>
            )}
            </>)}
          </svg>

          {loading && (
            <div className="absolute inset-0 flex items-center justify-center bg-slate-950/40 backdrop-blur-[1px] text-slate-300 text-sm gap-2">
              <Loader2 className="w-4 h-4 animate-spin" /> Loading plan…
            </div>
          )}

          {!level && !loading && (
            <div className="absolute inset-0 flex items-center justify-center p-6">
              <div className="max-w-sm w-full bg-slate-900/95 border border-slate-800 rounded-2xl p-6 shadow-2xl text-center space-y-4">
                <div className="w-12 h-12 rounded-xl bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 flex items-center justify-center mx-auto">
                  <Layers className="w-6 h-6" />
                </div>
                <div className="space-y-1">
                  <h3 className="text-base font-semibold text-white">No level</h3>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    Import a Sweet Home 3D file, restore a backup, or create a first floor or outdoor space.
                  </p>
                </div>
                <div className="space-y-2">
                  <button
                    type="button"
                    onClick={() => setShowWizard(true)}
                    className="w-full bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2.5 rounded-xl text-xs font-semibold shadow-lg shadow-indigo-500/30 transition-all cursor-pointer"
                  >
                    Import / Restore
                  </button>
                  <button
                    type="button"
                    onClick={() => setLevelsOpen(true)}
                    className="w-full bg-slate-800 hover:bg-slate-700 text-slate-200 px-4 py-2.5 rounded-xl text-xs font-semibold transition-all cursor-pointer"
                  >
                    Manage levels
                  </button>
                </div>
              </div>
            </div>
          )}

          {deviceToPlace && (
            <div className="absolute top-3 left-1/2 -translate-x-1/2 z-10 bg-indigo-950/90 backdrop-blur-md border border-indigo-500/50 text-indigo-100 text-xs px-3 py-1.5 rounded-full shadow-xl flex items-center gap-2">
              <Cpu className="w-3.5 h-3.5" />
              <span>
                Place <strong className="text-white">{deviceToPlace.name}</strong> on <strong className="text-indigo-300">{activeLayer}</strong>
              </span>
              <button type="button" onClick={() => setDeviceToPlace(null)} className="ml-1 text-indigo-300 hover:text-white cursor-pointer">
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
          )}

          <div className="absolute bottom-3 left-3 z-10 flex items-center gap-0.5 bg-slate-900/90 backdrop-blur-md border border-slate-800 p-1 rounded-xl shadow-xl">
            <button type="button" onClick={() => zoomBy(1.25)} title="Zoom in (+)" className="w-8 h-8 rounded-lg hover:bg-slate-800 text-slate-300 hover:text-white flex items-center justify-center cursor-pointer">
              <ZoomIn className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => zoomBy(1 / 1.25)} title="Zoom out (-)" className="w-8 h-8 rounded-lg hover:bg-slate-800 text-slate-300 hover:text-white flex items-center justify-center cursor-pointer">
              <ZoomOut className="w-4 h-4" />
            </button>
            <div className="h-4 w-px bg-slate-700 mx-0.5" />
            <button
              type="button"
              onClick={() => (alignMode ? fitToBounds(alignmentWalls(), [], []) : fitToBounds(plan.walls, plan.zones, placements))}
              title="Fit to plan"
              className="h-8 px-2.5 rounded-lg hover:bg-slate-800 text-indigo-300 hover:text-white text-[11px] font-semibold flex items-center gap-1.5 cursor-pointer"
            >
              <Maximize2 className="w-3.5 h-3.5" />
              Fit
            </button>
            <span className="px-2 text-[11px] font-mono text-slate-500">{Math.round((DEFAULT_VIEW.w / viewBox.w) * 100)}%</span>
          </div>

          <div className="absolute bottom-3 left-1/2 -translate-x-1/2 z-10 max-w-[60%] bg-slate-900/85 backdrop-blur-md border border-slate-800 text-slate-300 text-[11px] px-3 py-1.5 rounded-full shadow-xl truncate pointer-events-none">
            {statusText}
          </div>

          {cursorRaw && (
            <div className="absolute bottom-3 right-3 z-10 bg-slate-900/85 backdrop-blur-md border border-slate-800 text-slate-500 text-[11px] font-mono px-2.5 py-1.5 rounded-lg pointer-events-none">
              x {Math.round(cursorRaw.x)} · y {Math.round(cursorRaw.y)}
            </div>
          )}
        </div>

        {(alignMode || panelOpen) && (
          <aside className="w-80 shrink-0 bg-slate-900 border-l border-slate-800 flex flex-col min-h-0">
            {alignMode ? (
              <FloorAlignmentPanel
                levels={levels}
                config={overviewConfig}
                activeLevelId={alignSelectedId ?? level?.id ?? null}
                onSelect={setAlignSelectedId}
                onChange={(config) => onOverviewConfigChange?.(config)}
                onReset={resetAlignment}
              />
            ) : (
            <>
            <div className="flex items-center border-b border-slate-800 px-2 pt-2 gap-1">
              <PanelTab active={panelTab === 'inspector'} onClick={() => setPanelTab('inspector')} icon={<SlidersHorizontal className="w-3.5 h-3.5" />} label="Properties" />
              <PanelTab
                active={panelTab === 'devices'}
                onClick={() => setPanelTab('devices')}
                icon={<Cpu className="w-3.5 h-3.5" />}
                label="Devices"
                badge={`${placements.length}/${devices.length}`}
              />
            </div>
            <div className="flex-1 min-h-0">
              {panelTab === 'inspector' ? (
                <Inspector
                  plan={plan}
                  placements={placements}
                  devices={devices}
                  level={level}
                  selection={selection}
                  tool={tool}
                  settings={settings}
                  onSettings={(patch) => setSettings((s) => ({ ...s, ...patch }))}
                  zoneDraftCount={zoneDraft.length}
                  onCompleteZone={completeZone}
                  onCancelZone={() => setZoneDraft([])}
                  wallDrafting={!!wallStart}
                  onCancelWall={endWallChain}
                  onUpdateWall={updateWall}
                  onSetWallLength={setWallLengthById}
                  onSplitWall={splitWallById}
                  onUpdateOpening={updateOpening}
                  onUpdateZone={updateZone}
                  onUpdatePlacement={updatePlacement}
                  onDeleteSelection={deleteSelection}
                  onSelect={selectElement}
                />
              ) : (
                <DevicePalette
                  devices={devices}
                  loading={loadingDevices}
                  placedDeviceIds={placedDeviceIds}
                  deviceToPlace={deviceToPlace}
                  onPickDevice={(dev) => {
                    setDeviceToPlace(dev)
                    if (dev) setSelection(null)
                  }}
                  onRefresh={fetchDevices}
                />
              )}
            </div>
            </>
            )}
          </aside>
        )}
      </div>

      {level && <ImportPlanModal isOpen={showImport} levelId={level.id} onClose={() => setShowImport(false)} onImport={handleImport} />}
      {showWizard && (
        <OnboardingWizard
          onDone={async () => {
            setShowWizard(false)
            await onRefreshLevels()
          }}
          onClose={() => setShowWizard(false)}
        />
      )}
    </div>
  )
}

function PanelTab({ active, onClick, icon, label, badge }: { active: boolean; onClick: () => void; icon: ReactNode; label: string; badge?: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex-1 h-9 rounded-t-lg text-xs font-semibold flex items-center justify-center gap-1.5 border-b-2 transition-colors cursor-pointer ${
        active ? 'text-white border-indigo-500' : 'text-slate-400 border-transparent hover:text-white'
      }`}
    >
      {icon}
      <span>{label}</span>
      {badge && <span className="text-[10px] font-mono text-slate-500">{badge}</span>}
    </button>
  )
}

function OpeningPreview({ projection, width, upx }: { projection: geo.WallProjection; width: number; upx: number }) {
  const len = geo.wallLength(projection.wall)
  const half = Math.min(width, len) / 2
  const offset = Math.max(half, Math.min(len - half, projection.offset))
  const ux = Math.cos(projection.angle)
  const uy = Math.sin(projection.angle)
  const cx = projection.wall.x1 + ux * offset
  const cy = projection.wall.y1 + uy * offset
  const th = projection.wall.thickness || 12
  return (
    <g transform={`translate(${cx} ${cy}) rotate(${(projection.angle * 180) / Math.PI})`} className="pointer-events-none">
      <rect
        x={-half}
        y={-th / 2 - 3 * upx}
        width={half * 2}
        height={th + 6 * upx}
        rx={2 * upx}
        fill={CANVAS.preview}
        fillOpacity={0.3}
        stroke={CANVAS.preview}
        strokeWidth={1.5 * upx}
        strokeDasharray={`${4 * upx} ${2 * upx}`}
      />
    </g>
  )
}
