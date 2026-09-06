import React from 'react'
import { Power } from 'lucide-react'
import { WidgetFrame } from './WidgetFrame'

export interface ToggleWidgetProps {
  label: string
  on: boolean
  pending: boolean
  stale: boolean
  dense?: boolean
  onToggle: () => void
}

// Actuator widget, Display "toggle": the only Primary Action in this
// dashboard is a tap that toggles the bound entity, so the button fills the
// whole body and the touch target is the cell itself. `pending` mirrors the
// in-flight state the caller tracks via useRealtimeDevices' pendingDevices
// map, the same convention DeviceBadge3D uses for the 3D view.
export const ToggleWidget: React.FC<ToggleWidgetProps> = ({ label, on, pending, stale, dense = false, onToggle }) => {
  const handleTap = () => {
    if (!pending) onToggle()
  }

  return (
    <WidgetFrame label={label} stale={stale} dense={dense} icon={<Power className="text-indigo-400" />}>
      <button
        type="button"
        disabled={pending}
        onClick={handleTap}
        aria-pressed={on}
        title={pending ? `${label} • En cours de traitement (Home Assistant)...` : `${label} • ${on ? 'ON' : 'OFF'}`}
        className={`absolute inset-0 rounded-lg border-2 flex items-center justify-center transition-colors duration-200 select-none outline-none ${
          pending
            ? 'cursor-wait bg-slate-800/90 border-indigo-400/70 text-indigo-300'
            : on
            ? 'cursor-pointer bg-indigo-500 border-indigo-300 text-white shadow-lg shadow-indigo-500/40 active:bg-indigo-400'
            : 'cursor-pointer bg-slate-800/90 border-slate-600 text-slate-400 active:bg-slate-700'
        }`}
      >
        {pending && (
          <span className="absolute inset-1 rounded-md border-2 border-indigo-400 border-t-transparent animate-spin pointer-events-none" />
        )}
        <Power className={dense ? 'w-6 h-6' : 'w-8 h-8'} />
      </button>
    </WidgetFrame>
  )
}
