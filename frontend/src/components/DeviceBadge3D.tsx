import React from 'react'
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
import type { Device, DevicePlacement } from '../types'

interface DeviceBadge3DProps {
  placement: DevicePlacement
  device?: Device
  onToggle: (entityId: string) => void
}

const SVG_CENTER_X = 500
const SVG_CENTER_Y = 350
const SCALE = 0.04

function toWorldX(x: number): number {
  return (x - SVG_CENTER_X) * SCALE
}

function toWorldZ(y: number): number {
  return (y - SVG_CENTER_Y) * SCALE
}

// Light Halo: Warm Three.js point light + glowing translucent circular disc on ground
function LightHalo({ x, z }: { x: number; z: number }) {
  return (
    <group position={[x, 0, z]}>
      {/* Three.js warm point light illuminating surroundings */}
      <pointLight
        position={[0, 0.4, 0]}
        color="#fbbf24"
        intensity={3.2}
        distance={7.0}
        decay={1.6}
      />

      {/* Outer translucent luminous radial disc */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.035, 0]}>
        <circleGeometry args={[1.8, 32]} />
        <meshBasicMaterial
          color="#f59e0b"
          transparent
          opacity={0.35}
          side={THREE.DoubleSide}
          depthWrite={false}
        />
      </mesh>

      {/* Inner brighter warm core disc */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.038, 0]}>
        <circleGeometry args={[0.9, 32]} />
        <meshBasicMaterial
          color="#fef08a"
          transparent
          opacity={0.65}
          side={THREE.DoubleSide}
          depthWrite={false}
        />
      </mesh>
    </group>
  )
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

export function DeviceBadge3D({ placement, device, onToggle }: DeviceBadge3DProps) {
  const worldX = toWorldX(placement.x)
  const worldZ = toWorldZ(placement.y)

  const domain = device?.domain || placement.device_id.split('.')[0] || 'device'
  const state = device?.state || 'off'
  const isOn = state === 'on' || state === 'playing' || state === 'heat'

  const isActuator = domain === 'light' || domain === 'switch' || domain === 'media_player'
  const isSensor = domain === 'sensor' || domain === 'climate'

  const name = placement.custom_name || device?.name || placement.device_id

  // Format permanent sensor value and unit
  const formatSensorValue = () => {
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

  const handleTap = (e: React.MouseEvent | React.TouchEvent) => {
    e.stopPropagation()
    if (isActuator) {
      onToggle(placement.device_id)
    }
  }

  return (
    <group>
      {/* Light Halo: rendered only when light is ON */}
      {domain === 'light' && isOn && <LightHalo x={worldX} z={worldZ} />}

      {/* Billboarded 3D Badge facing camera */}
      <Html
        position={[worldX, 0.45, worldZ]}
        center
        style={{ pointerEvents: 'auto' }}
      >
        <div
          onClick={handleTap}
          onTouchEnd={handleTap}
          className={`group select-none flex items-center space-x-2 px-3 py-2 rounded-2xl backdrop-blur-md border shadow-2xl transition-all duration-200 cursor-pointer ${
            isActuator
              ? 'hover:scale-105 active:scale-95'
              : 'cursor-default'
          } ${
            isOn && domain === 'light'
              ? 'bg-amber-950/80 border-amber-400/80 text-amber-100 shadow-amber-500/40 ring-2 ring-amber-400/40'
              : isOn && domain === 'switch'
              ? 'bg-cyan-950/80 border-cyan-400/80 text-cyan-100 shadow-cyan-500/40 ring-2 ring-cyan-400/40'
              : isOn && domain === 'media_player'
              ? 'bg-purple-950/80 border-purple-400/80 text-purple-100 shadow-purple-500/40 ring-2 ring-purple-400/40'
              : isSensor
              ? 'bg-slate-900/90 border-emerald-500/40 text-emerald-100 shadow-emerald-500/20'
              : 'bg-slate-900/90 border-slate-700/80 text-slate-300 hover:border-slate-500 shadow-black/60'
          }`}
          style={{ minHeight: '44px', minWidth: '44px' }}
        >
          {/* Icon Circle */}
          <div
            className={`w-7 h-7 rounded-xl flex items-center justify-center shrink-0 transition-colors ${
              isOn && domain === 'light'
                ? 'bg-amber-500 text-slate-950 shadow-md shadow-amber-500/50'
                : isOn && domain === 'switch'
                ? 'bg-cyan-500 text-slate-950 shadow-md shadow-cyan-500/50'
                : isOn && domain === 'media_player'
                ? 'bg-purple-500 text-white shadow-md shadow-purple-500/50'
                : isSensor
                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                : 'bg-slate-800 text-slate-400 border border-slate-700'
            }`}
          >
            {getIcon(domain, placement.device_id, 'w-4 h-4')}
          </div>

          {/* Details Label */}
          <div className="flex flex-col text-left leading-tight whitespace-nowrap">
            <span className="text-[11px] font-semibold text-white/90 max-w-[120px] truncate">
              {name}
            </span>

            {/* Permanent Sensor Measure or Actuator State */}
            {isSensor ? (
              <span className="text-xs font-bold text-emerald-400 tracking-tight">
                {formatSensorValue() || '—'}
              </span>
            ) : isActuator ? (
              <div className="flex items-center space-x-1 mt-0.5">
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    isOn
                      ? domain === 'light'
                        ? 'bg-amber-400 shadow-sm shadow-amber-400'
                        : 'bg-cyan-400 shadow-sm shadow-cyan-400'
                      : 'bg-slate-500'
                  }`}
                />
                <span
                  className={`text-[10px] font-bold uppercase tracking-wider ${
                    isOn
                      ? domain === 'light'
                        ? 'text-amber-300'
                        : 'text-cyan-300'
                      : 'text-slate-400'
                  }`}
                >
                  {isOn ? 'ON' : 'OFF'}
                </span>
              </div>
            ) : (
              <span className="text-[10px] text-slate-400">{state}</span>
            )}
          </div>
        </div>
      </Html>
    </group>
  )
}
