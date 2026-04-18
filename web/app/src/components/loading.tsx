import { Spinner } from '@hanzo/gui'

const sizeMap = {
  xs: 'small' as const,
  sm: 'small' as const,
  md: 'large' as const,
  lg: 'large' as const,
}

export function Loading({ size = 'md' }: { size?: 'xs' | 'sm' | 'md' | 'lg' }) {
  return <Spinner size={sizeMap[size]} color="$color" />
}
