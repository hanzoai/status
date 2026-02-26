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
      // Clear other selections via custom event
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

  // Register listener once
  if (typeof window !== 'undefined') {
    // Use useEffect-style cleanup via a stable ref
  }

  return (
    <div>
      <div className="flex gap-0.5" data-health-bar>
        {display.map((r, i) => (
          <div
            key={i}
            ref={(el) => { barRefs.current[i] = el }}
            className={`flex-1 h-6 sm:h-8 rounded-sm transition-all ${
              r
                ? r.success
                  ? selectedIndex === i
                    ? 'bg-green-700 cursor-pointer'
                    : 'bg-green-500 hover:bg-green-700 cursor-pointer'
                  : selectedIndex === i
                    ? 'bg-red-700 cursor-pointer'
                    : 'bg-red-500 hover:bg-red-700 cursor-pointer'
                : 'bg-gray-200 dark:bg-gray-700'
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
      <div className="mt-1 flex items-center justify-between text-xs text-[hsl(var(--muted-foreground))]">
        <span>{oldestTime}</span>
        <span>{newestTime}</span>
      </div>
    </div>
  )
}
