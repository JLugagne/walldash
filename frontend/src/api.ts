let csrfTokenPromise: Promise<string | null> | null = null

async function fetchCsrfToken(): Promise<string | null> {
  try {
    const res = await fetch('/api/csrf-token')
    if (!res.ok) return null
    const payload = await res.json()
    return payload?.data?.csrf_token ?? null
  } catch {
    return null
  }
}

function getCsrfToken(): Promise<string | null> {
  if (!csrfTokenPromise) {
    csrfTokenPromise = fetchCsrfToken()
  }
  return csrfTokenPromise
}

/**
 * apiFetch wraps fetch and attaches the anti-CSRF token header required by the
 * backend CSRF middleware on state-changing requests (POST/PUT/DELETE/PATCH).
 * GET/HEAD requests bypass the token fetch entirely, matching the backend's
 * safe-method exemption.
 */
export async function apiFetch(input: RequestInfo | URL, init: RequestInit = {}): Promise<Response> {
  const method = (init.method || 'GET').toUpperCase()
  if (method === 'GET' || method === 'HEAD') {
    return fetch(input, init)
  }

  const token = await getCsrfToken()
  const headers = new Headers(init.headers)
  if (token) {
    headers.set('X-CSRF-Token', token)
  }

  const res = await fetch(input, { ...init, headers })
  if (res.status === 403) {
    csrfTokenPromise = null
  }
  return res
}
