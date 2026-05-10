import { apiClient } from './client'

export type CalendarProvider = 'google' | 'microsoft'

export interface CalendarConnection {
  id: string
  provider: CalendarProvider
  token_expiry?: string | null
}

export const calendarApi = {
  listConnections: async (): Promise<CalendarConnection[]> => {
    const { data } = await apiClient.get<{ data?: CalendarConnection[] | null }>('/calendar/connections')
    return Array.isArray(data?.data) ? data.data : []
  },
  disconnect: async (id: string): Promise<void> => {
    await apiClient.delete(`/calendar/connections/${id}`)
  },
  triggerSync: async (): Promise<void> => {
    await apiClient.post('/calendar/sync')
  },
}
