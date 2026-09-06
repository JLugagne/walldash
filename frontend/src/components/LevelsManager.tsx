import { useState } from 'react'
import { Plus, Trash2, ArrowUp, ArrowDown, Trees, Home, Layers, Check } from 'lucide-react'
import type { Level } from '../types'
import { apiFetch } from '../api'

interface LevelsManagerProps {
  levels: Level[]
  activeLevelId: string | null
  onSelectLevel: (id: string) => void
  onRefreshLevels: () => Promise<void>
}

export function LevelsManager({
  levels,
  activeLevelId,
  onSelectLevel,
  onRefreshLevels,
}: LevelsManagerProps) {
  const [name, setName] = useState('')
  const [isOutdoor, setIsOutdoor] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleCreateLevel = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    setIsSubmitting(true)
    setError(null)
    try {
      const res = await apiFetch('/api/levels', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: name.trim(),
          is_outdoor: isOutdoor,
        }),
      })
      const result = await res.json()
      if (res.ok && result.status === 'success') {
        setName('')
        setIsOutdoor(false)
        await onRefreshLevels()
        if (result.data?.id) {
          onSelectLevel(result.data.id)
        }
      } else {
        setError(result.error?.message || 'Erreur lors de la création du niveau')
      }
    } catch {
      setError('Impossible de joindre le serveur')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteLevel = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm('Voulez-vous supprimer ce niveau et son plan 2D ?')) return

    try {
      const res = await apiFetch(`/api/levels/${id}`, { method: 'DELETE' })
      if (res.ok) {
        await onRefreshLevels()
      }
    } catch (err) {
      console.error('Delete level failed:', err)
    }
  }

  const handleMove = async (index: number, direction: 'up' | 'down', e: React.MouseEvent) => {
    e.stopPropagation()
    const targetIndex = direction === 'up' ? index - 1 : index + 1
    if (targetIndex < 0 || targetIndex >= levels.length) return

    const newOrder = [...levels]
    const temp = newOrder[index]
    newOrder[index] = newOrder[targetIndex]
    newOrder[targetIndex] = temp

    const levelIds = newOrder.map((l) => l.id)

    try {
      const res = await apiFetch('/api/levels/reorder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ level_ids: levelIds }),
      })
      if (res.ok) {
        await onRefreshLevels()
      }
    } catch (err) {
      console.error('Reorder levels failed:', err)
    }
  }

  return (
    <div className="flex flex-col h-full space-y-4">
      <div className="flex items-center justify-between border-b border-slate-800 pb-3">
        <div className="flex items-center space-x-2">
          <Layers className="w-5 h-5 text-indigo-400" />
          <h2 className="text-base font-bold text-white tracking-tight">Niveaux & Étages</h2>
        </div>
        <span className="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full">
          {levels.length} {levels.length > 1 ? 'niveaux' : 'niveau'}
        </span>
      </div>

      {error && (
        <div className="p-2.5 rounded bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs">
          {error}
        </div>
      )}

      {/* Levels list */}
      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {levels.length === 0 ? (
          <div className="text-center py-8 text-slate-500 text-xs">
            Aucun niveau configuré. Créez un premier niveau ci-dessous.
          </div>
        ) : (
          levels.map((lvl, index) => {
            const isActive = lvl.id === activeLevelId
            return (
              <div
                key={lvl.id}
                onClick={() => onSelectLevel(lvl.id)}
                className={`group p-3 rounded-lg border transition-all cursor-pointer flex items-center justify-between ${
                  isActive
                    ? 'bg-indigo-950/40 border-indigo-500/80 shadow-md shadow-indigo-500/10'
                    : 'bg-slate-900/60 border-slate-800 hover:border-slate-700 hover:bg-slate-800/40'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <div
                    className={`p-1.5 rounded-md ${
                      lvl.is_outdoor
                        ? 'bg-emerald-500/20 text-emerald-400'
                        : 'bg-indigo-500/20 text-indigo-400'
                    }`}
                  >
                    {lvl.is_outdoor ? <Trees className="w-4 h-4" /> : <Home className="w-4 h-4" />}
                  </div>
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-semibold text-slate-200 group-hover:text-white">
                        {lvl.name}
                      </span>
                      {isActive && (
                        <span className="flex items-center text-[10px] text-indigo-400 bg-indigo-500/10 px-1.5 py-0.2 rounded border border-indigo-500/20">
                          <Check className="w-3 h-3 mr-0.5" /> Actif
                        </span>
                      )}
                    </div>
                    <span className="text-[11px] text-slate-400">
                      {lvl.is_outdoor ? 'Extérieur' : 'Intérieur'} • Ordre: {lvl.order}
                    </span>
                  </div>
                </div>

                <div className="flex items-center space-x-1">
                  <button
                    type="button"
                    disabled={index === 0}
                    onClick={(e) => handleMove(index, 'up', e)}
                    className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Monter"
                  >
                    <ArrowUp className="w-3.5 h-3.5" />
                  </button>
                  <button
                    type="button"
                    disabled={index === levels.length - 1}
                    onClick={(e) => handleMove(index, 'down', e)}
                    className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Descendre"
                  >
                    <ArrowDown className="w-3.5 h-3.5" />
                  </button>
                  <button
                    type="button"
                    onClick={(e) => handleDeleteLevel(lvl.id, e)}
                    className="p-1 rounded text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors ml-1"
                    title="Supprimer ce niveau"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )
          })
        )}
      </div>

      {/* Add level form */}
      <form onSubmit={handleCreateLevel} className="pt-3 border-t border-slate-800 space-y-3">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
          Nouveau Niveau
        </h3>
        <div>
          <input
            type="text"
            placeholder="Nom (ex: Rez-de-chaussée, Jardin...)"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full bg-slate-900 border border-slate-700 rounded-md px-3 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <label className="flex items-center space-x-2 text-xs text-slate-300 cursor-pointer">
          <input
            type="checkbox"
            checked={isOutdoor}
            onChange={(e) => setIsOutdoor(e.target.checked)}
            className="rounded bg-slate-900 border-slate-700 text-indigo-600 focus:ring-indigo-500 h-3.5 w-3.5"
          />
          <span>Espace extérieur (Jardin, Terrasse, Entrée)</span>
        </label>

        <button
          type="submit"
          disabled={!name.trim() || isSubmitting}
          className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded-md text-xs flex items-center justify-center space-x-2 shadow-md shadow-indigo-600/20 transition-all"
        >
          <Plus className="w-4 h-4" />
          <span>{isSubmitting ? 'Création...' : 'Ajouter le Niveau'}</span>
        </button>
      </form>
    </div>
  )
}
