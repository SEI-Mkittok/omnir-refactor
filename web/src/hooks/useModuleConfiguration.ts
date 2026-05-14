import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { moduleConfigurationApi } from '@/api/moduleConfiguration'
import type { CustomFieldEntityType, ModuleLayout, ModuleRelationshipDefinition } from '@/api/types'

export const moduleConfigurationKeys = {
  all: ['module-configuration'] as const,
  layouts: () => [...moduleConfigurationKeys.all, 'layouts'] as const,
  layout: (entityType: CustomFieldEntityType, admin = false) =>
    [...moduleConfigurationKeys.layouts(), entityType, admin ? 'admin' : 'runtime'] as const,
  relationships: () => [...moduleConfigurationKeys.all, 'relationships'] as const,
  relationshipList: (entityType?: CustomFieldEntityType, runtime = false) =>
    [...moduleConfigurationKeys.relationships(), entityType ?? null, runtime ? 'runtime' : 'admin'] as const,
}

export function useModuleLayout(entityType: CustomFieldEntityType) {
  return useQuery({
    queryKey: moduleConfigurationKeys.layout(entityType),
    queryFn: () => moduleConfigurationApi.getLayout(entityType),
    staleTime: 60_000,
  })
}

export function useAdminModuleLayout(entityType: CustomFieldEntityType) {
  return useQuery({
    queryKey: moduleConfigurationKeys.layout(entityType, true),
    queryFn: () => moduleConfigurationApi.getAdminLayout(entityType),
    staleTime: 30_000,
  })
}

export function useSaveModuleLayout(entityType: CustomFieldEntityType) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (layout: Pick<ModuleLayout, 'blocks'>) => moduleConfigurationApi.saveLayout(entityType, layout),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: moduleConfigurationKeys.layouts() })
    },
  })
}

export function useResetModuleLayout(entityType: CustomFieldEntityType) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => moduleConfigurationApi.resetLayout(entityType),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: moduleConfigurationKeys.layouts() })
    },
  })
}

export function useModuleRelationships(entityType?: CustomFieldEntityType) {
  return useQuery({
    queryKey: moduleConfigurationKeys.relationshipList(entityType),
    queryFn: () => moduleConfigurationApi.listRelationships(entityType),
    staleTime: 30_000,
  })
}

export function useRuntimeModuleRelationships(entityType: CustomFieldEntityType) {
  return useQuery({
    queryKey: moduleConfigurationKeys.relationshipList(entityType, true),
    queryFn: () => moduleConfigurationApi.listRuntimeRelationships(entityType),
    staleTime: 60_000,
  })
}

export function useSaveModuleRelationship() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (definition: ModuleRelationshipDefinition) => moduleConfigurationApi.saveRelationship(definition),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: moduleConfigurationKeys.relationships() })
    },
  })
}

export function useDeleteModuleRelationship() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => moduleConfigurationApi.deleteRelationship(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: moduleConfigurationKeys.relationships() })
    },
  })
}
