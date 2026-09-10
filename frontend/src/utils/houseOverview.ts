import type { Device, DevicePlacement, Level } from '../types'

export interface FloorOverviewConfig {
  x: number
  y: number
  visible: boolean
}

export type HouseOverviewConfig = Record<string, FloorOverviewConfig>

export const DEFAULT_FLOOR_CONFIG: FloorOverviewConfig = { x: 0, y: 0, visible: true }
export const HOUSE_FLOOR_HEIGHT = 3.2

export function floorElevation(index: number): number {
  return Math.max(0, index) * HOUSE_FLOOR_HEIGHT
}

export function loadHouseOverviewConfig(): HouseOverviewConfig {
  try {
    const raw = localStorage.getItem('ha_dash_house_overview')
    if (!raw) return {}
    const parsed = JSON.parse(raw) as HouseOverviewConfig
    if (!parsed || typeof parsed !== 'object') return {}
    return parsed
  } catch {
    return {}
  }
}

export function saveHouseOverviewConfig(config: HouseOverviewConfig) {
  try {
    localStorage.setItem('ha_dash_house_overview', JSON.stringify(config))
  } catch {
    // Local storage can be disabled; the editor remains usable for this session.
  }
}

export function configForLevel(config: HouseOverviewConfig, levelId: string): FloorOverviewConfig {
  const value = config[levelId]
  return {
    x: Number.isFinite(value?.x) ? value.x : DEFAULT_FLOOR_CONFIG.x,
    y: Number.isFinite(value?.y) ? value.y : DEFAULT_FLOOR_CONFIG.y,
    visible: value?.visible !== false,
  }
}

export function isLightPlacement(placement: DevicePlacement, device?: Device): boolean {
  return (placement.render_domain || device?.domain || placement.device_id.split('.')[0]) === 'light'
}

export function filterLightPlacements(
  placements: DevicePlacement[],
  devices: Record<string, Device>
): DevicePlacement[] {
  return placements.filter((placement) => isLightPlacement(placement, devices[placement.device_id]))
}

export function orderedLevels(levels: Level[]): Level[] {
  return [...levels].sort((a, b) => a.order - b.order)
}

export interface HouseCameraState {
  position: [number, number, number]
  target: [number, number, number]
}

const HOUSE_CAMERA_STORAGE_KEY = 'ha_dash_house_camera'

function isVec3(value: unknown): value is [number, number, number] {
  return (
    Array.isArray(value) &&
    value.length === 3 &&
    value.every((n) => typeof n === 'number' && Number.isFinite(n))
  )
}

export function loadHouseCameraState(): HouseCameraState | null {
  try {
    const raw = localStorage.getItem(HOUSE_CAMERA_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<HouseCameraState>
    if (!parsed || !isVec3(parsed.position) || !isVec3(parsed.target)) return null
    return { position: parsed.position, target: parsed.target }
  } catch {
    return null
  }
}

export function saveHouseCameraState(state: HouseCameraState) {
  try {
    localStorage.setItem(HOUSE_CAMERA_STORAGE_KEY, JSON.stringify(state))
  } catch {
    // Local storage can be disabled; the camera simply won't be restored.
  }
}
