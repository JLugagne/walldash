import { useEffect, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useApp } from '../useApp'
import { IsometricView } from './IsometricView'
import { resolveParamLevelId } from '../utils/levelSync'

export function ThreeDView() {
  const { levels, activeLevelId, setActiveLevelId, activeLevel } = useApp()
  const { levelId } = useParams<{ levelId: string }>()
  const navigate = useNavigate()

  // Same URL->state contract as AdminView: never re-run on activeLevelId,
  // otherwise a user selection (state set before navigate lands) is reverted
  // by the stale param and the two effects ping-pong forever.
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
      navigate(`/floor/${activeLevelId}`, { replace: true })
    }
  }, [activeLevelId, levelId, navigate])

  return (
    <IsometricView
      level={activeLevel}
      levels={levels}
      onSelectLevel={setActiveLevelId}
    />
  )
}