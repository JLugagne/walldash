import React from 'react'
import { BarChart3 } from 'lucide-react'
import { formatSensorValue } from '../format'
import { WidgetFrame } from './WidgetFrame'

export interface BarWidgetProps {
  label: string
  value: number | null
  min: number
  max: number
  unit?: string
  stale: boolean
  dense?: boolean
}

function clampRatio(value: number | null, min: number, max: number): number {
  if (value === null || max <= min) return 0
  const ratio = (value - min) / (max - min)
  return Math.min(1, Math.max(0, ratio))
}

// Sensor widget, Display "bar": a large reading over a horizontal fill between the Widget's Min
// and Max, with the bounds called out underneath. The track alone carries the meter semantics so
// the value stays a plain readout. Presentational only — bounds and freshness are resolved by the
// caller.
export const BarWidget: React.FC<BarWidgetProps> = ({
  label,
  value,
  min,
  max,
  unit,
  stale,
  dense = false,
}) => {
  const ratio = clampRatio(value, min, max)
  const formatted = formatSensorValue(value, unit)
  const bounds = { min: formatSensorValue(min, unit), max: formatSensorValue(max, unit) }

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      dense={dense}
      unit={unit}
      icon={<BarChart3 className="text-[#4bb8c9]" />}
      bodyClassName="flex flex-col justify-center gap-1"
    >
      <div
        className="flex items-baseline gap-1 min-w-0 whitespace-nowrap font-semibold tracking-tight text-white tabular-nums"
        style={{ fontSize: 'max(1rem, min(38cqh, 24cqw))' }}
      >
        <span className="truncate">{formatted.text}</span>
        {formatted.unit && <span className="shrink-0 font-medium text-slate-400 text-[0.55em]">{formatted.unit}</span>}
      </div>
      <div
        className="w-full h-2 shrink-0 rounded-md bg-slate-800 border border-slate-700 overflow-hidden"
        role="meter"
        aria-label={label}
        aria-valuemin={min}
        aria-valuemax={max}
        aria-valuenow={value ?? undefined}
      >
        <div className="h-full rounded-md bg-[#4bb8c9] transition-[width] duration-300" style={{ width: `${ratio * 100}%` }} />
      </div>
      {!dense && (
        <div className="shrink-0 flex justify-between text-[10px] leading-none text-slate-500 tabular-nums">
          <span>
            {bounds.min.text}
            {bounds.min.unit}
          </span>
          <span>
            {bounds.max.text}
            {bounds.max.unit}
          </span>
        </div>
      )}
    </WidgetFrame>
  )
}
