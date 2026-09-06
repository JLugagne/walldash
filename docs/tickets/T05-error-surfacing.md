# T05 — Surface server errors in the widget modal and overview CRUD

Kind: bug. Depends on: T00 (edits `OverviewsView.tsx` and `AddWidgetModal.tsx`).

## Problems (code review)
- `handleCreateWidget` / `handleUpdateWidgetContent` resolve normally on a non-2xx response.
  `AddWidgetModal.handleSubmit` only catches thrown errors, so on a 4xx (e.g. overlap, illegal
  display, actuator domain) the modal **closes and nothing was saved**, silently.
- `handleCreateOverview`, `handleRenameOverview`, `handleDeleteOverview` swallow non-2xx the same
  way (only `console.error` on network failure).
- `handleCreateDefaultOverview` posts an `automation_list` widget with
  `automations.slice(0, 3)`; with zero automations the domain rejects it (needs ≥ 1 entity) and
  the failure is silent.

## Scope
- Add `readApiError(res): Promise<string>` in `frontend/src/api.ts` that extracts
  `payload.message ?? payload.error ?? res.statusText`.
- `onCreate` / `onUpdate` throw an `Error(message)` on `!res.ok`. The modal catches it, shows the
  message in the footer slot (same place as `blockingReason`, red instead of amber) and stays
  open; the message clears on any field change.
- Overview create/rename: show the message inline next to the form; delete: reuse the layout
  toast (make the toast component accept `tone: 'error' | 'info'`).
- Default overview: only add the automation widget when `automations.length > 0`.

## Acceptance
- Mock `POST /api/overviews/:id/widgets` to return 400 `{ "status": "fail", "message": "x" }`:
  the modal stays open and displays `x`.
- No behavior change on success paths.
- `tsc` and `oxlint` clean.
