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
    aria-label="Valeur non fraîche"
    title="Dernière valeur connue, non rafraîchie récemment"
    className={`inline-flex items-center justify-center w-4 h-4 shrink-0 rounded-full bg-amber-500/15 border border-amber-500/40 text-amber-400 ${className}`}
  >
    <AlertTriangle className="w-2.5 h-2.5" />
  </span>
)
