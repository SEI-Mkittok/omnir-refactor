import apiClient from './client'

export interface UserPreferences {
  id?: string
  user_id?: string
  org_id?: string
  default_currency?: string
  number_format?: string
  default_record_view?: string
  landing_page?: string
  address_line1?: string
  address_line2?: string
  city?: string
  state?: string
  postal_code?: string
  country?: string
  photo_url?: string
  tags?: string[]
  service_preferences?: Record<string, unknown>
  calendar_start_day?: string
  calendar_date_format?: string
  calendar_time_zone?: string
  calendar_default_activity_status?: string
  calendar_default_duration_minutes?: number
  calendar_reminder_interval_minutes?: number
  calendar_default_view?: 'month' | 'week' | string
  calendar_day_start_hour?: number
  calendar_hour_format?: '12h' | '24h' | string
  calendar_default_activity_type?: string
  calendar_show_completed_events?: boolean
  created_at?: string
  updated_at?: string
}

export type UpdateUserPreferencesRequest = Partial<UserPreferences>

export const preferencesApi = {
  get: async (): Promise<UserPreferences> => {
    const { data } = await apiClient.get('/settings/preferences')
    return data
  },

  update: async (payload: UpdateUserPreferencesRequest): Promise<UserPreferences> => {
    const { data } = await apiClient.patch('/settings/preferences', payload)
    return data
  },

  getCalendar: async (): Promise<UserPreferences> => {
    const { data } = await apiClient.get('/settings/calendar-preferences')
    return data
  },

  updateCalendar: async (payload: UpdateUserPreferencesRequest): Promise<UserPreferences> => {
    const { data } = await apiClient.patch('/settings/calendar-preferences', payload)
    return data
  },
}
