import React, { useEffect, useRef, useState } from 'react'
import { Zap, Play, Check } from 'lucide-react'
import type { Widget, Automation } from '../../../types'
import { apiFetch } from '../../../api'
import { WidgetFrame } from './WidgetFrame'

export interface AutomationListWidgetProps {
  widget: Widget
  automations: Automation[]
  /** Called after Home Assistant accepted the trigger so the caller can refresh its automation poll early. */
  onTriggerSuccess?: (automationId: string) => void
}

const TRIGGERED_FEEDBACK_MS = 2000

function formatRelativeTime(isoString: string | null): string {
  if (!isoString) return 'Jamais'
  const date = new Date(isoString)
  if (isNaN(date.getTime())) return 'Jamais'
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

function selectAutomations(widget: Widget, automations: Automation[]): Automation[] {
  const configuredEntityIds = widget.config?.entity_ids || []
  if (configuredEntityIds.length === 0) return automations
  return automations.filter((a) => configuredEntityIds.includes(a.id))
}

/**
 * AutomationListWidget renders the "list" Display Mode for automations inside a WidgetFrame: one
 * 40 px row per automation with a status dot, its per-widget label (ADR 0005) or Home Assistant
 * name, the relative last-trigger time and an icon-only trigger button as the Primary Action. The
 * list scrolls inside the cell; nothing is painted outside the card. Deletion is not offered here:
 * like every other Widget it is removed through Edit Mode.
 */
export const AutomationListWidget: React.FC<AutomationListWidgetProps> = ({ widget, automations, onTriggerSuccess }) => {
  const [triggeringId, setTriggeringId] = useState<string | null>(null)
  const [triggeredIds, setTriggeredIds] = useState<Set<string>>(new Set())
  const feedbackTimersRef = useRef<Map<string, number>>(new Map())

  useEffect(() => {
    const timers = feedbackTimersRef.current
    return () => {
      timers.forEach((timer) => window.clearTimeout(timer))
      timers.clear()
    }
  }, [])

  const displayedAutomations = selectAutomations(widget, automations)
  const labels = widget.config?.labels ?? {}
  const label = widget.title || 'Automatisations'
  const count = displayedAutomations.length

  const markTriggered = (automationId: string) => {
    setTriggeredIds((prev) => new Set(prev).add(automationId))
    const previous = feedbackTimersRef.current.get(automationId)
    if (previous !== undefined) window.clearTimeout(previous)
    const timer = window.setTimeout(() => {
      feedbackTimersRef.current.delete(automationId)
      setTriggeredIds((prev) => {
        const next = new Set(prev)
        next.delete(automationId)
        return next
      })
    }, TRIGGERED_FEEDBACK_MS)
    feedbackTimersRef.current.set(automationId, timer)
  }

  const handleTrigger = async (automationId: string) => {
    if (triggeringId) return
    setTriggeringId(automationId)

    try {
      const res = await apiFetch(`/api/automations/${encodeURIComponent(automationId)}/trigger`, {
        method: 'POST',
      })
      if (res.ok) {
        markTriggered(automationId)
        onTriggerSuccess?.(automationId)
      }
    } catch (err) {
      console.error('Failed to trigger automation:', err)
    } finally {
      setTriggeringId(null)
    }
  }

  return (
    <WidgetFrame
      label={label}
      stale={false}
      icon={<Zap className="text-amber-400" />}
      trailing={
        <span className="text-[10px] leading-none font-medium text-slate-500 tabular-nums whitespace-nowrap">
          {count}
        </span>
      }
      bodyClassName="flex flex-col"
    >
      {count === 0 ? (
        <div className="flex-1 flex items-center justify-center text-center text-slate-500 text-xs px-2">
          Aucune automatisation configurée.
        </div>
      ) : (
        <ul className="flex-1 min-h-0 overflow-y-auto divide-y divide-slate-800/70 [scrollbar-width:thin]">
          {displayedAutomations.map((auto) => {
            const isCurrentlyRunning = auto.current > 0
            const isTriggering = triggeringId === auto.id
            const isTriggered = triggeredIds.has(auto.id)
            const name = labels[auto.id] ?? auto.name

            return (
              <li key={auto.id} className="h-10 flex items-center gap-2 min-w-0">
                <span
                  className={`w-2 h-2 shrink-0 rounded-full ${
                    isCurrentlyRunning ? 'bg-emerald-400 animate-pulse' : 'bg-slate-600'
                  }`}
                  aria-hidden="true"
                />
                <div className="min-w-0 flex-1 flex flex-col justify-center leading-none gap-1">
                  <span className="text-xs font-semibold text-white truncate" title={name}>
                    {name}
                  </span>
                  <span className="text-[10px] text-slate-400 truncate tabular-nums">
                    {isCurrentlyRunning ? `En cours (${auto.current})` : formatRelativeTime(auto.last_triggered)}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => handleTrigger(auto.id)}
                  disabled={isTriggering}
                  aria-label={`Déclencher ${name}`}
                  title="Déclencher"
                  className={`w-9 h-9 shrink-0 flex items-center justify-center rounded-lg transition-colors select-none touch-manipulation ${
                    isTriggered
                      ? 'bg-emerald-600 text-white'
                      : isTriggering
                        ? 'bg-slate-800 text-indigo-300 cursor-wait'
                        : 'bg-indigo-600 active:bg-indigo-700 text-white'
                  }`}
                >
                  {isTriggered ? (
                    <Check className="w-4 h-4" />
                  ) : isTriggering ? (
                    <span className="w-3.5 h-3.5 border-2 border-indigo-400/30 border-t-indigo-400 rounded-full animate-spin" />
                  ) : (
                    <Play className="w-4 h-4 fill-current" />
                  )}
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </WidgetFrame>
  )
}
