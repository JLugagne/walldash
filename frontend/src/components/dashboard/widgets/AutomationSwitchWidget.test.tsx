import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { AutomationSwitchWidget } from './AutomationSwitchWidget'
import type { Automation, Widget } from '../../../types'

function makeWidget(colSpan = 1, rowSpan = 1): Widget {
  return {
    id: 'w1',
    dashboard_id: 'd1',
    type: 'automation_switch',
    title: 'Evening',
    order: 0,
    col: 0,
    row: 0,
    col_span: colSpan,
    row_span: rowSpan,
    config: {
      entity_ids: [],
      display: 'switch',
      on_automation: 'automation.evening_on',
      off_automation: 'automation.evening_off',
    },
    created_at: '',
    updated_at: '',
  }
}

function automation(id: string, lastTriggered: string | null): Automation {
  return { id, name: id, state: 'on', current: 0, last_triggered: lastTriggered }
}

function okResponse(payload: unknown = { status: 'success' }): Promise<Response> {
  return Promise.resolve({
    ok: true,
    status: 200,
    json: async () => payload,
  } as Response)
}

function switchButton(): HTMLElement {
  return screen.getByRole('button')
}

describe('AutomationSwitchWidget', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/api/csrf-token')) return okResponse({ data: { csrf_token: 'token' } })
      return okResponse()
    }) as unknown as typeof fetch
  })

  it('shows On when the on automation was triggered most recently', () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget()}
        automations={[
          automation('automation.evening_on', '2026-09-12T20:00:00Z'),
          automation('automation.evening_off', '2026-09-12T08:00:00Z'),
        ]}
      />,
    )

    expect(switchButton().getAttribute('aria-pressed')).toBe('true')
  })

  it('shows Off when neither automation has ever triggered', () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget()}
        automations={[
          automation('automation.evening_on', null),
          automation('automation.evening_off', null),
        ]}
      />,
    )

    expect(switchButton().getAttribute('aria-pressed')).toBe('false')
  })

  it('triggers the on automation and moves the switch when it is off', async () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget()}
        automations={[
          automation('automation.evening_on', null),
          automation('automation.evening_off', null),
        ]}
      />,
    )

    fireEvent.click(switchButton())

    await waitFor(() =>
      expect(globalThis.fetch).toHaveBeenCalledWith(
        '/api/automations/automation.evening_on/trigger',
        expect.objectContaining({ method: 'POST' }),
      ),
    )
    await waitFor(() => expect(switchButton().getAttribute('aria-pressed')).toBe('true'))
  })

  it('triggers the off automation when the switch is on', async () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget()}
        automations={[
          automation('automation.evening_on', '2026-09-12T20:00:00Z'),
          automation('automation.evening_off', '2026-09-12T08:00:00Z'),
        ]}
      />,
    )

    fireEvent.click(switchButton())

    await waitFor(() =>
      expect(globalThis.fetch).toHaveBeenCalledWith(
        '/api/automations/automation.evening_off/trigger',
        expect.objectContaining({ method: 'POST' }),
      ),
    )
  })

  it('fills a 1x1 tile with a single exclusive OFF button', () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget(1, 1)}
        automations={[
          automation('automation.evening_on', null),
          automation('automation.evening_off', null),
        ]}
      />,
    )

    expect(screen.getAllByRole('button')).toHaveLength(1)
    expect(screen.getByText('OFF')).toBeDefined()
    expect(screen.queryByText('ON')).toBeNull()
  })

  it('keeps the reading and track layout on a wider tile', () => {
    render(
      <AutomationSwitchWidget
        widget={makeWidget(3, 1)}
        automations={[
          automation('automation.evening_on', '2026-09-12T20:00:00Z'),
          automation('automation.evening_off', '2026-09-12T08:00:00Z'),
        ]}
      />,
    )

    expect(screen.getAllByRole('button')).toHaveLength(1)
    expect(screen.getByText('On')).toBeDefined()
    expect(screen.queryByText('Off')).toBeNull()
  })
})
