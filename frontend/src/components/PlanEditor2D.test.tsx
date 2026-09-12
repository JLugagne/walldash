import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import { createMemoryRouter, RouterProvider, useNavigate } from 'react-router-dom'
import { PlanEditor2D } from './PlanEditor2D'
import type { Device, Level, WallSegment } from '../types'

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

const LEVEL: Level = {
  id: 'lvl-1',
  name: 'Ground',
  order: 0,
  is_outdoor: false,
  layers: [{ name: 'controls', hide_gauges: false }],
  created_at: '',
  updated_at: '',
}

const WALL: WallSegment = { id: 'w1', x1: 0, y1: 0, x2: 100, y2: 0, thickness: 8, openings: [] }

const DEVICE: Device = {
  id: 'light.1',
  name: 'Lamp',
  domain: 'light',
  state: 'on',
  attributes: {},
  last_updated: '',
}

function jsonResponse(payload: unknown): Promise<Response> {
  return Promise.resolve({ ok: true, json: async () => payload } as Response)
}

function Harness() {
  const navigate = useNavigate()
  return (
    <div>
      <button type="button" onClick={() => navigate('/dashboards')}>
        leave-editor
      </button>
      <PlanEditor2D
        level={LEVEL}
        levels={[LEVEL]}
        onSelectLevel={() => {}}
        onRefreshLevels={async () => {}}
      />
    </div>
  )
}

function renderEditor() {
  const router = createMemoryRouter(
    [
      { path: '/setup/plans/:levelId', element: <Harness /> },
      { path: '/dashboards', element: <div>Dashboards page</div> },
    ],
    { initialEntries: ['/setup/plans/lvl-1'] },
  )
  return render(<RouterProvider router={router} />)
}

async function loadPlanWithWall() {
  renderEditor()
  await waitFor(() => expect(document.querySelector('line.cursor-move')).toBeTruthy())
}

async function makeDirty() {
  fireEvent.click(await screen.findByTitle('More actions'))
  fireEvent.click(await screen.findByText('Clear walls and zones'))
  await waitFor(() => expect(document.querySelector('line.cursor-move')).toBeNull())
}

describe('PlanEditor2D navigation guard', () => {
  let confirmSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    vi.stubGlobal('ResizeObserver', ResizeObserverMock)
    confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    globalThis.fetch = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.endsWith('/plan')) return jsonResponse({ status: 'success', data: { walls: [WALL], zones: [] } })
      return jsonResponse({ status: 'success', data: [] })
    }) as unknown as typeof fetch
  })

  afterEach(() => {
    confirmSpy.mockRestore()
    vi.unstubAllGlobals()
  })

  it('asks before leaving the editor with unsaved changes', async () => {
    await loadPlanWithWall()
    await makeDirty()

    fireEvent.click(screen.getByText('leave-editor'))

    await waitFor(() =>
      expect(confirmSpy).toHaveBeenCalledWith(
        'Unsaved modifications. Leaving the editor will discard them. Continue?',
      ),
    )
    expect(await screen.findByText('Dashboards page')).toBeDefined()
  })

  it('leaves without asking when the plan is saved', async () => {
    await loadPlanWithWall()

    fireEvent.click(screen.getByText('leave-editor'))

    expect(await screen.findByText('Dashboards page')).toBeDefined()
    expect(confirmSpy).not.toHaveBeenCalledWith(
      'Unsaved modifications. Leaving the editor will discard them. Continue?',
    )
  })
})

describe('PlanEditor2D touch placement', () => {
  beforeEach(() => {
    vi.stubGlobal('ResizeObserver', ResizeObserverMock)
    globalThis.fetch = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)
      const method = (init?.method ?? 'GET').toUpperCase()
      if (url.endsWith('/plan')) return jsonResponse({ status: 'success', data: { walls: [WALL], zones: [] } })
      if (url.endsWith('/placements') && method === 'POST') {
        return jsonResponse({
          status: 'success',
          data: { id: 'pl-1', level_id: 'lvl-1', device_id: 'light.1', x: 500, y: 500, layer: 'controls' },
        })
      }
      if (url.endsWith('/api/devices')) return jsonResponse({ status: 'success', data: [DEVICE] })
      return jsonResponse({ status: 'success', data: [] })
    }) as unknown as typeof fetch
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('places a device dropped from the palette with a touch pointer', async () => {
    await loadPlanWithWall()

    fireEvent.click(screen.getByRole('button', { name: /devices/i }))
    const dragIcon = await waitFor(() => {
      const el = document.querySelector('[data-drag-icon]') as HTMLElement | null
      if (!el) throw new Error('palette icon not rendered yet')
      return el
    })

    // jsdom has no layout and no SVG CTM: give the plan an identity geometry so the drop lands.
    const svg = document.querySelector('svg.touch-none') as SVGSVGElement
    svg.getBoundingClientRect = () =>
      ({ left: 0, top: 0, right: 1000, bottom: 1000, width: 1000, height: 1000, x: 0, y: 0, toJSON: () => ({}) }) as DOMRect
    svg.getScreenCTM = () =>
      ({ a: 1, b: 0, c: 0, d: 1, e: 0, f: 0, inverse: () => ({ a: 1, b: 0, c: 0, d: 1, e: 0, f: 0 }) }) as unknown as DOMMatrix

    act(() => {
      const down = new Event('pointerdown', { bubbles: true, cancelable: true })
      Object.assign(down, { pointerType: 'touch', pointerId: 1, clientX: 10, clientY: 10, button: 0 })
      dragIcon.dispatchEvent(down)
    })
    act(() => {
      const up = new Event('pointerup', { bubbles: true, cancelable: true })
      Object.assign(up, { pointerType: 'touch', pointerId: 1, clientX: 500, clientY: 500 })
      window.dispatchEvent(up)
    })

    await waitFor(() =>
      expect(globalThis.fetch).toHaveBeenCalledWith(
        '/api/levels/lvl-1/placements',
        expect.objectContaining({ method: 'POST' }),
      ),
    )
  })
})
