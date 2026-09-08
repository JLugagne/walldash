import { useEffect, useState, useCallback, useMemo } from 'react'
import { Outlet } from 'react-router-dom'
import type { Level } from './types'
import type { AppContext } from './useApp'
import { OnboardingWizard } from './components/OnboardingWizard'

function App() {
  const [levels, setLevels] = useState<Level[]>([])
  const [activeLevelId, setActiveLevelId] = useState<string | null>(null)
  const [loaded, setLoaded] = useState(false)
  const [wizardDismissed, setWizardDismissed] = useState(false)

  const fetchLevels = useCallback(async () => {
    try {
      const res = await fetch('/api/levels')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setLevels(payload.data)
          setActiveLevelId((prev) => {
            if (prev && payload.data.some((l: Level) => l.id === prev)) {
              return prev
            }
            const hash = window.location.hash
            const match = hash.match(/\/(?:floor|admin)\/([^/?#]+)/)
            if (match) {
              const urlLevelId = decodeURIComponent(match[1])
              if (payload.data.some((l: Level) => l.id === urlLevelId)) {
                return urlLevelId
              }
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch levels:', err)
    } finally {
      setLoaded(true)
    }
  }, [])

  useEffect(() => {
    fetchLevels()
  }, [fetchLevels])

  useEffect(() => {
    if (levels.length > 0) setWizardDismissed(false)
  }, [levels.length])

  const activeLevel = useMemo(
    () => levels.find((l) => l.id === activeLevelId) || null,
    [levels, activeLevelId]
  )

  const context: AppContext = {
    levels,
    activeLevelId,
    setActiveLevelId,
    fetchLevels,
    activeLevel,
  }

  const showWizard = loaded && levels.length === 0 && !wizardDismissed

  return (
    <div className="h-screen w-screen bg-slate-950 text-slate-100 flex flex-col font-sans overflow-hidden">
      {showWizard && (
        <OnboardingWizard
          onDone={() => {
            setWizardDismissed(true)
            void fetchLevels()
          }}
          onClose={() => setWizardDismissed(true)}
        />
      )}
      <Outlet context={context} />
    </div>
  )
}

export default App