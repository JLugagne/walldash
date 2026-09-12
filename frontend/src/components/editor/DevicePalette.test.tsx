import { describe, it, expect, vi } from 'vitest'
import { render, fireEvent } from '@testing-library/react'
import { DevicePalette } from './DevicePalette'
import type { Device } from '../../types'

const DEVICE: Device = {
  id: 'light.1',
  name: 'Lamp',
  domain: 'light',
  state: 'on',
  attributes: {},
  last_updated: '',
}

function renderPalette(onDragDeviceStart = vi.fn()) {
  const { container } = render(
    <DevicePalette
      devices={[DEVICE]}
      loading={false}
      placedDeviceIds={new Set()}
      deviceToPlace={null}
      onPickDevice={vi.fn()}
      onRefresh={vi.fn()}
      onDragDeviceStart={onDragDeviceStart}
    />,
  )
  return { container, onDragDeviceStart }
}

function icon(container: HTMLElement): HTMLElement {
  return container.querySelector('[data-drag-icon]') as HTMLElement
}

function pointerDown(target: Element, init: Record<string, unknown>) {
  const event = new Event('pointerdown', { bubbles: true, cancelable: true })
  Object.assign(event, init)
  fireEvent(target, event)
}

describe('DevicePalette touch placement', () => {
  it('starts a pointer drag from the device icon for a touch pointer (iPad)', () => {
    const { container, onDragDeviceStart } = renderPalette()

    pointerDown(icon(container), { pointerType: 'touch', clientX: 10, clientY: 20, button: 0, pointerId: 1 })

    expect(onDragDeviceStart).toHaveBeenCalledWith(DEVICE, expect.anything())
  })

  it('leaves the native HTML5 drag alone for a mouse pointer', () => {
    const { container, onDragDeviceStart } = renderPalette()

    pointerDown(icon(container), { pointerType: 'mouse', clientX: 10, clientY: 20, button: 0, pointerId: 1 })

    expect(onDragDeviceStart).not.toHaveBeenCalled()
  })

  it('keeps the tap-to-place crosshair visible without hover (touch has none)', () => {
    const { container } = renderPalette()
    const crosshair = container.querySelector('button[title*="Place by clicking"]') as HTMLElement

    expect(crosshair.className).not.toContain('opacity-0')
  })
})
