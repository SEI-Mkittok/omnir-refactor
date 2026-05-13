import apiClient from './client'
import type {
  CRMEntityLink,
  CustomFieldEntityType,
  ModuleLayout,
  ModuleRelationshipDefinition,
} from './types'

export const moduleConfigurationApi = {
  getLayout: async (entityType: CustomFieldEntityType): Promise<ModuleLayout> => {
    const { data } = await apiClient.get(`/module-layouts/${entityType}`)
    return data
  },

  getAdminLayout: async (entityType: CustomFieldEntityType): Promise<ModuleLayout> => {
    const { data } = await apiClient.get(`/settings/module-layouts/${entityType}`)
    return data
  },

  saveLayout: async (entityType: CustomFieldEntityType, layout: Pick<ModuleLayout, 'blocks'>): Promise<ModuleLayout> => {
    const { data } = await apiClient.put(`/settings/module-layouts/${entityType}`, layout)
    return data
  },

  resetLayout: async (entityType: CustomFieldEntityType): Promise<ModuleLayout> => {
    const { data } = await apiClient.post(`/settings/module-layouts/${entityType}/reset`)
    return data
  },

  listRelationships: async (entityType?: CustomFieldEntityType): Promise<ModuleRelationshipDefinition[]> => {
    const { data } = await apiClient.get('/settings/module-relationships', {
      params: entityType ? { entity_type: entityType } : undefined,
    })
    return data
  },

  listRuntimeRelationships: async (entityType: CustomFieldEntityType): Promise<ModuleRelationshipDefinition[]> => {
    const { data } = await apiClient.get('/module-relationships', { params: { entity_type: entityType } })
    return data
  },

  saveRelationship: async (definition: ModuleRelationshipDefinition): Promise<ModuleRelationshipDefinition> => {
    const { data } = await apiClient.put('/settings/module-relationships', definition)
    return data
  },

  deleteRelationship: async (id: string): Promise<void> => {
    await apiClient.delete(`/settings/module-relationships/${id}`)
  },

  listEntityLinks: async (
    entityType: CustomFieldEntityType,
    entityId: string,
    relationshipDefinitionId?: string,
  ): Promise<CRMEntityLink[]> => {
    const { data } = await apiClient.get(`/entity-links/${entityType}/${entityId}`, {
      params: relationshipDefinitionId ? { relationship_definition_id: relationshipDefinitionId } : undefined,
    })
    return data
  },

  createEntityLink: async (payload: {
    relationship_definition_id: string
    from_entity_type: CustomFieldEntityType
    from_entity_id: string
    to_entity_type: CustomFieldEntityType
    to_entity_id: string
    metadata?: Record<string, unknown>
  }): Promise<CRMEntityLink> => {
    const { data } = await apiClient.post('/entity-links', payload)
    return data
  },

  deleteEntityLink: async (id: string): Promise<void> => {
    await apiClient.delete(`/entity-links/${id}`)
  },
}
