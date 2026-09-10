import React from 'react'
import { StaleBadge } from './StaleBadge'

export interface WidgetFrameProps {
  label: string
  stale: boolean
  /** Display Mode glyph, already coloured by the caller; rendered at 12 px left of the label. */
  icon?: React.ReactNode
  /** Right-aligned caption content (for example a Bar's value) shown before the StaleBadge. */
  trailing?: React.ReactNode
  /**
   * Compact chrome for 1-row Widgets: on a 12x8 tablet grid a row is about 75 px and the
   * caption + paddings must leave at least 40 px for the body.
   */
  dense?: boolean
  /** Extra classes for the body element, which is always `relative flex-1 min-h-0`. */
  bodyClassName?: string
  children: React.ReactNode
}

/**
 * WidgetFrame is the single card chrome shared by every Widget: a borderless 18 px caption row
 * (label, optional icon, optional trailing content, StaleBadge) followed by a body that takes the
 * remaining height and clips its overflow. The body is a size container, so children may size
 * text with `cqh` / `cqw` units without a ResizeObserver. Widgets render only their body.
 */
export const WidgetFrame: React.FC<WidgetFrameProps> = ({
  label,
  stale,
  icon,
  trailing,
  dense = false,
  bodyClassName = '',
  children,
}) => (
  <div
    className={`w-full h-full flex flex-col bg-slate-900/55 backdrop-blur-md border border-slate-800/80 shadow-sm overflow-hidden ${
      dense ? 'rounded-lg px-2 py-1 gap-0.5' : 'rounded-xl px-3 py-2 gap-1'
    }`}
  >
    <div className="h-[18px] shrink-0 flex items-center justify-between gap-1.5 min-w-0">
      <div className="flex items-center gap-1 min-w-0">
        {icon && <span className="w-3 h-3 shrink-0 flex items-center justify-center [&>svg]:w-3 [&>svg]:h-3">{icon}</span>}
        <span className="text-[11px] leading-none font-semibold uppercase tracking-wider text-slate-400 truncate" title={label}>
          {label}
        </span>
      </div>
      <div className="flex items-center gap-1 shrink-0">
        {trailing}
        {stale && <StaleBadge />}
      </div>
    </div>
    <div className={`relative flex-1 min-h-0 min-w-0 overflow-hidden [container-type:size] ${bodyClassName}`}>
      {children}
    </div>
  </div>
)
