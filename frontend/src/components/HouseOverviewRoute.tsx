import { useEffect, useState } from 'react'
import { Canvas } from '@react-three/fiber'
import { useNavigate } from 'react-router-dom'
import { HouseOverviewScene } from './HouseOverviewScene'
import { ViewTopSelectors } from './ViewTopSelectors'
import { ViewModeMenu } from './ViewModeMenu'
import { useApp } from '../useApp'
import type { Device, DevicePlacement, Plan } from '../types'
import { loadHouseOverviewConfig, type HouseOverviewConfig } from '../utils/houseOverview'

export function HouseOverviewRoute({ onClose }: { onClose?: () => void }) {
  const { levels, activeLevel, setActiveLevelId } = useApp()
  const navigate = useNavigate()
  const [plans, setPlans] = useState<Record<string, Plan>>({})
  const [placements, setPlacements] = useState<Record<string, DevicePlacement[]>>({})
  const [devices, setDevices] = useState<Record<string, Device>>({})
  const [config] = useState<HouseOverviewConfig>(() => loadHouseOverviewConfig())

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      const [deviceResponse, ...levelResponses] = await Promise.all([
        fetch('/api/devices').catch(() => null),
        ...levels.flatMap((level) => [
          fetch(`/api/levels/${level.id}/plan`).catch(() => null),
          fetch(`/api/levels/${level.id}/placements`).catch(() => null),
        ]),
      ])
      if (cancelled) return
      const devicePayload = await deviceResponse?.json().catch(() => null)
      const nextDevices: Record<string, Device> = {}
      for (const device of devicePayload?.data || []) nextDevices[device.id] = device
      const nextPlans: Record<string, Plan> = {}
      const nextPlacements: Record<string, DevicePlacement[]> = {}
      const levelPayloads = await Promise.all(levelResponses.map((response) => response?.json().catch(() => null)))
      levels.forEach((level, index) => {
        const planPayload = levelPayloads[index * 2]
        const placementPayload = levelPayloads[index * 2 + 1]
        nextPlans[level.id] = { level_id: level.id, walls: planPayload?.data?.walls || [], zones: planPayload?.data?.zones || [] }
        nextPlacements[level.id] = placementPayload?.data || []
      })
      setDevices(nextDevices)
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
      <div className="absolute bottom-6 right-6 z-20 pointer-events-none">
        <ViewModeMenu direction="up" />
      </div>
      <Canvas frameloop="demand" camera={{ position: [24, 24, 30], fov: 42, near: 0.5, far: 500 }} className="h-full w-full">
        <HouseOverviewScene levels={levels} plans={plans} placements={placements} devices={devices} config={config} />
      </Canvas>
      {levels.length === 0 && <div className="absolute inset-0 flex items-center justify-center text-sm text-slate-400">No floors configured.</div>}
    </section>
  )
}
