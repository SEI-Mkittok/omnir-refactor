import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { accountsApi } from '@/api/accounts'
import type {
  AccountListParams,
  CreateAccountRequest,
  UpdateAccountRequest,
  CreateNoteRequest,
} from '@/api/types'

export const accountKeys = {
  all: ['accounts'] as const,
  lists: () => [...accountKeys.all, 'list'] as const,
  list: (params?: AccountListParams) => [...accountKeys.lists(), params] as const,
  details: () => [...accountKeys.all, 'detail'] as const,
  detail: (id: string) => [...accountKeys.details(), id] as const,
  contacts: (id: string) => [...accountKeys.detail(id), 'contacts'] as const,
  deals: (id: string) => [...accountKeys.detail(id), 'deals'] as const,
  notes: (id: string) => [...accountKeys.detail(id), 'notes'] as const,
}

export function useAccounts(params?: AccountListParams) {
  return useQuery({
    queryKey: accountKeys.list(params),
    queryFn: () => accountsApi.list(params),
    staleTime: 30_000,
  })
}

export function useAccount(id: string) {
  return useQuery({
    queryKey: accountKeys.detail(id),
    queryFn: () => accountsApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useAccountContacts(id: string) {
  return useQuery({
    queryKey: accountKeys.contacts(id),
    queryFn: () => accountsApi.getContacts(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useAccountDeals(id: string) {
  return useQuery({
    queryKey: accountKeys.deals(id),
    queryFn: () => accountsApi.getDeals(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useAccountNotes(id: string) {
  return useQuery({
    queryKey: accountKeys.notes(id),
    queryFn: () => accountsApi.getNotes(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateAccountRequest) => accountsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: accountKeys.lists() })
    },
  })
}

export function useUpdateAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateAccountRequest }) =>
      accountsApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: accountKeys.lists() })
      qc.invalidateQueries({ queryKey: accountKeys.detail(id) })
    },
  })
}

export function useDeleteAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => accountsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: accountKeys.lists() })
    },
  })
}

export function useAddAccountNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      accountId,
      payload,
    }: {
      accountId: string
      payload: Omit<CreateNoteRequest, 'account_id'>
    }) => accountsApi.addNote(accountId, payload),
    onSuccess: (_, { accountId }) => {
      qc.invalidateQueries({ queryKey: accountKeys.notes(accountId) })
    },
  })
}
