import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { NumberWidget } from './NumberWidget'

function valueFontSize(text: string): string {
  return (screen.getByText(text).parentElement as HTMLElement).style.fontSize
}

describe('NumberWidget', () => {
  it('scales the value font with the widget size instead of capping it at a fixed size', () => {
    render(<NumberWidget label="Salon" value={24} unit="°C" stale={false} />)

    const fontSize = valueFontSize('24')
    expect(fontSize).toContain('cqh')
    expect(fontSize).toContain('cqw')
    // No fixed rem upper bound: the widget's own size is the only ceiling.
    expect(fontSize).not.toMatch(/,\s*[\d.]+rem\)$/)
  })

  it('keeps scaling the value font when bounds are present', () => {
    render(
      <NumberWidget
        label="Salon"
        value={24}
        unit="°C"
        min={10}
        max={30}
        stale={false}
      />,
    )

    const fontSize = valueFontSize('24')
    expect(fontSize).toContain('cqh')
    expect(fontSize).toContain('cqw')
    expect(fontSize).not.toMatch(/,\s*[\d.]+rem\)$/)
  })
})
