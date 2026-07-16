import type { ButtonHTMLAttributes, HTMLAttributes, InputHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

/**
 * Plain-React + Tailwind replacements for the handful of primitives the status
 * UI used from @hanzo/gui. @hanzo/gui@4.8.0 shipped an internally inconsistent
 * dependency tree (provider on @hanzogui/core@4.4.0, components on 3.0.x) whose
 * split theme React-context crashed the app at mount ("Missing theme."), and no
 * adjacent version is both installable and consistent. These cover exactly what
 * we render; all styling comes from the Tailwind theme tokens in index.css.
 */

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  bordered?: boolean
  padded?: boolean
}

export function Card({ bordered, padded = true, className, children, ...rest }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-lg bg-card text-card-foreground',
        bordered && 'border border-border',
        padded && 'p-4',
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  )
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** React-Native-Web-style press handler (mapped to onClick). */
  onPress?: () => void
  chromeless?: boolean
  circular?: boolean
  /** @hanzo/gui size token (e.g. "$3") — ignored, kept for prop compatibility. */
  size?: string
}

export function Button({
  onPress,
  chromeless,
  circular,
  size: _size,
  className,
  children,
  ...rest
}: ButtonProps) {
  return (
    <button
      type="button"
      {...rest}
      onClick={onPress}
      className={cn(
        'inline-flex items-center justify-center text-sm transition-colors',
        !chromeless && 'rounded-md border border-border bg-card hover:bg-accent',
        circular && 'rounded-full',
        className,
      )}
    >
      {children}
    </button>
  )
}

export function Separator({ vertical, className }: { vertical?: boolean; className?: string }) {
  return <div className={cn('bg-border', vertical ? 'w-px' : 'h-px w-full', className)} />
}

export function Spinner({ size = 'small', className }: { size?: 'small' | 'large'; color?: string; className?: string }) {
  return (
    <div
      role="status"
      aria-label="Loading"
      className={cn(
        'animate-spin rounded-full border-2 border-current border-t-transparent text-muted-foreground',
        size === 'large' ? 'h-8 w-8' : 'h-4 w-4',
        className,
      )}
    />
  )
}

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'size'> {
  /** React-Native-Web-style text handler (mapped to onChange). */
  onChangeText?: (text: string) => void
  size?: string
}

export function Input({ onChangeText, size: _size, className, ...rest }: InputProps) {
  return (
    <input
      className={cn(
        'h-10 rounded-md border border-border bg-card px-3 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-ring',
        className,
      )}
      onChange={(e) => onChangeText?.(e.target.value)}
      {...rest}
    />
  )
}

interface ImageProps {
  source: { uri: string }
  alt?: string
  width?: number
  height?: number
  className?: string
}

export function Image({ source, alt = '', width, height, className }: ImageProps) {
  return <img src={source.uri} alt={alt} width={width} height={height} className={className} />
}
