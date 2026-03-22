import { apiClient } from './client'

// ── Types ────────────────────────────────────────────────────────────────────

export interface WidgetPosition {
  x: number
  y: number
  w: number
  h: number
}

export interface Widget {
  type: string
  query_params: Record<string, unknown>
  position: WidgetPosition
}

export interface CustomDashboard {
  id: string
  org_id: string
  name: string
  widgets: Widget[]
  created_by: string
  created_at: string
  updated_at: string
}

export interface ScheduledReport {
  id: string
  org_id: string
  dashboard_id: string
  schedule: string
  recipients: string[]
  last_sent_at?: string
  created_at: string
  updated_at: string
}

export interface DashboardRunResult {
  dashboard_id: string
  widgets: Array<{ widget: Widget; data: unknown }>
}

// ── Dashboard CRUD ───────────────────────────────────────────────────────────

export async function listDashboards(): Promise<CustomDashboard[]> {
  const res = await apiClient.get<{ data: CustomDashboard[] }>('/dashboards')
  return res.data.data
}

export async function getDashboard(id: string): Promise<CustomDashboard> {
  const res = await apiClient.get<CustomDashboard>(`/dashboards/${id}`)
  return res.data
}

export async function createDashboard(payload: {
  name: string
  widgets: Widget[]
}): Promise<CustomDashboard> {
  const res = await apiClient.post<CustomDashboard>('/dashboards', payload)
  return res.data
}

export async function updateDashboard(
  id: string,
  patch: { name?: string; widgets?: Widget[] }
): Promise<CustomDashboard> {
  const res = await apiClient.patch<CustomDashboard>(`/dashboards/${id}`, patch)
  return res.data
}

export async function deleteDashboard(id: string): Promise<void> {
  await apiClient.delete(`/dashboards/${id}`)
}

export async function runDashboard(id: string): Promise<DashboardRunResult> {
  const res = await apiClient.post<DashboardRunResult>(`/dashboards/${id}/run`)
  return res.data
}

// ── Schedules CRUD ───────────────────────────────────────────────────────────

export async function listSchedules(): Promise<ScheduledReport[]> {
  const res = await apiClient.get<{ data: ScheduledReport[] }>('/reports/schedules')
  return res.data.data
}

export async function createSchedule(payload: {
  dashboard_id: string
  schedule: string
  recipients: string[]
}): Promise<ScheduledReport> {
  const res = await apiClient.post<ScheduledReport>('/reports/schedules', payload)
  return res.data
}

export async function updateSchedule(
  id: string,
  patch: { dashboard_id?: string; schedule?: string; recipients?: string[] }
): Promise<ScheduledReport> {
  const res = await apiClient.patch<ScheduledReport>(`/reports/schedules/${id}`, patch)
  return res.data
}

export async function deleteSchedule(id: string): Promise<void> {
  await apiClient.delete(`/reports/schedules/${id}`)
}
