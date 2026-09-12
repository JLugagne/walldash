import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'

vi.mock('./hooks/useRealtimeDevices', () => ({
  useRealtimeDeviceControl: () => ({ connected: false }),
  useRealtimeDevices: () => ({ deviceMap: {}, devices: [], pendingDevices: {}, toggleDevice: () => {} }),
}))

import App from './App'

function emptyLevelsResponse() {
  return Promise.resolve({
    ok: true,
    json: async () => ({ status: 'success', data: [] }),
  } as Response)
}

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: '',
    json: async () => body,
  } as Response
}

function renderApp() {
  const router = createMemoryRouter(
    [{ path: '/', element: <App />, children: [{ index: true, element: <div /> }] }],
    { initialEntries: ['/'] },
  )
  return render(<RouterProvider router={router} />)
}

describe('App shell', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(emptyLevelsResponse) as unknown as typeof fetch
  })

  it('fills the dynamic viewport height so iPad browser chrome does not crop the dashboard', () => {
    const { container } = renderApp()
    const shell = container.firstElementChild as HTMLElement
    expect(shell.className).toContain('app-viewport')
    expect(shell.className).not.toContain('h-screen')
  })

  it('returns to the login screen when the silent refresh fails', async () => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const raw =
        typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
      const pathname = new URL(raw, 'http://localhost').pathname
      const method = (init?.method ?? 'GET').toUpperCase()
      if (method === 'GET' && pathname === '/api/auth/me') return jsonResponse(401, { status: 'error' })
      if (method === 'GET' && pathname === '/api/csrf-token')
        return jsonResponse(200, { status: 'success', data: { csrf_token: 'csrf' } })
      if (method === 'POST' && pathname === '/api/auth/refresh')
        return jsonResponse(401, { status: 'error' })
      if (method === 'POST' && pathname === '/api/auth/connect')
        return jsonResponse(200, { status: 'success', data: { status: 'pending' } })
      throw new Error(`Unmocked ${method} ${pathname}`)
    }) as unknown as typeof fetch

    renderApp()

    expect(await screen.findByText('Waiting for approval')).toBeDefined()
  })
})
