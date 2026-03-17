import { apiClient } from './client'
import type { ReportsSummary } from './types'

export async function getReportsSummary(): Promise<ReportsSummary> {
  const res = await apiClient.get<ReportsSummary>('/reports')
  return res.data
}
