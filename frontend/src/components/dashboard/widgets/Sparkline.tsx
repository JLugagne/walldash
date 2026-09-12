import React from 'react'

export interface SparklineProps {
  /** Recent samples, oldest first. Nothing is drawn until there are at least two. */
  history: number[]
  /** Stroke and wash colour; the area fill reuses it at low opacity. */
  color?: string
  className?: string
}

/**
 * Sparkline renders a compact, non-interactive trend line for a numeric entity. It is a pure
 * function of the samples it receives: it never invents points, and returns null when fewer than
 * two samples exist so a single reading cannot masquerade as a trend. The viewBox is stretched
 * with `preserveAspectRatio="none"`, and `vectorEffect="non-scaling-stroke"` keeps the 2 px stroke
 * from being distorted by that stretch.
 */
export const Sparkline: React.FC<SparklineProps> = ({ history, color = '#4bb8c9', className = '' }) => {
  if (history.length < 2) return null

  const width = 100
  const height = 26
  const min = Math.min(...history)
  const max = Math.max(...history)
  const range = max - min || 1
  const stepX = width / (history.length - 1)

  const points = history.map((value, index) => ({
    x: index * stepX,
    y: height - 2 - ((value - min) / range) * (height - 4),
  }))

  const line = points
    .map((point, index) => `${index === 0 ? 'M' : 'L'}${point.x.toFixed(2)} ${point.y.toFixed(2)}`)
    .join(' ')
  const area = `${line} L${width} ${height} L0 ${height} Z`

  return (
    <svg
      className={className}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      aria-hidden="true"
      focusable="false"
    >
      <path d={area} fill={color} opacity={0.12} />
      <path
        d={line}
        fill="none"
        stroke={color}
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  )
}
