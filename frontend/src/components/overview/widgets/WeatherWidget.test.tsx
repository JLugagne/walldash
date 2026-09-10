import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { WeatherWidget, weatherIconName, weatherLabel } from './WeatherWidget'
import type { Widget } from '../../../types'

const CONFIG_UNCONFIGURED = {
  status: 'success',
  data: {
    configured: false,
    latitude: 0,
    longitude: 0,
    name: '',
    temperature_unit: 'celsius',
    wind_unit: 'kmh',
  },
}

const OPEN_METEO = {
  current: {
    temperature_2m: 21.4,
    relative_humidity_2m: 54,
    weather_code: 0,
    wind_speed_10m: 12.3,
    is_day: 1,
  },
  daily: {
    time: ['2026-09-10', '2026-09-11', '2026-09-12'],
    weather_code: [0, 3, 61],
    temperature_2m_max: [24.6, 22.1, 18.9],
    temperature_2m_min: [12.2, 11.4, 9.8],
  },
}

function jsonResponse(payload: unknown, ok = true, status = 200): Promise<Response> {
  return Promise.resolve({
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: async () => payload,
  } as Response)
}

function makeWidget(config: Widget['config']): Widget {
  return {
    id: 'w1',
    dashboard_id: 'd1',
    type: 'weather',
    title: '',
    order: 0,
    col: 0,
    row: 0,
    col_span: 2,
    row_span: 2,
    config,
    created_at: '',
    updated_at: '',
  }
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('weatherIconName', () => {
  it('maps WMO codes to bundled icons, with day/night variants', () => {
    expect(weatherIconName(0, true)).toBe('clear-day')
    expect(weatherIconName(0, false)).toBe('clear-night')
    expect(weatherIconName(1, false)).toBe('partly-cloudy-night')
    expect(weatherIconName(2, true)).toBe('partly-cloudy-day')
    expect(weatherIconName(3)).toBe('overcast')
    expect(weatherIconName(45)).toBe('fog')
    expect(weatherIconName(57)).toBe('drizzle')
    expect(weatherIconName(66)).toBe('rain')
    expect(weatherIconName(75)).toBe('snow')
    expect(weatherIconName(86)).toBe('snow')
    expect(weatherIconName(81)).toBe('showers-day')
    expect(weatherIconName(96)).toBe('thunderstorms')
    expect(weatherIconName(1234)).toBe('cloudy')
  })

  it('labels known codes and falls back to Cloudy', () => {
    expect(weatherLabel(0)).toBe('Clear')
    expect(weatherLabel(61)).toBe('Light rain')
    expect(weatherLabel(999)).toBe('Cloudy')
  })
})

describe('WeatherWidget', () => {
  it('renders current conditions resolved from the widget override', async () => {
    globalThis.fetch = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/api/weather/config')) return jsonResponse(CONFIG_UNCONFIGURED)
      return jsonResponse(OPEN_METEO)
    }) as unknown as typeof fetch

    render(
      <WeatherWidget
        widget={makeWidget({
          entity_ids: [],
          display: 'weather',
          weather_mode: 'current',
          latitude: 48.85,
          longitude: 2.35,
          location_name: 'Paris',
          units: 'metric',
        })}
      />
    )

    expect(await screen.findByText('21', undefined, { timeout: 5000 })).toBeDefined()
    expect(screen.getByText('Clear')).toBeDefined()
    expect(screen.getByText('°C')).toBeDefined()
    expect(screen.getByText('54%')).toBeDefined()
    expect(screen.getByText('Paris')).toBeDefined()
    expect(document.querySelector('img')?.getAttribute('src')).toBe('/weather/clear-day.svg')
  })

  it('shows the Set a location empty state when the endpoint is unconfigured', async () => {
    globalThis.fetch = vi.fn(() => jsonResponse(CONFIG_UNCONFIGURED)) as unknown as typeof fetch

    render(<WeatherWidget widget={makeWidget({ entity_ids: [], display: 'weather', weather_mode: 'current' })} />)

    expect(await screen.findByText('Set a location', undefined, { timeout: 5000 })).toBeDefined()
    expect(globalThis.fetch).toHaveBeenCalledTimes(1)
  })

  it('resolves location and units from the endpoint when the widget has no override', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/api/weather/config')) {
        return jsonResponse({
          status: 'success',
          data: {
            configured: true,
            latitude: 40.7,
            longitude: -74,
            name: 'New York',
            temperature_unit: 'fahrenheit',
            wind_unit: 'mph',
          },
        })
      }
      return jsonResponse(OPEN_METEO)
    })
    globalThis.fetch = fetchMock as unknown as typeof fetch

    render(<WeatherWidget widget={makeWidget({ entity_ids: [], display: 'weather', weather_mode: 'today' })} />)

    expect(await screen.findByText('New York', undefined, { timeout: 5000 })).toBeDefined()
    expect(screen.getByText(/25°/)).toBeDefined()

    const meteoUrl = fetchMock.mock.calls.map((call) => String(call[0])).find((url) => url.includes('open-meteo'))
    expect(meteoUrl).toContain('latitude=40.7')
    expect(meteoUrl).toContain('longitude=-74')
    expect(meteoUrl).toContain('temperature_unit=fahrenheit')
    expect(meteoUrl).toContain('wind_speed_unit=mph')
  })

  it('renders one column per requested day in ndays mode', async () => {
    globalThis.fetch = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/api/weather/config')) return jsonResponse(CONFIG_UNCONFIGURED)
      return jsonResponse(OPEN_METEO)
    }) as unknown as typeof fetch

    render(
      <WeatherWidget
        widget={makeWidget({
          entity_ids: [],
          display: 'weather',
          weather_mode: 'ndays',
          weather_days: 3,
          latitude: 48.85,
          longitude: 2.35,
          units: 'metric',
        })}
      />
    )

    expect(await screen.findByText('Thu', undefined, { timeout: 5000 })).toBeDefined()
    expect(screen.getByText('Fri')).toBeDefined()
    expect(screen.getByText('Sat')).toBeDefined()
  })

  it('shows a graceful error when the endpoint is unreachable and there is no override', async () => {
    globalThis.fetch = vi.fn(() => Promise.reject(new Error('offline'))) as unknown as typeof fetch

    render(<WeatherWidget widget={makeWidget({ entity_ids: [], display: 'weather', weather_mode: 'current' })} />)

    expect(await screen.findByText('Weather unavailable', undefined, { timeout: 5000 })).toBeDefined()
  })
})
