import { Plus } from 'lucide-react'
import type { Dashboard } from '../../types'

interface DashboardSelectorProps {
  dashboards: Dashboard[]
  activeDashboardId: string | null
  onSelectDashboard: (dashboardId: string) => void
  onCreateDashboard?: () => void
  /** Show the active dashboard pill even when there is only one dashboard. */
  showSingle?: boolean
}

export function DashboardSelector({
  dashboards,
  activeDashboardId,
  onSelectDashboard,
  onCreateDashboard,
  showSingle = false,
}: DashboardSelectorProps) {
  if (dashboards.length === 0 && !onCreateDashboard) return null

  const sortedDashboards = [...dashboards].sort((a, b) => a.order - b.order)
  const showPills = dashboards.length > 1 || showSingle

  return (
    <div
      role="group"
      aria-label="Dashboard selection"
      className="bg-slate-900/70 backdrop-blur-md border border-slate-800/80 rounded-xl p-1 flex items-center gap-1 pointer-events-auto"
    >
      {showPills &&
        sortedDashboards.map((dashboard) => {
          const isActive = dashboard.id === activeDashboardId
          return (
            <button
              key={dashboard.id}
              type="button"
              onClick={() => onSelectDashboard(dashboard.id)}
              aria-pressed={isActive}
              aria-label={`Dashboard ${dashboard.name}`}
              className={`px-3.5 py-1.5 max-sm:px-2.5 rounded-lg text-xs transition-all cursor-pointer whitespace-nowrap select-none ${
                isActive
                  ? 'bg-[#6d76e8] text-white font-semibold'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
              }`}
            >
              {dashboard.name}
            </button>
          )
        })}

      {onCreateDashboard && (
        <button
          type="button"
          onClick={onCreateDashboard}
          title="New dashboard"
          aria-label="New dashboard"
          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
        >
          <Plus className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  )
}
