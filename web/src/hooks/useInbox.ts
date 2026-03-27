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
    onMutate: async (threadId: string) => {
      await qc.cancelQueries({ queryKey: inboxKeys.threads() })
      const previous = qc.getQueriesData({ queryKey: inboxKeys.threads() })
      qc.setQueriesData({ queryKey: inboxKeys.threads() }, (old: unknown) => {
        if (!old || typeof old !== 'object') return old
        const page = old as { data: { id: string; unread: boolean }[]; meta: unknown }
        return {
          ...page,
          data: page.data.map((t) => (t.id === threadId ? { ...t, unread: false } : t)),
        }
      })
      return { previous }
    },
    onError: (_err, _threadId, context) => {
      if (context?.previous) {
        for (const [queryKey, data] of context.previous) {
          qc.setQueryData(queryKey, data)
        }
      }
    },
    onSettled: (_data, _err, threadId) => {
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
