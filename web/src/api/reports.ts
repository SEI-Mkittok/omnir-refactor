import { apiClient } from './client'
import type {
  ReportsSummary,
  TicketReport,
  LeadReport,
  ContactReport,
  DealReport,
  PipelineFunnelReport,
  ConversionRatesReport,
  RevenueProjectionReport,
  ActivitySummaryReport,
  ManagerDashboardReport,
} from './types'

export interface ReportsParams {
  from?: string // ISO 8601 date string
  to?: string   // ISO 8601 date string
}

export async function getReportsSummary(params?: ReportsParams): Promise<ReportsSummary> {
  const res = await apiClient.get<ReportsSummary>('/reports', { params })
  return res.data
}

export async function getTicketReport(params?: ReportsParams): Promise<TicketReport> {
  const res = await apiClient.get<TicketReport>('/reports/tickets', { params })
  return res.data
}

export async function getLeadReport(params?: ReportsParams): Promise<LeadReport> {
  const res = await apiClient.get<LeadReport>('/reports/leads', { params })
  return res.data
}

export async function getContactReport(params?: ReportsParams): Promise<ContactReport> {
  const res = await apiClient.get<ContactReport>('/reports/contacts', { params })
  return res.data
}

export async function getDealReport(params?: ReportsParams): Promise<DealReport> {
  const res = await apiClient.get<DealReport>('/reports/deals', { params })
  return res.data
}

export async function getPipelineFunnelReport(params?: ReportsParams): Promise<PipelineFunnelReport> {
  const res = await apiClient.get<PipelineFunnelReport>('/reports/pipeline-funnel', { params })
  return res.data
}

export async function getConversionRatesReport(params?: ReportsParams): Promise<ConversionRatesReport> {
  const res = await apiClient.get<ConversionRatesReport>('/reports/conversion-rates', { params })
  return res.data
}

export async function getRevenueProjectionReport(months = 3): Promise<RevenueProjectionReport> {
  const res = await apiClient.get<RevenueProjectionReport>('/reports/revenue-projection', { params: { months } })
  return res.data
}

export async function getActivitySummaryReport(params?: ReportsParams): Promise<ActivitySummaryReport> {
  const res = await apiClient.get<ActivitySummaryReport>('/reports/activity-summary', { params })
  return res.data
}

export async function getManagerDashboardReport(params?: ReportsParams): Promise<ManagerDashboardReport> {
  const res = await apiClient.get<ManagerDashboardReport>('/reports/manager-dashboard', { params })
  return res.data
}
