import { useMemo, useRef, useEffect } from 'react'
import { useThree, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import type { Zone, Device } from '../types'
import { toWorldX, toWorldZ, WALL_HEIGHT } from './wallGeometry'
import {
  getZoneHUDMetrics,
  calculateDiscDiameter,
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
  const centroid = useMemo(() => {
    const pts = zone.points || []
    if (pts.length === 0) return { x: 0, y: 0 }
    const sum = pts.reduce((acc, p) => ({ x: acc.x + p.x, y: acc.y + p.y }), { x: 0, y: 0 })
    return { x: sum.x / pts.length, y: sum.y / pts.length }
  }, [zone.points])
  const worldX = toWorldX(centroid.x)
  const worldZ = toWorldZ(centroid.y)

  const hud = useMemo(() => getZoneHUDMetrics(zone, deviceMap), [zone, deviceMap])
  const diameter = useMemo(() => calculateDiscDiameter(zone), [zone])

  const discRef = useRef<THREE.Group>(null)
  const { camera } = useThree()

  useFrame(() => {
    if (!discRef.current) return
    alignBillboardToCamera(discRef.current, camera)
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

    // Zone name at top
    ctx.font = '600 48px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif'
    ctx.fillStyle = '#cbd5e1'
    ctx.fillText(hud.zoneName, cx, cy - radius * 0.52)

    if (hud.hasSensors) {
      // Temperature value + °C
      ctx.font = '700 164px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Display", "Inter", sans-serif'
      ctx.fillStyle = hud.isAlert ? hud.accentColor : '#ffffff'
      ctx.shadowColor = 'rgba(0, 0, 0, 0.8)'
      ctx.shadowBlur = 14
      const tempText = hud.tempValue || '--'
      const tempMetrics = ctx.measureText(tempText)
      const unitFontSize = 48
      const gap = 12
      const totalWidth = tempMetrics.width + gap + ctx.measureText('°C').width * (unitFontSize / 164)
      const startX = cx - totalWidth / 2 + tempMetrics.width / 2
      ctx.fillText(tempText, startX, cy - radius * 0.05)
      ctx.shadowBlur = 0
      ctx.font = `600 ${unitFontSize}px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif`
      ctx.fillStyle = '#94a3b8'
      ctx.fillText('°C', startX + tempMetrics.width / 2 + gap + ctx.measureText('°C').width / 2, cy - radius * 0.05 - 14)

      // Humidity gauge arc on top border
      if (hud.humidityPct !== null) {
        const gaugeStart = Math.PI * 0.75
        const gaugeEnd = Math.PI * 0.25
        const gaugeSpan = 2 * Math.PI - gaugeStart + gaugeEnd
        const pct = Math.min(hud.humidityPct / 100, 1)
        const fillAngle = gaugeStart + gaugeSpan * pct

        // Color gradient: teal (dry) → blue (comfort) → orange (humid)
        function humidityColor(p: number) {
          if (p < 0.3) return '#22d3ee'
          if (p < 0.5) return '#60a5fa'
          if (p < 0.7) return '#818cf8'
          return '#f59e0b'
        }

        // Background track
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)'
        ctx.lineWidth = 44
        ctx.lineCap = 'round'
        ctx.beginPath()
        ctx.arc(cx, cy, radius, gaugeStart, gaugeEnd)
        ctx.stroke()

        // Filled portion
        ctx.strokeStyle = humidityColor(pct)
        ctx.lineWidth = 44
        ctx.lineCap = 'round'
        ctx.beginPath()
        ctx.arc(cx, cy, radius, gaugeStart, fillAngle)
        ctx.stroke()
      }

      // Status text at bottom
      ctx.font = '700 38px ui-sans-serif, -apple-system, BlinkMacSystemFont, "SF Pro Text", "Inter", sans-serif'
      ctx.fillStyle = hud.accentColor
      ctx.fillText(hud.statusText, cx, cy + radius * 0.48)
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
