export interface Layer {
  name: string
  hide_gauges: boolean
}

export interface Level {
  id: string
  name: string
  order: number
  is_outdoor: boolean
  layers?: Layer[]
  created_at: string
  updated_at: string
}

export interface RestoreSummary {
  levels: number
  plans: number
  placements: number
  overviews: number
  widgets: number
}

export interface Point2D {
  x: number
  y: number
}

export interface WallOpening {
  id: string
  type: 'door' | 'window'
  offset: number
  width: number
  flip_side?: boolean
  flip_hinge?: boolean
  /** Keep the hole in the wall but never render the door leaf/frame (passage). Doors only. */
  hide_door?: boolean
}


export interface WallSegment {
  id: string
  x1: number
  y1: number
  x2: number
  y2: number
  thickness: number
  openings?: WallOpening[]
}


export interface Zone {
  id: string
  name: string
  color: string
  points: Point2D[]
  /** User-adjustable label anchor; omitted for legacy zones. */
  label_position?: Point2D
  temp_sensor?: string
  temp_min?: number
  temp_max?: number
  humidity_sensor?: string
  humidity_min?: number
  humidity_max?: number
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
  last_updated: string
}

export interface DevicePlacement {
  id: string
  level_id: string
  device_id: string
  x: number
  y: number
  icon?: string
  render_domain?: string
  custom_name?: string
  layer?: string
  created_at?: string
  updated_at?: string
}

export interface SavePlacementRequest {
  id?: string
  device_id: string
  x: number
  y: number
  icon?: string
  render_domain?: string
  custom_name?: string
  layer?: string
}

/** What a Widget is bound to, and therefore what a tap on it does. */
export type WidgetType = 'sensor' | 'actuator' | 'automation_list' | 'weather'

/** How a Widget draws its data, independent of what that data is. */
export type DisplayMode = 'number' | 'arc' | 'bar' | 'toggle' | 'list' | 'weather'

/** Which slice of an Open-Meteo forecast a Weather Widget renders. */
export type WeatherMode = 'current' | 'today' | 'tomorrow' | 'ndays'

/** Unit system for a Weather Widget; the empty string defers to the server config. */
export type WeatherUnits = '' | 'metric' | 'imperial'

export interface WidgetConfig {
  entity_ids: string[]
  display: DisplayMode
  labels?: Record<string, string>
  min?: number
  max?: number
  unit?: string
  weather_mode?: WeatherMode
  weather_days?: number
  latitude?: number
  longitude?: number
  location_name?: string
  units?: WeatherUnits
}

export interface Widget {
  id: string
  dashboard_id: string
  type: WidgetType
  title: string
  order: number
  col: number
  row: number
  col_span: number
  row_span: number
  config: WidgetConfig
  created_at: string
  updated_at: string
}

export interface OverviewDashboard {
  id: string
  name: string
  order: number
  cols: number
  rows: number
  /** URL path of the background image rendered behind the Widget Grid, or empty for none. */
  background_image: string
  /** Image opacity in percent (0-100). */
  background_opacity: number
  /** Image blur radius in pixels (0-32). */
  background_blur: number
  /** Opacity in percent of the scrim drawn over the image (0-100). */
  background_dim: number
  created_at: string
  updated_at: string
  widgets: Widget[]
}

export interface Automation {
  id: string
  name: string
  state: string
  current: number
  last_triggered: string | null
}
