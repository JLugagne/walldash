import type { WallOpening, WallSegment } from '../../../types'
import { CANVAS } from '../constants'
import { formatMeters, sharedJoints, wallLength, type VertexRef, type WallEnd } from '../geometry'
import type { EditorSelection, OpeningDragMode } from '../types'
import { OpeningCutout, OpeningGlyph } from './OpeningGlyph'

interface WallLayerProps {
  walls: WallSegment[]
  selection: EditorSelection | null
  hoverWallId: string | null
  interactive: boolean
  showJoints: boolean
  jointTolerance: number
  upx: number
  onWallPointerDown: (wall: WallSegment, e: React.PointerEvent) => void
  onVertexPointerDown: (ref: VertexRef, e: React.PointerEvent) => void
  onOpeningPointerDown: (wall: WallSegment, opening: WallOpening, mode: OpeningDragMode, e: React.PointerEvent) => void
  onWallHover: (wallId: string | null) => void
}

export function WallLayer({
  walls,
  selection,
  hoverWallId,
  interactive,
  showJoints,
  jointTolerance,
  upx,
  onWallPointerDown,
  onVertexPointerDown,
  onOpeningPointerDown,
  onWallHover,
}: WallLayerProps) {
  const selectedWallId = selection?.type === 'wall' ? selection.id : selection?.type === 'opening' ? selection.wallId : null
  const selectedWall = walls.find((w) => w.id === selectedWallId && selection?.type === 'wall') ?? null
  const joints = showJoints ? sharedJoints(walls, jointTolerance) : []
  const outlineExtra = 2 * upx

  return (
    <g>
      <g className="pointer-events-none">
        {walls.map((w) => {
          const isSelected = w.id === selectedWallId
          return (
            <line
              key={w.id}
              x1={w.x1}
              y1={w.y1}
              x2={w.x2}
              y2={w.y2}
              stroke={isSelected ? CANVAS.wallSelectedOutline : CANVAS.wallOutline}
              strokeWidth={(w.thickness || 12) + outlineExtra}
              strokeLinecap="square"
            />
          )
        })}
      </g>

      <g className="pointer-events-none">
        {walls.map((w) => {
          const isSelected = w.id === selectedWallId
          const isHover = !isSelected && interactive && w.id === hoverWallId
          return (
            <line
              key={w.id}
              x1={w.x1}
              y1={w.y1}
              x2={w.x2}
              y2={w.y2}
              stroke={isSelected ? CANVAS.wallSelected : isHover ? CANVAS.wallHover : CANVAS.wallFill}
              strokeWidth={w.thickness || 12}
              strokeLinecap="square"
              filter={isSelected ? 'url(#editorGlow)' : undefined}
            />
          )
        })}
      </g>

      <g>
        {walls.map((w) => (w.openings ?? []).map((op) => <OpeningCutout key={op.id} wall={w} opening={op} />))}
      </g>

      <g>
        {walls.map((w) => (
          <line
            key={w.id}
            x1={w.x1}
            y1={w.y1}
            x2={w.x2}
            y2={w.y2}
            stroke="transparent"
            strokeWidth={Math.max(w.thickness || 12, 14 * upx)}
            strokeLinecap="square"
            style={{ pointerEvents: interactive ? 'stroke' : 'none' }}
            className={interactive ? 'cursor-move' : ''}
            onPointerDown={(e) => {
              e.stopPropagation()
              onWallPointerDown(w, e)
            }}
            onPointerEnter={() => onWallHover(w.id)}
            onPointerLeave={() => onWallHover(null)}
          />
        ))}
      </g>

      <g>
        {walls.map((w) =>
          (w.openings ?? []).map((op) => (
            <OpeningGlyph
              key={op.id}
              wall={w}
              opening={op}
              upx={upx}
              selected={selection?.type === 'opening' && selection.id === op.id}
              interactive={interactive}
              onPointerDown={(mode, e) => onOpeningPointerDown(w, op, mode, e)}
            />
          ))
        )}
      </g>

      {joints.length > 0 && (
        <g className="pointer-events-none">
          {joints.map((j, i) => (
            <circle key={i} cx={j.x} cy={j.y} r={2.5 * upx} fill={CANVAS.background} stroke={CANVAS.joint} strokeWidth={1.2 * upx} />
          ))}
        </g>
      )}

      {selectedWall && (
        <SelectedWallOverlay wall={selectedWall} upx={upx} onVertexPointerDown={onVertexPointerDown} />
      )}
    </g>
  )
}

function SelectedWallOverlay({
  wall,
  upx,
  onVertexPointerDown,
}: {
  wall: WallSegment
  upx: number
  onVertexPointerDown: (ref: VertexRef, e: React.PointerEvent) => void
}) {
  const len = wallLength(wall)
  const midX = (wall.x1 + wall.x2) / 2
  const midY = (wall.y1 + wall.y2) / 2
  const dx = wall.x2 - wall.x1
  const dy = wall.y2 - wall.y1
  const nx = len > 0 ? -dy / len : 0
  const ny = len > 0 ? dx / len : -1
  const labelOffset = (wall.thickness || 12) / 2 + 16 * upx
  const labelX = midX + nx * labelOffset * (ny > 0 ? -1 : 1)
  const labelY = midY + ny * labelOffset * (ny > 0 ? -1 : 1)

  return (
    <g>
      <DimensionLabel x={labelX} y={labelY} text={formatMeters(len)} upx={upx} />
      {(['start', 'end'] as WallEnd[]).map((end) => {
        const px = end === 'start' ? wall.x1 : wall.x2
        const py = end === 'start' ? wall.y1 : wall.y2
        return (
          <circle
            key={end}
            cx={px}
            cy={py}
            r={6 * upx}
            fill={CANVAS.handleFill}
            stroke={CANVAS.accent}
            strokeWidth={2 * upx}
            className="cursor-move"
            onPointerDown={(e) => {
              e.stopPropagation()
              onVertexPointerDown({ wallId: wall.id, end }, e)
            }}
          />
        )
      })}
    </g>
  )
}

export function DimensionLabel({ x, y, text, upx, color }: { x: number; y: number; text: string; upx: number; color?: string }) {
  const width = text.length * 6.5 + 14
  return (
    <g transform={`translate(${x} ${y}) scale(${upx})`} className="pointer-events-none select-none">
      <rect x={-width / 2} y={-10} width={width} height={20} rx={6} fill={CANVAS.labelBg} stroke={color ?? CANVAS.accent} strokeWidth={1} />
      <text x={0} y={3.5} textAnchor="middle" fill="#c7d2fe" fontSize={11} fontWeight={600} fontFamily="ui-monospace, monospace">
        {text}
      </text>
    </g>
  )
}
