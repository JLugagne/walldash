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

// Actuator widget, Display "toggle": the only Primary Action in this dashboard is a tap that
// toggles the bound entity, so the whole tile is the touch target. The button lives in the
// WidgetFrame overlay and stays transparent, letting the card show a clear switch and its state
// text: the caption icon and the low-alpha amber wash both turn on with the entity. `pending`
// mirrors the in-flight state the caller tracks via useRealtimeDevices' pendingDevices map, the
// same convention DeviceBadge3D uses for the 3D view, and swaps the knob for a spinner.
export const ToggleWidget: React.FC<ToggleWidgetProps> = ({ label, on, pending, stale, dense = false, onToggle }) => {
  const handleTap = () => {
    if (!pending) onToggle()
  }

  const trackClass = `relative inline-flex shrink-0 items-center rounded-full border transition-colors duration-200 ${
    dense ? 'w-9 h-[20px]' : 'w-11 h-6'
  } ${on ? 'bg-[#d99a3a] border-transparent' : 'bg-white/10 border-slate-700'}`

  return (
    <WidgetFrame
      label={label}
      stale={stale}
      dense={dense}
      icon={<Power className={on ? 'text-[#d99a3a]' : 'text-[#6d76e8]'} />}
      bodyClassName="flex items-center"
      overlay={
        <>
          {on && <div className="absolute inset-0 bg-[#d99a3a]/10" aria-hidden="true" />}
          <button
            type="button"
            disabled={pending}
            onClick={handleTap}
            aria-pressed={on}
            title={pending ? `${label} • Processing (Home Assistant)...` : `${label} • ${on ? 'ON' : 'OFF'}`}
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
            {on ? 'On' : 'Off'}
          </span>
          <span className={`truncate tabular-nums ${dense ? 'text-[10px]' : 'text-[11px]'} ${on ? 'text-[#d99a3a]' : 'text-slate-400'}`}>
            {pending ? 'Switching…' : on ? 'Active' : 'Inactive'}
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
              } ${on ? 'bg-white' : 'bg-slate-300'} ${on ? (dense ? 'left-[18px]' : 'left-[22px]') : 'left-0.5'}`}
            />
          )}
        </span>
      </div>
    </WidgetFrame>
  )
}
