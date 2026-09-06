import React, { useState } from 'react'
import { Zap, Play, CheckCircle2, Trash2, Clock, Activity } from 'lucide-react'
import type { Widget, Automation } from '../types'
import { apiFetch } from '../api'

interface AutomationListWidgetProps {
  widget: Widget
  automations: Automation[]
  isAdmin: boolean
  onDeleteWidget?: (widgetId: string) => void
  onTriggerSuccess?: (automationId: string) => void
}

function formatRelativeTime(isoString: string | null): string {
  if (!isoString) return 'Jamais exécutée'
  const date = new Date(isoString)
  if (isNaN(date.getTime())) return 'Jamais exécutée'
  const now = new Date()
  const diffSec = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diffSec < 10) return "À l'instant"
  if (diffSec < 60) return `Il y a ${diffSec} s`
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `Il y a ${diffMin} min`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `Il y a ${diffHours} h`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `Il y a ${diffDays} j`
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const AutomationListWidget: React.FC<AutomationListWidgetProps> = ({
  widget,
  automations,
  isAdmin,
  onDeleteWidget,
  onTriggerSuccess,
}) => {
  const [triggeringId, setTriggeringId] = useState<string | null>(null)
  const [triggeredIds, setTriggeredIds] = useState<Set<string>>(new Set())

  // Filter automations configured in widget (or show all if no explicit IDs specified)
  const configuredEntityIds = widget.config?.entity_ids || []
  const displayedAutomations =
    configuredEntityIds.length > 0
      ? automations.filter((a) => configuredEntityIds.includes(a.id))
      : automations

  const handleTrigger = async (automationId: string) => {
    if (triggeringId) return
    setTriggeringId(automationId)

    try {
      const res = await apiFetch(`/api/automations/${encodeURIComponent(automationId)}/trigger`, {
        method: 'POST',
      })
      if (res.ok) {
        setTriggeredIds((prev) => new Set(prev).add(automationId))
        setTimeout(() => {
          setTriggeredIds((prev) => {
            const next = new Set(prev)
            next.delete(automationId)
            return next
          })
        }, 2000)
        onTriggerSuccess?.(automationId)
      }
    } catch (err) {
      console.error('Failed to trigger automation:', err)
    } finally {
      setTriggeringId(null)
    }
  }

  return (
    <div className="bg-slate-900/80 rounded-2xl border border-slate-800 shadow-xl overflow-hidden flex flex-col">
      {/* Widget Header */}
      <div className="px-5 py-4 border-b border-slate-800/80 bg-slate-900/90 flex items-center justify-between">
        <div className="flex items-center space-x-2.5">
          <div className="p-1.5 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Zap className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-white tracking-wide">{widget.title}</h3>
            <p className="text-[11px] text-slate-400">
              {displayedAutomations.length} automatisation{displayedAutomations.length > 1 ? 's' : ''}
            </p>
          </div>
        </div>

        {isAdmin && onDeleteWidget && (
          <button
            type="button"
            onClick={() => onDeleteWidget(widget.id)}
            className="p-1.5 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
            title="Supprimer le widget"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Automations List */}
      <div className="p-4 space-y-3 flex-1 overflow-y-auto max-h-[480px]">
        {displayedAutomations.length === 0 ? (
          <div className="py-8 text-center text-slate-500 text-xs">
            Aucune automatisation configurée dans ce widget.
          </div>
        ) : (
          displayedAutomations.map((auto) => {
            const isCurrentlyRunning = auto.current > 0
            const isTriggering = triggeringId === auto.id
            const isTriggered = triggeredIds.has(auto.id)

            return (
              <div
                key={auto.id}
                className={`p-3.5 rounded-xl border transition-all flex items-center justify-between gap-3 ${
                  isCurrentlyRunning
                    ? 'bg-emerald-950/20 border-emerald-500/40 shadow-sm shadow-emerald-950/50'
                    : 'bg-slate-800/50 border-slate-800 hover:border-slate-700/80'
                }`}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-semibold text-white truncate">{auto.name}</span>
                    {isCurrentlyRunning && (
                      <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 animate-pulse">
                        <Activity className="w-3 h-3 animate-spin" />
                        <span>En cours ({auto.current})</span>
                      </span>
                    )}
                  </div>

                  <div className="flex items-center space-x-2 mt-1 text-[11px] text-slate-400">
                    <Clock className="w-3 h-3 text-slate-500" />
                    <span>{formatRelativeTime(auto.last_triggered)}</span>
                  </div>
                </div>

                {/* Tactile Trigger Button */}
                <button
                  type="button"
                  onClick={() => handleTrigger(auto.id)}
                  disabled={isTriggering}
                  className={`relative flex items-center justify-center space-x-1.5 px-4 py-2.5 rounded-xl text-xs font-bold transition-all active:scale-95 shadow-lg select-none min-h-[44px] min-w-[120px] ${
                    isTriggered
                      ? 'bg-emerald-600 text-white shadow-emerald-500/30 scale-105'
                      : isTriggering
                      ? 'bg-indigo-700 text-white/80 cursor-wait'
                      : 'bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white shadow-indigo-600/30'
                  }`}
                >
                  {isTriggered ? (
                    <>
                      <CheckCircle2 className="w-4 h-4 text-emerald-200 animate-bounce" />
                      <span>Lancé !</span>
                    </>
                  ) : isTriggering ? (
                    <>
                      <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                      <span>Envoi...</span>
                    </>
                  ) : (
                    <>
                      <Play className="w-3.5 h-3.5 fill-current" />
                      <span>Déclencher</span>
                    </>
                  )}
                </button>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
