import { useRef, useState } from 'react'
import { Archive, ArrowLeft, CheckCircle2, Loader2, Plus, Upload, X } from 'lucide-react'
import type { Level, RestoreSummary } from '../types'
import { apiFetch, readApiError } from '../api'

type WizardStep = 'choose' | 'sh3d' | 'restore' | 'blank' | 'done'

interface OnboardingWizardProps {
  onDone: () => void
  onClose: () => void
}

interface BackupDocument {
  version?: unknown
  levels?: unknown
  dashboards?: unknown
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

export function OnboardingWizard({ onDone, onClose }: OnboardingWizardProps) {
  const [step, setStep] = useState<WizardStep>('choose')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [resultTitle, setResultTitle] = useState('')
  const [resultDetail, setResultDetail] = useState('')
  const [blankName, setBlankName] = useState('')
  const [blankOutdoor, setBlankOutdoor] = useState(false)
  const [includeDevices, setIncludeDevices] = useState(false)
  const sh3dInputRef = useRef<HTMLInputElement>(null)
  const restoreInputRef = useRef<HTMLInputElement>(null)

  const finish = (title: string, detail: string) => {
    setResultTitle(title)
    setResultDetail(detail)
    setStep('done')
  }

  const goStep = (next: WizardStep) => {
    setError(null)
    setStep(next)
  }

  const handleSh3dFile = async (file: File) => {
    setBusy(true)
    setError(null)
    try {
      const form = new FormData()
      form.append('file', file)
      const res = await apiFetch('/api/levels/import/sh3d', { method: 'POST', body: form })
      if (!res.ok) {
        setError(await readApiError(res))
        return
      }
      const payload = await res.json().catch(() => ({}))
      const data: unknown = payload?.data
      if (payload?.status !== 'success' || !Array.isArray(data)) {
        setError('Import failed on the server side')
        return
      }
      const levels = data as Level[]
      finish(
        `${levels.length} level${levels.length > 1 ? 's' : ''} created`,
        levels.map((l) => l.name).join(' · ')
      )
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusy(false)
    }
  }

  const handleRestoreFile = async (file: File) => {
    setBusy(true)
    setError(null)
    try {
      const text = await file.text()
      let doc: BackupDocument
      try {
        doc = JSON.parse(text) as BackupDocument
      } catch {
        setError('This file is not valid JSON')
        return
      }
      if (!isRecord(doc) || (doc.version !== '1' && doc.version !== 1)) {
        setError('Unsupported backup version (expected version "1")')
        return
      }
      const res = await apiFetch('/api/restore', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          version: '1',
          levels: Array.isArray(doc.levels) ? doc.levels : [],
          dashboards: Array.isArray(doc.dashboards) ? doc.dashboards : [],
          include_devices: includeDevices,
        }),
      })
      if (!res.ok) {
        setError(await readApiError(res))
        return
      }
      const payload = await res.json().catch(() => ({}))
      const summary = payload?.data as RestoreSummary | undefined
      if (payload?.status !== 'success' || !summary || typeof summary.levels !== 'number') {
        setError('Restore failed on the server side')
        return
      }
      finish(
        `${summary.levels} level${summary.levels > 1 ? 's' : ''} restored`,
        `${summary.plans} plans · ${summary.placements} devices · ${summary.dashboards} dashboards · ${summary.widgets} widgets`
      )
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusy(false)
    }
  }

  const handleBlank = async (e: React.FormEvent) => {
    e.preventDefault()
    const name = blankName.trim()
    if (!name) {
      setError('Give the level a name')
      return
    }
    setBusy(true)
    setError(null)
    try {
      const res = await apiFetch('/api/levels', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, is_outdoor: blankOutdoor }),
      })
      if (!res.ok) {
        setError(await readApiError(res))
        return
      }
      const payload = await res.json().catch(() => ({}))
      if (payload?.status !== 'success') {
        setError('Creation failed on the server side')
        return
      }
      finish(`Level "${name}" created`, 'You can now draw its plan or import one from the editor.')
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusy(false)
    }
  }

  const stepTitle =
    step === 'choose'
      ? 'Welcome to Walldash'
      : step === 'sh3d'
        ? 'Import from Sweet Home 3D'
        : step === 'restore'
          ? 'Restore a backup'
          : step === 'blank'
            ? 'Create the first level'
            : 'All set'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
      <div className="w-full max-w-lg rounded-2xl border border-slate-700 bg-slate-900 shadow-2xl">
        <div className="flex items-center gap-2 px-5 py-4 border-b border-slate-800">
          {step !== 'choose' && step !== 'done' && (
            <button
              type="button"
              onClick={() => goStep('choose')}
              disabled={busy}
              title="Back"
              className="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-50 cursor-pointer"
            >
              <ArrowLeft className="w-4 h-4" />
            </button>
          )}
          <h2 className="text-sm font-semibold text-white flex-1">{stepTitle}</h2>
          <button
            type="button"
            onClick={onClose}
            title="Close"
            className="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5">
          {step === 'choose' && (
            <div className="space-y-2">
              <p className="text-xs text-slate-400 pb-1">
                No levels yet. Pick how to set up your home — every declared floor is created automatically.
              </p>
              <ChoiceButton
                icon={<Upload className="w-5 h-5 text-indigo-400" />}
                title="Import a Sweet Home 3D file"
                desc="One .sh3d file → all floors, walls, rooms and openings."
                onClick={() => goStep('sh3d')}
              />
              <ChoiceButton
                icon={<Archive className="w-5 h-5 text-emerald-400" />}
                title="Restore a backup"
                desc="Levels, plans, devices and dashboards from an export file."
                onClick={() => goStep('restore')}
              />
              <ChoiceButton
                icon={<Plus className="w-5 h-5 text-amber-400" />}
                title="Start blank"
                desc="Create the first level manually and draw its plan."
                onClick={() => goStep('blank')}
              />
            </div>
          )}

          {step === 'sh3d' && (
            <div className="space-y-3">
              <p className="text-xs text-slate-400">
                Drop your <span className="font-mono">.sh3d</span> file below. Each floor in the file becomes a level.
              </p>
              <DropZone
                accept=".sh3d"
                inputRef={sh3dInputRef}
                disabled={busy}
                label="Drop a .sh3d file here or click to browse"
                onFile={handleSh3dFile}
              />
            </div>
          )}

          {step === 'restore' && (
            <div className="space-y-3">
              <p className="text-xs text-slate-400">
                Drop a backup file previously downloaded from the dashboard export.
              </p>
              <DropZone
                accept=".json,application/json"
                inputRef={restoreInputRef}
                disabled={busy}
                label="Drop a backup .json file here or click to browse"
                onFile={handleRestoreFile}
              />
              <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeDevices}
                  disabled={busy}
                  onChange={(e) => setIncludeDevices(e.target.checked)}
                  className="w-4 h-4 accent-indigo-500"
                />
                Also restore placed devices (entity bindings)
              </label>
            </div>
          )}

          {step === 'blank' && (
            <form onSubmit={handleBlank} className="space-y-3">
              <div>
                <label className="text-xs font-semibold text-slate-300">Level name</label>
                <input
                  value={blankName}
                  onChange={(e) => setBlankName(e.target.value)}
                  disabled={busy}
                  placeholder="Ground floor"
                  className="mt-1 w-full h-9 bg-slate-950 border border-slate-700 rounded-lg px-3 text-sm text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
              <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={blankOutdoor}
                  disabled={busy}
                  onChange={(e) => setBlankOutdoor(e.target.checked)}
                  className="w-4 h-4 accent-indigo-500"
                />
                Outdoor area (garden, patio…)
              </label>
              <button
                type="submit"
                disabled={busy}
                className="w-full h-9 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-sm font-semibold text-white flex items-center justify-center gap-2 cursor-pointer"
              >
                {busy && <Loader2 className="w-4 h-4 animate-spin" />}
                Create level
              </button>
            </form>
          )}

          {step === 'done' && (
            <div className="space-y-3 text-center py-2">
              <CheckCircle2 className="w-10 h-10 text-emerald-400 mx-auto" />
              <div className="text-sm font-semibold text-white">{resultTitle}</div>
              {resultDetail && <div className="text-xs text-slate-400">{resultDetail}</div>}
              <button
                type="button"
                onClick={onDone}
                className="w-full h-9 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-sm font-semibold text-white cursor-pointer"
              >
                Open dashboard
              </button>
            </div>
          )}

          {busy && step !== 'blank' && step !== 'done' && (
            <div className="flex items-center gap-2 text-xs text-slate-400 pt-1">
              <Loader2 className="w-4 h-4 animate-spin" /> Working…
            </div>
          )}

          {error && (
            <div className="mt-3 p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs">
              {error}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function ChoiceButton({
  icon,
  title,
  desc,
  onClick,
}: {
  icon: React.ReactNode
  title: string
  desc: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full flex items-center gap-3 p-3 rounded-xl border border-slate-700 bg-slate-800/50 hover:bg-slate-800 hover:border-indigo-500/60 transition-colors text-left cursor-pointer"
    >
      <span className="shrink-0">{icon}</span>
      <span>
        <span className="block text-sm font-semibold text-white">{title}</span>
        <span className="block text-xs text-slate-400">{desc}</span>
      </span>
    </button>
  )
}

function DropZone({
  accept,
  inputRef,
  disabled,
  label,
  onFile,
}: {
  accept: string
  inputRef: React.RefObject<HTMLInputElement | null>
  disabled: boolean
  label: string
  onFile: (file: File) => void
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={label}
      onClick={() => {
        if (!disabled) inputRef.current?.click()
      }}
      onKeyDown={(e) => {
        if ((e.key === 'Enter' || e.key === ' ') && !disabled) inputRef.current?.click()
      }}
      onDragOver={(e) => e.preventDefault()}
      onDrop={(e) => {
        e.preventDefault()
        const file = e.dataTransfer.files?.[0]
        if (file && !disabled) void onFile(file)
      }}
      className="border-2 border-dashed border-slate-700 hover:border-indigo-500 rounded-xl p-8 text-center text-xs text-slate-400 hover:text-slate-200 transition-colors cursor-pointer"
    >
      <Upload className="w-6 h-6 mx-auto mb-2 text-slate-500" />
      {label}
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        disabled={disabled}
        onChange={(e) => {
          const file = e.target.files?.[0]
          e.target.value = ''
          if (file) void onFile(file)
        }}
      />
    </div>
  )
}
