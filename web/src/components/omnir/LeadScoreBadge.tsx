import { cn } from '@/lib/utils'

interface LeadScoreBadgeProps {
  score: number
  className?: string
}

/**
 * Displays a lead score (0–100) with color coding:
 *   0–30  → cold (gray)
 *   31–60 → warm (yellow)
 *   61–100 → hot (red)
 */
export function LeadScoreBadge({ score, className }: LeadScoreBadgeProps) {
  const tier =
    score <= 30 ? 'cold'
    : score <= 60 ? 'warm'
    : 'hot'

  const label =
    tier === 'cold' ? 'Cold'
    : tier === 'warm' ? 'Warm'
    : 'Hot'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold tabular-nums',
        tier === 'cold' && 'bg-slate-100 text-slate-600',
        tier === 'warm' && 'bg-yellow-100 text-yellow-700',
        tier === 'hot'  && 'bg-red-100 text-red-700',
        className,
      )}
      title={`Lead score: ${score}`}
    >
      {score}
      <span className="opacity-70">{label}</span>
    </span>
  )
}
