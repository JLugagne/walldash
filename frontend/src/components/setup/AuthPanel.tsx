import { useCallback, useEffect, useState } from 'react'
import {
  Check,
  Copy,
  Pencil,
  Plus,
  RefreshCw,
  ShieldCheck,
  Ticket,
  Trash2,
  User,
  X,
} from 'lucide-react'
import { readApiError } from '../../api'
import { authFetch, useAuth } from '../../useAuth'

type AccountRole = 'owner' | 'admin' | 'device'

interface PendingEnrollment {
  pending_id: string
  device_id: string
  label: string
  approved: boolean
  created_at: string
  expires_at: string
}

interface Invite {
  selector: string
  role: AccountRole
  created_by: string
  created_at: string
  expires_at: string
  consumed_at: string | null
  consumed_by: string
  revoked_at: string | null
}

interface CreatedInvite {
  selector: string
  token: string
  role: AccountRole
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
const INVITE_ROLES: AccountRole[] = ['device', 'admin']

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

function inviteState(invite: Invite): { label: string; className: string } {
  if (invite.revoked_at) return { label: 'revoked', className: 'border-rose-500/30 bg-rose-500/10 text-rose-300' }
  if (invite.consumed_at) return { label: 'used', className: 'border-slate-500/30 bg-slate-500/10 text-slate-300' }
  if (new Date(invite.expires_at).getTime() <= Date.now()) {
    return { label: 'expired', className: 'border-amber-500/30 bg-amber-500/10 text-amber-300' }
  }
  return { label: 'active', className: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300' }
}

/** Owner/admin setup screen: pending approvals, invitations, device roles and revocation. */
export function AuthPanel() {
  const { checking, account } = useAuth()
  const role = account?.role ?? ''
  const canManage = role === 'owner' || role === 'admin'

  const [pending, setPending] = useState<PendingEnrollment[]>([])
  const [devices, setDevices] = useState<AuthDevice[]>([])
  const [invites, setInvites] = useState<Invite[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [inviteRole, setInviteRole] = useState<AccountRole>('device')
  const [createdInvite, setCreatedInvite] = useState<CreatedInvite | null>(null)
  const [copied, setCopied] = useState(false)
  const [hideRevoked, setHideRevoked] = useState(true)

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

  const loadInvites = useCallback(async () => {
    try {
      const res = await authFetch('/api/setup/auth/invites')
      const payload = await res.json()
      if (res.ok && payload?.status === 'success' && Array.isArray(payload.data)) {
        setInvites(payload.data)
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
    void loadInvites()
    const timer = window.setInterval(() => {
      void loadPending()
      void loadInvites()
    }, 3000)
    return () => window.clearInterval(timer)
  }, [canManage, loadPending, loadInvites])

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

  const decidePending = async (entry: PendingEnrollment, decision: 'approve' | 'deny') => {
    setBusyId(entry.pending_id)
    setError(null)
    try {
      const res = await authFetch(`/api/setup/auth/pending/${entry.pending_id}/${decision}`, {
        method: 'POST',
      })
      if (res.ok) {
        await loadPending()
        if (decision === 'approve') await loadDevices()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const createInvite = async () => {
    setBusyId('new-invite')
    setError(null)
    setCreatedInvite(null)
    setCopied(false)
    try {
      const res = await authFetch('/api/setup/auth/invites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: inviteRole }),
      })
      const payload = await res.json()
      if (res.ok && payload?.status === 'success') {
        setCreatedInvite(payload.data as CreatedInvite)
        await loadInvites()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const revokeInvite = async (invite: Invite) => {
    setBusyId(invite.selector)
    setError(null)
    try {
      const res = await authFetch(`/api/setup/auth/invites/${invite.selector}/revoke`, {
        method: 'POST',
      })
      if (res.ok) {
        await loadInvites()
      } else {
        setError(await readApiError(res))
      }
    } catch {
      setError('Unable to reach the server')
    } finally {
      setBusyId(null)
    }
  }

  const copyInvite = async () => {
    if (!createdInvite) return
    const link = `${window.location.origin}/?invite=${encodeURIComponent(createdInvite.token)}`
    try {
      await navigator.clipboard.writeText(link)
      setCopied(true)
    } catch {
      setError('Copy failed — select the link manually.')
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

  const inviteLink = createdInvite
    ? `${window.location.origin}/?invite=${encodeURIComponent(createdInvite.token)}`
    : ''

  const visibleDevices = hideRevoked
    ? devices.filter((device) => device.status !== 'revoked')
    : devices

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
      <div className="shrink-0 flex items-center gap-3 px-4 py-3 border-b border-slate-800/80">
        <div className="space-y-0.5">
          <h1 className="text-sm font-semibold text-white">Access</h1>
          <p className="text-[11.5px] text-slate-400">Device approvals, invitations and permissions</p>
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
            <User className="w-3.5 h-3.5 text-[#8b93ee]" />
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
              Pending approvals
            </h2>
          </div>
          {pending.length === 0 ? (
            <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-5 text-center">
              <p className="text-sm text-slate-300">No device waiting for approval</p>
              <p className="text-[11.5px] text-slate-500 mt-1">
                New devices appear here and can be approved in one click.
              </p>
            </div>
          ) : (
            <ul className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md divide-y divide-slate-800/80 overflow-hidden">
              {pending.map((entry) => {
                const isBusy = busyId === entry.pending_id
                return (
                  <li key={entry.pending_id} className="flex items-center gap-3 px-4 py-3">
                    <span className="w-8 h-8 shrink-0 rounded-lg border border-slate-800/80 bg-slate-900/70 flex items-center justify-center text-slate-400">
                      <User className="w-4 h-4" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="text-sm text-slate-100 truncate">{entry.label || entry.device_id}</p>
                      <p className="text-[11px] uppercase tracking-[0.08em] text-slate-500">
                        {entry.approved ? 'Approved — waiting for device' : `Expires ${formatTimestamp(entry.expires_at)}`}
                      </p>
                    </div>
                    <button
                      type="button"
                      aria-label={`Approve ${entry.label || entry.device_id}`}
                      disabled={isBusy || entry.approved}
                      onClick={() => void decidePending(entry, 'approve')}
                      className="inline-flex items-center gap-1 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-1.5 text-xs font-semibold text-emerald-300 hover:text-emerald-200 hover:bg-emerald-500/20 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <Check className="w-3.5 h-3.5" />
                      <span>Approve</span>
                    </button>
                    <button
                      type="button"
                      aria-label={`Deny ${entry.label || entry.device_id}`}
                      disabled={isBusy}
                      onClick={() => void decidePending(entry, 'deny')}
                      className="inline-flex items-center gap-1 rounded-lg border border-slate-800/80 bg-slate-900/70 px-3 py-1.5 text-xs font-semibold text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <X className="w-3.5 h-3.5" />
                      <span>Deny</span>
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </section>

        <section className="space-y-2">
          <div className="flex items-center gap-2">
            <Ticket className="w-3.5 h-3.5 text-[#8b93ee]" />
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
              Invitations
            </h2>
          </div>
          <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-4 space-y-3">
            <p className="text-[11.5px] leading-relaxed text-slate-400">
              Create a single-use invitation to add a device. It expires after 15 minutes. Share the
              link or token with the device you want to enroll.
            </p>
            <div className="flex items-center gap-2">
              <select
                aria-label="Invitation role"
                value={inviteRole}
                onChange={(event) => setInviteRole(event.target.value as AccountRole)}
                className="rounded-lg border border-slate-800/80 bg-slate-950/60 px-2 py-1.5 text-xs text-slate-200 outline-none focus:border-[#6d76e8] cursor-pointer"
              >
                {INVITE_ROLES.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              <button
                type="button"
                onClick={() => void createInvite()}
                disabled={busyId === 'new-invite'}
                className="inline-flex items-center gap-1.5 rounded-lg bg-[#6d76e8] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[#7b83ea] transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Create invitation</span>
              </button>
            </div>

            {createdInvite && (
              <div className="rounded-lg border border-[#6d76e8]/40 bg-[#6d76e8]/10 p-3 space-y-2">
                <p className="text-[11px] font-semibold uppercase tracking-[0.08em] text-[#aab0f4]">
                  New {createdInvite.role} invitation — shown only once
                </p>
                <div className="flex items-center gap-2">
                  <input
                    readOnly
                    aria-label="Invitation link"
                    value={inviteLink}
                    onFocus={(event) => event.target.select()}
                    className="flex-1 rounded-md border border-slate-800/80 bg-slate-950/60 px-2 py-1.5 text-[11px] text-slate-200"
                  />
                  <button
                    type="button"
                    onClick={() => void copyInvite()}
                    className="inline-flex items-center gap-1 rounded-md border border-slate-800/80 bg-slate-900/70 px-2.5 py-1.5 text-[11px] font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
                  >
                    <Copy className="w-3.5 h-3.5" />
                    <span>{copied ? 'Copied' : 'Copy'}</span>
                  </button>
                </div>
                <p className="text-[10.5px] text-slate-500 break-all">Token: {createdInvite.token}</p>
              </div>
            )}

            {invites.length > 0 && (
              <div className="rounded-lg border border-slate-800/80 overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-slate-800/80 text-[10.5px] uppercase tracking-[0.08em] text-slate-500">
                        <th className="px-3 py-2 font-semibold">Role</th>
                        <th className="px-3 py-2 font-semibold">Created</th>
                        <th className="px-3 py-2 font-semibold">Expires</th>
                        <th className="px-3 py-2 font-semibold">State</th>
                        <th className="px-3 py-2 font-semibold text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/80">
                      {invites.map((invite) => {
                        const state = inviteState(invite)
                        const isBusy = busyId === invite.selector
                        return (
                          <tr key={invite.selector}>
                            <td className="px-3 py-2 text-slate-100">{invite.role}</td>
                            <td className="px-3 py-2 text-slate-400">{formatTimestamp(invite.created_at)}</td>
                            <td className="px-3 py-2 text-slate-400">{formatTimestamp(invite.expires_at)}</td>
                            <td className="px-3 py-2">
                              <span className={`rounded-md border px-2 py-0.5 text-[10.5px] font-semibold ${state.className}`}>
                                {state.label}
                              </span>
                            </td>
                            <td className="px-3 py-2 text-right">
                              <button
                                type="button"
                                aria-label={`Revoke invitation ${invite.selector}`}
                                disabled={isBusy || state.label !== 'active'}
                                onClick={() => void revokeInvite(invite)}
                                className="inline-flex items-center gap-1 rounded-lg border border-rose-500/30 bg-rose-500/10 px-2.5 py-1 text-[11px] font-semibold text-rose-300 hover:text-rose-200 hover:bg-rose-500/20 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                              >
                                <Trash2 className="w-3 h-3" />
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
          </div>
        </section>

        <section className="space-y-2">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-3.5 h-3.5 text-[#8b93ee]" />
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
              Devices
            </h2>
            <div className="flex-1" />
            <label className="flex items-center gap-1.5 text-[11px] text-slate-400 hover:text-slate-200 transition-colors cursor-pointer select-none">
              <input
                type="checkbox"
                checked={hideRevoked}
                onChange={(event) => setHideRevoked(event.target.checked)}
                className="h-3.5 w-3.5 rounded border-slate-700 bg-slate-900 accent-[#6d76e8] cursor-pointer"
              />
              <span>Hide revoked</span>
            </label>
          </div>
          {loading && devices.length === 0 ? (
            <div className="flex items-center gap-2 text-slate-400 text-sm">
              <RefreshCw className="w-4 h-4 animate-spin" />
              <span>Loading devices…</span>
            </div>
          ) : visibleDevices.length === 0 ? (
            <div className="rounded-xl border border-slate-800/80 bg-slate-900/70 px-4 py-6 text-center">
              <p className="text-sm text-slate-300">
                {devices.length === 0 ? 'No device accounts yet' : 'No devices match the current filter'}
              </p>
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
                    {visibleDevices.map((device) => {
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
                                  disabled={isBusy || ownerLocked}
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
                              disabled={isBusy || ownerLocked || device.status === 'revoked'}
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
