import { cn } from '@/lib/utils'

type Status = 'healthy' | 'unhealthy' | 'degraded' | 'unknown'

const variants: Record<Status, { bg: string; dot: string; label: string }> = {
  healthy: {
    bg: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
    dot: 'bg-green-500',
    label: 'Healthy',
  },
  unhealthy: {
    bg: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
    dot: 'bg-red-500',
    label: 'Unhealthy',
  },
  degraded: {
    bg: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20',
    dot: 'bg-yellow-500',
    label: 'Degraded',
  },
  unknown: {
    bg: 'bg-gray-500/10 text-gray-600 dark:text-gray-400 border-gray-500/20',
    dot: 'bg-gray-400',
    label: 'Unknown',
  },
}

export function StatusBadge({ status }: { status: Status }) {
  const v = variants[status] ?? variants.unknown
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium',
        v.bg,
      )}
    >
      <span className={cn('h-2 w-2 rounded-full', v.dot)} />
      {v.label}
    </span>
  )
}
