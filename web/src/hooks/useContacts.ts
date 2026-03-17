import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { contactsApi } from '@/api/contacts'
import type {
  ContactListParams,
  CreateContactRequest,
  UpdateContactRequest,
  CreateNoteRequest,
} from '@/api/types'

export const contactKeys = {
  all: ['contacts'] as const,
  lists: () => [...contactKeys.all, 'list'] as const,
  list: (params?: ContactListParams) => [...contactKeys.lists(), params] as const,
  details: () => [...contactKeys.all, 'detail'] as const,
  detail: (id: string) => [...contactKeys.details(), id] as const,
  notes: (id: string) => [...contactKeys.detail(id), 'notes'] as const,
}

export function useContacts(params?: ContactListParams) {
  return useQuery({
    queryKey: contactKeys.list(params),
    queryFn: () => contactsApi.list(params),
    staleTime: 30_000,
  })
}

export function useContact(id: string) {
  return useQuery({
    queryKey: contactKeys.detail(id),
    queryFn: () => contactsApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useContactNotes(id: string) {
  return useQuery({
    queryKey: contactKeys.notes(id),
    queryFn: () => contactsApi.getNotes(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateContact() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateContactRequest) => contactsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contactKeys.lists() })
    },
  })
}

export function useUpdateContact() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateContactRequest }) =>
      contactsApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: contactKeys.lists() })
      qc.invalidateQueries({ queryKey: contactKeys.detail(id) })
    },
  })
}

export function useDeleteContact() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => contactsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contactKeys.lists() })
    },
  })
}

export function useAddContactNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contactId,
      payload,
    }: {
      contactId: string
      payload: Omit<CreateNoteRequest, 'contact_id'>
    }) => contactsApi.addNote(contactId, payload),
    onSuccess: (_, { contactId }) => {
      qc.invalidateQueries({ queryKey: contactKeys.notes(contactId) })
    },
  })
}
