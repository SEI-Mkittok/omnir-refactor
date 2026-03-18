import apiClient from './client'
import type { Notification, NotificationListParams, PaginatedResponse, UnreadCountResponse } from './types'

export const notificationsApi = {
  list: async (params?: NotificationListParams): Promise<PaginatedResponse<Notification>> => {
    const { data } = await apiClient.get('/notifications', { params })
    return data
  },

  unreadCount: async (): Promise<UnreadCountResponse> => {
    const { data } = await apiClient.get('/notifications/unread-count')
    return data
  },

  markRead: async (id: string): Promise<void> => {
    await apiClient.post(`/notifications/${id}/read`)
  },

  markAllRead: async (): Promise<void> => {
    await apiClient.post('/notifications/read-all')
  },
}
