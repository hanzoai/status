export interface ConditionResult {
  condition: string
  success: boolean
}

export interface EndpointResult {
  status: number
  hostname: string
  duration: number
  success: boolean
  timestamp: string
  conditionResults?: ConditionResult[]
  errors?: string[]
}

export interface EndpointStatus {
  name: string
  group: string
  key: string
  results: EndpointResult[]
  events?: EndpointEvent[]
}

export interface EndpointEvent {
  type: 'HEALTHY' | 'UNHEALTHY' | 'START'
  timestamp: string
}

export interface ResponseTimeHistory {
  timestamps: number[]
  values: number[]
}

export interface Announcement {
  title: string
  description: string
  severity: string
  startTime: string
  endTime: string
  archived?: boolean
}

export interface AppConfig {
  oidc: boolean
  authenticated: boolean
  announcements: Announcement[]
}

export interface UIButton {
  name: string
  link: string
}

export interface WindowConfig {
  title: string
  logo: string
  header: string
  dashboardHeading: string
  dashboardSubheading: string
  link: string
  buttons: UIButton[]
  maximumNumberOfResults: string
  defaultSortBy: string
  defaultFilterBy: string
}

// Suite types
export interface SuiteEndpointResult {
  name: string
  success: boolean
  duration: number
}

export interface SuiteResult {
  success: boolean
  duration: number
  timestamp: string
  endpointResults?: SuiteEndpointResult[]
}

export interface SuiteStatus {
  name: string
  group: string
  key: string
  results: SuiteResult[]
}

declare global {
  interface Window {
    config?: WindowConfig
  }
}
