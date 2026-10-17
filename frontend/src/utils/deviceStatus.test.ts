import { describe, it, expect } from 'vitest'
import type { Device } from '../types'
import { isDeviceOffline } from './deviceStatus'

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

describe('isDeviceOffline', () => {
  it('treats an unavailable device as offline', () => {
    expect(isDeviceOffline(device('unavailable'))).toBe(true)
  })

  it('treats an unknown device as offline', () => {
    expect(isDeviceOffline(device('unknown'))).toBe(true)
  })

  it('is case-insensitive', () => {
    expect(isDeviceOffline(device('Unavailable'))).toBe(true)
  })

  it('does not treat on or off devices as offline', () => {
    expect(isDeviceOffline(device('on'))).toBe(false)
    expect(isDeviceOffline(device('off'))).toBe(false)
  })

  it('does not treat an absent device as offline', () => {
    expect(isDeviceOffline(undefined)).toBe(false)
  })
})
