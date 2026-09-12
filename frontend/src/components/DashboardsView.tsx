import React, { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { LayoutDashboard, Plus, RefreshCw, Settings } from 'lucide-react'
import type { Device, Widget } from '../types'
import { AutomationListWidget } from './dashboard/widgets/AutomationListWidget'
import { WeatherWidget } from './dashboard/widgets/WeatherWidget'
import { AddWidgetModal, type NewWidgetInput, type WidgetContentInput } from './AddWidgetModal'
import { apiFetch, readApiError } from '../api'
import { useRealtimeDevices } from '../hooks/useRealtimeDevices'
import { NumberWidget } from './dashboard/widgets/NumberWidget'
import { BarWidget } from './dashboard/widgets/BarWidget'
import { ArcWidget } from './dashboard/widgets/ArcWidget'
import { ToggleWidget } from './dashboard/widgets/ToggleWidget'
import { DashboardHeader } from './dashboard/DashboardHeader'
import { DashboardSelector } from './dashboard/DashboardSelector'
import { useTopBarSlot } from './TopBarSlot'
import { useSetupBannerSlot } from './setup/SetupBannerSlot'
import { WidgetGrid } from './dashboard/WidgetGrid'
import { Toast } from './dashboard/Toast'
import { useDashboardData } from './dashboard/useDashboardData'
import { DashboardBackgroundPanel, type BackgroundConfig } from './dashboard/DashboardBackgroundPanel'
import {
  DEFAULT_BACKGROUND_BLUR,
  DEFAULT_BACKGROUND_DIM,
  DEFAULT_BACKGROUND_OPACITY,
} from './dashboard/backgroundSamples'
import { widgetsToRects } from './dashboard/format'
import type { Rect } from './dashboard/grid'

interface DashboardsViewProps {
  mode?: 'display' | 'edit'
  dashboardId?: string
}

const STALE_MS = 30 * 60 * 1000
const TOAST_MS = 3000
const NETWORK_ERROR_MESSAGE = 'Unable to reach the server, change not saved.'
/** Rolling sample window per numeric entity, fed to the sensor sparklines. */
const HISTORY_LENGTH = 24
const HISTORY_SAMPLE_MS = 5000

function parseNumericState(state: string | undefined): number | null {
  if (state === undefined) return null
  const n = Number(state)
  return Number.isFinite(n) ? n : null
}

function isStaleDevice(device: Device | undefined): boolean {
  if (!device) return true
  if (device.state === 'unavailable' || device.state === 'unknown') return true
  if (!device.last_updated) return false
  const updatedAt = new Date(device.last_updated).getTime()
  if (Number.isNaN(updatedAt)) return false
  return Date.now() - updatedAt > STALE_MS
}

/** Remembers the active dashboard across setup <-> view navigation within the tab. */
function readStoredActiveDashboardId(): string | null {
  try {
    return window.sessionStorage.getItem('walldash:active-dashboard')
  } catch {
    return null
  }
}

export const DashboardsView: React.FC<DashboardsViewProps> = ({
  mode = 'display',
  dashboardId,
}) => {
  const navigate = useNavigate()
  const isEdit = mode === 'edit'
  const [activeDashboardId, setActiveDashboardId] = useState<string | null>(
    () => dashboardId ?? readStoredActiveDashboardId(),
  )
  const { dashboards, automations, loading, fetchDashboards, fetchAutomations } =
    useDashboardData(setActiveDashboardId)
  const [isEditMode, setIsEditMode] = useState(false)
  const [isAddWidgetOpen, setIsAddWidgetOpen] = useState(false)
  const [editingWidget, setEditingWidget] = useState<Widget | null>(null)
  const [toastMessage, setToastMessage] = useState<string | null>(null)
  const [isBackgroundOpen, setIsBackgroundOpen] = useState(false)
  const [previewBackground, setPreviewBackground] = useState<BackgroundConfig | null>(null)
  const backgroundTimerRef = useRef<number | null>(null)

  const [lastKnownValues, setLastKnownValues] = useState<Record<string, number>>({})
  const [histories, setHistories] = useState<Record<string, number[]>>({})

  const { deviceMap, devices, pendingDevices, toggleDevice } = useRealtimeDevices(null)

  useEffect(() => {
    if (dashboardId) setActiveDashboardId(dashboardId)
  }, [dashboardId])

  useEffect(() => {
    if (!activeDashboardId) return
    try {
      window.sessionStorage.setItem('walldash:active-dashboard', activeDashboardId)
    } catch {
      /* sessionStorage unavailable */
    }
  }, [activeDashboardId])

  useEffect(() => {
    setIsEditMode(false)
    setIsAddWidgetOpen(false)
    setEditingWidget(null)
    setIsBackgroundOpen(false)
    setPreviewBackground(null)
  }, [activeDashboardId])

  useEffect(() => {
    setLastKnownValues((prev) => {
      let changed = false
      const next = { ...prev }
      for (const device of Object.values(deviceMap)) {
        if (device.state !== 'unavailable' && device.state !== 'unknown') {
          const num = parseNumericState(device.state)
          if (num !== null && next[device.id] !== num) {
            next[device.id] = num
            changed = true
          }
        }
      }
      return changed ? next : prev
    })
  }, [deviceMap])

  // Roll a bounded per-entity sample window: on every device stream tick, plus a slow heartbeat so
  // a steady reading still yields a (flat) sparkline instead of a single point. Never fabricates a
  // value — it only records what the stream reports — and the caller renders nothing below two.
  useEffect(() => {
    const sample = () => {
      setHistories((prev) => {
        let changed = false
        const next: Record<string, number[]> = { ...prev }
        for (const device of Object.values(deviceMap)) {
          if (device.state === 'unavailable' || device.state === 'unknown') continue
          const num = parseNumericState(device.state)
          if (num === null) continue
          const existing = next[device.id] ?? []
          next[device.id] = [...existing, num].slice(-HISTORY_LENGTH)
          changed = true
        }
        return changed ? next : prev
      })
    }

    sample()
    const timer = window.setInterval(sample, HISTORY_SAMPLE_MS)
    return () => window.clearInterval(timer)
  }, [deviceMap])

  const activeDashboard = dashboards.find((d) => d.id === activeDashboardId) || null

  const backgroundConfig: BackgroundConfig | null =
    previewBackground ??
    (activeDashboard
      ? {
          image: activeDashboard.background_image,
          opacity: activeDashboard.background_opacity,
          blur: activeDashboard.background_blur,
          dim: activeDashboard.background_dim,
        }
      : null)

  const slot = useTopBarSlot()
  const setupBannerSlot = useSetupBannerSlot()

  const toastTimerRef = useRef<number | null>(null)

  const showToast = useCallback((message: string) => {
    if (toastTimerRef.current !== null) window.clearTimeout(toastTimerRef.current)
    setToastMessage(message)
    toastTimerRef.current = window.setTimeout(() => {
      toastTimerRef.current = null
      setToastMessage(null)
    }, TOAST_MS)
  }, [])

  useEffect(() => {
    return () => {
      if (toastTimerRef.current !== null) window.clearTimeout(toastTimerRef.current)
      if (backgroundTimerRef.current !== null) window.clearTimeout(backgroundTimerRef.current)
    }
  }, [])

  const persistBackground = useCallback(
    (config: BackgroundConfig) => {
      if (!activeDashboard) return
      setPreviewBackground(config)
      if (backgroundTimerRef.current !== null) window.clearTimeout(backgroundTimerRef.current)
      backgroundTimerRef.current = window.setTimeout(async () => {
        backgroundTimerRef.current = null
        try {
          const res = await apiFetch(`/api/dashboards/${activeDashboard.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              name: activeDashboard.name,
              order: activeDashboard.order,
              cols: activeDashboard.cols,
              rows: activeDashboard.rows,
              background_image: config.image,
              background_opacity: config.opacity,
              background_blur: config.blur,
              background_dim: config.dim,
            }),
          })
          if (res.ok) {
            await fetchDashboards()
          } else {
            showToast(await readApiError(res))
          }
        } catch (err) {
          console.error('Failed to save dashboard background:', err)
          showToast(NETWORK_ERROR_MESSAGE)
        }
      }, 350)
    },
    [activeDashboard, fetchDashboards, showToast]
  )

  const persistLayout = useCallback(
    async (rects: Rect[]): Promise<string | null> => {
      if (!activeDashboardId) return null
      try {
        const res = await apiFetch(`/api/dashboards/${activeDashboardId}/layout`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            positions: rects.map((r) => ({
              id: r.id,
              col: r.col,
              row: r.row,
              col_span: r.colSpan,
              row_span: r.rowSpan,
            })),
          }),
        })
        if (res.ok) {
          await fetchDashboards()
          return null
        }
        return await readApiError(res)
      } catch (err) {
        console.error('Failed to persist layout:', err)
        return NETWORK_ERROR_MESSAGE
      }
    },
    [activeDashboardId, fetchDashboards]
  )

  const handleRenameDashboard = async (name: string): Promise<boolean> => {
    if (!activeDashboardId) return false
    const res = await apiFetch(`/api/dashboards/${activeDashboardId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name,
        order: activeDashboard?.order || 0,
        cols: activeDashboard?.cols || 0,
        rows: activeDashboard?.rows || 0,
        background_image: activeDashboard?.background_image ?? '',
        background_opacity: activeDashboard?.background_opacity ?? DEFAULT_BACKGROUND_OPACITY,
        background_blur: activeDashboard?.background_blur ?? DEFAULT_BACKGROUND_BLUR,
        background_dim: activeDashboard?.background_dim ?? DEFAULT_BACKGROUND_DIM,
      }),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchDashboards()
    return true
  }

  const handleDeleteDashboard = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this dashboard?')) return

    try {
      const res = await apiFetch(`/api/dashboards/${id}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        if (isEdit) {
          navigate('/setup/dashboards')
          return
        }
        await fetchDashboards()
      } else {
        showToast(await readApiError(res))
      }
    } catch (err) {
      console.error('Failed to delete dashboard:', err)
      showToast(NETWORK_ERROR_MESSAGE)
    }
  }

  const handleCreateDashboard = async () => {
    try {
      const res = await apiFetch('/api/dashboards', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'New dashboard',
          order: dashboards.length,
          background_opacity: DEFAULT_BACKGROUND_OPACITY,
          background_blur: DEFAULT_BACKGROUND_BLUR,
          background_dim: DEFAULT_BACKGROUND_DIM,
        }),
      })
      if (res.ok) {
        const payload = await res.json()
        const createdId = payload?.data?.id as string | undefined
        await fetchDashboards()
        if (createdId) navigate(`/setup/dashboards/${createdId}`)
      } else {
        showToast(await readApiError(res))
      }
    } catch (err) {
      console.error('Failed to create dashboard:', err)
      showToast(NETWORK_ERROR_MESSAGE)
    }
  }

  const handleCreateWidget = async (input: NewWidgetInput) => {
    if (!activeDashboardId) return
    const res = await apiFetch(`/api/dashboards/${activeDashboardId}/widgets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchDashboards()
  }

  const handleUpdateWidgetContent = async (widgetId: string, input: WidgetContentInput) => {
    if (!activeDashboardId) return
    const res = await apiFetch(`/api/dashboards/${activeDashboardId}/widgets/${widgetId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchDashboards()
  }

  const handleDeleteWidget = async (widgetId: string) => {
    if (!activeDashboardId) return
    if (!window.confirm('Delete this widget?')) return

    try {
      const res = await apiFetch(`/api/dashboards/${activeDashboardId}/widgets/${widgetId}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        await fetchDashboards()
      } else {
        showToast(await readApiError(res))
      }
    } catch (err) {
      console.error('Failed to delete widget:', err)
      showToast(NETWORK_ERROR_MESSAGE)
    }
  }

  const renderWidgetBody = (widget: Widget): React.ReactNode => {
    if (widget.type === 'automation_list') {
      return (
        <AutomationListWidget widget={widget} automations={automations} onTriggerSuccess={() => fetchAutomations()} />
      )
    }

    if (widget.type === 'weather') {
      return <WeatherWidget widget={widget} />
    }

    const entityId = widget.config.entity_ids?.[0]
    const device = entityId ? deviceMap[entityId] : undefined
    const label =
      widget.title || (entityId ? widget.config.labels?.[entityId] : undefined) || device?.name || entityId || 'Widget'
    const unit = widget.config.unit || (device?.attributes?.unit_of_measurement as string | undefined) || ''
    const stale = isStaleDevice(device)
    const history = entityId ? histories[entityId] : undefined

    if (widget.type === 'sensor') {
      const live = parseNumericState(device?.state)
      const value = live !== null ? live : entityId ? lastKnownValues[entityId] ?? null : null
      switch (widget.config.display) {
        case 'number':
          return (
            <NumberWidget
              label={label}
              value={value ?? device?.state ?? null}
              unit={unit}
              min={widget.config.min}
              max={widget.config.max}
              history={history}
              stale={stale}
              dense={widget.row_span === 1}
            />
          )
        case 'bar':
          return (
            <BarWidget
              label={label}
              value={value}
              min={widget.config.min ?? 0}
              max={widget.config.max ?? 100}
              unit={unit}
              history={history}
              stale={stale}
              dense={widget.row_span === 1}
            />
          )
        case 'arc':
          return (
            <ArcWidget
              label={label}
              value={value}
              min={widget.config.min ?? 0}
              max={widget.config.max ?? 100}
              unit={unit}
              history={history}
              stale={stale}
            />
          )
        default:
          return null
      }
    }

    if (widget.type === 'actuator') {
      const isOn = device?.state === 'on'
      const isPending = entityId ? !!pendingDevices[entityId] : false
      return (
        <ToggleWidget
          label={label}
          on={isOn}
          pending={isPending}
          stale={stale}
          dense={widget.row_span === 1}
          onToggle={() => {
            if (entityId) toggleDevice(entityId)
          }}
        />
      )
    }

    return null
  }

  const renderGrid = () => {
    if (!activeDashboard) return null
    if (isEdit) {
      return (
        <WidgetGrid
          dashboard={activeDashboard}
          isEditMode={isEditMode}
          renderWidgetBody={renderWidgetBody}
          onEditWidget={setEditingWidget}
          onDeleteWidget={handleDeleteWidget}
          onPersistLayout={persistLayout}
          onMessage={showToast}
        />
      )
    }
    return (
      <WidgetGrid
        dashboard={activeDashboard}
        isEditMode={false}
        renderWidgetBody={renderWidgetBody}
        onEditWidget={() => {}}
        onDeleteWidget={() => {}}
        onPersistLayout={async () => null}
        onMessage={showToast}
      />
    )
  }

  const showSelector = !isEdit && dashboards.length > 1

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-slate-950">
      {showSelector &&
        slot &&
        createPortal(
          <div className="mx-auto flex h-14 min-w-0 max-w-full items-center gap-2 overflow-x-auto [scrollbar-width:none]">
            <DashboardSelector
              dashboards={dashboards}
              activeDashboardId={activeDashboardId}
              onSelectDashboard={setActiveDashboardId}
            />
          </div>,
          slot,
        )}

      {isEdit &&
        setupBannerSlot &&
        createPortal(
          <DashboardSelector
            dashboards={dashboards}
            activeDashboardId={activeDashboardId}
            showSingle
            onSelectDashboard={(id) => navigate(`/setup/dashboards/${id}`)}
            onCreateDashboard={() => void handleCreateDashboard()}
          />,
          setupBannerSlot,
        )}

      <main className="flex-1 min-h-0 relative overflow-hidden p-3">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(1200px_500px_at_50%_-10%,rgba(109,118,232,0.06),transparent_70%)]"
        />
        {backgroundConfig?.image && (
          <>
            <div
              aria-hidden="true"
              className="pointer-events-none absolute inset-0 bg-cover bg-center"
              style={{
                backgroundImage: `url(${backgroundConfig.image})`,
                filter: `blur(${backgroundConfig.blur}px)`,
                opacity: backgroundConfig.opacity / 100,
                transform: 'scale(1.1)',
              }}
            />
            <div
              aria-hidden="true"
              className="pointer-events-none absolute inset-0 bg-slate-950"
              style={{ opacity: backgroundConfig.dim / 100 }}
            />
          </>
        )}
        {toastMessage && <Toast message={toastMessage} />}
        <div className="relative z-10 h-full flex flex-col gap-3">
          {isEdit && (
            <div className="shrink-0 flex items-center">
              <DashboardHeader
                dashboard={activeDashboard}
                isEditMode={isEditMode}
                onRename={handleRenameDashboard}
                onDelete={() => {
                  if (activeDashboardId) void handleDeleteDashboard(activeDashboardId)
                }}
                onAddWidget={() => setIsAddWidgetOpen(true)}
                onToggleEditMode={() => setIsEditMode((v) => !v)}
                onToggleBackground={() => setIsBackgroundOpen((v) => !v)}
              />
            </div>
          )}

          <div className="flex-1 min-h-0">
            {loading ? (
              <div className="h-full flex items-center justify-center text-slate-400">
                <RefreshCw className="w-6 h-6 animate-spin mr-2" />
                <span>Loading dashboards...</span>
              </div>
            ) : dashboards.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-center p-8">
                <div className="p-4 rounded-xl bg-[#6d76e8]/10 text-[#8b93ee] border border-[#6d76e8]/20 mb-4">
                  <LayoutDashboard className="w-10 h-10" />
                </div>
                <h2 className="text-lg font-bold text-white mb-2">No Dashboard</h2>
                <p className="text-xs text-slate-400 max-w-md mb-6 leading-relaxed">
                  Dashboards group your automation and sensor widgets for quick
                  touch-based control.
                </p>
                <button
                  type="button"
                  onClick={() => navigate('/setup/dashboards')}
                  className="flex items-center space-x-2 px-5 py-2.5 bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 text-white text-xs font-bold rounded-lg transition-all"
                >
                  <Plus className="w-4 h-4" />
                  <span>Create a dashboard</span>
                </button>
              </div>
            ) : !activeDashboard ? (
              <div className="h-full flex flex-col items-center justify-center text-center p-8">
                <h2 className="text-lg font-bold text-white mb-2">Dashboard not found</h2>
                <button
                  type="button"
                  onClick={() => navigate('/setup/dashboards')}
                  className="mt-2 flex items-center space-x-2 px-5 py-2.5 bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 text-white text-xs font-bold rounded-lg transition-all"
                >
                  <span>Back to dashboards</span>
                </button>
              </div>
            ) : activeDashboard.widgets.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-center p-8">
                <div className="p-3 rounded-xl bg-slate-900/55 border border-slate-800/80 text-slate-400 mb-3">
                  <Settings className="w-8 h-8" />
                </div>
                <h3 className="text-base font-bold text-white mb-1">
                  Dashboard "{activeDashboard.name}" is empty
                </h3>
                <p className="text-xs text-slate-400 max-w-sm mb-5">
                  Add your first widget to display your sensors, control your devices or
                  trigger your favorite Home Assistant automations.
                </p>
                {isEdit && (
                  <button
                    type="button"
                    onClick={() => setIsAddWidgetOpen(true)}
                    className="flex items-center space-x-1.5 px-4 py-2 bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 text-white text-xs font-bold rounded-lg transition-all"
                  >
                    <Plus className="w-4 h-4" />
                    <span>Add a Widget</span>
                  </button>
                )}
              </div>
            ) : (
              renderGrid()
            )}
          </div>
        </div>
      </main>

      {isEdit && (
        <AddWidgetModal
          isOpen={isAddWidgetOpen || !!editingWidget}
          dashboardCols={activeDashboard?.cols ?? 12}
          dashboardRows={activeDashboard?.rows ?? 8}
          occupiedRects={activeDashboard ? widgetsToRects(activeDashboard.widgets) : []}
          nextOrder={activeDashboard?.widgets.length ?? 0}
          automations={automations}
          devices={devices}
          editingWidget={editingWidget}
          onClose={() => {
            setIsAddWidgetOpen(false)
            setEditingWidget(null)
          }}
          onCreate={handleCreateWidget}
          onUpdate={handleUpdateWidgetContent}
        />
      )}

      {isEdit && isBackgroundOpen && activeDashboard && backgroundConfig && (
        <DashboardBackgroundPanel
          value={backgroundConfig}
          onChange={persistBackground}
          onClose={() => {
            setIsBackgroundOpen(false)
            setPreviewBackground(null)
          }}
        />
      )}
    </div>
  )
}
