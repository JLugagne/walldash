import type { Device, DevicePlacement } from '../../../types'
import { domainStyle, isDeviceActive } from '../constants'
import type { EditorSelection } from '../types'

interface DeviceLayerProps {
  placements: DevicePlacement[]
  devices: Device[]
  selection: EditorSelection | null
  interactive: boolean
  upx: number
  onPlacementPointerDown: (placement: DevicePlacement, e: React.PointerEvent) => void
}

export function DeviceLayer({ placements, devices, selection, interactive, upx, onPlacementPointerDown }: DeviceLayerProps) {
  return (
    <g>
      {placements.map((p) => {
        const dev = devices.find((d) => d.id === p.device_id)
        const domain = p.render_domain || dev?.domain || p.icon || 'light'
        const style = domainStyle(domain)
        const Icon = style.Icon
        const selected = selection?.type === 'device' && selection.id === p.id
        const name = p.custom_name || dev?.name || p.device_id
        const active = isDeviceActive(dev?.state)
        return (
          <g
            key={p.id}
            transform={`translate(${p.x} ${p.y}) scale(${upx})`}
            className={`group select-none ${interactive ? 'cursor-move' : ''}`}
            style={{ pointerEvents: interactive ? 'all' : 'none' }}
            onPointerDown={(e) => {
              e.stopPropagation()
              onPlacementPointerDown(p, e)
            }}
          >
            <circle cx={0} cy={0} r={22} fill="transparent" />
            {selected && (
              <circle cx={0} cy={0} r={22} fill="none" stroke={style.ring} strokeWidth={2} strokeDasharray="4 3" className="animate-pulse pointer-events-none" />
            )}
            <circle
              cx={0}
              cy={0}
              r={19}
              fill="none"
              stroke={style.ring}
              strokeWidth={1.5}
              strokeDasharray="3 3"
              className="opacity-0 group-hover:opacity-70 transition-opacity pointer-events-none"
            />
            <g className="pointer-events-none">
              <circle cx={0} cy={0} r={15} fill="#0f172a" stroke={selected ? style.ring : style.fill} strokeWidth={selected ? 2.5 : 1.5} />
              <Icon x={-8} y={-8} width={16} height={16} stroke={style.fill} strokeWidth={2} />
              <circle cx={10.5} cy={-10.5} r={3.5} fill={active ? '#22c55e' : '#64748b'} stroke="#0f172a" strokeWidth={1.5} />
              <text x={0} y={27} textAnchor="middle" fill="#e2e8f0" fontSize={10} fontWeight={600}>
                {name}
              </text>
              {dev?.state && (
                <text x={0} y={37} textAnchor="middle" fill="#94a3b8" fontSize={8.5} fontFamily="ui-monospace, monospace">
                  {dev.state}
                </text>
              )}
            </g>
          </g>
        )
      })}
    </g>
  )
}
