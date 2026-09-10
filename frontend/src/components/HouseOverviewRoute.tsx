import { useEffect, useState } from 'react'
import { Canvas } from '@react-three/fiber'
import { useNavigate } from 'react-router-dom'
import { HouseOverviewScene } from './HouseOverviewScene'
import { ViewTopSelectors } from './ViewTopSelectors'
import { useApp } from '../useApp'
import { useRealtimeDeviceControl } from '../hooks/useRealtimeDevices'
import type { DevicePlacement, Plan } from '../types'
import { loadHouseOverviewConfig, type HouseOverviewConfig } from '../utils/houseOverview'

export function HouseOverviewRoute({ onClose }: { onClose?: () => void }) {
  const { levels, activeLevel, setActiveLevelId } = useApp()
  const navigate = useNavigate()
  const [plans, setPlans] = useState<Record<string, Plan>>({})
  const [placements, setPlacements] = useState<Record<string, DevicePlacement[]>>({})
  const [config] = useState<HouseOverviewConfig>(() => loadHouseOverviewConfig())

  // Live device state + lamp toggling: on/off colors follow Home Assistant and
  // lamps can be tapped in the overview. Placements and plans only change in the editor.
  const { deviceMap: devices, pendingDevices, toggleDevice } = useRealtimeDeviceControl()

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      const levelResponses = await Promise.all(
        levels.flatMap((level) => [
          fetch(`/api/levels/${level.id}/plan`).catch(() => null),
          fetch(`/api/levels/${level.id}/placements`).catch(() => null),
        ])
      )
      if (cancelled) return
      const nextPlans: Record<string, Plan> = {}
      const nextPlacements: Record<string, DevicePlacement[]> = {}
      const levelPayloads = await Promise.all(levelResponses.map((response) => response?.json().catch(() => null)))
      levels.forEach((level, index) => {
        const planPayload = levelPayloads[index * 2]
        const placementPayload = levelPayloads[index * 2 + 1]
        nextPlans[level.id] = { level_id: level.id, walls: planPayload?.data?.walls || [], zones: planPayload?.data?.zones || [] }
        nextPlacements[level.id] = placementPayload?.data || []
      })
      setPlans(nextPlans)
      setPlacements(nextPlacements)
    }
    void load()
    return () => { cancelled = true }
  }, [levels])

  const openFloor = (levelId: string) => {
    setActiveLevelId(levelId)
    if (onClose) onClose()
    else navigate(`/floor/${levelId}`)
  }

  const closeOverview = () => {
    if (onClose) {
      onClose()
      return
    }
    if (activeLevel) navigate(`/floor/${activeLevel.id}`)
  }

  return (
    <section className="relative flex min-h-[420px] flex-1 overflow-hidden bg-[#0b0f19]">
      <ViewTopSelectors
        levels={levels}
        activeLevelId={activeLevel?.id ?? null}
        onSelectLevel={openFloor}
        overviewActive
        onToggleOverview={closeOverview}
      />
      <Canvas frameloop="demand" camera={{ position: [24, 24, 30], fov: 42, near: 0.5, far: 500 }} className="h-full w-full">
        <HouseOverviewScene
          levels={levels}
          plans={plans}
          placements={placements}
          devices={devices}
          pendingDevices={pendingDevices}
          onToggleDevice={toggleDevice}
          config={config}
        />
      </Canvas>
      {levels.length === 0 && <div className="absolute inset-0 flex items-center justify-center text-sm text-slate-400">No floors configured.</div>}
    </section>
  )
}
