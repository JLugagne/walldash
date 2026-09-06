import type { ReactNode } from 'react'
import {
  AppWindow,
  BrickWall,
  Check,
  DoorOpen,
  Hexagon,
  Info,
  Minus,
  MousePointer2,
  Palette,
  Plus,
  Scissors,
  Thermometer,
  Trash2,
  X,
} from 'lucide-react'
import type { Device, DevicePlacement, Level, Plan, WallOpening, WallSegment, Zone } from '../../types'
import {
  DOOR_WIDTH_PRESETS,
  RENDER_DOMAIN_OPTIONS,
  WALL_LENGTH_PRESETS,
  WALL_THICKNESS_PRESETS,
  WINDOW_WIDTH_PRESETS,
  ZONE_COLOR_PRESETS,
  domainStyle,
  type ToolMode,
} from './constants'
import { cmToUnits, formatMeters, MIN_OPENING_WIDTH, unitsToCm, wallLength } from './geometry'
import type { EditorSelection } from './types'

export interface ToolSettings {
  wallThickness: number
  doorWidth: number
  windowWidth: number
  doorAsPassage: boolean
  zoneName: string
  zoneColor: string
}

interface InspectorProps {
  plan: Plan
  placements: DevicePlacement[]
  devices: Device[]
  level?: Level | null
  selection: EditorSelection | null
  tool: ToolMode
  settings: ToolSettings
  onSettings: (patch: Partial<ToolSettings>) => void
  zoneDraftCount: number
  onCompleteZone: () => void
  onCancelZone: () => void
  wallDrafting: boolean
  onCancelWall: () => void
  onUpdateWall: (id: string, patch: Partial<WallSegment>) => void
  onSetWallLength: (id: string, length: number) => void
  onSplitWall: (id: string) => void
  onUpdateOpening: (wallId: string, openingId: string, patch: Partial<WallOpening>) => void
  onUpdateZone: (id: string, patch: Partial<Zone>) => void
  onUpdatePlacement: (placement: DevicePlacement, patch: Partial<Pick<DevicePlacement, 'custom_name' | 'render_domain' | 'layer'>>) => void
  onDeleteSelection: () => void
  onSelect: (selection: EditorSelection | null) => void
}

export function Inspector(props: InspectorProps) {
  const { plan, selection, tool } = props

  if (selection?.type === 'wall') {
    const wall = plan.walls.find((w) => w.id === selection.id)
    if (wall) return <WallPanel wall={wall} {...props} />
  }
  if (selection?.type === 'opening') {
    const wall = plan.walls.find((w) => w.id === selection.wallId)
    const opening = wall?.openings?.find((o) => o.id === selection.id)
    if (wall && opening) return <OpeningPanel wall={wall} opening={opening} {...props} />
  }
  if (selection?.type === 'zone') {
    const zone = plan.zones.find((z) => z.id === selection.id)
    if (zone) return <ZonePanel zone={zone} {...props} />
  }
  if (selection?.type === 'device') {
    const placement = props.placements.find((p) => p.id === selection.id)
    if (placement) return <DevicePanel placement={placement} {...props} />
  }

  switch (tool) {
    case 'wall':
      return <WallToolPanel {...props} />
    case 'door':
    case 'window':
      return <OpeningToolPanel {...props} />
    case 'zone':
      return <ZoneToolPanel {...props} />
case 'pan':
        return (
          <PanelShell icon={<MousePointer2 className="w-4 h-4" />} title="Move view">
            <Hint
              items={[
                'Drag to move the plan.',
                'Scroll wheel to zoom around the cursor.',
                'Space, middle-click or right-click moves the view from any tool.',
              ]}
            />
          </PanelShell>
        )
    default:
      return <SelectToolPanel {...props} />
  }
}

function PanelShell({
  icon,
  title,
  subtitle,
  onClose,
  children,
}: {
  icon: ReactNode
  title: string
  subtitle?: string
  onClose?: () => void
  children: ReactNode
}) {
  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="px-4 py-3 border-b border-slate-800 flex items-center gap-2.5">
        <div className="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 flex items-center justify-center shrink-0">
          {icon}
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-sm font-semibold text-white truncate">{title}</div>
          {subtitle && <div className="text-[11px] text-slate-500 truncate font-mono">{subtitle}</div>}
        </div>
        {onClose && (
          <button
            type="button"
            onClick={onClose}
            title="Deselect (Escape)"
            className="w-7 h-7 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 flex items-center justify-center cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto p-4 space-y-5">{children}</div>
    </div>
  )
}

function Field({ label, trailing, children }: { label: string; trailing?: ReactNode; children: ReactNode }) {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">{label}</span>
        {trailing}
      </div>
      {children}
    </div>
  )
}

function Segmented<T extends string | number>({
  options,
  value,
  onChange,
}: {
  options: { label: string; value: T; hint?: string }[]
  value: T
  onChange: (value: T) => void
}) {
  return (
    <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}>
      {options.map((opt) => (
        <button
          key={String(opt.value)}
          type="button"
          onClick={() => onChange(opt.value)}
          title={opt.hint}
          className={`px-1 py-1.5 rounded-lg text-[11px] font-medium leading-tight transition-colors cursor-pointer truncate ${
            opt.value === value ? 'bg-indigo-600 text-white' : 'bg-slate-800/80 text-slate-400 hover:text-white hover:bg-slate-800'
          }`}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}

function Stepper({
  value,
  onChange,
  step,
  min,
  max,
  unit,
}: {
  value: number
  onChange: (value: number) => void
  step: number
  min: number
  max?: number
  unit: string
}) {
  const clamp = (v: number) => Math.max(min, max !== undefined ? Math.min(max, v) : v)
  return (
    <div className="flex items-center gap-1">
      <button
        type="button"
        onClick={() => onChange(clamp(value - step))}
        className="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 flex items-center justify-center cursor-pointer"
      >
        <Minus className="w-3.5 h-3.5" />
      </button>
      <div className="relative flex-1">
        <input
          type="number"
          value={value}
          min={min}
          max={max}
          onChange={(e) => {
            const v = Number(e.target.value)
            if (Number.isFinite(v)) onChange(clamp(v))
          }}
          className="w-full h-8 bg-slate-950 border border-slate-800 rounded-lg pl-2 pr-8 text-xs text-white font-mono text-right focus:outline-none focus:border-indigo-500"
        />
        <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-slate-500">{unit}</span>
      </div>
      <button
        type="button"
        onClick={() => onChange(clamp(value + step))}
        className="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 flex items-center justify-center cursor-pointer"
      >
        <Plus className="w-3.5 h-3.5" />
      </button>
    </div>
  )
}

function Toggle({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean
  onChange: (checked: boolean) => void
  label: string
  description?: string
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className="w-full flex items-center gap-3 p-2.5 rounded-lg bg-slate-800/50 hover:bg-slate-800 text-left transition-colors cursor-pointer"
    >
      <span
        className={`relative w-9 h-5 rounded-full transition-colors shrink-0 ${checked ? 'bg-indigo-600' : 'bg-slate-700'}`}
      >
        <span
          className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-all ${checked ? 'left-[18px]' : 'left-0.5'}`}
        />
      </span>
      <span className="min-w-0">
        <span className="block text-xs font-medium text-slate-200">{label}</span>
        {description && <span className="block text-[11px] text-slate-500 leading-snug">{description}</span>}
      </span>
    </button>
  )
}

function Hint({ items }: { items: string[] }) {
  return (
    <div className="rounded-lg bg-slate-800/40 border border-slate-800 p-3 space-y-1.5">
      {items.map((item, i) => (
        <div key={i} className="flex items-start gap-2 text-[11px] text-slate-400 leading-snug">
          <Info className="w-3 h-3 mt-0.5 text-slate-500 shrink-0" />
          <span>{item}</span>
        </div>
      ))}
    </div>
  )
}

function DangerButton({ onClick, children }: { onClick: () => void; children: ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full h-9 rounded-lg bg-rose-600/15 hover:bg-rose-600 border border-rose-600/30 text-rose-300 hover:text-white text-xs font-semibold flex items-center justify-center gap-2 transition-colors cursor-pointer"
    >
      <Trash2 className="w-3.5 h-3.5" />
      {children}
    </button>
  )
}

function shortId(id: string): string {
  return id.length > 18 ? `…${id.slice(-12)}` : id
}

function WallPanel({ wall, onUpdateWall, onSetWallLength, onSplitWall, onDeleteSelection, onSelect }: InspectorProps & { wall: WallSegment }) {
  const len = Math.round(wallLength(wall))
  const thickness = wall.thickness || 12
  const thicknessKnown = WALL_THICKNESS_PRESETS.some((p) => p.value === thickness)
  const openings = wall.openings ?? []

  return (
    <PanelShell icon={<BrickWall className="w-4 h-4" />} title="Wall" subtitle={shortId(wall.id)} onClose={() => onSelect(null)}>
      <Field label="Length" trailing={<span className="text-[11px] font-mono text-indigo-300">{formatMeters(len)}</span>}>
        <Stepper value={unitsToCm(len)} onChange={(cm) => onSetWallLength(wall.id, cmToUnits(cm))} step={10} min={20} unit="cm" />
        <div className="grid grid-cols-5 gap-1">
          {WALL_LENGTH_PRESETS.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => onSetWallLength(wall.id, p.value)}
              className={`py-1 rounded-md text-[10px] font-mono transition-colors cursor-pointer ${
                len === p.value ? 'bg-indigo-600 text-white' : 'bg-slate-800/80 text-slate-400 hover:text-white'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </Field>

      <Field label="Thickness" trailing={<span className="text-[11px] font-mono text-slate-400">{unitsToCm(thickness)} cm</span>}>
        <Segmented
          options={WALL_THICKNESS_PRESETS.map((p) => ({ label: p.label, value: p.value, hint: `${unitsToCm(p.value)} cm` }))}
          value={thicknessKnown ? thickness : -1}
          onChange={(v) => onUpdateWall(wall.id, { thickness: v })}
        />
        {!thicknessKnown && (
          <Stepper value={unitsToCm(thickness)} onChange={(cm) => onUpdateWall(wall.id, { thickness: Math.max(2, cmToUnits(cm)) })} step={5} min={5} unit="cm" />
        )}
      </Field>

      <Field label={`Openings (${openings.length})`}>
        {openings.length === 0 ? (
          <p className="text-[11px] text-slate-500">No doors or windows. Use the Door (D) or Window (F) tools.</p>
        ) : (
          <div className="space-y-1">
            {openings.map((op) => {
              const Icon = op.type === 'window' ? AppWindow : DoorOpen
              return (
                <button
                  key={op.id}
                  type="button"
                  onClick={() => onSelect({ type: 'opening', id: op.id, wallId: wall.id })}
                  className="w-full flex items-center gap-2 px-2.5 py-2 rounded-lg bg-slate-800/50 hover:bg-slate-800 text-left cursor-pointer"
                >
                  <Icon className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                  <span className="text-xs text-slate-200 flex-1">
                    {op.type === 'window' ? 'Window' : op.hide_door ? 'Passage' : 'Door'}
                  </span>
                  <span className="text-[10px] font-mono text-slate-500">
                    {unitsToCm(op.width)} cm · {formatMeters(op.offset)}
                  </span>
                </button>
              )
            })}
          </div>
        )}
      </Field>

      <Field label="Actions">
        <button
          type="button"
          onClick={() => onSplitWall(wall.id)}
          className="w-full h-9 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium flex items-center justify-center gap-2 cursor-pointer"
        >
          <Scissors className="w-3.5 h-3.5" />
          Split at midpoint
        </button>
        <DangerButton onClick={onDeleteSelection}>Delete wall</DangerButton>
      </Field>

      <Hint
        items={[
          'Drag the wall to move it: connected walls follow.',
          'Hold Alt while dragging to detach the wall from its joints.',
          'Drag an endpoint handle to stretch or reorient the wall.',
        ]}
      />
    </PanelShell>
  )
}

function OpeningPanel({ wall, opening, onUpdateOpening, onDeleteSelection, onSelect }: InspectorProps & { wall: WallSegment; opening: WallOpening }) {
  const isWindow = opening.type === 'window'
  const len = wallLength(wall)
  const half = opening.width / 2
  const minOffset = Math.round(half)
  const maxOffset = Math.max(minOffset, Math.round(len - half))
  const presets = isWindow ? WINDOW_WIDTH_PRESETS : DOOR_WIDTH_PRESETS
  const update = (patch: Partial<WallOpening>) => onUpdateOpening(wall.id, opening.id, patch)
  const title = isWindow ? 'Window' : opening.hide_door ? 'Passage' : 'Door'

  return (
    <PanelShell
      icon={isWindow ? <AppWindow className="w-4 h-4" /> : <DoorOpen className="w-4 h-4" />}
      title={title}
      subtitle={shortId(opening.id)}
      onClose={() => onSelect(null)}
    >
      <Field label="Type">
        <Segmented
          options={[
            { label: 'Door', value: 'door' },
            { label: 'Window', value: 'window' },
          ]}
          value={opening.type}
          onChange={(type) => update({ type: type as WallOpening['type'] })}
        />
      </Field>

      <Field label="Width">
        <Stepper
          value={unitsToCm(opening.width)}
          onChange={(cm) => update({ width: Math.max(MIN_OPENING_WIDTH, cmToUnits(cm)) })}
          step={5}
          min={unitsToCm(MIN_OPENING_WIDTH)}
          max={unitsToCm(len)}
          unit="cm"
        />
        <div className="grid grid-cols-5 gap-1">
          {presets.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => update({ width: p.value })}
              className={`py-1 rounded-md text-[10px] font-mono transition-colors cursor-pointer ${
                opening.width === p.value ? 'bg-indigo-600 text-white' : 'bg-slate-800/80 text-slate-400 hover:text-white'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </Field>

      <Field label="Position on wall" trailing={<span className="text-[11px] font-mono text-slate-400">{formatMeters(opening.offset)}</span>}>
        <input
          type="range"
          min={minOffset}
          max={maxOffset}
          value={opening.offset}
          onChange={(e) => update({ offset: Number(e.target.value) })}
          className="w-full accent-indigo-500 cursor-pointer"
        />
        <div className="grid grid-cols-3 gap-1">
          <button type="button" onClick={() => update({ offset: minOffset })} className="py-1 rounded-md text-[10px] bg-slate-800/80 text-slate-400 hover:text-white cursor-pointer">
            Start
          </button>
          <button type="button" onClick={() => update({ offset: Math.round(len / 2) })} className="py-1 rounded-md text-[10px] bg-slate-800/80 text-slate-400 hover:text-white cursor-pointer">
            Center
          </button>
          <button type="button" onClick={() => update({ offset: maxOffset })} className="py-1 rounded-md text-[10px] bg-slate-800/80 text-slate-400 hover:text-white cursor-pointer">
            End
          </button>
        </div>
      </Field>

      {!isWindow && (
        <Field label="Render">
          <div className="space-y-1.5">
            <Toggle
              checked={!!opening.hide_door}
              onChange={(v) => update({ hide_door: v })}
              label="Passage without door"
              description="Keeps the wall opening but hides the door and frame, in both 2D and 3D."
            />
            {!opening.hide_door && (
              <>
                <Toggle checked={!!opening.flip_side} onChange={(v) => update({ flip_side: v })} label="Opens on the other side of the wall" />
                <Toggle checked={!!opening.flip_hinge} onChange={(v) => update({ flip_hinge: v })} label="Hinge at the other end" />
              </>
            )}
          </div>
        </Field>
      )}

      <Field label="Actions">
        <button
          type="button"
          onClick={() => onSelect({ type: 'wall', id: wall.id })}
          className="w-full h-9 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium flex items-center justify-center gap-2 cursor-pointer"
        >
          <BrickWall className="w-3.5 h-3.5" />
          View the parent wall
        </button>
        <DangerButton onClick={onDeleteSelection}>Delete {isWindow ? 'window' : 'opening'}</DangerButton>
      </Field>

      <Hint items={['Drag the opening along the wall to reposition it.', 'Drag its handles to adjust the width.']} />
    </PanelShell>
  )
}

function isTemperatureSensor(d: Device): boolean {
  if (d.domain !== 'sensor') return false
  const dc = (d.attributes?.device_class || '').toLowerCase()
  const unit = (d.attributes?.unit_of_measurement || '').toLowerCase()
  const id = d.id.toLowerCase()
  const name = (d.name || '').toLowerCase()
  return (
    dc === 'temperature' ||
    unit.includes('°c') ||
    unit.includes('°f') ||
    id.includes('temp') ||
    name.includes('temp')
  )
}

function isHumiditySensor(d: Device): boolean {
  if (d.domain !== 'sensor') return false
  const dc = (d.attributes?.device_class || '').toLowerCase()
  const id = d.id.toLowerCase()
  const name = (d.name || '').toLowerCase()
  return dc === 'humidity' || id.includes('humid') || name.includes('humid')
}

function ZonePanel({ zone, devices, onUpdateZone, onDeleteSelection, onSelect }: InspectorProps & { zone: Zone }) {
  const tempCandidates = devices.filter(isTemperatureSensor)
  const allSensors = devices.filter((d) => d.domain === 'sensor')
  const baseTempList = tempCandidates.length > 0 ? tempCandidates : allSensors
  const tempOptions = [...baseTempList]
  if (zone.temp_sensor && !tempOptions.some((d) => d.id === zone.temp_sensor)) {
    const found = devices.find((d) => d.id === zone.temp_sensor)
    tempOptions.unshift(found || { id: zone.temp_sensor, name: zone.temp_sensor, domain: 'sensor', state: '', attributes: {}, last_updated: '' })
  }

  const humCandidates = devices.filter(isHumiditySensor)
  const baseHumList = humCandidates.length > 0 ? humCandidates : allSensors
  const humOptions = [...baseHumList]
  if (zone.humidity_sensor && !humOptions.some((d) => d.id === zone.humidity_sensor)) {
    const found = devices.find((d) => d.id === zone.humidity_sensor)
    humOptions.unshift(found || { id: zone.humidity_sensor, name: zone.humidity_sensor, domain: 'sensor', state: '', attributes: {}, last_updated: '' })
  }

  const hasThresholds = zone.temp_min !== undefined || zone.temp_max !== undefined
  const hasHumThresholds = zone.humidity_min !== undefined || zone.humidity_max !== undefined

  return (
    <PanelShell icon={<Hexagon className="w-4 h-4" />} title={zone.name || 'Zone'} subtitle={`${zone.points.length} vertices`} onClose={() => onSelect(null)}>
      <Field label="Name">
        <input
          type="text"
          value={zone.name}
          onChange={(e) => onUpdateZone(zone.id, { name: e.target.value })}
          className="w-full h-9 bg-slate-950 border border-slate-800 rounded-lg px-3 text-xs text-white focus:outline-none focus:border-indigo-500"
        />
      </Field>
      <Field label="Color">
        <ColorPicker value={zone.color} onChange={(color) => onUpdateZone(zone.id, { color })} />
      </Field>

      <div className="pt-2 border-t border-slate-800/80 space-y-3">
        <div className="flex items-center gap-1.5 text-xs font-semibold text-slate-300">
          <Thermometer className="w-3.5 h-3.5 text-indigo-400" />
          <span>Ambient sensors</span>
        </div>

        <Field label="Temperature sensor">
          <div className="flex items-center gap-1.5">
            <select
              value={zone.temp_sensor || ''}
              onChange={(e) => onUpdateZone(zone.id, { temp_sensor: e.target.value || undefined })}
              className="flex-1 min-w-0 h-9 bg-slate-950 border border-slate-800 rounded-lg px-2.5 text-xs text-white focus:outline-none focus:border-indigo-500 truncate"
            >
              <option value="">No sensor</option>
              {tempOptions.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name ? `${d.name} (${d.id})` : d.id}
                </option>
              ))}
            </select>
            {zone.temp_sensor && (
              <button
                type="button"
                onClick={() => onUpdateZone(zone.id, { temp_sensor: undefined })}
                title="Remove temperature sensor"
                className="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white flex items-center justify-center cursor-pointer shrink-0"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
        </Field>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">Temperature alert thresholds</span>
            {hasThresholds && (
              <button
                type="button"
                onClick={() => onUpdateZone(zone.id, { temp_min: undefined, temp_max: undefined })}
                className="text-[10px] text-slate-500 hover:text-slate-300 underline cursor-pointer"
              >
                Clear thresholds
              </button>
            )}
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label className="block text-[10px] text-slate-400">Cold alert &lt;</label>
              <div className="relative">
                <input
                  type="number"
                  step="0.5"
                  placeholder="Ex: 18"
                  value={zone.temp_min ?? ''}
                  onChange={(e) => {
                    const val = e.target.value === '' ? undefined : Number(e.target.value)
                    onUpdateZone(zone.id, { temp_min: val !== undefined && Number.isFinite(val) ? val : undefined })
                  }}
                  className="w-full h-8 bg-slate-950 border border-slate-800 rounded-lg pl-2 pr-7 text-xs text-white font-mono text-right focus:outline-none focus:border-indigo-500"
                />
                <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-slate-500">°C</span>
              </div>
            </div>
            <div className="space-y-1">
              <label className="block text-[10px] text-slate-400">Hot alert &gt;</label>
              <div className="relative">
                <input
                  type="number"
                  step="0.5"
                  placeholder="Ex: 26"
                  value={zone.temp_max ?? ''}
                  onChange={(e) => {
                    const val = e.target.value === '' ? undefined : Number(e.target.value)
                    onUpdateZone(zone.id, { temp_max: val !== undefined && Number.isFinite(val) ? val : undefined })
                  }}
                  className="w-full h-8 bg-slate-950 border border-slate-800 rounded-lg pl-2 pr-7 text-xs text-white font-mono text-right focus:outline-none focus:border-indigo-500"
                />
                <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-slate-500">°C</span>
              </div>
            </div>
          </div>
          {zone.temp_min !== undefined && zone.temp_max !== undefined && zone.temp_min > zone.temp_max && (
            <p className="text-[11px] text-rose-400 font-medium pt-0.5">
              Warning: the cold threshold ({zone.temp_min}°C) is higher than the hot threshold ({zone.temp_max}°C).
            </p>
          )}
        </div>

        <Field label="Humidity sensor">
          <div className="flex items-center gap-1.5">
            <select
              value={zone.humidity_sensor || ''}
              onChange={(e) => onUpdateZone(zone.id, { humidity_sensor: e.target.value || undefined })}
              className="flex-1 min-w-0 h-9 bg-slate-950 border border-slate-800 rounded-lg px-2.5 text-xs text-white focus:outline-none focus:border-indigo-500 truncate"
            >
              <option value="">No sensor</option>
              {humOptions.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name ? `${d.name} (${d.id})` : d.id}
                </option>
              ))}
            </select>
            {zone.humidity_sensor && (
              <button
                type="button"
                onClick={() => onUpdateZone(zone.id, { humidity_sensor: undefined })}
                title="Remove humidity sensor"
                className="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white flex items-center justify-center cursor-pointer shrink-0"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
        </Field>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">Humidity alert thresholds</span>
            {hasHumThresholds && (
              <button
                type="button"
                onClick={() => onUpdateZone(zone.id, { humidity_min: undefined, humidity_max: undefined })}
                className="text-[10px] text-slate-500 hover:text-slate-300 underline cursor-pointer"
              >
                Clear thresholds
              </button>
            )}
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label className="block text-[10px] text-slate-400">Dry alert &lt;</label>
              <div className="relative">
                <input
                  type="number"
                  step="1"
                  placeholder="Ex: 30"
                  value={zone.humidity_min ?? ''}
                  onChange={(e) => {
                    const val = e.target.value === '' ? undefined : Number(e.target.value)
                    onUpdateZone(zone.id, { humidity_min: val !== undefined && Number.isFinite(val) ? val : undefined })
                  }}
                  className="w-full h-8 bg-slate-950 border border-slate-800 rounded-lg pl-2 pr-7 text-xs text-white font-mono text-right focus:outline-none focus:border-indigo-500"
                />
                <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-slate-500">%</span>
              </div>
            </div>
            <div className="space-y-1">
              <label className="block text-[10px] text-slate-400">Humid alert &gt;</label>
              <div className="relative">
                <input
                  type="number"
                  step="1"
                  placeholder="Ex: 70"
                  value={zone.humidity_max ?? ''}
                  onChange={(e) => {
                    const val = e.target.value === '' ? undefined : Number(e.target.value)
                    onUpdateZone(zone.id, { humidity_max: val !== undefined && Number.isFinite(val) ? val : undefined })
                  }}
                  className="w-full h-8 bg-slate-950 border border-slate-800 rounded-lg pl-2 pr-7 text-xs text-white font-mono text-right focus:outline-none focus:border-indigo-500"
                />
                <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-slate-500">%</span>
              </div>
            </div>
          </div>
          {zone.humidity_min !== undefined && zone.humidity_max !== undefined && zone.humidity_min > zone.humidity_max && (
            <p className="text-[11px] text-rose-400 font-medium pt-0.5">
              Warning: the dry threshold ({zone.humidity_min}%) is higher than the humid threshold ({zone.humidity_max}%).
            </p>
          )}
        </div>
      </div>

      <Field label="Actions">
        <DangerButton onClick={onDeleteSelection}>Delete zone</DangerButton>
      </Field>
      <Hint items={['Drag the zone to move it.', 'Drag a vertex to adjust the outline.']} />
    </PanelShell>
  )
}

function ColorPicker({ value, onChange, onPickName }: { value: string; onChange: (color: string) => void; onPickName?: (name: string) => void }) {
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      {ZONE_COLOR_PRESETS.map((preset) => (
        <button
          key={preset.value}
          type="button"
          title={preset.name}
          onClick={() => {
            onChange(preset.value)
            onPickName?.(preset.name)
          }}
          className={`w-6 h-6 rounded-full border-2 transition-transform cursor-pointer ${
            value.toLowerCase() === preset.value ? 'border-white scale-110' : 'border-transparent hover:scale-110'
          }`}
          style={{ backgroundColor: preset.value }}
        />
      ))}
      <label className="relative w-6 h-6 rounded-full border-2 border-dashed border-slate-600 hover:border-slate-400 flex items-center justify-center cursor-pointer overflow-hidden" title="Custom color">
        <Palette className="w-3 h-3 text-slate-400" />
        <input type="color" value={value} onChange={(e) => onChange(e.target.value)} className="absolute inset-0 opacity-0 cursor-pointer" />
      </label>
    </div>
  )
}

function DevicePanel({ placement, devices, level, onUpdatePlacement, onDeleteSelection, onSelect }: InspectorProps & { placement: DevicePlacement }) {
  const device = devices.find((d) => d.id === placement.device_id)
  const domain = placement.render_domain || device?.domain || placement.icon || 'light'
  const style = domainStyle(domain)
  const Icon = style.Icon
  const currentLayer = placement.layer || 'controls'
  const levelLayers = level?.layers && level.layers.length > 0 ? level.layers : ['controls', 'sensors']
  const allLayers = Array.from(new Set([...levelLayers, currentLayer]))

  return (
    <PanelShell icon={<Icon className="w-4 h-4" />} title={placement.custom_name || device?.name || placement.device_id} subtitle={placement.device_id} onClose={() => onSelect(null)}>
      <Field label="Display name">
        <input
          type="text"
          defaultValue={placement.custom_name || device?.name || ''}
          key={placement.id}
          onBlur={(e) => {
            const v = e.target.value.trim()
            if (v !== (placement.custom_name || '')) onUpdatePlacement(placement, { custom_name: v || undefined })
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') (e.target as HTMLInputElement).blur()
          }}
          className="w-full h-9 bg-slate-950 border border-slate-800 rounded-lg px-3 text-xs text-white focus:outline-none focus:border-indigo-500"
        />
      </Field>

      <Field label="Display Layer" trailing={<span className="text-[11px] font-mono text-indigo-300">{currentLayer}</span>}>
        <select
          value={currentLayer}
          onChange={(e) => onUpdatePlacement(placement, { layer: e.target.value })}
          className="w-full h-9 bg-slate-950 border border-slate-800 rounded-lg px-2 text-xs text-white focus:outline-none focus:border-indigo-500"
        >
          {allLayers.map((l) => (
            <option key={l} value={l}>
              {l} {l === 'controls' ? '(default)' : ''}
            </option>
          ))}
        </select>
        <p className="text-[11px] text-slate-500 leading-snug mt-1.5">
          Manage layers in the <span className="text-indigo-400">Layers</span> panel.
        </p>
      </Field>

      <Field label="Displayed as" trailing={<span className="text-[11px] text-slate-500">actual: {device?.domain ?? '—'}</span>}>
        <select
          value={placement.render_domain || ''}
          onChange={(e) => onUpdatePlacement(placement, { render_domain: e.target.value || undefined })}
          className="w-full h-9 bg-slate-950 border border-slate-800 rounded-lg px-2 text-xs text-white focus:outline-none focus:border-indigo-500"
        >
          <option value="">Automatic</option>{RENDER_DOMAIN_OPTIONS.map((opt) => (
            <option key={opt.key} value={opt.key}>
              {opt.label}
            </option>
          ))}
        </select>
        <p className="text-[11px] text-slate-500 leading-snug">
          Useful for an outlet controlling a lamp: it will display as a light in the views.
        </p>
      </Field>

      <Field label="State" trailing={<span className="text-[11px] font-mono text-slate-400">{device?.state ?? 'unknown'}</span>}>
        <div className="text-[11px] font-mono text-slate-500">
          x {Math.round(placement.x)} · y {Math.round(placement.y)}
        </div>
      </Field>

      <Field label="Actions">
        <DangerButton onClick={onDeleteSelection}>Remove from plan</DangerButton>
      </Field>
    </PanelShell>
  )
}

function WallToolPanel({ settings, onSettings, wallDrafting, onCancelWall }: InspectorProps) {
  return (
    <PanelShell icon={<BrickWall className="w-4 h-4" />} title="Draw walls" subtitle="Shortcut W">
      <Field label="New wall thickness" trailing={<span className="text-[11px] font-mono text-slate-400">{unitsToCm(settings.wallThickness)} cm</span>}>
        <Segmented
          options={WALL_THICKNESS_PRESETS.map((p) => ({ label: p.label, value: p.value, hint: `${unitsToCm(p.value)} cm` }))}
          value={settings.wallThickness}
          onChange={(v) => onSettings({ wallThickness: v })}
        />
      </Field>

      {wallDrafting && (
        <button
          type="button"
          onClick={onCancelWall}
          className="w-full h-9 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium flex items-center justify-center gap-2 cursor-pointer"
        >
          <X className="w-3.5 h-3.5" />
          Finish drawing (Escape)
        </button>
      )}

      <Hint
        items={[
          'Click to place the start point, then each corner: walls chain together.',
          'Cursor snaps to existing endpoints for clean room closure.',
          'Shift: lock to horizontal / vertical.',
          'Escape, Enter or double-click: finish the chain.',
        ]}
      />
    </PanelShell>
  )
}

function OpeningToolPanel({ tool, settings, onSettings }: InspectorProps) {
  const isDoor = tool === 'door'
  const width = isDoor ? settings.doorWidth : settings.windowWidth
  const presets = isDoor ? DOOR_WIDTH_PRESETS : WINDOW_WIDTH_PRESETS
  const setWidth = (v: number) => onSettings(isDoor ? { doorWidth: v } : { windowWidth: v })

  return (
    <PanelShell
      icon={isDoor ? <DoorOpen className="w-4 h-4" /> : <AppWindow className="w-4 h-4" />}
      title={isDoor ? 'Place a door' : 'Place a window'}
      subtitle={isDoor ? 'Shortcut D' : 'Shortcut F'}
    >
      <Field label="Default width">
        <Stepper value={unitsToCm(width)} onChange={(cm) => setWidth(Math.max(MIN_OPENING_WIDTH, cmToUnits(cm)))} step={5} min={unitsToCm(MIN_OPENING_WIDTH)} unit="cm" />
        <div className="grid grid-cols-5 gap-1">
          {presets.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => setWidth(p.value)}
              className={`py-1 rounded-md text-[10px] font-mono transition-colors cursor-pointer ${
                width === p.value ? 'bg-indigo-600 text-white' : 'bg-slate-800/80 text-slate-400 hover:text-white'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </Field>

      {isDoor && (
        <Field label="Render">
          <Toggle
            checked={settings.doorAsPassage}
            onChange={(v) => onSettings({ doorAsPassage: v })}
            label="Create passages without doors"
            description="New openings cut through the wall without showing a door."
          />
        </Field>
      )}

      <Hint items={['Hover over a wall: a preview snaps onto it.', 'Click to place the opening, then adjust it from the Selection tool.']} />
    </PanelShell>
  )
}

function ZoneToolPanel({ settings, onSettings, zoneDraftCount, onCompleteZone, onCancelZone }: InspectorProps) {
  return (
    <PanelShell icon={<Hexagon className="w-4 h-4" />} title="Draw a zone" subtitle="Shortcut Z">
      <Field label="Name">
        <input
          type="text"
          value={settings.zoneName}
          onChange={(e) => onSettings({ zoneName: e.target.value })}
          placeholder="Living room, Kitchen…"
          className="w-full h-9 bg-slate-950 border border-slate-800 rounded-lg px-3 text-xs text-white focus:outline-none focus:border-indigo-500 placeholder:text-slate-600"
        />
      </Field>
      <Field label="Color">
        <ColorPicker value={settings.zoneColor} onChange={(zoneColor) => onSettings({ zoneColor })} onPickName={(zoneName) => onSettings({ zoneName })} />
      </Field>

      <Field label="Ongoing outline" trailing={<span className="text-[11px] font-mono text-slate-400">{zoneDraftCount} points</span>}>
        <div className="grid grid-cols-2 gap-1.5">
          <button
            type="button"
            disabled={zoneDraftCount < 3}
            onClick={onCompleteZone}
            className="h-9 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white text-xs font-semibold flex items-center justify-center gap-1.5 cursor-pointer"
          >
            <Check className="w-3.5 h-3.5" />
            Close
          </button>
          <button
            type="button"
            disabled={zoneDraftCount === 0}
            onClick={onCancelZone}
            className="h-9 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-200 text-xs font-medium flex items-center justify-center gap-1.5 cursor-pointer"
          >
            <X className="w-3.5 h-3.5" />
            Cancel
          </button>
        </div>
      </Field>

      <Hint items={['Click to place each vertex.', 'Click the first point, press Enter or double-click to close the zone.']} />
    </PanelShell>
  )
}

function SelectToolPanel({ plan, placements }: InspectorProps) {
  const openings = plan.walls.reduce((acc, w) => acc + (w.openings?.length ?? 0), 0)
  const stats = [
    { label: 'Walls', value: plan.walls.length },
    { label: 'Openings', value: openings },
    { label: 'Zones', value: plan.zones.length },
    { label: 'Devices', value: placements.length },
  ]
  return (
    <PanelShell icon={<MousePointer2 className="w-4 h-4" />} title="Selection" subtitle="Shortcut V">
      <div className="grid grid-cols-2 gap-2">
        {stats.map((s) => (
          <div key={s.label} className="rounded-lg bg-slate-800/50 border border-slate-800 p-3">
            <div className="text-lg font-bold text-white leading-none">{s.value}</div>
            <div className="text-[11px] text-slate-500 mt-1">{s.label}</div>
          </div>
        ))}
      </div>
      <Hint
        items={[
          'Click a wall, opening, zone or device to edit it here.',
          'Drag to move. Walls pull their joints (Alt to detach).',
          'Delete clears the selection, Ctrl+Z undoes, Ctrl+S saves.',
        ]}
      />
      <div className="rounded-lg border border-slate-800 p-3 space-y-1.5">
        <div className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">Shortcuts</div>
        {[
          ['V', 'Selection'],
          ['W', 'Wall'],
          ['D', 'Door'],
          ['F', 'Window'],
          ['Z', 'Zone'],
          ['H', 'Move view'],
          ['G', 'Snap'],
          ['Space', 'Move (hold)'],
        ].map(([key, label]) => (
          <div key={key} className="flex items-center justify-between text-[11px]">
            <span className="text-slate-400">{label}</span>
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 border border-slate-700 text-slate-300 font-mono text-[10px]">{key}</kbd>
          </div>
        ))}
      </div>
    </PanelShell>
  )
}
