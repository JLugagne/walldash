import { useState } from 'react'
import { ArrowDown, ArrowUp, Check, Home, Layers, Pencil, Plus, Trash2, Trees, Upload, X, Gauge, SlidersHorizontal } from 'lucide-react'
import type { Level } from '../types'
import { apiFetch } from '../api'

interface LevelsManagerProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (id: string) => void
  onRefreshLevels: () => Promise<void>
  onOpenWizard?: () => void
}

export function LevelsManager({ levels, activeLevelId, onSelectLevel, onRefreshLevels, onOpenWizard }: LevelsManagerProps) {
  const [name, setName] = useState('')
  const [isOutdoor, setIsOutdoor] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [newLayer, setNewLayer] = useState('')

  const sorted = [...levels].sort((a, b) => a.order - b.order)
  const activeLevel = levels.find((l) => l.id === activeLevelId)
  const activeLayers = activeLevel?.layers && activeLevel.layers.length > 0 ? activeLevel.layers : [
    { name: 'controls', hide_gauges: false },
    { name: 'sensors', hide_gauges: false },
  ]

  const request = async (input: string, init: RequestInit, failure: string) => {
    setBusy(true)
    setError(null)
    try {
      const res = await apiFetch(input, init)
      const result = res.status === 204 ? { status: 'success' } : await res.json().catch(() => ({}))
      if (!res.ok || (result.status && result.status !== 'success')) {
        setError(result.error?.message || failure)
        return null
      }
      await onRefreshLevels()
      return result
    } catch {
      setError('Unable to reach the server')
      return null
    } finally {
      setBusy(false)
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    const result = await request(
      '/api/levels',
      { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: name.trim(), is_outdoor: isOutdoor }) },
      'Error while creating the level'
    )
    if (result) {
      setName('')
      setIsOutdoor(false)
      if (result.data?.id) onSelectLevel(result.data.id)
    }
  }

  const handleRename = async (lvl: Level) => {
    const trimmed = editName.trim()
    setEditingId(null)
    if (!trimmed || trimmed === lvl.name) return
    await request(
      `/api/levels/${lvl.id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: trimmed, is_outdoor: lvl.is_outdoor, layers: lvl.layers }),
      },
      'Error while renaming the level'
    )
  }

  const handleToggleOutdoor = async (lvl: Level) => {
    await request(
      `/api/levels/${lvl.id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: lvl.name, is_outdoor: !lvl.is_outdoor, layers: lvl.layers }),
      },
      'Error while updating the level'
    )
  }

  const handleAddLayer = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!activeLevel) return
    const trimmed = newLayer.trim().toLowerCase()
    if (!trimmed) return
    if (activeLayers.some((l) => l.name === trimmed)) {
      setError(`Layer "${trimmed}" already exists`)
      return
    }
    const updatedLayers = [...activeLayers, { name: trimmed, hide_gauges: false }]
    const result = await request(
      `/api/levels/${activeLevel.id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: activeLevel.name, is_outdoor: activeLevel.is_outdoor, layers: updatedLayers }),
      },
      'Error while adding the layer'
    )
    if (result) {
      setNewLayer('')
    }
  }

  const handleToggleHideGauges = async (layerName: string) => {
    if (!activeLevel) return
    const updatedLayers = activeLayers.map((l) =>
      l.name === layerName ? { ...l, hide_gauges: !l.hide_gauges } : l
    )
    await request(
      `/api/levels/${activeLevel.id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: activeLevel.name, is_outdoor: activeLevel.is_outdoor, layers: updatedLayers }),
      },
      'Error while updating layer'
    )
  }

  const handleDeleteLayer = async (layerToDelete: string) => {
    if (!activeLevel) return
    if (layerToDelete === 'controls') {
      setError('The "controls" layer is the default layer and cannot be deleted.')
      return
    }
    if (!confirm(`Delete layer "${layerToDelete}"? Associated devices will be reassigned to "controls".`)) return
    const updatedLayers = activeLayers.filter((l) => l.name !== layerToDelete)
    await request(
      `/api/levels/${activeLevel.id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: activeLevel.name, is_outdoor: activeLevel.is_outdoor, layers: updatedLayers }),
      },
      'Error while deleting the layer'
    )
  }

  const handleDelete = async (lvl: Level) => {
    if (!confirm(`Delete "${lvl.name}" along with its plan and placed devices?`)) return
    // If the deleted level is selected, move the selection to a sibling
    // *before* the refresh lands so the editor never points at a ghost id
    // (which would fetch /plan + /placements for a missing level, then get
    // bounced back by the URL sync).
    const fallback = sorted.find((l) => l.id !== lvl.id) ?? null
    if (fallback && lvl.id === activeLevelId) onSelectLevel(fallback.id)
    const result = await request(`/api/levels/${lvl.id}`, { method: 'DELETE' }, 'Error while deleting the level')
    if (result && lvl.id === activeLevelId && fallback) onSelectLevel(fallback.id)
  }

  const handleMove = async (index: number, direction: -1 | 1) => {
    const target = index + direction
    if (target < 0 || target >= sorted.length) return
    const order = [...sorted]
    ;[order[index], order[target]] = [order[target], order[index]]
    await request(
      '/api/levels/reorder',
      { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ level_ids: order.map((l) => l.id) }) },
      'Error while reordering'
    )
  }

  return (
    <div className="flex flex-col min-h-0">
      <div className="px-4 py-3 border-b border-slate-800 flex items-center gap-2">
        <Layers className="w-4 h-4 text-indigo-400" />
        <h2 className="text-sm font-semibold text-white flex-1">Levels</h2>
        <span className="text-[11px] text-slate-500">{levels.length} level{levels.length > 1 ? 's' : ''}</span>
        {onOpenWizard && (
          <button
            type="button"
            onClick={onOpenWizard}
            disabled={busy}
            title="Import levels or restore a backup"
            className="h-7 px-2 rounded-lg flex items-center gap-1 text-[11px] text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 border border-slate-700 disabled:opacity-50 cursor-pointer"
          >
            <Upload className="w-3.5 h-3.5" />
            Import
          </button>
        )}
      </div>

      {error && (
        <div className="mx-3 mt-3 p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center gap-2">
          <span className="flex-1">{error}</span>
          <button type="button" onClick={() => setError(null)} className="hover:text-white cursor-pointer">
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      <div className="flex-1 min-h-0 overflow-y-auto p-2 space-y-1 max-h-[40vh]">
        {sorted.length === 0 && <div className="text-center py-6 text-slate-500 text-xs">No levels. Create the first one below.</div>}
        {sorted.map((lvl, index) => {
          const active = lvl.id === activeLevelId
          const Icon = lvl.is_outdoor ? Trees : Home
          const editing = editingId === lvl.id
          return (
            <div
              key={lvl.id}
              className={`group flex items-center gap-2 px-2 py-1.5 rounded-lg border transition-colors ${
                active ? 'bg-indigo-950/40 border-indigo-500/50' : 'border-transparent hover:bg-slate-800/60'
              }`}
            >
              <button
                type="button"
                onClick={() => handleToggleOutdoor(lvl)}
                disabled={busy}
                title={lvl.is_outdoor ? 'Outdoor (click to switch to indoor)' : 'Indoor (click to switch to outdoor)'}
                className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 cursor-pointer ${
                  lvl.is_outdoor ? 'bg-emerald-500/15 text-emerald-400' : 'bg-indigo-500/15 text-indigo-400'
                }`}
              >
                <Icon className="w-4 h-4" />
              </button>

              {editing ? (
                <input
                  autoFocus
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  onBlur={() => handleRename(lvl)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleRename(lvl)
                    if (e.key === 'Escape') setEditingId(null)
                  }}
                  className="flex-1 min-w-0 h-8 bg-slate-950 border border-indigo-500 rounded-lg px-2 text-xs text-white focus:outline-none"
                />
              ) : (
                <button
                  type="button"
                  onClick={() => onSelectLevel(lvl.id)}
                  className="flex-1 min-w-0 text-left cursor-pointer"
                >
                  <div className="text-xs font-semibold text-slate-200 truncate flex items-center gap-1.5">
                    {lvl.name}
                    {active && <Check className="w-3 h-3 text-indigo-400" />}
                  </div>
                  <div className="text-[10px] text-slate-500">{lvl.is_outdoor ? 'Outdoor' : 'Indoor'}</div>
                </button>
              )}

              <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                <SmallButton title="Rename" onClick={() => {
                  setEditingId(lvl.id)
                  setEditName(lvl.name)
                }}>
                  <Pencil className="w-3 h-3" />
                </SmallButton>
                <SmallButton title="Move up" disabled={index === 0 || busy} onClick={() => handleMove(index, -1)}>
                  <ArrowUp className="w-3 h-3" />
                </SmallButton>
                <SmallButton title="Move down" disabled={index === sorted.length - 1 || busy} onClick={() => handleMove(index, 1)}>
                  <ArrowDown className="w-3 h-3" />
                </SmallButton>
                <SmallButton title="Delete" danger disabled={busy} onClick={() => handleDelete(lvl)}>
                  <Trash2 className="w-3 h-3" />
                </SmallButton>
              </div>
            </div>
          )
        })}
      </div>

      {activeLevel && (
        <div className="p-3 border-t border-slate-800 space-y-2 bg-slate-950/20">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <Layers className="w-3.5 h-3.5 text-indigo-400" />
              Display Layers · {activeLevel.name}
            </span>
            <span className="text-[10px] text-slate-500 font-mono">{activeLayers.length}</span>
          </div>

          <div className="flex flex-wrap gap-1.5 py-1">
            {activeLayers.map((layer) => (
              <span
                key={layer.name}
                className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-slate-800/90 border border-slate-700/60 text-xs text-slate-200"
              >
                <span>{layer.name}</span>
                <button
                  type="button"
                  onClick={() => handleToggleHideGauges(layer.name)}
                  title={layer.hide_gauges ? 'Show gauges for zones in this layer' : 'Hide gauges for zones in this layer'}
                  className={`w-5 h-5 rounded flex items-center justify-center cursor-pointer ml-0.5 transition-colors ${
                    layer.hide_gauges
                      ? 'bg-amber-900/30 text-amber-400 hover:bg-amber-900/50'
                      : 'hover:bg-slate-700 text-slate-400 hover:text-white'
                  }`}
                >
                  {layer.hide_gauges ? <SlidersHorizontal className="w-3 h-3" /> : <Gauge className="w-3 h-3" />}
                </button>
                {layer.name === 'controls' ? (
                  <span className="text-[9px] font-mono text-indigo-400 uppercase tracking-wider">(default)</span>
                ) : (
                  <button
                    type="button"
                    onClick={() => handleDeleteLayer(layer.name)}
                    disabled={busy}
                    title={`Delete layer "${layer.name}" (reassigns to controls)`}
                    className="text-slate-400 hover:text-rose-400 cursor-pointer ml-0.5"
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </span>
            ))}
          </div>

          <form onSubmit={handleAddLayer} className="flex items-center gap-1.5 pt-0.5">
            <input
              type="text"
              placeholder="New layer (e.g. lights, hvac…)"
              value={newLayer}
              onChange={(e) => setNewLayer(e.target.value)}
              className="flex-1 min-w-0 h-8 bg-slate-950 border border-slate-800 rounded-lg px-2.5 text-xs text-white placeholder:text-slate-600 focus:outline-none focus:border-indigo-500"
            />
            <button
              type="submit"
              disabled={!newLayer.trim() || busy}
              className="h-8 px-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white text-xs font-semibold flex items-center gap-1 cursor-pointer shrink-0"
            >
              <Plus className="w-3.5 h-3.5" />
              Add
            </button>
          </form>
        </div>
      )}

      <form onSubmit={handleCreate} className="p-3 border-t border-slate-800 space-y-2 bg-slate-950/40">
        <div className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">New level</div>
        <div className="flex items-center gap-1.5">
          <input
            type="text"
            placeholder="Ground floor, Garden…"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="flex-1 min-w-0 h-9 bg-slate-950 border border-slate-800 rounded-lg px-3 text-xs text-white placeholder:text-slate-600 focus:outline-none focus:border-indigo-500"
          />
          <button
            type="button"
            onClick={() => setIsOutdoor((v) => !v)}
            title={isOutdoor ? 'Outdoor space' : 'Indoor space'}
            className={`w-9 h-9 rounded-lg flex items-center justify-center shrink-0 transition-colors cursor-pointer ${
              isOutdoor ? 'bg-emerald-500/15 text-emerald-400' : 'bg-slate-800 text-slate-400 hover:text-white'
            }`}
          >
            {isOutdoor ? <Trees className="w-4 h-4" /> : <Home className="w-4 h-4" />}
          </button>
          <button
            type="submit"
            disabled={!name.trim() || busy}
            className="h-9 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white text-xs font-semibold flex items-center gap-1 cursor-pointer"
          >
            <Plus className="w-3.5 h-3.5" />
            Create
          </button>
        </div>
      </form>
    </div>
  )
}

function SmallButton({
  children,
  onClick,
  title,
  disabled,
  danger,
}: {
  children: React.ReactNode
  onClick: () => void
  title: string
  disabled?: boolean
  danger?: boolean
}) {
  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      title={title}
      disabled={disabled}
      className={`w-6 h-6 rounded flex items-center justify-center disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer ${
        danger ? 'text-slate-500 hover:text-rose-400 hover:bg-rose-500/10' : 'text-slate-400 hover:text-white hover:bg-slate-700'
      }`}
    >
      {children}
    </button>
  )
}