import { useCallback, useEffect } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { ChevronRight, Cpu, LayoutDashboard, Layers, ShieldCheck, SlidersHorizontal, type LucideIcon } from 'lucide-react'
import { useApp } from '../../useApp'
import { SetupBannerSlotProvider, useSetSetupBannerSlot } from './SetupBannerSlot'

interface SetupTask {
  to: string
  label: string
  icon: LucideIcon
}

const TASKS: SetupTask[] = [
  { to: '/setup/plans', label: 'Floors & Plans', icon: Layers },
  { to: '/setup/devices', label: 'Devices', icon: Cpu },
  { to: '/setup/dashboards', label: 'Dashboards', icon: LayoutDashboard },
  { to: '/setup/access', label: 'Access', icon: ShieldCheck },
  { to: '/setup/settings', label: 'Settings', icon: SlidersHorizontal },
]

function isTypingTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  const tag = el?.tagName?.toLowerCase()
  return tag === 'input' || tag === 'textarea' || tag === 'select' || !!el?.isContentEditable
}

export function SetupShell() {
  return (
    <SetupBannerSlotProvider>
      <SetupShellContent />
    </SetupBannerSlotProvider>
  )
}

function SetupShellContent() {
  const navigate = useNavigate()
  const location = useLocation()
  const app = useApp()
  const setBannerSlot = useSetSetupBannerSlot()
  // The focused editors are section-specific: hide the cross-section task nav there so they get
  // the full height and preview the view-mode result.
  const hideTaskNav =
    /^\/setup\/dashboards\/[^/]+/.test(location.pathname) ||
    location.pathname.startsWith('/setup/plans')

  const exitSetup = useCallback(() => {
    if (location.pathname.startsWith('/setup/dashboards')) {
      navigate('/dashboards')
      return
    }
    const planMatch = location.pathname.match(/^\/setup\/plans\/([^/]+)/)
    if (planMatch) {
      navigate(`/floor/${planMatch[1]}`)
      return
    }
    navigate('/')
  }, [navigate, location.pathname])

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Escape' || isTypingTarget(e.target)) return
      exitSetup()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [exitSetup])

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden bg-[#08090c] text-slate-100">
      <div className="shrink-0 border-t-2 border-[#6d76e8] border-b border-b-slate-800/80 bg-[#6d76e8]/[0.07] backdrop-blur-md px-4 py-2.5 flex items-center gap-3 select-none">
        <span className="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.12em] text-[#6d76e8]">
          <span className="w-1.5 h-1.5 rounded-full bg-[#6d76e8]" />
          Setup mode
        </span>
        <span className="text-[11.5px] text-slate-400">You are configuring, not viewing</span>
        <div ref={setBannerSlot} className="flex-1 min-w-0 flex items-center justify-center" />
        <button
          type="button"
          onClick={exitSetup}
          className="flex items-center gap-1.5 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 px-3.5 py-2 text-xs font-semibold text-white transition-all cursor-pointer"
        >
          <ChevronRight className="w-3.5 h-3.5" />
          <span>Exit setup</span>
        </button>
      </div>

      {!hideTaskNav && (
        <nav
          aria-label="Setup tasks"
          className="shrink-0 flex items-center gap-1 border-b border-slate-800/80 bg-slate-900/70 backdrop-blur-md px-4 py-2 select-none"
        >
          {TASKS.map((task) => {
            const Icon = task.icon
            return (
              <NavLink
                key={task.to}
                to={task.to}
                className={({ isActive }) =>
                  `flex items-center gap-2 rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors ${
                    isActive
                      ? 'bg-[#6d76e8] text-white'
                      : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
                  }`
                }
              >
                <Icon className="w-3.5 h-3.5" />
                <span>{task.label}</span>
              </NavLink>
            )
          })}
          <div className="flex-1" />
          <span className="text-[11.5px] text-slate-500">Esc to exit</span>
        </nav>
      )}

      <div className="flex-1 min-h-0 flex flex-col overflow-hidden">
        <Outlet context={app} />
      </div>
    </div>
  )
}
