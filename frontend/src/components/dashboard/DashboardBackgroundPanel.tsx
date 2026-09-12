import React from 'react'
import { X } from 'lucide-react'
import { BACKGROUND_SAMPLES } from './backgroundSamples'

export interface BackgroundConfig {
  image: string
  opacity: number
  blur: number
  dim: number
}

interface DashboardBackgroundPanelProps {
  value: BackgroundConfig
  onChange: (config: BackgroundConfig) => void
  onClose: () => void
}

/**
 * DashboardBackgroundPanel is the edit-mode background picker: bundled samples, a "no
 * background" reset, and the opacity / blur / dim sliders. It is controlled — every change is
 * forwarded through `onChange`, which previews live and persists (debounced) upstream.
 */
export const DashboardBackgroundPanel: React.FC<DashboardBackgroundPanelProps> = ({
  value,
  onChange,
  onClose,
}) => {
  const update = (patch: Partial<BackgroundConfig>) => onChange({ ...value, ...patch })

  return (
    <div className="fixed right-4 top-16 z-50 w-80 rounded-2xl border border-slate-800/80 bg-slate-900/70 backdrop-blur-md p-4 shadow-xl">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-[11px] font-semibold uppercase tracking-wider text-white">
          Dashboard background
        </h3>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          title="Close"
          className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-800 hover:text-white"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="mb-2 text-[11px] text-slate-400">Samples</div>
      <div className="mb-3 grid grid-cols-2 gap-2">
        {BACKGROUND_SAMPLES.map((sample) => {
          const selected = value.image === sample.url
          return (
            <button
              key={sample.id}
              type="button"
              onClick={() => update({ image: sample.url })}
              style={{ backgroundImage: `url(${sample.url})` }}
              className={`relative h-14 overflow-hidden rounded-lg border bg-cover bg-center text-left transition-colors ${
                selected
                  ? 'border-[#8b93ee] ring-2 ring-[#6d76e8]/50'
                  : 'border-slate-800/80 hover:border-slate-600'
              }`}
            >
              <span className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent px-1.5 py-0.5 text-[10px] font-semibold text-white">
                {sample.name}
              </span>
            </button>
          )
        })}
      </div>

      <button
        type="button"
        onClick={() => update({ image: '' })}
        className={`mb-3 h-8 w-full rounded-lg border text-[11px] font-semibold transition-colors ${
          value.image === ''
            ? 'border-[#8b93ee] bg-[#6d76e8]/20 text-white'
            : 'border-slate-700 bg-slate-800/60 text-slate-300 hover:text-white'
        }`}
      >
        No background
      </button>

      <BackgroundSlider
        label="Image opacity"
        value={value.opacity}
        min={0}
        max={100}
        suffix="%"
        disabled={value.image === ''}
        onChange={(v) => update({ opacity: v })}
      />
      <BackgroundSlider
        label="Blur"
        value={value.blur}
        min={0}
        max={32}
        suffix="px"
        disabled={value.image === ''}
        onChange={(v) => update({ blur: v })}
      />
      <BackgroundSlider
        label="Dim"
        value={value.dim}
        min={0}
        max={85}
        suffix="%"
        disabled={value.image === ''}
        onChange={(v) => update({ dim: v })}
      />

      <p className="mt-2 text-[10px] leading-relaxed text-slate-500">
        Widgets blur and dim the image behind them so their content stays readable.
      </p>
    </div>
  )
}

interface BackgroundSliderProps {
  label: string
  value: number
  min: number
  max: number
  suffix: string
  disabled?: boolean
  onChange: (value: number) => void
}

const BackgroundSlider: React.FC<BackgroundSliderProps> = ({
  label,
  value,
  min,
  max,
  suffix,
  disabled,
  onChange,
}) => (
  <label className="mb-3 block text-[11px] text-slate-400">
    <span className="mb-1 flex justify-between">
      <span>{label}</span>
      <span className="font-semibold text-slate-200">
        {value}
        {suffix}
      </span>
    </span>
    <input
      type="range"
      min={min}
      max={max}
      value={value}
      disabled={disabled}
      onChange={(e) => onChange(Number(e.target.value))}
      className="w-full accent-[#6d76e8] disabled:opacity-40"
    />
  </label>
)
