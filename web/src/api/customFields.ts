import apiClient from './client'
import type {
  CustomFieldDefinition,
  CreateCustomFieldDefinitionRequest,
  UpdateCustomFieldDefinitionRequest,
  CustomFieldEntityType,
} from './types'

export const customFieldsApi = {
  list: async (entity_type?: CustomFieldEntityType): Promise<CustomFieldDefinition[]> => {
    const { data } = await apiClient.get('/custom-fields', {
    const { data } = await apiClient.get('/custom-fields', {
      params: entity_type ? { entity_type } : undefined,
    })
    return data
  },

  get: async (id: string): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.get(`/custom-fields/${id}`)
    const { data } = await apiClient.get(`/custom-fields/${id}`)
    return data
  },

  create: async (payload: CreateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.post('/custom-fields', payload)
    const { data } = await apiClient.post('/custom-fields', payload)
    return data
  },

  update: async (id: string, payload: UpdateCustomFieldDefinitionRequest): Promise<CustomFieldDefinition> => {
    const { data } = await apiClient.patch(`/custom-fields/${id}`, payload)
    const { data } = await apiClient.patch(`/custom-fields/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/custom-fields/${id}`)
    await apiClient.delete(`/custom-fields/${id}`)
  },
}
