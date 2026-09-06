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

// Sensor widget, Display "arc": the single circular rendering in this
// dashboard, a dial open at the bottom (roughly 270 degrees), never a
// full-circle gauge. Drawn as an SVG with a square viewBox so
// `preserveAspectRatio` (the SVG default, xMidYMid meet) always centres
// it in a square regardless of the cell's actual aspect ratio (ADR 0004).
export const ArcWidget: React.FC<ArcWidgetProps> = ({ label, value, min, max, unit, stale }) => {
  const ratio = clampRatio(value, min, max)
  const trackPath = describeArc(START_ANGLE_DEG, SWEEP_DEG)
  const valuePath = ratio > 0 ? describeArc(START_ANGLE_DEG, ratio * SWEEP_DEG) : ''
  const formatted = formatSensorValue(value, unit)

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      icon={<CircleGauge className="text-cyan-400" />}
      bodyClassName="flex items-center justify-center"
    >
      <svg
        viewBox="0 0 100 100"
        className="w-full h-full"
        role="img"
        aria-label={`${label}: ${formatted.text}${formatted.unit ? ` ${formatted.unit}` : ''}`}
      >
        <path d={trackPath} fill="none" stroke="#1e293b" strokeWidth={9} strokeLinecap="round" />
        {valuePath && <path d={valuePath} fill="none" stroke="#22d3ee" strokeWidth={9} strokeLinecap="round" />}
        <text
          x={CENTER}
          y={CENTER}
          textAnchor="middle"
          dominantBaseline="central"
          className="fill-white font-extrabold tabular-nums"
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
    </WidgetFrame>
  )
}
