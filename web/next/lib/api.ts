import type {
  AppConfig,
  EndpointStatus,
  ResponseTimeHistory,
  SuiteStatus,
} from './types'

const fetchOpts: RequestInit = { credentials: 'include' }

export async function fetchConfig(): Promise<AppConfig> {
  const res = await fetch('/api/v1/config', fetchOpts)
  if (!res.ok) throw new Error(`config: ${res.status}`)
  return res.json()
}

export async function fetchEndpointStatuses(
  page = 1,
  pageSize = 50,
): Promise<EndpointStatus[]> {
  const res = await fetch(
    `/api/v1/endpoints/statuses?page=${page}&pageSize=${pageSize}`,
    fetchOpts,
  )
  if (!res.ok) throw new Error(`endpoints: ${res.status}`)
  return res.json()
}

export async function fetchEndpointStatus(
  key: string,
  page = 1,
  pageSize = 50,
): Promise<EndpointStatus> {
  const res = await fetch(
    `/api/v1/endpoints/${key}/statuses?page=${page}&pageSize=${pageSize}`,
    fetchOpts,
  )
  if (!res.ok) throw new Error(`endpoint ${key}: ${res.status}`)
  return res.json()
}

export async function fetchResponseTimeHistory(
  key: string,
  duration: string,
): Promise<ResponseTimeHistory> {
  const res = await fetch(
    `/api/v1/endpoints/${key}/response-times/${duration}/history`,
    fetchOpts,
  )
  if (!res.ok) throw new Error(`history ${key}: ${res.status}`)
  return res.json()
}

export async function fetchSuiteStatuses(
  page = 1,
  pageSize = 50,
): Promise<SuiteStatus[]> {
  const res = await fetch(
    `/api/v1/suites/statuses?page=${page}&pageSize=${pageSize}`,
    fetchOpts,
  )
  if (!res.ok) throw new Error(`suites: ${res.status}`)
  return res.json()
}

export async function fetchSuiteStatus(
  key: string,
  page = 1,
  pageSize = 50,
): Promise<SuiteStatus> {
  const res = await fetch(
    `/api/v1/suites/${key}/statuses?page=${page}&pageSize=${pageSize}`,
    fetchOpts,
  )
  if (!res.ok) throw new Error(`suite ${key}: ${res.status}`)
  return res.json()
}
