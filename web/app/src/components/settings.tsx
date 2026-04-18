import { useState, useEffect, useRef, useCallback } from 'react'
import { Button, Separator } from '@hanzo/gui'
import { RefreshCw, Sun, Moon } from 'lucide-react'
import { getTheme, toggleTheme } from '@/lib/theme'

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
  const [showMenu, setShowMenu] = useState(false)
  const [interval, setIntervalValue] = useState(DEFAULT_INTERVAL)
  const [isDark, setIsDark] = useState(() => getTheme() === 'dark')
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const settingsRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const stored = localStorage.getItem('gatus:refresh-interval')
    if (stored) {
      const v = parseInt(stored)
      if (v >= 10 && INTERVALS.some((i) => i.value === v)) {
        setIntervalValue(v)
      }
    }
  }, [])

  useEffect(() => {
    if (timerRef.current) clearInterval(timerRef.current)
    timerRef.current = setInterval(onRefresh, interval * 1000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [interval, onRefresh])

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

  const handleToggleTheme = () => {
    const next = toggleTheme()
    setIsDark(next === 'dark')
  }

  const formatLabel = (value: number): string => {
    const found = INTERVALS.find((i) => i.value === value)
    return found ? found.label : `${value}s`
  }

  return (
    <div ref={settingsRef} className="fixed bottom-4 left-4 z-50">
      <div className="flex items-center gap-0.5 rounded-full border border-border bg-card/95 p-1 shadow-lg backdrop-blur-sm">
        <Button
          size="$2"
          chromeless
          circular
          onPress={() => setShowMenu(!showMenu)}
          className="relative flex items-center gap-1.5 px-2.5 py-1.5"
        >
          <RefreshCw className="h-3 w-3 text-muted-foreground" />
          <span className="font-mono text-[11px] font-medium text-muted-foreground">{formatLabel(interval)}</span>

          {showMenu && (
            <div
              className="absolute bottom-full left-0 mb-2 overflow-hidden rounded-xl border border-border bg-popover shadow-xl"
              onClick={(e: React.MouseEvent) => e.stopPropagation()}
            >
              {INTERVALS.map((i) => (
                <Button
                  key={i.value}
                  size="$2"
                  chromeless
                  onPress={() => selectInterval(i.value)}
                  className={`block w-full px-4 py-2 text-left font-mono text-[11px] ${
                    interval === i.value ? 'bg-accent' : ''
                  }`}
                >
                  {i.label}
                </Button>
              ))}
            </div>
          )}
        </Button>

        <Separator vertical className="h-4" />

        <Button
          size="$2"
          chromeless
          circular
          onPress={handleToggleTheme}
          aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {isDark ? (
            <Sun className="h-3 w-3 text-muted-foreground" />
          ) : (
            <Moon className="h-3 w-3 text-muted-foreground" />
          )}
        </Button>
      </div>
    </div>
  )
}
