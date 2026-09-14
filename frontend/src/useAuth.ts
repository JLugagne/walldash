import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from './api'

export interface Account {
  id: string
  label: string
  role: string
  status?: string
}

export interface AuthState {
  authenticated: boolean
  account: Account | null
  checking: boolean
  refresh: () => Promise<void>
  logout: () => Promise<void>
}

let refreshPromise: Promise<boolean> | null = null

function parseAccount(payload: unknown): Account | null {
  if (!payload || typeof payload !== 'object') return null
  const data = (payload as { data?: unknown }).data
  if (!data || typeof data !== 'object') return null
  const record = data as Record<string, unknown>
  if (typeof record.id !== 'string') return null
  return {
    id: record.id,
    label: typeof record.label === 'string' ? record.label : '',
    role: typeof record.role === 'string' ? record.role : '',
    status: typeof record.status === 'string' ? record.status : undefined,
  }
}

/**
 * refreshSession performs the silent POST /api/auth/refresh, collapsing concurrent
 * callers onto a single in-flight request so a 401 storm triggers one rotation only.
 */
export function refreshSession(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = apiFetch('/api/auth/refresh', { method: 'POST' })
      .then((res) => res.ok)
      .catch(() => false)
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

/**
 * authFetch wraps apiFetch and retries the request exactly once after a successful
 * silent refresh when the access cookie is stale: 401 means it expired, and 403 means
 * the route is gated on a scope the token does not carry yet. The 403 case is how a
 * freshly promoted device picks up its new scopes — a role change never invalidates an
 * issued token (the server re-checks the account itself for every privileged operation),
 * so the smaller scope set is rotated away on the first gated request that needs it. A
 * genuinely forbidden operation just answers 403 again. Both statuses are written by the
 * authentication middleware before any handler runs, so retrying cannot replay a side
 * effect. It returns the final response, so callers still decide whether to treat a 401
 * as unauthenticated.
 */
export async function authFetch(
  input: RequestInfo | URL,
  init: RequestInit = {},
): Promise<Response> {
  const res = await apiFetch(input, init)
  if (res.status !== 401 && res.status !== 403) return res
  if (!(await refreshSession())) return res
  return apiFetch(input, init)
}

/**
 * useAuth bootstraps the session from GET /api/auth/me, keeps the in-memory account
 * in sync, and exposes refresh/logout. Tokens stay in HttpOnly cookies; this hook
 * never reads or stores them.
 */
export function useAuth(): AuthState {
  const [account, setAccount] = useState<Account | null>(null)
  const [checking, setChecking] = useState(true)

  const bootstrap = useCallback(async () => {
    try {
      const res = await authFetch('/api/auth/me')
      if (res.ok) {
        const payload = await res.json()
        setAccount(parseAccount(payload))
      } else {
        setAccount(null)
      }
    } catch {
      setAccount(null)
    } finally {
      setChecking(false)
    }
  }, [])

  useEffect(() => {
    void bootstrap()
  }, [bootstrap])

  const refresh = useCallback(async () => {
    await bootstrap()
  }, [bootstrap])

  const logout = useCallback(async () => {
    try {
      await apiFetch('/api/auth/logout', { method: 'POST' })
    } catch {
      setAccount(null)
      return
    }
    setAccount(null)
  }, [])

  return { authenticated: account !== null, account, checking, refresh, logout }
}
