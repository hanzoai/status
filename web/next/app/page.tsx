'use client'

import { useState, useEffect } from 'react'
import { Dashboard } from '@/components/dashboard'
import { EndpointDetail } from '@/components/endpoint-detail'
import { SuiteDetail } from '@/components/suite-detail'
import { Loading } from '@/components/loading'
import { fetchConfig as fetchAppConfig } from '@/lib/api'
import type { AppConfig } from '@/lib/types'
import { LogIn } from 'lucide-react'

/**
 * Single-page app that reads window.location.pathname to determine which view
 * to render. The Go backend serves this same index.html for all routes:
 *   /                  -> Dashboard
 *   /endpoints/:key    -> Endpoint detail
 *   /suites/:key       -> Suite detail
 */
export default function App() {
  const [config, setConfig] = useState<AppConfig | null>(null)
  const [configLoaded, setConfigLoaded] = useState(false)
  const [oidcLoading, setOidcLoading] = useState(false)

  // Determine the current route from the URL
  const [route, setRoute] = useState<{ view: 'dashboard' | 'endpoint' | 'suite'; key?: string }>({ view: 'dashboard' })

  useEffect(() => {
    const resolveRoute = () => {
      const path = window.location.pathname
      const endpointMatch = path.match(/^\/endpoints\/(.+)$/)
      if (endpointMatch) {
        setRoute({ view: 'endpoint', key: endpointMatch[1] })
        return
      }
      const suiteMatch = path.match(/^\/suites\/(.+)$/)
      if (suiteMatch) {
        setRoute({ view: 'suite', key: suiteMatch[1] })
        return
      }
      setRoute({ view: 'dashboard' })
    }

    resolveRoute()

    // Listen for pushState/popState for client-side navigation
    window.addEventListener('popstate', resolveRoute)
    return () => window.removeEventListener('popstate', resolveRoute)
  }, [])

  // Load config
  useEffect(() => {
    fetchAppConfig()
      .then((c) => {
        setConfig(c)
        setConfigLoaded(true)
      })
      .catch(() => setConfigLoaded(true))

    const id = setInterval(() => {
      fetchAppConfig().then(setConfig).catch(() => {})
    }, 600_000)
    return () => clearInterval(id)
  }, [])

  // Client-side navigation helper
  const navigate = (path: string) => {
    window.history.pushState({}, '', path)
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  // Loading config
  if (!configLoaded) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[hsl(var(--background))] text-[hsl(var(--foreground))]">
        <Loading size="lg" />
      </div>
    )
  }

  // OIDC login screen
  if (config?.oidc && !config.authenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[hsl(var(--background))] p-4 text-[hsl(var(--foreground))]">
        <div className="w-full max-w-md rounded-lg border border-[hsl(var(--border))] bg-[hsl(var(--card))]">
          <div className="p-6 text-center">
            <svg
              viewBox="0 0 67 67"
              className="mx-auto mb-4 h-20 w-20 dark:invert"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path d="M22.21 67V44.6369H0V67H22.21Z" fill="#000" />
              <path d="M0 44.6369L22.21 46.8285V44.6369H0Z" fill="#999" />
              <path d="M66.7038 22.3184H22.2534L0.0878906 44.6367H44.4634L66.7038 22.3184Z" fill="#000" />
              <path d="M22.21 0H0V22.3184H22.21V0Z" fill="#000" />
              <path d="M66.7198 0H44.5098V22.3184H66.7198V0Z" fill="#000" />
              <path d="M66.6753 22.3185L44.5098 20.0822V22.3185H66.6753Z" fill="#999" />
              <path d="M66.7198 67V44.6369H44.5098V67H66.7198Z" fill="#000" />
            </svg>
            <h2 className="text-3xl font-bold">Status</h2>
            <p className="mt-2 text-[hsl(var(--muted-foreground))]">Sign in to continue</p>
          </div>
          <div className="px-6 pb-6">
            {typeof window !== 'undefined' &&
              new URLSearchParams(window.location.search).get('error') && (
                <div className="mb-6 rounded-md border border-red-500/20 bg-red-500/10 p-3">
                  <p className="text-center text-sm text-red-600 dark:text-red-400">
                    {new URLSearchParams(window.location.search).get('error') === 'access_denied'
                      ? 'You do not have access to this status page'
                      : new URLSearchParams(window.location.search).get('error')}
                  </p>
                </div>
              )}
            <a
              href="/oidc/login"
              onClick={() => setOidcLoading(true)}
              className="flex h-11 w-full items-center justify-center gap-2 rounded-md bg-[hsl(var(--primary))] px-8 text-sm font-medium text-[hsl(var(--primary-foreground))] transition-colors hover:bg-[hsl(var(--primary)/.9)]"
            >
              {oidcLoading ? (
                <Loading size="xs" />
              ) : (
                <>
                  <LogIn className="h-4 w-4" />
                  Login with OIDC
                </>
              )}
            </a>
          </div>
        </div>
      </div>
    )
  }

  // Main app with routing
  switch (route.view) {
    case 'endpoint':
      return (
        <EndpointDetail
          endpointKey={route.key!}
          announcements={config?.announcements ?? []}
          navigate={navigate}
        />
      )
    case 'suite':
      return (
        <SuiteDetail
          suiteKey={route.key!}
          navigate={navigate}
        />
      )
    default:
      return (
        <Dashboard
          announcements={config?.announcements ?? []}
          navigate={navigate}
        />
      )
  }
}
