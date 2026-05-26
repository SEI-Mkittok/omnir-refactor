import apiClient from './client'
import type { MailConverterPreview, MailConverterRule, MailConverterRun, PaginatedResponse } from './types'

export const mailConverterApi = {
  listRules(params?: { page?: number; limit?: number; status?: string }): Promise<PaginatedResponse<MailConverterRule>> {
    return apiClient.get('/mail-converter/rules', { params }).then((r) => r.data)
  },

  createRule(payload: Partial<MailConverterRule>): Promise<MailConverterRule> {
    return apiClient.post('/mail-converter/rules', payload).then((r) => r.data)
  },

  updateRule(id: string, payload: Partial<MailConverterRule>): Promise<MailConverterRule> {
    return apiClient.patch(`/mail-converter/rules/${id}`, payload).then((r) => r.data)
  },

  deleteRule(id: string): Promise<void> {
    return apiClient.delete(`/mail-converter/rules/${id}`).then(() => undefined)
  },

  previewRule(id: string, limit = 50): Promise<MailConverterPreview> {
    return apiClient.post(`/mail-converter/rules/${id}/preview`, null, { params: { limit } }).then((r) => r.data)
  },

  scanRule(id: string, limit = 250): Promise<MailConverterRun> {
    return apiClient.post(`/mail-converter/rules/${id}/scan`, null, { params: { limit } }).then((r) => r.data)
  },

  runs(id: string): Promise<{ data: MailConverterRun[] }> {
    return apiClient.get(`/mail-converter/rules/${id}/runs`).then((r) => r.data)
  },
}
