import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { LayerSelector } from './LayerSelector'
import type { Layer } from '../types'

describe('LayerSelector', () => {
  const defaultLayers: Layer[] = [
    { name: 'controls', hide_gauges: false },
    { name: 'sensors', hide_gauges: false },
  ]

  const customLayers: Layer[] = [
    { name: 'controls', hide_gauges: false },
    { name: 'security', hide_gauges: false },
    { name: 'climate', hide_gauges: false },
  ]

  it('renders default layers ("Controls" and "Sensors") when layers prop is empty or undefined', () => {
    const handleSelect = vi.fn()
    render(<LayerSelector activeLayer="controls" onSelectLayer={handleSelect} />)

    expect(screen.getByRole('button', { name: /controls/i })).toBeDefined()
    expect(screen.getByRole('button', { name: /sensors/i })).toBeDefined()
  })

  it('renders custom layers passed in props with capitalized labels', () => {
    const handleSelect = vi.fn()
    render(
      <LayerSelector
        layers={customLayers}
        activeLayer="controls"
        onSelectLayer={handleSelect}
      />
    )

    expect(screen.getByText('Controls')).toBeDefined()
    expect(screen.getByText('Security')).toBeDefined()
    expect(screen.getByText('Climate')).toBeDefined()
  })

  it('marks the active layer button with active styling and aria-pressed=true', () => {
    const handleSelect = vi.fn()
    render(
      <LayerSelector
        layers={defaultLayers}
        activeLayer="sensors"
        onSelectLayer={handleSelect}
      />
    )

    const sensorsBtn = screen.getByRole('button', { name: /sensors/i })
    const controlsBtn = screen.getByRole('button', { name: /controls/i })

    expect(sensorsBtn.getAttribute('aria-pressed')).toBe('true')
    expect(sensorsBtn.className).toContain('bg-indigo-600')
    expect(sensorsBtn.className).toContain('text-white')

    expect(controlsBtn.getAttribute('aria-pressed')).toBe('false')
    expect(controlsBtn.className).toContain('text-slate-400')
  })

  it('calls onSelectLayer with the clicked layer name when an inactive button is clicked', () => {
    const handleSelect = vi.fn()
    render(
      <LayerSelector
        layers={[
          { name: 'controls', hide_gauges: false },
          { name: 'sensors', hide_gauges: false },
          { name: 'security', hide_gauges: false },
        ]}
        activeLayer="controls"
        onSelectLayer={handleSelect}
      />
    )

    const securityBtn = screen.getByRole('button', { name: /security/i })
    fireEvent.click(securityBtn)

    expect(handleSelect).toHaveBeenCalledTimes(1)
    expect(handleSelect).toHaveBeenCalledWith('security')
  })

  it('does not re-trigger or calls onSelectLayer even if already active button is clicked', () => {
    const handleSelect = vi.fn()
    render(
      <LayerSelector
        layers={defaultLayers}
        activeLayer="controls"
        onSelectLayer={handleSelect}
      />
    )

    const controlsBtn = screen.getByRole('button', { name: /controls/i })
    fireEvent.click(controlsBtn)

    expect(handleSelect).toHaveBeenCalledWith('controls')
  })
})