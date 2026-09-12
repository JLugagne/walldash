import { useCallback, useEffect, useState } from 'react'
import { Check, KeyRound, Pencil, RefreshCw, ShieldCheck, Trash2, User, X } from 'lucide-react'
import { readApiError } from '../../api'
import { authFetch, useAuth } from '../../useAuth'

type AccountRole = 'owner' | 'admin' | 'device'

interface PendingEnrollment {
  device_id: string
  label: string
  code: string
  expires_at: string
}

interface AuthDevice {
  id: string
  label: string
  role: AccountRole
  status: string
  created_at: string
  last_seen: string | null
}

const ALL_ROLES: AccountRole[] = ['owner', 'admin', 'device']

function roleChoices(currentRole: string): AccountRole[] {
  return currentRole === 'owner' ? ALL_ROLES : ['admin', 'device']
}

function formatTimestamp(value: string | null | undefined): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString()
}

function statusClass(status: string): string {
  return status === 'revoked'
    ? 'border-rose-500/30 bg-rose-500/10 text-rose-300'
    : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
}

/** Owner/admin setup screen: pending enrollment codes plus device roles and revocation. */
export function AuthPanel() {
  const { checking, account } = useAuth()
  const role = account?.role ?? ''
  const canManage = role === 'owner' || role === 'admin'

  const [pending, setPending] = useState<PendingEnrollment[]>([])
  const [devices, setDevices] = useState<AuthDevice[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')

  const loadPending = useCallback(async () => {
    try {
      const res = await authFetch('/api/setup/auth/pending')
      const payload = await res.json()
      if (res.ok && payload?.status === 'success' && Array.isArray(payload.data)) {
        setPending(payload.data)
      }
    } catch {
      return
    }
  }, [])

  const loadDevices = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await authFetch('/api/setup/auth/devices')
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
    if (!canManage) return
    void loadPending()
    const timer = window.setInterval(() => {
      void loadPending()
    }, 3000)
    return () => window.clearInterval(timer)
  }, [canManage, loadPending])

  useEffect(() => {
    if (!canManage) return
    void loadDevices()
  }, [canManage, loadDevices])

  const changeRole = async (device: AuthDevice, nextRole: string) => {
    if (nextRole === device.role) return
    setBusyId(device.id)
    setError(null)
    try {
      const res = await authFetch(`/api/setup/auth/devices/${device.id}/role`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: nextRole }),
      })
      if (res.ok) {
        await loadDevices()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const startRename = (device: AuthDevice) => {
    setRenamingId(device.id)
    setRenameValue(device.label)
    setError(null)
  }

  const cancelRename = () => {
    setRenamingId(null)
    setRenameValue('')
  }

  const renameDevice = async (device: AuthDevice) => {
    const label = renameValue.trim()
    if (!label) return
    setBusyId(device.id)
    setError(null)
    try {
      const res = await authFetch(`/api/setup/auth/devices/${device.id}/label`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ label }),
      })
      if (res.ok) {
        cancelRename()
        await loadDevices()
      } else {
        setError(await readApiError(res))
        cancelRename()
      }
    } catch {
      setError('Unable to reach the server')
      cancelRename()
    } finally {
      setBusyId(null)
    }
  }

  const revokeDevice = async (device: AuthDevice) => {
    setBusyId(device.id)
    setError(null)
    try {
      const res = await authFetch(`/api/setup/auth/devices/${device.id}/revoke`, { method: 'POST' })
      if (res.ok) {
        await loadDevices()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  if (checking) {
    return (
      <div className="flex-1 flex items-center justify-center p-6">
        <div className="flex items-center gap-2 text-slate-400 text-sm">
          <RefreshCw className="w-4 h-4 animate-spin" />
          <span>Checking access…</span>
        </div>
      </div>
    )
  }

  if (!canManage) {
    return (
      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
          <ShieldCheck className="w-6 h-6 mx-auto text-slate-500 mb-2" />
          <p className="text-sm text-slate-300">Access management is not available for this account</p>
          <p className="text-[11.5px] text-slate-500 mt-1">
            Only owners and administrators can review devices.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <div className="shrink-0 flex items-center gap-3 px-4 py-3 border-b border-slate-800/80">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Access</h1>
          <p className="text-[11.5px] text-slate-400">Device enrollments and permissions</p>
        </div>
        <div className="flex-1" />
        <button
          type="button"
          onClick={() => void loadDevices()}
          disabled={loading}
          className="flex items-center gap-1.5 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-2 text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-800/60 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-4 space-y-4">
        {error && (
          <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-300">
            {error}
          </div>
        )}

        <section className="space-y-2">
          <div className="flex items-center gap-2">
            <KeyRound className="w-3.5 h-3.5 text-[#8b93ee]" />
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
              Pending enrollments
            </h2>
          </div>
          {pending.length === 0 ? (
            <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-5 text-center">
              <p className="text-sm text-slate-300">No pending enrollment</p>
              <p className="text-[11.5px] text-slate-500 mt-1">
                New devices appear here with a 6-digit code.
              </p>
            </div>
          ) : (
            <ul className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md divide-y divide-slate-800/80 overflow-hidden">
              {pending.map((entry) => (
                <li key={entry.device_id} className="flex items-center gap-3 px-4 py-3">
                  <span className="w-8 h-8 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
                    <User className="w-4 h-4" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-slate-100 truncate">{entry.label || entry.device_id}</p>
                    <p className="text-[11px] uppercase tracking-[0.08em] text-slate-500">
                      Expires {formatTimestamp(entry.expires_at)}
                    </p>
                  </div>
                  <span className="shrink-0 rounded-md border border-[#6d76e8]/40 bg-[#6d76e8]/10 px-3 py-1.5 text-lg font-semibold tracking-[0.3em] text-white tabular-nums">
                    {entry.code}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="space-y-2">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-3.5 h-3.5 text-[#8b93ee]" />
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
              Devices
            </h2>
          </div>
          {loading && devices.length === 0 ? (
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <RefreshCw className="w-4 h-4 animate-spin" />
              <span>Loading devices…</span>
            </div>
          ) : devices.length === 0 ? (
            <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
              <p className="text-sm text-slate-300">No device accounts yet</p>
            </div>
          ) : (
            <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-800/80 text-[10.5px] uppercase tracking-[0.08em] text-slate-500">
                      <th className="px-4 py-2.5 font-semibold">Label</th>
                      <th className="px-4 py-2.5 font-semibold">Role</th>
                      <th className="px-4 py-2.5 font-semibold">Status</th>
                      <th className="px-4 py-2.5 font-semibold">Created</th>
                      <th className="px-4 py-2.5 font-semibold">Last seen</th>
                      <th className="px-4 py-2.5 font-semibold text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/80">
                    {devices.map((device) => {
                      const isBusy = busyId === device.id
                      const ownerLocked = device.role === 'owner' && role !== 'owner'
                      return (
                        <tr key={device.id}>
                          <td className="px-4 py-3 text-slate-100">
                            {renamingId === device.id ? (
                              <div className="flex items-center gap-1.5">
                                <input
                                  type="text"
                                  aria-label={`Label for ${device.label || device.id}`}
                                  value={renameValue}
                                  autoFocus
                                  disabled={isBusy}
                                  maxLength={64}
                                  onChange={(event) => setRenameValue(event.target.value)}
                                  onKeyDown={(event) => {
                                    if (event.key === 'Enter') void renameDevice(device)
                                    if (event.key === 'Escape') cancelRename()
                                  }}
                                  className="h-8 w-full min-w-40 max-w-xs rounded-lg border border-[#6d76e8] bg-slate-800 px-2 text-xs text-white focus:outline-none disabled:opacity-60"
                                />
                                <button
                                  type="button"
                                  aria-label={`Save label for ${device.label || device.id}`}
                                  disabled={isBusy || !renameValue.trim()}
                                  onClick={() => void renameDevice(device)}
                                  className="shrink-0 rounded-md border border-emerald-500/30 bg-emerald-500/10 p-1.5 text-emerald-300 hover:text-emerald-200 hover:bg-emerald-500/20 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
                                >
                                  <Check className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  type="button"
                                  aria-label={`Cancel label for ${device.label || device.id}`}
                                  disabled={isBusy}
                                  onClick={cancelRename}
                                  className="shrink-0 rounded-md border border-slate-800/80 bg-slate-900/70 p-1.5 text-slate-400 hover:text-white hover:bg-slate-800/60 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
                                >
                                  <X className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            ) : (
                              <div className="flex items-center gap-1.5">
                                <span className="truncate">{device.label || device.id}</span>
                                <button
                                  type="button"
                                  aria-label={`Rename ${device.label || device.id}`}
                                  disabled={isBusy}
                                  onClick={() => startRename(device)}
                                  className="shrink-0 rounded-md p-1 text-slate-500 hover:text-white hover:bg-slate-800/60 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
                                >
                                  <Pencil className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            )}
                          </td>
                          <td className="px-4 py-3">
                            {ownerLocked ? (
                              <span className="rounded-md border border-slate-800/80 bg-slate-900/70 px-2.5 py-1 text-[11px] font-semibold text-slate-300">
                                {device.role}
                              </span>
                            ) : (
                              <select
                                aria-label={`Role for ${device.label || device.id}`}
                                value={device.role}
                                disabled={isBusy}
                                onChange={(event) => void changeRole(device, event.target.value)}
                                className="rounded-lg border border-slate-800/80 bg-slate-950/60 px-2 py-1.5 text-xs text-slate-200 outline-none focus:border-[#6d76e8] disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                              >
                                {roleChoices(role).map((option) => (
                                  <option key={option} value={option}>
                                    {option}
                                  </option>
                                ))}
                              </select>
                            )}
                          </td>
                          <td className="px-4 py-3">
                            <span
                              className={`rounded-md border px-2.5 py-1 text-[11px] font-semibold ${statusClass(device.status)}`}
                            >
                              {device.status}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-slate-400">{formatTimestamp(device.created_at)}</td>
                          <td className="px-4 py-3 text-slate-400">{formatTimestamp(device.last_seen)}</td>
                          <td className="px-4 py-3 text-right">
                            <button
                              type="button"
                              aria-label={`Revoke ${device.label || device.id}`}
                              disabled={isBusy || device.status === 'revoked'}
                              onClick={() => void revokeDevice(device)}
                              className="inline-flex items-center gap-1 rounded-lg border border-rose-500/30 bg-rose-500/10 px-3 py-1.5 text-xs font-semibold text-rose-300 hover:text-rose-200 hover:bg-rose-500/20 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                              <span>Revoke</span>
                            </button>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
