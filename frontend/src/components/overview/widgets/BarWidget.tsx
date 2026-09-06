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

// Sensor widget, Display "bar": a horizontal fill between admin-entered
// Min and Max. The value lives in the caption so the body is the 8 px
// track alone and a 2x1 cell fits; the min/max legend only appears when
// the widget spans several rows. Presentational only — bounds and
// freshness are resolved by the caller.
export const BarWidget: React.FC<BarWidgetProps> = ({ label, value, min, max, unit, stale, dense = false }) => {
  const ratio = clampRatio(value, min, max)
  const formatted = formatSensorValue(value, unit)
  const bounds = { min: formatSensorValue(min).text, max: formatSensorValue(max).text }

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      dense={dense}
      icon={<BarChart3 className="text-cyan-400" />}
      trailing={
        <span className="flex items-baseline gap-0.5 whitespace-nowrap">
          <span className="text-xs leading-none font-bold text-white tabular-nums">{formatted.text}</span>
          {formatted.unit && <span className="text-[10px] leading-none font-medium text-slate-400">{formatted.unit}</span>}
        </span>
      }
      bodyClassName="flex flex-col justify-center gap-1"
    >
      <div
        className="w-full h-2 rounded-full bg-slate-800 border border-slate-700 overflow-hidden"
        role="meter"
        aria-label={label}
        aria-valuemin={min}
        aria-valuemax={max}
        aria-valuenow={value ?? undefined}
      >
        <div className="h-full rounded-full bg-cyan-500 transition-[width] duration-300" style={{ width: `${ratio * 100}%` }} />
      </div>
      {!dense && (
        <div className="flex justify-between text-[9px] leading-none text-slate-500 tabular-nums">
          <span>{bounds.min}</span>
          <span>{bounds.max}</span>
        </div>
      )}
    </WidgetFrame>
  )
}
