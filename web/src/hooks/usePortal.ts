import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { portalApi } from '@/api/portal'
import type { CreatePortalTicketRequest, PortalTicketListParams } from '@/api/types'

export const portalKeys = {
  all: ['portal'] as const,
  tickets: () => [...portalKeys.all, 'tickets'] as const,
  ticketList: (params?: PortalTicketListParams) => [...portalKeys.tickets(), 'list', params] as const,
  ticket: (id: string) => [...portalKeys.tickets(), 'detail', id] as const,
  comments: (ticketId: string) => [...portalKeys.ticket(ticketId), 'comments'] as const,
}

export function usePortalTickets(params?: PortalTicketListParams) {
  return useQuery({
    queryKey: portalKeys.ticketList(params),
    queryFn: () => portalApi.listTickets(params),
    staleTime: 30_000,
  })
}

export function usePortalTicket(id: string) {
  return useQuery({
    queryKey: portalKeys.ticket(id),
    queryFn: () => portalApi.getTicket(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function usePortalTicketComments(ticketId: string) {
  return useQuery({
    queryKey: portalKeys.comments(ticketId),
    queryFn: () => portalApi.getComments(ticketId),
    staleTime: 15_000,
    enabled: !!ticketId,
  })
}

export function useCreatePortalTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreatePortalTicketRequest) => portalApi.createTicket(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: portalKeys.tickets() })
    },
  })
}

export function useAddPortalComment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ ticketId, body }: { ticketId: string; body: string }) =>
      portalApi.addComment(ticketId, body),
    onSuccess: (_, { ticketId }) => {
      qc.invalidateQueries({ queryKey: portalKeys.comments(ticketId) })
    },
  })
}
