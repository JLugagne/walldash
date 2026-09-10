import React, { useCallback, useEffect, useRef, useState } from 'react'
import { PencilLine, Trash2, X } from 'lucide-react'
import type { DisplayMode, OverviewDashboard, Widget } from '../../types'
import { widgetsToRects } from './format'
import { clampToMinSize, reflow, type GridSize, type Rect } from './grid'

interface DragInfo {
  widgetId: string
  display: DisplayMode
  mode: 'move' | 'resize'
  startX: number
  startY: number
  cellWidth: number
  cellHeight: number
  originRects: Rect[]
  origin: Rect
}

function widgetLabel(widget: Widget): string {
  const entityId = widget.config.entity_ids?.[0]
  return widget.title || (entityId ? widget.config.labels?.[entityId] : undefined) || entityId || 'Widget'
}

function sameLayout(a: Rect[], b: Rect[]): boolean {
  if (a.length !== b.length) return false
  const byId = new Map(b.map((r) => [r.id, r]))
  return a.every((r) => {
    const o = byId.get(r.id)
    return (
      o !== undefined &&
      o.col === r.col &&
      o.row === r.row &&
      o.colSpan === r.colSpan &&
      o.rowSpan === r.rowSpan
    )
  })
}

export interface WidgetGridProps {
  overview: OverviewDashboard
  isEditMode: boolean
  renderWidgetBody: (widget: Widget) => React.ReactNode
  onEditWidget: (widget: Widget) => void
  onDeleteWidget: (widgetId: string) => void
  /**
   * Persists a full grid layout (per ADR 0004, all-or-nothing). Resolves to `null` when the
   * server accepted it, otherwise to the user-facing message explaining the refusal or the
   * network failure; WidgetGrid forwards that message to `onMessage`.
   */
  onPersistLayout: (rects: Rect[]) => Promise<string | null>
  /** Shows a transient banner; the view owns the single toast so two surfaces never overlap. */
  onMessage: (message: string) => void
}

/**
 * WidgetGrid owns one Overview Dashboard's Widget Grid: the cell layout, the edit-mode guides,
 * the drag/move and drag/resize gesture end to end (including its live reflow preview and the
 * refusal ring). Refusals, local or server-side, are reported through `onMessage`.
 *
 * The gesture is pointer-based and touch-safe: the move overlay and the resize handle opt out of
 * browser panning, `pointercancel` and window `blur` snap the layout back silently, a tap or a
 * drag that ends where it started persists nothing, and a resize never previews a span below the
 * Display Mode minimum so the server is never asked for a layout it would reject.
 *
 * Edit Mode keeps every Widget's content readable: pressing a Widget selects it (indigo ring),
 * and only the selected one shows a resize handle, which overhangs its bottom-right corner into
 * the grid gap. Edit and delete live in a floating action bar at the bottom of the grid, never
 * over the card, so a 1-row Widget's value, track or button stays visible while laying out.
 */
export const WidgetGrid: React.FC<WidgetGridProps> = ({
  overview,
  isEditMode,
  renderWidgetBody,
  onEditWidget,
  onDeleteWidget,
  onPersistLayout,
  onMessage,
}) => {
  const gridRef = useRef<HTMLDivElement>(null)
  const [dragInfo, setDragInfo] = useState<DragInfo | null>(null)
  const [previewRects, setPreviewRects] = useState<Rect[] | null>(null)
  const [dragRefused, setDragRefused] = useState(false)
  const previewRectsRef = useRef<Rect[] | null>(null)
  const dragRefusedRef = useRef(false)
  const [selectedWidgetId, setSelectedWidgetId] = useState<string | null>(null)
  const selectedWidget = isEditMode ? (overview.widgets.find((w) => w.id === selectedWidgetId) ?? null) : null

  const setPreview = useCallback((rects: Rect[] | null, refused: boolean) => {
    previewRectsRef.current = rects
    dragRefusedRef.current = refused
    setPreviewRects(rects)
    setDragRefused(refused)
  }, [])

  const startDrag = (widget: Widget, mode: 'move' | 'resize', e: React.PointerEvent) => {
    if (!isEditMode || !gridRef.current) return
    e.preventDefault()
    const rect = gridRef.current.getBoundingClientRect()
    setSelectedWidgetId(widget.id)
    setDragInfo({
      widgetId: widget.id,
      display: widget.config.display,
      mode,
      startX: e.clientX,
      startY: e.clientY,
      cellWidth: rect.width / overview.cols,
      cellHeight: rect.height / overview.rows,
      originRects: widgetsToRects(overview.widgets),
      origin: { id: widget.id, col: widget.col, row: widget.row, colSpan: widget.col_span, rowSpan: widget.row_span },
    })
    setPreview(null, false)
  }

  useEffect(() => {
    if (!dragInfo) return

    const gridSize: GridSize = { cols: overview.cols, rows: overview.rows }

    const resetDrag = () => {
      setDragInfo(null)
      setPreview(null, false)
    }

    const handleMove = (ev: PointerEvent) => {
      const deltaCols = Math.round((ev.clientX - dragInfo.startX) / dragInfo.cellWidth)
      const deltaRows = Math.round((ev.clientY - dragInfo.startY) / dragInfo.cellHeight)

      let target: { col: number; row: number; colSpan: number; rowSpan: number }
      if (dragInfo.mode === 'move') {
        const col = Math.min(Math.max(dragInfo.origin.col + deltaCols, 0), gridSize.cols - dragInfo.origin.colSpan)
        const row = Math.min(Math.max(dragInfo.origin.row + deltaRows, 0), gridSize.rows - dragInfo.origin.rowSpan)
        target = { col, row, colSpan: dragInfo.origin.colSpan, rowSpan: dragInfo.origin.rowSpan }
      } else {
        const colSpan = Math.min(Math.max(dragInfo.origin.colSpan + deltaCols, 1), gridSize.cols - dragInfo.origin.col)
        const rowSpan = Math.min(Math.max(dragInfo.origin.rowSpan + deltaRows, 1), gridSize.rows - dragInfo.origin.row)
        const span = clampToMinSize(dragInfo.display, colSpan, rowSpan)
        target = { col: dragInfo.origin.col, row: dragInfo.origin.row, colSpan: span.colSpan, rowSpan: span.rowSpan }
      }

      const result = reflow(dragInfo.originRects, gridSize, dragInfo.widgetId, target)
      if (result.ok) {
        setPreview(result.rects, false)
      } else {
        setPreview(previewRectsRef.current, true)
      }
    }

    const detach = () => {
      window.removeEventListener('pointermove', handleMove)
      window.removeEventListener('pointerup', handleUp)
      window.removeEventListener('pointercancel', handleCancel)
      window.removeEventListener('blur', handleCancel)
    }

    const handleCancel = () => {
      detach()
      resetDrag()
    }

    const handleUp = () => {
      detach()
      const finalRects = previewRectsRef.current
      const refused = dragRefusedRef.current
      resetDrag()
      if (refused) {
        onMessage(
          dragInfo.mode === 'resize'
            ? 'Resize refused: this size would push a widget outside the grid.'
            : 'Move refused: this placement would push a widget outside the grid.',
        )
        return
      }
      if (!finalRects || sameLayout(finalRects, dragInfo.originRects)) return
      void onPersistLayout(finalRects).then((message) => {
        if (message) onMessage(message)
      })
    }

    window.addEventListener('pointermove', handleMove)
    window.addEventListener('pointerup', handleUp)
    window.addEventListener('pointercancel', handleCancel)
    window.addEventListener('blur', handleCancel)
    return detach
  }, [dragInfo, overview, onPersistLayout, onMessage, setPreview])

  const effectiveRects: Rect[] = dragInfo && previewRects ? previewRects : widgetsToRects(overview.widgets)
  const rectById = new Map(effectiveRects.map((r) => [r.id, r]))
  const selectedRect = selectedWidget ? rectById.get(selectedWidget.id) : undefined
  const actionBarAtTop = selectedRect !== undefined && selectedRect.row + selectedRect.rowSpan >= overview.rows

  return (
    <>
      <div
        ref={gridRef}
        className="relative w-full h-full grid gap-3"
        style={{
          gridTemplateColumns: `repeat(${overview.cols}, minmax(0, 1fr))`,
          gridTemplateRows: `repeat(${overview.rows}, minmax(0, 1fr))`,
        }}
        onPointerDown={(e) => {
          if (e.target === e.currentTarget) setSelectedWidgetId(null)
        }}
      >
        {isEditMode && (
          <div
            className="absolute inset-0 grid pointer-events-none"
            style={{
              gridTemplateColumns: `repeat(${overview.cols}, minmax(0, 1fr))`,
              gridTemplateRows: `repeat(${overview.rows}, minmax(0, 1fr))`,
              gap: '0.75rem',
            }}
          >
            {Array.from({ length: overview.cols * overview.rows }).map((_, i) => (
              <div key={i} className="border border-dashed border-slate-700/40 rounded-lg" />
            ))}
          </div>
        )}

        {overview.widgets.map((widget) => {
          const rect = rectById.get(widget.id) || {
            id: widget.id,
            col: widget.col,
            row: widget.row,
            colSpan: widget.col_span,
            rowSpan: widget.row_span,
          }
          const isDragTarget = dragInfo?.widgetId === widget.id
          const isSelected = selectedWidget?.id === widget.id
          const radius = rect.rowSpan === 1 ? 'rounded-lg' : 'rounded-xl'

          return (
            <div
              key={widget.id}
              className={`relative min-w-0 min-h-0 ${isSelected ? 'z-20' : 'z-10'}`}
              style={{
                gridColumn: `${rect.col + 1} / span ${rect.colSpan}`,
                gridRow: `${rect.row + 1} / span ${rect.rowSpan}`,
              }}
            >
              <div className="absolute inset-0">{renderWidgetBody(widget)}</div>

              {isEditMode && (
                <>
                  <div
                    className={`absolute inset-0 z-10 cursor-move touch-none select-none ${radius}`}
                    onPointerDown={(e) => startDrag(widget, 'move', e)}
                  />
                  {isSelected && (
                    <div
                      className="absolute -bottom-[22px] -right-[22px] z-20 w-11 h-11 flex items-center justify-center cursor-se-resize touch-none select-none"
                      onPointerDown={(e) => startDrag(widget, 'resize', e)}
                      title="Resize"
                    >
                      <div className="w-6 h-6 rounded-md bg-[#6d76e8] border border-[#8b93ee] shadow-sm" />
                    </div>
                  )}
                  {(isSelected || isDragTarget) && (
                    <div
                      className={`absolute inset-0 z-30 ring-2 pointer-events-none ${radius} ${
                        isDragTarget && dragRefused ? 'ring-red-500' : 'ring-indigo-400'
                      }`}
                    />
                  )}
                </>
              )}
            </div>
          )
        })}
      </div>

      {selectedWidget && (
        <div
          className={`absolute left-1/2 -translate-x-1/2 z-40 flex items-center gap-1 h-12 pl-4 pr-1.5 rounded-xl bg-slate-900/70 backdrop-blur-md border border-slate-800/80 shadow-xl ${
            actionBarAtTop ? 'top-4' : 'bottom-4'
          }`}
        >
          <span className="text-xs font-semibold text-slate-300 truncate max-w-[12rem]" title={widgetLabel(selectedWidget)}>
            {widgetLabel(selectedWidget)}
          </span>
          <span className="w-px h-6 bg-slate-700 mx-1" aria-hidden="true" />
          <button
            type="button"
            onClick={() => onEditWidget(selectedWidget)}
            className="h-10 px-3 rounded-lg flex items-center gap-1.5 text-xs font-semibold text-slate-200 hover:bg-[#6d76e8]/30 active:bg-[#6d76e8]/40 transition-colors"
            title="Edit widget content"
          >
            <PencilLine className="w-4 h-4" />
            Edit
          </button>
          <button
            type="button"
            onClick={() => onDeleteWidget(selectedWidget.id)}
            className="h-10 px-3 rounded-lg flex items-center gap-1.5 text-xs font-semibold text-rose-300 hover:bg-rose-500/20 active:bg-rose-500/30 transition-colors"
            title="Delete widget"
          >
            <Trash2 className="w-4 h-4" />
            Delete
          </button>
          <button
            type="button"
            onClick={() => setSelectedWidgetId(null)}
            className="h-10 w-10 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
            title="Deselect"
            aria-label="Deselect"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

    </>
  )
}
