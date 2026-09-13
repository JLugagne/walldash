import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { KeyRound, Loader2, RefreshCw, ShieldCheck } from 'lucide-react'
import { apiFetch, readApiError } from '../../api'

interface LoginScreenProps {
  onAuthenticated: () => void | Promise<void>
}

const TOO_MANY_ATTEMPTS = 'Too many attempts. Please wait a moment and try again.'
const INVALID_INVITE = 'That invitation is invalid or has already been used.'
const APPROVAL_EXPIRED =
  'This approval request expired. Ask an owner for a new invitation, or try again.'

function inviteTokenFromURL(): string {
  if (typeof window === 'undefined') return ''
  return new URLSearchParams(window.location.search).get('invite')?.trim() ?? ''
}

function ErrorBox({ message }: { message: string }) {
  return (
    <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3.5 py-2.5 text-[12px] text-rose-300">
      {message}
    </div>
  )
}

export function LoginScreen({ onAuthenticated }: LoginScreenProps) {
  const [phase, setPhase] = useState<'connecting' | 'pending' | 'invite'>('connecting')
  const [error, setError] = useState<string | null>(null)
  const [invite, setInvite] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const pollRef = useRef<number | null>(null)

  const stopPolling = useCallback(() => {
    if (pollRef.current !== null) {
      window.clearInterval(pollRef.current)
      pollRef.current = null
    }
  }, [])

  const redeem = useCallback(async () => {
    try {
      const res = await apiFetch('/api/auth/redeem', { method: 'POST' })
      // 202 means the enrollment is still awaiting approval: keep polling.
      if (res.status === 202) return
      if (res.ok) {
        stopPolling()
        await onAuthenticated()
        return
      }
      stopPolling()
      setPhase('invite')
      setError(res.status === 401 ? APPROVAL_EXPIRED : await readApiError(res))
    } catch {
      // Transient network errors: keep polling.
    }
  }, [onAuthenticated, stopPolling])

  const startPolling = useCallback(() => {
    stopPolling()
    pollRef.current = window.setInterval(() => {
      void redeem()
    }, 3000)
  }, [redeem, stopPolling])

  const connect = useCallback(async () => {
    setPhase('connecting')
    setError(null)
    try {
      const res = await apiFetch('/api/auth/connect', { method: 'POST' })
      if (res.status === 429) {
        setPhase('invite')
        setError(TOO_MANY_ATTEMPTS)
        return
      }
      if (!res.ok) {
        setPhase('invite')
        setError(await readApiError(res))
        return
      }
      const payload = await res.json()
      if (payload?.data?.status === 'authenticated') {
        await onAuthenticated()
        return
      }
      setPhase('pending')
      startPolling()
    } catch {
      setPhase('invite')
      setError('Unable to reach the server. Check your connection and try again.')
    }
  }, [onAuthenticated, startPolling])

  const redeemInvite = useCallback(
    async (token: string) => {
      const trimmed = token.trim()
      if (!trimmed || submitting) return
      setSubmitting(true)
      setError(null)
      try {
        const res = await apiFetch('/api/auth/invite/redeem', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ token: trimmed }),
        })
        if (res.status === 429) {
          setPhase('invite')
          setError(TOO_MANY_ATTEMPTS)
        } else if (res.status === 401) {
          setPhase('invite')
          setError(INVALID_INVITE)
        } else if (res.ok) {
          stopPolling()
          window.history.replaceState({}, '', window.location.pathname)
          await onAuthenticated()
          return
        } else {
          setPhase('invite')
          setError(await readApiError(res))
        }
      } catch {
        setPhase('invite')
        setError('Unable to reach the server. Check your connection and try again.')
      } finally {
        setSubmitting(false)
      }
    },
    [onAuthenticated, stopPolling, submitting],
  )

  const resumeOrConnect = useCallback(async () => {
    setPhase('connecting')
    setError(null)
    // Resume an existing pending enrollment (for example after a page refresh)
    // instead of minting a brand-new approval request.
    try {
      const res = await apiFetch('/api/auth/redeem', { method: 'POST' })
      if (res.status === 202) {
        setPhase('pending')
        startPolling()
        return
      }
      if (res.ok) {
        await onAuthenticated()
        return
      }
    } catch {
      // fall through to a fresh connect
    }
    await connect()
  }, [connect, onAuthenticated, startPolling])

  useEffect(() => {
    const token = inviteTokenFromURL()
    if (token) {
      void redeemInvite(token)
    } else {
      void resumeOrConnect()
    }
    return () => stopPolling()
    // Runs once on mount: the invite token is read from the URL.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onInviteSubmit = useCallback(
    (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()
      void redeemInvite(invite)
    },
    [invite, redeemInvite],
  )

  const inviteForm = (
    <form className="space-y-3" onSubmit={onInviteSubmit}>
      <div className="space-y-1.5">
        <label
          htmlFor="invite-token"
          className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400"
        >
          <KeyRound className="h-3.5 w-3.5" />
          Invitation token
        </label>
        <input
          id="invite-token"
          name="invite"
          type="text"
          autoComplete="off"
          placeholder="Paste the invitation token or link"
          value={invite}
          onChange={(event) => setInvite(event.target.value)}
          className="w-full rounded-xl border border-slate-800/80 bg-slate-950/60 px-4 py-3 text-sm text-white outline-none focus:border-[#6d76e8]"
        />
      </div>
      <button
        type="submit"
        disabled={!invite.trim() || submitting}
        className="flex w-full items-center justify-center gap-2 rounded-xl bg-[#6d76e8] px-4 py-3 text-sm font-semibold text-white transition-all hover:bg-[#7b83ea] active:scale-95 disabled:cursor-not-allowed disabled:opacity-40"
      >
        {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
        <span>{submitting ? 'Signing in…' : 'Sign in with invitation'}</span>
      </button>
    </form>
  )

  return (
    <div className="flex-1 flex items-center justify-center p-6">
      <div className="w-full max-w-sm rounded-2xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-6 shadow-2xl">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#6d76e8]/15 text-[#6d76e8]">
            <ShieldCheck className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-sm font-semibold text-white">Device access</h1>
            <p className="text-[11.5px] text-slate-400">Walldash sign-in</p>
          </div>
        </div>

        {phase === 'connecting' ? (
          <div className="mt-6 flex items-center gap-2 text-sm text-slate-400">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Starting sign-in…</span>
          </div>
        ) : phase === 'pending' ? (
          <div className="mt-6 space-y-4">
            <div className="flex items-center gap-2 text-sm font-medium text-slate-100">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Waiting for approval</span>
            </div>
            <p className="text-[11.5px] leading-relaxed text-slate-400">
              An owner or administrator must approve this device from the setup panel. This page
              continues automatically once it is approved.
            </p>

            {error && <ErrorBox message={error} />}

            {inviteForm}

            <button
              type="button"
              onClick={() => void connect()}
              className="flex w-full items-center justify-center gap-1.5 text-[11.5px] font-medium text-slate-400 transition-colors hover:text-slate-200"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Start over</span>
            </button>
          </div>
        ) : (
          <div className="mt-6 space-y-4">
            <p className="text-[11.5px] leading-relaxed text-slate-400">
              This device is not signed in yet. An owner can approve it from the setup panel, or
              you can enter an invitation token below.
            </p>

            {error && <ErrorBox message={error} />}

            {inviteForm}

            <button
              type="button"
              onClick={() => void connect()}
              className="flex w-full items-center justify-center gap-1.5 text-[11.5px] font-medium text-slate-400 transition-colors hover:text-slate-200"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Wait for owner approval</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
