import { useEffect, useState } from 'react'

/**
 * Phone breakpoint for the mobile Dashboard flow. 639px is Tailwind's `sm` minus one, so the
 * `max-sm:` utilities and this hook always agree: phones (portrait) switch to the flow, small
 * tablets and iPads keep the fixed Widget Grid.
 */
export const MOBILE_QUERY = '(max-width: 639px)'

function matchesMobile(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  return window.matchMedia(MOBILE_QUERY).matches
}

/**
 * useIsMobile reports whether the viewport is at the phone breakpoint and re-renders when it
 * crosses it (rotation, window resize). It resolves synchronously on first render so a phone never
 * flashes the desktop grid, and it degrades to `false` where `matchMedia` is unavailable (jsdom).
 */
export function useIsMobile(): boolean {
  const [isMobile, setIsMobile] = useState(matchesMobile)

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return
    const mql = window.matchMedia(MOBILE_QUERY)
    const onChange = () => setIsMobile(mql.matches)
    onChange()
    mql.addEventListener('change', onChange)
    return () => mql.removeEventListener('change', onChange)
  }, [])

  return isMobile
}
