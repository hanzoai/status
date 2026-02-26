'use client'

import { useMemo } from 'react'
import type { EndpointResult, EndpointStatus } from '@/lib/types'
import { StatusBadge } from './status-badge'
import { HealthBar } from './health-bar'

interface EndpointCardProps {
  endpoint: EndpointStatus
  maxResults?: number
  showAverageResponseTime?: boolean
  onNavigate: (key: string) => void
  onTooltip?: (result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => void
}

export function EndpointCard({
  endpoint,
  maxResults = 50,
  showAverageResponseTime = true,
  onNavigate,
  onTooltip,
}: EndpointCardProps) {
  const results = endpoint.results ?? []
  const latest = results.length > 0 ? results[results.length - 1] : null
  const status = latest ? (latest.success ? 'healthy' : 'unhealthy') : 'unknown'
  const hostname = latest?.hostname ?? null

  const responseTimeText = useMemo(() => {
    if (results.length === 0) return 'N/A'
    let total = 0
    let count = 0
    let min = Infinity
    let max = 0
    for (const r of results) {
      if (r.duration) {
        const ms = r.duration / 1_000_000
        total += ms
        count++
        min = Math.min(min, ms)
        max = Math.max(max, ms)
      }
    }
    if (count === 0) return 'N/A'
    if (showAverageResponseTime) {
      return `~${Math.round(total / count)}ms`
    }
    const minMs = Math.trunc(min)
    const maxMs = Math.trunc(max)
    return minMs === maxMs ? `${minMs}ms` : `${minMs}-${maxMs}ms`
  }, [results, showAverageResponseTime])

  return (
    <div className="flex h-full flex-col rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))] transition hover:scale-[1.01] hover:shadow-lg dark:hover:border-gray-700">
      {/* Header */}
      <div className="space-y-0 px-3 pb-2 pt-3 sm:px-6 sm:pt-6">
        <div className="flex items-start justify-between gap-2 sm:gap-3">
          <div className="min-w-0 flex-1 overflow-hidden">
            <button
              onClick={() => onNavigate(endpoint.key)}
              className="block truncate text-left text-sm font-semibold hover:text-[hsl(var(--primary))] hover:underline sm:text-base"
              title={endpoint.name}
            >
              {endpoint.name}
            </button>
            <div className="flex min-h-[1.25rem] items-center gap-2 text-xs text-[hsl(var(--muted-foreground))] sm:text-sm">
              {endpoint.group && (
                <span className="truncate" title={endpoint.group}>
                  {endpoint.group}
                </span>
              )}
              {endpoint.group && hostname && <span>{'\u2022'}</span>}
              {hostname && (
                <span className="truncate" title={hostname}>
                  {hostname}
                </span>
              )}
            </div>
          </div>
          <div className="ml-2 flex-shrink-0">
            <StatusBadge status={status} />
          </div>
        </div>
      </div>
      {/* Content */}
      <div className="flex-1 px-3 pb-3 pt-2 sm:px-6 sm:pb-4">
        <div className="space-y-2">
          <div className="mb-1 flex items-center justify-between">
            <div className="flex-1" />
            <p
              className="text-xs text-[hsl(var(--muted-foreground))]"
              title={showAverageResponseTime ? 'Average response time' : 'Min-max response time'}
            >
              {responseTimeText}
            </p>
          </div>
          <HealthBar
            results={results}
            maxResults={maxResults}
            onTooltip={onTooltip}
          />
        </div>
      </div>
    </div>
  )
}
