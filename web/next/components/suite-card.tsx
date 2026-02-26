'use client'

import { useMemo } from 'react'
import type { SuiteStatus, EndpointResult } from '@/lib/types'
import { StatusBadge } from './status-badge'
import { HealthBar } from './health-bar'

interface SuiteCardProps {
  suite: SuiteStatus
  maxResults?: number
  onNavigate: (key: string) => void
  onTooltip?: (result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => void
}

export function SuiteCard({
  suite,
  maxResults = 50,
  onNavigate,
  onTooltip,
}: SuiteCardProps) {
  const results = suite.results ?? []
  const latest = results.length > 0 ? results[results.length - 1] : null
  const status = latest ? (latest.success ? 'healthy' : 'unhealthy') : 'unknown'

  // Adapt suite results to look like endpoint results for HealthBar
  const adaptedResults = useMemo(() => {
    return results.map((r) => ({
      status: 0,
      hostname: '',
      duration: r.duration,
      success: r.success,
      timestamp: r.timestamp,
    })) as EndpointResult[]
  }, [results])

  return (
    <div className="flex h-full flex-col rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))] transition hover:scale-[1.01] hover:shadow-lg dark:hover:border-gray-700">
      <div className="space-y-0 px-3 pb-2 pt-3 sm:px-6 sm:pt-6">
        <div className="flex items-start justify-between gap-2 sm:gap-3">
          <div className="min-w-0 flex-1 overflow-hidden">
            <button
              onClick={() => onNavigate(suite.key)}
              className="block truncate text-left text-sm font-semibold hover:text-[hsl(var(--primary))] hover:underline sm:text-base"
              title={suite.name}
            >
              {suite.name}
            </button>
            {suite.group && (
              <div className="flex min-h-[1.25rem] items-center text-xs text-[hsl(var(--muted-foreground))] sm:text-sm">
                <span className="truncate" title={suite.group}>
                  {suite.group}
                </span>
              </div>
            )}
          </div>
          <div className="ml-2 flex-shrink-0">
            <StatusBadge status={status} />
          </div>
        </div>
      </div>
      <div className="flex-1 px-3 pb-3 pt-2 sm:px-6 sm:pb-4">
        <HealthBar
          results={adaptedResults}
          maxResults={maxResults}
          onTooltip={onTooltip}
        />
      </div>
    </div>
  )
}
