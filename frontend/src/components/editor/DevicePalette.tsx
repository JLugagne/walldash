import { useMemo, useState } from 'react'
import { Check, Crosshair, Loader2, RefreshCw, Search, X } from 'lucide-react'
import type { Device } from '../../types'
import { DOMAIN_CATEGORIES, domainStyle, isDeviceActive } from './constants'

interface DevicePaletteProps {
  devices: Device[]
  loading: boolean
  placedDeviceIds: Set<string>
  deviceToPlace: Device | null
  onPickDevice: (device: Device | null) => void
  onRefresh: () => void
}

export function DevicePalette({ devices, loading, placedDeviceIds, deviceToPlace, onPickDevice, onRefresh }: DevicePaletteProps) {
  const [category, setCategory] = useState('all')
  const [search, setSearch] = useState('')
  const [hidePlaced, setHidePlaced] = useState(false)

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return devices.filter((d) => {
      if (category !== 'all' && d.domain !== category) return false
      if (hidePlaced && placedDeviceIds.has(d.id)) return false
      if (!q) return true
      return d.name.toLowerCase().includes(q) || d.id.toLowerCase().includes(q)
    })
  }, [devices, category, search, hidePlaced, placedDeviceIds])

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="p-3 border-b border-slate-800 space-y-2">
        <div className="relative">
          <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search a device…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 text-slate-200 text-xs rounded-lg pl-8 pr-7 py-2 focus:outline-none focus:border-indigo-500 placeholder:text-slate-600"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white cursor-pointer"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
        <div className="flex flex-wrap gap-1">
          {DOMAIN_CATEGORIES.map((cat) => (
            <button
              key={cat.key}
              type="button"
              onClick={() => setCategory(cat.key)}
              className={`px-2 py-1 rounded-md text-[11px] font-medium transition-colors cursor-pointer ${
                category === cat.key ? 'bg-indigo-600 text-white' : 'bg-slate-800/80 text-slate-400 hover:text-white'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
        <div className="flex items-center justify-between text-[11px] text-slate-500">
          <label className="flex items-center gap-1.5 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={hidePlaced}
              onChange={(e) => setHidePlaced(e.target.checked)}
              className="w-3 h-3 accent-indigo-500"
            />
            <span>Hide placed devices</span>
          </label>
          <button
            type="button"
            onClick={onRefresh}
            title="Reload from Home Assistant"
            className="p-1 rounded hover:bg-slate-800 hover:text-white cursor-pointer"
          >
            <RefreshCw className={`w-3 h-3 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-2 space-y-1">
        {loading && devices.length === 0 ? (
          <div className="p-6 text-center text-xs text-slate-500 flex items-center justify-center gap-2">
            <Loader2 className="w-3.5 h-3.5 animate-spin" /> Loading…
          </div>
        ) : filtered.length === 0 ? (
          <div className="p-6 text-center text-xs text-slate-500">No devices found</div>
        ) : (
          filtered.map((dev) => {
            const style = domainStyle(dev.domain)
            const Icon = style.Icon
            const placed = placedDeviceIds.has(dev.id)
            const placing = deviceToPlace?.id === dev.id
            const active = isDeviceActive(dev.state)
            return (
              <div
                key={dev.id}
                draggable
                onDragStart={(e) => {
                  e.dataTransfer.setData('text/plain', dev.id)
                  e.dataTransfer.effectAllowed = 'copy'
                  const iconEl = e.currentTarget.querySelector('[data-drag-icon]') as HTMLElement | null
                  if (iconEl) {
                    const rect = iconEl.getBoundingClientRect()
                    e.dataTransfer.setDragImage(iconEl, rect.width / 2, rect.height / 2)
                  }
                }}
                className={`group flex items-center gap-2.5 p-2 rounded-lg border transition-all cursor-grab active:cursor-grabbing ${
                  placing
                    ? 'border-indigo-500 bg-indigo-950/40'
                    : 'border-transparent hover:border-slate-700 hover:bg-slate-800/60'
                }`}
              >
                <div
                  data-drag-icon
                  className={`relative w-9 h-9 rounded-xl border flex items-center justify-center shrink-0 bg-slate-950 ${style.bg} ${style.border} ${style.text}`}
                >
                  <Icon className="w-4 h-4" />
                  <span
                    className={`absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full ring-2 ring-slate-900 ${
                      active ? 'bg-emerald-400' : 'bg-slate-600'
                    }`}
                  />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-xs font-semibold text-slate-200 truncate" title={dev.name}>
                    {dev.name}
                  </div>
                  <div className="text-[10px] text-slate-500 font-mono truncate" title={dev.id}>
                    {dev.id}
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  {placed && !placing && (
                    <span className="text-emerald-400" title="Already placed on this level">
                      <Check className="w-3.5 h-3.5" />
                    </span>
                  )}
                  <button
                    type="button"
                    onClick={() => onPickDevice(placing ? null : dev)}
                    title={placing ? 'Cancel placement' : 'Place by clicking on the plan'}
                    className={`w-7 h-7 rounded-lg flex items-center justify-center transition-colors cursor-pointer ${
                      placing
                        ? 'bg-rose-600 text-white hover:bg-rose-500'
                        : 'text-slate-400 hover:text-white hover:bg-indigo-600 opacity-0 group-hover:opacity-100'
                    }`}
                  >
                    {placing ? <X className="w-3.5 h-3.5" /> : <Crosshair className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )
          })
        )}
      </div>

      <div className="px-3 py-2 border-t border-slate-800 text-[11px] text-slate-500 leading-relaxed">
        Drag a device onto the plan, or click the crosshair then click the desired location.
      </div>
    </div>
  )
}
