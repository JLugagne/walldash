import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { MobileWidgetFlow } from './MobileWidgetFlow'
import type { Dashboard, Widget } from '../../types'

function widget(id: string, col: number, row: number, colSpan: number, rowSpan: number): Widget {
  return {
    id,
    dashboard_id: 'd1',
    type: 'sensor',
    title: id,
    order: 0,
    col,
    row,
    col_span: colSpan,
    row_span: rowSpan,
    config: { entity_ids: [], display: 'number' },
    created_at: '',
    updated_at: '',
  }
}

function dashboard(widgets: Widget[]): Dashboard {
  return {
    id: 'd1',
    name: 'Home',
    order: 0,
    cols: 12,
    rows: 5,
    background_image: '',
    background_opacity: 0,
    background_blur: 0,
    background_dim: 0,
    created_at: '',
    updated_at: '',
    widgets,
  }
}

function renderFlow(widgets: Widget[]) {
  return render(
    <MobileWidgetFlow dashboard={dashboard(widgets)} renderWidgetBody={(w) => <div>{w.id}</div>} />,
  )
}

function flowCells(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>('div[style*="grid-column"]'))
}

describe('MobileWidgetFlow', () => {
  it('renders widgets in desktop reading order', () => {
    const { container } = renderFlow([
      widget('c', 0, 1, 1, 1),
      widget('a', 1, 0, 1, 1),
      widget('b', 0, 0, 3, 1),
    ])

    expect(flowCells(container).map((cell) => cell.textContent)).toEqual(['b', 'a', 'c'])
  })

  it('maps a single-cell widget to half width and wider widgets to full width', () => {
    const { container } = renderFlow([widget('half', 0, 0, 1, 1), widget('full', 1, 0, 4, 2)])
    const [half, full] = flowCells(container)

    expect(half.style.gridColumn).toContain('span 1')
    expect(full.style.gridColumn).toContain('span 2')
    expect(full.style.gridRow).toContain('span 2')
  })
})
