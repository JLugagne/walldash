#!/usr/bin/env node
// Generates the Walldash user-manual screenshots from example data.
//
// It never talks to a real Home Assistant instance: it starts the Walldash
// backend with an empty HA_URL / HA_TOKEN, which makes it serve its built-in
// example devices ("demo mode"), imports a Sweet Home 3D demo plan, seeds a few
// device placements and an overview dashboard, then drives the UI with
// Playwright and writes the screenshots to ../static/images.
//
// Usage:
//   SH3D_FILE=/path/to/plan.sh3d node generate.mjs
//   node generate.mjs --skip-build        # reuse an existing frontend/dist and backend binary
//
// Environment:
//   SH3D_FILE    Path to a .sh3d plan (default: ./demo/alps-hotel.sh3d)
//   MANUAL_PORT  Backend port for the demo server (default: 18080)

import { spawn } from 'node:child_process'
import { mkdir, readFile, rm } from 'node:fs/promises'
import { existsSync, readdirSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const MANUAL_DIR = path.resolve(__dirname, '..')
const REPO_ROOT = path.resolve(__dirname, '..', '..', '..')
const FRONTEND_DIR = path.join(REPO_ROOT, 'frontend')
const WORK_DIR = path.join(__dirname, '.work')
const IMAGES_DIR = path.join(MANUAL_DIR, 'static', 'images')
const DEMO_DIR = path.join(__dirname, 'demo')

const PORT = Number(process.env.MANUAL_PORT || '18080')
const BASE = `http://127.0.0.1:${PORT}`
const SKIP_BUILD = process.argv.includes('--skip-build')

const SH3D_FILE =
  process.env.SH3D_FILE || path.join(DEMO_DIR, 'alps-hotel.sh3d')

// ---------------------------------------------------------------------------
// Small process helpers
// ---------------------------------------------------------------------------

function run(cmd, args, opts = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, { stdio: 'inherit', ...opts })
    child.on('error', reject)
    child.on('exit', (code) =>
      code === 0 ? resolve() : reject(new Error(`${cmd} exited with code ${code}`))
    )
  })
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

// Playwright pins one Chromium build per release. On machines where that exact
// build is not downloaded but another revision is cached (common in CI and in
// Playwright dev environments), fall back to the newest cached full Chromium.
function findCachedChromium() {
  if (process.env.CHROMIUM_PATH) return process.env.CHROMIUM_PATH
  const root = path.join(os.homedir(), '.cache', 'ms-playwright')
  if (!existsSync(root)) return null
  const candidates = []
  for (const entry of readdirSync(root)) {
    if (!entry.startsWith('chromium-')) continue
    const rev = Number.parseInt(entry.split('-')[1], 10) || 0
    for (const rel of ['chrome-linux64/chrome', 'chrome-linux/chrome']) {
      const p = path.join(root, entry, rel)
      if (existsSync(p)) candidates.push({ rev, p })
    }
  }
  candidates.sort((a, b) => b.rev - a.rev)
  return candidates[0]?.p ?? null
}

const CHROMIUM_ARGS = [
  '--enable-unsafe-swiftshader',
  '--use-gl=angle',
  '--use-angle=swiftshader',
  '--ignore-gpu-blocklist',
]

async function launchBrowser() {
  try {
    return await chromium.launch({ args: CHROMIUM_ARGS })
  } catch (err) {
    const fallback = findCachedChromium()
    if (!fallback) throw err
    console.warn(`  default Chromium unavailable, using cached build: ${fallback}`)
    return await chromium.launch({ args: CHROMIUM_ARGS, executablePath: fallback })
  }
}

async function waitForHealth(timeoutMs = 20000) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${BASE}/api/health`)
      if (res.ok) return
    } catch {
      /* not up yet */
    }
    await sleep(250)
  }
  throw new Error('Walldash backend did not become healthy in time')
}

// ---------------------------------------------------------------------------
// Backend API helpers (CSRF is bypassed with X-Requested-With, as the SPA does)
// ---------------------------------------------------------------------------

async function api(pathname, { method = 'GET', body } = {}) {
  const headers = {}
  let payload
  if (method !== 'GET') headers['X-Requested-With'] = 'XMLHttpRequest'
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }
  const res = await fetch(`${BASE}${pathname}`, { method, headers, body: payload })
  const text = await res.text()
  let json
  try {
    json = JSON.parse(text)
  } catch {
    /* non-JSON */
  }
  if (!res.ok) {
    throw new Error(`${method} ${pathname} → HTTP ${res.status}: ${text.slice(0, 400)}`)
  }
  return json
}

async function importSh3d() {
  const buf = await readFile(SH3D_FILE)
  const form = new FormData()
  form.append('file', new Blob([buf]), path.basename(SH3D_FILE))
  const res = await fetch(`${BASE}/api/levels/import/sh3d`, {
    method: 'POST',
    headers: { 'X-Requested-With': 'XMLHttpRequest' },
    body: form,
  })
  const text = await res.text()
  if (!res.ok) throw new Error(`sh3d import → HTTP ${res.status}: ${text.slice(0, 400)}`)
  return JSON.parse(text).data
}

// ---------------------------------------------------------------------------
// Demo data
// ---------------------------------------------------------------------------

// Each fallback device is anchored to a distinct room of the demo plan so the
// 3D view shows one marker per room.
const DEVICE_PLACEMENTS = [
  { device: 'light.salon_plafond', zone: 'Hall', name: 'Hall Chandelier', render: 'light', layer: 'controls' },
  { device: 'light.cuisine_spot', zone: 'Kitchen', name: 'Kitchen Spotlights', render: 'light', layer: 'controls' },
  { device: 'light.chambre_chevet', zone: 'Office', name: 'Office Desk Lamp', render: 'light', layer: 'controls' },
  { device: 'switch.machine_a_cafe', zone: 'Breakfast room', name: 'Coffee Machine', render: 'switch', layer: 'controls' },
  { device: 'switch.prise_tv', zone: 'Ski shed', name: 'Ski Shed Plug', render: 'switch', layer: 'controls' },
  { device: 'sensor.temperature_salon', zone: "Owner's flat", name: 'Flat Temperature', render: 'sensor', layer: 'sensors' },
  { device: 'sensor.humidite_sdb', zone: 'Room 9', name: 'Room 9 Humidity', render: 'sensor', layer: 'sensors' },
  { device: 'climate.thermostat_salon', zone: 'Room 10', name: 'Room 10 Thermostat', render: 'climate', layer: 'controls' },
  { device: 'media_player.enceinte_salon', zone: 'Room 7', name: 'Room 7 Speaker', render: 'media_player', layer: 'controls' },
]

// Zones that get a temperature / humidity sensor bound, so the Sensors layer
// has something meaningful to display.
const ZONE_SENSOR_BINDINGS = {
  Hall: {
    temp_sensor: 'sensor.temperature_salon',
    temp_min: 18,
    temp_max: 24,
    humidity_sensor: 'sensor.humidite_sdb',
    humidity_min: 40,
    humidity_max: 65,
  },
  'Breakfast room': { temp_sensor: 'sensor.temperature_salon', temp_min: 19, temp_max: 25 },
  "Owner's flat": { temp_sensor: 'sensor.temperature_salon', temp_min: 18, temp_max: 24 },
  'Ski shed': { humidity_sensor: 'sensor.humidite_sdb', humidity_min: 40, humidity_max: 70 },
}

const AUTOMATIONS = [
  'automation.eteindre_toutes_les_lumieres',
  'automation.scenario_depart_maison',
  'automation.arrosage_automatique_jardin',
  'automation.simulation_presence',
]

// The demo plan is an Alpine hotel, so the showcase dashboards use Chamonix weather.
const WEATHER_LOCATION = { latitude: 45.9237, longitude: 6.8694, location_name: 'Chamonix' }

const HOME_WIDGETS = [
  { type: 'sensor', title: 'Hall temperature', config: { entity_ids: ['sensor.temperature_salon'], display: 'arc', min: 10, max: 30, unit: '°C' }, col: 0, row: 0, col_span: 3, row_span: 3 },
  { type: 'weather', title: 'Chamonix', config: { display: 'weather', weather_mode: 'current', units: 'metric', ...WEATHER_LOCATION }, col: 3, row: 0, col_span: 3, row_span: 3 },
  { type: 'automation_list', title: 'Scenes', config: { entity_ids: AUTOMATIONS, display: 'list' }, col: 6, row: 0, col_span: 3, row_span: 3 },
  { type: 'actuator', title: 'Hall lights', config: { entity_ids: ['light.salon_plafond'], display: 'toggle' }, col: 9, row: 0, col_span: 3, row_span: 1 },
  { type: 'actuator', title: 'Coffee machine', config: { entity_ids: ['switch.machine_a_cafe'], display: 'toggle' }, col: 9, row: 1, col_span: 3, row_span: 1 },
  { type: 'actuator', title: 'Kitchen lights', config: { entity_ids: ['light.cuisine_spot'], display: 'toggle' }, col: 9, row: 2, col_span: 3, row_span: 1 },
  { type: 'sensor', title: 'Ski shed humidity', config: { entity_ids: ['sensor.humidite_sdb'], display: 'number', unit: '%' }, col: 0, row: 3, col_span: 3, row_span: 2 },
  { type: 'sensor', title: 'Office temperature', config: { entity_ids: ['sensor.temperature_salon'], display: 'bar', min: 10, max: 30, unit: '°C' }, col: 3, row: 3, col_span: 3, row_span: 2 },
  { type: 'weather', title: '5-day forecast', config: { display: 'weather', weather_mode: 'ndays', weather_days: 5, units: 'metric', ...WEATHER_LOCATION }, col: 6, row: 3, col_span: 6, row_span: 2 },
]

const WEATHER_WIDGETS = [
  { type: 'weather', title: 'Chamonix', config: { display: 'weather', weather_mode: 'current', units: 'metric', ...WEATHER_LOCATION }, col: 0, row: 0, col_span: 4, row_span: 3 },
  { type: 'weather', title: '5-day forecast', config: { display: 'weather', weather_mode: 'ndays', weather_days: 5, units: 'metric', ...WEATHER_LOCATION }, col: 4, row: 0, col_span: 8, row_span: 3 },
  { type: 'sensor', title: 'Ski shed humidity', config: { entity_ids: ['sensor.humidite_sdb'], display: 'number', unit: '%' }, col: 0, row: 3, col_span: 4, row_span: 2 },
  { type: 'sensor', title: 'Office temperature', config: { entity_ids: ['sensor.temperature_salon'], display: 'bar', min: 10, max: 30, unit: '°C' }, col: 4, row: 3, col_span: 4, row_span: 2 },
  { type: 'automation_list', title: 'Scenes', config: { entity_ids: AUTOMATIONS, display: 'list' }, col: 8, row: 3, col_span: 4, row_span: 2 },
]

function zoneCenter(zone) {
  const n = zone.points.length
  const cx = zone.points.reduce((s, p) => s + p.x, 0) / n
  const cy = zone.points.reduce((s, p) => s + p.y, 0) / n
  return { x: Math.max(cx, 1), y: Math.max(cy, 1) }
}

async function createOverviewWithWidgets(name, order, backgroundImage, widgets) {
  const overview = (
    await api('/api/overviews', {
      method: 'POST',
      body: {
        name,
        order,
        cols: 12,
        rows: 5,
        background_image: backgroundImage,
        background_opacity: 95,
        background_blur: 12,
        background_dim: 45,
      },
    })
  ).data
  for (const w of widgets) {
    await api(`/api/overviews/${overview.id}/widgets`, {
      method: 'POST',
      body: { type: w.type, title: w.title, config: w.config, col: w.col, row: w.row, col_span: w.col_span, row_span: w.row_span },
    })
  }
  return overview
}

async function seed() {
  console.log('→ importing demo plan:', SH3D_FILE)
  const levels = await importSh3d()
  const ground = levels.find((l) => /ground/i.test(l.name)) || levels[0]
  console.log(`  created ${levels.length} levels (using "${ground.name}")`)

  // 1. Bind sensors to a few zones of the ground floor.
  const plan = (await api(`/api/levels/${ground.id}/plan`)).data
  plan.zones = plan.zones.map((z) =>
    ZONE_SENSOR_BINDINGS[z.name] ? { ...z, ...ZONE_SENSOR_BINDINGS[z.name] } : z
  )
  await api(`/api/levels/${ground.id}/plan`, { method: 'PUT', body: plan })
  console.log(`  bound sensors on ${Object.keys(ZONE_SENSOR_BINDINGS).length} zones`)

  // 2. Place example devices.
  const zonesByName = new Map(plan.zones.map((z) => [z.name, z]))
  let placed = 0
  for (const p of DEVICE_PLACEMENTS) {
    const zone = zonesByName.get(p.zone)
    if (!zone) {
      console.warn(`  ! zone "${p.zone}" not found, skipping ${p.device}`)
      continue
    }
    const { x, y } = zoneCenter(zone)
    await api(`/api/levels/${ground.id}/placements`, {
      method: 'POST',
      body: {
        device_id: p.device,
        x,
        y,
        render_domain: p.render,
        custom_name: p.name,
        layer: p.layer,
      },
    })
    placed += 1
  }
  console.log(`  placed ${placed} devices`)

  // 3. Create the showcase overview dashboards.
  const home = await createOverviewWithWidgets('Home', 0, '/backgrounds/mountains-lake.jpg', HOME_WIDGETS)
  const weather = await createOverviewWithWidgets('Weather', 1, '/backgrounds/desert-night.jpg', WEATHER_WIDGETS)
  console.log(`  created overviews "${home.name}" and "${weather.name}"`)

  return { groundId: ground.id, levels }
}

// ---------------------------------------------------------------------------
// Playwright capture
// ---------------------------------------------------------------------------

const IPAD_LANDSCAPE = { width: 1194, height: 834 }
// Screenshots are captured at CSS-pixel size: deviceScaleFactor 1 keeps the PNGs small.
const DEVICE_SCALE_FACTOR = 1

async function capture({ groundId }) {
  await mkdir(IMAGES_DIR, { recursive: true })
  const browser = await launchBrowser()
  const context = await browser.newContext({
    viewport: IPAD_LANDSCAPE,
    deviceScaleFactor: DEVICE_SCALE_FACTOR,
    isMobile: false,
    hasTouch: true,
    colorScheme: 'dark',
    locale: 'en-US',
  })

  const results = []

  async function shot(name, hash, prepare) {
    const file = path.join(IMAGES_DIR, `${name}.png`)
    let lastError = null
    for (let attempt = 1; attempt <= 2; attempt += 1) {
      const page = await context.newPage()
      try {
        await page.goto(`${BASE}/#${hash}`, { waitUntil: 'load' })
        await page.waitForSelector('canvas', { timeout: 15000 }).catch(() => {})
        if (prepare) await prepare(page)
        await page.waitForTimeout(2800)
        await page.screenshot({ path: file, timeout: 60000, animations: 'disabled' })
        await page.close()
        console.log(`  ✓ ${name}.png`)
        results.push({ name, ok: true })
        return
      } catch (err) {
        lastError = err
        await page.close().catch(() => {})
        await sleep(1000)
      }
    }
    console.warn(`  ✗ ${name}.png — ${lastError.message}`)
    results.push({ name, ok: false, error: lastError.message })
  }

  await shot('house-overview', '/')
  await shot('floor-controls', `/floor/${groundId}`)
  await shot('floor-sensors', `/floor/${groundId}`, async (page) => {
    const sensors = page.locator('[aria-label="Layer Sensors"]').first()
    await sensors.waitFor({ timeout: 8000 })
    await sensors.click()
  })
  await shot('plan-editor', `/admin/${groundId}`)
  await shot('dashboard', '/overviews')
  await shot('dashboard-edit', '/overviews', async (page) => {
    const menu = page.locator('[title="Overview menu"]').first()
    await menu.waitFor({ timeout: 8000 })
    await menu.click()
    await page.getByRole('menuitem', { name: 'Admin mode' }).click()
    await menu.click()
    await page.getByRole('menuitem', { name: 'Edit layout' }).click()
  })
  await shot('setup-devices', '/setup/devices')
  await shot('setup-dashboards', '/setup/dashboards')
  await shot('setup-settings', '/setup/settings')

  await browser.close()

  const failed = results.filter((r) => !r.ok)
  if (failed.length) {
    throw new Error(`failed screenshots: ${failed.map((f) => f.name).join(', ')}`)
  }
}

// ---------------------------------------------------------------------------
// Orchestration
// ---------------------------------------------------------------------------

let backend = null

function stopBackend() {
  if (backend && !backend.killed) {
    backend.kill('SIGTERM')
    backend = null
  }
}

async function main() {
  if (!existsSync(SH3D_FILE)) {
    throw new Error(
      `Sweet Home 3D file not found: ${SH3D_FILE}\n` +
        `Set SH3D_FILE=/path/to/plan.sh3d or drop a plan at tools/demo/alps-hotel.sh3d.`
    )
  }
  if (!existsSync(FRONTEND_DIR)) throw new Error(`frontend not found at ${FRONTEND_DIR}`)

  await mkdir(WORK_DIR, { recursive: true })
  const binPath = path.join(WORK_DIR, 'walldash')
  const dbPath = path.join(WORK_DIR, 'manual.db')

  if (!SKIP_BUILD) {
    console.log('→ building backend…')
    await run('go', ['build', '-o', binPath, './cmd'], { cwd: REPO_ROOT })
    console.log('→ building frontend…')
    await run('npm', ['run', 'build'], { cwd: FRONTEND_DIR })
  }

  console.log('→ starting demo backend (no Home Assistant configured)…')
  await rm(dbPath, { force: true })

  backend = spawn(binPath, [], {
    cwd: WORK_DIR, // deliberately isolated so no repository .env is loaded
    env: {
      ...process.env,
      HA_URL: '',
      HA_TOKEN: '',
      PORT: String(PORT),
      DB_PATH: dbPath,
      LOG_LEVEL: 'warn',
      FRONTEND_DIR: path.join(FRONTEND_DIR, 'dist'),
    },
    stdio: ['ignore', 'inherit', 'inherit'],
  })

  process.on('SIGINT', () => {
    stopBackend()
    process.exit(130)
  })
  process.on('SIGTERM', () => {
    stopBackend()
    process.exit(143)
  })

  await waitForHealth()
  console.log('  backend is up')

  // Screenshot the onboarding wizard before any data exists.
  await captureEmptyOnboarding()

  const ctx = await seed()
  await capture(ctx)

  stopBackend()
  console.log('\nDone. Screenshots written to docs/user-manual/static/images/')
}

// Captures the onboarding wizard on a freshly initialised (empty) database.
async function captureEmptyOnboarding() {
  await mkdir(IMAGES_DIR, { recursive: true })
  const browser = await launchBrowser()
  const context = await browser.newContext({
    viewport: IPAD_LANDSCAPE,
    deviceScaleFactor: DEVICE_SCALE_FACTOR,
    hasTouch: true,
    colorScheme: 'dark',
    locale: 'en-US',
  })
  const page = await context.newPage()
  await page.goto(`${BASE}/`, { waitUntil: 'load' })
  await page.getByText('Welcome to Walldash').waitFor({ timeout: 10000 })
  await page.waitForTimeout(600)
  await page.screenshot({ path: path.join(IMAGES_DIR, 'onboarding.png') })
  await browser.close()
  console.log('  ✓ onboarding.png')
}

main().catch((err) => {
  stopBackend()
  console.error('\nGeneration failed:', err.message)
  process.exit(1)
})
