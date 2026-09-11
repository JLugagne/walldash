import React, { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { LayoutDashboard, Plus, RefreshCw, Settings } from 'lucide-react'
import type { Device, Widget } from '../types'
import { AutomationListWidget } from './overview/widgets/AutomationListWidget'
import { WeatherWidget } from './overview/widgets/WeatherWidget'
import { AddWidgetModal, type NewWidgetInput, type WidgetContentInput } from './AddWidgetModal'
import { apiFetch, readApiError } from '../api'
import { useRealtimeDevices } from '../hooks/useRealtimeDevices'
import { NumberWidget } from './overview/widgets/NumberWidget'
import { BarWidget } from './overview/widgets/BarWidget'
import { ArcWidget } from './overview/widgets/ArcWidget'
import { ToggleWidget } from './overview/widgets/ToggleWidget'
import { OverviewHeader } from './overview/OverviewHeader'
import { useTopBarSlot } from './TopBarSlot'
import { WidgetGrid } from './overview/WidgetGrid'
import { Toast } from './overview/Toast'
import { useOverviewData } from './overview/useOverviewData'
import { OverviewBackgroundPanel, type BackgroundConfig } from './overview/OverviewBackgroundPanel'
import {
  DEFAULT_BACKGROUND_BLUR,
  DEFAULT_BACKGROUND_DIM,
  DEFAULT_BACKGROUND_OPACITY,
} from './overview/backgroundSamples'
import { widgetsToRects } from './overview/format'
import type { Rect } from './overview/grid'

interface OverviewsViewProps {
  initialIsAdmin?: boolean
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

export const OverviewsView: React.FC<OverviewsViewProps> = ({ initialIsAdmin = false }) => {
  const [activeOverviewId, setActiveOverviewId] = useState<string | null>(null)
  const { overviews, automations, loading, fetchOverviews, fetchAutomations } =
    useOverviewData(setActiveOverviewId)
  const [isAdmin, setIsAdmin] = useState(initialIsAdmin)
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
    if (!isAdmin) setIsEditMode(false)
  }, [isAdmin])

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

  const activeOverview = overviews.find((o) => o.id === activeOverviewId) || null

  const backgroundConfig: BackgroundConfig | null =
    previewBackground ??
    (activeOverview
      ? {
          image: activeOverview.background_image,
          opacity: activeOverview.background_opacity,
          blur: activeOverview.background_blur,
          dim: activeOverview.background_dim,
        }
      : null)

  useEffect(() => {
    setPreviewBackground(null)
    setIsBackgroundOpen(false)
    if (activeOverviewId && window.sessionStorage.getItem('walldash:open-background') === '1') {
      window.sessionStorage.removeItem('walldash:open-background')
      setIsBackgroundOpen(true)
    }
  }, [activeOverviewId])

  const slot = useTopBarSlot()

  useEffect(() => {
    const handler = () => setIsBackgroundOpen((v) => !v)
    window.addEventListener('walldash:open-background', handler)
    return () => window.removeEventListener('walldash:open-background', handler)
  }, [])

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
      if (!activeOverview) return
      setPreviewBackground(config)
      if (backgroundTimerRef.current !== null) window.clearTimeout(backgroundTimerRef.current)
      backgroundTimerRef.current = window.setTimeout(async () => {
        backgroundTimerRef.current = null
        try {
          const res = await apiFetch(`/api/overviews/${activeOverview.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              name: activeOverview.name,
              order: activeOverview.order,
              cols: activeOverview.cols,
              rows: activeOverview.rows,
              background_image: config.image,
              background_opacity: config.opacity,
              background_blur: config.blur,
              background_dim: config.dim,
            }),
          })
          if (res.ok) {
            await fetchOverviews()
          } else {
            showToast(await readApiError(res))
          }
        } catch (err) {
          console.error('Failed to save dashboard background:', err)
          showToast(NETWORK_ERROR_MESSAGE)
        }
      }, 350)
    },
    [activeOverview, fetchOverviews, showToast]
  )

  const persistLayout = useCallback(
    async (rects: Rect[]): Promise<string | null> => {
      if (!activeOverviewId) return null
      try {
        const res = await apiFetch(`/api/overviews/${activeOverviewId}/layout`, {
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
          await fetchOverviews()
          return null
        }
        return await readApiError(res)
      } catch (err) {
        console.error('Failed to persist layout:', err)
        return NETWORK_ERROR_MESSAGE
      }
    },
    [activeOverviewId, fetchOverviews]
  )

  const handleCreateDefaultOverview = async () => {
    try {
      const res = await apiFetch('/api/overviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'Overview',
          order: 0,
          background_opacity: DEFAULT_BACKGROUND_OPACITY,
          background_blur: DEFAULT_BACKGROUND_BLUR,
          background_dim: DEFAULT_BACKGROUND_DIM,
        }),
      })
      if (res.ok) {
        const payload = await res.json()
        if (payload?.data?.id && automations.length > 0) {
          const widgetRes = await apiFetch(`/api/overviews/${payload.data.id}/widgets`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              type: 'automation_list',
              title: 'Quick Automations',
              order: 0,
              config: {
                entity_ids: automations.slice(0, 3).map((a) => a.id),
                display: 'list',
              },
              col: 0,
              row: 0,
              col_span: 2,
              row_span: 2,
            }),
          })
          if (!widgetRes.ok) {
            showToast(await readApiError(widgetRes))
          }
        }
        await fetchOverviews()
      } else {
        showToast(await readApiError(res))
      }
    } catch (err) {
      console.error('Failed to create default overview:', err)
      showToast(NETWORK_ERROR_MESSAGE)
    }
  }

  const handleCreateOverview = async (name: string): Promise<boolean> => {
    const res = await apiFetch('/api/overviews', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name,
        order: overviews.length,
        background_opacity: DEFAULT_BACKGROUND_OPACITY,
        background_blur: DEFAULT_BACKGROUND_BLUR,
        background_dim: DEFAULT_BACKGROUND_DIM,
      }),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchOverviews()
    return true
  }

  const handleRenameOverview = async (name: string): Promise<boolean> => {
    if (!activeOverviewId) return false
    const res = await apiFetch(`/api/overviews/${activeOverviewId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name,
        order: activeOverview?.order || 0,
        cols: activeOverview?.cols || 0,
        rows: activeOverview?.rows || 0,
        background_image: activeOverview?.background_image ?? '',
        background_opacity: activeOverview?.background_opacity ?? DEFAULT_BACKGROUND_OPACITY,
        background_blur: activeOverview?.background_blur ?? DEFAULT_BACKGROUND_BLUR,
        background_dim: activeOverview?.background_dim ?? DEFAULT_BACKGROUND_DIM,
      }),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchOverviews()
    return true
  }

  const handleDeleteOverview = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this dashboard?')) return

    try {
      const res = await apiFetch(`/api/overviews/${id}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        await fetchOverviews()
      } else {
        showToast(await readApiError(res))
      }
    } catch (err) {
      console.error('Failed to delete overview:', err)
      showToast(NETWORK_ERROR_MESSAGE)
    }
  }

  const handleCreateWidget = async (input: NewWidgetInput) => {
    if (!activeOverviewId) return
    const res = await apiFetch(`/api/overviews/${activeOverviewId}/widgets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchOverviews()
  }

  const handleUpdateWidgetContent = async (widgetId: string, input: WidgetContentInput) => {
    if (!activeOverviewId) return
    const res = await apiFetch(`/api/overviews/${activeOverviewId}/widgets/${widgetId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchOverviews()
  }

  const handleDeleteWidget = async (widgetId: string) => {
    if (!activeOverviewId) return
    if (!window.confirm('Delete this widget?')) return

    try {
      const res = await apiFetch(`/api/overviews/${activeOverviewId}/widgets/${widgetId}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        await fetchOverviews()
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

  const overviewHeader = (
    <OverviewHeader
      overviews={overviews}
      activeOverviewId={activeOverviewId}
      isAdmin={isAdmin}
      isEditMode={isEditMode}
      onSelectOverview={setActiveOverviewId}
      onCreateOverview={handleCreateOverview}
      onRenameOverview={handleRenameOverview}
      onDeleteOverview={handleDeleteOverview}
      onToggleEditMode={() => setIsEditMode((v) => !v)}
      onAddWidget={() => setIsAddWidgetOpen(true)}
      onToggleBackground={() => setIsBackgroundOpen((v) => !v)}
      onToggleAdmin={() => setIsAdmin((v) => !v)}
    />
  )

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-slate-950">
      {slot ? (
        createPortal(
          <div className="ml-auto flex h-14 items-center select-none">{overviewHeader}</div>,
          slot,
        )
      ) : (
        <div className="relative z-20 h-14 shrink-0 select-none flex items-center justify-end px-4">
          {overviewHeader}
        </div>
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
        <div className="relative z-10 h-full">
        {loading ? (
          <div className="h-full flex items-center justify-center text-slate-400">
            <RefreshCw className="w-6 h-6 animate-spin mr-2" />
            <span>Loading Overviews...</span>
          </div>
        ) : overviews.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center p-8">
            <div className="p-4 rounded-xl bg-[#6d76e8]/10 text-[#8b93ee] border border-[#6d76e8]/20 mb-4">
              <LayoutDashboard className="w-10 h-10" />
            </div>
            <h2 className="text-lg font-bold text-white mb-2">No Overview Dashboard</h2>
            <p className="text-xs text-slate-400 max-w-md mb-6 leading-relaxed">
              Overview Dashboards group your automation and sensor widgets
              for quick touch-based control.
            </p>
            <button
              type="button"
              onClick={handleCreateDefaultOverview}
              className="flex items-center space-x-2 px-5 py-2.5 bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 text-white text-xs font-bold rounded-lg transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Create my first Overview</span>
            </button>
          </div>
        ) : !activeOverview ? (
          <div className="text-center py-12 text-slate-500 text-sm">
            Select an Overview above.
          </div>
        ) : activeOverview.widgets.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center p-8">
            <div className="p-3 rounded-xl bg-slate-900/55 border border-slate-800/80 text-slate-400 mb-3">
              <Settings className="w-8 h-8" />
            </div>
            <h3 className="text-base font-bold text-white mb-1">
              Dashboard "{activeOverview.name}" is empty
            </h3>
            <p className="text-xs text-slate-400 max-w-sm mb-5">
              Add your first widget to display your sensors, control your devices or
              trigger your favorite Home Assistant automations.
            </p>
            {isAdmin && (
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
          <WidgetGrid
            overview={activeOverview}
            isEditMode={isEditMode}
            renderWidgetBody={renderWidgetBody}
            onEditWidget={setEditingWidget}
            onDeleteWidget={handleDeleteWidget}
            onPersistLayout={persistLayout}
            onMessage={showToast}
          />
        )}
        </div>
      </main>

      <AddWidgetModal
        isOpen={isAddWidgetOpen || !!editingWidget}
        overviewCols={activeOverview?.cols ?? 12}
        overviewRows={activeOverview?.rows ?? 8}
        occupiedRects={activeOverview ? widgetsToRects(activeOverview.widgets) : []}
        nextOrder={activeOverview?.widgets.length ?? 0}
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

      {isBackgroundOpen && activeOverview && backgroundConfig && (
        <OverviewBackgroundPanel
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
