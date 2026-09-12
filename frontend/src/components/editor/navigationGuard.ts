/** Route prefix that owns the Plan Editor (see `routes.tsx`). */
export const PLAN_EDITOR_PREFIX = '/setup/plans'

/**
 * shouldBlockPlanExit reports whether a navigation must be intercepted because the Plan Editor has
 * unsaved changes and the target leaves the editor.
 *
 * Level switches *inside* `/setup/plans` are deliberately not blocked: the editor owns its own
 * confirmation for discarding the current level before loading another one.
 */
export function shouldBlockPlanExit(
  currentPathname: string,
  nextPathname: string,
  isDirty: boolean,
): boolean {
  return (
    isDirty &&
    currentPathname.startsWith(PLAN_EDITOR_PREFIX) &&
    !nextPathname.startsWith(PLAN_EDITOR_PREFIX)
  )
}
