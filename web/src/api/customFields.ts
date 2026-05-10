import apiClient from './client'
import type {
  CustomFieldDefinition,
  CreateCustomFieldDefinitionRequest,
  UpdateCustomFieldDefinitionRequest,
  CustomFieldEntityType,
} from './types'

export const customFieldsApi = {
  list: async (
    entity_type?: CustomFieldEntityType,
    options?: { activeOptionsOnly?: boolean },
  ): Promise<CustomFieldDefinition[]> => {
    const params: Record<string, string> = {}
    if (entity_type) {
      params.entity_type = entity_type
    }
    if (options?.activeOptionsOnly) {
      params.active_options_only = 'true'
    }
    const { data } = await apiClient.get('/custom-fields', {
      params: Object.keys(params).length > 0 ? params : undefined,
    })
    return data
  },

  get: async (id: string): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.get(`/custom-fields/${id}`)
    return data
  },

  create: async (payload: CreateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.post('/custom-fields', payload)
    return data
  },

  update: async (id: string, payload: UpdateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.patch(`/custom-fields/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/custom-fields/${id}`)
  },
}
