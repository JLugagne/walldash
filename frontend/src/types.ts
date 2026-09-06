export interface Level {
  id: string
  name: string
  order: number
  is_outdoor: boolean
  created_at: string
  updated_at: string
}

export interface Point2D {
  x: number
  y: number
}

export interface WallSegment {
  id: string
  x1: number
  y1: number
  x2: number
  y2: number
  thickness: number
}

export interface Zone {
  id: string
  name: string
  color: string
  points: Point2D[]
}

export interface Plan {
  level_id: string
  walls: WallSegment[]
  zones: Zone[]
}

export interface Device {
  id: string
  name: string
  domain: string
  state: string
  attributes: Record<string, any>
}

export interface DevicePlacement {
  id: string
  level_id: string
  device_id: string
  x: number
  y: number
  icon?: string
  custom_name?: string
  created_at?: string
  updated_at?: string
}

export interface SavePlacementRequest {
  id?: string
  device_id: string
  x: number
  y: number
  icon?: string
  custom_name?: string
}
