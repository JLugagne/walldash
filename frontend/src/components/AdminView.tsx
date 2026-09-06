import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useApp } from '../useApp'
import { PlanEditor2D } from './PlanEditor2D'
import { ViewModeMenu } from './ViewModeMenu'

export function AdminView() {
  const { levels, activeLevelId, setActiveLevelId, fetchLevels, activeLevel } = useApp()
  const { levelId } = useParams<{ levelId: string }>()
  const navigate = useNavigate()

  useEffect(() => {
    if (levelId && levels.length > 0 && activeLevelId !== levelId) {
      setActiveLevelId(levelId)
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