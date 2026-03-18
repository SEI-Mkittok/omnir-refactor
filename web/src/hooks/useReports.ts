import { useQuery } from '@tanstack/react-query'
import {
  getReportsSummary,
  getTicketReport,
  getLeadReport,
  getContactReport,
  getDealReport,
  type ReportsParams,
} from '@/api/reports'
import type {
  ReportsSummary,
  TicketReport,
  LeadReport,
  ContactReport,
  DealReport,
} from '@/api/types'

export type { ReportsParams }

export const reportKeys = {
  summary: ['reports', 'summary'] as const,
  tickets: (params: ReportsParams) => ['reports', 'tickets', params] as const,
  leads: (params: ReportsParams) => ['reports', 'leads', params] as const,
  contacts: (params: ReportsParams) => ['reports', 'contacts', params] as const,
  deals: (params: ReportsParams) => ['reports', 'deals', params] as const,
}

// ---- Mock data (placeholder until OMN-222 backend is ready) ----

function buildMockTicketReport(): TicketReport {
  const today = new Date()
  const over_time = Array.from({ length: 30 }, (_, i) => {
    const d = new Date(today)
    d.setDate(d.getDate() - (29 - i))
    return {
      date: d.toISOString().slice(0, 10),
      count: Math.floor(Math.random() * 8) + 1,
    }
  })
  return {
    open_count: 34,
    avg_resolution_hours: 18.5,
    breach_rate: 0.12,
    over_time,
    by_status: [
      { status: 'open', count: 34 },
      { status: 'pending', count: 12 },
      { status: 'resolved', count: 87 },
      { status: 'closed', count: 210 },
    ],
    total_closed: 297,
  }
}

function buildMockLeadReport(): LeadReport {
  return {
    new_count: 48,
    converted_count: 17,
    conversion_rate: 0.354,
    funnel: [
      { stage: 'new', label: 'New', count: 48 },
      { stage: 'contacted', label: 'Contacted', count: 36 },
      { stage: 'qualified', label: 'Qualified', count: 24 },
      { stage: 'converted', label: 'Converted', count: 17 },
    ],
  }
}

function buildMockContactReport(): ContactReport {
  return {
    new_count: 31,
    total_count: 284,
    over_time: [
      { month: '2025-10', count: 18 },
      { month: '2025-11', count: 22 },
      { month: '2025-12', count: 15 },
      { month: '2026-01', count: 27 },
      { month: '2026-02', count: 24 },
      { month: '2026-03', count: 31 },
    ],
  }
}

function buildMockDealReport(): DealReport {
  return {
    pipeline_value_cents: 28450000,
    by_stage: [
      { stage: 'lead', count: 12, total_value_cents: 1800000 },
      { stage: 'qualified', count: 8, total_value_cents: 4200000 },
      { stage: 'proposal', count: 5, total_value_cents: 6750000 },
      { stage: 'negotiation', count: 3, total_value_cents: 5500000 },
      { stage: 'closed_won', count: 7, total_value_cents: 10200000 },
      { stage: 'closed_lost', count: 4, total_value_cents: 0 },
    ],
    won_count: 7,
    lost_count: 4,
  }
}

// ---- Hooks ----

export function useReportsSummary() {
  return useQuery<ReportsSummary>({
    queryKey: reportKeys.summary,
    queryFn: getReportsSummary,
    staleTime: 5 * 60 * 1000,
  })
}

export function useTicketReport(params: ReportsParams = {}) {
  return useQuery<TicketReport>({
    queryKey: reportKeys.tickets(params),
    queryFn: () => getTicketReport(params),
    staleTime: 5 * 60 * 1000,
    placeholderData: buildMockTicketReport,
  })
}

export function useLeadReport(params: ReportsParams = {}) {
  return useQuery<LeadReport>({
    queryKey: reportKeys.leads(params),
    queryFn: () => getLeadReport(params),
    staleTime: 5 * 60 * 1000,
    placeholderData: buildMockLeadReport,
  })
}

export function useContactReport(params: ReportsParams = {}) {
  return useQuery<ContactReport>({
    queryKey: reportKeys.contacts(params),
    queryFn: () => getContactReport(params),
    staleTime: 5 * 60 * 1000,
    placeholderData: buildMockContactReport,
  })
}

export function useDealReport(params: ReportsParams = {}) {
  return useQuery<DealReport>({
    queryKey: reportKeys.deals(params),
    queryFn: () => getDealReport(params),
    staleTime: 5 * 60 * 1000,
    placeholderData: buildMockDealReport,
  })
}
