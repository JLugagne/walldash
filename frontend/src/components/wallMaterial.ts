import * as THREE from 'three'
import { WALL_HEIGHT } from './wallGeometry'

export interface WallFadeUniforms {
  uWallHeight: { value: number }
  uFadeStart: { value: number }
  uMinZ: { value: number }
  uMaxZ: { value: number }
  uCapAlpha: { value: number }
  uBackAlpha: { value: number }
  uFrontAlpha: { value: number }
  uSideAlpha: { value: number }
}

export interface WallMaterialOptions {
  color: string
  capAlpha: number
}

const VARYINGS = `
varying vec3 vWallPos;
varying vec3 vWallNormal;
`

const FRAGMENT_UNIFORMS = `
uniform float uWallHeight;
uniform float uFadeStart;
uniform float uMinZ;
uniform float uMaxZ;
uniform float uCapAlpha;
uniform float uBackAlpha;
uniform float uFrontAlpha;
uniform float uSideAlpha;
`

const FADE_SNIPPET = `
float wallH = clamp(vWallPos.y / uWallHeight, 0.0, 1.0);
float facingZ = smoothstep(0.35, 0.9, abs(vWallNormal.z));
float frontness = smoothstep(uMinZ, uMaxZ, vWallPos.z);
float zMin = mix(uBackAlpha, uFrontAlpha, frontness);
float sideMin = mix(uSideAlpha, zMin, facingZ);
float isCap = step(0.9, abs(vWallNormal.y));
float minAlpha = mix(sideMin, uCapAlpha, isCap);
float fade = smoothstep(uFadeStart, 1.0, wallH);
diffuseColor.a *= mix(1.0, minAlpha, fade);
`

// Standard lit material whose alpha fades with height, stronger on faces turned toward the camera
export function createWallMaterial(options: WallMaterialOptions): {
  material: THREE.MeshStandardMaterial
  uniforms: WallFadeUniforms
} {
  const uniforms: WallFadeUniforms = {
    uWallHeight: { value: WALL_HEIGHT },
    uFadeStart: { value: 0.25 },
    uMinZ: { value: -10 },
    uMaxZ: { value: 10 },
    uCapAlpha: { value: options.capAlpha },
    uBackAlpha: { value: 0.65 },
    uFrontAlpha: { value: 0.1 },
    uSideAlpha: { value: 0.92 },
  }

  const material = new THREE.MeshStandardMaterial({
    color: options.color,
    roughness: 0.6,
    metalness: 0.05,
    transparent: true,
    depthWrite: false,
    side: THREE.FrontSide,
  })

  material.onBeforeCompile = (shader) => {
    Object.assign(shader.uniforms, uniforms)
    shader.vertexShader = shader.vertexShader
      .replace('#include <common>', `#include <common>${VARYINGS}`)
      .replace(
        '#include <begin_vertex>',
        `#include <begin_vertex>
vWallPos = (modelMatrix * vec4(transformed, 1.0)).xyz;
vWallNormal = normalize(mat3(modelMatrix) * objectNormal);`
      )
    shader.fragmentShader = shader.fragmentShader
      .replace('#include <common>', `#include <common>${VARYINGS}${FRAGMENT_UNIFORMS}`)
      .replace('#include <alphamap_fragment>', `#include <alphamap_fragment>${FADE_SNIPPET}`)
  }
  material.customProgramCacheKey = () => `wall-fade-${options.capAlpha}`

  return { material, uniforms }
}
