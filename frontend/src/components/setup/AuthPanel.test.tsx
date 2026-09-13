import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AuthPanel } from './AuthPanel'

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

function emptyInvites(): Handler {
  return () => jsonResponse(200, { status: 'success', data: [] })
}

const PENDING = [
  {
    pending_id: 'p1',
    device_id: 'dev1',
    label: 'Kitchen tablet',
    approved: false,
    created_at: '2026-09-12T09:45:00Z',
    expires_at: '2026-09-12T10:00:00Z',
  },
]

const DEVICES = [
  {
    id: 'd1',
    label: 'Hall panel',
    role: 'owner',
    status: 'active',
    created_at: '2026-09-01T10:00:00Z',
    last_seen: '2026-09-12T09:00:00Z',
  },
  {
    id: 'd2',
    label: 'Garage tablet',
    role: 'admin',
    status: 'active',
    created_at: '2026-09-02T10:00:00Z',
    last_seen: null,
  },
]

const DEVICES_WITH_REVOKED = [
  ...DEVICES,
  {
    id: 'd3',
    label: 'Old phone',
    role: 'device',
    status: 'revoked',
    created_at: '2026-08-20T10:00:00Z',
    last_seen: '2026-08-21T10:00:00Z',
  },
]

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('AuthPanel', () => {
  it('renders pending approvals without any code', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: PENDING }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
    })

    render(<AuthPanel />)

    expect(await screen.findByText('Kitchen tablet', {}, { timeout: 5000 })).toBeDefined()
    expect(screen.getByRole('button', { name: /approve kitchen tablet/i })).toBeDefined()
    expect(screen.getByRole('button', { name: /deny kitchen tablet/i })).toBeDefined()
  })

  it('approves a pending enrollment through the setup endpoint', async () => {
    const fetchMock = mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: PENDING }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
      'POST /api/setup/auth/pending/p1/approve': () =>
        jsonResponse(200, { status: 'success', data: { id: 'dev1', label: 'Kitchen tablet', role: 'device' } }),
    })

    render(<AuthPanel />)

    fireEvent.click(
      await screen.findByRole('button', { name: /approve kitchen tablet/i }, { timeout: 5000 }),
    )

    await waitFor(() => {
      const calls = fetchMock.mock.calls.map(
        ([input, init]) => `${((init as RequestInit | undefined)?.method ?? 'GET').toUpperCase()} ${String(input)}`,
      )
      expect(calls).toContain('POST /api/setup/auth/pending/p1/approve')
    })
  })

  it('creates an invitation and shows the one-time link', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
      'POST /api/setup/auth/invites': () =>
        jsonResponse(200, {
          status: 'success',
          data: { selector: 'sel1', token: 'tok.abc', role: 'device', expires_at: '2026-09-12T10:15:00Z' },
        }),
    })

    render(<AuthPanel />)

    fireEvent.click(await screen.findByRole('button', { name: /create invitation/i }, { timeout: 5000 }))

    expect(await screen.findByDisplayValue(/invite=tok.abc/)).toBeDefined()
  })

  it('reloads devices after a silent token refresh on 401', async () => {
    let deviceLoads = 0
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => {
        deviceLoads += 1
        if (deviceLoads === 1) return jsonResponse(401, { status: 'error' })
        return jsonResponse(200, { status: 'success', data: DEVICES })
      },
      'POST /api/auth/refresh': () => jsonResponse(204, {}),
    })

    render(<AuthPanel />)

    expect(await screen.findByText('Hall panel', {}, { timeout: 5000 })).toBeDefined()
    expect(deviceLoads).toBe(2)
  })

  it('revokes a device through the setup endpoint', async () => {
    const fetchMock = mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
      'POST /api/setup/auth/devices/d2/revoke': () => jsonResponse(204, {}),
    })

    render(<AuthPanel />)

    fireEvent.click(
      await screen.findByRole('button', { name: /revoke garage tablet/i }, { timeout: 5000 }),
    )

    await waitFor(() => {
      const calls = fetchMock.mock.calls.map(
        ([input, init]) => `${((init as RequestInit | undefined)?.method ?? 'GET').toUpperCase()} ${String(input)}`,
      )
      expect(calls).toContain('POST /api/setup/auth/devices/d2/revoke')
    })
  })

  it('renames a device and reloads the list', async () => {
    let deviceLoads = 0
    const fetchMock = mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => {
        deviceLoads += 1
        return jsonResponse(200, { status: 'success', data: DEVICES })
      },
      'POST /api/setup/auth/devices/d2/label': () =>
        jsonResponse(200, { status: 'success', data: { id: 'd2', label: 'Kitchen panel' } }),
    })

    render(<AuthPanel />)

    fireEvent.click(
      await screen.findByRole('button', { name: /rename garage tablet/i }, { timeout: 5000 }),
    )
    const input = screen.getByRole('textbox', { name: /label for garage tablet/i })
    fireEvent.change(input, { target: { value: 'Kitchen panel' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    await waitFor(() => {
      expect(deviceLoads).toBe(2)
    })

    const labelCall = fetchMock.mock.calls.find(
      ([input, init]) =>
        String(input) === '/api/setup/auth/devices/d2/label' &&
        (init as RequestInit | undefined)?.method === 'POST',
    )
    expect(labelCall).toBeDefined()
    const init = (labelCall?.[1] ?? {}) as RequestInit
    expect(JSON.parse(String(init.body))).toEqual({
      label: 'Kitchen panel',
    })
  })

  it('surfaces the error and keeps the old label when renaming fails', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
      'POST /api/setup/auth/devices/d2/label': () =>
        jsonResponse(400, { status: 'error', message: 'label must be 1..64 characters' }),
    })

    render(<AuthPanel />)

    fireEvent.click(
      await screen.findByRole('button', { name: /rename garage tablet/i }, { timeout: 5000 }),
    )
    const input = screen.getByRole('textbox', { name: /label for garage tablet/i })
    fireEvent.change(input, { target: { value: 'Bad' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(await screen.findByText('label must be 1..64 characters')).toBeDefined()
    expect(screen.getByText('Garage tablet')).toBeDefined()
  })

  it('hides revoked devices by default and reveals them when toggled', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('owner'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES_WITH_REVOKED }),
    })

    render(<AuthPanel />)

    expect(await screen.findByText('Hall panel', {}, { timeout: 5000 })).toBeDefined()
    expect(screen.queryByText('Old phone')).toBeNull()

    fireEvent.click(screen.getByRole('checkbox', { name: /hide revoked/i }))

    expect(await screen.findByText('Old phone')).toBeDefined()
  })

  it('restricts the role selector for non-owner accounts', async () => {
    mockRoutes({
      'GET /api/auth/me': meRoute('admin'),
      'GET /api/setup/auth/pending': () => jsonResponse(200, { status: 'success', data: [] }),
      'GET /api/setup/auth/invites': emptyInvites(),
      'GET /api/setup/auth/devices': () => jsonResponse(200, { status: 'success', data: DEVICES }),
    })

    render(<AuthPanel />)

    const select = await screen.findByLabelText(/role for garage tablet/i, {}, { timeout: 5000 })
    const options = Array.from(select.querySelectorAll('option')).map((option) =>
      option.getAttribute('value'),
    )

    expect(options).toContain('admin')
    expect(options).toContain('device')
    expect(options).not.toContain('owner')

    expect(screen.queryByLabelText(/role for hall panel/i)).toBeNull()
    const renameOwner = screen.getByRole('button', { name: /rename hall panel/i }) as HTMLButtonElement
    const revokeOwner = screen.getByRole('button', { name: /revoke hall panel/i }) as HTMLButtonElement
    expect(renameOwner.disabled).toBe(true)
    expect(revokeOwner.disabled).toBe(true)
  })
})
