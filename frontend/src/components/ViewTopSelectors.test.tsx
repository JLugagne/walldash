import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ViewTopSelectors } from './ViewTopSelectors'
import type { Level } from '../types'

const levels: Level[] = [
  { id: 'l1', name: 'Ground floor', order: 0, is_outdoor: false, created_at: '', updated_at: '' },
  { id: 'l2', name: '1st floor', order: 1, is_outdoor: false, created_at: '', updated_at: '' },
]

describe('ViewTopSelectors', () => {
  it('marks only the overview action active while the house overview is shown', () => {
    render(
      <ViewTopSelectors
        levels={levels}
        activeLevelId="l2"
        onSelectLevel={vi.fn()}
        overviewActive
        onToggleOverview={vi.fn()}
      />
    )

    const pressed = screen
      .getAllByRole('button')
      .filter((button) => button.getAttribute('aria-pressed') === 'true')

    expect(pressed).toHaveLength(1)
    expect(pressed[0].getAttribute('aria-label')).toBe('Back to floor view')
  })

  it('marks the active level when the house overview is not shown', () => {
    render(
      <ViewTopSelectors
        levels={levels}
        activeLevelId="l2"
        onSelectLevel={vi.fn()}
        overviewActive={false}
        onToggleOverview={vi.fn()}
      />
    )

    const levelButton = screen.getByRole('button', { name: 'Level 1st floor' })
    expect(levelButton.getAttribute('aria-pressed')).toBe('true')
  })
})
