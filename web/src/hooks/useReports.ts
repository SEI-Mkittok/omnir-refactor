import { useQuery } from '@tanstack/react-query'
import {
  getReportsSummary,
  getTicketReport,
  getLeadReport,
  getContactReport,
  getDealReport,
  getPipelineFunnelReport,
  getActivitySummaryReport,
  type ReportsParams,
} from '@/api/reports'
import type {
  ReportsSummary,
  TicketReport,
  LeadReport,
  ContactReport,
  DealReport,
  PipelineFunnelReport,
  ActivitySummaryReport,
} from '@/api/types'

export type { ReportsParams }

export const reportKeys = {
  summary: ['reports', 'summary'] as const,
  tickets: (params: ReportsParams) => ['reports', 'tickets', params] as const,
  leads: (params: ReportsParams) => ['reports', 'leads', params] as const,
  contacts: (params: ReportsParams) => ['reports', 'contacts', params] as const,
  deals: (params: ReportsParams) => ['reports', 'deals', params] as const,
  pipelineFunnel: (params: ReportsParams) => ['reports', 'pipeline-funnel', params] as const,
  activitySummary: (params: ReportsParams) => ['reports', 'activity-summary', params] as const,
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

export function useActivitySummaryReport(params: ReportsParams = {}) {
  return useQuery<ActivitySummaryReport>({
    queryKey: reportKeys.activitySummary(params),
    queryFn: () => getActivitySummaryReport(params),
    staleTime: 5 * 60 * 1000,
  })
}
