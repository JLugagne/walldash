import React, { useEffect, useMemo, useState } from 'react'
import { X, Check, Search, AlertTriangle } from 'lucide-react'
import type {
  Automation,
  Device,
  DisplayMode,
  WeatherMode,
  WeatherUnits,
  Widget,
  WidgetConfig,
  WidgetType,
} from '../types'
import { MIN_SIZE, findFreeArea, type Rect } from './overview/grid'

// Mirrors the legal Widget Type / Display Mode matrix from ADR 0004 and the
// Go domain (internal/dashboard/domain/overview.go). The Go matrix stays the
// single source of truth; this table only keeps the UI honest and lets the
// modal offer the right choices before the server ever sees a request.
const DISPLAY_OPTIONS_BY_TYPE: Record<WidgetType, DisplayMode[]> = {
  sensor: ['number', 'arc', 'bar'],
  actuator: ['toggle'],
  automation_list: ['list'],
  weather: ['weather'],
}

// Mirrors domain.AllowedActionDomains: the only entity domains an Actuator
// widget may bind to, since its Primary Action is a toggle service call.
const ACTUATOR_ALLOWED_DOMAINS = ['light', 'switch', 'media_player', 'automation']

const TYPE_LABELS: Record<WidgetType, string> = {
  sensor: 'Sensor',
  actuator: 'Actuator',
  automation_list: 'Automation list',
  weather: 'Weather',
}

const WEATHER_MODE_LABELS: Record<WeatherMode, string> = {
  current: 'Current',
  today: 'Today',
  tomorrow: 'Tomorrow',
  ndays: 'Forecast',
}

const WEATHER_UNIT_LABELS: Record<WeatherUnits, string> = {
  '': 'Auto',
  metric: 'Metric',
  imperial: 'Imperial',
}

const DEFAULT_WEATHER_DAYS = 5
const MIN_WEATHER_DAYS = 1
const MAX_WEATHER_DAYS = 14

function suggestUnit(devices: Device[], entityId: string | undefined): string {
  const attrUnit = devices.find((d) => d.id === entityId)?.attributes?.unit_of_measurement
  return typeof attrUnit === 'string' ? attrUnit : ''
}

const DISPLAY_LABELS: Record<DisplayMode, string> = {
  number: 'Number',
  arc: 'Gauge',
  bar: 'Bar',
  toggle: 'Toggle',
  list: 'List',
  weather: 'Weather',
}

export interface NewWidgetInput {
  type: WidgetType
  title: string
  order: number
  config: WidgetConfig
  col: number
  row: number
  col_span: number
  row_span: number
}

export interface WidgetContentInput {
  title: string
  config: WidgetConfig
}

interface AddWidgetModalProps {
  isOpen: boolean
  overviewCols: number
  overviewRows: number
  occupiedRects: Rect[]
  nextOrder: number
  automations: Automation[]
  devices: Device[]
  editingWidget: Widget | null
  onClose: () => void
  onCreate: (input: NewWidgetInput) => Promise<void>
  onUpdate: (widgetId: string, input: WidgetContentInput) => Promise<void>
}

export const AddWidgetModal: React.FC<AddWidgetModalProps> = ({
  isOpen,
  overviewCols,
  overviewRows,
  occupiedRects,
  nextOrder,
  automations,
  devices,
  editingWidget,
  onClose,
  onCreate,
  onUpdate,
}) => {
  const isEditing = !!editingWidget

  const [type, setType] = useState<WidgetType>('sensor')
  const [display, setDisplay] = useState<DisplayMode>('number')
  const [title, setTitle] = useState('')
  const [selectedEntityIds, setSelectedEntityIds] = useState<string[]>([])
  const [labels, setLabels] = useState<Record<string, string>>({})
  const [minValue, setMinValue] = useState('')
  const [maxValue, setMaxValue] = useState('')
  const [unit, setUnit] = useState('')
  const [weatherMode, setWeatherMode] = useState<WeatherMode>('current')
  const [weatherDays, setWeatherDays] = useState(String(DEFAULT_WEATHER_DAYS))
  const [weatherUnits, setWeatherUnits] = useState<WeatherUnits>('')
  const [locationName, setLocationName] = useState('')
  const [latitude, setLatitude] = useState('')
  const [longitude, setLongitude] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [saving, setSaving] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  useEffect(() => {
    if (!isOpen) return
    setSubmitError(null)
    if (editingWidget) {
      setType(editingWidget.type)
      setDisplay(editingWidget.config.display)
      setTitle(editingWidget.title)
      setSelectedEntityIds(editingWidget.config.entity_ids || [])
      setLabels(editingWidget.config.labels || {})
      setMinValue(editingWidget.config.min !== undefined ? String(editingWidget.config.min) : '')
      setMaxValue(editingWidget.config.max !== undefined ? String(editingWidget.config.max) : '')
      setUnit(editingWidget.config.unit || '')
      setWeatherMode(editingWidget.config.weather_mode || 'current')
      setWeatherDays(
        editingWidget.config.weather_days !== undefined
          ? String(editingWidget.config.weather_days)
          : String(DEFAULT_WEATHER_DAYS)
      )
      setWeatherUnits(editingWidget.config.units || '')
      setLocationName(editingWidget.config.location_name || '')
      setLatitude(editingWidget.config.latitude !== undefined ? String(editingWidget.config.latitude) : '')
      setLongitude(editingWidget.config.longitude !== undefined ? String(editingWidget.config.longitude) : '')
    } else {
      setType('sensor')
      setDisplay('number')
      setTitle('')
      setSelectedEntityIds([])
      setLabels({})
      setMinValue('')
      setMaxValue('')
      setUnit('')
      setWeatherMode('current')
      setWeatherDays(String(DEFAULT_WEATHER_DAYS))
      setWeatherUnits('')
      setLocationName('')
      setLatitude('')
      setLongitude('')
    }
    setSearchTerm('')
  }, [isOpen, editingWidget])

  useEffect(() => {
    if (type !== 'sensor') return
    setUnit((current) => current || suggestUnit(devices, selectedEntityIds[0]))
  }, [selectedEntityIds, type, devices])

  useEffect(() => {
    setSubmitError(null)
  }, [
    type,
    display,
    title,
    selectedEntityIds,
    labels,
    minValue,
    maxValue,
    unit,
    weatherMode,
    weatherDays,
    weatherUnits,
    locationName,
    latitude,
    longitude,
  ])

  const minSize = MIN_SIZE[display]

  const eligibleDevices = useMemo(() => {
    if (type === 'actuator') {
      return devices.filter((d) => ACTUATOR_ALLOWED_DOMAINS.includes(d.domain))
    }
    return devices
  }, [devices, type])

  const filteredAutomations = useMemo(
    () =>
      automations.filter(
        (a) =>
          a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          a.id.toLowerCase().includes(searchTerm.toLowerCase())
      ),
    [automations, searchTerm]
  )

  const filteredDevices = useMemo(
    () =>
      eligibleDevices.filter(
        (d) =>
          d.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          d.id.toLowerCase().includes(searchTerm.toLowerCase())
      ),
    [eligibleDevices, searchTerm]
  )

  const placement = useMemo(() => {
    if (isEditing) return null
    return findFreeArea(occupiedRects, { cols: overviewCols, rows: overviewRows }, minSize.cols, minSize.rows)
  }, [isEditing, occupiedRects, overviewCols, overviewRows, minSize.cols, minSize.rows])

  if (!isOpen) return null

  const availableDisplays = DISPLAY_OPTIONS_BY_TYPE[type]

  const handleTypeChange = (nextType: WidgetType) => {
    setType(nextType)
    setDisplay(DISPLAY_OPTIONS_BY_TYPE[nextType][0])
    setSelectedEntityIds([])
    setUnit('')
  }

  const handleDisplayChange = (nextDisplay: DisplayMode) => {
    setDisplay(nextDisplay)
    if (nextDisplay !== 'arc' && nextDisplay !== 'bar') {
      setMinValue('')
      setMaxValue('')
    }
  }

  const toggleEntity = (id: string) => {
    setSelectedEntityIds((prev) => {
      if (type === 'automation_list') {
        return prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
      }
      return prev.includes(id) ? [] : [id]
    })
  }

  const isWeather = type === 'weather'

  const cardinalityOk = isWeather
    ? true
    : type === 'automation_list'
      ? selectedEntityIds.length >= 1
      : selectedEntityIds.length === 1

  const boundsRequired = display === 'arc' || display === 'bar'
  const parsedMin = minValue.trim() === '' ? undefined : Number(minValue)
  const parsedMax = maxValue.trim() === '' ? undefined : Number(maxValue)
  const boundsOk =
    !boundsRequired ||
    (parsedMin !== undefined &&
      parsedMax !== undefined &&
      !Number.isNaN(parsedMin) &&
      !Number.isNaN(parsedMax) &&
      parsedMin < parsedMax)

  const parsedWeatherDays = weatherDays.trim() === '' ? undefined : Number(weatherDays)
  const weatherDaysOk =
    !isWeather ||
    weatherMode !== 'ndays' ||
    (parsedWeatherDays !== undefined &&
      Number.isInteger(parsedWeatherDays) &&
      parsedWeatherDays >= MIN_WEATHER_DAYS &&
      parsedWeatherDays <= MAX_WEATHER_DAYS)

  const hasLatitude = latitude.trim() !== ''
  const hasLongitude = longitude.trim() !== ''
  const parsedLatitude = hasLatitude ? Number(latitude) : undefined
  const parsedLongitude = hasLongitude ? Number(longitude) : undefined
  const locationOk =
    !isWeather ||
    (hasLatitude === hasLongitude &&
      (!hasLatitude ||
        (parsedLatitude !== undefined &&
          !Number.isNaN(parsedLatitude) &&
          parsedLatitude >= -90 &&
          parsedLatitude <= 90 &&
          parsedLongitude !== undefined &&
          !Number.isNaN(parsedLongitude) &&
          parsedLongitude >= -180 &&
          parsedLongitude <= 180)))

  const noRoomReason =
    !isEditing && !placement
      ? "No space available on the grid for this widget. Free up space or choose a more compact display mode."
      : null

  const tooSmallForEditReason =
    isEditing && editingWidget && (minSize.cols > editingWidget.col_span || minSize.rows > editingWidget.row_span)
      ? "This display mode needs a larger widget. Resize it in edit mode before changing the display."
      : null

  const cardinalityReason = !cardinalityOk
    ? type === 'automation_list'
      ? 'Select at least one automation.'
      : 'Select exactly one entity for this widget.'
    : null

  const boundsReason = !boundsOk ? 'Enter a minimum strictly lower than the maximum.' : null

  const weatherDaysReason = !weatherDaysOk
    ? `Enter a whole number of days between ${MIN_WEATHER_DAYS} and ${MAX_WEATHER_DAYS}.`
    : null

  const locationReason = !locationOk
    ? hasLatitude !== hasLongitude
      ? 'Enter both latitude and longitude, or leave both empty.'
      : 'Enter a latitude between -90 and 90 and a longitude between -180 and 180.'
    : null

  const blockingReason =
    noRoomReason || tooSmallForEditReason || cardinalityReason || boundsReason || weatherDaysReason || locationReason

  const canSubmit = !saving && !blockingReason

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return

    setSaving(true)
    setSubmitError(null)
    try {
      const config: WidgetConfig = isWeather
        ? {
            entity_ids: [],
            display: 'weather',
            weather_mode: weatherMode,
            ...(weatherMode === 'ndays' && parsedWeatherDays !== undefined
              ? { weather_days: parsedWeatherDays }
              : {}),
            ...(weatherUnits ? { units: weatherUnits } : {}),
            ...(locationName.trim() ? { location_name: locationName.trim() } : {}),
            ...(parsedLatitude !== undefined && parsedLongitude !== undefined
              ? { latitude: parsedLatitude, longitude: parsedLongitude }
              : {}),
          }
        : {
            entity_ids: selectedEntityIds,
            display,
            ...(Object.keys(labels).length > 0 ? { labels } : {}),
            ...(parsedMin !== undefined ? { min: parsedMin } : {}),
            ...(parsedMax !== undefined ? { max: parsedMax } : {}),
            ...(unit.trim() ? { unit: unit.trim() } : {}),
          }

      if (isEditing && editingWidget) {
        await onUpdate(editingWidget.id, { title: title.trim(), config })
      } else if (placement) {
        await onCreate({
          type,
          title: title.trim(),
          order: nextOrder,
          config,
          col: placement.col,
          row: placement.row,
          col_span: minSize.cols,
          row_span: minSize.rows,
        })
      }
      onClose()
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Failed to save the widget.")
    } finally {
      setSaving(false)
    }
  }

  const entityDisplayName = (id: string): string => {
    if (type === 'automation_list') {
      return automations.find((a) => a.id === id)?.name || id
    }
    return devices.find((d) => d.id === id)?.name || id
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-2xl w-full max-w-xl shadow-xl overflow-hidden flex flex-col max-h-[85vh]">
        <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between bg-slate-900/60">
          <div>
            <h2 className="text-base font-bold text-white">
              {isEditing ? 'Edit Widget' : 'Add Widget'}
            </h2>
            <p className="text-xs text-slate-400">
              {isEditing
                ? 'The widget type and position cannot be changed here.'
                : 'Choose what the widget displays; its size is calculated automatically.'}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 text-slate-400 hover:text-white rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 flex flex-col overflow-hidden">
          <div className="p-6 space-y-4 flex-1 overflow-y-auto">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Widget Type</label>
              {isEditing ? (
                <div className="px-3.5 py-2 rounded-lg bg-slate-800/50 border border-slate-800 text-sm text-slate-300">
                  {TYPE_LABELS[type]}
                </div>
              ) : (
                <div className="grid grid-cols-2 gap-2">
                  {(Object.keys(DISPLAY_OPTIONS_BY_TYPE) as WidgetType[]).map((t) => (
                    <button
                      key={t}
                      type="button"
                      onClick={() => handleTypeChange(t)}
                      className={`px-3 py-2 rounded-lg text-xs font-semibold border transition-colors ${
                        type === t
                          ? 'bg-[#6d76e8]/20 border-[#6d76e8] text-white'
                          : 'bg-slate-800/50 border-slate-800 text-slate-400 hover:border-slate-700'
                      }`}
                    >
                      {TYPE_LABELS[t]}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Display mode</label>
              <div className="flex flex-wrap gap-2">
                {availableDisplays.map((d) => (
                  <button
                    key={d}
                    type="button"
                    onClick={() => handleDisplayChange(d)}
                    className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors ${
                      display === d
                        ? 'bg-[#6d76e8]/20 border-[#6d76e8] text-white'
                        : 'bg-slate-800/50 border-slate-800 text-slate-400 hover:border-slate-700'
                    }`}
                  >
                    {DISPLAY_LABELS[d]}
                  </button>
                ))}
              </div>
            </div>

            <div>
<label className="block text-xs font-semibold text-slate-300 mb-1.5">
                  Widget Title <span className="text-slate-500 font-normal">(optional)</span>
                </label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder={isWeather ? 'Uses the location name, otherwise' : 'Uses the entity label, or its Home Assistant name, otherwise'}
                  className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3.5 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8] focus:ring-1 focus:ring-[#6d76e8] transition-colors"
              />
            </div>

            {isWeather && (
              <div className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Weather mode</label>
                  <div className="flex flex-wrap gap-2">
                    {(Object.keys(WEATHER_MODE_LABELS) as WeatherMode[]).map((m) => (
                      <button
                        key={m}
                        type="button"
                        onClick={() => setWeatherMode(m)}
                        className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors ${
                          weatherMode === m
                            ? 'bg-[#6d76e8]/20 border-[#6d76e8] text-white'
                            : 'bg-slate-800/50 border-slate-800 text-slate-400 hover:border-slate-700'
                        }`}
                      >
                        {WEATHER_MODE_LABELS[m]}
                      </button>
                    ))}
                  </div>
                </div>

                {weatherMode === 'ndays' && (
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1.5">Forecast days</label>
                    <input
                      type="number"
                      min={MIN_WEATHER_DAYS}
                      max={MAX_WEATHER_DAYS}
                      value={weatherDays}
                      onChange={(e) => setWeatherDays(e.target.value)}
                      className="w-32 bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#6d76e8]"
                    />
                  </div>
                )}

                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Units</label>
                  <div className="flex flex-wrap gap-2">
                    {(Object.keys(WEATHER_UNIT_LABELS) as WeatherUnits[]).map((u) => (
                      <button
                        key={u || 'auto'}
                        type="button"
                        onClick={() => setWeatherUnits(u)}
                        className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors ${
                          weatherUnits === u
                            ? 'bg-[#6d76e8]/20 border-[#6d76e8] text-white'
                            : 'bg-slate-800/50 border-slate-800 text-slate-400 hover:border-slate-700'
                        }`}
                      >
                        {WEATHER_UNIT_LABELS[u]}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">
                    Location override <span className="text-slate-500 font-normal">(optional)</span>
                  </label>
                  <div className="grid grid-cols-2 gap-3 mb-2">
                    <input
                      type="number"
                      step="any"
                      value={latitude}
                      onChange={(e) => setLatitude(e.target.value)}
                      placeholder="Latitude"
                      className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8]"
                    />
                    <input
                      type="number"
                      step="any"
                      value={longitude}
                      onChange={(e) => setLongitude(e.target.value)}
                      placeholder="Longitude"
                      className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8]"
                    />
                  </div>
                  <input
                    type="text"
                    value={locationName}
                    onChange={(e) => setLocationName(e.target.value)}
                    placeholder="Location name (optional)"
                    className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8]"
                  />
                </div>
              </div>
            )}

            {boundsRequired && (
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Min</label>
                  <input
                    type="number"
                    value={minValue}
                    onChange={(e) => setMinValue(e.target.value)}
                    className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#6d76e8]"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Max</label>
                  <input
                    type="number"
                    value={maxValue}
                    onChange={(e) => setMaxValue(e.target.value)}
                    className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#6d76e8]"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Unit</label>
                  <input
                    type="text"
                    value={unit}
                    onChange={(e) => setUnit(e.target.value)}
                    placeholder="°C"
                    className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#6d76e8]"
                  />
                </div>
              </div>
            )}

            {display === 'number' && (
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1.5">
                  Unit <span className="text-slate-500 font-normal">(optional)</span>
                </label>
                <input
                  type="text"
                  value={unit}
                  onChange={(e) => setUnit(e.target.value)}
                  placeholder="%"
                  className="w-full bg-slate-800/80 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#6d76e8]"
                />
              </div>
            )}

            {!isWeather && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="text-xs font-semibold text-slate-300">
                  {type === 'automation_list'
                    ? `Automations (${selectedEntityIds.length} selected${selectedEntityIds.length > 1 ? '' : ''})`
                    : 'Linked Entity'}
                </label>
              </div>

              <div className="relative mb-3">
                <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder={type === 'automation_list' ? 'Search automation...' : 'Search device...'}
                  className="w-full bg-slate-800/50 border border-slate-700/80 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8]"
                />
              </div>

              <div className="space-y-2 max-h-48 overflow-y-auto border border-slate-800 rounded-lg p-2 bg-slate-950/40">
                {type === 'automation_list' ? (
                  filteredAutomations.length === 0 ? (
                    <div className="text-center py-6 text-xs text-slate-500">No automations found</div>
                  ) : (
                    filteredAutomations.map((auto) => {
                      const isSelected = selectedEntityIds.includes(auto.id)
                      return (
                        <div
                          key={auto.id}
                          onClick={() => toggleEntity(auto.id)}
                          className={`p-3 rounded-lg border cursor-pointer transition-all flex items-center justify-between select-none ${
                            isSelected
                              ? 'bg-[#6d76e8]/15 border-[#6d76e8]/40 text-white'
                              : 'bg-slate-900/60 border-slate-800/70 text-slate-300 hover:border-slate-700'
                          }`}
                        >
                          <div className="min-w-0 flex-1 pr-3">
                            <p className="text-xs font-semibold truncate">{auto.name}</p>
                            <p className="text-[10px] text-slate-500 font-mono truncate">{auto.id}</p>
                          </div>
                          <div
                            className={`w-5 h-5 rounded-md border flex items-center justify-center transition-colors ${
                              isSelected ? 'bg-[#6d76e8] border-[#6d76e8] text-white' : 'border-slate-700 bg-slate-800'
                            }`}
                          >
                            {isSelected && <Check className="w-3.5 h-3.5" />}
                          </div>
                        </div>
                      )
                    })
                  )
                ) : filteredDevices.length === 0 ? (
                  <div className="text-center py-6 text-xs text-slate-500">No devices found</div>
                ) : (
                  filteredDevices.map((device) => {
                    const isSelected = selectedEntityIds.includes(device.id)
                    return (
                      <div
                        key={device.id}
                        onClick={() => toggleEntity(device.id)}
                        className={`p-3 rounded-lg border cursor-pointer transition-all flex items-center justify-between select-none ${
                          isSelected
                            ? 'bg-[#6d76e8]/15 border-[#6d76e8]/40 text-white'
                            : 'bg-slate-900/60 border-slate-800/70 text-slate-300 hover:border-slate-700'
                        }`}
                      >
                        <div className="min-w-0 flex-1 pr-3">
                          <p className="text-xs font-semibold truncate">{device.name}</p>
                          <p className="text-[10px] text-slate-500 font-mono truncate">
                            {device.id} • {device.domain}
                          </p>
                        </div>
                        <div
                          className={`w-5 h-5 rounded-full border flex items-center justify-center transition-colors ${
                            isSelected ? 'bg-[#6d76e8] border-[#6d76e8] text-white' : 'border-slate-700 bg-slate-800'
                          }`}
                        >
                          {isSelected && <Check className="w-3.5 h-3.5" />}
                        </div>
                      </div>
                    )
                  })
                )}
              </div>
            </div>
            )}

            {selectedEntityIds.length > 0 && (
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1.5">
                  Custom labels <span className="text-slate-500 font-normal">(optional)</span>
                </label>
                <div className="space-y-2">
                  {selectedEntityIds.map((id) => (
                    <div key={id} className="flex items-center gap-2">
                      <span className="text-[11px] text-slate-500 w-1/3 truncate" title={entityDisplayName(id)}>
                        {entityDisplayName(id)}
                      </span>
                      <input
                        type="text"
                        value={labels[id] || ''}
                        onChange={(e) => setLabels((prev) => ({ ...prev, [id]: e.target.value }))}
                        placeholder={entityDisplayName(id)}
                        className="flex-1 bg-slate-800/80 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-[#6d76e8]"
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="px-6 py-4 border-t border-slate-800/80 bg-slate-900/60 flex items-center justify-between gap-3">
            {submitError ? (
              <div className="flex items-center gap-1.5 text-[11px] text-red-400 min-w-0">
                <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                <span className="truncate">{submitError}</span>
              </div>
            ) : blockingReason ? (
              <div className="flex items-center gap-1.5 text-[11px] text-[#e0b060] min-w-0">
                <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                <span className="truncate">{blockingReason}</span>
              </div>
            ) : (
              <span />
            )}
            <div className="flex items-center space-x-3 shrink-0">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={!canSubmit}
                title={blockingReason || undefined}
                className="px-5 py-2 text-xs font-bold bg-[#6d76e8] hover:bg-[#7b83ea] active:bg-[#5b64d4] text-white rounded-lg transition-all active:scale-95 disabled:opacity-50"
              >
                {saving ? 'Saving...' : isEditing ? 'Save' : 'Create Widget'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}
