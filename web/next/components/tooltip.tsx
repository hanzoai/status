'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import type { EndpointResult } from '@/lib/types'
import { formatTimestamp } from '@/lib/utils'

interface TooltipProps {
  result: EndpointResult | null
  anchor: HTMLElement | null
  persistent: boolean
  onDismiss: () => void
}

export function Tooltip({ result, anchor, persistent, onDismiss }: TooltipProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState({ top: 0, left: 0 })
  const [visible, setVisible] = useState(false)

  const updatePosition = useCallback(() => {
    if (!anchor || !ref.current) return
    const r = anchor.getBoundingClientRect()
    const t = ref.current.getBoundingClientRect()
    const scrollTop = window.scrollY
    const scrollLeft = window.scrollX

    let top = r.bottom + scrollTop + 8
    let left = r.left + scrollLeft

    // Flip above if no room below
    if (window.innerHeight - r.bottom < t.height + 20 && r.top > t.height + 20) {
      top = r.top + scrollTop - t.height - 8
    }

    // Prevent overflow right
    if (window.innerWidth - r.left < t.width + 20) {
      left = r.right + scrollLeft - t.width
      if (left < scrollLeft + 10) left = scrollLeft + 10
    }

    setPos({ top: Math.round(top), left: Math.round(left) })
  }, [anchor])

  useEffect(() => {
    if (result && anchor) {
      setVisible(true)
      // Give the DOM a tick to measure
      requestAnimationFrame(updatePosition)
    } else if (!persistent) {
      setVisible(false)
    }
  }, [result, anchor, persistent, updatePosition])

  useEffect(() => {
    if (!visible) return
    window.addEventListener('resize', updatePosition)
    return () => window.removeEventListener('resize', updatePosition)
  }, [visible, updatePosition])

  // Click outside to dismiss persistent tooltip
  useEffect(() => {
    if (!persistent) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        const bar = (e.target as HTMLElement).closest('[data-health-bar]')
        if (!bar) onDismiss()
      }
    }
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [persistent, onDismiss])

  if (!visible || !result) return null

  return (
    <div
      ref={ref}
      className="absolute z-50 rounded-md border border-[hsl(var(--border))] bg-[hsl(var(--popover))] px-3 py-2 text-sm shadow-lg text-[hsl(var(--popover-foreground))]"
      style={{ top: pos.top, left: pos.left }}
    >
      <div className="space-y-2">
        <div>
          <div className="text-xs font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">
            Timestamp
          </div>
          <div className="font-mono text-xs">{formatTimestamp(result.timestamp)}</div>
        </div>
        <div>
          <div className="text-xs font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">
            Response Time
          </div>
          <div className="font-mono text-xs">
            {Math.trunc(result.duration / 1_000_000)}ms
          </div>
        </div>
        {result.conditionResults && result.conditionResults.length > 0 && (
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">
              Conditions
            </div>
            <div className="space-y-0.5 font-mono text-xs">
              {result.conditionResults.map((cr, i) => (
                <div key={i} className="flex items-start gap-1">
                  <span className={cr.success ? 'text-green-500' : 'text-red-500'}>
                    {cr.success ? '\u2713' : '\u2717'}
                  </span>
                  <span className="break-all">{cr.condition}</span>
                </div>
              ))}
            </div>
          </div>
        )}
        {result.errors && result.errors.length > 0 && (
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">
              Errors
            </div>
            <div className="space-y-0.5 font-mono text-xs">
              {result.errors.map((err, i) => (
                <div key={i} className="text-red-500">
                  {'\u2022'} {err}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
