import { useEffect, useRef, useState } from 'react'
import { Box, LayoutDashboard, Edit3, Layers, type LucideIcon } from 'lucide-react'

export type ViewMode = '3d' | 'admin' | 'overviews'

interface ViewModeOption {
  key: ViewMode
  label: string
  icon: LucideIcon
}

const OPTIONS: ViewModeOption[] = [
  { key: '3d', label: '3D View', icon: Box },
  { key: 'overviews', label: 'Overviews', icon: LayoutDashboard },
  { key: 'admin', label: '2D Editor', icon: Edit3 },
]

interface ViewModeMenuProps {
  mode: ViewMode
  onSelect: (mode: ViewMode) => void
  direction?: 'up' | 'down'
  useLayersIcon?: boolean
}

export function ViewModeMenu({
  mode,
  onSelect,
  direction = 'down',
  useLayersIcon,
}: ViewModeMenuProps) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const activeOption = OPTIONS.find((o) => o.key === mode) ?? OPTIONS[0]
  const showLayersIcon = useLayersIcon ?? (direction === 'up')
  const TriggerIcon = showLayersIcon ? Layers : activeOption.icon

  useEffect(() => {
    if (!open) return
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  const dropdownPositionClass =
    direction === 'up'
      ? 'absolute bottom-full right-0 mb-2.5'
      : 'absolute top-full left-0 mt-2'

  return (
    <div ref={containerRef} className="relative pointer-events-auto">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        title="Switch view (Layers)"
        aria-label="Switch view"
        aria-expanded={open}
        className="w-12 h-12 rounded-2xl bg-slate-900/90 backdrop-blur-md border border-slate-800/80 shadow-2xl shadow-black/50 flex items-center justify-center text-indigo-400 hover:text-white hover:bg-slate-800/90 active:scale-95 transition-all cursor-pointer"
      >
        <TriggerIcon className="w-5 h-5" />
      </button>

      {open && (
        <div className={`${dropdownPositionClass} bg-slate-900/95 backdrop-blur-md border border-slate-800/80 rounded-2xl shadow-2xl shadow-black/50 p-1.5 flex flex-col space-y-1 min-w-[180px] z-30`}>
          {OPTIONS.map((option) => {
            const Icon = option.icon
            const isActive = option.key === mode
            return (
              <button
                key={option.key}
                type="button"
                onClick={() => {
                  onSelect(option.key)
                  setOpen(false)
                }}
                className={`flex items-center space-x-2.5 px-3 py-2.5 rounded-xl text-xs font-semibold transition-all active:scale-95 cursor-pointer ${
                  isActive
                    ? 'bg-indigo-600 text-white shadow-md shadow-indigo-500/30'
                    : 'text-slate-300 hover:text-white hover:bg-slate-800/70'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{option.label}</span>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
