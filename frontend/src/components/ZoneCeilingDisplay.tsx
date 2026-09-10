import { useMemo, useRef, useEffect } from 'react'
import { useThree, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import type { Zone, Device } from '../types'
import { toWorldX, toWorldZ, WALL_HEIGHT } from './wallGeometry'
import {
  getZoneHUDMetrics,
  calculateDiscDiameter,
  getZoneLabelPosition,
} from '../utils/ceilingDisplay'

export interface ZoneCeilingDisplayProps {
  zone: Zone
  deviceMap?: Record<string, Device>
  ceilingY?: number
}

/**
 * Aligns the billboard HUD disc directly with the camera's view plane.
 * Keeping the disc parallel to the camera's projection plane ensures circular
 * elements project onto screen space as perfect circles without foreshortening/tilting.
 */
export function alignBillboardToCamera(
  target: { quaternion: THREE.Quaternion },
  camera: { quaternion: THREE.Quaternion }
) {
  target.quaternion.copy(camera.quaternion)
}

export function ZoneCeilingDisplay({
  zone,
  deviceMap = {},
  ceilingY = WALL_HEIGHT,
}: ZoneCeilingDisplayProps) {
  const labelPosition = useMemo(() => getZoneLabelPosition(zone), [zone])
  const worldX = toWorldX(labelPosition.x)
  const worldZ = toWorldZ(labelPosition.y)

  // Only recompute HUD metrics (and the 1024px canvas texture below) when this
  // zone's own sensors change. deviceMap gets a new identity on every WS
  // state_changed for any device; depending on the whole map would repaint
  // every zone disc on each update.
  const tempDevice = zone.temp_sensor ? deviceMap[zone.temp_sensor] : undefined
  const humidityDevice = zone.humidity_sensor ? deviceMap[zone.humidity_sensor] : undefined
  const hud = useMemo(
    () => getZoneHUDMetrics(zone, deviceMap),
    // deviceMap intentionally omitted: it gets a new identity on every WS
    // update for any device, while tempDevice/humidityDevice only change when
    // this zone's own sensors change (unchanged entries keep their reference
    // through the setDeviceMap spread).
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [zone, tempDevice, humidityDevice]
  )
  const diameter = useMemo(() => calculateDiscDiameter(zone), [zone])

  const discRef = useRef<THREE.Group>(null)
  const { camera } = useThree()

  // In demand frameloop mode this only runs when a frame is already scheduled
  // (camera move, data change), which is exactly when re-alignment is needed.
  // The equality check skips redundant quaternion writes.
  useFrame(() => {
    const disc = discRef.current
    if (!disc) return
    if (disc.quaternion.equals(camera.quaternion)) return
    alignBillboardToCamera(disc, camera)
  })

  // Subtle optical distance compensation for deeper rooms
  const compScale = useMemo(() => 1.0 + Math.max(0, -worldZ) * 0.025, [worldZ])

  const texture = useMemo(() => {
    if (typeof document === 'undefined' || !document.createElement) {
      return null
    }

    const canvas = document.createElement('canvas')
    const size = 1024
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext('2d')
    if (!ctx) return null

    ctx.clearRect(0, 0, size, size)

    const cx = size / 2
    const cy = size / 2
    const radius = size * 0.44

    // 1. Translucent smoked glass circular lens
    const bgGrad = ctx.createRadialGradient(cx, cy, radius * 0.1, cx, cy, radius)
    bgGrad.addColorStop(0, 'rgba(15, 23, 42, 0.72)')
    bgGrad.addColorStop(0.8, 'rgba(10, 16, 32, 0.65)')
    bgGrad.addColorStop(1, 'rgba(10, 16, 32, 0.45)')
    ctx.fillStyle = bgGrad
    ctx.beginPath()
    ctx.arc(cx, cy, radius, 0, Math.PI * 2)
    ctx.fill()

    // 2. Fine circular border (1px)
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.22)'
    ctx.lineWidth = 4
    ctx.beginPath()
    ctx.arc(cx, cy, radius, 0, Math.PI * 2)
    ctx.stroke()

    // 3. Colored accent arc if sensors configured
    if (hud.hasSensors) {
      const arcStart = Math.PI * 0.7
      const arcEnd = Math.PI * 1.35

      ctx.strokeStyle = hud.accentColor
      ctx.lineWidth = 10
      ctx.lineCap = 'round'
      ctx.beginPath()
      ctx.arc(cx, cy, radius, arcStart, arcEnd)
      ctx.stroke()

      // Subtle specular shine on arc
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.5)'
      ctx.lineWidth = 3
      ctx.beginPath()
      ctx.arc(cx, cy, radius, arcStart + 0.05, arcEnd - 0.05)
      ctx.stroke()
    }

    // 4. Centered typography
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    if (hud.hasSensors) {
      // Temperature value + °C centered in upper/middle area
      ctx.font = '700 176px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Display", "Inter", sans-serif'
      ctx.fillStyle = hud.isAlert ? hud.accentColor : '#ffffff'
      ctx.shadowColor = 'rgba(0, 0, 0, 0.8)'
      ctx.shadowBlur = 14
      const tempText = hud.tempValue || '--'
      const tempMetrics = ctx.measureText(tempText)
      const unitFontSize = 52
      const gap = 12
      const totalWidth = tempMetrics.width + gap + ctx.measureText('°C').width * (unitFontSize / 176)
      const startX = cx - totalWidth / 2 + tempMetrics.width / 2
      ctx.fillText(tempText, startX, cy - radius * 0.12)
      ctx.shadowBlur = 0
      ctx.font = `600 ${unitFontSize}px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif`
      ctx.fillStyle = '#94a3b8'
      ctx.fillText('°C', startX + tempMetrics.width / 2 + gap + ctx.measureText('°C').width / 2, cy - radius * 0.12 - 16)

      // Humidity gauge arc on top border
      if (hud.humidityPct !== null) {
        const gaugeStart = Math.PI * 0.75
        const gaugeEnd = Math.PI * 0.25
        const gaugeSpan = 2 * Math.PI - gaugeStart + gaugeEnd
        const pct = Math.min(hud.humidityPct / 100, 1)
        const fillAngle = gaugeStart + gaugeSpan * pct

        // Background track
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)'
        ctx.lineWidth = 44
        ctx.lineCap = 'round'
        ctx.beginPath()
        ctx.arc(cx, cy, radius, gaugeStart, gaugeEnd)
        ctx.stroke()

        // Filled portion
        ctx.strokeStyle = hud.humidityColor
        ctx.lineWidth = 44
        ctx.lineCap = 'round'
        ctx.beginPath()
        ctx.arc(cx, cy, radius, gaugeStart, fillAngle)
        ctx.stroke()
      }

      // Status text (COMFORT, TOO COOL, etc.) prominently displayed and centered below
      ctx.font = '700 56px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif'
      ctx.fillStyle = hud.accentColor
      ctx.letterSpacing = '2px'
      ctx.fillText(hud.statusText, cx, cy + radius * 0.42)
    } else {
      // If no sensors, display zone name in center
      ctx.font = '600 56px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif'
      ctx.fillStyle = '#cbd5e1'
      ctx.fillText(hud.zoneName, cx, cy)
    }

    const tex = new THREE.CanvasTexture(canvas)
    tex.anisotropy = 4
    tex.needsUpdate = true
    return tex
  }, [hud])

  useEffect(() => {
    return () => {
      if (texture) {
        texture.dispose()
      }
    }
  }, [texture])

  const stemHeight = ceilingY + 0.13

  return (
    <group name="zone-ceiling-display" position={[worldX, 0, worldZ]}>
      {/* Floor Anchor Ring */}
      <mesh
        name="zone-ceiling-floor-ring"
        rotation={[-Math.PI / 2, 0, 0]}
        position={[0, 0.02, 0]}
      >
        <ringGeometry args={[0.11, 0.14, 32]} />
        <meshBasicMaterial
          color={hud.accentColor}
          transparent
          opacity={0.65}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Floor Anchor Center Dot */}
      <mesh
        name="zone-ceiling-floor-dot"
        rotation={[-Math.PI / 2, 0, 0]}
        position={[0, 0.022, 0]}
      >
        <circleGeometry args={[0.025, 16]} />
        <meshBasicMaterial
          color={hud.accentColor}
          transparent
          opacity={0.75}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Vertical Stem Line from floor to ceiling */}
      <mesh
        name="zone-ceiling-stem"
        position={[0, stemHeight / 2 + 0.02, 0]}
      >
        <cylinderGeometry args={[0.005, 0.005, stemHeight, 8]} />
        <meshBasicMaterial
          color={hud.accentColor}
          transparent
          opacity={hud.isAlert ? 0.65 : 0.35}
        />
      </mesh>

      {/* Ceiling Floating HUD Disc */}
      <group
        ref={discRef}
        name="zone-ceiling-disc-group"
        position={[0, ceilingY + 0.15, 0]}
        scale={[compScale, compScale, compScale]}
      >
        <mesh name="zone-ceiling-disc" renderOrder={999}>
          <planeGeometry args={[diameter, diameter]} />
          <meshBasicMaterial
            map={texture || undefined}
            transparent
            depthWrite={false}
            depthTest={false}
            side={THREE.DoubleSide}
          />
        </mesh>
      </group>
    </group>
  )
}
