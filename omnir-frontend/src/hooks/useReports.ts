import { useQuery } from '@tanstack/react-query'
import { getReportsSummary } from '@/api/reports'

export const reportKeys = {
  summary: ['reports', 'summary'] as const,
}

export function useReportsSummary() {
  return useQuery({
    queryKey: reportKeys.summary,
    queryFn: getReportsSummary,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}
