import React from 'react'
import { AlertTriangle } from 'lucide-react'

interface StaleBadgeProps {
  className?: string
}

// Single warning badge shown when a widget's value is unavailable, unknown,
// or older than the 30-minute freshness threshold. The caller decides
// whether the value is stale; this component only renders the indicator.
// Sized to sit inside WidgetFrame's 18 px caption row.
export const StaleBadge: React.FC<StaleBadgeProps> = ({ className = '' }) => (
  <span
    role="img"
    aria-label="Stale value"
    title="Last known value, not recently refreshed"
    className={`inline-flex items-center justify-center w-4 h-4 shrink-0 rounded-full bg-[#d99a3a]/15 border border-[#d99a3a]/40 text-[#d99a3a] ${className}`}
  >
    <AlertTriangle className="w-2.5 h-2.5" />
  </span>
)
