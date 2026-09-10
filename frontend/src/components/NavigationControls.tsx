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
      <div className="hidden sm:flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-900/70 backdrop-blur-md border border-slate-800/80 text-[11px] text-slate-400">
        <Move className="w-3 h-3 text-slate-400" />
        <span>Drag 1 finger / mouse to move</span>
      </div>

      <div className="bg-slate-900/70 backdrop-blur-md p-1.5 rounded-xl border border-slate-800/80 flex flex-col space-y-1">
        {/* Zoom In */}
        <button
          type="button"
          onClick={onZoomIn}
          title="Zoom in (+)"
          aria-label="Zoom in"
          className="w-11 h-11 rounded-lg bg-slate-800/60 hover:bg-slate-800/80 active:scale-95 text-slate-300 hover:text-white flex items-center justify-center transition-all border border-slate-800/80"
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
          className="w-11 h-11 rounded-lg bg-slate-800/60 hover:bg-slate-800/80 active:scale-95 text-slate-300 hover:text-white flex items-center justify-center transition-all border border-slate-800/80"
        >
          <ZoomOut className="w-5 h-5" />
        </button>

        <div className="w-full h-px bg-slate-800/80 my-0.5" />

        {/* Recenter Button */}
        <button
          type="button"
          onClick={onRecenter}
          title="Reset view"
          aria-label="Reset view"
          className="w-11 h-11 rounded-lg bg-[#6d76e8] hover:bg-[#7b83ea] active:scale-95 text-white flex items-center justify-center transition-all border border-[#6d76e8]/40"
        >
          <RotateCcw className="w-5 h-5" />
        </button>
      </div>
    </div>
  )
}
