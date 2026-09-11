import React from 'react'
import { Gauge } from 'lucide-react'
import { formatSensorValue } from '../format'
import { WidgetFrame } from './WidgetFrame'
import { Sparkline } from './Sparkline'

export interface NumberWidgetProps {
  label: string
  /** Raw state: a number, a non-numeric HA state such as `cloudy`, or null when nothing is known. */
  value: number | string | null
  unit?: string
  min?: number
  max?: number
  /** Recent samples for the optional sparkline; nothing is drawn below two samples. */
  history?: number[]
  stale: boolean
  dense?: boolean
}

const CHAR_WIDTH_EM = 0.62
const UNIT_SCALE = 0.55
const UNIT_GAP_EM = 0.25
const MIN_FONT = '0.625rem'
const MAX_FONT = '2.25rem'

function valueFontSize(text: string, unit: string, heightCap: number): string {
  const unitWidthEm = unit ? UNIT_GAP_EM + unit.length * CHAR_WIDTH_EM * UNIT_SCALE : 0
  const lineWidthEm = Math.max(text.length * CHAR_WIDTH_EM + unitWidthEm, CHAR_WIDTH_EM)
  const widthFit = (100 / lineWidthEm).toFixed(1)
  return `clamp(${MIN_FONT}, min(${heightCap}cqh, ${widthFit}cqw), ${MAX_FONT})`
}

// Sensor widget, Display "number": a single centred value line whose font follows the body size
// through container-query units, so a 1x1 cell and a 4x2 cell both fill without clipping. The
// width term is scaled by the line's character count so a long word such as "Cloudy" shrinks to
// fit a 1x1 cell instead of being ellipsized. Variant A adds the reading's dimension as a caption
// chip, explicit min/max bounds when the Widget config sets them, and a sparkline of the recent
// samples the caller has collected. Presentational only — the caller resolves the label, the last
// known value, its history, and whether it is stale.
export const NumberWidget: React.FC<NumberWidgetProps> = ({
  label,
  value,
  unit,
  min,
  max,
  history = [],
  stale,
  dense = false,
}) => {
  const formatted = formatSensorValue(value, unit)
  const hasBounds = min !== undefined && max !== undefined
  const hasHistory = history.length >= 2
  const boundsMin = hasBounds ? formatSensorValue(min, unit) : null
  const boundsMax = hasBounds ? formatSensorValue(max, unit) : null
  const heightCap = hasBounds || hasHistory ? 38 : 45

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      dense={dense}
      unit={unit}
      icon={<Gauge className="text-[#42a67d]" />}
      bodyClassName="flex flex-col"
    >
      <div className="h-full w-full flex flex-col min-h-0">
        <div className="flex-1 min-h-0 flex items-center justify-center">
          <div
            className="flex items-baseline justify-center gap-1 min-w-0 max-w-full whitespace-nowrap"
            style={{ fontSize: valueFontSize(formatted.text, formatted.unit, heightCap) }}
          >
            <span className="min-w-0 font-semibold tracking-tight text-white leading-none tabular-nums truncate">
              {formatted.text}
            </span>
            {formatted.unit && (
              <span className="shrink-0 font-medium text-slate-400 leading-none text-[0.55em]">{formatted.unit}</span>
            )}
          </div>
        </div>
        {boundsMin && boundsMax && (
          <div className="shrink-0 flex items-center justify-between px-0.5 text-[10px] leading-none text-slate-500 tabular-nums">
            <span>
              {boundsMin.text}
              {boundsMin.unit}
            </span>
            <span>
              {boundsMax.text}
              {boundsMax.unit}
            </span>
          </div>
        )}
        {hasHistory && <Sparkline history={history} color="#42a67d" className="w-full h-5 shrink-0 mt-0.5" />}
      </div>
    </WidgetFrame>
  )
}
