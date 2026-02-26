import type { WindowConfig } from './types'

export function getConfig(): WindowConfig {
  if (typeof window === 'undefined') {
    return {
      logo: '',
      header: 'Status',
      dashboardHeading: '',
      dashboardSubheading: '',
      link: '',
      buttons: [],
      maximumNumberOfResults: '20',
      defaultSortBy: 'name',
      defaultFilterBy: 'none',
    }
  }
  return (
    window.config ?? {
      logo: '',
      header: 'Status',
      dashboardHeading: '',
      dashboardSubheading: '',
      link: '',
      buttons: [],
      maximumNumberOfResults: '20',
      defaultSortBy: 'name',
      defaultFilterBy: 'none',
    }
  )
}

/** Returns the logo URL, or empty string if it's an unrendered Go template placeholder. */
export function getLogo(): string {
  const cfg = getConfig()
  if (!cfg.logo || cfg.logo.includes('{{')) return ''
  return cfg.logo
}

/** Returns the brand link, or null if not configured. */
export function getLink(): string | null {
  const cfg = getConfig()
  if (!cfg.link || cfg.link.includes('{{')) return null
  return cfg.link
}

/** Extract hostname from a URL for display. */
export function getBrandName(link: string | null): string {
  if (!link) return ''
  try {
    return new URL(link).hostname.replace('www.', '')
  } catch {
    return ''
  }
}
