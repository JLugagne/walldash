import { useCallback, useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { KeyRound, Loader2, RefreshCw, ShieldCheck } from 'lucide-react'
import { apiFetch, readApiError } from '../../api'

interface LoginScreenProps {
  onAuthenticated: () => void | Promise<void>
}

const TOO_MANY_ATTEMPTS = 'Too many attempts. Please wait a moment and try again.'
const INVALID_CODE = 'That code did not match. Check the 6-digit code and try again.'

export function LoginScreen({ onAuthenticated }: LoginScreenProps) {
  const [connecting, setConnecting] = useState(true)
  const [pending, setPending] = useState(false)
  const [code, setCode] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const connect = useCallback(async () => {
    setConnecting(true)
    setError(null)
    try {
      const res = await apiFetch('/api/auth/connect', { method: 'POST' })
      if (res.status === 429) {
        setPending(false)
        setError(TOO_MANY_ATTEMPTS)
      } else if (res.ok) {
        setPending(true)
        setCode('')
      } else {
        setPending(false)
        setError(await readApiError(res))
      }
    } catch {
      setPending(false)
      setError('Unable to reach the server. Check your connection and try again.')
    } finally {
      setConnecting(false)
    }
  }, [])

  useEffect(() => {
    void connect()
  }, [connect])

  const verify = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()
      if (code.length !== 6 || submitting) return
      setSubmitting(true)
      setError(null)
      try {
        const res = await apiFetch('/api/auth/verify', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ code }),
        })
        if (res.status === 429) {
          setError(TOO_MANY_ATTEMPTS)
        } else if (res.status === 401) {
          setError(INVALID_CODE)
        } else if (res.ok) {
          await onAuthenticated()
          return
        } else {
          setError(await readApiError(res))
        }
      } catch {
        setError('Unable to reach the server. Check your connection and try again.')
      } finally {
        setSubmitting(false)
      }
    },
    [code, submitting, onAuthenticated],
  )

  const onCodeChange = useCallback((value: string) => {
    setCode(value.replace(/\D/g, '').slice(0, 6))
  }, [])

  return (
    <div className="flex-1 flex items-center justify-center p-6">
      <div className="w-full max-w-sm rounded-2xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-6 shadow-2xl">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#6d76e8]/15 text-[#6d76e8]">
            <ShieldCheck className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-sm font-semibold text-white">Device approval</h1>
            <p className="text-[11.5px] text-slate-400">Walldash sign-in</p>
          </div>
        </div>

        {connecting ? (
          <div className="mt-6 flex items-center gap-2 text-sm text-slate-400">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Starting sign-in…</span>
          </div>
        ) : pending ? (
          <form className="mt-6 space-y-4" onSubmit={verify}>
            <div className="space-y-1">
              <p className="text-sm font-medium text-slate-100">Waiting for approval</p>
              <p className="text-[11.5px] leading-relaxed text-slate-400">
                Ask the owner to read the 6-digit code from the server logs or the setup tab, then
                enter it below.
              </p>
            </div>

            <div className="space-y-1.5">
              <label htmlFor="otp-code" className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-400">
                <KeyRound className="h-3.5 w-3.5" />
                6-digit code
              </label>
              <input
                id="otp-code"
                name="code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                autoFocus
                placeholder="000000"
                value={code}
                onChange={(event) => onCodeChange(event.target.value)}
                className="w-full rounded-xl border border-slate-800/80 bg-slate-950/60 px-4 py-3 text-center text-2xl font-semibold tracking-[0.5em] text-white outline-none focus:border-[#6d76e8]"
              />
            </div>

            {error && (
              <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3.5 py-2.5 text-[12px] text-rose-300">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={code.length !== 6 || submitting}
              className="flex w-full items-center justify-center gap-2 rounded-xl bg-[#6d76e8] px-4 py-3 text-sm font-semibold text-white transition-all hover:bg-[#7b83ea] active:scale-95 disabled:cursor-not-allowed disabled:opacity-40"
            >
              {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{submitting ? 'Verifying…' : 'Sign in'}</span>
            </button>

            <button
              type="button"
              onClick={() => void connect()}
              className="flex w-full items-center justify-center gap-1.5 text-[11.5px] font-medium text-slate-400 transition-colors hover:text-slate-200"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Request a new code</span>
            </button>
          </form>
        ) : (
          <div className="mt-6 space-y-4">
            {error && (
              <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3.5 py-2.5 text-[12px] text-rose-300">
                {error}
              </div>
            )}
            <button
              type="button"
              onClick={() => void connect()}
              className="flex w-full items-center justify-center gap-2 rounded-xl bg-[#6d76e8] px-4 py-3 text-sm font-semibold text-white transition-all hover:bg-[#7b83ea] active:scale-95"
            >
              <RefreshCw className="h-4 w-4" />
              <span>Try again</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
