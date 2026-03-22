import apiClient from './client'
import type { AuditLog, AuditLogListParams, PaginatedResponse } from './types'

export async function listAuditLog(
  params: AuditLogListParams = {}
): Promise<PaginatedResponse<AuditLog>> {
  const { data } = await apiClient.get('/admin/audit-log', { params })
  return data
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
