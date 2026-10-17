import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import type { Device, DevicePlacement } from '../types'

vi.mock('@react-three/drei', () => ({
  Html: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}))

import { DeviceBadge3D } from './DeviceBadge3D'

const placement: DevicePlacement = { id: 'p1', level_id: 'l1', device_id: 'light.kitchen', x: 1, y: 1 }

function device(state: string): Device {
  return {
    id: 'light.kitchen',
    name: 'Kitchen lamp',
    domain: 'light',
    state,
    attributes: {},
    last_updated: '2026-01-01T00:00:00Z',
  }
}

describe('DeviceBadge3D offline indicator', () => {
  it('shows the offline indicator and disables the toggle for an unavailable device', () => {
    const onToggle = vi.fn()
    render(<DeviceBadge3D placement={placement} device={device('unavailable')} onToggle={onToggle} ceilingY={2.5} />)

    expect(screen.getByLabelText('Device offline')).toBeDefined()

    const button = screen.getByRole('button') as HTMLButtonElement
    expect(button.getAttribute('title')).toContain('Offline')
    expect(button.disabled).toBe(true)

    button.click()
    expect(onToggle).not.toHaveBeenCalled()
  })

  it('shows the offline indicator for an unknown device too', () => {
    render(<DeviceBadge3D placement={placement} device={device('unknown')} onToggle={() => {}} ceilingY={2.5} />)
    expect(screen.getByLabelText('Device offline')).toBeDefined()
  })

  it('shows no offline indicator and keeps the toggle active for a healthy device', () => {
    const onToggle = vi.fn()
    render(<DeviceBadge3D placement={placement} device={device('on')} onToggle={onToggle} ceilingY={2.5} />)

    expect(screen.queryByLabelText('Device offline')).toBeNull()

    const button = screen.getByRole('button') as HTMLButtonElement
    expect(button.getAttribute('title')).toContain('ON')
    expect(button.disabled).toBe(false)

    button.click()
    expect(onToggle).toHaveBeenCalledWith('light.kitchen')
  })
})
