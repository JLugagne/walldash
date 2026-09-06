import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useApp } from '../useApp'
import { IsometricView } from './IsometricView'

export function ThreeDView() {
  const { levels, activeLevelId, setActiveLevelId, activeLevel } = useApp()
  const { levelId } = useParams<{ levelId: string }>()
  const navigate = useNavigate()

  useEffect(() => {
    if (levelId && levels.length > 0 && activeLevelId !== levelId) {
      setActiveLevelId(levelId)
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