import apiClient from './client'
import type {
  Activity,
  ActivityType,
  PaginatedResponse,
} from './types'

export interface ActivityListParams {
  page?: number
  per_page?: number
  contact_id?: string
  deal_id?: string
  account_id?: string
  type?: ActivityType
  completed?: boolean
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface CreateActivityRequest {
  type: ActivityType
  subject: string
  description?: string
  due_date?: string
  completed?: boolean
  contact_id?: string
  account_id?: string
  deal_id?: string
  owner_id?: string
}

export interface UpdateActivityRequest extends Partial<CreateActivityRequest> {}

export const activitiesApi = {
  list: async (params?: ActivityListParams): Promise<PaginatedResponse<Activity>> => {
    const { data } = await apiClient.get('/activities', { params })
    return data
  },

  get: async (id: string): Promise<Activity> => {
    const { data } = await apiClient.get(`/activities/${id}`)
    return data
  },

  create: async (payload: CreateActivityRequest): Promise<Activity> => {
    const { data } = await apiClient.post('/activities', payload)
    return data
  },

  update: async (id: string, payload: UpdateActivityRequest): Promise<Activity> => {
    const { data } = await apiClient.patch(`/activities/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/activities/${id}`)
  },
}
