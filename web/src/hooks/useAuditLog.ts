import { useQuery } from '@tanstack/react-query'
import { listAuditLog, getAuditLogEntry } from '@/api/auditLog'
import type { AuditLogListParams } from '@/api/types'

const auditLogKeys = {
  all: ['audit-log'] as const,
  lists: () => [...auditLogKeys.all, 'list'] as const,
  list: (params?: AuditLogListParams) => [...auditLogKeys.lists(), params] as const,
  detail: (id: string) => [...auditLogKeys.all, 'detail', id] as const,
}

export function useAuditLog(params?: AuditLogListParams) {
  return useQuery({
    queryKey: auditLogKeys.list(params),
    queryFn: () => listAuditLog(params),
    staleTime: 30_000,
  })
}

export function useAuditLogEntry(id: string | null) {
  return useQuery({
    queryKey: auditLogKeys.detail(id ?? ''),
    queryFn: () => getAuditLogEntry(id!),
    enabled: !!id,
    staleTime: 60_000,
  })
}
