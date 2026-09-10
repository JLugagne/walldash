import React, { useMemo } from 'react'
import { Html } from '@react-three/drei'
import * as THREE from 'three'
import {
  Lightbulb,
  Power,
  Thermometer,
  Flame,
  Music,
  Cpu,
  Droplets,
} from 'lucide-react'
import type { Device, DevicePlacement, WallSegment } from '../types'
import { toWorldX, toWorldZ, snapToNearestWall, type WallSnap } from './wallGeometry'
import { getRadialGlowTexture } from './floorTexture'

const FLOOR_MOUNT_HEIGHT = 0.3
const WALL_MOUNT_HEIGHT = 1.45
const FREE_MOUNT_HEIGHT = 1.0
const WALL_SNAP_DISTANCE = 2.0
const WALL_OFFSET = 0.09

type MountKind = 'ceiling' | 'floor' | 'wall' | 'free'

interface Mount {
  kind: MountKind
  x: number
  y: number
  z: number
  snap: WallSnap | null
}

function mountKindFor(domain: string): MountKind {
  if (domain === 'light') return 'ceiling'
  if (domain === 'switch') return 'floor'
  if (domain === 'sensor' || domain === 'climate' || domain === 'binary_sensor') return 'wall'
  return 'free'
}

// Lights hang from the ceiling, plugs sit near the floor, sensors stick to the closest wall at eye level
function resolveMount(domain: string, walls: WallSegment[], x: number, z: number, ceilingY: number, lightMountHeight?: number): Mount {
  const kind = mountKindFor(domain)
  if (kind === 'ceiling') return { kind, x, y: lightMountHeight ?? ceilingY - 0.05, z, snap: null }
  if (kind === 'floor') return { kind, x, y: FLOOR_MOUNT_HEIGHT, z, snap: null }
  if (kind === 'wall') {
    const snap = snapToNearestWall(walls, x, z)
    if (snap && snap.distance <= WALL_SNAP_DISTANCE) {
      return {
        kind,
        x: snap.x + snap.normalX * WALL_OFFSET,
        y: WALL_MOUNT_HEIGHT,
        z: snap.z + snap.normalZ * WALL_OFFSET,
        snap,
      }
    }
    return { kind: 'free', x, y: WALL_MOUNT_HEIGHT, z, snap: null }
  }
  return { kind, x, y: FREE_MOUNT_HEIGHT, z, snap: null }
}

interface DeviceBadge3DProps {
  placement: DevicePlacement
  device?: Device
  isPending?: boolean
  onToggle: (entityId: string) => void
  ceilingY: number
  walls?: WallSegment[]
  interactive?: boolean
  /** Overrides the mount height for lamp-type devices (used by the house overview). */
  lightMountHeight?: number
}

// Warm Three.js point light mounted at the lamp, illuminating nearby walls/floor
function CeilingLight({ x, y, z }: { x: number; y: number; z: number }) {
  return (
    <pointLight
      position={[x, y, z]}
      color="#fbbf24"
      intensity={28}
      distance={8}
      decay={2}
    />
  )
}

// Additive light pool projected on the floor under an active light
function FloorLightPool({ x, z }: { x: number; z: number }) {
  const texture = useMemo(() => getRadialGlowTexture(), [])
  if (!texture) return null
  return (
    <mesh rotation={[-Math.PI / 2, 0, 0]} position={[x, 0.04, z]} renderOrder={1}>
      <circleGeometry args={[2.1, 48]} />
      <meshBasicMaterial
        map={texture}
        color="#f59e0b"
        transparent
        opacity={0.8}
        blending={THREE.AdditiveBlending}
        depthWrite={false}
      />
    </mesh>
  )
}

// Floor anchor ring plus a thin stem up to the mount height, so the badge visibly points to its floor position
function FloorPin({ x, z, height, color, faint }: { x: number; z: number; height: number; color: string; faint?: boolean }) {
  return (
    <group position={[x, 0, z]}>
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.045, 0]} renderOrder={4}>
        <ringGeometry args={[0.17, 0.25, 32]} />
        <meshBasicMaterial color={color} transparent opacity={faint ? 0.55 : 0.9} depthWrite={false} />
      </mesh>
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.045, 0]} renderOrder={4}>
        <circleGeometry args={[0.06, 16]} />
        <meshBasicMaterial color={color} transparent opacity={faint ? 0.7 : 1} depthWrite={false} />
      </mesh>
      <mesh position={[0, height / 2, 0]} renderOrder={4}>
        <cylinderGeometry args={[faint ? 0.012 : 0.02, faint ? 0.012 : 0.02, height, 8]} />
        <meshBasicMaterial color={color} transparent opacity={faint ? 0.4 : 0.75} depthWrite={false} />
      </mesh>
    </group>
  )
}

// Small plate flush against the wall face where a sensor is mounted
function WallPlate({ mount, color }: { mount: Mount; color: string }) {
  if (!mount.snap) return null
  const { snap } = mount
  return (
    <group position={[snap.x + snap.normalX * 0.02, WALL_MOUNT_HEIGHT, snap.z + snap.normalZ * 0.02]} rotation={[0, snap.rotY, 0]}>
      <mesh>
        <boxGeometry args={[0.24, 0.3, 0.04]} />
        <meshStandardMaterial color="#e2e8f0" roughness={0.4} metalness={0.1} />
      </mesh>
      <mesh position={[0, -0.09, 0.025]}>
        <boxGeometry args={[0.12, 0.02, 0.01]} />
        <meshBasicMaterial color={color} />
      </mesh>
    </group>
  )
}

function pinColor(domain: string, isOn: boolean, isSensor: boolean, isPending: boolean): string {
  if (isPending) return '#818cf8'
  if (isSensor) return '#34d399'
  if (!isOn) return '#64748b'
  if (domain === 'light') return '#f59e0b'
  if (domain === 'switch') return '#06b6d4'
  if (domain === 'media_player') return '#a855f7'
  return '#94a3b8'
}

function getIcon(domain: string, entityId: string, className = 'w-4 h-4') {
  if (domain === 'light') return <Lightbulb className={className} />
  if (domain === 'switch') return <Power className={className} />
  if (domain === 'media_player') return <Music className={className} />
  if (domain === 'climate') return <Flame className={className} />
  if (domain === 'sensor') {
    if (entityId.includes('humid') || entityId.includes('hygro')) {
      return <Droplets className={className} />
    }
    return <Thermometer className={className} />
  }
  return <Cpu className={className} />
}

// Format permanent sensor value and unit, used as the hover/accessibility label
function formatSensorValue(domain: string, device?: Device): string | null {
  if (!device) return null
  if (domain === 'climate') {
    const curTemp = device.attributes?.current_temperature
    const targetTemp = device.attributes?.temperature
    if (curTemp !== undefined) {
      return `${curTemp}°C` + (targetTemp ? ` → ${targetTemp}°C` : '')
    }
    return device.state
  }

  const unit = device.attributes?.unit_of_measurement || ''
  return `${device.state}${unit ? ' ' + unit : ''}`
}

export function DeviceBadge3D({ placement, device, isPending = false, onToggle, ceilingY, walls = [], interactive = true, lightMountHeight }: DeviceBadge3DProps) {
  const worldX = toWorldX(placement.x)
  const worldZ = toWorldZ(placement.y)

  const domain = placement.render_domain || device?.domain || placement.device_id.split('.')[0] || 'device'
  const mount = useMemo(
    () => resolveMount(domain, walls, worldX, worldZ, ceilingY, lightMountHeight),
    [domain, walls, worldX, worldZ, ceilingY, lightMountHeight]
  )
  const state = device?.state || 'off'
  const isOn = state === 'on' || state === 'playing' || state === 'heat'

  const isActuator = domain === 'light' || domain === 'switch' || domain === 'media_player'
  const isSensor = domain === 'sensor' || domain === 'climate'

  const name = placement.custom_name || device?.name || placement.device_id
  const sensorValue = isSensor ? formatSensorValue(domain, device) : null
  const title = isPending
    ? `${name} • En cours de traitement (Home Assistant)...`
    : isSensor && sensorValue
    ? `${name}: ${sensorValue}`
    : `${name} • ${isOn ? 'ON' : 'OFF'}`

  const color = pinColor(domain, isOn, isSensor, isPending)

  const handleTap = (e: React.SyntheticEvent) => {
    e.stopPropagation()
    if (interactive && isActuator && !isPending) {
      onToggle(placement.device_id)
    }
  }

  const handlePointerStop = (e: React.PointerEvent | React.MouseEvent | React.TouchEvent) => {
    e.stopPropagation()
  }

  return (
    <group>
      {/* Light Halo: rendered only when light is confirmed ON and not pending */}
      {domain === 'light' && isOn && !isPending && (
        <>
          <CeilingLight x={worldX} y={mount.y} z={worldZ} />
          <FloorLightPool x={worldX} z={worldZ} />
        </>
      )}

      {mount.kind === 'wall' ? (
        <WallPlate mount={mount} color={color} />
      ) : (
        <FloorPin x={worldX} z={worldZ} height={mount.y} color={color} faint={mount.kind === 'ceiling'} />
      )}

      {/* Billboarded badge at the mount point; floor and free mounts sit on top of the stem */}
      <Html
        position={[mount.x, mount.y, mount.z]}
        zIndexRange={[100, 51]}
        style={{ pointerEvents: 'auto' }}
      >
        <div
          className={`relative flex flex-col items-center justify-center -translate-x-1/2 ${
            mount.kind === 'floor' || mount.kind === 'free' ? '-translate-y-full' : '-translate-y-1/2'
          }`}
        >
          {/* Warm glow directly behind the icon, always pixel-aligned with it */}
          {domain === 'light' && isOn && !isPending && (
            <div className="absolute w-14 h-14 rounded-full bg-amber-400/40 blur-xl pointer-events-none" />
          )}

          {/* Intermediate processing indicator: rotating glowing ring */}
          {isPending && (
            <div className="absolute -inset-1.5 rounded-full border-2 border-indigo-400 border-t-transparent animate-spin pointer-events-none" />
          )}

          <button
            type="button"
            data-interactive="true"
            aria-label={title}
             disabled={!interactive || !isActuator || isPending}
            onClick={handleTap}
            onPointerDown={handlePointerStop}
            onPointerUp={handlePointerStop}
            onMouseDown={handlePointerStop}
            onTouchStart={handlePointerStop}
            onTouchEnd={handleTap}
            title={title}
            className={`relative select-none flex items-center justify-center rounded-full border-2 shadow-lg transition-all duration-200 outline-none whitespace-nowrap ${
              isSensor && sensorValue ? 'h-8 min-w-[2.5rem] px-2' : 'w-10 h-10'
            } ${
              isPending
                ? 'cursor-wait bg-indigo-950/90 border-indigo-400/90 text-indigo-300 shadow-indigo-500/50'
                 : interactive && isActuator
                ? 'cursor-pointer hover:scale-110 active:scale-95'
                : 'cursor-default'
            } ${
              !isPending && isOn && domain === 'light'
                ? 'bg-amber-500 border-amber-300 text-slate-950 shadow-amber-500/60'
                : !isPending && isOn && domain === 'switch'
                ? 'bg-cyan-500 border-cyan-300 text-slate-950 shadow-cyan-500/60'
                : !isPending && isOn && domain === 'media_player'
                ? 'bg-purple-500 border-purple-300 text-white shadow-purple-500/60'
                : !isPending && isSensor
                ? 'bg-slate-900/95 border-emerald-500/70 text-emerald-300 shadow-emerald-500/30'
                : !isPending
                ? 'bg-slate-800/90 border-slate-600 text-slate-400 shadow-black/50'
                : ''
            }`}
          >
            {isSensor && sensorValue ? (
              <span className="text-[11px] font-bold leading-none">{sensorValue}</span>
            ) : (
              getIcon(domain, placement.device_id, 'w-5 h-5')
            )}
          </button>

          {/* Discrete "En cours..." pill badge below button when processing */}
          {isPending && (
            <div className="absolute -bottom-5 whitespace-nowrap bg-indigo-950/95 text-indigo-200 border border-indigo-500/50 text-[9px] font-semibold px-1.5 py-0.5 rounded shadow-lg backdrop-blur-md pointer-events-none animate-pulse">
              En cours...
            </div>
          )}
        </div>
      </Html>
    </group>
  )
}
