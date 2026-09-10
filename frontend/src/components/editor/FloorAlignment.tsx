import type { PointerEvent as ReactPointerEvent } from 'react'
import { Eye, EyeOff, RotateCcw } from 'lucide-react'
import type { Level, Plan } from '../../types'
import type { HouseOverviewConfig } from '../../utils/houseOverview'
import { configForLevel, orderedLevels } from '../../utils/houseOverview'

interface AlignmentLayerProps {
  levels: Level[]
  plans: Record<string, Plan>
  config: HouseOverviewConfig
  activeLevelId: string | null
  upx: number
  onFloorPointerDown: (levelId: string, event: ReactPointerEvent<SVGGElement>) => void
}

function labelAnchor(plan: Plan | undefined): { x: number; y: number } {
  if (!plan || plan.walls.length === 0) return { x: 0, y: 0 }
  let minX = Infinity
  let maxX = -Infinity
  let minY = Infinity
  for (const wall of plan.walls) {
    minX = Math.min(minX, wall.x1, wall.x2)
    maxX = Math.max(maxX, wall.x1, wall.x2)
    minY = Math.min(minY, wall.y1, wall.y2)
  }
  return { x: (minX + maxX) / 2, y: minY }
}

export function AlignmentLayer({ levels, plans, config, activeLevelId, upx, onFloorPointerDown }: AlignmentLayerProps) {
  const activeStroke = 10 * upx
  const idleStroke = 6 * upx
  const hitStroke = 26 * upx

  return (
    <g>
      {orderedLevels(levels).map((level) => {
        const floor = configForLevel(config, level.id)
        if (!floor.visible) return null
        const plan = plans[level.id]
        const active = level.id === activeLevelId
        const label = active ? labelAnchor(plan) : null
        return (
          <g
            key={level.id}
            transform={`translate(${floor.x} ${floor.y})`}
            className="cursor-move"
            onPointerDown={(event) => onFloorPointerDown(level.id, event)}
          >
            {(plan?.walls || []).map((wall) => (
              <g key={wall.id}>
                <line
                  x1={wall.x1}
                  y1={wall.y1}
                  x2={wall.x2}
                  y2={wall.y2}
                  stroke="transparent"
                  strokeWidth={hitStroke}
                  strokeLinecap="round"
                  pointerEvents="stroke"
                />
                <line
                  x1={wall.x1}
                  y1={wall.y1}
                  x2={wall.x2}
                  y2={wall.y2}
                  stroke={active ? '#818cf8' : '#94a3b8'}
                  strokeWidth={active ? activeStroke : idleStroke}
                  strokeLinecap="round"
                  opacity={active ? 1 : 0.4}
                  pointerEvents="none"
                />
              </g>
            ))}
            {label && (
              <text
                x={label.x}
                y={label.y - 16 * upx}
                fill="#c7d2fe"
                fontSize={18 * upx}
                textAnchor="middle"
                className="pointer-events-none select-none"
              >
                {level.name}
              </text>
            )}
          </g>
        )
      })}
    </g>
  )
}

interface FloorAlignmentPanelProps {
  levels: Level[]
  config: HouseOverviewConfig
  activeLevelId: string | null
  onSelect: (id: string) => void
  onChange: (config: HouseOverviewConfig) => void
  onReset: () => void
}

export function FloorAlignmentPanel({ levels, config, activeLevelId, onSelect, onChange, onReset }: FloorAlignmentPanelProps) {
  const ordered = orderedLevels(levels)
  const selected = ordered.find((level) => level.id === activeLevelId) || ordered[0]
  const selectedConfig = selected ? configForLevel(config, selected.id) : { x: 0, y: 0, visible: true }

  const update = (id: string, patch: Partial<ReturnType<typeof configForLevel>>) => {
    onChange({ ...config, [id]: { ...configForLevel(config, id), ...patch } })
  }

  if (!selected) return <div className="p-4 text-xs text-slate-500">No floors configured.</div>

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center justify-between border-b border-slate-800 px-3 py-2">
        <span className="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Floors</span>
        <button
          type="button"
          onClick={onReset}
          className="flex items-center gap-1 rounded-md px-2 py-1 text-[11px] text-slate-400 transition-colors hover:bg-slate-800 hover:text-white"
        >
          <RotateCcw className="h-3 w-3" /> Reset
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-2">
        {ordered.map((level) => {
          const floor = configForLevel(config, level.id)
          const active = level.id === selected.id
          return (
            <div
              key={level.id}
              className={`mb-1 flex items-center gap-1 rounded-lg ${active ? 'bg-indigo-600 text-white' : 'text-slate-300 hover:bg-slate-800'}`}
            >
              <button
                type="button"
                onClick={() => update(level.id, { visible: !floor.visible })}
                title={floor.visible ? 'Hide in overview' : 'Show in overview'}
                className="p-2"
              >
                {floor.visible ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5 text-slate-500" />}
              </button>
              <button type="button" onClick={() => onSelect(level.id)} className="flex-1 truncate py-2 text-left text-xs font-medium">
                {level.name}
              </button>
              <span className="pr-2 font-mono text-[10px] opacity-60">{level.order + 1}</span>
            </div>
          )
        })}
      </div>
      <div className="space-y-3 border-t border-slate-800 p-3">
        <p className="text-xs font-semibold text-white">{selected.name}</p>
        <label className="block text-[11px] text-slate-400">
          Horizontal offset
          <input
            type="number"
            value={selectedConfig.x}
            onChange={(event) => update(selected.id, { x: Number(event.target.value) || 0 })}
            className="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-2 py-1.5 text-xs text-white"
          />
        </label>
        <label className="block text-[11px] text-slate-400">
          Vertical offset
          <input
            type="number"
            value={selectedConfig.y}
            onChange={(event) => update(selected.id, { y: Number(event.target.value) || 0 })}
            className="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-2 py-1.5 text-xs text-white"
          />
        </label>
        <p className="text-[11px] leading-relaxed text-slate-500">Drag a floor on the canvas to align it with the others.</p>
      </div>
    </div>
  )
}
