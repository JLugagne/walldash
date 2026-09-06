import React from 'react'
import { Gauge } from 'lucide-react'
import { formatSensorValue } from '../format'
import { WidgetFrame } from './WidgetFrame'

export interface NumberWidgetProps {
  label: string
  /** Raw state: a number, a non-numeric HA state such as `cloudy`, or null when nothing is known. */
  value: number | string | null
  unit?: string
  stale: boolean
  dense?: boolean
}

const CHAR_WIDTH_EM = 0.62
const UNIT_SCALE = 0.55
const UNIT_GAP_EM = 0.25
const MIN_FONT = '0.625rem'
const MAX_FONT = '2.25rem'

function valueFontSize(text: string, unit: string): string {
  const unitWidthEm = unit ? UNIT_GAP_EM + unit.length * CHAR_WIDTH_EM * UNIT_SCALE : 0
  const lineWidthEm = Math.max(text.length * CHAR_WIDTH_EM + unitWidthEm, CHAR_WIDTH_EM)
  const widthFit = (100 / lineWidthEm).toFixed(1)
  return `clamp(${MIN_FONT}, min(45cqh, ${widthFit}cqw), ${MAX_FONT})`
}

// Sensor widget, Display "number": a single centred value line whose font
// follows the body size through container-query units, so a 1x1 cell and a
// 4x2 cell both fill without clipping. The width term is scaled by the line's
// character count so a long word such as "Cloudy" shrinks to fit a 1x1 cell
// instead of being ellipsized. Presentational only — the caller resolves the
// label, the last known value, and whether it is stale.
export const NumberWidget: React.FC<NumberWidgetProps> = ({ label, value, unit, stale, dense = false }) => {
  const formatted = formatSensorValue(value, unit)

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      dense={dense}
      icon={<Gauge className="text-emerald-400" />}
      bodyClassName="flex items-center justify-center"
    >
      <div
        className="flex items-baseline justify-center gap-1 min-w-0 max-w-full whitespace-nowrap"
        style={{ fontSize: valueFontSize(formatted.text, formatted.unit) }}
      >
        <span className="min-w-0 font-extrabold text-white leading-none tabular-nums truncate">{formatted.text}</span>
        {formatted.unit && (
          <span className="shrink-0 font-medium text-slate-400 leading-none text-[0.55em]">{formatted.unit}</span>
        )}
      </div>
    </WidgetFrame>
  )
}
