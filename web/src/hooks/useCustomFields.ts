import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { customFieldsApi } from '@/api/customFields'
import type {
  CustomFieldEntityType,
  CreateCustomFieldDefinitionRequest,
  UpdateCustomFieldDefinitionRequest,
} from '@/api/types'

export const customFieldKeys = {
  all: ['custom-field-definitions'] as const,
  lists: () => [...customFieldKeys.all, 'list'] as const,
  list: (entity_type?: CustomFieldEntityType) => [...customFieldKeys.lists(), entity_type] as const,
  detail: (id: string) => [...customFieldKeys.all, 'detail', id] as const,
}

export function useCustomFieldDefinitions(entity_type?: CustomFieldEntityType) {
  return useQuery({
    queryKey: customFieldKeys.list(entity_type),
    queryFn: () => customFieldsApi.list(entity_type),
    staleTime: 60_000,
  })
}

export function useCreateCustomFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateCustomFieldDefinitionRequest) => customFieldsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: customFieldKeys.lists() })
    },
  })
}

export function useUpdateCustomFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateCustomFieldDefinitionRequest }) =>
      customFieldsApi.update(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: customFieldKeys.lists() })
    },
  })
}

export function useDeleteCustomFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => customFieldsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: customFieldKeys.lists() })
    },
  })
}
