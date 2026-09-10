import { useLocation, useNavigate } from 'react-router-dom'
import { Box, LayoutDashboard, Image, Settings, type LucideIcon } from 'lucide-react'
import { useRealtimeDeviceControl } from '../hooks/useRealtimeDevices'
import { useSetTopBarSlot } from './TopBarSlot'

type ViewMode = '3d' | 'house' | 'overviews'

interface ViewSegment {
  key: ViewMode
  label: string
  icon: LucideIcon
  to: string
}

const SEGMENTS: ViewSegment[] = [
  { key: '3d', label: '3D View', icon: Box, to: '/' },
  { key: 'overviews', label: 'Overviews', icon: LayoutDashboard, to: '/overviews' },
]

function resolveMode(pathname: string): ViewMode {
  if (pathname.startsWith('/house')) return 'house'
  if (pathname.startsWith('/overviews')) return 'overviews'
  return '3d'
}

export function AppTopBar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { connected } = useRealtimeDeviceControl()
  const setSlot = useSetTopBarSlot()

  const mode = resolveMode(location.pathname)

  return (
    <header className="h-14 shrink-0 bg-slate-900/70 backdrop-blur-md border-b border-slate-800/80 flex items-center px-3 gap-3 select-none">
      <nav
        aria-label="View switcher"
        className="rounded-xl border border-slate-800/80 bg-slate-900/70 p-1 flex gap-1"
      >
        {SEGMENTS.map((segment) => {
          const Icon = segment.icon
          const isActive = segment.key === mode
          return (
            <button
              key={segment.key}
              type="button"
              onClick={() => navigate(segment.to)}
              aria-current={isActive ? 'page' : undefined}
              className={
                isActive
                  ? 'bg-[#6d76e8] text-white rounded-lg px-3 py-1.5 text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800/60 rounded-lg px-3 py-1.5 text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer'
              }
            >
              <Icon className="w-3.5 h-3.5" />
              <span>{segment.label}</span>
            </button>
          )
        })}
      </nav>

      <div ref={setSlot} className="flex-1 min-w-0 overflow-x-auto" />

      {mode === 'overviews' && (
        <button
          type="button"
          onClick={() => window.dispatchEvent(new CustomEvent('walldash:open-background'))}
          title="Dashboard background"
          aria-label="Dashboard background"
          className="h-9 w-9 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
        >
          <Image className="w-4 h-4" />
        </button>
      )}

      <div className="flex items-center gap-1.5 rounded-lg border border-slate-800/80 bg-slate-900/70 px-2.5 py-1 text-[11px] font-semibold text-slate-300">
        <span className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-[#42a67d]' : 'bg-slate-500'}`} />
        <span>{connected ? 'Connected' : 'Offline'}</span>
      </div>

      <button
        type="button"
        onClick={() => navigate('/setup')}
        title="Open setup mode"
        className="h-9 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 flex items-center gap-1.5 text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
      >
        <Settings className="w-3.5 h-3.5" />
        <span>Setup</span>
      </button>
    </header>
  )
}
