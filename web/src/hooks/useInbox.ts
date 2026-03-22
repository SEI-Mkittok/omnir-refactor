import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { inboxApi } from '@/api/inbox'
import type { InboxListParams, SendInboxEmailRequest } from '@/api/types'

export const inboxKeys = {
  accounts: () => ['inbox', 'accounts'] as const,
  threads: (params?: InboxListParams) => ['inbox', 'threads', params] as const,
  thread: (id: string) => ['inbox', 'thread', id] as const,
  templates: (q?: string) => ['inbox', 'templates', q] as const,
}

export function useEmailAccounts() {
  return useQuery({
    queryKey: inboxKeys.accounts(),
    queryFn: () => inboxApi.listAccounts(),
  })
}

export function useInboxThreads(params?: InboxListParams) {
  return useQuery({
    queryKey: inboxKeys.threads(params),
    queryFn: () => inboxApi.listThreads(params),
  })
}

export function useInboxThread(threadId: string | null) {
  return useQuery({
    queryKey: inboxKeys.thread(threadId ?? ''),
    queryFn: () => inboxApi.getThread(threadId!),
    enabled: !!threadId,
  })
}

export function useEmailTemplates(q?: string) {
  return useQuery({
    queryKey: inboxKeys.templates(q),
    queryFn: () => inboxApi.listTemplates(q),
  })
}

export function useMarkThreadRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (threadId: string) => inboxApi.markRead(threadId),
    onSuccess: (_data, threadId) => {
      qc.invalidateQueries({ queryKey: inboxKeys.threads() })
      qc.invalidateQueries({ queryKey: inboxKeys.thread(threadId) })
    },
  })
}

export function useSendEmail() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: SendInboxEmailRequest) => inboxApi.sendEmail(payload),
    onSuccess: (_data, payload) => {
      qc.invalidateQueries({ queryKey: inboxKeys.threads() })
      if (payload.thread_id) {
        qc.invalidateQueries({ queryKey: inboxKeys.thread(payload.thread_id) })
      }
    },
  })
}

export function useConnectEmailAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: inboxApi.connectAccount,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: inboxKeys.accounts() })
    },
  })
}

export function useDisconnectEmailAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (accountId: string) => inboxApi.disconnectAccount(accountId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: inboxKeys.accounts() })
    },
  })
}
