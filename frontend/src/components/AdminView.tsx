import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useApp } from '../useApp'
import { PlanEditor2D } from './PlanEditor2D'
import { resolveParamLevelId } from '../utils/levelSync'
import type { Plan } from '../types'
import { loadHouseOverviewConfig, saveHouseOverviewConfig, type HouseOverviewConfig } from '../utils/houseOverview'

export function AdminView() {
  const { levels, activeLevelId, setActiveLevelId, fetchLevels, activeLevel } = useApp()
  const { levelId } = useParams<{ levelId: string }>()
  const navigate = useNavigate()
  const basePath = '/setup/plans'
  const [alignMode, setAlignMode] = useState(false)
  const [overviewPlans, setOverviewPlans] = useState<Record<string, Plan>>({})
  const [overviewConfig, setOverviewConfig] = useState<HouseOverviewConfig>(() => loadHouseOverviewConfig())

  // URL -> state only. Deps intentionally exclude activeLevelId: when the user
  // picks a level we set state first and navigate second, so for one render
  // the (stale) param differs from state. If this effect re-ran on that
  // transient render it would read the stale param as intent and revert the
  // user's choice, then the state->URL effect would navigate back: infinite
  // /plan + /placements ping-pong with blinking URL. The ref gives the effect
  // the current selection without subscribing to it.
  const activeRef = useRef(activeLevelId)
  activeRef.current = activeLevelId

  useEffect(() => {
    const next = resolveParamLevelId(levelId, levels, activeRef.current)
    if (next) {
      setActiveLevelId(next)
    }
  }, [levelId, levels, setActiveLevelId])

  useEffect(() => {
    if (activeLevelId && activeLevelId !== levelId) {
      navigate(`${basePath}/${activeLevelId}`, { replace: true })
    }
  }, [activeLevelId, levelId, navigate, basePath])

  useEffect(() => {
    if (!alignMode) return
    let cancelled = false
    Promise.all(levels.map(async (level) => {
      const response = await fetch(`/api/levels/${level.id}/plan`).catch(() => null)
      const payload = await response?.json().catch(() => null)
      return [level.id, { level_id: level.id, walls: payload?.data?.walls || [], zones: payload?.data?.zones || [] }] as const
    })).then((entries) => {
      if (!cancelled) setOverviewPlans(Object.fromEntries(entries))
    })
    return () => { cancelled = true }
  }, [alignMode, levels])

  useEffect(() => {
    saveHouseOverviewConfig(overviewConfig)
  }, [overviewConfig])

  return (
    <main className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <PlanEditor2D
        level={activeLevel}
        levels={levels}
        onSelectLevel={setActiveLevelId}
        onRefreshLevels={fetchLevels}
        alignMode={alignMode}
        onToggleAlignMode={() => setAlignMode((value) => !value)}
        overviewPlans={overviewPlans}
        overviewConfig={overviewConfig}
        onOverviewConfigChange={setOverviewConfig}
      />
    </main>
  )
}
