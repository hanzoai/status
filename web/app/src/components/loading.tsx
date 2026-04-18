import { cn } from '@/lib/utils'

const sizes = {
  xs: 'h-4 w-4 border',
  sm: 'h-6 w-6 border-2',
  md: 'h-8 w-8 border-2',
  lg: 'h-12 w-12 border-2',
}

export function Loading({ size = 'md' }: { size?: keyof typeof sizes }) {
  return (
    <div
      className={cn(
        'animate-spin rounded-full border-foreground/20 border-t-foreground',
        sizes[size],
      )}
    />
  )
}
