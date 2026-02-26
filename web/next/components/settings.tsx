'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { RefreshCw, Sun, Moon } from 'lucide-react'
import { useTheme } from 'next-themes'

const INTERVALS = [
  { value: 10, label: '10s' },
  { value: 30, label: '30s' },
  { value: 60, label: '1m' },
  { value: 120, label: '2m' },
  { value: 300, label: '5m' },
  { value: 600, label: '10m' },
]

const DEFAULT_INTERVAL = 300

interface SettingsProps {
  onRefresh: () => void
}

export function Settings({ onRefresh }: SettingsProps) {
  const { resolvedTheme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)
  const [showMenu, setShowMenu] = useState(false)
  const [interval, setIntervalValue] = useState(DEFAULT_INTERVAL)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const settingsRef = useRef<HTMLDivElement>(null)

  useEffect(() => setMounted(true), [])

  // Load stored interval
  useEffect(() => {
    const stored = localStorage.getItem('gatus:refresh-interval')
    if (stored) {
      const v = parseInt(stored)
      if (v >= 10 && INTERVALS.some((i) => i.value === v)) {
        setIntervalValue(v)
      }
    }
  }, [])

  // Set up auto-refresh
  useEffect(() => {
    if (timerRef.current) clearInterval(timerRef.current)
    timerRef.current = setInterval(onRefresh, interval * 1000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [interval, onRefresh])

  // Click outside to close menu
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (settingsRef.current && !settingsRef.current.contains(e.target as Node)) {
        setShowMenu(false)
      }
    }
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [])

  const selectInterval = useCallback(
    (value: number) => {
      setIntervalValue(value)
      localStorage.setItem('gatus:refresh-interval', String(value))
      setShowMenu(false)
      onRefresh()
    },
    [onRefresh],
  )

  const toggleTheme = () => {
    const next = resolvedTheme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    document.cookie = `theme=${next}; path=/; max-age=31536000; samesite=strict`
  }

  const formatLabel = (value: number): string => {
    const found = INTERVALS.find((i) => i.value === value)
    return found ? found.label : `${value}s`
  }

  const isDark = resolvedTheme === 'dark'

  return (
    <div ref={settingsRef} className="fixed bottom-4 left-4 z-50">
      <div className="flex items-center gap-0.5 rounded-full border border-[hsl(var(--border))] bg-[hsl(var(--card)/.95)] p-1 shadow-lg backdrop-blur-sm">
        {/* Refresh interval button */}
        <button
          onClick={() => setShowMenu(!showMenu)}
          className="relative flex items-center gap-1.5 rounded-full px-2.5 py-1.5 transition-colors hover:bg-[hsl(var(--accent))]"
        >
          <RefreshCw className="h-3 w-3 text-[hsl(var(--muted-foreground))]" />
          <span className="font-mono text-[11px] font-medium text-[hsl(var(--muted-foreground))]">{formatLabel(interval)}</span>

          {showMenu && (
            <div
              className="absolute bottom-full left-0 mb-2 overflow-hidden rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--popover))] shadow-xl"
              onClick={(e) => e.stopPropagation()}
            >
              {INTERVALS.map((i) => (
                <button
                  key={i.value}
                  onClick={() => selectInterval(i.value)}
                  className={`block w-full px-4 py-2 text-left font-mono text-[11px] transition-colors hover:bg-[hsl(var(--accent))] ${
                    interval === i.value ? 'bg-[hsl(var(--accent))] text-[hsl(var(--foreground))]' : 'text-[hsl(var(--muted-foreground))]'
                  }`}
                >
                  {i.label}
                </button>
              ))}
            </div>
          )}
        </button>

        {/* Divider */}
        <div className="h-4 w-px bg-[hsl(var(--border))]" />

        {/* Theme toggle */}
        {mounted && (
          <button
            onClick={toggleTheme}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            className="rounded-full p-1.5 transition-colors hover:bg-[hsl(var(--accent))]"
          >
            {isDark ? (
              <Sun className="h-3 w-3 text-[hsl(var(--muted-foreground))]" />
            ) : (
              <Moon className="h-3 w-3 text-[hsl(var(--muted-foreground))]" />
            )}
          </button>
        )}
      </div>
    </div>
  )
}
