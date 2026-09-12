import React from 'react'
import { Power } from 'lucide-react'
import { WidgetFrame } from './WidgetFrame'

export interface ToggleWidgetProps {
  label: string
  on: boolean
  pending: boolean
  stale: boolean
  dense?: boolean
  /**
   * 1x1 layout: the body becomes a single exclusive ON/OFF button that fills the whole tile
   * instead of the reading + track pair. Used for one-cell switches so the state is unmistakable.
   */
  fullTile?: boolean
  /**
   * Caption glyph, already coloured by the caller. Defaults to the Power icon; the Automation
   * switch passes Zap so the two keep distinct identities while sharing the exact same switch UI.
   */
  icon?: React.ReactNode
  onToggle: () => void
}

// The single switch UI shared by every Widget whose Primary Action is a toggle: actuator devices
// and automation switches both render through this component, so a tile always shows one state
// (On/Off) and one control, never a pair of ON/OFF segments. The whole tile is the touch target:
// the transparent button lives in the WidgetFrame overlay, while the caption icon and the
// low-alpha amber wash turn on with the state. `pending` mirrors the in-flight state the caller
// tracks (useRealtimeDevices' pendingDevices map for devices, the local optimistic state for
// automations) and swaps the knob for a spinner. On a 1x1 tile (`fullTile`) the state itself is
// the button and fills the body.
export const ToggleWidget: React.FC<ToggleWidgetProps> = ({
  label,
  on,
  pending,
  stale,
  dense = false,
  fullTile = false,
  icon,
  onToggle,
}) => {
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
      icon={icon ?? <Power className={on ? 'text-[#d99a3a]' : 'text-[#6d76e8]'} />}
      bodyClassName={fullTile ? 'flex items-stretch' : 'flex items-center'}
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
      {fullTile ? (
        <div className="pointer-events-none w-full h-full flex items-stretch">
          <span
            className={`flex-1 min-w-0 rounded-lg border flex items-center justify-center font-bold tracking-tight transition-colors duration-200 ${
              on
                ? 'border-[#d99a3a]/50 bg-[#d99a3a]/20 text-[#f0c274]'
                : 'border-slate-600 bg-white/[0.06] text-white'
            }`}
            style={{ fontSize: 'clamp(0.875rem, min(34cqh, 24cqw), 2.25rem)' }}
          >
            {pending ? (
              <span className="w-5 h-5 rounded-full border-2 border-[#8b93ee]/30 border-t-[#8b93ee] animate-spin" />
            ) : (
              (on ? 'ON' : 'OFF')
            )}
          </span>
        </div>
      ) : (
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
      )}
    </WidgetFrame>
  )
}
