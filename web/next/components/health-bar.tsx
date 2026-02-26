'use client'

import { useMemo, useRef, useState, useCallback } from 'react'
import type { EndpointResult } from '@/lib/types'
import { timeAgo } from '@/lib/utils'

interface HealthBarProps {
  results: EndpointResult[]
  maxResults?: number
  onTooltip?: (result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => void
}

export function HealthBar({ results, maxResults = 50, onTooltip }: HealthBarProps) {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)
  const barRefs = useRef<(HTMLDivElement | null)[]>([])

  // Pad results to maxResults with null at the front
  const display = useMemo(() => {
    const arr: (EndpointResult | null)[] = [...results]
    while (arr.length < maxResults) arr.unshift(null)
    return arr.slice(-maxResults)
  }, [results, maxResults])

  const oldestTime = results.length > 0
    ? timeAgo(results[Math.max(0, results.length - maxResults)].timestamp)
    : ''
  const newestTime = results.length > 0
    ? timeAgo(results[results.length - 1].timestamp)
    : ''

  const handleMouseEnter = useCallback(
    (r: EndpointResult | null, i: number) => {
      if (r && onTooltip) onTooltip(r, barRefs.current[i], 'hover')
    },
    [onTooltip],
  )

  const handleMouseLeave = useCallback(() => {
    if (onTooltip) onTooltip(null, null, 'hover')
  }, [onTooltip])

  const handleClick = useCallback(
    (r: EndpointResult | null, i: number) => {
      if (!r || !onTooltip) return
      window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
      if (selectedIndex === i) {
        setSelectedIndex(null)
        onTooltip(null, null, 'click')
      } else {
        setSelectedIndex(i)
        onTooltip(r, barRefs.current[i], 'click')
      }
    },
    [onTooltip, selectedIndex],
  )

  // Listen for clear events from other cards
  const clearRef = useRef<(() => void) | null>(null)
  clearRef.current = () => setSelectedIndex(null)

  return (
    <div>
      <div className="flex gap-[3px]" data-health-bar>
        {display.map((r, i) => (
          <div
            key={i}
            ref={(el) => { barRefs.current[i] = el }}
            className={`flex-1 h-7 rounded-[3px] transition-all duration-150 ${
              r
                ? r.success
                  ? selectedIndex === i
                    ? 'bg-emerald-600 cursor-pointer'
                    : 'bg-emerald-500/80 hover:bg-emerald-500 cursor-pointer'
                  : selectedIndex === i
                    ? 'bg-red-600 cursor-pointer'
                    : 'bg-red-500/80 hover:bg-red-500 cursor-pointer'
                : 'bg-[hsl(var(--muted))]'
            }`}
            onMouseEnter={() => handleMouseEnter(r, i)}
            onMouseLeave={handleMouseLeave}
            onClick={(e) => {
              e.stopPropagation()
              handleClick(r, i)
            }}
          />
        ))}
      </div>
      <div className="mt-1.5 flex items-center justify-between font-mono text-[10px] text-[hsl(var(--muted-foreground)/.6)]">
        <span>{oldestTime}</span>
        <span>{newestTime}</span>
      </div>
    </div>
  )
}
