import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Box, LayoutDashboard, Settings, type LucideIcon } from 'lucide-react'
import { useRealtimeDeviceControl } from '../hooks/useRealtimeDevices'
import { useSetTopBarSlot } from './TopBarSlot'

type ViewMode = '3d' | 'house' | 'dashboards'

interface ViewSegment {
  key: ViewMode
  label: string
  icon: LucideIcon
  to: string
}

const SEGMENTS: ViewSegment[] = [
  { key: '3d', label: '3D View', icon: Box, to: '/' },
  { key: 'dashboards', label: 'Dashboards', icon: LayoutDashboard, to: '/dashboards' },
]

function resolveMode(pathname: string): ViewMode {
  if (pathname.startsWith('/house')) return 'house'
  if (pathname.startsWith('/dashboards')) return 'dashboards'
  return '3d'
}

interface AppTopBarProps {
  activeLevelId: string | null
}

export function AppTopBar({ activeLevelId }: AppTopBarProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const { connected } = useRealtimeDeviceControl()
  const setSlot = useSetTopBarSlot()

  const mode = resolveMode(location.pathname)

  const openSetup = () => {
    if (mode === 'dashboards') {
      let activeDashboardId: string | null = null
      try {
        activeDashboardId = window.sessionStorage.getItem('walldash:active-dashboard')
      } catch {
        activeDashboardId = null
      }
      navigate(activeDashboardId ? `/setup/dashboards/${activeDashboardId}` : '/setup/dashboards')
      return
    }
    navigate(mode === '3d' && activeLevelId ? `/setup/plans/${activeLevelId}` : '/setup/plans')
  }

  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 30_000)
    return () => window.clearInterval(timer)
  }, [])
  const time = now.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
  const date = now.toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long' })

  return (
    <header className="relative z-30 h-14 shrink-0 bg-slate-900/70 backdrop-blur-md border-b border-slate-800/80 flex items-center px-3 gap-3 select-none">
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

      <div ref={setSlot} className="flex-1 min-w-0 flex items-center" />

      <div
        className="flex shrink-0 items-center gap-2.5"
        title={connected ? 'Connected' : 'Offline'}
      >
        <span className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-[#42a67d]' : 'bg-slate-500'}`} />
        <span className="text-sm font-semibold leading-none tabular-nums text-white">{time}</span>
        <span className="text-xs leading-none text-slate-400">{date}</span>
      </div>

      <button
        type="button"
        onClick={openSetup}
        title="Open setup mode"
        className="h-9 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 flex items-center gap-1.5 text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
      >
        <Settings className="w-3.5 h-3.5" />
        <span>Setup</span>
      </button>
    </header>
  )
}
