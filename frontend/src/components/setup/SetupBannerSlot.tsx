import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'

interface SetupBannerSlotContextValue {
  slot: HTMLElement | null
  setSlot: (node: HTMLElement | null) => void
}

const SetupBannerSlotContext = createContext<SetupBannerSlotContextValue | null>(null)

const noop = () => {}

export function SetupBannerSlotProvider({ children }: { children: ReactNode }) {
  const [slot, setSlot] = useState<HTMLElement | null>(null)
  const value = useMemo(() => ({ slot, setSlot }), [slot])
  return <SetupBannerSlotContext.Provider value={value}>{children}</SetupBannerSlotContext.Provider>
}

export function useSetupBannerSlot(): HTMLElement | null {
  const ctx = useContext(SetupBannerSlotContext)
  return ctx ? ctx.slot : null
}

export function useSetSetupBannerSlot(): (node: HTMLElement | null) => void {
  const ctx = useContext(SetupBannerSlotContext)
  return ctx ? ctx.setSlot : noop
}
