import { useCallback, useEffect, useState } from 'react'
import { Cpu, RefreshCw } from 'lucide-react'
import type { Device } from '../../types'

export function DevicesList() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/devices')
      const payload = await res.json()
      if (res.ok && payload?.status === 'success' && Array.isArray(payload.data)) {
        setDevices(payload.data)
      } else {
        setError('Unable to load devices')
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <div className="shrink-0 flex items-center gap-3 px-4 py-3 border-b border-slate-800/80">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Devices</h1>
          <p className="text-[11.5px] text-slate-400">
            {devices.length} device{devices.length === 1 ? '' : 's'} from Home Assistant
          </p>
        </div>
        <div className="flex-1" />
        <button
          type="button"
          onClick={() => void load()}
          disabled={loading}
          className="flex items-center gap-1.5 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-2 text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-800/60 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        {loading && devices.length === 0 ? (
          <div className="flex items-center gap-2 text-slate-400 text-sm">
            <RefreshCw className="w-4 h-4 animate-spin" />
            <span>Loading devices…</span>
          </div>
        ) : error ? (
          <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-300">
            {error}
          </div>
        ) : devices.length === 0 ? (
          <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
            <Cpu className="w-6 h-6 mx-auto text-slate-500 mb-2" />
            <p className="text-sm text-slate-300">No devices found</p>
            <p className="text-[11.5px] text-slate-500 mt-1">
              Devices are discovered from your Home Assistant instance.
            </p>
          </div>
        ) : (
          <ul className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md divide-y divide-slate-800/80 overflow-hidden">
            {devices.map((device) => (
              <li key={device.id} className="flex items-center gap-3 px-4 py-3">
                <span className="w-8 h-8 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
                  <Cpu className="w-4 h-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-slate-100 truncate">{device.name || device.id}</p>
                  <p className="text-[11px] uppercase tracking-[0.08em] text-slate-500 truncate">
                    {device.domain}
                  </p>
                </div>
                <span className="shrink-0 rounded-md border border-slate-800/80 bg-slate-900/70 px-2.5 py-1 text-[11px] font-semibold text-slate-300 tabular-nums">
                  {device.state}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
