import * as THREE from 'three'

let tileTexture: THREE.Texture | null = null
let glowTexture: THREE.Texture | null = null

// Subtle tile pattern multiplied with the zone color; one texture repeat equals one tile
export function getFloorTileTexture(): THREE.Texture | null {
  if (tileTexture) return tileTexture
  if (typeof document === 'undefined') return null
  const size = 256
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) return null

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, size, size)

  const grad = ctx.createLinearGradient(0, 0, size, size)
  grad.addColorStop(0, 'rgba(0,0,0,0)')
  grad.addColorStop(1, 'rgba(0,0,0,0.07)')
  ctx.fillStyle = grad
  ctx.fillRect(0, 0, size, size)

  ctx.strokeStyle = 'rgba(0,0,0,0.16)'
  ctx.lineWidth = 3
  ctx.strokeRect(1.5, 1.5, size - 3, size - 3)

  ctx.strokeStyle = 'rgba(255,255,255,0.35)'
  ctx.lineWidth = 2
  ctx.strokeRect(5, 5, size - 10, size - 10)

  const texture = new THREE.CanvasTexture(canvas)
  texture.wrapS = THREE.RepeatWrapping
  texture.wrapT = THREE.RepeatWrapping
  texture.repeat.set(1 / 0.6, 1 / 0.6)
  texture.anisotropy = 4
  texture.colorSpace = THREE.SRGBColorSpace
  tileTexture = texture
  return texture
}

// Radial white-to-transparent disc used for additive light pools on the floor
export function getRadialGlowTexture(): THREE.Texture | null {
  if (glowTexture) return glowTexture
  if (typeof document === 'undefined') return null
  const size = 256
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) return null
  const grad = ctx.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2)
  grad.addColorStop(0, 'rgba(255,255,255,1)')
  grad.addColorStop(0.35, 'rgba(255,255,255,0.45)')
  grad.addColorStop(1, 'rgba(255,255,255,0)')
  ctx.fillStyle = grad
  ctx.fillRect(0, 0, size, size)
  const texture = new THREE.CanvasTexture(canvas)
  texture.colorSpace = THREE.SRGBColorSpace
  glowTexture = texture
  return texture
}

// Picks the interior point farthest from any edge, so labels stay inside L-shaped rooms
export function poleOfInaccessibility(points: { x: number; y: number }[]): { x: number; y: number } {
  if (points.length === 0) return { x: 0, y: 0 }
  let minX = Infinity
  let maxX = -Infinity
  let minY = Infinity
  let maxY = -Infinity
  for (const p of points) {
    minX = Math.min(minX, p.x)
    maxX = Math.max(maxX, p.x)
    minY = Math.min(minY, p.y)
    maxY = Math.max(maxY, p.y)
  }
  const centroid = {
    x: points.reduce((s, p) => s + p.x, 0) / points.length,
    y: points.reduce((s, p) => s + p.y, 0) / points.length,
  }
  if (points.length < 3) return centroid

  const inside = (x: number, y: number): boolean => {
    let result = false
    for (let i = 0, j = points.length - 1; i < points.length; j = i++) {
      const pi = points[i]
      const pj = points[j]
      if (pi.y > y !== pj.y > y && x < ((pj.x - pi.x) * (y - pi.y)) / (pj.y - pi.y) + pi.x) {
        result = !result
      }
    }
    return result
  }

  const edgeDistance = (x: number, y: number): number => {
    let best = Infinity
    for (let i = 0, j = points.length - 1; i < points.length; j = i++) {
      const a = points[j]
      const b = points[i]
      const dx = b.x - a.x
      const dy = b.y - a.y
      const lenSq = dx * dx + dy * dy
      let t = lenSq > 0 ? ((x - a.x) * dx + (y - a.y) * dy) / lenSq : 0
      t = Math.max(0, Math.min(1, t))
      best = Math.min(best, Math.hypot(x - (a.x + t * dx), y - (a.y + t * dy)))
    }
    return best
  }

  const steps = 28
  let bestPoint = inside(centroid.x, centroid.y) ? centroid : null
  let bestDist = bestPoint ? edgeDistance(centroid.x, centroid.y) : -1
  for (let i = 1; i < steps; i++) {
    for (let j = 1; j < steps; j++) {
      const x = minX + ((maxX - minX) * i) / steps
      const y = minY + ((maxY - minY) * j) / steps
      if (!inside(x, y)) continue
      const d = edgeDistance(x, y)
      if (d > bestDist) {
        bestDist = d
        bestPoint = { x, y }
      }
    }
  }
  return bestPoint || centroid
}
