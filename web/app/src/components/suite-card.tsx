import { useMemo } from 'react'
import { Card } from '@hanzo/gui'
import type { SuiteStatus, EndpointResult } from '@/lib/types'
import { StatusBadge } from './status-badge'
import { HealthBar } from './health-bar'

interface SuiteCardProps {
  suite: SuiteStatus
  maxResults?: number
  onNavigate: (key: string) => void
  onTooltip?: (result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => void
}

export function SuiteCard({ suite, maxResults = 50, onNavigate, onTooltip }: SuiteCardProps) {
  const results = suite.results ?? []
  const latest = results.length > 0 ? results[results.length - 1] : null
  const status = latest ? (latest.success ? 'healthy' : 'unhealthy') : 'unknown'

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
    <Card className="flex h-full flex-col transition-all duration-200 hover:opacity-90" bordered padded={false}>
      <div className="space-y-0 px-4 pb-2 pt-4">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1 overflow-hidden">
            <button
              onClick={() => onNavigate(suite.key)}
              className="block truncate text-left text-sm font-medium text-foreground transition-colors hover:text-brand sm:text-[15px]"
              title={suite.name}
            >
              {suite.name}
            </button>
            {suite.group && (
              <p className="truncate text-xs text-muted-foreground" title={suite.group}>{suite.group}</p>
            )}
          </div>
          <div className="ml-2 flex-shrink-0">
            <StatusBadge status={status} />
          </div>
        </div>
      </div>
      <div className="flex-1 px-4 pb-4 pt-2">
        <HealthBar results={adaptedResults} maxResults={maxResults} onTooltip={onTooltip} />
      </div>
    </Card>
  )
}
