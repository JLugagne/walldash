import {
  AppWindow,
  BrickWall,
  Cpu,
  DoorOpen,
  Flame,
  Hand,
  Hexagon,
  Lightbulb,
  MousePointer2,
  Music,
  Power,
  Thermometer,
  type LucideIcon,
} from 'lucide-react'
import { cmToUnits } from './geometry'

export type ToolMode = 'select' | 'wall' | 'door' | 'window' | 'zone' | 'pan'

export interface ToolDef {
  key: ToolMode
  label: string
  icon: LucideIcon
  shortcut: string
}

export const TOOLS: ToolDef[] = [
  { key: 'select', label: 'Selection', icon: MousePointer2, shortcut: 'V' },
  { key: 'wall', label: 'Wall', icon: BrickWall, shortcut: 'W' },
  { key: 'door', label: 'Door', icon: DoorOpen, shortcut: 'D' },
  { key: 'window', label: 'Window', icon: AppWindow, shortcut: 'F' },
  { key: 'zone', label: 'Zone', icon: Hexagon, shortcut: 'Z' },
  { key: 'pan', label: 'Move view', icon: Hand, shortcut: 'H' },
]

export const GRID_SIZES = [10, 20, 40]

export interface SizePreset {
  label: string
  value: number
}

export const WALL_THICKNESS_PRESETS: SizePreset[] = [
  { label: 'Partition', value: cmToUnits(15) },
  { label: 'Indoor', value: cmToUnits(20) },
  { label: 'Load-bearing', value: cmToUnits(30) },
  { label: 'Outdoor', value: cmToUnits(40) },
  { label: 'Thick', value: cmToUnits(50) },
]

export const DOOR_WIDTH_PRESETS: SizePreset[] = [
  { label: '70', value: cmToUnits(70) },
  { label: '80', value: cmToUnits(80) },
  { label: '90', value: cmToUnits(90) },
  { label: '120', value: cmToUnits(120) },
  { label: '160', value: cmToUnits(160) },
]

export const WINDOW_WIDTH_PRESETS: SizePreset[] = [
  { label: '60', value: cmToUnits(60) },
  { label: '90', value: cmToUnits(90) },
  { label: '120', value: cmToUnits(120) },
  { label: '160', value: cmToUnits(160) },
  { label: '200', value: cmToUnits(200) },
]

export const WALL_LENGTH_PRESETS: SizePreset[] = [
  { label: '2 m', value: 80 },
  { label: '3 m', value: 120 },
  { label: '4 m', value: 160 },
  { label: '5 m', value: 200 },
  { label: '6 m', value: 240 },
]

export const ZONE_COLOR_PRESETS = [
  { name: 'Living room', value: '#3b82f6' },
  { name: 'Kitchen', value: '#f59e0b' },
  { name: 'Bedroom', value: '#8b5cf6' },
  { name: 'Bathroom', value: '#06b6d4' },
  { name: 'Office', value: '#10b981' },
  { name: 'Hallway', value: '#64748b' },
  { name: 'Terrace', value: '#059669' },
  { name: 'Garage', value: '#ec4899' },
]

export const DOMAIN_CATEGORIES = [
  { key: 'all', label: 'All' },
  { key: 'light', label: 'Lights' },
  { key: 'switch', label: 'Outlets' },
  { key: 'sensor', label: 'Sensors' },
  { key: 'climate', label: 'Climate' },
  { key: 'media_player', label: 'Media' },
]

export const RENDER_DOMAIN_OPTIONS = [
  { key: 'light', label: 'Light' },
  { key: 'switch', label: 'Switch' },
  { key: 'media_player', label: 'Media' },
  { key: 'climate', label: 'Climate' },
  { key: 'sensor', label: 'Sensor' },
]

export interface DomainStyle {
  fill: string
  ring: string
  text: string
  bg: string
  border: string
  Icon: LucideIcon
}

const DOMAIN_STYLES: Record<string, DomainStyle> = {
  light: { fill: '#f59e0b', ring: '#fbbf24', text: 'text-amber-400', bg: 'bg-amber-500/10', border: 'border-amber-500/30', Icon: Lightbulb },
  switch: { fill: '#06b6d4', ring: '#22d3ee', text: 'text-cyan-400', bg: 'bg-cyan-500/10', border: 'border-cyan-500/30', Icon: Power },
  sensor: { fill: '#10b981', ring: '#34d399', text: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/30', Icon: Thermometer },
  climate: { fill: '#f43f5e', ring: '#fb7185', text: 'text-rose-400', bg: 'bg-rose-500/10', border: 'border-rose-500/30', Icon: Flame },
  media_player: { fill: '#a855f7', ring: '#c084fc', text: 'text-purple-400', bg: 'bg-purple-500/10', border: 'border-purple-500/30', Icon: Music },
}

const DEFAULT_DOMAIN_STYLE: DomainStyle = {
  fill: '#6366f1',
  ring: '#818cf8',
  text: 'text-indigo-400',
  bg: 'bg-indigo-500/10',
  border: 'border-indigo-500/30',
  Icon: Cpu,
}

export function domainStyle(domain: string): DomainStyle {
  return DOMAIN_STYLES[domain] ?? DEFAULT_DOMAIN_STYLE
}

export function isDeviceActive(state: string | undefined): boolean {
  if (!state) return false
  return state !== 'off' && state !== 'idle' && state !== 'unavailable' && state !== 'unknown'
}

/** Canvas palette shared by every SVG layer so the editor reads as one surface. */
export const CANVAS = {
  background: '#0b1120',
  gridLine: '#1e293b',
  gridDot: '#334155',
  axis: '#334155',
  wallFill: '#94a3b8',
  wallOutline: '#0f172a',
  wallHover: '#c7d2e3',
  wallSelected: '#818cf8',
  wallSelectedOutline: '#4338ca',
  accent: '#6366f1',
  accentSoft: '#a5b4fc',
  handleFill: '#ffffff',
  joint: '#475569',
  doorGlyph: '#cbd5e1',
  windowGlass: '#38bdf8',
  preview: '#34d399',
  label: '#e2e8f0',
  labelBg: '#1e1b4b',
}
