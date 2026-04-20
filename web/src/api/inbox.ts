import { apiClient, mgmtClient } from './client'
import type {
  ConnectEmailAccountRequest,
  EmailAccount,
  EmailTemplate,
  InboxListParams,
  InboxMessage,
  InboxThread,
  PaginatedResponse,
  SendInboxEmailRequest,
} from './types'

interface InboxThreadDetailResponse {
  thread: BackendInboxThread
  messages: BackendInboxMessage[]
}

interface BackendInboxMessage {
  id: string
  org_id: string
  connection_id: string
  thread_id: string
  message_id: string
  direction: 'inbound' | 'outbound'
  from_addr: string
  to_addrs?: string[]
  subject: string
  body_html?: string
  body_text?: string
  contact_id?: string
  sent_at: string
  created_at: string
}

interface BackendInboxThread {
  thread_id: string
  org_id: string
  connection_id: string
  subject: string
  participants?: string[]
  snippet?: string
  unread: boolean
  message_count: number
  last_message_at: string
  contact_id?: string
}

interface EmailTemplateListResponse {
  data?: EmailTemplate[]
}

function normalizeMessage(message: BackendInboxMessage): InboxMessage {
  const bodyText = message.body_text ?? ''

  return {
    id: message.id,
    org_id: message.org_id,
    connection_id: message.connection_id,
    thread_id: message.thread_id,
    message_id: message.message_id,
    direction: message.direction,
    from_addr: message.from_addr,
    to_addrs: message.to_addrs ?? [],
    subject: message.subject,
    body_html: message.body_html,
    body_text: bodyText,
    snippet: bodyText.slice(0, 160),
    has_attachments: false,
    contact_id: message.contact_id,
    sent_at: message.sent_at,
    created_at: message.created_at,
  }
}

function normalizeThread(thread: BackendInboxThread): InboxThread {
  return {
    thread_id: thread.thread_id,
    org_id: thread.org_id,
    connection_id: thread.connection_id,
    subject: thread.subject,
    participants: thread.participants ?? [],
    snippet: thread.snippet ?? '',
    unread: thread.unread,
    message_count: thread.message_count,
    last_message_at: thread.last_message_at,
    contact_id: thread.contact_id,
  }
}

export const inboxApi = {
  listAccounts: async (): Promise<EmailAccount[]> => {
    const { data } = await mgmtClient.get<EmailAccount[]>('/integrations/email/connections')
    return data
  },

  connectAccount: async (payload: ConnectEmailAccountRequest): Promise<{ redirect_url: string }> => {
    void payload.redirect_uri
    const providerRoute = payload.provider === 'outlook' ? 'microsoft' : payload.provider
    return { redirect_url: `/api/integrations/email/auth/${providerRoute}` }
  },

  disconnectAccount: async (accountId: string): Promise<void> => {
    await mgmtClient.delete(`/integrations/email/connections/${accountId}`)
  },

  listThreads: async (params?: InboxListParams): Promise<PaginatedResponse<InboxThread>> => {
    const query = params
      ? {
          connection_id: params.connection_id,
          unread_only: params.unread_only,
          page: params.page,
          limit: params.limit,
        }
      : undefined
    const { data } = await apiClient.get<PaginatedResponse<BackendInboxThread>>('/emails/threads', { params: query })
    return {
      ...data,
      data: (data.data ?? []).map(normalizeThread),
    }
  },

  getThread: async (threadId: string): Promise<InboxThread & { messages: InboxMessage[] }> => {
    const { data } = await apiClient.get<InboxThreadDetailResponse>(`/emails/threads/${threadId}`)
    return {
      ...normalizeThread(data.thread),
      messages: (data.messages ?? []).map(normalizeMessage),
    }
  },

  markRead: async (threadId: string): Promise<void> => {
    await apiClient.patch(`/emails/${threadId}/read`)
  },

  sendEmail: async (payload: SendInboxEmailRequest): Promise<InboxMessage> => {
    const { data } = await apiClient.post<BackendInboxMessage>('/emails/send', payload)
    return normalizeMessage(data)
  },

  listTemplates: async (q?: string): Promise<EmailTemplate[]> => {
    const { data } = await apiClient.get<EmailTemplateListResponse>('/email-templates', { params: q ? { q } : undefined })
    return data.data ?? []
  },
}
