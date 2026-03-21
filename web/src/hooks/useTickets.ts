import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ticketsApi } from '@/api/tickets'
import type {
  TicketListParams,
  CreateTicketRequest,
  UpdateTicketRequest,
  CreateTicketCommentRequest,
} from '@/api/types'

export const ticketKeys = {
  all: ['tickets'] as const,
  lists: () => [...ticketKeys.all, 'list'] as const,
  list: (params?: TicketListParams) => [...ticketKeys.lists(), params] as const,
  details: () => [...ticketKeys.all, 'detail'] as const,
  detail: (id: string) => [...ticketKeys.details(), id] as const,
  comments: (id: string) => [...ticketKeys.detail(id), 'comments'] as const,
  attachments: (id: string) => [...ticketKeys.detail(id), 'attachments'] as const,
}

export function useTickets(params?: TicketListParams) {
  return useQuery({
    queryKey: ticketKeys.list(params),
    queryFn: () => ticketsApi.list(params),
    staleTime: 30_000,
  })
}

export function useTicket(id: string) {
  return useQuery({
    queryKey: ticketKeys.detail(id),
    queryFn: () => ticketsApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useTicketComments(id: string) {
  return useQuery({
    queryKey: ticketKeys.comments(id),
    queryFn: () => ticketsApi.getComments(id),
    staleTime: 15_000,
    enabled: !!id,
  })
}

export function useTicketAttachments(id: string) {
  return useQuery({
    queryKey: ticketKeys.attachments(id),
    queryFn: () => ticketsApi.getAttachments(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateTicketRequest) => ticketsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ticketKeys.lists() })
    },
  })
}

export function useUpdateTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateTicketRequest }) =>
      ticketsApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ticketKeys.lists() })
      qc.invalidateQueries({ queryKey: ticketKeys.detail(id) })
    },
  })
}

export function useDeleteTicket() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => ticketsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ticketKeys.lists() })
    },
  })
}

export function useAddTicketComment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      ticketId,
      payload,
    }: {
      ticketId: string
      payload: CreateTicketCommentRequest
    }) => ticketsApi.addComment(ticketId, payload),
    onSuccess: (_, { ticketId }) => {
      qc.invalidateQueries({ queryKey: ticketKeys.comments(ticketId) })
    },
  })
}

export function useUpdateTicketContact() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ ticketId, contactId }: { ticketId: string; contactId: string | null }) =>
      ticketsApi.patchContact(ticketId, contactId),
    onSuccess: (_, { ticketId }) => {
      qc.invalidateQueries({ queryKey: ticketKeys.detail(ticketId) })
      qc.invalidateQueries({ queryKey: ticketKeys.lists() })
    },
  })
}

export function useUploadTicketAttachment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ ticketId, file }: { ticketId: string; file: File }) =>
      ticketsApi.uploadAttachment(ticketId, file),
    onSuccess: (_, { ticketId }) => {
      qc.invalidateQueries({ queryKey: ticketKeys.attachments(ticketId) })
    },
  })
}
