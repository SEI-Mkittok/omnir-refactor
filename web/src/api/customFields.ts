import apiClient from './client'
import type {
  CustomFieldDefinition,
  CreateCustomFieldDefinitionRequest,
  UpdateCustomFieldDefinitionRequest,
  CustomFieldEntityType,
} from './types'

export const customFieldsApi = {
  list: async (entity_type?: CustomFieldEntityType): Promise<CustomFieldDefinition[]> => {
    const { data } = await apiClient.get('/custom-field-definitions', {
      params: entity_type ? { entity_type } : undefined,
    })
    return data
  },

  get: async (id: string): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.get(`/custom-field-definitions/${id}`)
    return data
  },

  create: async (payload: CreateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.post('/custom-field-definitions', payload)
    return data
  },

  update: async (id: string, payload: UpdateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.patch(`/custom-field-definitions/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/custom-field-definitions/${id}`)
  },
}
