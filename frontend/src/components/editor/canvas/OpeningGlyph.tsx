import type { WallOpening, WallSegment } from '../../../types'
import { CANVAS } from '../constants'
import { wallAngleDeg } from '../geometry'
import type { OpeningDragMode } from '../types'

interface OpeningGlyphProps {
  wall: WallSegment
  opening: WallOpening
  upx: number
  selected: boolean
  interactive: boolean
  onPointerDown?: (mode: OpeningDragMode, e: React.PointerEvent) => void
}

function openingFrame(wall: WallSegment, opening: WallOpening) {
  const dx = wall.x2 - wall.x1
  const dy = wall.y2 - wall.y1
  const len = Math.hypot(dx, dy) || 1
  return {
    cx: wall.x1 + (dx / len) * opening.offset,
    cy: wall.y1 + (dy / len) * opening.offset,
    angle: wallAngleDeg(wall),
  }
}

/** Hole punched through the wall stroke; drawn before the glyph so the grid shows through. */
export function OpeningCutout({ wall, opening }: { wall: WallSegment; opening: WallOpening }) {
  const { cx, cy, angle } = openingFrame(wall, opening)
  const th = wall.thickness || 12
  const hw = opening.width / 2
  return (
    <g transform={`translate(${cx} ${cy}) rotate(${angle})`} className="pointer-events-none">
      <rect x={-hw} y={-th / 2 - 1.5} width={opening.width} height={th + 3} fill={CANVAS.background} />
      <rect x={-hw} y={-th / 2 - 1.5} width={opening.width} height={th + 3} fill="url(#editorGrid)" />
    </g>
  )
}

export function OpeningGlyph({ wall, opening, upx, selected, interactive, onPointerDown }: OpeningGlyphProps) {
  const { cx, cy, angle } = openingFrame(wall, opening)
  const th = wall.thickness || 12
  const hw = opening.width / 2
  const thin = 1.2 * upx
  const bold = 2.2 * upx
  const jambColor = selected ? CANVAS.accentSoft : CANVAS.doorGlyph
  const isWindow = opening.type === 'window'

  return (
    <g transform={`translate(${cx} ${cy}) rotate(${angle})`} className={interactive ? 'cursor-pointer' : ''}>
      {isWindow ? (
        <g className="pointer-events-none">
          <rect x={-hw} y={-th / 2} width={opening.width} height={th} fill="#1e293b" stroke={jambColor} strokeWidth={thin} />
          <line x1={-hw} y1={0} x2={hw} y2={0} stroke={CANVAS.windowGlass} strokeWidth={Math.min(th * 0.28, 3 * upx)} />
          <line x1={0} y1={-th / 2} x2={0} y2={th / 2} stroke="#7dd3fc" strokeWidth={thin} />
          <line x1={-hw} y1={-th / 2 - 1.5 * upx} x2={-hw} y2={th / 2 + 1.5 * upx} stroke={jambColor} strokeWidth={bold} />
          <line x1={hw} y1={-th / 2 - 1.5 * upx} x2={hw} y2={th / 2 + 1.5 * upx} stroke={jambColor} strokeWidth={bold} />
        </g>
      ) : (
        <g className="pointer-events-none">
          <line x1={-hw} y1={-th / 2 - 2 * upx} x2={-hw} y2={th / 2 + 2 * upx} stroke={jambColor} strokeWidth={bold} />
          <line x1={hw} y1={-th / 2 - 2 * upx} x2={hw} y2={th / 2 + 2 * upx} stroke={jambColor} strokeWidth={bold} />
          <line
            x1={-hw}
            y1={0}
            x2={hw}
            y2={0}
            stroke="#64748b"
            strokeWidth={thin}
            strokeDasharray={`${2 * upx} ${2 * upx}`}
          />
          {opening.hide_door ? (
            <PassageMarks hw={hw} th={th} upx={upx} />
          ) : (
            <DoorSwing opening={opening} hw={hw} th={th} upx={upx} />
          )}
        </g>
      )}

      {selected && (
        <g>
          <rect
            x={-hw - 3 * upx}
            y={-th / 2 - 4 * upx}
            width={opening.width + 6 * upx}
            height={th + 8 * upx}
            rx={3 * upx}
            fill="none"
            stroke={CANVAS.accent}
            strokeWidth={1.5 * upx}
            strokeDasharray={`${3 * upx} ${2 * upx}`}
            className="pointer-events-none"
          />
          {(['resize-start', 'resize-end'] as OpeningDragMode[]).map((mode) => (
            <circle
              key={mode}
              cx={mode === 'resize-start' ? -hw : hw}
              cy={0}
              r={5 * upx}
              fill={CANVAS.handleFill}
              stroke={CANVAS.accent}
              strokeWidth={1.5 * upx}
              className="cursor-ew-resize"
              onPointerDown={(e) => {
                e.stopPropagation()
                onPointerDown?.(mode, e)
              }}
            />
          ))}
        </g>
      )}

      <rect
        x={-hw}
        y={-th / 2 - 6 * upx}
        width={opening.width}
        height={th + 12 * upx}
        fill="transparent"
        style={{ pointerEvents: interactive ? 'all' : 'none' }}
        onPointerDown={(e) => {
          e.stopPropagation()
          onPointerDown?.('move', e)
        }}
      />
    </g>
  )
}

function DoorSwing({ opening, hw, th, upx }: { opening: WallOpening; hw: number; th: number; upx: number }) {
  const side = opening.flip_side ? 1 : -1
  const yBase = side * (th / 2)
  const thin = 1 * upx
  const leaf = 1.6 * upx
  const isDouble = opening.width >= 50

  if (isDouble) {
    const leftSweep = side === -1 ? 1 : 0
    const rightSweep = side === -1 ? 0 : 1
    return (
      <>
        <line x1={0} y1={-th / 2} x2={0} y2={th / 2} stroke={CANVAS.doorGlyph} strokeWidth={leaf} />
        <line x1={-hw} y1={yBase} x2={-hw} y2={yBase + side * hw} stroke={CANVAS.doorGlyph} strokeWidth={leaf} />
        <path
          d={`M ${-hw} ${yBase + side * hw} A ${hw} ${hw} 0 0 ${leftSweep} 0 ${yBase}`}
          fill="none"
          stroke="#64748b"
          strokeWidth={thin}
          strokeDasharray={`${2 * upx} ${2 * upx}`}
        />
        <line x1={hw} y1={yBase} x2={hw} y2={yBase + side * hw} stroke={CANVAS.doorGlyph} strokeWidth={leaf} />
        <path
          d={`M ${hw} ${yBase + side * hw} A ${hw} ${hw} 0 0 ${rightSweep} 0 ${yBase}`}
          fill="none"
          stroke="#64748b"
          strokeWidth={thin}
          strokeDasharray={`${2 * upx} ${2 * upx}`}
        />
      </>
    )
  }

  const hingeX = opening.flip_hinge ? hw : -hw
  const targetX = opening.flip_hinge ? -hw : hw
  const tipY = yBase + side * opening.width
  const sweep = opening.flip_hinge ? (side === -1 ? 0 : 1) : side === -1 ? 1 : 0
  return (
    <>
      <line x1={hingeX} y1={yBase} x2={hingeX} y2={tipY} stroke={CANVAS.doorGlyph} strokeWidth={leaf} />
      <path
        d={`M ${hingeX} ${tipY} A ${opening.width} ${opening.width} 0 0 ${sweep} ${targetX} ${yBase}`}
        fill="none"
        stroke="#64748b"
        strokeWidth={thin}
        strokeDasharray={`${2 * upx} ${2 * upx}`}
      />
    </>
  )
}

function PassageMarks({ hw, th, upx }: { hw: number; th: number; upx: number }) {
  const arrow = Math.min(hw * 0.4, 6 * upx)
  const reach = th / 2 + 6 * upx
  return (
    <g stroke="#94a3b8" strokeWidth={1.2 * upx} fill="none" strokeLinecap="round" strokeLinejoin="round">
      <line x1={0} y1={-reach} x2={0} y2={reach} strokeDasharray={`${1.5 * upx} ${2 * upx}`} />
      <path d={`M ${-arrow} ${-reach + arrow} L 0 ${-reach} L ${arrow} ${-reach + arrow}`} />
      <path d={`M ${-arrow} ${reach - arrow} L 0 ${reach} L ${arrow} ${reach - arrow}`} />
    </g>
  )
}
