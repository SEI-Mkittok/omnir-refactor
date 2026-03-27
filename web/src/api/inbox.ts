// TODO(OMN-529): Some endpoints below are still scaffolded against the expected
// API contract. Remove remaining mock data when the backend stubs land.

import apiClient from './client'
import type {
  EmailAccount,
  InboxThread,
  InboxMessage,
  EmailTemplate,
  PaginatedResponse,
  InboxListParams,
  SendInboxEmailRequest,
  ConnectEmailAccountRequest,
} from './types'

// ── Mock data (remove when OMN-528 merges) ──────────────────────────────────

const MOCK_ACCOUNTS: EmailAccount[] = [
  {
    id: 'acc-1',
    org_id: 'org-1',
    user_id: 'user-1',
    provider: 'gmail',
    email_address: 'alex@company.com',
    display_name: 'Alex (Work)',
    status: 'connected',
    last_synced_at: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
    created_at: new Date().toISOString(),
  },
]

const MOCK_THREADS: InboxThread[] = [
  {
    id: 'thread-1',
    org_id: 'org-1',
    account_id: 'acc-1',
    subject: 'Re: Proposal follow-up',
    participants: ['sarah@acme.com', 'alex@company.com'],
    snippet: 'Thanks for sending over the proposal. I had a chance to review it with our team…',
    unread: true,
    message_count: 4,
    last_message_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
    contact_id: 'contact-1',
  },
  {
    id: 'thread-2',
    org_id: 'org-1',
    account_id: 'acc-1',
    subject: 'Q3 Contract Review',
    participants: ['marcus@webb.co', 'alex@company.com'],
    snippet: 'Please find attached the updated contract terms for Q3. Let me know if you have any questions.',
    unread: false,
    message_count: 2,
    last_message_at: new Date(Date.now() - 26 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 'thread-3',
    org_id: 'org-1',
    account_id: 'acc-1',
    subject: 'Introduction — TechCorp Partnership',
    participants: ['jennifer@techcorp.com', 'alex@company.com'],
    snippet: 'Hi Alex, I was referred to you by David at GrowthCo. We are looking for a CRM solution.',
    unread: true,
    message_count: 1,
    last_message_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    contact_id: 'contact-2',
  },
]

const MOCK_MESSAGES: InboxMessage[] = [
  {
    id: 'msg-1',
    org_id: 'org-1',
    account_id: 'acc-1',
    thread_id: 'thread-1',
    message_id: '<msg-1@gmail.com>',
    direction: 'outbound',
    from_addr: 'alex@company.com',
    from_name: 'Alex',
    to_addrs: ['sarah@acme.com'],
    subject: 'Proposal follow-up',
    body_text: 'Hi Sarah,\n\nI wanted to follow up on the proposal I sent last week. Please let me know if you have any questions.\n\nBest,\nAlex',
    snippet: 'I wanted to follow up on the proposal I sent last week.',
    has_attachments: false,
    sent_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 'msg-2',
    org_id: 'org-1',
    account_id: 'acc-1',
    thread_id: 'thread-1',
    message_id: '<msg-2@gmail.com>',
    direction: 'inbound',
    from_addr: 'sarah@acme.com',
    from_name: 'Sarah Chen',
    to_addrs: ['alex@company.com'],
    subject: 'Re: Proposal follow-up',
    body_text: 'Thanks for sending over the proposal. I had a chance to review it with our team and we have a few questions. Can we schedule a call this week?\n\nBest,\nSarah',
    snippet: 'Thanks for sending over the proposal. I had a chance to review it with our team…',
    has_attachments: false,
    contact_id: 'contact-1',
    sent_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
    created_at: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
  },
]

const MOCK_TEMPLATES: EmailTemplate[] = [
  {
    id: 'tmpl-1',
    org_id: 'org-1',
    name: 'Follow-up after demo',
    subject: 'Great speaking with you!',
    body_html: '<p>Hi {first_name},</p><p>Thanks for joining us for the demo today. I wanted to follow up and answer any questions you might have.</p><p>Best,<br>{sender_name}</p>',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'tmpl-2',
    org_id: 'org-1',
    name: 'Proposal sent confirmation',
    body_html: '<p>Attached please find our proposal for {company_name}. Please review at your convenience and let me know if you have any questions.</p>',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'tmpl-3',
    org_id: 'org-1',
    name: 'Meeting request',
    body_html: '<p>Hi {first_name},</p><p>I\'d love to schedule 30 minutes to discuss how we can help {company_name}. Are you available this week?</p>',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

// ── API client (real calls, uncomment when OMN-528 merges) ──────────────────

export const inboxApi = {
  // List connected email accounts
  listAccounts: async (): Promise<EmailAccount[]> => {
    // TODO(OMN-528): return (await apiClient.get<EmailAccount[]>('/inbox/accounts')).data
    return Promise.resolve(MOCK_ACCOUNTS)
  },

  // Connect a new email account — returns OAuth redirect URL
  connectAccount: async (payload: ConnectEmailAccountRequest): Promise<{ redirect_url: string }> => {
    // TODO(OMN-528): return (await apiClient.post<{ redirect_url: string }>('/inbox/accounts/connect', payload)).data
    return Promise.resolve({ redirect_url: `https://accounts.${payload.provider}.com/oauth` })
  },

  // Disconnect an account
  disconnectAccount: async (accountId: string): Promise<void> => {
    // TODO(OMN-528): await apiClient.delete(`/inbox/accounts/${accountId}`)
    void accountId
  },

  // List threads (unified inbox)
  listThreads: async (params?: InboxListParams): Promise<PaginatedResponse<InboxThread>> => {
    return (await apiClient.get<PaginatedResponse<InboxThread>>('/emails/threads', { params })).data
  },

  // Get a single thread with messages
  getThread: async (threadId: string): Promise<InboxThread & { messages: InboxMessage[] }> => {
    // TODO(OMN-528): return (await apiClient.get<InboxThread & { messages: InboxMessage[] }>(`/inbox/threads/${threadId}`)).data
    const thread = MOCK_THREADS.find((t) => t.id === threadId) ?? MOCK_THREADS[0]
    return Promise.resolve({ ...thread, messages: MOCK_MESSAGES.filter((m) => m.thread_id === threadId) })
  },

  // Mark thread as read
  markRead: async (threadId: string): Promise<void> => {
    await apiClient.patch(`/emails/${threadId}/read`)
  },

  // Send / reply
  sendEmail: async (payload: SendInboxEmailRequest): Promise<InboxMessage> => {
    // TODO(OMN-528): return (await apiClient.post<InboxMessage>('/inbox/send', payload)).data
    const msg: InboxMessage = {
      id: `msg-${Date.now()}`,
      org_id: 'org-1',
      account_id: payload.account_id,
      thread_id: payload.thread_id ?? `thread-${Date.now()}`,
      message_id: `<${Date.now()}@mock.local>`,
      direction: 'outbound',
      from_addr: MOCK_ACCOUNTS[0]?.email_address ?? '',
      to_addrs: payload.to,
      cc_addrs: payload.cc,
      subject: payload.subject,
      body_text: payload.body_html.replace(/<[^>]+>/g, ''),
      snippet: payload.body_html.replace(/<[^>]+>/g, '').slice(0, 100),
      has_attachments: false,
      sent_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
    }
    return Promise.resolve(msg)
  },

  // List email templates
  listTemplates: async (q?: string): Promise<EmailTemplate[]> => {
    // TODO(OMN-528): return (await apiClient.get<EmailTemplate[]>('/inbox/templates', { params: { q } })).data
    if (q) {
      return Promise.resolve(MOCK_TEMPLATES.filter((t) => t.name.toLowerCase().includes(q.toLowerCase())))
    }
    return Promise.resolve(MOCK_TEMPLATES)
  },
}
