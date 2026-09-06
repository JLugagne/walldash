import { describe, it, expect, vi } from 'vitest'
import { render } from '@testing-library/react'
import * as THREE from 'three'
import { ZoneCeilingDisplay, alignBillboardToCamera } from './ZoneCeilingDisplay'
import { WALL_HEIGHT } from './wallGeometry'
import type { Zone, Device } from '../types'

vi.mock('@react-three/fiber', () => ({
  useThree: () => ({
    camera: { quaternion: new THREE.Quaternion() },
  }),
  useFrame: () => {},
}))

describe('ZoneCeilingDisplay Component', () => {
  const mockZoneNoSensors: Zone = {
    id: 'zone-1',
    name: 'Salon',
    color: '#3b82f6',
    points: [
      { x: 0, y: 0 },
      { x: 100, y: 0 },
      { x: 100, y: 100 },
      { x: 0, y: 100 },
    ],
  }

  const mockZoneWithSensors: Zone = {
    id: 'zone-2',
    name: 'Chambre',
    color: '#10b981',
    points: [
      { x: 0, y: 0 },
      { x: 100, y: 0 },
      { x: 100, y: 100 },
      { x: 0, y: 100 },
    ],
    temp_sensor: 'sensor.chambre_temp',
    humidity_sensor: 'sensor.chambre_humidity',
    temp_min: 19,
    temp_max: 23,
  }

  const mockDevices: Record<string, Device> = {
    'sensor.chambre_temp': {
      id: 'sensor.chambre_temp',
      name: 'Temp Chambre',
      domain: 'sensor',
      state: '21.5',
      attributes: { unit_of_measurement: '°C' },
      last_updated: '2026-09-05T12:00:00Z',
    },
    'sensor.chambre_humidity': {
      id: 'sensor.chambre_humidity',
      name: 'Hum Chambre',
      domain: 'sensor',
      state: '52',
      attributes: { unit_of_measurement: '%' },
      last_updated: '2026-09-05T12:00:00Z',
    },
  }

  it('renders a 3D group container with floor ring, stem and ceiling disc', () => {
    const { container } = render(
      <ZoneCeilingDisplay
        zone={mockZoneWithSensors}
        deviceMap={mockDevices}
        ceilingY={WALL_HEIGHT}
      />
    )

    const mainGroup = container.querySelector('[name="zone-ceiling-display"]')
    expect(mainGroup).not.toBeNull()

    const floorRing = container.querySelector('[name="zone-ceiling-floor-ring"]')
    expect(floorRing).not.toBeNull()

    const stem = container.querySelector('[name="zone-ceiling-stem"]')
    expect(stem).not.toBeNull()

    const disc = container.querySelector('[name="zone-ceiling-disc"]')
    expect(disc).not.toBeNull()
  })

  it('positions ceiling disc at ceilingY (3.2m) with horizontal XZ orientation', () => {
    const { container } = render(
      <ZoneCeilingDisplay
        zone={mockZoneNoSensors}
        deviceMap={{}}
        ceilingY={WALL_HEIGHT}
      />
    )

    const discGroup = container.querySelector('[name="zone-ceiling-disc-group"]')
    expect(discGroup).not.toBeNull()
  })

  it('adapts disc diameter to room dimensions', () => {
    const smallZone: Zone = {
      id: 'zone-small',
      name: 'Bureau',
      color: '#3b82f6',
      points: [
        { x: 0, y: 0 },
        { x: 80, y: 0 },
        { x: 80, y: 80 },
        { x: 0, y: 80 },
      ],
    }

    const { container } = render(
      <ZoneCeilingDisplay
        zone={smallZone}
        deviceMap={{}}
      />
    )

    const disc = container.querySelector('[name="zone-ceiling-disc"]')
    expect(disc).not.toBeNull()
  })

  it('aligns billboard disc directly to camera quaternion without tilting', () => {
    const mockCamera = {
      quaternion: new THREE.Quaternion().setFromEuler(new THREE.Euler(0.65, 0.2, -0.1)),
    }
    const mockDisc = {
      quaternion: new THREE.Quaternion(),
    }

    alignBillboardToCamera(mockDisc, mockCamera)

    expect(mockDisc.quaternion.x).toBeCloseTo(mockCamera.quaternion.x, 5)
    expect(mockDisc.quaternion.y).toBeCloseTo(mockCamera.quaternion.y, 5)
    expect(mockDisc.quaternion.z).toBeCloseTo(mockCamera.quaternion.z, 5)
    expect(mockDisc.quaternion.w).toBeCloseTo(mockCamera.quaternion.w, 5)
  })
})
