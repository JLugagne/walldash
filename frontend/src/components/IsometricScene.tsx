import { useEffect, useMemo } from 'react'
import { useThree } from '@react-three/fiber'
import { Html } from '@react-three/drei'
import * as THREE from 'three'
import type { Plan, WallSegment, Zone, Device, DevicePlacement } from '../types'
import { DeviceBadge3D } from './DeviceBadge3D'

interface IsometricSceneProps {
  plan: Plan | null
  zoom: number
  pan: { x: number; z: number }
  placements?: DevicePlacement[]
  deviceMap?: Record<string, Device>
  onToggleDevice?: (entityId: string) => void
}

const SVG_CENTER_X = 500
const SVG_CENTER_Y = 350
const SCALE = 0.04
const WALL_HEIGHT = 2.4

function toWorldX(x: number): number {
  return (x - SVG_CENTER_X) * SCALE
}

function toWorldZ(y: number): number {
  return (y - SVG_CENTER_Y) * SCALE
}

// Controller maintaining strictly fixed isometric camera
function FixedIsometricCamera({ zoom }: { zoom: number }) {
  const { camera } = useThree()

  useEffect(() => {
    // Fixed isometric position: [30, 30, 30] looking at [0, 0, 0]
    // In this orientation:
    // +Z is the front facade (bottom of 2D plan, y=700)
    // -Z is the back facade (top of 2D plan, y=0)
    // +X is the right facade (x=1000)
    // -X is the left facade (x=0)
    camera.position.set(30, 30, 30)
    camera.lookAt(0, 0, 0)
    if ('zoom' in camera) {
      camera.zoom = zoom
      camera.updateProjectionMatrix()
    }
  }, [camera, zoom])

  return null
}

// Extruded 3D wall segment with corner cylinders
function WallMesh({ wall }: { wall: WallSegment }) {
  const wx1 = toWorldX(wall.x1)
  const wz1 = toWorldZ(wall.y1)
  const wx2 = toWorldX(wall.x2)
  const wz2 = toWorldZ(wall.y2)

  const dx = wx2 - wx1
  const dz = wz2 - wz1
  const length = Math.hypot(dx, dz)
  if (length < 0.001) return null

  const midX = (wx1 + wx2) / 2
  const midZ = (wz1 + wz2) / 2
  const rotY = -Math.atan2(dz, dx)
  const thickness3D = Math.max((wall.thickness || 12) * SCALE, 0.25)
  const radius = thickness3D / 2

  return (
    <group>
      {/* Wall Extruded Box */}
      <mesh position={[midX, WALL_HEIGHT / 2, midZ]} rotation={[0, rotY, 0]}>
        <boxGeometry args={[length, WALL_HEIGHT, thickness3D]} />
        <meshStandardMaterial color="#64748b" roughness={0.4} metalness={0.15} />
      </mesh>

      {/* Joint Cylinders at Endpoints for seamless corner connections */}
      <mesh position={[wx1, WALL_HEIGHT / 2, wz1]}>
        <cylinderGeometry args={[radius, radius, WALL_HEIGHT, 16]} />
        <meshStandardMaterial color="#64748b" roughness={0.4} metalness={0.15} />
      </mesh>
      <mesh position={[wx2, WALL_HEIGHT / 2, wz2]}>
        <cylinderGeometry args={[radius, radius, WALL_HEIGHT, 16]} />
        <meshStandardMaterial color="#64748b" roughness={0.4} metalness={0.15} />
      </mesh>
    </group>
  )
}

// 3D Floor Mesh for a Zone
function ZoneMesh({ zone }: { zone: Zone }) {
  const shape = useMemo(() => {
    if (!zone.points || zone.points.length < 3) return null
    const s = new THREE.Shape()
    const p0 = zone.points[0]
    s.moveTo(toWorldX(p0.x), -toWorldZ(p0.y))
    for (let i = 1; i < zone.points.length; i++) {
      const p = zone.points[i]
      s.lineTo(toWorldX(p.x), -toWorldZ(p.y))
    }
    s.closePath()
    return s
  }, [zone.points])

  const outlineGeometry = useMemo(() => {
    if (!zone.points || zone.points.length < 3) return null
    const pts = zone.points.map(
      (p) => new THREE.Vector3(toWorldX(p.x), 0.03, toWorldZ(p.y))
    )
    pts.push(pts[0]) // close polygon loop
    return new THREE.BufferGeometry().setFromPoints(pts)
  }, [zone.points])

  const centroid = useMemo(() => {
    if (!zone.points || zone.points.length === 0) return { x: 0, z: 0 }
    let sumX = 0
    let sumY = 0
    for (const pt of zone.points) {
      sumX += pt.x
      sumY += pt.y
    }
    return {
      x: toWorldX(sumX / zone.points.length),
      z: toWorldZ(sumY / zone.points.length),
    }
  }, [zone.points])

  const lineMesh = useMemo(() => {
    if (!outlineGeometry) return null
    const mat = new THREE.LineBasicMaterial({ color: zone.color, transparent: true, opacity: 0.75 })
    return new THREE.Line(outlineGeometry, mat)
  }, [outlineGeometry, zone.color])

  if (!shape) return null

  return (
    <group>
      {/* Floor surface */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.02, 0]}>
        <shapeGeometry args={[shape]} />
        <meshStandardMaterial
          color={zone.color}
          transparent
          opacity={0.38}
          roughness={0.6}
          metalness={0.05}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Perimeter line */}
      {lineMesh && <primitive object={lineMesh} />}

      {/* Room / Zone Title Label */}
      <Html
        position={[centroid.x, 0.08, centroid.z]}
        center
        style={{ pointerEvents: 'none' }}
      >
        <div className="px-2.5 py-1 rounded-full bg-slate-900/85 backdrop-blur-md border border-slate-700/60 text-slate-200 text-xs font-semibold shadow-lg shadow-black/50 flex items-center space-x-1.5 whitespace-nowrap select-none">
          <span
            className="w-2 h-2 rounded-full ring-1 ring-white/40"
            style={{ backgroundColor: zone.color }}
          />
          <span>{zone.name}</span>
        </div>
      </Html>
    </group>
  )
}

export function IsometricScene({
  plan,
  zoom,
  pan,
  placements = [],
  deviceMap = {},
  onToggleDevice = () => {},
}: IsometricSceneProps) {
  const walls = plan?.walls || []
  const zones = plan?.zones || []

  return (
    <>
      <FixedIsometricCamera zoom={zoom} />

      {/* Lighting: Soft ambient + directional lights */}
      <ambientLight intensity={0.7} />
      <directionalLight position={[20, 35, 20]} intensity={1.2} />
      <directionalLight position={[-15, 20, -15]} intensity={0.35} />

      {/* Ground plane & subtle grid */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.01, 0]}>
        <planeGeometry args={[100, 100]} />
        <meshStandardMaterial color="#0b0f19" roughness={0.9} />
      </mesh>
      <gridHelper args={[60, 60, '#334155', '#1e293b']} position={[0, 0, 0]} />

      {/* Panned House Content */}
      <group position={[pan.x, 0, pan.z]}>
        {/* Zone Floor Meshes */}
        {zones.map((zone) => (
          <ZoneMesh key={zone.id} zone={zone} />
        ))}

        {/* Extruded 3D Walls */}
        {walls.map((wall) => (
          <WallMesh key={wall.id} wall={wall} />
        ))}

        {/* 3D Device Badges, Sensors & Light Halos */}
        {placements.map((placement) => (
          <DeviceBadge3D
            key={placement.id}
            placement={placement}
            device={deviceMap[placement.device_id]}
            onToggle={onToggleDevice}
          />
        ))}
      </group>
    </>
  )
}
