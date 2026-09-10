import type { Point2D, Zone } from '../../../types'
import { poleOfInaccessibility } from '../../floorTexture'
import { CANVAS } from '../constants'
import type { EditorSelection } from '../types'

interface ZoneLayerProps {
  zones: Zone[]
  selection: EditorSelection | null
  interactive: boolean
  upx: number
  onZonePointerDown: (zone: Zone, e: React.PointerEvent) => void
  onZoneLabelPointerDown: (zone: Zone, e: React.PointerEvent) => void
  onZoneVertexPointerDown: (zone: Zone, index: number, e: React.PointerEvent) => void
}

export function ZoneLayer({ zones, selection, interactive, upx, onZonePointerDown, onZoneLabelPointerDown, onZoneVertexPointerDown }: ZoneLayerProps) {
  return (
    <g>
      {zones.map((zone) => {
        const selected = selection?.type === 'zone' && selection.id === zone.id
        const pointsStr = zone.points.map((p) => `${p.x},${p.y}`).join(' ')
        const label = zone.label_position ?? poleOfInaccessibility(zone.points)
        return (
          <g key={zone.id}>
            <polygon
              points={pointsStr}
              fill={zone.color}
              fillOpacity={selected ? 0.32 : 0.16}
              stroke={selected ? '#ffffff' : zone.color}
              strokeOpacity={selected ? 0.9 : 0.7}
              strokeWidth={(selected ? 1.8 : 1.2) * upx}
              strokeDasharray={selected ? `${4 * upx} ${2 * upx}` : undefined}
              strokeLinejoin="round"
              style={{ pointerEvents: interactive ? 'all' : 'none' }}
              className={interactive ? 'cursor-move' : ''}
              onPointerDown={(e) => {
                e.stopPropagation()
                onZonePointerDown(zone, e)
              }}
            />
            <ZoneLabel
              x={label.x}
              y={label.y}
              name={zone.name}
              color={zone.color}
              upx={upx}
              interactive={interactive}
              onPointerDown={(e) => onZoneLabelPointerDown(zone, e)}
            />
            {selected &&
              zone.points.map((p, i) => (
                <circle
                  key={i}
                  cx={p.x}
                  cy={p.y}
                  r={5 * upx}
                  fill={CANVAS.handleFill}
                  stroke={zone.color}
                  strokeWidth={2 * upx}
                  className="cursor-move"
                  onPointerDown={(e) => {
                    e.stopPropagation()
                    onZoneVertexPointerDown(zone, i, e)
                  }}
                />
              ))}
          </g>
        )
      })}
    </g>
  )
}

function ZoneLabel({ x, y, name, color, upx, interactive, onPointerDown }: { x: number; y: number; name: string; color: string; upx: number; interactive: boolean; onPointerDown: (e: React.PointerEvent) => void }) {
  const width = name.length * 6.2 + 26
  return (
    <g
      transform={`translate(${x} ${y}) scale(${upx})`}
      className={`${interactive ? 'cursor-move' : 'pointer-events-none'} select-none`}
      style={{ pointerEvents: interactive ? 'all' : 'none' }}
      onPointerDown={(e) => {
        e.stopPropagation()
        onPointerDown(e)
      }}
    >
      <rect x={-width / 2} y={-10} width={width} height={20} rx={10} fill="#0f172a" fillOpacity={0.85} stroke="#334155" strokeWidth={1} />
      <circle cx={-width / 2 + 10} cy={0} r={3.5} fill={color} stroke="#ffffff" strokeOpacity={0.5} strokeWidth={1} />
      <text x={5} y={3.5} textAnchor="middle" fill={CANVAS.label} fontSize={11} fontWeight={600}>
        {name}
      </text>
    </g>
  )
}

export function ZoneDraft({ points, cursor, color, upx }: { points: Point2D[]; cursor: Point2D | null; color: string; upx: number }) {
  if (points.length === 0) return null
  const last = points[points.length - 1]
  return (
    <g className="pointer-events-none">
      {points.length > 2 && (
        <polygon points={points.map((p) => `${p.x},${p.y}`).join(' ')} fill={color} fillOpacity={0.12} stroke="none" />
      )}
      {points.length > 1 && (
        <polyline
          points={points.map((p) => `${p.x},${p.y}`).join(' ')}
          fill="none"
          stroke={color}
          strokeWidth={2 * upx}
          strokeDasharray={`${4 * upx} ${3 * upx}`}
        />
      )}
      {cursor && (
        <line x1={last.x} y1={last.y} x2={cursor.x} y2={cursor.y} stroke={color} strokeWidth={1.5 * upx} strokeDasharray={`${2 * upx} ${2 * upx}`} />
      )}
      {points.map((p, i) => (
        <circle
          key={i}
          cx={p.x}
          cy={p.y}
          r={(i === 0 ? 6 : 4) * upx}
          fill={i === 0 ? '#38bdf8' : color}
          stroke="#ffffff"
          strokeWidth={1.5 * upx}
        />
      ))}
    </g>
  )
}
