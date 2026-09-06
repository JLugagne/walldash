import { useCallback, useEffect, useState } from 'react'
import type { Dispatch, SetStateAction } from 'react'
import type { Automation, OverviewDashboard } from '../../types'

export interface OverviewData {
  overviews: OverviewDashboard[]
  automations: Automation[]
  loading: boolean
  fetchOverviews: () => Promise<void>
  fetchAutomations: () => Promise<void>
}

/**
 * useOverviewData fetches the Overview Dashboards and the Automations list, and keeps
 * automations fresh with a 5 s poll for live status. It never owns the active overview id:
 * callers pass their own setter so `fetchOverviews` can keep the current selection (or fall
 * back to the first dashboard) without this hook holding that piece of state itself.
 */
export function useOverviewData(
  setActiveOverviewId: Dispatch<SetStateAction<string | null>>
): OverviewData {
  const [overviews, setOverviews] = useState<OverviewDashboard[]>([])
  const [automations, setAutomations] = useState<Automation[]>([])
  const [loading, setLoading] = useState(true)

  const fetchOverviews = useCallback(async () => {
    try {
      const res = await fetch('/api/overviews')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setOverviews(payload.data)
          setActiveOverviewId((prev) => {
            if (prev && payload.data.some((o: OverviewDashboard) => o.id === prev)) {
              return prev
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch overviews:', err)
    }
  }, [setActiveOverviewId])

  const fetchAutomations = useCallback(async () => {
    try {
      const res = await fetch('/api/automations')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setAutomations(payload.data)
        }
      }
    } catch (err) {
      console.error('Failed to fetch automations:', err)
    }
  }, [])

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchOverviews(), fetchAutomations()]).finally(() => setLoading(false))

    const interval = setInterval(fetchAutomations, 5000)
    return () => clearInterval(interval)
  }, [fetchOverviews, fetchAutomations])

  return { overviews, automations, loading, fetchOverviews, fetchAutomations }
}
