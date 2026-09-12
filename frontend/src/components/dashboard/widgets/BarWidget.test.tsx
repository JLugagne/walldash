import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BarWidget } from './BarWidget'

describe('BarWidget', () => {
  it('scales the value font with the widget size instead of capping it at a fixed size', () => {
    render(<BarWidget label="Salon" value={24} min={10} max={30} unit="°C" stale={false} />)

    const fontSize = (screen.getByText('24').parentElement as HTMLElement).style.fontSize
    expect(fontSize).toContain('cqh')
    expect(fontSize).toContain('cqw')
    // No fixed rem upper bound: the widget's own size is the only ceiling.
    expect(fontSize).not.toMatch(/,\s*[\d.]+rem\)$/)
  })
})
