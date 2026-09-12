import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LoginScreen } from './LoginScreen'

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

function csrfRoute(): Handler {
  return () => jsonResponse(200, { status: 'success', data: { csrf_token: 'csrf' } })
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('LoginScreen', () => {
  it('runs the pending -> code -> success flow', async () => {
    const onAuthenticated = vi.fn()
    mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(200, { status: 'success', data: { status: 'pending' } }),
      'GET /api/csrf-token': csrfRoute(),
      'POST /api/auth/verify': () =>
        jsonResponse(200, {
          status: 'success',
          data: { device: { id: 'd1', label: 'Tablet', role: 'owner' } },
        }),
    })

    render(<LoginScreen onAuthenticated={onAuthenticated} />)

    expect(await screen.findByText('Waiting for approval')).toBeDefined()

    fireEvent.change(screen.getByLabelText(/6-digit code/i), { target: { value: '004217' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(onAuthenticated).toHaveBeenCalledTimes(1))
  })

  it('shows a friendly message when the code is rejected', async () => {
    mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(200, { status: 'success', data: { status: 'pending' } }),
      'GET /api/csrf-token': csrfRoute(),
      'POST /api/auth/verify': () => jsonResponse(401, { status: 'error', code: 'invalid_code' }),
    })

    render(<LoginScreen onAuthenticated={vi.fn()} />)

    expect(await screen.findByText('Waiting for approval')).toBeDefined()

    fireEvent.change(screen.getByLabelText(/6-digit code/i), { target: { value: '000000' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    expect(await screen.findByText(/did not match/i)).toBeDefined()
  })

  it('shows a rate limit message when connect is throttled', async () => {
    mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(429, { status: 'error' }),
      'GET /api/csrf-token': csrfRoute(),
    })

    render(<LoginScreen onAuthenticated={vi.fn()} />)

    expect(await screen.findByText(/too many attempts/i)).toBeDefined()
  })
})
