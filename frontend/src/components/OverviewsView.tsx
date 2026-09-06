import React, { useState, useEffect, useCallback } from 'react'
import {
  LayoutDashboard,
  Plus,
  Trash2,
  Edit2,
  Check,
  X,
  Settings,
  RefreshCw,
  SlidersHorizontal,
} from 'lucide-react'
import type { OverviewDashboard, Automation } from '../types'
import { AutomationListWidget } from './AutomationListWidget'
import { AddWidgetModal } from './AddWidgetModal'

interface OverviewsViewProps {
  initialIsAdmin?: boolean
}

export const OverviewsView: React.FC<OverviewsViewProps> = ({ initialIsAdmin = false }) => {
  const [overviews, setOverviews] = useState<OverviewDashboard[]>([])
  const [activeOverviewId, setActiveOverviewId] = useState<string | null>(null)
  const [automations, setAutomations] = useState<Automation[]>([])
  const [loading, setLoading] = useState(true)
  const [isAdmin, setIsAdmin] = useState(initialIsAdmin)
  const [isAddWidgetOpen, setIsAddWidgetOpen] = useState(false)

  // Creation / Editing states
  const [isCreatingOverview, setIsCreatingOverview] = useState(false)
  const [newOverviewName, setNewOverviewName] = useState('')
  const [isRenamingOverview, setIsRenamingOverview] = useState(false)
  const [renameValue, setRenameValue] = useState('')

  const fetchOverviews = useCallback(async () => {
    try {
      const res = await fetch('/api/overviews')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setOverviews(payload.data)
          setActiveOverviewId((prev) => {
            if (prev && payload.data.some((o: OverviewDashboard) => o.id === prev)) {
              return prev
            }
            return payload.data.length > 0 ? payload.data[0].id : null
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch overviews:', err)
    }
  }, [])

  const fetchAutomations = useCallback(async () => {
    try {
      const res = await fetch('/api/automations')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setAutomations(payload.data)
        }
      }
    } catch (err) {
      console.error('Failed to fetch automations:', err)
    }
  }, [])

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchOverviews(), fetchAutomations()]).finally(() => setLoading(false))

    // Refresh automations state every 5 seconds for live status
    const interval = setInterval(fetchAutomations, 5000)
    return () => clearInterval(interval)
  }, [fetchOverviews, fetchAutomations])

  const activeOverview = overviews.find((o) => o.id === activeOverviewId) || null

  // Create an initial default overview if none exists
  const handleCreateDefaultOverview = async () => {
    try {
      const res = await fetch('/api/overviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'Vue Synthétique', order: 0 }),
      })
      if (res.ok) {
        const payload = await res.json()
        if (payload?.data?.id) {
          // Also add a default automation list widget
          const widgetRes = await fetch(`/api/overviews/${payload.data.id}/widgets`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              type: 'automation_list',
              title: 'Automatisations Rapides',
              order: 0,
              config: {
                entity_ids: automations.slice(0, 3).map((a) => a.id),
              },
            }),
          })
          if (widgetRes.ok) {
            await fetchOverviews()
            return
          }
        }
        await fetchOverviews()
      }
    } catch (err) {
      console.error('Failed to create default overview:', err)
    }
  }

  const handleCreateOverview = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newOverviewName.trim()) return

    try {
      const res = await fetch('/api/overviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newOverviewName.trim(), order: overviews.length }),
      })
      if (res.ok) {
        setNewOverviewName('')
        setIsCreatingOverview(false)
        await fetchOverviews()
      }
    } catch (err) {
      console.error('Failed to create overview:', err)
    }
  }

  const handleRenameOverview = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!activeOverviewId || !renameValue.trim()) return

    try {
      const res = await fetch(`/api/overviews/${activeOverviewId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: renameValue.trim(), order: activeOverview?.order || 0 }),
      })
      if (res.ok) {
        setIsRenamingOverview(false)
        await fetchOverviews()
      }
    } catch (err) {
      console.error('Failed to rename overview:', err)
    }
  }

  const handleDeleteOverview = async (id: string) => {
    if (!window.confirm('Êtes-vous sûr de vouloir supprimer ce tableau de bord ?')) return

    try {
      const res = await fetch(`/api/overviews/${id}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        await fetchOverviews()
      }
    } catch (err) {
      console.error('Failed to delete overview:', err)
    }
  }

  const handleAddWidget = async (title: string, selectedEntityIds: string[]) => {
    if (!activeOverviewId) return

    const res = await fetch(`/api/overviews/${activeOverviewId}/widgets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'automation_list',
        title,
        order: (activeOverview?.widgets?.length || 0) + 1,
        config: {
          entity_ids: selectedEntityIds,
        },
      }),
    })

    if (res.ok) {
      await fetchOverviews()
    }
  }

  const handleDeleteWidget = async (widgetId: string) => {
    if (!activeOverviewId) return
    if (!window.confirm('Supprimer ce widget ?')) return

    try {
      const res = await fetch(`/api/overviews/${activeOverviewId}/widgets/${widgetId}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        await fetchOverviews()
      }
    } catch (err) {
      console.error('Failed to delete widget:', err)
    }
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center text-slate-400">
        <RefreshCw className="w-6 h-6 animate-spin mr-2" />
        <span>Chargement des Overviews...</span>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-slate-950">
      {/* Sub-header Navigation & Controls */}
      <div className="border-b border-slate-800/80 bg-slate-900/60 backdrop-blur px-6 py-3 flex flex-wrap items-center justify-between gap-4">
        {/* Overview Tabs */}
        <div className="flex items-center space-x-2 overflow-x-auto py-1">
          {overviews.map((ov) => (
            <button
              key={ov.id}
              type="button"
              onClick={() => {
                setActiveOverviewId(ov.id)
                setIsRenamingOverview(false)
              }}
              className={`px-4 py-2 rounded-xl text-xs font-bold transition-all whitespace-nowrap active:scale-95 ${
                activeOverviewId === ov.id
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30'
                  : 'bg-slate-800/60 text-slate-300 hover:bg-slate-800 hover:text-white'
              }`}
            >
              {ov.name}
            </button>
          ))}

          {/* Add Overview Button in Admin Mode */}
          {isAdmin && !isCreatingOverview && (
            <button
              type="button"
              onClick={() => setIsCreatingOverview(true)}
              className="flex items-center space-x-1 px-3 py-2 rounded-xl text-xs font-semibold bg-slate-800/40 hover:bg-slate-800 text-slate-400 hover:text-white border border-dashed border-slate-700 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Nouvel Overview</span>
            </button>
          )}

          {/* Inline Create Overview Input */}
          {isAdmin && isCreatingOverview && (
            <form onSubmit={handleCreateOverview} className="flex items-center space-x-1.5">
              <input
                type="text"
                value={newOverviewName}
                onChange={(e) => setNewOverviewName(e.target.value)}
                placeholder="Nom de l'overview..."
                autoFocus
                className="bg-slate-800 border border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none"
              />
              <button
                type="submit"
                className="p-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg"
              >
                <Check className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                onClick={() => setIsCreatingOverview(false)}
                className="p-1.5 bg-slate-800 hover:bg-slate-700 text-slate-400 rounded-lg"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </form>
          )}
        </div>

        {/* Action Controls (Admin Toggle + Dashboard Actions) */}
        <div className="flex items-center space-x-3">
          {activeOverview && isAdmin && !isRenamingOverview && (
            <div className="flex items-center space-x-1.5 border-r border-slate-800 pr-3 mr-1">
              <button
                type="button"
                onClick={() => {
                  setRenameValue(activeOverview.name)
                  setIsRenamingOverview(true)
                }}
                className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
                title="Renommer l'overview"
              >
                <Edit2 className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                onClick={() => handleDeleteOverview(activeOverview.id)}
                className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                title="Supprimer l'overview"
              >
                <Trash2 className="w-3.5 h-3.5" />
              </button>
            </div>
          )}

          {isRenamingOverview && (
            <form onSubmit={handleRenameOverview} className="flex items-center space-x-1.5">
              <input
                type="text"
                value={renameValue}
                onChange={(e) => setRenameValue(e.target.value)}
                autoFocus
                className="bg-slate-800 border border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none"
              />
              <button
                type="submit"
                className="p-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg"
              >
                <Check className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                onClick={() => setIsRenamingOverview(false)}
                className="p-1.5 bg-slate-800 hover:bg-slate-700 text-slate-400 rounded-lg"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </form>
          )}

          {activeOverview && isAdmin && (
            <button
              type="button"
              onClick={() => setIsAddWidgetOpen(true)}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white rounded-xl text-xs font-bold shadow-lg shadow-indigo-600/30 transition-all"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Ajouter un Widget</span>
            </button>
          )}

          {/* Admin / User Mode Toggle Button */}
          <button
            type="button"
            onClick={() => setIsAdmin(!isAdmin)}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold border transition-all ${
              isAdmin
                ? 'bg-amber-500/20 text-amber-300 border-amber-500/40 shadow-sm shadow-amber-500/20'
                : 'bg-slate-800/60 text-slate-400 border-slate-700 hover:text-slate-200'
            }`}
          >
            <SlidersHorizontal className="w-3.5 h-3.5" />
            <span>{isAdmin ? 'Mode Admin Actif' : 'Mode Utilisateur'}</span>
          </button>
        </div>
      </div>

      {/* Main Grid View */}
      <main className="flex-1 p-6 overflow-y-auto">
        {overviews.length === 0 ? (
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
              Ajoutez votre premier widget pour afficher et déclencher vos automatisations
              Home Assistant préférées.
            </p>
            <button
              type="button"
              onClick={() => setIsAddWidgetOpen(true)}
              className="flex items-center space-x-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 active:scale-95 text-white text-xs font-bold rounded-xl shadow-lg shadow-indigo-600/30 transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Ajouter un Widget Automatisations</span>
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {activeOverview.widgets.map((widget) => (
              <AutomationListWidget
                key={widget.id}
                widget={widget}
                automations={automations}
                isAdmin={isAdmin}
                onDeleteWidget={handleDeleteWidget}
                onTriggerSuccess={() => {
                  fetchAutomations()
                }}
              />
            ))}
          </div>
        )}
      </main>

      {/* Add Widget Modal */}
      <AddWidgetModal
        isOpen={isAddWidgetOpen}
        automations={automations}
        onClose={() => setIsAddWidgetOpen(false)}
        onAddWidget={handleAddWidget}
      />
    </div>
  )
}
