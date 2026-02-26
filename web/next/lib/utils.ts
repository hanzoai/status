import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Human-readable relative time, e.g. "2 hours ago". */
export function timeAgo(timestamp: string | Date): string {
  const diff = Date.now() - new Date(timestamp).getTime()
  if (diff < 500) return 'now'
  if (diff > 3 * 86400000) {
    const days = Math.round(diff / 86400000)
    return `${days} day${days !== 1 ? 's' : ''} ago`
  }
  if (diff > 3600000) {
    const hours = Math.round(diff / 3600000)
    return `${hours} hour${hours !== 1 ? 's' : ''} ago`
  }
  if (diff > 60000) {
    const minutes = Math.round(diff / 60000)
    return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`
  }
  const seconds = Math.round(diff / 1000)
  return `${seconds} second${seconds !== 1 ? 's' : ''} ago`
}

/** Pretty time difference between two timestamps. */
export function timeDifference(start: string | Date, end: string | Date): string {
  const ms = new Date(start).getTime() - new Date(end).getTime()
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    const rem = minutes % 60
    const h = `${hours} hour${hours === 1 ? '' : 's'}`
    return rem > 0 ? `${h} ${rem} minute${rem === 1 ? '' : 's'}` : h
  }
  if (minutes > 0) {
    const rem = seconds % 60
    const m = `${minutes} minute${minutes === 1 ? '' : 's'}`
    return rem > 0 ? `${m} ${rem} second${rem === 1 ? '' : 's'}` : m
  }
  return `${seconds} second${seconds === 1 ? '' : 's'}`
}

/** Format timestamp as YYYY-MM-DD HH:mm:ss. */
export function formatTimestamp(timestamp: string | Date): string {
  const d = new Date(timestamp)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/** Duration in nanoseconds to milliseconds string. */
export function durationMs(ns: number): string {
  return `${Math.round(ns / 1_000_000)}ms`
}

/** Format refresh interval seconds into human label. */
export function formatRefreshInterval(seconds: number): string {
  if (seconds >= 60) return `${seconds / 60}m`
  return `${seconds}s`
}
