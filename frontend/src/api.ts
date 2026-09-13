/**
 * apiFetch performs an authenticated fetch. Session credentials travel in
 * httpOnly cookies, which the browser attaches automatically for same-origin
 * requests.
 */
export function apiFetch(input: RequestInfo | URL, init: RequestInit = {}): Promise<Response> {
  return fetch(input, init)
}

/**
 * readApiError extracts a human-readable message from a failed API response, preferring the
 * envelope's `message` or `error` field and falling back to the HTTP status line. Safe to call
 * on any `Response`, even one whose body is not JSON or already consumed.
 */
export async function readApiError(res: Response): Promise<string> {
  const fallback = res.statusText || `Erreur HTTP ${res.status}`
  try {
    const payload = await res.json()
    const message = payload?.message ?? payload?.error
    return typeof message === 'string' && message ? message : fallback
  } catch {
    return fallback
  }
}
