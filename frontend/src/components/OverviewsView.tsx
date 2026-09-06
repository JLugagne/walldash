import React, { useCallback, useEffect, useRef, useState } from 'react'
import { LayoutDashboard, Plus, RefreshCw, Settings } from 'lucide-react'
import type { Device, Widget } from '../types'
import { AutomationListWidget } from './overview/widgets/AutomationListWidget'
import { AddWidgetModal, type NewWidgetInput, type WidgetContentInput } from './AddWidgetModal'
import { apiFetch, readApiError } from '../api'
import { useRealtimeDevices } from '../hooks/useRealtimeDevices'
import { NumberWidget } from './overview/widgets/NumberWidget'
import { BarWidget } from './overview/widgets/BarWidget'
import { ArcWidget } from './overview/widgets/ArcWidget'
import { ToggleWidget } from './overview/widgets/ToggleWidget'
import { OverviewHeader } from './overview/OverviewHeader'
import { WidgetGrid } from './overview/WidgetGrid'
import { Toast } from './overview/Toast'
import { useOverviewData } from './overview/useOverviewData'
import { widgetsToRects } from './overview/format'
import type { Rect } from './overview/grid'

interface OverviewsViewProps {
  initialIsAdmin?: boolean
  viewModeMenu?: React.ReactNode
}

const STALE_MS = 30 * 60 * 1000
const TOAST_MS = 3000
const NETWORK_ERROR_MESSAGE = 'Connexion au serveur impossible, modification non enregistrée.'

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

export const OverviewsView: React.FC<OverviewsViewProps> = ({
  initialIsAdmin = false,
  viewModeMenu,
}) => {
  const [activeOverviewId, setActiveOverviewId] = useState<string | null>(null)
  const { overviews, automations, loading, fetchOverviews, fetchAutomations } =
    useOverviewData(setActiveOverviewId)
  const [isAdmin, setIsAdmin] = useState(initialIsAdmin)
  const [isEditMode, setIsEditMode] = useState(false)
  const [isAddWidgetOpen, setIsAddWidgetOpen] = useState(false)
  const [editingWidget, setEditingWidget] = useState<Widget | null>(null)
  const [toastMessage, setToastMessage] = useState<string | null>(null)

  const [lastKnownValues, setLastKnownValues] = useState<Record<string, number>>({})

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

  const activeOverview = overviews.find((o) => o.id === activeOverviewId) || null

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
    }
  }, [])

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
        body: JSON.stringify({ name: 'Vue Synthétique', order: 0 }),
      })
      if (res.ok) {
        const payload = await res.json()
        if (payload?.data?.id && automations.length > 0) {
          const widgetRes = await apiFetch(`/api/overviews/${payload.data.id}/widgets`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              type: 'automation_list',
              title: 'Automatisations Rapides',
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
      body: JSON.stringify({ name, order: overviews.length }),
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
      body: JSON.stringify({ name, order: activeOverview?.order || 0 }),
    })
    if (!res.ok) {
      throw new Error(await readApiError(res))
    }
    await fetchOverviews()
    return true
  }

  const handleDeleteOverview = async (id: string) => {
    if (!window.confirm('Êtes-vous sûr de vouloir supprimer ce tableau de bord ?')) return

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
    if (!window.confirm('Supprimer ce widget ?')) return

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

    const entityId = widget.config.entity_ids?.[0]
    const device = entityId ? deviceMap[entityId] : undefined
    const label =
      widget.title || (entityId ? widget.config.labels?.[entityId] : undefined) || device?.name || entityId || 'Widget'
    const unit = widget.config.unit || (device?.attributes?.unit_of_measurement as string | undefined) || ''
    const stale = isStaleDevice(device)

    if (widget.type === 'sensor') {
      const live = parseNumericState(device?.state)
      const value = live !== null ? live : entityId ? lastKnownValues[entityId] ?? null : null
      switch (widget.config.display) {
        case 'number':
          return <NumberWidget label={label} value={value ?? device?.state ?? null} unit={unit} stale={stale} dense={widget.row_span === 1} />
        case 'bar':
          return (
            <BarWidget
              label={label}
              value={value}
              min={widget.config.min ?? 0}
              max={widget.config.max ?? 100}
              unit={unit}
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

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-slate-950">
      <OverviewHeader
        overviews={overviews}
        activeOverviewId={activeOverviewId}
        isAdmin={isAdmin}
        isEditMode={isEditMode}
        viewModeMenu={viewModeMenu}
        onSelectOverview={setActiveOverviewId}
        onCreateOverview={handleCreateOverview}
        onRenameOverview={handleRenameOverview}
        onDeleteOverview={handleDeleteOverview}
        onToggleEditMode={() => setIsEditMode((v) => !v)}
        onAddWidget={() => setIsAddWidgetOpen(true)}
        onToggleAdmin={() => setIsAdmin((v) => !v)}
      />

      <main className="flex-1 min-h-0 relative overflow-hidden p-3">
        {toastMessage && <Toast message={toastMessage} />}
        {loading ? (
          <div className="h-full flex items-center justify-center text-slate-400">
            <RefreshCw className="w-6 h-6 animate-spin mr-2" />
            <span>Chargement des Overviews...</span>
          </div>
        ) : overviews.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center p-8">
            <div className="p-4 rounded-2xl bg-indigo-600/10 text-indigo-400 border border-indigo-500/20 mb-4 shadow-xl">
              <LayoutDashboard className="w-10 h-10" />
            </div>
            <h2 className="text-lg font-bold text-white mb-2">Aucun Overview Dashboard</h2>
            <p className="text-xs text-slate-400 max-w-md mb-6 leading-relaxed">
              Les Overview Dashboards regroupent vos widgets d'automatisations et capteurs
              pour un pilotage tactile rapide.
            </p>
            <button
              type="button"
              onClick={handleCreateDefaultOverview}
              className="flex items-center space-x-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white text-xs font-bold rounded-xl shadow-xl shadow-indigo-600/30 transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Créer mon premier Overview</span>
            </button>
          </div>
        ) : !activeOverview ? (
          <div className="text-center py-12 text-slate-500 text-sm">
            Sélectionnez un Overview ci-dessus.
          </div>
        ) : activeOverview.widgets.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center p-8">
            <div className="p-3 rounded-xl bg-slate-800 text-slate-400 mb-3">
              <Settings className="w-8 h-8" />
            </div>
            <h3 className="text-base font-bold text-white mb-1">
              Tableau "{activeOverview.name}" vide
            </h3>
            <p className="text-xs text-slate-400 max-w-sm mb-5">
              Ajoutez votre premier widget pour afficher vos capteurs, piloter vos appareils ou
              déclencher vos automatisations Home Assistant préférées.
            </p>
            {isAdmin && (
              <button
                type="button"
                onClick={() => setIsAddWidgetOpen(true)}
                className="flex items-center space-x-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white text-xs font-bold rounded-xl shadow-lg shadow-indigo-600/30 transition-all"
              >
                <Plus className="w-4 h-4" />
                <span>Ajouter un Widget</span>
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
    </div>
  )
}
