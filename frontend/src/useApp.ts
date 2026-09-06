import { useOutletContext } from 'react-router-dom'
import type { Level } from './types'

export interface AppContext {
  levels: Level[]
  activeLevelId: string | null
  setActiveLevelId: (id: string) => void
  fetchLevels: () => Promise<void>
  activeLevel: Level | null
}

export function useApp(): AppContext {
  return useOutletContext<AppContext>()
}