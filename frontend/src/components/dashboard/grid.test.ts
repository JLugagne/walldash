import { describe, it, expect } from 'vitest'
import { mobileFlowItems, MOBILE_COLUMNS, type Rect } from './grid'

function rect(id: string, row: number, col: number, colSpan: number, rowSpan: number): Rect {
  return { id, row, col, colSpan, rowSpan }
}

describe('mobileFlowItems', () => {
  it('orders widgets by reading order (row, then column)', () => {
    const items = mobileFlowItems([
      rect('c', 1, 0, 1, 1),
      rect('a', 0, 1, 1, 1),
      rect('b', 0, 0, 3, 1),
    ])

    expect(items.map((i) => i.rect.id)).toEqual(['b', 'a', 'c'])
  })

  it('collapses a single-column widget to half width and anything wider to full width', () => {
    const items = mobileFlowItems([
      rect('one', 0, 0, 1, 1),
      rect('two', 0, 1, 2, 2),
      rect('wide', 0, 3, 6, 1),
    ])
    const colsById = new Map(items.map((i) => [i.rect.id, i.mobileCols]))

    expect(colsById.get('one')).toBe(1)
    expect(colsById.get('two')).toBe(MOBILE_COLUMNS)
    expect(colsById.get('wide')).toBe(MOBILE_COLUMNS)
  })

  it('keeps the row span to drive the height and never mutates the input', () => {
    const input = [rect('tall', 0, 0, 1, 3)]
    const items = mobileFlowItems(input)

    expect(items[0].rect.rowSpan).toBe(3)
    expect(input[0]).toEqual(rect('tall', 0, 0, 1, 3))
  })
})
