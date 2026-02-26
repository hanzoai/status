import { cn } from '@/lib/utils'

type Status = 'healthy' | 'unhealthy' | 'degraded' | 'unknown'

const variants: Record<Status, { bg: string; dot: string; label: string }> = {
  healthy: {
    bg: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20',
    dot: 'bg-emerald-500',
    label: 'Healthy',
  },
  unhealthy: {
    bg: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
    dot: 'bg-red-500',
    label: 'Unhealthy',
  },
  degraded: {
    bg: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
    dot: 'bg-amber-500',
    label: 'Degraded',
  },
  unknown: {
    bg: 'bg-[hsl(var(--muted))] text-[hsl(var(--muted-foreground))] border-[hsl(var(--border))]',
    dot: 'bg-[hsl(var(--muted-foreground)/.5)]',
    label: 'Unknown',
  },
}

export function StatusBadge({ status }: { status: Status }) {
  const v = variants[status] ?? variants.unknown
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[11px] font-medium',
        v.bg,
      )}
    >
      <span className={cn('h-1.5 w-1.5 rounded-full', v.dot)} />
      {v.label}
    </span>
  )
}
