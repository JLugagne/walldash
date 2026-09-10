import { useCallback, useEffect, useMemo, useRef, type ComponentRef } from 'react'
import { useThree } from '@react-three/fiber'
import { OrbitControls } from '@react-three/drei'
import * as THREE from 'three'
import type { Device, DevicePlacement, Level, Plan } from '../types'
import type { HouseOverviewConfig } from '../utils/houseOverview'
import {
  configForLevel,
  filterLightPlacements,
  floorElevation,
  loadHouseCameraState,
  orderedLevels,
  saveHouseCameraState,
} from '../utils/houseOverview'
import { IsometricScene, WALL_HEIGHT } from './IsometricScene'
import { SCALE, toWorldX, toWorldZ } from './wallGeometry'

interface HouseOverviewSceneProps {
  levels: Level[]
  plans: Record<string, Plan>
  placements: Record<string, DevicePlacement[]>
  devices: Record<string, Device>
  pendingDevices?: Record<string, boolean>
  onToggleDevice?: (entityId: string) => void
  config: HouseOverviewConfig
}

function OverviewCamera({ levels, plans, config }: Pick<HouseOverviewSceneProps, 'levels' | 'plans' | 'config'>) {
  const { camera, size, invalidate } = useThree()
  const controlsRef = useRef<ComponentRef<typeof OrbitControls> | null>(null)
  const restoredRef = useRef(false)

  const bounds = useMemo(() => {
    let minX = -10
    let maxX = 10
    let minZ = -7
    let maxZ = 7
    let maxY = WALL_HEIGHT
    const visibleLevels = orderedLevels(levels).filter((level) => configForLevel(config, level.id).visible)
    visibleLevels.forEach((level, index) => {
      const floor = configForLevel(config, level.id)
      const plan = plans[level.id]
      for (const wall of plan?.walls || []) {
        minX = Math.min(minX, toWorldX(wall.x1) + floor.x * SCALE, toWorldX(wall.x2) + floor.x * SCALE)
        maxX = Math.max(maxX, toWorldX(wall.x1) + floor.x * SCALE, toWorldX(wall.x2) + floor.x * SCALE)
        minZ = Math.min(minZ, toWorldZ(wall.y1) + floor.y * SCALE, toWorldZ(wall.y2) + floor.y * SCALE)
        maxZ = Math.max(maxZ, toWorldZ(wall.y1) + floor.y * SCALE, toWorldZ(wall.y2) + floor.y * SCALE)
      }
      maxY = Math.max(maxY, floorElevation(index) + WALL_HEIGHT)
    })
    return { minX, maxX, minZ, maxZ, minY: 0, maxY }
  }, [config, levels, plans])

  const fitTarget = useMemo(() => ({
    x: (bounds.minX + bounds.maxX) / 2,
    y: (bounds.minY + bounds.maxY) / 2,
    z: (bounds.minZ + bounds.maxZ) / 2,
  }), [bounds])

  // Plans arrive asynchronously; until they do we cannot auto-fit meaningfully.
  const dataReady = Object.keys(plans).length > 0

  // Keep the projection aspect in sync with the viewport without moving the camera.
  useEffect(() => {
    if (!(camera instanceof THREE.PerspectiveCamera)) return
    camera.aspect = size.width / Math.max(size.height, 1)
    camera.updateProjectionMatrix()
    invalidate()
  }, [camera, invalidate, size.height, size.width])

  // Restore the last user camera; fall back to an auto-fit the first time.
  // Runs once per mount, so re-renders (e.g. live device updates) never move the camera.
  useEffect(() => {
    if (!(camera instanceof THREE.PerspectiveCamera)) return
    if (restoredRef.current) return

    const controls = controlsRef.current
    const saved = loadHouseCameraState()

    if (saved) {
      camera.position.set(saved.position[0], saved.position[1], saved.position[2])
      if (controls) {
        controls.target.set(saved.target[0], saved.target[1], saved.target[2])
        controls.update()
      } else {
        camera.lookAt(saved.target[0], saved.target[1], saved.target[2])
      }
      restoredRef.current = true
      invalidate()
      return
    }

    if (!dataReady) return

    const span = Math.max(bounds.maxX - bounds.minX, bounds.maxZ - bounds.minZ, bounds.maxY - bounds.minY)
    const verticalFov = (camera.fov * Math.PI) / 180
    const distance = Math.max(12, (span / (2 * Math.tan(verticalFov / 2))) * 1.35)
    const direction = new THREE.Vector3(1, 0.8, 1).normalize()
    camera.position.set(fitTarget.x, fitTarget.y, fitTarget.z).addScaledVector(direction, distance)
    if (controls) {
      controls.target.set(fitTarget.x, fitTarget.y, fitTarget.z)
      controls.update()
    } else {
      camera.lookAt(fitTarget.x, fitTarget.y, fitTarget.z)
    }
    restoredRef.current = true
    invalidate()
  }, [bounds, camera, dataReady, fitTarget, invalidate])

  // Remember the camera the user settles on, so reopening the overview restores it.
  const persistCamera = useCallback(() => {
    if (!(camera instanceof THREE.PerspectiveCamera)) return
    const controls = controlsRef.current
    saveHouseCameraState({
      position: [camera.position.x, camera.position.y, camera.position.z],
      target: [controls?.target.x ?? 0, controls?.target.y ?? 0, controls?.target.z ?? 0],
    })
  }, [camera])

  return (
    <OrbitControls
      ref={controlsRef}
      enableDamping
      dampingFactor={0.08}
      enablePan
      enableZoom
      minDistance={8}
      maxDistance={180}
      minPolarAngle={0.12}
      maxPolarAngle={Math.PI - 0.12}
      onEnd={persistCamera}
    />
  )
}

export function HouseOverviewScene({ levels, plans, placements, devices, pendingDevices = {}, onToggleDevice, config }: HouseOverviewSceneProps) {
  const visibleLevels = orderedLevels(levels).filter((level) => configForLevel(config, level.id).visible)

  return (
    <>
      <OverviewCamera levels={levels} plans={plans} config={config} />
      <hemisphereLight args={['#dbeafe', '#172033', 1.5]} />
      <ambientLight intensity={1.4} />
      <directionalLight position={[25, 45, 25]} intensity={3.2} />
      <directionalLight position={[-25, 20, -15]} intensity={1.2} color="#93c5fd" />
      <color attach="background" args={['#0b0f19']} />
      {visibleLevels.map((level, index) => {
        const floor = configForLevel(config, level.id)
        const plan = plans[level.id] || { level_id: level.id, walls: [], zones: [] }
        const floorPlacements = filterLightPlacements(placements[level.id] || [], devices)
        return (
          <group key={level.id} position={[floor.x * SCALE, floorElevation(index), floor.y * SCALE]}>
            <IsometricScene
              plan={plan}
              placements={floorPlacements}
              deviceMap={devices}
              pendingDevices={pendingDevices}
              onToggleDevice={onToggleDevice}
              activeLayer={undefined}
              layers={level.layers}
              onlyLights
              showCamera={false}
              showGround={false}
              showLighting={false}
              lightMountHeight={WALL_HEIGHT / 2}
            />
          </group>
        )
      })}
    </>
  )
}
