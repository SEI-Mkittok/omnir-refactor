import apiClient from './client'
import type { PaginatedResponse, Webform, WebformPreviewResponse, WebformSubmission } from './types'

export const webformsApi = {
  list(params?: { page?: number; limit?: number; status?: string; target_module?: string; campaign_id?: string; q?: string }): Promise<PaginatedResponse<Webform>> {
    return apiClient.get('/webforms', { params }).then((r) => r.data)
  },

  create(payload: Partial<Webform>): Promise<Webform> {
    return apiClient.post('/webforms', payload).then((r) => r.data)
  },

  update(id: string, payload: Partial<Webform>): Promise<Webform> {
    return apiClient.patch(`/webforms/${id}`, payload).then((r) => r.data)
  },

  delete(id: string): Promise<void> {
    return apiClient.delete(`/webforms/${id}`).then(() => undefined)
  },

  preview(id: string, payload: Record<string, unknown>): Promise<WebformPreviewResponse> {
    return apiClient.post(`/webforms/${id}/preview`, { payload }).then((r) => r.data)
  },

  submissions(params?: { page?: number; limit?: number; webform_id?: string; campaign_id?: string }): Promise<PaginatedResponse<WebformSubmission>> {
    return apiClient.get('/submissions', { params }).then((r) => r.data)
  },
}
