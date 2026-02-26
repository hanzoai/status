'use client'

import { useState, useEffect, useMemo, useCallback } from 'react'
import {
  Activity,
  Timer,
  RefreshCw,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  CheckCircle,
} from 'lucide-react'
import { fetchEndpointStatuses, fetchSuiteStatuses } from '@/lib/api'
import type { EndpointStatus, EndpointResult, Announcement, SuiteStatus } from '@/lib/types'
import { Header } from './header'
import { Footer } from './footer'
import { Loading } from './loading'
import { EndpointCard } from './endpoint-card'
import { SuiteCard } from './suite-card'
import { SearchBar } from './search-bar'
import { Settings } from './settings'
import { Tooltip } from './tooltip'
import { AnnouncementBanner, PastAnnouncements } from './announcement-banner'

const RESULT_PAGE_SIZE = 50

interface DashboardProps {
  announcements: Announcement[]
  navigate: (path: string) => void
}

export function Dashboard({ announcements, navigate }: DashboardProps) {
  const [endpoints, setEndpoints] = useState<EndpointStatus[]>([])
  const [suites, setSuites] = useState<SuiteStatus[]>([])
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [filter, setFilter] = useState('none')
  const [sortBy, setSortBy] = useState('name')
  const [showAvgResponseTime, setShowAvgResponseTime] = useState(true)
  const [uncollapsedGroups, setUncollapsedGroups] = useState<Set<string>>(new Set())

  // Tooltip state
  const [tooltipResult, setTooltipResult] = useState<EndpointResult | null>(null)
  const [tooltipAnchor, setTooltipAnchor] = useState<HTMLElement | null>(null)
  const [tooltipPersistent, setTooltipPersistent] = useState(false)

  // Load preferences
  useEffect(() => {
    const stored = localStorage.getItem('gatus:show-average-response-time')
    setShowAvgResponseTime(stored !== 'false')
    try {
      const saved = localStorage.getItem('gatus:uncollapsed-groups')
      if (saved) setUncollapsedGroups(new Set(JSON.parse(saved)))
    } catch { /* ignore */ }
  }, [])

  // Fetch data
  const fetchData = useCallback(async () => {
    const isInitial = endpoints.length === 0 && suites.length === 0
    if (isInitial) setLoading(true)
    try {
      const [endpointData, suiteData] = await Promise.allSettled([
        fetchEndpointStatuses(1, RESULT_PAGE_SIZE),
        fetchSuiteStatuses(1, RESULT_PAGE_SIZE),
      ])
      if (endpointData.status === 'fulfilled') setEndpoints(endpointData.value)
      if (suiteData.status === 'fulfilled') setSuites(suiteData.value ?? [])
    } catch (err) {
      console.error('[Dashboard] fetch error:', err)
    } finally {
      if (isInitial) setLoading(false)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const refreshData = useCallback(() => {
    setEndpoints([])
    setSuites([])
    fetchData()
  }, [fetchData])

  // Filtering
  const filteredEndpoints = useMemo(() => {
    let filtered = [...endpoints]
    if (searchQuery) {
      const q = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (e) =>
          e.name.toLowerCase().includes(q) ||
          (e.group && e.group.toLowerCase().includes(q)),
      )
    }
    if (filter === 'failing') {
      filtered = filtered.filter((e) => {
        if (!e.results?.length) return false
        return !e.results[e.results.length - 1].success
      })
    } else if (filter === 'unstable') {
      filtered = filtered.filter((e) => {
        if (!e.results?.length) return false
        return e.results.some((r) => !r.success)
      })
    }
    if (sortBy === 'health') {
      filtered.sort((a, b) => {
        const aOk = a.results?.length ? a.results[a.results.length - 1].success : true
        const bOk = b.results?.length ? b.results[b.results.length - 1].success : true
        if (!aOk && bOk) return -1
        if (aOk && !bOk) return 1
        return a.name.localeCompare(b.name)
      })
    }
    return filtered
  }, [endpoints, searchQuery, filter, sortBy])

  const filteredSuites = useMemo(() => {
    let filtered = [...suites]
    if (searchQuery) {
      const q = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (s) =>
          s.name.toLowerCase().includes(q) ||
          (s.group && s.group.toLowerCase().includes(q)),
      )
    }
    if (filter === 'failing') {
      filtered = filtered.filter((s) => {
        if (!s.results?.length) return false
        return !s.results[s.results.length - 1].success
      })
    } else if (filter === 'unstable') {
      filtered = filtered.filter((s) => {
        if (!s.results?.length) return false
        return s.results.some((r) => !r.success)
      })
    }
    if (sortBy === 'health') {
      filtered.sort((a, b) => {
        const aOk = a.results?.length ? a.results[a.results.length - 1].success : true
        const bOk = b.results?.length ? b.results[b.results.length - 1].success : true
        if (!aOk && bOk) return -1
        if (aOk && !bOk) return 1
        return a.name.localeCompare(b.name)
      })
    }
    return filtered
  }, [suites, searchQuery, filter, sortBy])

  // Grouping
  const isGrouped = sortBy === 'group'

  const combinedGroups = useMemo(() => {
    if (!isGrouped) return null
    const groups: Record<string, { endpoints: EndpointStatus[]; suites: SuiteStatus[] }> = {}
    for (const ep of filteredEndpoints) {
      const g = ep.group || 'No Group'
      if (!groups[g]) groups[g] = { endpoints: [], suites: [] }
      groups[g].endpoints.push(ep)
    }
    for (const s of filteredSuites) {
      const g = s.group || 'No Group'
      if (!groups[g]) groups[g] = { endpoints: [], suites: [] }
      groups[g].suites.push(s)
    }
    const sorted = Object.keys(groups).sort((a, b) => {
      if (a === 'No Group') return 1
      if (b === 'No Group') return -1
      return a.localeCompare(b)
    })
    const result: [string, { endpoints: EndpointStatus[]; suites: SuiteStatus[] }][] = []
    for (const key of sorted) result.push([key, groups[key]])
    return result
  }, [isGrouped, filteredEndpoints, filteredSuites])

  // Announcements
  const activeAnnouncements = useMemo(
    () => (announcements ?? []).filter((a) => !a.archived),
    [announcements],
  )
  const archivedAnnouncements = useMemo(
    () => (announcements ?? []).filter((a) => a.archived),
    [announcements],
  )

  // Tooltip handlers
  const handleTooltip = useCallback(
    (result: EndpointResult | null, anchor: HTMLElement | null, action: 'hover' | 'click') => {
      if (action === 'click') {
        if (!result) {
          setTooltipResult(null)
          setTooltipAnchor(null)
          setTooltipPersistent(false)
        } else {
          setTooltipResult(result)
          setTooltipAnchor(anchor)
          setTooltipPersistent(true)
        }
      } else {
        if (!tooltipPersistent) {
          setTooltipResult(result)
          setTooltipAnchor(anchor)
        }
      }
    },
    [tooltipPersistent],
  )

  const dismissTooltip = useCallback(() => {
    setTooltipResult(null)
    setTooltipAnchor(null)
    setTooltipPersistent(false)
    window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
  }, [])

  // Toggle group collapse
  const toggleGroup = useCallback((group: string) => {
    setUncollapsedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(group)) next.delete(group)
      else next.add(group)
      localStorage.setItem('gatus:uncollapsed-groups', JSON.stringify([...next]))
      return next
    })
  }, [])

  const toggleResponseTimeDisplay = () => {
    setShowAvgResponseTime((prev) => {
      const next = !prev
      localStorage.setItem('gatus:show-average-response-time', next ? 'true' : 'false')
      return next
    })
  }

  const unhealthyCount = (eps: EndpointStatus[]) =>
    eps.filter((e) => e.results?.length && !e.results[e.results.length - 1].success).length

  const failingSuitesCount = (ss: SuiteStatus[]) =>
    ss.filter((s) => s.results?.length && !s.results[s.results.length - 1].success).length

  return (
    <div className="relative min-h-screen">
      <Header />
      <main className="relative">
        <div className="container mx-auto max-w-7xl px-4 py-8">
          <div className="mb-6">
            <div className="mb-4 flex items-center justify-end gap-2">
              <button
                onClick={toggleResponseTimeDisplay}
                title={showAvgResponseTime ? 'Show min-max response time' : 'Show average response time'}
                className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[hsl(var(--accent))] transition-colors"
              >
                {showAvgResponseTime ? <Activity className="h-5 w-5" /> : <Timer className="h-5 w-5" />}
              </button>
              <button
                onClick={refreshData}
                title="Refresh data"
                className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[hsl(var(--accent))] transition-colors"
              >
                <RefreshCw className="h-5 w-5" />
              </button>
            </div>

            <AnnouncementBanner announcements={activeAnnouncements} />
            <SearchBar onSearch={setSearchQuery} onFilterChange={setFilter} onSortChange={setSortBy} />
          </div>

          {loading && (
            <div className="flex items-center justify-center py-20">
              <Loading size="lg" />
            </div>
          )}

          {!loading && filteredEndpoints.length === 0 && filteredSuites.length === 0 && (
            <div className="py-20 text-center">
              <AlertCircle className="mx-auto mb-4 h-12 w-12 text-[hsl(var(--muted-foreground))]" />
              <h3 className="mb-2 text-lg font-semibold">No endpoints or suites found</h3>
              <p className="text-[hsl(var(--muted-foreground))]">
                {searchQuery || filter !== 'none'
                  ? 'Try adjusting your filters'
                  : 'No endpoints or suites are configured'}
              </p>
            </div>
          )}

          {!loading && (filteredEndpoints.length > 0 || filteredSuites.length > 0) && (
            <>
              {isGrouped && combinedGroups ? (
                <div className="space-y-6">
                  {combinedGroups.map(([group, items]) => (
                    <div key={group} className="overflow-hidden rounded-lg border border-[hsl(var(--border))]">
                      <button
                        onClick={() => toggleGroup(group)}
                        className="flex w-full items-center justify-between border-b border-[hsl(var(--border))] bg-[hsl(var(--card))] p-4 transition-colors hover:bg-[hsl(var(--accent)/.5)]"
                      >
                        <div className="flex items-center gap-3">
                          {uncollapsedGroups.has(group) ? (
                            <ChevronDown className="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
                          ) : (
                            <ChevronUp className="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
                          )}
                          <h2 className="text-xl font-semibold">{group}</h2>
                        </div>
                        <div className="flex items-center gap-2">
                          {unhealthyCount(items.endpoints) + failingSuitesCount(items.suites) > 0 ? (
                            <span className="rounded-full bg-red-600 px-2 py-1 text-sm font-medium text-white">
                              {unhealthyCount(items.endpoints) + failingSuitesCount(items.suites)}
                            </span>
                          ) : (
                            <CheckCircle className="h-6 w-6 text-green-600" />
                          )}
                        </div>
                      </button>

                      {uncollapsedGroups.has(group) && (
                        <div className="p-4">
                          {items.suites.length > 0 && (
                            <div className="mb-4">
                              <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">Suites</h3>
                              <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                                {items.suites.map((s) => (
                                  <SuiteCard key={s.key} suite={s} maxResults={RESULT_PAGE_SIZE} onNavigate={(k) => navigate(`/suites/${k}`)} onTooltip={handleTooltip} />
                                ))}
                              </div>
                            </div>
                          )}
                          {items.endpoints.length > 0 && (
                            <div>
                              {items.suites.length > 0 && (
                                <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-[hsl(var(--muted-foreground))]">Endpoints</h3>
                              )}
                              <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                                {items.endpoints.map((ep) => (
                                  <EndpointCard key={ep.key} endpoint={ep} maxResults={RESULT_PAGE_SIZE} showAverageResponseTime={showAvgResponseTime} onNavigate={(k) => navigate(`/endpoints/${k}`)} onTooltip={handleTooltip} />
                                ))}
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div>
                  {filteredSuites.length > 0 && (
                    <div className="mb-6">
                      <h2 className="mb-3 text-lg font-semibold">Suites</h2>
                      <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                        {filteredSuites.map((s) => (
                          <SuiteCard key={s.key} suite={s} maxResults={RESULT_PAGE_SIZE} onNavigate={(k) => navigate(`/suites/${k}`)} onTooltip={handleTooltip} />
                        ))}
                      </div>
                    </div>
                  )}
                  {filteredEndpoints.length > 0 && (
                    <div>
                      {filteredSuites.length > 0 && <h2 className="mb-3 text-lg font-semibold">Endpoints</h2>}
                      <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                        {filteredEndpoints.map((ep) => (
                          <EndpointCard key={ep.key} endpoint={ep} maxResults={RESULT_PAGE_SIZE} showAverageResponseTime={showAvgResponseTime} onNavigate={(k) => navigate(`/endpoints/${k}`)} onTooltip={handleTooltip} />
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </>
          )}

          {archivedAnnouncements.length > 0 && (
            <div className="mt-12 pb-8">
              <PastAnnouncements announcements={archivedAnnouncements} />
            </div>
          )}
        </div>
      </main>
      <Footer />
      <Settings onRefresh={fetchData} />
      <Tooltip result={tooltipResult} anchor={tooltipAnchor} persistent={tooltipPersistent} onDismiss={dismissTooltip} />
    </div>
  )
}
