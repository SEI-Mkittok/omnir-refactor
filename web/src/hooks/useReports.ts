import { useQuery } from '@tanstack/react-query'
import {
  getReportsSummary,
  getTicketReport,
  getLeadReport,
  getContactReport,
  getDealReport,
  getPipelineFunnelReport,
  getConversionRatesReport,
  getRevenueProjectionReport,
  getActivitySummaryReport,
  getManagerDashboardReport,
  type ReportsParams,
} from '@/api/reports'
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
} from '@/api/types'

export type { ReportsParams }

export const reportKeys = {
  summary: ['reports', 'summary'] as const,
  tickets: (params: ReportsParams) => ['reports', 'tickets', params] as const,
  leads: (params: ReportsParams) => ['reports', 'leads', params] as const,
  contacts: (params: ReportsParams) => ['reports', 'contacts', params] as const,
  deals: (params: ReportsParams) => ['reports', 'deals', params] as const,
  pipelineFunnel: (params: ReportsParams) => ['reports', 'pipeline-funnel', params] as const,
  conversionRates: (params: ReportsParams) => ['reports', 'conversion-rates', params] as const,
  revenueProjection: (months: number) => ['reports', 'revenue-projection', months] as const,
  activitySummary: (params: ReportsParams) => ['reports', 'activity-summary', params] as const,
  managerDashboard: (params: ReportsParams) => ['reports', 'manager-dashboard', params] as const,
}

// ---- Hooks ----

export function useReportsSummary(params: ReportsParams = {}) {
  return useQuery<ReportsSummary>({
    queryKey: [...reportKeys.summary, params] as const,
    queryFn: () => getReportsSummary(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useTicketReport(params: ReportsParams = {}) {
  return useQuery<TicketReport>({
    queryKey: reportKeys.tickets(params),
    queryFn: () => getTicketReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useLeadReport(params: ReportsParams = {}) {
  return useQuery<LeadReport>({
    queryKey: reportKeys.leads(params),
    queryFn: () => getLeadReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useContactReport(params: ReportsParams = {}) {
  return useQuery<ContactReport>({
    queryKey: reportKeys.contacts(params),
    queryFn: () => getContactReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useDealReport(params: ReportsParams = {}) {
  return useQuery<DealReport>({
    queryKey: reportKeys.deals(params),
    queryFn: () => getDealReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function usePipelineFunnelReport(params: ReportsParams = {}) {
  return useQuery<PipelineFunnelReport>({
    queryKey: reportKeys.pipelineFunnel(params),
    queryFn: () => getPipelineFunnelReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useConversionRatesReport(params: ReportsParams = {}) {
  return useQuery<ConversionRatesReport>({
    queryKey: reportKeys.conversionRates(params),
    queryFn: () => getConversionRatesReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useRevenueProjectionReport(months = 3) {
  return useQuery<RevenueProjectionReport>({
    queryKey: reportKeys.revenueProjection(months),
    queryFn: () => getRevenueProjectionReport(months),
    staleTime: 5 * 60 * 1000,
  })
}

export function useActivitySummaryReport(params: ReportsParams = {}) {
  return useQuery<ActivitySummaryReport>({
    queryKey: reportKeys.activitySummary(params),
    queryFn: () => getActivitySummaryReport(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useManagerDashboardReport(params: ReportsParams = {}) {
  return useQuery<ManagerDashboardReport>({
    queryKey: reportKeys.managerDashboard(params),
    queryFn: () => getManagerDashboardReport(params),
    staleTime: 5 * 60 * 1000,
  })
}
