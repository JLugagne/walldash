import React from 'react'
import type { Dashboard, Widget } from '../../types'
import { widgetsToRects } from './format'
import { mobileFlowItems } from './grid'

export interface MobileWidgetFlowProps {
  dashboard: Dashboard
  renderWidgetBody: (widget: Widget) => React.ReactNode
}

/** Flow row height in pixels; a Widget's desktop row_span multiplies it. Chosen so a 1x1 tile
 * stays tappable (~180x120 on a 390 px phone) while 2-row Widgets keep their proportion. */
const FLOW_ROW_PX = 120

/**
 * MobileWidgetFlow is the phone layout of a Dashboard: one vertical scroll of Widgets laid out in
 * at most two columns. It replaces the fixed Widget Grid below the `sm` breakpoint (see
 * `useIsMobile`), where the persisted 12-column grid would collapse into unusable strips.
 *
 * Widgets keep the reading order of the desktop layout; their column span maps to one or two flow
 * columns and their row span drives the height. The flow is read-only — phone users view and tap,
 * layout is authored on a tablet or desktop — and the page scrolls vertically, an exception to the
 * no-scroll invariant of ADR 0004 that applies to the fixed grid only.
 */
export const MobileWidgetFlow: React.FC<MobileWidgetFlowProps> = ({ dashboard, renderWidgetBody }) => {
  const widgetById = new Map(dashboard.widgets.map((w) => [w.id, w]))
  const items = mobileFlowItems(widgetsToRects(dashboard.widgets))

  return (
    <div className="h-full overflow-y-auto overscroll-contain pb-2">
      <div
        className="grid grid-cols-2 gap-3"
        style={{ gridAutoRows: `${FLOW_ROW_PX}px` }}
      >
        {items.map(({ rect, mobileCols }) => {
          const widget = widgetById.get(rect.id)
          if (!widget) return null
          return (
            <div
              key={widget.id}
              className="min-w-0 min-h-0"
              style={{ gridColumn: `span ${mobileCols}`, gridRow: `span ${rect.rowSpan}` }}
            >
              {renderWidgetBody(widget)}
            </div>
          )
        })}
      </div>
    </div>
  )
}
