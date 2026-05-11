import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessSettingsApi } from '@/api/accessSettings'
import type {
  ACLGroup,
  ACLProfile,
  ACLProfileFieldPermission,
  ACLProfilePermission,
  ACLRole,
  ACLSharingRules,
} from '@/api/types'

export const accessSettingsKeys = {
  all: ['access-settings'] as const,
  roles: () => [...accessSettingsKeys.all, 'roles'] as const,
  profiles: () => [...accessSettingsKeys.all, 'profiles'] as const,
  profilePermissions: (id: string) => [...accessSettingsKeys.profiles(), id, 'permissions'] as const,
  catalog: () => [...accessSettingsKeys.profiles(), 'catalog'] as const,
  groups: () => [...accessSettingsKeys.all, 'groups'] as const,
  sharing: () => [...accessSettingsKeys.all, 'sharing'] as const,
}

export function useACLRoles() {
  return useQuery({
    queryKey: accessSettingsKeys.roles(),
    queryFn: accessSettingsApi.listRoles,
    staleTime: 60_000,
  })
}

export function useACLProfiles() {
  return useQuery({
    queryKey: accessSettingsKeys.profiles(),
    queryFn: accessSettingsApi.listProfiles,
    staleTime: 60_000,
  })
}

export function usePermissionCatalog() {
  return useQuery({
    queryKey: accessSettingsKeys.catalog(),
    queryFn: accessSettingsApi.getPermissionCatalog,
    staleTime: 5 * 60_000,
  })
}

export function useProfilePermissions(profileId?: string) {
  return useQuery({
    queryKey: accessSettingsKeys.profilePermissions(profileId ?? ''),
    queryFn: () => accessSettingsApi.getProfilePermissions(profileId!),
    enabled: Boolean(profileId),
  })
}

export function useACLGroups() {
  return useQuery({
    queryKey: accessSettingsKeys.groups(),
    queryFn: accessSettingsApi.listGroups,
    staleTime: 60_000,
  })
}

export function useSharingRules() {
  return useQuery({
    queryKey: accessSettingsKeys.sharing(),
    queryFn: accessSettingsApi.getSharingRules,
    staleTime: 30_000,
  })
}

export function useCreateACLRole() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Pick<ACLRole, 'name' | 'description' | 'parent_id'>) =>
      accessSettingsApi.createRole(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.roles() }),
  })
}

export function useMoveACLRole() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, parent_id }: { id: string; parent_id: string | null }) =>
      accessSettingsApi.moveRole(id, parent_id),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.roles() }),
  })
}

export function useDeleteACLRole() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: accessSettingsApi.deleteRole,
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.roles() }),
  })
}

export function useCreateACLProfile() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Pick<ACLProfile, 'name' | 'description'>) =>
      accessSettingsApi.createProfile(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.profiles() }),
  })
}

export function useSaveProfilePermissions() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      profileId,
      permissions,
      field_permissions,
    }: {
      profileId: string
      permissions: ACLProfilePermission[]
      field_permissions: ACLProfileFieldPermission[]
    }) => accessSettingsApi.replaceProfilePermissions(profileId, { permissions, field_permissions }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: accessSettingsKeys.profilePermissions(vars.profileId) })
    },
  })
}

export function useCreateACLGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Pick<ACLGroup, 'name' | 'description' | 'user_ids'>) =>
      accessSettingsApi.createGroup(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.groups() }),
  })
}

export function useSaveGroupMembers() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, user_ids }: { id: string; user_ids: string[] }) =>
      accessSettingsApi.replaceGroupMembers(id, user_ids),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.groups() }),
  })
}

export function useDeleteACLGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: accessSettingsApi.deleteGroup,
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.groups() }),
  })
}

export function useSaveSharingRules() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: ACLSharingRules) => accessSettingsApi.replaceSharingRules(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: accessSettingsKeys.sharing() }),
  })
}
