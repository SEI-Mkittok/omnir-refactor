import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { webhooksApi } from '@/api/webhooks'
import type { CreateWebhookRequest, UpdateWebhookRequest } from '@/api/types'

export const webhookKeys = {
  all: ['webhooks'] as const,
  lists: () => [...webhookKeys.all, 'list'] as const,
  detail: (id: string) => [...webhookKeys.all, 'detail', id] as const,
  deliveries: (id: string) => [...webhookKeys.all, 'deliveries', id] as const,
}

export function useWebhooks() {
  return useQuery({
    queryKey: webhookKeys.lists(),
    queryFn: () => webhooksApi.list(),
    staleTime: 30_000,
  })
}

export function useCreateWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateWebhookRequest) => webhooksApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: webhookKeys.lists() })
    },
  })
}

export function useUpdateWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateWebhookRequest }) =>
      webhooksApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: webhookKeys.lists() })
      qc.invalidateQueries({ queryKey: webhookKeys.detail(id) })
    },
  })
}

export function useDeleteWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => webhooksApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: webhookKeys.lists() })
    },
  })
}

export function useTestWebhook() {
  return useMutation({
    mutationFn: (id: string) => webhooksApi.test(id),
  })
}

export function useWebhookDeliveries(webhookId: string) {
  return useQuery({
    queryKey: webhookKeys.deliveries(webhookId),
    queryFn: () => webhooksApi.listDeliveries(webhookId),
    staleTime: 15_000,
  })
}
