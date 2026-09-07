import { useEffect, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useApp } from '../useApp'
import { PlanEditor2D } from './PlanEditor2D'
import { ViewModeMenu } from './ViewModeMenu'
import { resolveParamLevelId } from '../utils/levelSync'

export function AdminView() {
  const { levels, activeLevelId, setActiveLevelId, fetchLevels, activeLevel } = useApp()
  const { levelId } = useParams<{ levelId: string }>()
  const navigate = useNavigate()

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
      navigate(`/admin/${activeLevelId}`, { replace: true })
    }
  }, [activeLevelId, levelId, navigate])

  return (
    <main className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <PlanEditor2D
        level={activeLevel}
        levels={levels}
        onSelectLevel={setActiveLevelId}
        onRefreshLevels={fetchLevels}
        viewModeMenu={<ViewModeMenu />}
      />
    </main>
  )
}