import type { Widget } from '../../types'
import type { Rect } from './grid'

export function widgetsToRects(widgets: Widget[]): Rect[] {
  return widgets.map((w) => ({ id: w.id, col: w.col, row: w.row, colSpan: w.col_span, rowSpan: w.row_span }))
}

export interface FormattedSensorValue {
  text: string
  unit: string
}

const NUMBER_FORMAT = new Intl.NumberFormat('fr-FR', { maximumFractionDigits: 1 })

const EMPTY_STATES = new Set(['', 'unavailable', 'unknown', 'none'])

/**
 * formatSensorValue turns a raw Home Assistant state into what a Widget prints.
 * Numbers are rounded to one decimal in French locale (`21.3456` -> `21,3`) and keep the unit;
 * `null`, an empty string and the HA sentinels `unavailable` / `unknown` / `none` become `—` and
 * keep the unit so a stale temperature still reads as a temperature; any other text (`cloudy`)
 * is passed through with a capital first letter and drops the unit, which only makes sense next
 * to a number. Pure and allocation-light: safe to call on every render.
 */
export function formatSensorValue(state: string | number | null | undefined, unit?: string): FormattedSensorValue {
  const safeUnit = unit ?? ''
  if (state === null || state === undefined) return { text: '—', unit: safeUnit }
  if (typeof state === 'number') {
    return Number.isFinite(state) ? { text: NUMBER_FORMAT.format(state), unit: safeUnit } : { text: '—', unit: safeUnit }
  }
  const trimmed = state.trim()
  if (EMPTY_STATES.has(trimmed.toLowerCase())) return { text: '—', unit: safeUnit }
  const asNumber = Number(trimmed)
  if (Number.isFinite(asNumber)) return { text: NUMBER_FORMAT.format(asNumber), unit: safeUnit }
  return { text: trimmed.charAt(0).toUpperCase() + trimmed.slice(1), unit: '' }
}
