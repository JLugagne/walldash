import { RotateCcw, ZoomIn, ZoomOut, Move } from 'lucide-react'

interface NavigationControlsProps {
  zoom: number
  defaultZoom: number
  onZoomIn: () => void
  onZoomOut: () => void
  onRecenter: () => void
}

export function NavigationControls({
  zoom,
  defaultZoom,
  onZoomIn,
  onZoomOut,
  onRecenter,
}: NavigationControlsProps) {
  const zoomPercent = Math.round((zoom / defaultZoom) * 100)

  return (
    <div className="flex flex-col items-end space-y-2 pointer-events-auto select-none">
      {/* Quick pan tooltip indicator */}
      <div className="hidden sm:flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-900/80 backdrop-blur border border-slate-800 text-[11px] text-slate-400">
        <Move className="w-3 h-3 text-slate-400" />
        <span>Drag 1 finger / mouse to move</span>
      </div>

      <div className="bg-slate-900/90 backdrop-blur-md p-1.5 rounded-2xl border border-slate-800/80 shadow-2xl shadow-black/50 flex flex-col space-y-1">
        {/* Zoom In */}
        <button
          type="button"
          onClick={onZoomIn}
          title="Zoom in (+)"
          aria-label="Zoom in"
          className="w-11 h-11 rounded-xl bg-slate-800/80 hover:bg-slate-700/80 active:scale-95 text-slate-200 hover:text-white flex items-center justify-center transition-all border border-slate-700/50"
        >
          <ZoomIn className="w-5 h-5" />
        </button>

        {/* Zoom Level Readout */}
        <div className="text-[10px] font-mono text-center text-slate-400 py-0.5 font-semibold select-none">
          {zoomPercent}%
        </div>

        {/* Zoom Out */}
        <button
          type="button"
          onClick={onZoomOut}
          title="Zoom out (-)"
          aria-label="Zoom out"
          className="w-11 h-11 rounded-xl bg-slate-800/80 hover:bg-slate-700/80 active:scale-95 text-slate-200 hover:text-white flex items-center justify-center transition-all border border-slate-700/50"
        >
          <ZoomOut className="w-5 h-5" />
        </button>

        <div className="w-full h-px bg-slate-800 my-0.5" />

        {/* Recenter Button */}
        <button
          type="button"
          onClick={onRecenter}
          title="Reset view"
          aria-label="Reset view"
          className="w-11 h-11 rounded-xl bg-indigo-600/90 hover:bg-indigo-500 active:scale-95 text-white flex items-center justify-center transition-all shadow-lg shadow-indigo-600/30 border border-indigo-400/30"
        >
          <RotateCcw className="w-5 h-5" />
        </button>
      </div>
    </div>
  )
}
