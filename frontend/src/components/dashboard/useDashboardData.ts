import { useCallback, useEffect, useState } from 'react'
import type { Dispatch, SetStateAction } from 'react'
import type { Automation, Dashboard } from '../../types'

export interface DashboardData {
  dashboards: Dashboard[]
  automations: Automation[]
  loading: boolean
  fetchDashboards: () => Promise<void>
  fetchAutomations: () => Promise<void>
}

/**
 * useDashboardData fetches the Dashboards and the Automations list, and keeps
 * automations fresh with a 5 s poll for live status. It never owns the active dashboard id:
 * callers pass their own setter so `fetchDashboards` can keep the current selection (or fall
 * back to the first dashboard) without this hook holding that piece of state itself.
 */
export function useDashboardData(
  setActiveDashboardId: Dispatch<SetStateAction<string | null>>
): DashboardData {
  const [dashboards, setDashboards] = useState<Dashboard[]>([])
  const [automations, setAutomations] = useState<Automation[]>([])
  const [loading, setLoading] = useState(true)

  const fetchDashboards = useCallback(async () => {
    try {
      const res = await fetch('/api/dashboards')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setDashboards(payload.data)
          setActiveDashboardId((prev) => {
            if (prev && payload.data.some((d: Dashboard) => d.id === prev)) {
              return prev
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch dashboards:', err)
    }
  }, [setActiveDashboardId])

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
    Promise.all([fetchDashboards(), fetchAutomations()]).finally(() => setLoading(false))

    const interval = setInterval(fetchAutomations, 5000)
    return () => clearInterval(interval)
  }, [fetchDashboards, fetchAutomations])

  return { dashboards, automations, loading, fetchDashboards, fetchAutomations }
}
