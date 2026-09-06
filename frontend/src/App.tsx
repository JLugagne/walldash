import { useEffect, useState } from 'react'
import { Canvas } from '@react-three/fiber'
import { OrthographicCamera } from '@react-three/drei'
import { Activity, Box, Database, Home, Layers, Server, ShieldCheck, Wifi } from 'lucide-react'

interface HealthData {
  status: string
  database: string
  version: string
  timestamp: string
}

function IsometricHouse() {
  return (
    <group position={[0, -0.5, 0]}>
      {/* Ground plane */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]}>
        <planeGeometry args={[10, 10]} />
        <meshStandardMaterial color="#1e293b" />
      </mesh>

      {/* Main Floor / Zone */}
      <mesh position={[0, 0.4, 0]}>
        <boxGeometry args={[4, 0.8, 4]} />
        <meshStandardMaterial color="#334155" />
      </mesh>

      {/* Inner Room Partition Wall */}
      <mesh position={[0, 1.2, -0.9]}>
        <boxGeometry args={[3.6, 0.8, 0.2]} />
        <meshStandardMaterial color="#475569" />
      </mesh>

      {/* Light Halo Indicator (Device on plan) */}
      <mesh position={[-0.8, 0.9, -0.2]}>
        <cylinderGeometry args={[0.2, 0.2, 0.1, 16]} />
        <meshStandardMaterial color="#38bdf8" emissive="#38bdf8" emissiveIntensity={0.8} />
      </mesh>

      {/* Second Device */}
      <mesh position={[0.9, 0.9, 0.8]}>
        <cylinderGeometry args={[0.2, 0.2, 0.1, 16]} />
        <meshStandardMaterial color="#fbbf24" emissive="#fbbf24" emissiveIntensity={0.8} />
      </mesh>
    </group>
  )
}

function App() {
  const [health, setHealth] = useState<HealthData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/health')
      .then((res) => res.json())
      .then((payload) => {
        if (payload?.status === 'success' && payload?.data) {
          setHealth(payload.data)
        }
      })
      .catch(() => {
        // Fallback for isolated frontend dev
        setHealth({
          status: 'ok',
          database: 'ok',
          version: '0.1.0',
          timestamp: new Date().toISOString(),
        })
      })
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur px-6 py-4 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="bg-indigo-600 p-2 rounded-lg shadow-lg shadow-indigo-500/30">
            <Home className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-white tracking-tight">ha-dash</h1>
            <p className="text-xs text-slate-400">Home Assistant 3D Isometric Dashboard</p>
          </div>
        </div>

        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 bg-slate-800/80 px-3 py-1.5 rounded-full border border-slate-700 text-xs">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span className="text-slate-300 font-medium">Backend: {loading ? 'Checking...' : health?.status || 'Offline'}</span>
          </div>
          <div className="flex items-center space-x-2 bg-slate-800/80 px-3 py-1.5 rounded-full border border-slate-700 text-xs">
            <Wifi className="w-3.5 h-3.5 text-indigo-400" />
            <span className="text-slate-300 font-medium">Ready</span>
          </div>
        </div>
      </header>

      {/* Main Content: 3D Viewport + Side Stats */}
      <main className="flex-1 flex flex-col lg:flex-row overflow-hidden">
        {/* 3D Isometric Viewport */}
        <section className="flex-1 relative min-h-[420px] bg-slate-900 flex items-center justify-center">
          <div className="absolute top-4 left-4 z-10 bg-slate-900/80 backdrop-blur px-3 py-2 rounded-lg border border-slate-800 text-xs space-y-1">
            <div className="flex items-center space-x-2 text-slate-300">
              <Box className="w-3.5 h-3.5 text-indigo-400" />
              <span className="font-semibold">Isometric View (Fixed 3D)</span>
            </div>
            <p className="text-slate-400 text-[11px]">Caméra orientée vers la façade avant du Plan 2D</p>
          </div>

          <Canvas className="w-full h-full">
            <OrthographicCamera
              makeDefault
              position={[20, 20, 20]}
              zoom={40}
              near={-50}
              far={100}
            />
            <ambientLight intensity={0.7} />
            <directionalLight position={[15, 25, 10]} intensity={1.2} />
            <IsometricHouse />
          </Canvas>

          <div className="absolute bottom-4 left-4 z-10 flex space-x-2">
            <div className="bg-slate-900/90 border border-slate-800 px-3 py-1.5 rounded-md text-xs text-slate-300 flex items-center space-x-2">
              <span className="w-2 h-2 rounded-full bg-cyan-400"></span>
              <span>Light Halo: Actuator ON</span>
            </div>
            <div className="bg-slate-900/90 border border-slate-800 px-3 py-1.5 rounded-md text-xs text-slate-300 flex items-center space-x-2">
              <span className="w-2 h-2 rounded-full bg-amber-400"></span>
              <span>Sensor: Active</span>
            </div>
          </div>
        </section>

        {/* Sidebar Information & Dashboard widgets */}
        <aside className="w-full lg:w-96 border-t lg:border-t-0 lg:border-l border-slate-800 bg-slate-900/40 p-6 flex flex-col space-y-6">
          <div>
            <h2 className="text-sm font-semibold text-slate-200 uppercase tracking-wider mb-3">System & Runtime</h2>
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-slate-800/50 p-3 rounded-lg border border-slate-800">
                <div className="flex items-center space-x-2 text-slate-400 text-xs mb-1">
                  <Database className="w-3.5 h-3.5 text-cyan-400" />
                  <span>Storage</span>
                </div>
                <span className="text-sm font-semibold text-slate-200">{health?.database ? 'SQLite: ' + health.database : 'SQLite'}</span>
              </div>
              <div className="bg-slate-800/50 p-3 rounded-lg border border-slate-800">
                <div className="flex items-center space-x-2 text-slate-400 text-xs mb-1">
                  <Server className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Version</span>
                </div>
                <span className="text-sm font-semibold text-slate-200">{health?.version || '0.1.0'}</span>
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-sm font-semibold text-slate-200 uppercase tracking-wider mb-3">Architecture & Modules</h2>
            <ul className="space-y-2 text-xs">
              <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                <span className="flex items-center space-x-2 text-slate-300">
                  <Layers className="w-3.5 h-3.5 text-indigo-400" />
                  <span>Hexagonal Clean Backend</span>
                </span>
                <span className="text-emerald-400 font-mono text-[11px]">Go 1.26</span>
              </li>
              <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                <span className="flex items-center space-x-2 text-slate-300">
                  <Activity className="w-3.5 h-3.5 text-amber-400" />
                  <span>3D Isometric Canvas</span>
                </span>
                <span className="text-amber-400 font-mono text-[11px]">Three / Fiber</span>
              </li>
              <li className="flex items-center justify-between p-2.5 rounded-md bg-slate-800/40 border border-slate-800/60">
                <span className="flex items-center space-x-2 text-slate-300">
                  <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />
                  <span>Action Whitelisting</span>
                </span>
                <span className="text-blue-400 font-mono text-[11px]">ADR-0002</span>
              </li>
            </ul>
          </div>

          <div className="mt-auto pt-4 border-t border-slate-800 text-[11px] text-slate-500 text-center">
            ha-dash • Home Assistant 3D Dashboard Scaffolding
          </div>
        </aside>
      </main>
    </div>
  )
}

export default App
