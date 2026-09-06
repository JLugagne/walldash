import { useRef, useState } from 'react'
import { AlertCircle, Check, Copy, FileJson, Sparkles, Upload, X } from 'lucide-react'
import type { Plan } from '../types'
import { parseImportedPlan, PlanImportError, PLAN_IMPORT_PROMPT } from '../planImport'

interface ImportPlanModalProps {
  isOpen: boolean
  levelId: string
  onClose: () => void
  onImport: (plan: Plan) => void
}

export function ImportPlanModal({ isOpen, levelId, onClose, onImport }: ImportPlanModalProps) {
  const [jsonText, setJsonText] = useState('')
  const [issues, setIssues] = useState<string[]>([])
  const [promptCopied, setPromptCopied] = useState(false)
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  if (!isOpen) return null

  const handleClose = () => {
    setJsonText('')
    setIssues([])
    onClose()
  }

  const handleCopyPrompt = async () => {
    try {
      await navigator.clipboard.writeText(PLAN_IMPORT_PROMPT)
      setPromptCopied(true)
      setTimeout(() => setPromptCopied(false), 2000)
    } catch {
      setIssues(['Impossible de copier le prompt dans le presse-papiers.'])
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      setJsonText(typeof reader.result === 'string' ? reader.result : '')
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  const handleImport = () => {
    setIssues([])
    if (!jsonText.trim()) {
      setIssues(['Collez ou importez un JSON avant de continuer.'])
      return
    }
    try {
      const plan = parseImportedPlan(jsonText, levelId)
      onImport(plan)
      handleClose()
    } catch (err) {
      if (err instanceof PlanImportError) {
        setIssues(err.issues)
      } else {
        setIssues(['Erreur inattendue lors de la lecture du JSON.'])
      }
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center space-x-2">
            <div className="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <FileJson className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">Importer un plan depuis un JSON</h2>
              <p className="text-xs text-slate-400">
                Générez le JSON avec une IA à partir d'une photo de votre plan, puis collez-le ici.
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={handleClose}
            className="p-1.5 text-slate-400 hover:text-white rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-6 space-y-4 flex-1 overflow-y-auto">
          {/* Step 1: copy prompt */}
          <div className="bg-indigo-950/40 border border-indigo-500/30 rounded-xl p-4 space-y-2">
            <div className="flex items-center space-x-2 text-indigo-300 font-semibold text-xs">
              <Sparkles className="w-4 h-4" />
              <span>1. Copiez le prompt et envoyez-le à une IA avec une photo de votre plan</span>
            </div>
            <p className="text-xs text-slate-400">
              Le prompt (en anglais) explique à l'IA le format JSON attendu : murs, portes, fenêtres, pièces et
              dimensions réelles (mètres / centimètres).
            </p>
            <button
              type="button"
              onClick={handleCopyPrompt}
              className="flex items-center space-x-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium px-3 py-1.5 rounded-md transition-all"
            >
              {promptCopied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{promptCopied ? 'Prompt copié !' : 'Copier le prompt pour l\'IA'}</span>
            </button>
          </div>

          {/* Step 2: paste / upload JSON */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs font-semibold text-slate-300">
                2. Collez la réponse JSON de l'IA
              </label>
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="flex items-center space-x-1.5 text-xs text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 px-2.5 py-1 rounded-md border border-slate-700"
              >
                <Upload className="w-3.5 h-3.5" />
                <span>Importer un fichier .json</span>
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept=".json,application/json"
                className="hidden"
                onChange={handleFileChange}
              />
            </div>
            <textarea
              value={jsonText}
              onChange={(e) => setJsonText(e.target.value)}
              placeholder='{"version": 1, "unit": "m", "walls": [...], "zones": [...]}'
              spellCheck={false}
              className="w-full h-56 bg-slate-950/60 border border-slate-700 rounded-xl px-3.5 py-2.5 text-xs font-mono text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors resize-none"
            />
          </div>

          {/* Errors */}
          {issues.length > 0 && (
            <div className="bg-rose-950/40 border border-rose-500/30 rounded-xl p-3 space-y-1">
              <div className="flex items-center space-x-1.5 text-rose-400 text-xs font-semibold">
                <AlertCircle className="w-3.5 h-3.5" />
                <span>Le JSON contient des erreurs :</span>
              </div>
              <ul className="text-xs text-rose-300 list-disc list-inside space-y-0.5">
                {issues.map((issue, i) => (
                  <li key={i}>{issue}</li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-slate-800 bg-slate-900/80 flex items-center justify-between">
          <p className="text-[11px] text-slate-500">
            L'import remplace les murs, ouvertures et zones actuels de ce niveau (sauvegardez avant si besoin).
          </p>
          <div className="flex items-center space-x-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white rounded-xl transition-colors"
            >
              Annuler
            </button>
            <button
              type="button"
              onClick={handleImport}
              disabled={!jsonText.trim()}
              className="px-5 py-2 text-xs font-bold bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white rounded-xl shadow-lg shadow-indigo-600/30 transition-all active:scale-95 disabled:opacity-50"
            >
              Importer le plan
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
