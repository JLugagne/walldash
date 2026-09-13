import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, act } from '@testing-library/react'
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

function setUrl(search: string) {
  window.history.replaceState({}, '', `/${search}`)
}

beforeEach(() => {
  vi.restoreAllMocks()
  vi.useRealTimers()
  setUrl('')
})

describe('LoginScreen', () => {
  it('signs the first device in automatically as owner', async () => {
    const onAuthenticated = vi.fn()
    mockRoutes({
      'POST /api/auth/connect': () =>
        jsonResponse(200, {
          status: 'success',
          data: { status: 'authenticated', device: { id: 'd1', label: 'Tablet', role: 'owner' } },
        }),
    })

    render(<LoginScreen onAuthenticated={onAuthenticated} />)

    await waitFor(() => expect(onAuthenticated).toHaveBeenCalledTimes(1))
  })

  it('shows the waiting state when an owner must approve the device', async () => {
    mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(200, { status: 'success', data: { status: 'pending' } }),
      'POST /api/auth/redeem': () => jsonResponse(202, { status: 'success', data: { status: 'pending' } }),
    })

    render(<LoginScreen onAuthenticated={vi.fn()} />)

    expect(await screen.findByText('Waiting for approval')).toBeDefined()
  })

  it('resumes an existing pending enrollment instead of creating a new one, and keeps polling', async () => {
    vi.useFakeTimers()
    const onAuthenticated = vi.fn()
    let redeemCalls = 0
    const fetchMock = mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(200, { status: 'success', data: { status: 'pending' } }),
      'POST /api/auth/redeem': () => {
        redeemCalls += 1
        if (redeemCalls <= 2) {
          return jsonResponse(202, { status: 'success', data: { status: 'pending' } })
        }
        return jsonResponse(200, {
          status: 'success',
          data: { status: 'authenticated', device: { id: 'd1', label: 'Tablet', role: 'owner' } },
        })
      },
    })

    render(<LoginScreen onAuthenticated={onAuthenticated} />)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    expect(screen.getByText('Waiting for approval')).toBeDefined()
    // The existing pending cookie is reused: no new approval request is created.
    const createdNewEnrollment = fetchMock.mock.calls.some(([input]) =>
      String(input).includes('/api/auth/connect'),
    )
    expect(createdNewEnrollment).toBe(false)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })
    expect(onAuthenticated).not.toHaveBeenCalled()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })
    expect(onAuthenticated).toHaveBeenCalledTimes(1)

    vi.useRealTimers()
  })

  it('redeems an invitation carried by the URL', async () => {
    setUrl('?invite=abc.def')
    const onAuthenticated = vi.fn()
    mockRoutes({
      'POST /api/auth/invite/redeem': () =>
        jsonResponse(200, {
          status: 'success',
          data: { status: 'authenticated', device: { id: 'd2', label: 'Tablet', role: 'device' } },
        }),
    })

    render(<LoginScreen onAuthenticated={onAuthenticated} />)

    await waitFor(() => expect(onAuthenticated).toHaveBeenCalledTimes(1))
  })

  it('shows a friendly message when the invitation is rejected', async () => {
    setUrl('?invite=bad-token')
    mockRoutes({
      'POST /api/auth/invite/redeem': () => jsonResponse(401, { status: 'error', code: 'invalid_invite' }),
    })

    render(<LoginScreen onAuthenticated={vi.fn()} />)

    expect(await screen.findByText(/invalid or has already been used/i)).toBeDefined()
  })

  it('shows a rate limit message when connect is throttled', async () => {
    mockRoutes({
      'POST /api/auth/connect': () => jsonResponse(429, { status: 'error' }),
    })

    render(<LoginScreen onAuthenticated={vi.fn()} />)

    expect(await screen.findByText(/too many attempts/i)).toBeDefined()
  })
})
