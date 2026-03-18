import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { emailsApi } from '@/api/emails'
import type { SendEmailRequest } from '@/api/types'

export const emailKeys = {
  all: ['emails'] as const,
  byContact: (contactId: string) => [...emailKeys.all, 'contact', contactId] as const,
}

export function useContactEmails(contactId: string) {
  return useQuery({
    queryKey: emailKeys.byContact(contactId),
    queryFn: () => emailsApi.listByContact(contactId),
    staleTime: 30_000,
    enabled: !!contactId,
  })
}

export function useSendEmail() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: SendEmailRequest) => emailsApi.send(payload),
    onSuccess: (_, payload) => {
      if (payload.contact_id) {
        qc.invalidateQueries({ queryKey: emailKeys.byContact(payload.contact_id) })
      }
    },
  })
}
