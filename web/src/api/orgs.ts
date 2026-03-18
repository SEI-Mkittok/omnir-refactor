import { apiClient } from './client'
import type { Org, CreateOrgRequest, OrgListParams, PaginatedResponse } from './types'

export const orgsApi = {
  list: (params?: OrgListParams) =>
    apiClient.get<PaginatedResponse<Org>>('/orgs', { params }).then((r) => r.data),

  get: (id: string) =>
    apiClient.get<Org>(`/orgs/${id}`).then((r) => r.data),

  create: (payload: CreateOrgRequest) =>
    apiClient.post<Org>('/orgs', payload).then((r) => r.data),

  switchTo: (id: string) =>
    apiClient.post<{ user: import('./types').User }>(`/orgs/${id}/switch`).then((r) => r.data),
}
