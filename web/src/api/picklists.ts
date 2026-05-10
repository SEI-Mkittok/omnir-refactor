import apiClient from './client'
import type { CustomFieldDefinition, CustomFieldEntityType } from './types'

export interface PicklistValue {
  id: string
  org_id: string
  custom_field_id: string
  value: string
  display_label: string
  order_idx: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface PicklistFieldBundle {
  field: CustomFieldDefinition
  values: PicklistValue[]
}

export interface PicklistValueInput {
  value: string
  display_label: string
  order_idx: number
  is_active: boolean
}

export interface PicklistDependency {
  id: string
  org_id: string
  entity_type: CustomFieldEntityType
  source_field_id: string
  target_field_id: string
  mapping: Record<string, string[]>
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface PicklistDependencyInput {
  entity_type: CustomFieldEntityType
  source_field_id: string
  target_field_id: string
  mapping: Record<string, string[]>
  is_active: boolean
}

export const picklistsApi = {
  listFields: async (entity_type?: CustomFieldEntityType): Promise<PicklistFieldBundle[]> => {
    const { data } = await apiClient.get('/settings/picklists', {
      params: entity_type ? { entity_type } : undefined,
    })
    return data.data ?? []
  },

  listValues: async (fieldID: string): Promise<PicklistValue[]> => {
    const { data } = await apiClient.get(`/settings/picklists/${fieldID}/values`)
    return data.values ?? []
  },

  upsertValues: async (fieldID: string, values: PicklistValueInput[]): Promise<PicklistValue[]> => {
    const { data } = await apiClient.put(`/settings/picklists/${fieldID}/values`, { values })
    return data.values ?? []
  },

  remapDelete: async (fieldID: string, from_value: string, to_value?: string): Promise<PicklistValue[]> => {
    const payload: Record<string, unknown> = { from_value }
    if (typeof to_value === 'string') {
      payload.to_value = to_value
    }
    const { data } = await apiClient.post(`/settings/picklists/${fieldID}/remap-delete`, payload)
    return data.values ?? []
  },

  listDependencies: async (entity_type?: CustomFieldEntityType): Promise<PicklistDependency[]> => {
    const { data } = await apiClient.get('/settings/picklist-dependencies', {
      params: entity_type ? { entity_type } : undefined,
    })
    return data.data ?? []
  },

  upsertDependency: async (payload: PicklistDependencyInput): Promise<PicklistDependency> => {
    const { data } = await apiClient.put('/settings/picklist-dependencies', payload)
    return data
  },

  deleteDependency: async (dependencyID: string): Promise<void> => {
    await apiClient.delete(`/settings/picklist-dependencies/${dependencyID}`)
  },
}
