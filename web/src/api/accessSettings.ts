import apiClient from './client'
import type {
  ACLGroup,
  ACLProfile,
  ACLProfileFieldPermission,
  ACLProfilePermission,
  ACLRole,
  ACLSharingRules,
  PermissionCatalog,
} from './types'

interface ListResponse<T> {
  data: T[]
}

export const accessSettingsApi = {
  listRoles: async (): Promise<ACLRole[]> => {
    const { data } = await apiClient.get<ListResponse<ACLRole>>('/settings/roles')
    return data.data
  },

  createRole: async (payload: Pick<ACLRole, 'name' | 'description' | 'parent_id'>): Promise<ACLRole> => {
    const { data } = await apiClient.post<ACLRole>('/settings/roles', payload)
    return data
  },

  updateRole: async (id: string, payload: Partial<Pick<ACLRole, 'name' | 'description' | 'parent_id'>>): Promise<ACLRole> => {
    const { data } = await apiClient.patch<ACLRole>(`/settings/roles/${id}`, payload)
    return data
  },

  moveRole: async (id: string, parent_id: string | null): Promise<ACLRole> => {
    const { data } = await apiClient.patch<ACLRole>(`/settings/roles/${id}/parent`, { parent_id })
    return data
  },

  deleteRole: async (id: string): Promise<void> => {
    await apiClient.delete(`/settings/roles/${id}`)
  },

  listProfiles: async (): Promise<ACLProfile[]> => {
    const { data } = await apiClient.get<ListResponse<ACLProfile>>('/settings/profiles')
    return data.data
  },

  createProfile: async (payload: Pick<ACLProfile, 'name' | 'description'>): Promise<ACLProfile> => {
    const { data } = await apiClient.post<ACLProfile>('/settings/profiles', payload)
    return data
  },

  updateProfile: async (id: string, payload: Partial<Pick<ACLProfile, 'name' | 'description'>>): Promise<ACLProfile> => {
    const { data } = await apiClient.patch<ACLProfile>(`/settings/profiles/${id}`, payload)
    return data
  },

  deleteProfile: async (id: string): Promise<void> => {
    await apiClient.delete(`/settings/profiles/${id}`)
  },

  getPermissionCatalog: async (): Promise<PermissionCatalog> => {
    const { data } = await apiClient.get<PermissionCatalog>('/settings/profiles/catalog')
    return data
  },

  getProfilePermissions: async (id: string): Promise<{
    permissions: ACLProfilePermission[]
    field_permissions: ACLProfileFieldPermission[]
  }> => {
    const { data } = await apiClient.get(`/settings/profiles/${id}/permissions`)
    return data
  },

  replaceProfilePermissions: async (
    id: string,
    payload: { permissions: ACLProfilePermission[]; field_permissions: ACLProfileFieldPermission[] }
  ): Promise<{ permissions: ACLProfilePermission[]; field_permissions: ACLProfileFieldPermission[] }> => {
    const { data } = await apiClient.put(`/settings/profiles/${id}/permissions`, payload)
    return data
  },

  listGroups: async (): Promise<ACLGroup[]> => {
    const { data } = await apiClient.get<ListResponse<ACLGroup>>('/settings/groups')
    return data.data
  },

  createGroup: async (payload: Pick<ACLGroup, 'name' | 'description' | 'user_ids'>): Promise<ACLGroup> => {
    const { data } = await apiClient.post<ACLGroup>('/settings/groups', payload)
    return data
  },

  updateGroup: async (id: string, payload: Partial<Pick<ACLGroup, 'name' | 'description'>>): Promise<ACLGroup> => {
    const { data } = await apiClient.patch<ACLGroup>(`/settings/groups/${id}`, payload)
    return data
  },

  deleteGroup: async (id: string): Promise<void> => {
    await apiClient.delete(`/settings/groups/${id}`)
  },

  replaceGroupMembers: async (id: string, user_ids: string[]): Promise<void> => {
    await apiClient.put(`/settings/groups/${id}/members`, { user_ids })
  },

  getSharingRules: async (): Promise<ACLSharingRules> => {
    const { data } = await apiClient.get<ACLSharingRules>('/settings/sharing-rules')
    return data
  },

  replaceSharingRules: async (payload: ACLSharingRules): Promise<ACLSharingRules> => {
    const { data } = await apiClient.patch<ACLSharingRules>('/settings/sharing-rules', payload)
    return data
  },
}
