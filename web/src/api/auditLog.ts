import apiClient from './client'
import type { AuditLog, AuditLogListParams, PaginatedResponse } from './types'

function buildAuditLogResponse(data: {
  data: AuditLog[]
  total: number
  page: number
  limit: number
  total_pages: number
}): PaginatedResponse<AuditLog> {
  return {
    data: data.data,
    meta: {
      page: data.page,
      per_page: data.limit,
      total: data.total,
      total_pages: data.total_pages,
    },
  }
}

export async function listAuditLog(
  params: AuditLogListParams = {}
): Promise<PaginatedResponse<AuditLog>> {
  const { data } = await apiClient.get('/admin/audit-log', { params })
  return buildAuditLogResponse(data)
}

export async function getAuditLogEntry(id: string): Promise<AuditLog> {
  const { data } = await apiClient.get<AuditLog>(`/admin/audit-log/${id}`)
  return data
}

export async function downloadAuditLogCsv(params: AuditLogListParams = {}): Promise<void> {
  const { data } = await apiClient.get<Blob>('/admin/audit-log/export', {
    params,
    responseType: 'blob',
  })
  const url = URL.createObjectURL(data)
  const a = document.createElement('a')
  a.href = url
  a.download = `audit-log-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
