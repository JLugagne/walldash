import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'

interface TopBarSlotContextValue {
  slot: HTMLElement | null
  setSlot: (node: HTMLElement | null) => void
}

const TopBarSlotContext = createContext<TopBarSlotContextValue | null>(null)

const noop = () => {}

export function TopBarSlotProvider({ children }: { children: ReactNode }) {
  const [slot, setSlot] = useState<HTMLElement | null>(null)
  const value = useMemo(() => ({ slot, setSlot }), [slot])
  return <TopBarSlotContext.Provider value={value}>{children}</TopBarSlotContext.Provider>
}

export function useTopBarSlot(): HTMLElement | null {
  const ctx = useContext(TopBarSlotContext)
  return ctx ? ctx.slot : null
}

export function useSetTopBarSlot(): (node: HTMLElement | null) => void {
  const ctx = useContext(TopBarSlotContext)
  return ctx ? ctx.setSlot : noop
}
