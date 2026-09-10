import React, { useEffect, useState } from 'react'
import { CloudSun, Droplets, MapPin, Wind } from 'lucide-react'
import type { WeatherMode, WeatherUnits, Widget } from '../../../types'
import { apiFetch } from '../../../api'
import { WidgetFrame } from './WidgetFrame'

const REFRESH_MS = 15 * 60 * 1000
const DEFAULT_DAYS = 5
const MIN_DAYS = 1
const MAX_DAYS = 14
const OPEN_METEO_URL = 'https://api.open-meteo.com/v1/forecast'

export type TemperatureUnit = 'celsius' | 'fahrenheit'
export type WindUnit = 'kmh' | 'mph'

interface WeatherConfig {
  configured: boolean
  latitude: number
  longitude: number
  name: string
  temperature_unit: TemperatureUnit
  wind_unit: WindUnit
}

interface CurrentWeather {
  temperature_2m: number
  relative_humidity_2m: number
  weather_code: number
  wind_speed_10m: number
  is_day?: number
}

interface DailyWeather {
  time?: string[]
  weather_code?: number[]
  temperature_2m_max?: number[]
  temperature_2m_min?: number[]
}

interface Forecast {
  current?: CurrentWeather
  daily?: DailyWeather
}

type WeatherStatus = 'loading' | 'ready' | 'error' | 'unconfigured'

interface ResolvedUnits {
  temperature: TemperatureUnit
  wind: WindUnit
}

const WEATHER_LABELS: Record<number, string> = {
  0: 'Clear',
  1: 'Mainly clear',
  2: 'Partly cloudy',
  3: 'Overcast',
  45: 'Fog',
  48: 'Rime fog',
  51: 'Light drizzle',
  53: 'Drizzle',
  55: 'Dense drizzle',
  56: 'Freezing drizzle',
  57: 'Freezing drizzle',
  61: 'Light rain',
  63: 'Rain',
  65: 'Heavy rain',
  66: 'Freezing rain',
  67: 'Freezing rain',
  71: 'Light snow',
  73: 'Snow',
  75: 'Heavy snow',
  77: 'Snow grains',
  80: 'Light showers',
  81: 'Showers',
  82: 'Violent showers',
  85: 'Snow showers',
  86: 'Heavy snow showers',
  95: 'Thunderstorm',
  96: 'Thunderstorm with hail',
  99: 'Thunderstorm with hail',
}

export function weatherLabel(code: number): string {
  return WEATHER_LABELS[code] ?? 'Cloudy'
}

export function weatherIconName(code: number, isDay = true): string {
  if (code === 0) return isDay ? 'clear-day' : 'clear-night'
  if (code === 1 || code === 2) return isDay ? 'partly-cloudy-day' : 'partly-cloudy-night'
  if (code === 3) return 'overcast'
  if (code === 45 || code === 48) return 'fog'
  if (code >= 51 && code <= 57) return 'drizzle'
  if (code >= 61 && code <= 67) return 'rain'
  if ((code >= 71 && code <= 77) || code === 85 || code === 86) return 'snow'
  if (code >= 80 && code <= 82) return 'showers-day'
  if (code >= 95 && code <= 99) return 'thunderstorms'
  return 'cloudy'
}

function iconSrc(code: number, isDay: boolean): string {
  return `/weather/${weatherIconName(code, isDay)}.svg`
}

function rounded(value: number | undefined): string {
  return value === undefined || Number.isNaN(value) ? '—' : String(Math.round(value))
}

function weekdayLabel(iso: string | undefined): string {
  if (!iso) return ''
  const date = new Date(`${iso}T00:00:00`)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('en-US', { weekday: 'short' })
}

function currentIsDay(current: CurrentWeather): boolean {
  return typeof current.is_day === 'number' ? current.is_day === 1 : true
}

export interface WeatherWidgetProps {
  widget: Widget
}

/**
 * WeatherWidget binds no Home Assistant entity: it resolves a location (a per-widget override or
 * the server-wide `/api/weather/config`) and calls Open-Meteo directly from the browser, which
 * allows CORS and needs no API key. It refreshes every 15 minutes, and a failed refresh keeps the
 * last good forecast visible while reporting staleness through the frame.
 */
export const WeatherWidget: React.FC<WeatherWidgetProps> = ({ widget }) => {
  const config = widget.config
  const mode: WeatherMode = config.weather_mode ?? 'current'
  const units: WeatherUnits = config.units ?? ''
  const overrideLat = typeof config.latitude === 'number' ? config.latitude : undefined
  const overrideLon = typeof config.longitude === 'number' ? config.longitude : undefined
  const locationName = config.location_name?.trim() ?? ''
  const configuredDays = Math.min(
    MAX_DAYS,
    Math.max(MIN_DAYS, Math.trunc(config.weather_days ?? DEFAULT_DAYS)),
  )
  const forecastDays = Math.max(2, mode === 'ndays' ? configuredDays : 1)

  const [status, setStatus] = useState<WeatherStatus>('loading')
  const [forecast, setForecast] = useState<Forecast | null>(null)
  const [resolvedName, setResolvedName] = useState(locationName)
  const [resolvedUnits, setResolvedUnits] = useState<ResolvedUnits>({
    temperature: units === 'imperial' ? 'fahrenheit' : 'celsius',
    wind: units === 'imperial' ? 'mph' : 'kmh',
  })
  const [stale, setStale] = useState(false)

  useEffect(() => {
    let cancelled = false
    let timer: number | null = null

    const load = async () => {
      let serverConfig: WeatherConfig | null = null
      let configOk = false
      try {
        const configRes = await apiFetch('/api/weather/config')
        if (configRes.ok) {
          configOk = true
          const payload = await configRes.json()
          serverConfig = (payload?.data ?? null) as WeatherConfig | null
        }
      } catch {
        configOk = false
      }

      try {
        const hasOverride = overrideLat !== undefined && overrideLon !== undefined
        let latitude: number | undefined
        let longitude: number | undefined
        if (hasOverride) {
          latitude = overrideLat
          longitude = overrideLon
        } else if (configOk && serverConfig?.configured) {
          latitude = serverConfig.latitude
          longitude = serverConfig.longitude
        }

        if (latitude === undefined || longitude === undefined) {
          if (configOk && !hasOverride) {
            if (!cancelled) setStatus('unconfigured')
            return
          }
          throw new Error('Weather location is unavailable')
        }

        const temperature: TemperatureUnit =
          units === 'metric'
            ? 'celsius'
            : units === 'imperial'
              ? 'fahrenheit'
              : (serverConfig?.temperature_unit ?? 'celsius')
        const wind: WindUnit =
          units === 'metric' ? 'kmh' : units === 'imperial' ? 'mph' : (serverConfig?.wind_unit ?? 'kmh')

        const params = new URLSearchParams({
          latitude: String(latitude),
          longitude: String(longitude),
          current: 'temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m,is_day',
          daily: 'weather_code,temperature_2m_max,temperature_2m_min',
          timezone: 'auto',
          forecast_days: String(forecastDays),
          temperature_unit: temperature,
          wind_speed_unit: wind,
        })

        const res = await fetch(`${OPEN_METEO_URL}?${params.toString()}`)
        if (!res.ok) throw new Error(`Open-Meteo responded ${res.status}`)
        const data = (await res.json()) as Forecast

        if (!cancelled) {
          setForecast(data)
          setResolvedName(locationName || serverConfig?.name || '')
          setResolvedUnits({ temperature, wind })
          setStale(false)
          setStatus('ready')
        }
      } catch {
        if (!cancelled) {
          setStatus((previous) => {
            if (previous === 'ready') {
              setStale(true)
              return 'ready'
            }
            return 'error'
          })
        }
      } finally {
        if (!cancelled) timer = window.setTimeout(load, REFRESH_MS)
      }
    }

    setStatus('loading')
    setStale(false)
    load()
    return () => {
      cancelled = true
      if (timer !== null) window.clearTimeout(timer)
    }
  }, [mode, units, overrideLat, overrideLon, locationName, forecastDays])

  const label = widget.title || resolvedName || 'Weather'
  const dense = widget.row_span === 1
  const frameIcon = <CloudSun className="text-[#4bb8c9]" />

  const errorFrame = (
    <WidgetFrame label={label} stale={false} dense={dense} icon={frameIcon} bodyClassName="flex items-center justify-center">
      <span className="text-xs text-slate-500 text-center px-2">Weather unavailable</span>
    </WidgetFrame>
  )

  if (status === 'unconfigured') {
    return (
      <WidgetFrame label={label} stale={false} dense={dense} icon={frameIcon} bodyClassName="flex items-center justify-center">
        <div className="flex flex-col items-center justify-center gap-1 text-center px-2">
          <MapPin className="w-5 h-5 text-slate-500" />
          <span className="text-xs text-slate-400">Set a location</span>
        </div>
      </WidgetFrame>
    )
  }

  if (status === 'loading') {
    return (
      <WidgetFrame label={label} stale={false} dense={dense} icon={frameIcon} bodyClassName="flex items-center justify-center">
        <span className="w-4 h-4 border-2 border-[#4bb8c9]/30 border-t-[#4bb8c9] rounded-full animate-spin" />
      </WidgetFrame>
    )
  }

  if (status === 'error' || !forecast) return errorFrame

  const tempUnitLabel = resolvedUnits.temperature === 'fahrenheit' ? '°F' : '°C'
  const windUnitLabel = resolvedUnits.wind === 'mph' ? 'mph' : 'km/h'

  if (mode === 'current') {
    const current = forecast.current
    if (!current) return errorFrame
    return (
      <WidgetFrame label={label} stale={stale} dense={dense} icon={frameIcon} bodyClassName="flex items-center justify-center">
        <div className="h-full w-full flex flex-col items-center justify-center gap-0.5">
          <img
            src={iconSrc(current.weather_code, currentIsDay(current))}
            alt=""
            className="w-[40%] h-auto max-h-[38%] object-contain"
          />
          <div className="flex items-start justify-center leading-none">
            <span
              className="font-semibold tracking-tight text-white tabular-nums"
              style={{ fontSize: 'clamp(1.5rem, min(34cqh, 30cqw), 3rem)' }}
            >
              {rounded(current.temperature_2m)}
            </span>
            <span
              className="font-medium text-slate-400 ml-0.5"
              style={{ fontSize: 'clamp(0.75rem, min(13cqh, 10cqw), 1.25rem)' }}
            >
              {tempUnitLabel}
            </span>
          </div>
          <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-400 truncate max-w-full">
            {weatherLabel(current.weather_code)}
          </span>
          <div className="flex items-center gap-3 text-[10px] text-slate-400 tabular-nums">
            <span className="flex items-center gap-0.5">
              <Droplets className="w-3 h-3 text-[#4bb8c9]" />
              {current.relative_humidity_2m}%
            </span>
            <span className="flex items-center gap-0.5">
              <Wind className="w-3 h-3 text-slate-300" />
              {rounded(current.wind_speed_10m)} {windUnitLabel}
            </span>
          </div>
        </div>
      </WidgetFrame>
    )
  }

  const daily = forecast.daily
  if (!daily) return errorFrame

  if (mode === 'today' || mode === 'tomorrow') {
    const index = mode === 'tomorrow' ? 1 : 0
    const code = daily.weather_code?.[index]
    return (
      <WidgetFrame label={label} stale={stale} dense={dense} icon={frameIcon} bodyClassName="flex items-center justify-center">
        <div className="h-full w-full flex items-center justify-center gap-2 px-1">
          <img src={iconSrc(code ?? -1, true)} alt="" className="w-[32%] h-auto max-h-[72%] object-contain" />
          <div className="min-w-0 flex flex-col justify-center gap-1">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-400 truncate">{weatherLabel(code ?? -1)}</span>
            <span className="text-sm font-semibold tracking-tight text-white tabular-nums leading-none whitespace-nowrap">
              {rounded(daily.temperature_2m_max?.[index])}°
              <span className="text-slate-400 font-medium text-xs ml-1">{rounded(daily.temperature_2m_min?.[index])}°</span>
            </span>
          </div>
        </div>
      </WidgetFrame>
    )
  }

  const shownDays = Math.min(configuredDays, daily.time?.length ?? configuredDays)
  return (
    <WidgetFrame label={label} stale={stale} dense={dense} icon={frameIcon} bodyClassName="flex items-stretch">
      <div className="h-full w-full flex items-stretch gap-0.5 overflow-x-auto">
        {Array.from({ length: shownDays }, (_, index) => {
          const code = daily.weather_code?.[index]
          return (
            <div key={daily.time?.[index] ?? index} className="flex-1 min-w-0 flex flex-col items-center justify-center gap-1">
              <span className="text-[9px] font-semibold uppercase tracking-wider text-slate-400 truncate max-w-full">
                {weekdayLabel(daily.time?.[index])}
              </span>
              <img src={iconSrc(code ?? -1, true)} alt="" className="w-6 h-6 max-w-full object-contain" />
              <span className="text-[10px] font-semibold tracking-tight text-white tabular-nums leading-none">
                {rounded(daily.temperature_2m_max?.[index])}°
              </span>
              <span className="text-[9px] text-slate-500 tabular-nums leading-none">
                {rounded(daily.temperature_2m_min?.[index])}°
              </span>
            </div>
          )
        })}
      </div>
    </WidgetFrame>
  )
}
