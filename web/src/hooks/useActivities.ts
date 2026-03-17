import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { activitiesApi } from '@/api/activities'
import type { ActivityListParams, CreateActivityRequest, UpdateActivityRequest } from '@/api/activities'

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (params?: ActivityListParams) => [...activityKeys.lists(), params] as const,
  details: () => [...activityKeys.all, 'detail'] as const,
  detail: (id: string) => [...activityKeys.details(), id] as const,
  forContact: (contactId: string) => [...activityKeys.lists(), { contact_id: contactId }] as const,
  forDeal: (dealId: string) => [...activityKeys.lists(), { deal_id: dealId }] as const,
}

export function useActivities(params?: ActivityListParams) {
  return useQuery({
    queryKey: activityKeys.list(params),
    queryFn: () => activitiesApi.list(params),
    staleTime: 30_000,
  })
}

export function useContactActivities(contactId: string) {
  return useQuery({
    queryKey: activityKeys.forContact(contactId),
    queryFn: () =>
      activitiesApi.list({
        contact_id: contactId,
        per_page: 100,
        sort_by: 'created_at',
        sort_dir: 'desc',
      }),
    staleTime: 30_000,
    enabled: !!contactId,
  })
}

export function useDealActivities(dealId: string) {
  return useQuery({
    queryKey: activityKeys.forDeal(dealId),
    queryFn: () =>
      activitiesApi.list({
        deal_id: dealId,
        per_page: 100,
        sort_by: 'created_at',
        sort_dir: 'desc',
      }),
    staleTime: 30_000,
    enabled: !!dealId,
  })
}

export function useCreateActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateActivityRequest) => activitiesApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function useUpdateActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateActivityRequest }) =>
      activitiesApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: activityKeys.lists() })
      qc.invalidateQueries({ queryKey: activityKeys.detail(id) })
    },
  })
}

export function useDeleteActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => activitiesApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}
