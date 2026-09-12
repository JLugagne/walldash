import React, { useEffect, useState } from 'react'
import { Zap } from 'lucide-react'
import type { Automation, Widget } from '../../../types'
import { apiFetch } from '../../../api'
import { WidgetFrame } from './WidgetFrame'

export interface AutomationSwitchWidgetProps {
  widget: Widget
  automations: Automation[]
  onTriggerSuccess?: (automationId: string) => void
}

type Position = 'on' | 'off'

function triggeredAt(automation: Automation | undefined): number {
  if (!automation?.last_triggered) return 0
  const time = new Date(automation.last_triggered).getTime()
  return Number.isNaN(time) ? 0 : time
}

function derivePosition(onAutomation: Automation | undefined, offAutomation: Automation | undefined): Position {
  return triggeredAt(onAutomation) > triggeredAt(offAutomation) ? 'on' : 'off'
}

export const AutomationSwitchWidget: React.FC<AutomationSwitchWidgetProps> = ({
  widget,
  automations,
  onTriggerSuccess,
}) => {
  const onId = widget.config.on_automation
  const offId = widget.config.off_automation
  const onAutomation = automations.find((a) => a.id === onId)
  const offAutomation = automations.find((a) => a.id === offId)

  const derived = derivePosition(onAutomation, offAutomation)
  const [optimistic, setOptimistic] = useState<Position | null>(null)
  const [pending, setPending] = useState<Position | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    if (optimistic !== null && derived === optimistic) setOptimistic(null)
  }, [derived, optimistic])

  const position = optimistic ?? derived
  const isOn = position === 'on'
  const dense = widget.row_span === 1

  const handleTap = async () => {
    if (pending !== null || !onId || !offId) return
    const target: Position = isOn ? 'off' : 'on'
    const targetId = target === 'on' ? onId : offId
    setPending(target)
    setOptimistic(target)
    setFailed(false)
    try {
      const res = await apiFetch(`/api/automations/${encodeURIComponent(targetId)}/trigger`, { method: 'POST' })
      if (!res.ok) {
        setOptimistic(null)
        setFailed(true)
        return
      }
      onTriggerSuccess?.(targetId)
    } catch (err) {
      console.error('Failed to trigger automation:', err)
      setOptimistic(null)
      setFailed(true)
    } finally {
      setPending(null)
    }
  }

  const label = widget.title || 'Automations'
  const activeAutomation = isOn ? onAutomation : offAutomation
  const stateText = pending ? 'Switching…' : failed ? 'Failed' : isOn ? 'On' : 'Off'
  const subtitle = activeAutomation?.name || (isOn ? onId : offId) || 'Not configured'

  const trackClass = `relative inline-flex shrink-0 items-center rounded-full border transition-colors duration-200 ${
    dense ? 'w-9 h-[20px]' : 'w-11 h-6'
  } ${isOn ? 'bg-[#d99a3a] border-transparent' : 'bg-white/10 border-slate-700'}`

  return (
    <WidgetFrame
      label={label}
      stale={false}
      dense={dense}
      icon={<Zap className={isOn ? 'text-[#d99a3a]' : 'text-[#6d76e8]'} />}
      bodyClassName="flex items-center"
      overlay={
        <>
          {isOn && <div className="absolute inset-0 bg-[#d99a3a]/10" aria-hidden="true" />}
          <button
            type="button"
            disabled={pending !== null}
            onClick={handleTap}
            aria-pressed={isOn}
            aria-label={`${label} • ${isOn ? 'On' : 'Off'}`}
            title={pending ? `${label} • Processing (Home Assistant)...` : `${label} • ${isOn ? 'On' : 'Off'}`}
            className={`absolute inset-0 w-full h-full outline-none select-none transition-colors duration-200 ${
              pending ? 'cursor-wait bg-slate-900/30' : 'cursor-pointer active:bg-white/5'
            }`}
          />
        </>
      }
    >
      <div className="pointer-events-none w-full flex items-center gap-3 min-w-0">
        <div className="min-w-0 flex-1 flex flex-col justify-center leading-none gap-1">
          <span className={`font-semibold tracking-tight text-white truncate ${dense ? 'text-xs' : 'text-sm'}`}>
            {stateText}
          </span>
          <span
            className={`truncate ${dense ? 'text-[10px]' : 'text-[11px]'} ${isOn ? 'text-[#d99a3a]' : 'text-slate-400'}`}
          >
            {subtitle}
          </span>
        </div>
        <span className={trackClass} aria-hidden="true">
          {pending ? (
            <span
              className={`absolute inset-0 m-auto rounded-full border-2 border-[#8b93ee]/30 border-t-[#8b93ee] animate-spin ${
                dense ? 'w-3 h-3' : 'w-3.5 h-3.5'
              }`}
            />
          ) : (
            <span
              className={`absolute top-1/2 -translate-y-1/2 rounded-full transition-all duration-200 ${
                dense ? 'w-3.5 h-3.5' : 'w-4 h-4'
              } ${isOn ? 'bg-white' : 'bg-slate-300'} ${isOn ? (dense ? 'left-[18px]' : 'left-[22px]') : 'left-0.5'}`}
            />
          )}
        </span>
      </div>
    </WidgetFrame>
  )
}
