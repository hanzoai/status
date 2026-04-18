import type { WindowConfig } from './types'

const defaults: WindowConfig = {
  title: 'Status',
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

export function getConfig(): WindowConfig {
  return window.config ?? defaults
}

/** Returns the logo URL, or empty string if it is an unrendered Go template placeholder. */
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
