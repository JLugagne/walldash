import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'

vi.mock('../hooks/useRealtimeDevices', () => ({
  useRealtimeDeviceControl: () => ({ connected: false }),
  useRealtimeDevices: () => ({ deviceMap: {}, devices: [], pendingDevices: {}, toggleDevice: () => {} }),
}))

import { AppTopBar } from './AppTopBar'

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: '',
    json: async () => body,
  } as Response
}

type Handler = () => Response | Promise<Response>

function mockRoutes(routes: Record<string, Handler>): ReturnType<typeof vi.fn> {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const raw = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
    const pathname = new URL(raw, 'http://localhost').pathname
    const method = (init?.method ?? 'GET').toUpperCase()
    const handler = routes[`${method} ${pathname}`]
    if (!handler) throw new Error(`Unmocked ${method} ${pathname}`)
    return handler()
  })
  globalThis.fetch = fetchMock as unknown as typeof fetch
  return fetchMock
}

function meRoute(role: string): Handler {
  return () => jsonResponse(200, { status: 'success', data: { id: 'me', label: 'Owner', role } })
}

const PENDING = [
  { device_id: 'p1', label: 'Kitchen tablet', code: '004217', expires_at: '2026-09-12T10:00:00Z' },
  { device_id: 'p2', label: 'Hall panel', code: '771122', expires_at: '2026-09-12T10:05:00Z' },
]

function renderTopBar() {
  const router = createMemoryRouter(
    [
      { path: '/', element: <AppTopBar activeLevelId={null} /> },
      { path: '/setup/access', element: <div>Access page</div> },
    ],
    { initialEntries: ['/'] },
  )
  return render(<RouterProvider router={router} />)
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('AppTopBar access entrypoint', () => {
  it('shows the Access button for owners and navigates to the access route', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
    })

    renderTopBar()

    fireEvent.click(await screen.findByRole('button', { name: /access/i }))

    expect(await screen.findByText('Access page')).toBeDefined()
  })

  it('shows the Access button for admins', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('admin'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
    })

    renderTopBar()

    expect(await screen.findByRole('button', { name: /access/i })).toBeDefined()
  })

  it('hides the Access button for device accounts', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('device'),
    })

    renderTopBar()

    await screen.findByRole('button', { name: /setup/i })
    expect(screen.queryByRole('button', { name: /access/i })).toBeNull()
  })

  it('badges the pending enrollment count with a descriptive title', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: PENDING }),
    })

    renderTopBar()

    const access = await screen.findByRole('button', { name: /access/i })

    await waitFor(() => expect(access.textContent).toContain('2'))
    expect(access.getAttribute('title')).toBe('2 enrollment requests pending')
  })

  it('offers no badge when there is no pending enrollment', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
    })

    renderTopBar()

    const access = await screen.findByRole('button', { name: /access/i })

    await waitFor(() => expect(access.getAttribute('title')).toBe('Open access management'))
    expect(access.textContent).not.toContain('0')
  })
})
