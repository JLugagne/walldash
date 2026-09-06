import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { IsometricView } from './IsometricView'
import type { Level, DevicePlacement, Layer } from '../types'

vi.mock('@react-three/fiber', () => ({
  Canvas: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="mock-canvas">
      {React.Children.map(children, (child) => {
        if (React.isValidElement(child) && typeof child.type === 'string' && child.type === 'color') {
          return null
        }
        return child
      })}
    </div>
  ),
}))

let lastSceneProps: any = null
vi.mock('./IsometricScene', () => ({
  IsometricScene: (props: any) => {
    lastSceneProps = props
    return <div data-testid="mock-isometric-scene" />
  },
}))

const mockPlacements: DevicePlacement[] = [
  { id: 'p1', level_id: 'l1', device_id: 'light.1', x: 0, y: 0, layer: 'controls' },
  { id: 'p2', level_id: 'l1', device_id: 'sensor.1', x: 0, y: 0, layer: 'sensors' },
  { id: 'p3', level_id: 'l1', device_id: 'camera.1', x: 0, y: 0, layer: 'security' },
]

vi.mock('../hooks/useRealtimeDevices', () => ({
  useRealtimeDevices: () => ({
    deviceMap: {},
    placements: mockPlacements,
    pendingDevices: {},
    toggleDevice: vi.fn(),
  }),
}))

function renderInRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

describe('IsometricView Layer Selection and Filtering', () => {
  const defaultLayers: Layer[] = [
    { name: 'controls', hide_gauges: false },
    { name: 'sensors', hide_gauges: false },
    { name: 'security', hide_gauges: false },
  ]

  const gardenLayers: Layer[] = [
    { name: 'garden_sensors', hide_gauges: false },
    { name: 'irrigation', hide_gauges: false },
  ]

  const level1: Level = {
    id: 'l1',
    name: 'Ground Floor',
    order: 0,
    is_outdoor: false,
    layers: defaultLayers,
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
  }

  const level2: Level = {
    id: 'l2',
    name: 'Jardin',
    order: 1,
    is_outdoor: true,
    layers: gardenLayers,
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
  }

  beforeEach(() => {
    localStorage.clear()
    lastSceneProps = null
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'success', data: { level_id: 'l1', walls: [], zones: [] } }),
    }))
  })

  it('renders LayerSelector at top center with level layers', async () => {
    renderInRouter(
      <IsometricView
        level={level1}
        levels={[level1]}
        onSelectLevel={vi.fn()}
      />
    )

    await waitFor(() => {
      expect(screen.getByRole('group', { name: /display layers/i })).toBeDefined()
    })
    expect(screen.getByRole('button', { name: /controls/i })).toBeDefined()
    expect(screen.getByRole('button', { name: /sensors/i })).toBeDefined()
    expect(screen.getByRole('button', { name: /security/i })).toBeDefined()
  })

  it('filters placements passed to IsometricScene or passes activeLayer', async () => {
    renderInRouter(
      <IsometricView
        level={level1}
        levels={[level1]}
        onSelectLevel={vi.fn()}
      />
    )

    expect(lastSceneProps).not.toBeNull()
    expect(lastSceneProps.activeLayer).toBe('controls')
    if (lastSceneProps.placements) {
      const visible = lastSceneProps.activeLayer
        ? lastSceneProps.placements.filter((p: DevicePlacement) => (p.layer || 'controls') === lastSceneProps.activeLayer)
        : lastSceneProps.placements
      expect(visible.map((p: DevicePlacement) => p.id)).toEqual(['p1'])
    }

    const sensorsBtn = screen.getByRole('button', { name: /sensors/i })
    fireEvent.click(sensorsBtn)

    await waitFor(() => {
      expect(lastSceneProps.activeLayer).toBe('sensors')
    })
  })

  it('switches activeLayer when level changes to one without the previous layer', async () => {
    const { rerender } = renderInRouter(
      <IsometricView
        level={level1}
        levels={[level1, level2]}
        onSelectLevel={vi.fn()}
      />
    )

    expect(screen.getByRole('button', { name: /controls/i }).getAttribute('aria-pressed')).toBe('true')

    rerender(
      <MemoryRouter>
        <IsometricView
          level={level2}
          levels={[level1, level2]}
          onSelectLevel={vi.fn()}
        />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /garden_sensors/i }).getAttribute('aria-pressed')).toBe('true')
      expect(lastSceneProps.activeLayer).toBe('garden_sensors')
    })
  })
})