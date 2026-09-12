import React, { useEffect, useState } from 'react'
import { Zap } from 'lucide-react'
import type { Automation, Widget } from '../../../types'
import { apiFetch } from '../../../api'
import { ToggleWidget } from './ToggleWidget'

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

// Automation switch Widget: its Primary Action triggers one of two bound automations (the "on" one
// when currently off, the "off" one when currently on). It owns only the automation-specific logic
// — deriving the state from the two `last_triggered` timestamps and firing the right trigger — and
// renders through ToggleWidget so it looks and behaves exactly like an actuator device switch.
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

  useEffect(() => {
    if (optimistic !== null && derived === optimistic) setOptimistic(null)
  }, [derived, optimistic])

  const isOn = (optimistic ?? derived) === 'on'
  const label = widget.title || 'Automations'

  const handleTap = async () => {
    if (pending !== null || !onId || !offId) return
    const target: Position = isOn ? 'off' : 'on'
    const targetId = target === 'on' ? onId : offId
    setPending(target)
    setOptimistic(target)
    try {
      const res = await apiFetch(`/api/automations/${encodeURIComponent(targetId)}/trigger`, { method: 'POST' })
      if (!res.ok) {
        setOptimistic(null)
        return
      }
      onTriggerSuccess?.(targetId)
    } catch (err) {
      console.error('Failed to trigger automation:', err)
      setOptimistic(null)
    } finally {
      setPending(null)
    }
  }

  return (
    <ToggleWidget
      label={label}
      on={isOn}
      pending={pending !== null}
      stale={false}
      dense={widget.row_span === 1}
      fullTile={widget.col_span === 1 && widget.row_span === 1}
      icon={<Zap className={isOn ? 'text-[#d99a3a]' : 'text-[#6d76e8]'} />}
      onToggle={handleTap}
    />
  )
}
