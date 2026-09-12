import { useNavigate } from 'react-router-dom'
import { ChevronRight, Image, Info, SlidersHorizontal } from 'lucide-react'

export function SettingsPanel() {
  const navigate = useNavigate()

  const openDashboards = () => {
    navigate('/setup/dashboards')
  }

  return (
    <div className="flex-1 min-h-0 overflow-y-auto p-4">
      <div className="max-w-2xl space-y-4">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Settings</h1>
          <p className="text-[11.5px] text-slate-400">Application configuration</p>
        </div>

        <section className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-4">
          <div className="flex items-start gap-3">
            <span className="w-9 h-9 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-[#8b93ee]">
              <Image className="w-4 h-4" />
            </span>
            <div className="min-w-0 flex-1">
              <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
                Dashboard backgrounds
              </h2>
              <p className="text-xs text-slate-400 mt-1.5 leading-relaxed">
                Backgrounds are configured per dashboard inside the dashboard editor. Open the
                dashboards manager to choose a dashboard and edit it.
              </p>
              <button
                type="button"
                onClick={openDashboards}
                className="mt-3 flex items-center gap-1.5 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 px-3 py-2 text-xs font-semibold text-white transition-all cursor-pointer"
              >
                <SlidersHorizontal className="w-3.5 h-3.5" />
                <span>Open dashboards</span>
                <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </section>

        <section className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-4">
          <div className="flex items-start gap-3">
            <span className="w-9 h-9 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
              <Info className="w-4 h-4" />
            </span>
            <div className="min-w-0 flex-1">
              <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
                About
              </h2>
              <dl className="mt-2 space-y-1.5 text-xs">
                <div className="flex items-center justify-between gap-4">
                  <dt className="text-slate-500">Application</dt>
                  <dd className="text-slate-200 font-semibold">Walldash</dd>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <dt className="text-slate-500">Configuration</dt>
                  <dd className="text-slate-200">Stored on your Home Assistant server</dd>
                </div>
              </dl>
              <p className="text-[11.5px] text-slate-500 mt-3 leading-relaxed">
                Floors, devices, placements and Dashboards are served by the local
                backend and are not versioned in the browser.
              </p>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}
