import apiClient from './client'
import type { Campaign, CampaignMember, CampaignMemberInput, PaginatedResponse } from './types'

export const campaignsApi = {
  list(params?: { page?: number; limit?: number; status?: string; q?: string }): Promise<PaginatedResponse<Campaign>> {
    return apiClient.get('/campaigns', { params }).then((r) => r.data)
  },

  create(payload: Partial<Campaign>): Promise<Campaign> {
    return apiClient.post('/campaigns', payload).then((r) => r.data)
  },

  update(id: string, payload: Partial<Campaign>): Promise<Campaign> {
    return apiClient.patch(`/campaigns/${id}`, payload).then((r) => r.data)
  },

  delete(id: string): Promise<void> {
    return apiClient.delete(`/campaigns/${id}`).then(() => undefined)
  },

  members(id: string): Promise<{ data: CampaignMember[] }> {
    return apiClient.get(`/campaigns/${id}/members`).then((r) => r.data)
  },

  addMembers(id: string, members: CampaignMemberInput[]): Promise<{ data: CampaignMember[] }> {
    return apiClient.post(`/campaigns/${id}/members`, { members }).then((r) => r.data)
  },
}
