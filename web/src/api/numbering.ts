import apiClient from './client'

export interface NumberingSettings {
  org_id: string
  quote_number_start: number
  ticket_number_start: number
  kb_article_number_start: number
  invoice_number_start: number
  quote_number_prefix: string
  ticket_number_prefix: string
  kb_article_number_prefix: string
  invoice_number_prefix: string
  quote_number_current: number
  ticket_number_current: number
  kb_article_number_current: number
  invoice_number_current: number
}

export interface UpdateNumberingSettingsRequest {
  quote_number_start?: number
  ticket_number_start?: number
  kb_article_number_start?: number
  invoice_number_start?: number
  quote_number_prefix?: string
  ticket_number_prefix?: string
  kb_article_number_prefix?: string
  invoice_number_prefix?: string
  quote_number_current?: number
  ticket_number_current?: number
  kb_article_number_current?: number
  invoice_number_current?: number
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
