import apiClient from './client'

export type LeadConversionTarget = 'contact' | 'account' | 'deal'

export interface LeadConversionMapping {
  id?: string
  org_id?: string
  lead_field: string
  target_entity: LeadConversionTarget
  target_field: string
  is_active: boolean
  created_at?: string
  updated_at?: string
}

export const leadConversionMappingApi = {
  list: async (): Promise<LeadConversionMapping[]> => {
    const { data } = await apiClient.get('/settings/lead-conversion-mapping')
    return data.data ?? []
  },

  replace: async (mappings: LeadConversionMapping[]): Promise<LeadConversionMapping[]> => {
    const { data } = await apiClient.put('/settings/lead-conversion-mapping', { mappings })
    return data.data ?? []
  },
}
