import { useRef, useState } from 'react'
import { AlertCircle, Check, Copy, FileArchive, Sparkles, Upload, X } from 'lucide-react'
import type { Plan } from '../types'
import { PLAN_IMPORT_PROMPT } from '../planImport'
import { apiFetch } from '../api'

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
  const [importing, setImporting] = useState(false)
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
      setIssues(['Unable to copy the prompt to the clipboard.'])
    }
  }

  const extractError = (payload: any): string => {
    if (!payload) return 'Import failed'
    if (typeof payload === 'string') return payload
    if (payload.message && typeof payload.message === 'string') return payload.message
    if (payload.error) {
      if (typeof payload.error === 'string') return payload.error
      if (typeof payload.error.message === 'string') return payload.error.message
    }
    return 'Import failed'
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.name.endsWith('.sh3d')) {
      submitFile(file)
    } else {
      const reader = new FileReader()
      reader.onload = () => {
        setJsonText(typeof reader.result === 'string' ? reader.result : '')
      }
      reader.readAsText(file)
    }
    e.target.value = ''
  }

  const submitFile = async (file: File) => {
    setIssues([])
    setImporting(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const res = await apiFetch(`/api/levels/${levelId}/plan/import`, {
        method: 'POST',
        body: formData,
      })
      if (!res.ok) {
        const payload = await res.json().catch(() => null)
        setIssues([extractError(payload)])
        return
      }
      const payload = await res.json()
      const plan = payload?.data
      if (plan) {
        onImport(plan as Plan)
        handleClose()
      }
    } catch (err) {
      setIssues([(err as Error).message || 'Failed to import .sh3d file'])
    } finally {
      setImporting(false)
    }
  }

  const handleImport = async () => {
    setIssues([])
    if (!jsonText.trim()) {
      setIssues(['Paste a JSON plan or upload a .sh3d file before continuing.'])
      return
    }
    setImporting(true)
    try {
      const res = await apiFetch(`/api/levels/${levelId}/plan/import`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: jsonText,
      })
      if (!res.ok) {
        const payload = await res.json().catch(() => null)
        setIssues([extractError(payload)])
        return
      }
      const payload = await res.json()
      const plan = payload?.data
      if (plan) {
        onImport(plan as Plan)
        handleClose()
      }
    } catch (err) {
      setIssues([(err as Error).message || 'Unexpected error while importing.'])
    } finally {
      setImporting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center space-x-2">
            <div className="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <FileArchive className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white">Import a plan</h2>
              <p className="text-xs text-slate-400">
                Upload a SweetHome3D .sh3d file or paste a plan JSON (AI-generated from a photo).
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
          {/* Step 1: copy prompt (JSON only) */}
          <div className="bg-indigo-950/40 border border-indigo-500/30 rounded-xl p-4 space-y-2">
            <div className="flex items-center space-x-2 text-indigo-300 font-semibold text-xs">
              <Sparkles className="w-4 h-4" />
              <span>1. For JSON: copy the prompt and send it to an AI with a photo of your plan</span>
            </div>
            <p className="text-xs text-slate-400">
              The prompt (in English) explains to the AI the expected JSON format: walls, doors, windows, rooms and
              real dimensions (meters / centimeters).
            </p>
            <button
              type="button"
              onClick={handleCopyPrompt}
              className="flex items-center space-x-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium px-3 py-1.5 rounded-md transition-all"
            >
              {promptCopied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{promptCopied ? 'Prompt copied!' : "Copy prompt for AI"}</span>
            </button>
          </div>

          {/* Step 2: paste JSON or upload file */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs font-semibold text-slate-300">
                2. Paste the JSON or upload a SweetHome3D file
              </label>
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="flex items-center space-x-1.5 text-xs text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 px-2.5 py-1 rounded-md border border-slate-700"
              >
                <Upload className="w-3.5 h-3.5" />
                <span>Upload .json or .sh3d</span>
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept=".json,.sh3d,application/json"
                className="hidden"
                onChange={handleFileChange}
              />
            </div>
            <textarea
              value={jsonText}
              onChange={(e) => setJsonText(e.target.value)}
              placeholder='{"version": 1, "unit": "m", "walls": [...], "zones": [...]}'
              spellCheck={false}
              disabled={importing}
              className="w-full h-56 bg-slate-950/60 border border-slate-700 rounded-xl px-3.5 py-2.5 text-xs font-mono text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors resize-none disabled:opacity-50"
            />
          </div>

          {/* Errors */}
          {issues.length > 0 && (
            <div className="bg-rose-950/40 border border-rose-500/30 rounded-xl p-3 space-y-1">
              <div className="flex items-center space-x-1.5 text-rose-400 text-xs font-semibold">
                <AlertCircle className="w-3.5 h-3.5" />
                <span>Import failed:</span>
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
            The import replaces the current walls, openings and zones for this level (save first if needed).
          </p>
          <div className="flex items-center space-x-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white rounded-xl transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleImport}
              disabled={!jsonText.trim() || importing}
              className="px-5 py-2 text-xs font-bold bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white rounded-xl shadow-lg shadow-indigo-600/30 transition-all active:scale-95 disabled:opacity-50"
            >
              {importing ? 'Importing...' : 'Import plan'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
