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
  if (!isoString) return 'Never'
  const date = new Date(isoString)
  if (isNaN(date.getTime())) return 'Never'
  const now = new Date()
  const diffSec = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diffSec < 10) return 'Just now'
  if (diffSec < 60) return `${diffSec}s ago`
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin}min ago`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString('en-US', {
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
  const label = widget.title || 'Automations'
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
      icon={<Zap className="text-[#d99a3a]" />}
      trailing={
        <span className="text-[10px] leading-none font-medium text-slate-500 tabular-nums whitespace-nowrap">
          {count}
        </span>
      }
      bodyClassName="flex flex-col"
    >
      {count === 0 ? (
        <div className="flex-1 flex items-center justify-center text-center text-slate-500 text-xs px-2">
          No automations configured.
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
                    isCurrentlyRunning ? 'bg-[#42a67d] animate-pulse' : 'bg-slate-600'
                  }`}
                  aria-hidden="true"
                />
                <div className="min-w-0 flex-1 flex flex-col justify-center leading-none gap-1">
                  <span className="text-xs font-semibold text-white truncate" title={name}>
                    {name}
                  </span>
                  <span className="text-[10px] text-slate-500 truncate tabular-nums">
                    {isCurrentlyRunning ? `Running (${auto.current})` : formatRelativeTime(auto.last_triggered)}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => handleTrigger(auto.id)}
                  disabled={isTriggering}
                  aria-label={`Trigger ${name}`}
                  title="Trigger"
                  className={`w-9 h-9 shrink-0 flex items-center justify-center rounded-lg transition-colors select-none touch-manipulation ${
                    isTriggered
                      ? 'bg-[#42a67d] text-white'
                      : isTriggering
                        ? 'bg-slate-800 text-[#8b93ee] cursor-wait'
                        : 'bg-[#6d76e8] active:bg-[#5b64d4] text-white'
                  }`}
                >
                  {isTriggered ? (
                    <Check className="w-4 h-4" />
                  ) : isTriggering ? (
                    <span className="w-3.5 h-3.5 border-2 border-[#8b93ee]/30 border-t-[#8b93ee] rounded-full animate-spin" />
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
