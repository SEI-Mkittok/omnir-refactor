import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { campaignsApi } from '@/api/campaigns'
import type { Campaign, CampaignMemberInput } from '@/api/types'

const keys = {
  all: ['campaigns'] as const,
  list: (params?: unknown) => [...keys.all, 'list', params] as const,
  members: (id: string) => [...keys.all, 'members', id] as const,
}

export function useCampaigns(params?: { page?: number; limit?: number; status?: string; q?: string }) {
  return useQuery({
    queryKey: keys.list(params),
    queryFn: () => campaignsApi.list(params),
    staleTime: 30_000,
  })
}

export function useCreateCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<Campaign>) => campaignsApi.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useUpdateCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<Campaign> }) => campaignsApi.update(id, payload),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: keys.all })
      qc.invalidateQueries({ queryKey: keys.members(id) })
    },
  })
}

export function useDeleteCampaign() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => campaignsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useCampaignMembers(id: string) {
  return useQuery({
    queryKey: keys.members(id),
    queryFn: () => campaignsApi.members(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}

export function useAddCampaignMembers() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, members }: { id: string; members: CampaignMemberInput[] }) => campaignsApi.addMembers(id, members),
    onSuccess: (_data, { id }) => qc.invalidateQueries({ queryKey: keys.members(id) }),
  })
}
