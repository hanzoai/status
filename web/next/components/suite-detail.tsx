'use client'

import { useState, useEffect, useCallback } from 'react'
import { ArrowLeft, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react'
import { fetchSuiteStatus } from '@/lib/api'
import type { SuiteStatus, EndpointResult } from '@/lib/types'
import { timeAgo } from '@/lib/utils'
import { Header } from './header'
import { Footer } from './footer'
import { Loading } from './loading'
import { StatusBadge } from './status-badge'
import { HealthBar } from './health-bar'
import { Settings } from './settings'
import { Tooltip } from './tooltip'

const PAGE_SIZE = 50

interface SuiteDetailProps {
  suiteKey: string
  navigate: (path: string) => void
}

export function SuiteDetail({ suiteKey, navigate }: SuiteDetailProps) {
  const [suite, setSuite] = useState<SuiteStatus | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [tooltipResult, setTooltipResult] = useState<EndpointResult | null>(null)
  const [tooltipAnchor, setTooltipAnchor] = useState<HTMLElement | null>(null)
  const [tooltipPersistent, setTooltipPersistent] = useState(false)

  const fetchData = useCallback(async () => {
    setIsRefreshing(true)
    try {
      const data = await fetchSuiteStatus(suiteKey, currentPage, PAGE_SIZE)
      setSuite(data)
    } catch (err) {
      console.error('[SuiteDetail] fetch error:', err)
    } finally {
      setIsRefreshing(false)
    }
  }, [suiteKey, currentPage])

  useEffect(() => { fetchData() }, [fetchData])

  const latestResult = suite?.results?.length ? suite.results[suite.results.length - 1] : null
  const healthStatus = latestResult ? (latestResult.success ? 'healthy' : 'unhealthy') : 'unknown'

  const adaptedResults: EndpointResult[] = (suite?.results ?? []).map((r) => ({
    status: 0, hostname: '', duration: r.duration, success: r.success, timestamp: r.timestamp,
  }))

  const handleTooltip = useCallback((result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => {
    if (action === 'click') {
      if (!result) { setTooltipResult(null); setTooltipAnchor(null); setTooltipPersistent(false) }
      else { setTooltipResult(result); setTooltipAnchor(anchor); setTooltipPersistent(true) }
    } else if (!tooltipPersistent) { setTooltipResult(result); setTooltipAnchor(anchor) }
  }, [tooltipPersistent])

  const dismissTooltip = useCallback(() => {
    setTooltipResult(null); setTooltipAnchor(null); setTooltipPersistent(false)
    window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
  }, [])

  const lastCheckTime = latestResult ? timeAgo(latestResult.timestamp) : 'Never'

  return (
    <div className="relative min-h-screen">
      <Header />
      <main className="relative">
        <div className="container mx-auto max-w-7xl px-4 py-8">
          <button onClick={() => navigate('/')} className="mb-4 inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-[hsl(var(--accent))]">
            <ArrowLeft className="h-4 w-4" />Back to Dashboard
          </button>

          {suite?.name ? (
            <div className="space-y-6">
              <div className="flex items-start justify-between">
                <div>
                  <h1 className="text-4xl font-bold tracking-tight">{suite.name}</h1>
                  {suite.group && <div className="mt-2 text-[hsl(var(--muted-foreground))]">Group: {suite.group}</div>}
                </div>
                <StatusBadge status={healthStatus as 'healthy' | 'unhealthy' | 'unknown'} />
              </div>

              <div className="grid gap-6 md:grid-cols-2">
                <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                  <div className="px-6 pb-2 pt-6"><p className="text-sm font-medium text-[hsl(var(--muted-foreground))]">Current Status</p></div>
                  <div className="px-6 pb-6"><div className="text-2xl font-bold">{healthStatus === 'healthy' ? 'All Passing' : 'Failures Detected'}</div></div>
                </div>
                <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                  <div className="px-6 pb-2 pt-6"><p className="text-sm font-medium text-[hsl(var(--muted-foreground))]">Last Check</p></div>
                  <div className="px-6 pb-6"><div className="text-2xl font-bold">{lastCheckTime}</div></div>
                </div>
              </div>

              <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                <div className="flex items-center justify-between px-6 pt-6">
                  <h2 className="text-lg font-semibold">Recent Checks</h2>
                  <button onClick={fetchData} disabled={isRefreshing} title="Refresh data" className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[hsl(var(--accent))] transition-colors">
                    <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
                  </button>
                </div>
                <div className="p-6">
                  {adaptedResults.length > 0 && <HealthBar results={adaptedResults} maxResults={PAGE_SIZE} onTooltip={handleTooltip} />}
                  <div className="mt-4 flex items-center justify-center gap-2 border-t border-[hsl(var(--border))] pt-4">
                    <button disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => Math.max(1, p - 1))} className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-[hsl(var(--border))] disabled:opacity-50 hover:bg-[hsl(var(--accent))]"><ChevronLeft className="h-4 w-4" /></button>
                    <span className="text-sm text-[hsl(var(--muted-foreground))]">Page {currentPage}</span>
                    <button disabled={!suite.results?.length || suite.results.length < PAGE_SIZE} onClick={() => setCurrentPage((p) => p + 1)} className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-[hsl(var(--border))] disabled:opacity-50 hover:bg-[hsl(var(--accent))]"><ChevronRight className="h-4 w-4" /></button>
                  </div>
                </div>
              </div>

              <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                <div className="px-6 pt-6"><h2 className="text-lg font-semibold">Current Health</h2></div>
                <div className="p-6 text-center"><img src={`/api/v1/endpoints/${suiteKey}/health/badge.svg`} alt="health badge" className="mx-auto" /></div>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center py-20"><Loading size="lg" /></div>
          )}
        </div>
      </main>
      <Footer />
      <Settings onRefresh={fetchData} />
      <Tooltip result={tooltipResult} anchor={tooltipAnchor} persistent={tooltipPersistent} onDismiss={dismissTooltip} />
    </div>
  )
}
