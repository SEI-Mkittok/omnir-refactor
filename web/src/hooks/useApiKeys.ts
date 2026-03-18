import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiKeysApi } from '@/api/apiKeys'
import type { CreateAPIKeyRequest } from '@/api/types'

export const apiKeyKeys = {
  all: ['api-keys'] as const,
  lists: () => [...apiKeyKeys.all, 'list'] as const,
}

export function useApiKeys() {
  return useQuery({
    queryKey: apiKeyKeys.lists(),
    queryFn: () => apiKeysApi.list(),
    staleTime: 30_000,
  })
}

export function useCreateApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateAPIKeyRequest) => apiKeysApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: apiKeyKeys.lists() })
    },
  })
}

export function useRevokeApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.revoke(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: apiKeyKeys.lists() })
    },
  })
}
