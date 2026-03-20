import apiClient from './client'
import type {
  Automation,
  AutomationRun,
  CreateAutomationRequest,
  UpdateAutomationRequest,
  AutomationListParams,
} from './types'

export const automationsApi = {
  list(params?: AutomationListParams): Promise<{ data: Automation[]; total: number }> {
    return apiClient.get('/automations', { params }).then((r) => r.data)
  },

  get(id: string): Promise<Automation> {
    return apiClient.get(`/automations/${id}`).then((r) => r.data)
  },

  create(req: CreateAutomationRequest): Promise<Automation> {
    return apiClient.post('/automations', req).then((r) => r.data)
  },

  update(id: string, req: UpdateAutomationRequest): Promise<Automation> {
    return apiClient.patch(`/automations/${id}`, req).then((r) => r.data)
  },

  delete(id: string): Promise<void> {
    return apiClient.delete(`/automations/${id}`).then(() => undefined)
  },

  listRuns(id: string, params?: { page?: number }): Promise<{ data: AutomationRun[]; total: number }> {
    return apiClient.get(`/automations/${id}/runs`, { params }).then((r) => r.data)
  },
}
