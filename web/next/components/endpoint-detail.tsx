'use client'

import { useState, useEffect, useMemo, useCallback } from 'react'
import {
  ArrowLeft,
  RefreshCw,
  ArrowUpCircle,
  ArrowDownCircle,
  PlayCircle,
  Activity,
  Timer,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { fetchEndpointStatus } from '@/lib/api'
import type { EndpointStatus, EndpointResult, EndpointEvent, Announcement } from '@/lib/types'
import { timeAgo, timeDifference, formatTimestamp } from '@/lib/utils'
import { Header } from './header'
import { Footer } from './footer'
import { Loading } from './loading'
import { StatusBadge } from './status-badge'
import { HealthBar } from './health-bar'
import { Settings } from './settings'
import { Tooltip } from './tooltip'
import { ResponseTimeChart } from './response-time-chart'

const PAGE_SIZE = 50

interface ProcessedEvent extends EndpointEvent {
  fancyText: string
  fancyTimeAgo: string
}

interface EndpointDetailProps {
  endpointKey: string
  announcements: Announcement[]
  navigate: (path: string) => void
}

export function EndpointDetail({ endpointKey, navigate }: EndpointDetailProps) {
  const [endpoint, setEndpoint] = useState<EndpointStatus | null>(null)
  const [currentStatus, setCurrentStatus] = useState<EndpointStatus | null>(null)
  const [events, setEvents] = useState<ProcessedEvent[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [showAvgResponseTime, setShowAvgResponseTime] = useState(true)
  const [chartDuration, setChartDuration] = useState('24h')
  const [showChart, setShowChart] = useState(false)

  const [tooltipResult, setTooltipResult] = useState<EndpointResult | null>(null)
  const [tooltipAnchor, setTooltipAnchor] = useState<HTMLElement | null>(null)
  const [tooltipPersistent, setTooltipPersistent] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem('gatus:show-average-response-time')
    setShowAvgResponseTime(stored !== 'false')
  }, [])

  const fetchData = useCallback(async () => {
    setIsRefreshing(true)
    try {
      const data = await fetchEndpointStatus(endpointKey, currentPage, PAGE_SIZE)
      setEndpoint(data)
      if (currentPage === 1) setCurrentStatus(data)

      if (data.events?.length) {
        const processed: ProcessedEvent[] = []
        for (let i = data.events.length - 1; i >= 0; i--) {
          const ev = data.events[i]
          let fancyText = ''
          if (i === data.events.length - 1) {
            fancyText = ev.type === 'UNHEALTHY' ? 'Endpoint is unhealthy' : ev.type === 'HEALTHY' ? 'Endpoint is healthy' : 'Monitoring started'
          } else {
            if (ev.type === 'HEALTHY') {
              fancyText = 'Endpoint became healthy'
            } else if (ev.type === 'UNHEALTHY') {
              const next = data.events[i + 1]
              fancyText = next ? `Endpoint was unhealthy for ${timeDifference(next.timestamp, ev.timestamp)}` : 'Endpoint became unhealthy'
            } else {
              fancyText = 'Monitoring started'
            }
          }
          processed.push({ ...ev, fancyText, fancyTimeAgo: timeAgo(ev.timestamp) })
        }
        setEvents(processed)
      }

      if (data.results?.some((r) => r.duration > 0)) setShowChart(true)
    } catch (err) {
      console.error('[EndpointDetail] fetch error:', err)
    } finally {
      setIsRefreshing(false)
    }
  }, [endpointKey, currentPage])

  useEffect(() => { fetchData() }, [fetchData])

  const latestResult = currentStatus?.results?.length ? currentStatus.results[currentStatus.results.length - 1] : null
  const healthStatus = latestResult ? (latestResult.success ? 'healthy' : 'unhealthy') : 'unknown'
  const hostname = latestResult?.hostname ?? null

  const avgResponseTime = useMemo(() => {
    if (!endpoint?.results?.length) return 'N/A'
    let total = 0, count = 0
    for (const r of endpoint.results) { if (r.duration) { total += r.duration; count++ } }
    return count > 0 ? `${Math.round(total / count / 1_000_000)}ms` : 'N/A'
  }, [endpoint])

  const responseTimeRange = useMemo(() => {
    if (!endpoint?.results?.length) return 'N/A'
    let min = Infinity, max = 0, hasData = false
    for (const r of endpoint.results) { if (r.duration) { min = Math.min(min, r.duration); max = Math.max(max, r.duration); hasData = true } }
    if (!hasData) return 'N/A'
    const minMs = Math.trunc(min / 1_000_000), maxMs = Math.trunc(max / 1_000_000)
    return minMs === maxMs ? `${minMs}ms` : `${minMs}-${maxMs}ms`
  }, [endpoint])

  const lastCheckTime = currentStatus?.results?.length ? timeAgo(currentStatus.results[currentStatus.results.length - 1].timestamp) : 'Never'

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

  const toggleResponseTimeDisplay = () => {
    setShowAvgResponseTime((prev) => { const next = !prev; localStorage.setItem('gatus:show-average-response-time', next ? 'true' : 'false'); return next })
  }

  const healthBadgeUrl = `/api/v1/endpoints/${endpointKey}/health/badge.svg`
  const uptimeBadgeUrl = (d: string) => `/api/v1/endpoints/${endpointKey}/uptimes/${d}/badge.svg`
  const responseTimeBadgeUrl = (d: string) => `/api/v1/endpoints/${endpointKey}/response-times/${d}/badge.svg`

  return (
    <div className="relative min-h-screen">
      <Header />
      <main className="relative">
        <div className="container mx-auto max-w-7xl px-4 py-8">
          <button onClick={() => navigate('/')} className="mb-4 inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-[hsl(var(--accent))]">
            <ArrowLeft className="h-4 w-4" />Back to Dashboard
          </button>

          {endpoint?.name ? (
            <div className="space-y-6">
              <div className="flex items-start justify-between">
                <div>
                  <h1 className="text-4xl font-bold tracking-tight">{endpoint.name}</h1>
                  <div className="mt-2 flex items-center gap-3 text-[hsl(var(--muted-foreground))]">
                    {endpoint.group && <span>Group: {endpoint.group}</span>}
                    {endpoint.group && hostname && <span>{'\u2022'}</span>}
                    {hostname && <span>{hostname}</span>}
                  </div>
                </div>
                <StatusBadge status={healthStatus as 'healthy' | 'unhealthy' | 'unknown'} />
              </div>

              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                {[
                  { label: 'Current Status', value: healthStatus === 'healthy' ? 'Operational' : 'Issues Detected' },
                  { label: 'Avg Response Time', value: avgResponseTime },
                  { label: 'Response Time Range', value: responseTimeRange },
                  { label: 'Last Check', value: lastCheckTime },
                ].map((card) => (
                  <div key={card.label} className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                    <div className="px-6 pb-2 pt-6"><p className="text-sm font-medium text-[hsl(var(--muted-foreground))]">{card.label}</p></div>
                    <div className="px-6 pb-6"><div className="text-2xl font-bold">{card.value}</div></div>
                  </div>
                ))}
              </div>

              <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                <div className="flex items-center justify-between px-6 pt-6">
                  <h2 className="text-lg font-semibold">Recent Checks</h2>
                  <div className="flex items-center gap-2">
                    <button onClick={toggleResponseTimeDisplay} title={showAvgResponseTime ? 'Show min-max response time' : 'Show average response time'} className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[hsl(var(--accent))] transition-colors">
                      {showAvgResponseTime ? <Activity className="h-5 w-5" /> : <Timer className="h-5 w-5" />}
                    </button>
                    <button onClick={fetchData} disabled={isRefreshing} title="Refresh data" className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[hsl(var(--accent))] transition-colors">
                      <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
                    </button>
                  </div>
                </div>
                <div className="p-6">
                  {endpoint.results?.length > 0 && <HealthBar results={endpoint.results} maxResults={PAGE_SIZE} onTooltip={handleTooltip} />}
                  <div className="mt-4 flex items-center justify-center gap-2 border-t border-[hsl(var(--border))] pt-4">
                    <button disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => Math.max(1, p - 1))} className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-[hsl(var(--border))] disabled:opacity-50 hover:bg-[hsl(var(--accent))]"><ChevronLeft className="h-4 w-4" /></button>
                    <span className="text-sm text-[hsl(var(--muted-foreground))]">Page {currentPage}</span>
                    <button disabled={!endpoint.results?.length || endpoint.results.length < PAGE_SIZE} onClick={() => setCurrentPage((p) => p + 1)} className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-[hsl(var(--border))] disabled:opacity-50 hover:bg-[hsl(var(--accent))]"><ChevronRight className="h-4 w-4" /></button>
                  </div>
                </div>
              </div>

              {showChart && (
                <div className="space-y-6">
                  <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                    <div className="flex items-center justify-between px-6 pt-6">
                      <h2 className="text-lg font-semibold">Response Time Trend</h2>
                      <select value={chartDuration} onChange={(e) => setChartDuration(e.target.value)} className="rounded-md border border-[hsl(var(--border))] bg-[hsl(var(--background))] px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]">
                        <option value="24h">24 hours</option>
                        <option value="7d">7 days</option>
                        <option value="30d">30 days</option>
                      </select>
                    </div>
                    <div className="p-6"><ResponseTimeChart endpointKey={endpointKey} duration={chartDuration} /></div>
                  </div>

                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    {['30d', '7d', '24h', '1h'].map((period) => (
                      <div key={period} className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                        <div className="px-6 pb-2 pt-6 text-center"><p className="text-sm font-medium text-[hsl(var(--muted-foreground))]">{period === '30d' ? 'Last 30 days' : period === '7d' ? 'Last 7 days' : period === '24h' ? 'Last 24 hours' : 'Last hour'}</p></div>
                        <div className="px-6 pb-6 text-center"><img src={responseTimeBadgeUrl(period)} alt={`${period} response time`} className="mx-auto mt-2" /></div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                <div className="px-6 pt-6"><h2 className="text-lg font-semibold">Uptime Statistics</h2></div>
                <div className="p-6">
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    {['30d', '7d', '24h', '1h'].map((period) => (
                      <div key={period} className="text-center">
                        <p className="mb-2 text-sm text-[hsl(var(--muted-foreground))]">{period === '30d' ? 'Last 30 days' : period === '7d' ? 'Last 7 days' : period === '24h' ? 'Last 24 hours' : 'Last hour'}</p>
                        <img src={uptimeBadgeUrl(period)} alt={`${period} uptime`} className="mx-auto" />
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                <div className="px-6 pt-6"><h2 className="text-lg font-semibold">Current Health</h2></div>
                <div className="p-6 text-center"><img src={healthBadgeUrl} alt="health badge" className="mx-auto" /></div>
              </div>

              {events.length > 0 && (
                <div className="rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
                  <div className="px-6 pt-6"><h2 className="text-lg font-semibold">Events</h2></div>
                  <div className="p-6 space-y-4">
                    {events.map((ev, i) => (
                      <div key={i} className="flex items-start gap-4 border-b border-[hsl(var(--border))] pb-4 last:border-0">
                        <div className="mt-1">
                          {ev.type === 'HEALTHY' ? <ArrowUpCircle className="h-5 w-5 text-green-500" /> : ev.type === 'UNHEALTHY' ? <ArrowDownCircle className="h-5 w-5 text-red-500" /> : <PlayCircle className="h-5 w-5 text-[hsl(var(--muted-foreground))]" />}
                        </div>
                        <div className="flex-1">
                          <p className="font-medium">{ev.fancyText}</p>
                          <p className="text-sm text-[hsl(var(--muted-foreground))]">{formatTimestamp(ev.timestamp)} {'\u2022'} {ev.fancyTimeAgo}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
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
