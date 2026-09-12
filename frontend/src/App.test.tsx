import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render } from '@testing-library/react'
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
})
