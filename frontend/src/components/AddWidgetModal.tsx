import React, { useState } from 'react'
import { X, Check, Search, Zap } from 'lucide-react'
import type { Automation } from '../types'

interface AddWidgetModalProps {
  isOpen: boolean
  automations: Automation[]
  onClose: () => void
  onAddWidget: (title: string, selectedEntityIds: string[]) => Promise<void>
}

export const AddWidgetModal: React.FC<AddWidgetModalProps> = ({
  isOpen,
  automations,
  onClose,
  onAddWidget,
}) => {
  const [title, setTitle] = useState('Automatisations Favorites')
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [searchTerm, setSearchTerm] = useState('')
  const [saving, setSaving] = useState(false)

  if (!isOpen) return null

  const filteredAutomations = automations.filter(
    (a) =>
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.id.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const selectAll = () => {
    setSelectedIds(new Set(automations.map((a) => a.id)))
  }

  const selectNone = () => {
    setSelectedIds(new Set())
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || saving) return

    setSaving(true)
    try {
      await onAddWidget(title.trim(), Array.from(selectedIds))
      onClose()
    } catch (err) {
      console.error('Failed to add widget:', err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center space-x-2">
            <div className="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <Zap className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">Ajouter un Widget Automatisations</h2>
              <p className="text-xs text-slate-400">
                Sélectionnez les scénarios Home Assistant à afficher dans ce widget.
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 text-slate-400 hover:text-white rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="flex-1 flex flex-col overflow-hidden">
          <div className="p-6 space-y-4 flex-1 overflow-y-auto">
            {/* Title Input */}
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">
                Titre du Widget
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="ex: Scénarios du Salon"
                required
                className="w-full bg-slate-800/80 border border-slate-700 rounded-xl px-3.5 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
              />
            </div>

            {/* Selection Controls */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="text-xs font-semibold text-slate-300">
                  Automatisations ({selectedIds.size}/{automations.length} sélectionnée{selectedIds.size > 1 ? 's' : ''})
                </label>
                <div className="space-x-3 text-xs">
                  <button
                    type="button"
                    onClick={selectAll}
                    className="text-indigo-400 hover:text-indigo-300 font-medium"
                  >
                    Tout cocher
                  </button>
                  <span className="text-slate-600">•</span>
                  <button
                    type="button"
                    onClick={selectNone}
                    className="text-slate-400 hover:text-slate-300"
                  >
                    Décocher tout
                  </button>
                </div>
              </div>

              {/* Search Bar */}
              <div className="relative mb-3">
                <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Rechercher une automatisation..."
                  className="w-full bg-slate-800/50 border border-slate-700/80 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                />
              </div>

              {/* Automations Checklist */}
              <div className="space-y-2 max-h-60 overflow-y-auto border border-slate-800 rounded-xl p-2 bg-slate-950/40">
                {filteredAutomations.length === 0 ? (
                  <div className="text-center py-6 text-xs text-slate-500">
                    Aucune automatisation trouvée
                  </div>
                ) : (
                  filteredAutomations.map((auto) => {
                    const isSelected = selectedIds.has(auto.id)
                    return (
                      <div
                        key={auto.id}
                        onClick={() => toggleSelect(auto.id)}
                        className={`p-3 rounded-lg border cursor-pointer transition-all flex items-center justify-between select-none ${
                          isSelected
                            ? 'bg-indigo-600/15 border-indigo-500/40 text-white'
                            : 'bg-slate-900/60 border-slate-800/70 text-slate-300 hover:border-slate-700'
                        }`}
                      >
                        <div className="min-w-0 flex-1 pr-3">
                          <p className="text-xs font-semibold truncate">{auto.name}</p>
                          <p className="text-[10px] text-slate-500 font-mono truncate">{auto.id}</p>
                        </div>
                        <div
                          className={`w-5 h-5 rounded-md border flex items-center justify-center transition-colors ${
                            isSelected
                              ? 'bg-indigo-600 border-indigo-500 text-white shadow-sm'
                              : 'border-slate-700 bg-slate-800'
                          }`}
                        >
                          {isSelected && <Check className="w-3.5 h-3.5" />}
                        </div>
                      </div>
                    )
                  })
                )}
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-slate-800 bg-slate-900/80 flex items-center justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white rounded-xl transition-colors"
            >
              Annuler
            </button>
            <button
              type="submit"
              disabled={saving || !title.trim()}
              className="px-5 py-2 text-xs font-bold bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white rounded-xl shadow-lg shadow-indigo-600/30 transition-all active:scale-95 disabled:opacity-50"
            >
              {saving ? 'Création...' : 'Créer le Widget'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
