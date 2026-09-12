import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { useAuth } from './useAuth'

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: '',
    json: async () => body,
  } as Response
}

function emptyResponse(status: number): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: '',
    json: async () => {
      throw new Error('empty body')
    },
  } as unknown as Response
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

function csrfRoute(): Handler {
  return () => jsonResponse(200, { status: 'success', data: { csrf_token: 'csrf' } })
}

const ACCOUNT = { id: 'd1', label: 'Tablet', role: 'owner', status: 'active' }

function Probe() {
  const auth = useAuth()
  return (
    <div>
      <span data-testid="checking">{String(auth.checking)}</span>
      <span data-testid="authenticated">{String(auth.authenticated)}</span>
      <span data-testid="account">{auth.account ? `${auth.account.id}:${auth.account.role}` : ''}</span>
    </div>
  )
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('useAuth', () => {
  it('refreshes once and retries after a 401', async () => {
    let meCalls = 0
    const fetchMock = mockRoutes({
      'GET /api/auth/me': () => {
        meCalls += 1
        return meCalls === 1
          ? jsonResponse(401, { status: 'error' })
          : jsonResponse(200, { status: 'success', data: ACCOUNT })
      },
      'GET /api/csrf-token': csrfRoute(),
      'POST /api/auth/refresh': () => emptyResponse(204),
    })

    render(<Probe />)

    await waitFor(() => expect(screen.getByTestId('authenticated').textContent).toBe('true'))
    expect(screen.getByTestId('account').textContent).toBe('d1:owner')
    expect(meCalls).toBe(2)

    const calls = fetchMock.mock.calls.map(([input]) => String(input))
    expect(calls).toContain('/api/auth/refresh')
  })

  it('stays unauthenticated when the silent refresh fails', async () => {
    mockRoutes({
      'GET /api/auth/me': () => jsonResponse(401, { status: 'error' }),
      'GET /api/csrf-token': csrfRoute(),
      'POST /api/auth/refresh': () => jsonResponse(401, { status: 'error' }),
    })

    render(<Probe />)

    await waitFor(() => expect(screen.getByTestId('checking').textContent).toBe('false'))
    expect(screen.getByTestId('authenticated').textContent).toBe('false')
  })
})
