import React from 'react'
import { CircleGauge } from 'lucide-react'
import { formatSensorValue } from '../format'
import { WidgetFrame } from './WidgetFrame'

export interface ArcWidgetProps {
  label: string
  value: number | null
  min: number
  max: number
  unit?: string
  stale: boolean
}

const CENTER = 50
const RADIUS = 40
const START_ANGLE_DEG = 135
const SWEEP_DEG = 270

function polarToCartesian(centerDeg: number) {
  const rad = (centerDeg * Math.PI) / 180
  return {
    x: CENTER + RADIUS * Math.cos(rad),
    y: CENTER + RADIUS * Math.sin(rad),
  }
}

function describeArc(startDeg: number, sweepDeg: number): string {
  const start = polarToCartesian(startDeg)
  const end = polarToCartesian(startDeg + sweepDeg)
  const largeArcFlag = sweepDeg > 180 ? 1 : 0
  return `M ${start.x} ${start.y} A ${RADIUS} ${RADIUS} 0 ${largeArcFlag} 1 ${end.x} ${end.y}`
}

function clampRatio(value: number | null, min: number, max: number): number {
  if (value === null || max <= min) return 0
  return Math.min(1, Math.max(0, (value - min) / (max - min)))
}

// Sensor widget, Display "arc": the single circular rendering in this dashboard, a dial open at
// the bottom (roughly 270 degrees), never a full-circle gauge. Drawn as an SVG with a square
// viewBox so `preserveAspectRatio` (the SVG default, xMidYMid meet) always centres it in a square
// regardless of the cell's actual aspect ratio (ADR 0004). Variant A adds the reading's unit as a
// caption chip and the min/max bounds under the dial.
export const ArcWidget: React.FC<ArcWidgetProps> = ({ label, value, min, max, unit, stale }) => {
  const ratio = clampRatio(value, min, max)
  const trackPath = describeArc(START_ANGLE_DEG, SWEEP_DEG)
  const valuePath = ratio > 0 ? describeArc(START_ANGLE_DEG, ratio * SWEEP_DEG) : ''
  const formatted = formatSensorValue(value, unit)
  const bounds = { min: formatSensorValue(min, unit), max: formatSensorValue(max, unit) }

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      unit={unit}
      icon={<CircleGauge className="text-[#4bb8c9]" />}
      bodyClassName="flex flex-col"
    >
      <div className="h-full w-full flex flex-col min-h-0">
        <div className="flex-1 min-h-0 flex items-center justify-center">
          <svg
            viewBox="0 0 100 100"
            className="w-full h-full"
            role="img"
            aria-label={`${label}: ${formatted.text}${formatted.unit ? ` ${formatted.unit}` : ''}`}
          >
            <path d={trackPath} fill="none" stroke="#1e293b" strokeWidth={9} strokeLinecap="round" />
            {valuePath && <path d={valuePath} fill="none" stroke="#4bb8c9" strokeWidth={9} strokeLinecap="round" />}
            <text
              x={CENTER}
              y={CENTER}
              textAnchor="middle"
              dominantBaseline="central"
              className="fill-white font-semibold tracking-tight tabular-nums"
              style={{ fontSize: formatted.unit ? 20 : 24 }}
            >
              {formatted.text}
            </text>
            {formatted.unit && (
              <text
                x={CENTER}
                y={CENTER + 15}
                textAnchor="middle"
                dominantBaseline="central"
                className="fill-slate-400 font-medium"
                style={{ fontSize: 9 }}
              >
                {formatted.unit}
              </text>
            )}
          </svg>
        </div>
        <div className="shrink-0 flex justify-between px-0.5 text-[10px] leading-none text-slate-500 tabular-nums">
          <span>
            {bounds.min.text}
            {bounds.min.unit}
          </span>
          <span>
            {bounds.max.text}
            {bounds.max.unit}
          </span>
        </div>
      </div>
    </WidgetFrame>
  )
}
