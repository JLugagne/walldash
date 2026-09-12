import React from 'react'
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { DashboardHeader } from './DashboardHeader'
import type { Dashboard } from '../../types'

function dashboard(): Dashboard {
  return {
    id: 'd1',
    name: 'Home',
    order: 0,
    cols: 12,
    rows: 8,
    background_image: '',
    background_opacity: 0,
    background_blur: 0,
    background_dim: 0,
    created_at: '',
    updated_at: '',
    widgets: [],
  }
}

function renderHeader(overrides: Partial<React.ComponentProps<typeof DashboardHeader>> = {}) {
  const props = {
    dashboard: dashboard(),
    isEditMode: false,
    onRename: vi.fn().mockResolvedValue(true),
    onResize: vi.fn().mockResolvedValue(true),
    onDelete: vi.fn(),
    onAddWidget: vi.fn(),
    onToggleEditMode: vi.fn(),
    onToggleBackground: vi.fn(),
    ...overrides,
  }
  return { ...render(<DashboardHeader {...props} />), props }
}

describe('DashboardHeader grid size', () => {
  it('opens with the dashboard grid size and applies new dimensions', async () => {
    const { props } = renderHeader()

    fireEvent.click(screen.getByRole('button', { name: /grid/i }))

    const cols = screen.getByLabelText('Columns') as HTMLInputElement
    const rows = screen.getByLabelText('Rows') as HTMLInputElement
    expect(cols.value).toBe('12')
    expect(rows.value).toBe('8')

    fireEvent.change(cols, { target: { value: '6' } })
    fireEvent.change(rows, { target: { value: '4' } })
    fireEvent.click(screen.getByRole('button', { name: /apply/i }))

    await waitFor(() => expect(props.onResize).toHaveBeenCalledWith(6, 4))
    await waitFor(() => expect(screen.queryByLabelText('Columns')).toBeNull())
  })

  it('keeps the editor open and shows the server message when a resize is refused', async () => {
    renderHeader({
      onResize: vi.fn().mockRejectedValue(new Error('widget w-1 overflows the grid height')),
    })

    fireEvent.click(screen.getByRole('button', { name: /grid/i }))
    fireEvent.click(screen.getByRole('button', { name: /apply/i }))

    expect(await screen.findByText('widget w-1 overflows the grid height')).toBeDefined()
    expect(screen.getByLabelText('Columns')).toBeDefined()
  })
})
