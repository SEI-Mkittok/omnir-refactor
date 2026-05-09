import apiClient from './client'

export interface NumberingSettings {
  id: string
  org_id: string
  quote_number_start: number
  ticket_number_start: number
  kb_article_number_start: number
  invoice_number_start: number
  created_at: string
  updated_at: string
}

export interface UpdateNumberingSettingsRequest {
  quote_number_start?: number
  ticket_number_start?: number
  kb_article_number_start?: number
  invoice_number_start?: number
}

export const numberingApi = {
  get: async (): Promise<NumberingSettings> => {
    const { data } = await apiClient.get('/settings/numbering')
    return data
  },

  update: async (payload: UpdateNumberingSettingsRequest): Promise<NumberingSettings> => {
    const { data } = await apiClient.patch('/settings/numbering', payload)
    return data
  },
}
