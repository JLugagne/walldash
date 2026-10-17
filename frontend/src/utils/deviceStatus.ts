import type { Device } from '../types'

const OFFLINE_STATES = new Set(['unavailable', 'unknown'])

/**
 * A device is offline when Home Assistant reports it as `unavailable` (the
 * integration lost contact with it, e.g. a lamp or plug dropped off the
 * network) or `unknown` (its state was never reported). An absent device is
 * not offline: nothing has been loaded yet to judge, so the caller keeps its
 * normal rendering.
 */
export function isDeviceOffline(device: Device | undefined): boolean {
  if (!device) return false
  return OFFLINE_STATES.has(device.state?.toLowerCase() ?? '')
}
