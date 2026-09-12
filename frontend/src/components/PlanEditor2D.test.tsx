import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { createMemoryRouter, RouterProvider, useNavigate } from 'react-router-dom'
import { PlanEditor2D } from './PlanEditor2D'
import type { Level, WallSegment } from '../types'

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
