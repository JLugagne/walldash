import { useEffect, useMemo, useRef } from 'react'
import { useThree } from '@react-three/fiber'
import * as THREE from 'three'
import type { Plan, WallSegment, Zone, Device, DevicePlacement, Layer } from '../types'
import { DeviceBadge3D } from './DeviceBadge3D'
import { ZoneCeilingDisplay } from './ZoneCeilingDisplay'
import { buildWallGeometry, toWorldX, toWorldZ, WALL_HEIGHT } from './wallGeometry'
import { createWallMaterial } from './wallMaterial'
import { getFloorTileTexture } from './floorTexture'
import { hasConfiguredSensors, isCeilingHiddenForActiveLayer } from '../utils/ceilingDisplay'

export { WALL_HEIGHT }

interface IsometricSceneProps {
  plan: Plan | null
  zoom?: number
  pan?: { x: number; z: number }
  viewAngle?: number
  placements?: DevicePlacement[]
  deviceMap?: Record<string, Device>
  pendingDevices?: Record<string, boolean>
  onToggleDevice?: (entityId: string) => void
  activeLayer?: string
  layers?: Layer[]
  onlyLights?: boolean
  interactiveDevices?: boolean
  showCamera?: boolean
  showGround?: boolean
  showLighting?: boolean
  solidWalls?: boolean
  /** Overrides the mount height for lamp-type devices (used by the house overview). */
  lightMountHeight?: number
}

interface PlanBounds {
  minX: number
  maxX: number
  minZ: number
  maxZ: number
}

const VIEW_ANGLE_MAX = 1.45
const VIEW_ANGLE_MIN = 0.05

// Front perspective camera that iteratively re-centers and fits the projected house bounds to the viewport
function PerspectiveSceneCamera({ bounds, viewAngle = 0.6 }: { bounds: PlanBounds; viewAngle?: number }) {
  const { camera, size, invalidate } = useThree()

  useEffect(() => {
    if (!(camera instanceof THREE.PerspectiveCamera)) return

    const aspect = size.width / size.height
    camera.aspect = aspect
    camera.updateProjectionMatrix()

    const t = Math.max(0, Math.min(1, viewAngle))
    const polarAngle = VIEW_ANGLE_MAX - t * (VIEW_ANGLE_MAX - VIEW_ANGLE_MIN)
    const dir = new THREE.Vector3(0, Math.sin(polarAngle), Math.cos(polarAngle))
    const corners: THREE.Vector3[] = []
    for (const x of [bounds.minX, bounds.maxX]) {
      for (const y of [0, WALL_HEIGHT]) {
        for (const z of [bounds.minZ, bounds.maxZ]) corners.push(new THREE.Vector3(x, y, z))
      }
    }

    const target = new THREE.Vector3(
      (bounds.minX + bounds.maxX) / 2,
      WALL_HEIGHT / 2,
      (bounds.minZ + bounds.maxZ) / 2
    )
    const tanHalfV = Math.tan((camera.fov * Math.PI) / 360)
    let dist = Math.max(bounds.maxX - bounds.minX, bounds.maxZ - bounds.minZ)
    const MARGIN = 1.05
    const right = new THREE.Vector3()
    const up = new THREE.Vector3()
    const projected = new THREE.Vector3()

    for (let iteration = 0; iteration < 6; iteration++) {
      camera.position.copy(target).addScaledVector(dir, dist)
      camera.lookAt(target)
      camera.updateMatrixWorld(true)

      let minX = Infinity
      let maxX = -Infinity
      let minY = Infinity
      let maxY = -Infinity
      for (const c of corners) {
        projected.copy(c).project(camera)
        minX = Math.min(minX, projected.x)
        maxX = Math.max(maxX, projected.x)
        minY = Math.min(minY, projected.y)
        maxY = Math.max(maxY, projected.y)
      }

      const halfH = dist * tanHalfV
      const halfW = halfH * aspect
      right.setFromMatrixColumn(camera.matrixWorld, 0)
      up.setFromMatrixColumn(camera.matrixWorld, 1)
      target.addScaledVector(right, ((minX + maxX) / 2) * halfW)
      target.addScaledVector(up, ((minY + maxY) / 2) * halfH)

      const extent = Math.max((maxX - minX) / 2, (maxY - minY) / 2)
      dist = Math.max(dist * extent * MARGIN, 8)
    }

    camera.position.copy(target).addScaledVector(dir, dist)
    camera.lookAt(target)
    camera.updateMatrixWorld(true)
    camera.updateProjectionMatrix()
    // In demand mode no frame is scheduled automatically: request one so the
    // re-fitted camera (and billboards synced to it) actually get drawn.
    invalidate()
  }, [camera, invalidate, size.width, size.height, bounds.minX, bounds.maxX, bounds.minZ, bounds.maxZ, viewAngle])

  return null
}

// 3D Window Mesh with frame, glass pane, and cross mullions (croisillons)
function Window3DMesh({ width, thickness }: { width: number; thickness: number }) {
  const height = 1.2
  const fT = 0.04 // frame profile thickness
  const fD = Math.max(thickness * 0.75, 0.12) // frame depth

  return (
    <group>
      {/* Outer Window Frame (Off-white / Slate) */}
      {/* Top frame bar */}
      <mesh position={[0, height / 2 - fT / 2, 0]}>
        <boxGeometry args={[width, fT, fD]} />
        <meshStandardMaterial color="#f8fafc" roughness={0.3} metalness={0.15} />
      </mesh>
      {/* Bottom frame bar */}
      <mesh position={[0, -height / 2 + fT / 2, 0]}>
        <boxGeometry args={[width, fT, fD]} />
        <meshStandardMaterial color="#f8fafc" roughness={0.3} metalness={0.15} />
      </mesh>
      {/* Left jamb */}
      <mesh position={[-width / 2 + fT / 2, 0, 0]}>
        <boxGeometry args={[fT, height - 2 * fT, fD]} />
        <meshStandardMaterial color="#f8fafc" roughness={0.3} metalness={0.15} />
      </mesh>
      {/* Right jamb */}
      <mesh position={[width / 2 - fT / 2, 0, 0]}>
        <boxGeometry args={[fT, height - 2 * fT, fD]} />
        <meshStandardMaterial color="#f8fafc" roughness={0.3} metalness={0.15} />
      </mesh>

      {/* Croisillons (Cross Bars): Vertical central mullion & Horizontal transom */}
      <mesh position={[0, 0, 0]}>
        <boxGeometry args={[fT * 0.65, height - 2 * fT, fD * 0.6]} />
        <meshStandardMaterial color="#f1f5f9" roughness={0.3} metalness={0.15} />
      </mesh>
      <mesh position={[0, 0, 0]}>
        <boxGeometry args={[width - 2 * fT, fT * 0.65, fD * 0.6]} />
        <meshStandardMaterial color="#f1f5f9" roughness={0.3} metalness={0.15} />
      </mesh>

      {/* Glass Pane with distinct opacity and sky-blue glass reflection */}
      <mesh position={[0, 0, 0]}>
        <boxGeometry args={[width - 2 * fT, height - 2 * fT, 0.02]} />
        <meshStandardMaterial
          color="#7dd3fc"
          transparent
          opacity={0.45}
          roughness={0.08}
          metalness={0.85}
          depthWrite={false}
        />
      </mesh>
    </group>
  )
}

// 3D Door Mesh with frame, door panel, handle, and central mullion if wide
function Door3DMesh({
  width,
  thickness,
  flipSide,
  flipHinge,
}: {
  width: number
  thickness: number
  flipSide?: boolean
  flipHinge?: boolean
}) {
  const height = 2.1
  const fT = 0.045 // frame profile thickness
  const fD = Math.max(thickness * 0.8, 0.14) // frame depth
  const isWide = width >= 1.15 // Wide door threshold (~1.15m in 3D)
  const sideZ = flipSide ? -1 : 1
  const handleZ = sideZ * (fD / 2 + 0.02)

  return (
    <group>
      {/* Door Frame */}
      {/* Top header */}
      <mesh position={[0, height / 2 - fT / 2, 0]}>
        <boxGeometry args={[width, fT, fD]} />
        <meshStandardMaterial color="#334155" roughness={0.4} metalness={0.25} />
      </mesh>
      {/* Left jamb */}
      <mesh position={[-width / 2 + fT / 2, 0, 0]}>
        <boxGeometry args={[fT, height - fT, fD]} />
        <meshStandardMaterial color="#334155" roughness={0.4} metalness={0.25} />
      </mesh>
      {/* Right jamb */}
      <mesh position={[width / 2 - fT / 2, 0, 0]}>
        <boxGeometry args={[fT, height - fT, fD]} />
        <meshStandardMaterial color="#334155" roughness={0.4} metalness={0.25} />
      </mesh>
      {/* Subtle floor threshold bar */}
      <mesh position={[0, -height / 2 + 0.012, 0]}>
        <boxGeometry args={[width, 0.024, fD * 1.05]} />
        <meshStandardMaterial color="#1e293b" roughness={0.5} metalness={0.2} />
      </mesh>

      {/* Door Panels & Central Mullion */}
      {isWide ? (
        // Wide door: central mullion (montant au milieu) + two leaves
        <>
          {/* Montant vertical central */}
          <mesh position={[0, 0, 0]}>
            <boxGeometry args={[fT, height - fT - 0.024, fD]} />
            <meshStandardMaterial color="#334155" roughness={0.4} metalness={0.25} />
          </mesh>

          {/* Left Door Leaf (different opacity than wall) */}
          {(() => {
            const leafW = (width - 3 * fT) / 2
            const leafH = height - fT - 0.03
            const leftCenterX = -fT / 2 - leafW / 2
            const rightCenterX = fT / 2 + leafW / 2
            return (
              <>
                <mesh position={[leftCenterX, -0.012, 0]}>
                  <boxGeometry args={[leafW, leafH, 0.035]} />
                  <meshStandardMaterial
                    color="#475569"
                    transparent
                    opacity={0.72}
                    roughness={0.4}
                    metalness={0.2}
                  />
                </mesh>
                {/* Left Door Handle */}
                <mesh position={[leftCenterX + leafW / 2 - 0.05, -0.05, handleZ]}>
                  <boxGeometry args={[0.08, 0.02, 0.02]} />
                  <meshStandardMaterial color="#cbd5e1" roughness={0.2} metalness={0.9} />
                </mesh>

                {/* Right Door Leaf */}
                <mesh position={[rightCenterX, -0.012, 0]}>
                  <boxGeometry args={[leafW, leafH, 0.035]} />
                  <meshStandardMaterial
                    color="#475569"
                    transparent
                    opacity={0.72}
                    roughness={0.4}
                    metalness={0.2}
                  />
                </mesh>
                {/* Right Door Handle */}
                <mesh position={[rightCenterX - leafW / 2 + 0.05, -0.05, handleZ]}>
                  <boxGeometry args={[0.08, 0.02, 0.02]} />
                  <meshStandardMaterial color="#cbd5e1" roughness={0.2} metalness={0.9} />
                </mesh>
              </>
            )
          })()}
        </>
      ) : (
        // Standard single door
        <>
          {(() => {
            const leafW = width - 2 * fT
            const leafH = height - fT - 0.03
            const handleX = flipHinge ? -leafW / 2 + 0.07 : leafW / 2 - 0.07
            return (
              <>
                <mesh position={[0, -0.012, 0]}>
                  <boxGeometry args={[leafW, leafH, 0.035]} />
                  <meshStandardMaterial
                    color="#475569"
                    transparent
                    opacity={0.72}
                    roughness={0.4}
                    metalness={0.2}
                  />
                </mesh>
                {/* Single Door Handle */}
                <mesh position={[handleX, -0.05, handleZ]}>
                  <boxGeometry args={[0.08, 0.02, 0.02]} />
                  <meshStandardMaterial color="#cbd5e1" roughness={0.2} metalness={0.9} />
                </mesh>
              </>
            )
          })()}
        </>
      )}
    </group>
  )
}


// Key light with a shadow frustum fitted to the plan, plus soft sky/fill lights
function SceneLighting({ bounds }: { bounds: PlanBounds }) {
  const keyRef = useRef<THREE.DirectionalLight>(null)
  const centerX = (bounds.minX + bounds.maxX) / 2
  const centerZ = (bounds.minZ + bounds.maxZ) / 2
  const radius = Math.hypot(bounds.maxX - bounds.minX, bounds.maxZ - bounds.minZ) / 2 + 3

  useEffect(() => {
    const light = keyRef.current
    if (!light) return
    light.target.position.set(centerX, 0, centerZ)
    light.target.updateMatrixWorld()
    const cam = light.shadow.camera
    cam.left = -radius
    cam.right = radius
    cam.top = radius
    cam.bottom = -radius
    cam.near = 1
    cam.far = 120
    cam.updateProjectionMatrix()
    light.shadow.needsUpdate = true
  }, [centerX, centerZ, radius])

  return (
    <>
      <hemisphereLight args={['#c7d2fe', '#0b0f19', 0.6]} />
      <ambientLight intensity={0.3} />
      <directionalLight
        ref={keyRef}
        castShadow
        position={[centerX + 6, 40, centerZ + 10]}
        intensity={1.7}
        shadow-mapSize={[1024, 1024]}
        shadow-bias={-0.0004}
        shadow-normalBias={0.03}
      />
      <directionalLight position={[centerX - 20, 25, centerZ + 15]} intensity={0.35} />
      <directionalLight position={[centerX + 20, 25, centerZ - 15]} intensity={0.25} />
    </>
  )
}

// All walls of the level merged into one volume, so overlapping corners never double-blend
function MergedWalls({ walls, bounds, solid }: { walls: WallSegment[]; bounds: PlanBounds; solid?: boolean }) {
  const geometry = useMemo(() => buildWallGeometry(walls), [walls])
  const sideMat = useMemo(() => createWallMaterial({ color: solid ? '#64748b' : '#8593a8', capAlpha: solid ? 0.9 : 0.55, opaque: solid }), [solid])
  const capMat = useMemo(() => createWallMaterial({ color: solid ? '#cbd5e1' : '#dde5f0', capAlpha: solid ? 0.9 : 0.55, opaque: solid }), [solid])

  useEffect(() => {
    for (const { uniforms } of [sideMat, capMat]) {
      uniforms.uMinZ.value = bounds.minZ
      uniforms.uMaxZ.value = bounds.maxZ
    }
  }, [bounds.minZ, bounds.maxZ, sideMat, capMat])

  useEffect(() => {
    return () => {
      geometry.sides?.dispose()
      geometry.caps?.dispose()
    }
  }, [geometry])

  useEffect(() => {
    return () => {
      sideMat.material.dispose()
      capMat.material.dispose()
    }
  }, [sideMat, capMat])

  return (
    <group>
      {geometry.sides && (
        <mesh geometry={geometry.sides} material={sideMat.material} castShadow receiveShadow renderOrder={2} />
      )}
      {geometry.caps && <mesh geometry={geometry.caps} material={capMat.material} renderOrder={3} />}
      {geometry.openings.map((op) => (
        <group key={op.key} position={[op.x, 0, op.z]} rotation={[0, op.rotY, 0]}>
          {op.type === 'window' ? (
            <group position={[0, 1.5, 0]}>
              <Window3DMesh width={op.width} thickness={op.thickness} />
            </group>
          ) : op.hideDoor ? null : (
            <group position={[0, 1.05, 0]}>
              <Door3DMesh
                width={op.width}
                thickness={op.thickness}
                flipSide={op.flipSide}
                flipHinge={op.flipHinge}
              />
            </group>
          )}
        </group>
      ))}
    </group>
  )
}

// Detect closed interior planar faces (rooms) from connected wall segments
function findClosedRoomsFromWalls(walls: WallSegment[]): { id: string; points: { x: number; y: number }[] }[] {
  if (!walls || walls.length < 3) return []

  const SNAP_DIST = 10
  const vertices: { x: number; y: number }[] = []

  function getVertexId(x: number, y: number): number {
    for (let i = 0; i < vertices.length; i++) {
      if (Math.hypot(vertices[i].x - x, vertices[i].y - y) <= SNAP_DIST) {
        return i
      }
    }
    vertices.push({ x, y })
    return vertices.length - 1
  }

  // Build undirected adjacency list
  const adj = new Map<number, Set<number>>()
  for (const w of walls) {
    const v1 = getVertexId(w.x1, w.y1)
    const v2 = getVertexId(w.x2, w.y2)
    if (v1 === v2) continue
    if (!adj.has(v1)) adj.set(v1, new Set())
    if (!adj.has(v2)) adj.set(v2, new Set())
    adj.get(v1)!.add(v2)
    adj.get(v2)!.add(v1)
  }

  // Extract outgoing directed edges for each vertex with angle
  const outgoing = new Map<number, { to: number; angle: number }[]>()
  for (const [u, neighbors] of adj.entries()) {
    const uPt = vertices[u]
    const list: { to: number; angle: number }[] = []
    for (const v of neighbors) {
      const vPt = vertices[v]
      list.push({ to: v, angle: Math.atan2(vPt.y - uPt.y, vPt.x - uPt.x) })
    }
    outgoing.set(u, list)
  }

  // Planar face traversal tracing interior faces
  const visited = new Set<string>()
  const rooms: { id: string; points: { x: number; y: number }[] }[] = []

  for (const [u, edges] of outgoing.entries()) {
    for (const { to: v } of edges) {
      const edgeKey = `${u}->${v}`
      if (visited.has(edgeKey)) continue

      const path: number[] = [u]
      let currU = u
      let currV = v
      let closed = false

      for (let step = 0; step < 60; step++) {
        visited.add(`${currU}->${currV}`)
        path.push(currV)

        const outEdges = outgoing.get(currV)
        if (!outEdges || outEdges.length === 0) break

        const inAngle = Math.atan2(vertices[currU].y - vertices[currV].y, vertices[currU].x - vertices[currV].x)
        let bestEdge = null
        let minDiff = Infinity
        for (const edge of outEdges) {
          let diff = inAngle - edge.angle
          while (diff <= 1e-6) diff += 2 * Math.PI
          if (diff < minDiff) {
            minDiff = diff
            bestEdge = edge
          }
        }
        if (!bestEdge) break

        const nextV = bestEdge.to
        if (nextV === v && currV === u) {
          closed = true
          break
        }
        if (currV === u) {
          closed = true
          break
        }

        currU = currV
        currV = nextV
      }

      if (closed && path.length >= 4) {
        const uniqueVertices = path.slice(0, -1)
        if (uniqueVertices.length >= 3) {
          const pts = uniqueVertices.map((idx) => vertices[idx])
          let area = 0
          for (let i = 0; i < pts.length; i++) {
            const j = (i + 1) % pts.length
            area += pts[i].x * pts[j].y - pts[j].x * pts[i].y
          }
          area = area / 2

          // Positive area in screen coordinates corresponds to interior faces
          if (area > 150 && area < 500000) {
            rooms.push({
              id: `room-${rooms.length}-${Math.round(pts[0].x)}-${Math.round(pts[0].y)}`,
              points: pts,
            })
          }
        }
      }
    }
  }

  // Deduplicate overlapping room polygons
  const uniqueRooms: { id: string; points: { x: number; y: number }[] }[] = []
  for (const r of rooms) {
    let cx = 0
    let cy = 0
    for (const p of r.points) {
      cx += p.x
      cy += p.y
    }
    cx /= r.points.length
    cy /= r.points.length

    const exists = uniqueRooms.some((existing) => {
      let ecx = 0
      let ecy = 0
      for (const p of existing.points) {
        ecx += p.x
        ecy += p.y
      }
      ecx /= existing.points.length
      ecy /= existing.points.length
      return Math.hypot(cx - ecx, cy - ecy) < 20
    })
    if (!exists) {
      uniqueRooms.push(r)
    }
  }

  return uniqueRooms
}

// 3D Solid Interior Floor for Wall-Enclosed Rooms
function ClosedRoomMesh({ points }: { points: { x: number; y: number }[] }) {
  const shape = useMemo(() => {
    if (!points || points.length < 3) return null
    const s = new THREE.Shape()
    const p0 = points[0]
    s.moveTo(toWorldX(p0.x), -toWorldZ(p0.y))
    for (let i = 1; i < points.length; i++) {
      const p = points[i]
      s.lineTo(toWorldX(p.x), -toWorldZ(p.y))
    }
    s.closePath()
    return s
  }, [points])

  const outlineGeometry = useMemo(() => {
    if (!points || points.length < 3) return null
    const pts = points.map(
      (p) => new THREE.Vector3(toWorldX(p.x), 0.022, toWorldZ(p.y))
    )
    pts.push(pts[0])
    return new THREE.BufferGeometry().setFromPoints(pts)
  }, [points])

  const lineMesh = useMemo(() => {
    if (!outlineGeometry) return null
    const mat = new THREE.LineBasicMaterial({ color: '#334155', transparent: true, opacity: 0.85 })
    return new THREE.Line(outlineGeometry, mat)
  }, [outlineGeometry])

  if (!shape) return null

  return (
    <group>
      {/* Opaque solid interior floor slab */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.018, 0]} receiveShadow>
        <shapeGeometry args={[shape]} />
        <meshStandardMaterial
          color="#1e293b"
          roughness={0.65}
          metalness={0.08}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Perimeter boundary line */}
      {lineMesh && <primitive object={lineMesh} />}
    </group>
  )
}

// Tiled floor slab tinted with the zone color; the editable label is rendered in the 2D editor.
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

  const floorColor = useMemo(() => {
    const c = new THREE.Color(zone.color)
    return c.lerp(new THREE.Color('#1e293b'), 0.4)
  }, [zone.color])

  const tileTexture = useMemo(() => getFloorTileTexture(), [])

  const outlineGeometry = useMemo(() => {
    if (!zone.points || zone.points.length < 3) return null
    const pts = zone.points.map(
      (p) => new THREE.Vector3(toWorldX(p.x), 0.03, toWorldZ(p.y))
    )
    pts.push(pts[0])
    return new THREE.BufferGeometry().setFromPoints(pts)
  }, [zone.points])

  const lineMesh = useMemo(() => {
    if (!outlineGeometry) return null
    const mat = new THREE.LineBasicMaterial({ color: zone.color, transparent: true, opacity: 0.95 })
    return new THREE.Line(outlineGeometry, mat)
  }, [outlineGeometry, zone.color])

  if (!shape) return null

  return (
    <group>
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.02, 0]} receiveShadow>
        <shapeGeometry args={[shape]} />
        <meshStandardMaterial
          color={floorColor}
          map={tileTexture ?? undefined}
          roughness={0.85}
          metalness={0.02}
          side={THREE.DoubleSide}
        />
      </mesh>

      {lineMesh && <primitive object={lineMesh} />}
    </group>
  )
}

export function IsometricScene({
  plan,
  pan,
  viewAngle = 0.6,
  placements = [],
  deviceMap = {},
  pendingDevices = {},
  onToggleDevice = () => {},
  activeLayer,
  layers = [],
  onlyLights = false,
  interactiveDevices = true,
  showCamera = true,
  showGround = true,
  showLighting = true,
  solidWalls = false,
  lightMountHeight,
}: IsometricSceneProps) {
  const walls = plan?.walls || []
  const zones = plan?.zones || []

  // Build a map of layer_name -> hide_gauges
  const layerHideGaugesMap = useMemo(() => {
    const map: Record<string, boolean> = {}
    for (const layer of layers) {
      map[layer.name] = layer.hide_gauges
    }
    return map
  }, [layers])

  // Ceiling visibility follows the active (viewed) layer: each layer's
  // hide_gauges flag controls its own view independently.
  const ceilingHiddenForActiveLayer = useMemo(() => {
    return isCeilingHiddenForActiveLayer(activeLayer, layerHideGaugesMap)
  }, [activeLayer, layerHideGaugesMap])

  // Filter placements by active layer (if specified)
  const visiblePlacements = useMemo(() => {
    if (!activeLayer) return placements
    return placements.filter((p) => (p.layer || 'controls') === activeLayer)
  }, [placements, activeLayer])

  const renderedPlacements = useMemo(() => {
    if (!onlyLights) return visiblePlacements
    return visiblePlacements.filter((placement) => {
      const domain = placement.render_domain || deviceMap[placement.device_id]?.domain || placement.device_id.split('.')[0]
      return domain === 'light'
    })
  }, [deviceMap, onlyLights, visiblePlacements])

  // Bounding box dimensions of the plan in 3D world space
  const planBounds = useMemo(() => {
    if (!plan || (plan.walls.length === 0 && plan.zones.length === 0)) {
      return { minX: -10, maxX: 10, minZ: -7, maxZ: 7 }
    }
    let minX = Infinity
    let maxX = -Infinity
    let minZ = Infinity
    let maxZ = -Infinity
    for (const w of plan.walls) {
      minX = Math.min(minX, toWorldX(w.x1), toWorldX(w.x2))
      maxX = Math.max(maxX, toWorldX(w.x1), toWorldX(w.x2))
      minZ = Math.min(minZ, toWorldZ(w.y1), toWorldZ(w.y2))
      maxZ = Math.max(maxZ, toWorldZ(w.y1), toWorldZ(w.y2))
    }
    for (const z of plan.zones) {
      for (const p of z.points) {
        minX = Math.min(minX, toWorldX(p.x))
        maxX = Math.max(maxX, toWorldX(p.x))
        minZ = Math.min(minZ, toWorldZ(p.y))
        maxZ = Math.max(maxZ, toWorldZ(p.y))
      }
    }
    for (const p of placements) {
      const wx = toWorldX(p.x)
      const wz = toWorldZ(p.y)
      minX = Math.min(minX, wx)
      maxX = Math.max(maxX, wx)
      minZ = Math.min(minZ, wz)
      maxZ = Math.max(maxZ, wz)
    }
    if (!isFinite(minX)) return { minX: -10, maxX: 10, minZ: -7, maxZ: 7 }

    // Expand by 0.6 units to include outer wall faces and cylinder corners
    const wallPadding = 0.6
    return {
      minX: minX - wallPadding,
      maxX: maxX + wallPadding,
      minZ: minZ - wallPadding,
      maxZ: maxZ + wallPadding,
    }
  }, [plan, placements])

  // Detect closed room volumes from walls
  const detectedRooms = useMemo(() => {
    return findClosedRoomsFromWalls(plan?.walls || [])
  }, [plan?.walls])

  return (
    <>
      {showCamera && <PerspectiveSceneCamera bounds={planBounds} viewAngle={viewAngle} />}

      {showLighting && <SceneLighting bounds={planBounds} />}

      {/* Exterior ground plane */}
      {showGround && (
        <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.01, 0]} receiveShadow>
          <planeGeometry args={[2000, 2000]} />
          <meshStandardMaterial color="#0b0f19" roughness={0.95} />
        </mesh>
      )}

      {/* House Content */}
      <group position={[pan?.x || 0, 0, pan?.z || 0]}>
        {/* Closed Rooms Floor Slabs (detected from walls) */}
        {detectedRooms.map((room) => (
          <ClosedRoomMesh key={room.id} points={room.points} />
        ))}

        {/* Zone Floor Meshes */}
        {zones.map((zone) => (
          <ZoneMesh key={zone.id} zone={zone} />
        ))}

        {/* 3D Ceiling Displays at y = WALL_HEIGHT (hidden when the active layer has hide_gauges=true) */}
        {zones.map((zone) => {
          if (onlyLights || ceilingHiddenForActiveLayer || !hasConfiguredSensors(zone)) return null
          return (
            <ZoneCeilingDisplay
              key={`ceiling-${zone.id}`}
              zone={zone}
              deviceMap={deviceMap}
            />
          )
        })}

        <MergedWalls walls={walls} bounds={planBounds} solid={solidWalls} />

        {/* 3D Device Badges, Sensors & Light Halos */}
        {renderedPlacements.map((placement) => (
          <DeviceBadge3D
            key={placement.id}
            placement={placement}
            device={deviceMap[placement.device_id]}
            isPending={!!pendingDevices[placement.device_id]}
            onToggle={onToggleDevice}
            interactive={interactiveDevices}
            ceilingY={WALL_HEIGHT}
            walls={walls}
            lightMountHeight={lightMountHeight}
          />
        ))}
      </group>
    </>
  )
}
