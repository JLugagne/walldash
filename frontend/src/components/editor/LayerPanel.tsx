import { useState } from 'react'
import { Check, Eye, EyeOff, Layers, Pencil, Plus, Trash2, X } from 'lucide-react'
import type { DevicePlacement, Level } from '../../types'
import { apiFetch } from '../../api'

interface LayerPanelProps {
  level: Level | null
  placements: DevicePlacement[]
  activeLayer: string
  onSelectLayer: (layer: string) => void
  onRefreshLevels: () => Promise<void>
  onRefreshPlacements: () => Promise<void>
}

export function LayerPanel({ level, placements, activeLayer, onSelectLayer, onRefreshLevels, onRefreshPlacements }: LayerPanelProps) {
  const [newName, setNewName] = useState('')
  const [editingLayer, setEditingLayer] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  if (!level) {
    return (
      <div className="flex-1 flex items-center justify-center p-6">
        <div className="text-center space-y-3">
          <Layers className="w-8 h-8 text-slate-600 mx-auto" />
          <p className="text-xs text-slate-500">Select a level to manage its layers.</p>
        </div>
      </div>
    )
  }

  const layers = level.layers && level.layers.length > 0 ? level.layers : ['controls', 'sensors']

  const countForLayer = (layer: string) => placements.filter((p) => (p.layer || 'controls') === layer).length

  const updateLevelLayers = async (newLayers: string[]) => {
    if (!level) return
    setBusy(true)
    setError(null)
    try {
      const res = await apiFetch(`/api/levels/${level.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: level.name, is_outdoor: level.is_outdoor, layers: newLayers }),
      })
      const payload = await res.json().catch(() => ({}))
      if (!res.ok || (payload.status && payload.status !== 'success')) {
        setError(payload.error?.message || 'Failed to update layers')
        return
      }
      await onRefreshLevels()
      await onRefreshPlacements()
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusy(false)
    }
  }

  const handleAddLayer = () => {
    const trimmed = newName.trim().toLowerCase()
    if (!trimmed) return
    if (layers.includes(trimmed)) {
      setError(`Layer "${trimmed}" already exists`)
      return
    }
    updateLevelLayers([...layers, trimmed])
    setNewName('')
  }

  const handleDeleteLayer = (layer: string) => {
    if (layer === 'controls') return
    const updated = layers.filter((l) => l !== layer)
    updateLevelLayers(updated)
  }

  const handleRenameLayer = async (oldName: string) => {
    const trimmed = editValue.trim().toLowerCase()
    setEditingLayer(null)
    if (!trimmed || trimmed === oldName) return
    if (layers.includes(trimmed)) {
      setError(`Layer "${trimmed}" already exists`)
      return
    }

    if (!level) return
    setBusy(true)
    setError(null)
    try {
      const updatedLayers = layers.map((l) => (l === oldName ? trimmed : l))
      const res = await apiFetch(`/api/levels/${level.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: level.name, is_outdoor: level.is_outdoor, layers: updatedLayers }),
      })
      const payload = await res.json().catch(() => ({}))
      if (!res.ok || (payload.status && payload.status !== 'success')) {
        setError(payload.error?.message || 'Failed to rename layer')
        return
      }

      const affected = placements.filter((p) => (p.layer || 'controls') === oldName)
      for (const p of affected) {
        await apiFetch(`/api/levels/${level.id}/placements`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: p.id,
            device_id: p.device_id,
            x: p.x,
            y: p.y,
            icon: p.icon,
            custom_name: p.custom_name,
            render_domain: p.render_domain,
            layer: trimmed,
          }),
        })
      }

      await onRefreshLevels()
      await onRefreshPlacements()
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusy(false)
    }
  }

  const startRename = (layer: string) => {
    setEditingLayer(layer)
    setEditValue(layer)
  }

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="px-3 py-2 border-b border-slate-800 flex items-center justify-between">
        <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Layers</span>
        <span className="text-[10px] font-mono text-slate-600">{layers.length} layer{layers.length > 1 ? 's' : ''}</span>
      </div>

      {error && (
        <div className="bg-rose-500/10 border-b border-rose-500/20 px-3 py-1.5 text-[11px] text-rose-300 flex items-center gap-2">
          <span className="flex-1">{error}</span>
          <button type="button" onClick={() => setError(null)} className="text-rose-300 hover:text-white cursor-pointer">
            <X className="w-3 h-3" />
          </button>
        </div>
      )}

      <div className="flex-1 overflow-y-auto">
        {layers.map((layer, idx) => {
          const isActive = layer === activeLayer
          const count = countForLayer(layer)
          const isEditing = editingLayer === layer

          return (
            <div
              key={layer}
              onClick={() => {
                if (!isEditing) onSelectLayer(layer)
              }}
              className={`group flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-slate-800/50 transition-colors ${
                isActive
                  ? 'bg-indigo-600/10 border-l-2 border-l-indigo-500'
                  : 'border-l-2 border-l-transparent hover:bg-slate-800/50'
              }`}
            >
              <div className={`shrink-0 ${isActive ? 'text-indigo-400' : 'text-slate-600 group-hover:text-slate-400'}`}>
                {isActive ? <Eye className="w-3.5 h-3.5" /> : <EyeOff className="w-3.5 h-3.5" />}
              </div>

              {idx < 9 && (
                <span className="text-[9px] font-mono text-slate-600 w-3 text-right shrink-0">{idx + 1}</span>
              )}

              <div className="flex-1 min-w-0">
                {isEditing ? (
                  <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="text"
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') handleRenameLayer(layer)
                        if (e.key === 'Escape') setEditingLayer(null)
                      }}
                      autoFocus
                      className="flex-1 min-w-0 h-7 bg-slate-950 border border-indigo-500 rounded-md px-2 text-xs text-white focus:outline-none"
                    />
                    <button
                      type="button"
                      onClick={() => handleRenameLayer(layer)}
                      className="shrink-0 w-6 h-6 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white flex items-center justify-center cursor-pointer"
                    >
                      <Check className="w-3 h-3" />
                    </button>
                    <button
                      type="button"
                      onClick={() => setEditingLayer(null)}
                      className="shrink-0 w-6 h-6 rounded-md hover:bg-slate-800 text-slate-400 hover:text-white flex items-center justify-center cursor-pointer"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                ) : (
                  <div className="flex items-center gap-1.5">
                    <span className={`text-xs font-medium truncate ${isActive ? 'text-indigo-200' : 'text-slate-300'}`}>
                      {layer}
                    </span>
                    {layer === 'controls' && (
                      <span className="text-[9px] font-mono text-slate-600 bg-slate-800/50 px-1 rounded">default</span>
                    )}
                  </div>
                )}
              </div>

              <span className="text-[10px] font-mono text-slate-600 shrink-0">{count}</span>

              {!isEditing && (
                <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation()
                      startRename(layer)
                    }}
                    title="Rename layer"
                    className="w-6 h-6 rounded-md hover:bg-slate-800 text-slate-500 hover:text-white flex items-center justify-center cursor-pointer"
                  >
                    <Pencil className="w-3 h-3" />
                  </button>
                  {layer !== 'controls' && (
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteLayer(layer)
                      }}
                      title="Delete layer"
                      className="w-6 h-6 rounded-md hover:bg-rose-900/50 text-slate-500 hover:text-rose-400 flex items-center justify-center cursor-pointer"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>

      <div className="p-3 border-t border-slate-800">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleAddLayer()
          }}
          className="flex items-center gap-1.5"
        >
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="New layer name…"
            disabled={busy}
            className="flex-1 min-w-0 h-8 bg-slate-950 border border-slate-800 rounded-lg px-2.5 text-xs text-white placeholder:text-slate-600 focus:outline-none focus:border-indigo-500"
          />
          <button
            type="submit"
            disabled={busy || !newName.trim()}
            className="h-8 px-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white text-xs font-semibold flex items-center gap-1 cursor-pointer shrink-0 disabled:cursor-default"
          >
            <Plus className="w-3.5 h-3.5" />
          </button>
        </form>
      </div>
    </div>
  )
}