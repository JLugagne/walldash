import type { Level } from '../types'

/**
 * Resolves a route-param level id to a selection update, if any.
 *
 * Returns the param id only when it differs from the active level AND matches
 * a known level. Stale params (e.g. a level deleted in another view, or a
 * bookmarked id) resolve to null so the caller never resurrects a ghost
 * selection — which would otherwise fight the URL-sync effect back and forth
 * (param pulls selection toward the ghost, URL effect pushes the URL back).
 */
export function resolveParamLevelId(
  levelId: string | undefined,
  levels: Level[],
  activeLevelId: string | null
): string | null {
  if (!levelId || activeLevelId === levelId) return null
  if (!levels.some((l) => l.id === levelId)) return null
  return levelId
}
