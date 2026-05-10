import apiClient from './client'

export interface OrgCurrency {
  id?: string
  org_id?: string
  code: string
  display_name: string
  symbol: string
  decimal_places: number
  is_active: boolean
  is_default: boolean
  created_at?: string
  updated_at?: string
}

export interface OrgCurrencyInput {
  code: string
  display_name: string
  symbol: string
  decimal_places: number
  is_active: boolean
}

export interface CurrenciesSettings {
  default_code: string
  currencies: OrgCurrency[]
}

export interface UpdateCurrenciesRequest {
  default_code: string
  currencies: OrgCurrencyInput[]
}

export const currenciesApi = {
  get: async (): Promise<CurrenciesSettings> => {
    const { data } = await apiClient.get('/settings/currencies')
    return data
  },

  update: async (payload: UpdateCurrenciesRequest): Promise<CurrenciesSettings> => {
    const { data } = await apiClient.patch('/settings/currencies', payload)
    return data
  },
}
