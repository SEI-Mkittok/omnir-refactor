import { useEffect, useMemo, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  ChevronRight,
  Building2,
  Globe,
  Phone,
  MapPin,
  Users,
  Briefcase,
  Trash2,
  ExternalLink,
  Plus,
  X,
  Ticket,
  FileText,
  Mail,
  Zap,
  Search,
} from 'lucide-react'
import {
  useAccount,
  useAccountContacts,
  useAccountNotes,
  useDeleteAccount,
  useAddAccountNote,
  useLinkContactToAccount,
  useLinkDealToAccount,
  useLinkTicketToAccount,
} from '@/hooks/useAccounts'
import { useDeals } from '@/hooks/useDeals'
import { useTickets } from '@/hooks/useTickets'
import { useQuotes } from '@/hooks/useQuotes'
import { contactsApi } from '@/api/contacts'
import { dealsApi } from '@/api/deals'
import { inboxApi } from '@/api/inbox'
import { sequencesApi } from '@/api/sequences'
import { ticketsApi } from '@/api/tickets'

import { Button } from '@/components/ui/Button'
import {
  RelationshipEditor,
  type RelationshipRow,
  createRelationshipRow,
  normalizeRelationshipRows,
} from '@/components/omnir/RelationshipEditor'
import { formatDate, formatRelativeTime, formatCurrency } from '@/lib/utils'
import { mapCrmLinkError } from '@/lib/crmLinkErrors'
import { Spinner } from '@/components/ui/Spinner'
import { EntityLinkModal } from '@/components/omnir/EntityLinkModal'
import { UnifiedTimeline } from '@/components/omnir/UnifiedTimeline'
import type { Contact, Deal, EmailSequence, InboxThread, Note, Ticket as TicketType } from '@/api/types'

// ── Helpers ───────────────────────────────────────────────────────────────────

function accountInitials(name: string) {
  return name
    .split(' ')
    .map((w) => w[0] ?? '')
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function createAccountRelationshipRows(contacts: Contact[] | undefined): RelationshipRow[] {
  if (!contacts?.length) return []

  return normalizeRelationshipRows(
    contacts.map((contact, index) =>
      createRelationshipRow({
        id: `account-relationship-${contact.id}`,
        entityId: contact.id,
        label: `${contact.first_name} ${contact.last_name}`.trim(),
        meta: contact.title ?? contact.email ?? '',
        role: index === 0 ? 'primary' : 'billing',
        isPrimary: index === 0,
      })
    )
  )
}

// ── Notes Panel ───────────────────────────────────────────────────────────────
function NotesPanel({ accountId }: { accountId: string }) {
  const { data: notes, isLoading } = useAccountNotes(accountId)
  const addNote = useAddAccountNote()
  const [noteText, setNoteText] = useState('')

  const handleSubmit = async () => {
    const content = noteText.trim()
    if (!content) return
    await addNote.mutateAsync({ accountId, payload: { content } })
    setNoteText('')
  }

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <p
        className="mb-3 text-[11px] font-semibold uppercase tracking-widest"
        style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
      >
        Notes
      </p>

      {/* Add note */}
      <div className="mb-4">
        <textarea
          value={noteText}
          onChange={(e) => setNoteText(e.target.value)}
          placeholder="Add a note…"
          rows={3}
          className="w-full rounded-lg border text-sm resize-none px-3 py-2 focus:outline-none"
          style={{
            borderColor: 'var(--border-default)',
            color: 'var(--text-primary)',
            background: 'var(--surface-app)',
          }}
          onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--border-focus)' }}
          onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)' }}
        />
        <div className="flex justify-end mt-1.5">
          <Button
            size="sm"
            disabled={!noteText.trim() || addNote.isPending}
            onClick={handleSubmit}
          >
            <Plus className="h-3.5 w-3.5" />
            Add Note
          </Button>
        </div>
      </div>

      {/* Note list */}
      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="h-10 bg-slate-100 animate-pulse rounded" />
          ))}
        </div>
      ) : !notes?.length ? (
        <p className="text-sm" style={{ color: 'var(--text-label)' }}>
          No notes yet.
        </p>
      ) : (
        <ul className="space-y-3">
          {notes.map((n: Note) => (
            <li
              key={n.id}
              className="border-b pb-3 last:border-0 last:pb-0"
              style={{ borderColor: 'var(--border-subtle)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-primary)', whiteSpace: 'pre-wrap' }}>
                {n.content}
              </p>
              <time
                className="mt-1 block text-xs"
                style={{ color: 'var(--text-label)' }}
                title={formatDate(n.created_at)}
              >
                {formatRelativeTime(n.created_at)}
              </time>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

// ── Linked Entities ───────────────────────────────────────────────────────────

type LinkedEntitiesTab = 'contacts' | 'deals' | 'quotes' | 'tickets' | 'sequences' | 'inbox'

type LinkedEntityPageState = Record<LinkedEntitiesTab, number>

const LINKED_ENTITY_PAGE_SIZE = 5

const LINKED_TAB_ORDER: LinkedEntitiesTab[] = [
  'contacts',
  'deals',
  'quotes',
  'tickets',
  'sequences',
  'inbox',
]

const LINKED_TAB_META: Record<
  LinkedEntitiesTab,
  { label: string; route: string; icon: typeof Users; empty: string }
> = {
  contacts: { label: 'Contacts', route: '/contacts', icon: Users, empty: 'No contacts linked yet.' },
  deals: { label: 'Deals', route: '/deals', icon: Briefcase, empty: 'No deals linked yet.' },
  quotes: { label: 'Quotes', route: '/quotes', icon: FileText, empty: 'No quotes linked yet.' },
  tickets: { label: 'Tickets', route: '/tickets', icon: Ticket, empty: 'No tickets linked yet.' },
  sequences: { label: 'Sequences', route: '/sequences', icon: Zap, empty: "No sequences are linked through this account's contacts yet." },
  inbox: { label: 'Inbox', route: '/inbox', icon: Mail, empty: "No inbox threads are linked through this account's contacts yet." },
}

const STAGE_STYLES: Record<string, { bg: string; text: string }> = {
  lead: { bg: 'var(--color-primary-light)', text: 'var(--color-primary)' },
  qualified: { bg: 'var(--color-info-light)', text: 'var(--color-info)' },
  proposal: { bg: 'var(--color-warning-light)', text: 'var(--color-warning)' },
  negotiation: { bg: 'var(--surface-app)', text: 'var(--text-secondary)' },
  closed_won: { bg: 'var(--color-success-light)', text: 'var(--color-success)' },
  closed_lost: { bg: 'var(--color-danger-light)', text: 'var(--color-danger)' },
}

const TICKET_STATUS_STYLES: Record<string, { bg: string; text: string }> = {
  open: { bg: 'var(--color-warning-light)', text: 'var(--color-warning)' },
  in_progress: { bg: 'var(--color-info-light)', text: 'var(--color-info)' },
  resolved: { bg: 'var(--color-success-light)', text: 'var(--color-success)' },
  closed: { bg: 'var(--surface-app)', text: 'var(--text-secondary)' },
}

function buildAccountScopedPath(path: string, accountId: string, accountName: string) {
  const params = new URLSearchParams({ account_id: accountId, account_name: accountName })
  return `${path}?${params.toString()}`
}

function PaginationControls({
  page,
  totalPages,
  onPageChange,
}: {
  page: number
  totalPages: number
  onPageChange: (page: number) => void
}) {
  if (totalPages <= 1) return null

  return (
    <div
      className="mt-4 flex items-center justify-between border-t pt-3"
      style={{ borderColor: 'var(--border-subtle)' }}
    >
      <p className="text-xs" style={{ color: 'var(--text-label)' }}>
        Page {page} of {totalPages}
      </p>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
          Previous
        </Button>
        <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>
          Next
        </Button>
      </div>
    </div>
  )
}

function LinkedPanelLoading() {
  return (
    <div className="flex items-center justify-center py-10">
      <Spinner />
    </div>
  )
}

function LinkedPanelError({ onRetry }: { onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-10 text-center">
      <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
        We couldn’t load this panel.
      </p>
      {onRetry && (
        <Button variant="outline" size="sm" onClick={onRetry}>
          Retry
        </Button>
      )}
    </div>
  )
}

function LinkedPanelEmpty({ message, ctaHref, ctaLabel }: { message: string; ctaHref: string; ctaLabel: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-10 text-center">
      <p className="text-sm" style={{ color: 'var(--text-label)' }}>
        {message}
      </p>
      <Link to={ctaHref} className="text-sm font-medium" style={{ color: 'var(--color-primary)' }}>
        {ctaLabel}
      </Link>
    </div>
  )
}

function LinkedEntitiesSection({ accountId, accountName }: { accountId: string; accountName: string }) {
  const [activeTab, setActiveTab] = useState<LinkedEntitiesTab>('contacts')
  const [pages, setPages] = useState<LinkedEntityPageState>({
    contacts: 1,
    deals: 1,
    quotes: 1,
    tickets: 1,
    sequences: 1,
    inbox: 1,
  })

  const contactsQuery = useAccountContacts(accountId)
  const dealsQuery = useDeals({
    account_id: accountId,
    page: pages.deals,
    per_page: LINKED_ENTITY_PAGE_SIZE,
    sort_by: 'created_at',
    sort_dir: 'desc',
  })
  const quotesQuery = useQuotes({
    account_id: accountId,
    page: pages.quotes,
    limit: LINKED_ENTITY_PAGE_SIZE,
  })
  const ticketsQuery = useTickets({
    account_id: accountId,
    page: pages.tickets,
    per_page: LINKED_ENTITY_PAGE_SIZE,
    sort_by: 'created_at',
    sort_dir: 'desc',
  })

  const contactIds = useMemo(
    () => (contactsQuery.data ?? []).map((contact) => contact.id),
    [contactsQuery.data]
  )

  const sequencesQuery = useQuery({
    queryKey: ['accounts', accountId, 'linked-sequences', contactIds],
    enabled: contactIds.length > 0,
    staleTime: 30_000,
    queryFn: async () => {
      const result = await sequencesApi.list({ page: 1, limit: 200 })
      const matches: EmailSequence[] = []
      const contactIdSet = new Set(contactIds)

      await Promise.all(
        (result.data ?? []).map(async (sequence) => {
          const enrollments = await sequencesApi.listEnrollments(sequence.id)
          if ((enrollments.data ?? []).some((enrollment) => contactIdSet.has(enrollment.contact_id))) {
            matches.push(sequence)
          }
        })
      )

      return matches.sort(
        (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
      )
    },
  })

  const inboxQuery = useQuery({
    queryKey: ['accounts', accountId, 'linked-inbox', contactIds],
    enabled: contactIds.length > 0,
    staleTime: 30_000,
    queryFn: async () => {
      const byThreadId = new Map<string, InboxThread>()

      await Promise.all(
        contactIds.map(async (contactId) => {
          const response = await inboxApi.listThreads({ contact_id: contactId, page: 1, limit: 50 })
          for (const thread of response.data ?? []) {
            const existing = byThreadId.get(thread.thread_id)
            if (!existing || new Date(thread.last_message_at) > new Date(existing.last_message_at)) {
              byThreadId.set(thread.thread_id, thread)
            }
          }
        })
      )

      return Array.from(byThreadId.values()).sort(
        (a, b) => new Date(b.last_message_at).getTime() - new Date(a.last_message_at).getTime()
      )
    },
  })

  const linkContact = useLinkContactToAccount()
  const linkDeal = useLinkDealToAccount()
  const linkTicket = useLinkTicketToAccount()
  const [activeLinkModal, setActiveLinkModal] = useState<'contacts' | 'deals' | 'tickets' | null>(null)

  const contactsTotal = contactsQuery.data?.length ?? 0
  const contactsPageItems = useMemo(() => {
    const allContacts = contactsQuery.data ?? []
    const start = (pages.contacts - 1) * LINKED_ENTITY_PAGE_SIZE
    return allContacts.slice(start, start + LINKED_ENTITY_PAGE_SIZE)
  }, [contactsQuery.data, pages.contacts])
  const contactsTotalPages = Math.max(1, Math.ceil(contactsTotal / LINKED_ENTITY_PAGE_SIZE))

  const sequences = useMemo(() => sequencesQuery.data ?? [], [sequencesQuery.data])
  const sequencesTotal = sequences.length
  const sequencesPageItems = useMemo(() => {
    const start = (pages.sequences - 1) * LINKED_ENTITY_PAGE_SIZE
    return sequences.slice(start, start + LINKED_ENTITY_PAGE_SIZE)
  }, [pages.sequences, sequences])
  const sequencesTotalPages = Math.max(1, Math.ceil(sequencesTotal / LINKED_ENTITY_PAGE_SIZE))

  const inboxThreads = useMemo(() => inboxQuery.data ?? [], [inboxQuery.data])
  const inboxTotal = inboxThreads.length
  const inboxPageItems = useMemo(() => {
    const start = (pages.inbox - 1) * LINKED_ENTITY_PAGE_SIZE
    return inboxThreads.slice(start, start + LINKED_ENTITY_PAGE_SIZE)
  }, [inboxThreads, pages.inbox])
  const inboxTotalPages = Math.max(1, Math.ceil(inboxTotal / LINKED_ENTITY_PAGE_SIZE))

  const counts: Record<LinkedEntitiesTab, number> = {
    contacts: contactsTotal,
    deals: dealsQuery.data?.meta?.total ?? 0,
    quotes: quotesQuery.data?.meta?.total ?? 0,
    tickets: ticketsQuery.data?.meta?.total ?? 0,
    sequences: sequencesTotal,
    inbox: inboxTotal,
  }

  const setPageFor = (tab: LinkedEntitiesTab, page: number) => {
    setPages((prev) => ({ ...prev, [tab]: page }))
  }

  const activeRoute = buildAccountScopedPath(LINKED_TAB_META[activeTab].route, accountId, accountName)

  const renderContactsPanel = () => {
    if (contactsQuery.isLoading) return <LinkedPanelLoading />
    if (contactsQuery.isError) return <LinkedPanelError onRetry={() => void contactsQuery.refetch()} />
    if (!contactsTotal) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.contacts.empty} ctaHref={activeRoute} ctaLabel="Open Contacts" />
    }

    return (
      <>
        <ul className="space-y-2">
          {contactsPageItems.map((contact) => (
            <li key={contact.id} className="flex items-center gap-2">
              <Link
                to={`/contacts/${contact.id}`}
                className="flex flex-1 items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-[var(--surface-app)]"
              >
                <div
                  className="flex h-8 w-8 items-center justify-center rounded-full text-[11px] font-bold"
                  style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
                >
                  {((contact.first_name?.[0] ?? '') + (contact.last_name?.[0] ?? '')).toUpperCase() || '?'}
                </div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {contact.first_name} {contact.last_name}
                  </p>
                  <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
                    {contact.email ?? contact.title ?? 'No email yet'}
                  </p>
                </div>
              </Link>
              <button
                onClick={() => linkContact.mutate({ contactId: contact.id, accountId: null })}
                className="rounded p-1 transition-colors hover:bg-[var(--surface-app)]"
                aria-label="Unlink contact"
              >
                <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
              </button>
            </li>
          ))}
        </ul>
        <PaginationControls
          page={pages.contacts}
          totalPages={contactsTotalPages}
          onPageChange={(page) => setPageFor('contacts', page)}
        />
      </>
    )
  }

  const renderDealsPanel = () => {
    if (dealsQuery.isLoading) return <LinkedPanelLoading />
    if (dealsQuery.isError) return <LinkedPanelError onRetry={() => void dealsQuery.refetch()} />
    const deals = dealsQuery.data?.data ?? []
    const totalPages = dealsQuery.data?.meta?.total_pages ?? 1
    if (!deals.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.deals.empty} ctaHref={activeRoute} ctaLabel="Open Deals" />
    }

    return (
      <>
        <ul className="space-y-2">
          {deals.map((deal) => {
            const stageStyle = STAGE_STYLES[deal.stage] ?? { bg: 'var(--surface-app)', text: 'var(--text-label)' }
            return (
              <li key={deal.id} className="flex items-center gap-2 rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {deal.title}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
                    {formatCurrency((deal.value_cents ?? 0) / 100)}
                  </p>
                </div>
                <span
                  className="shrink-0 rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase"
                  style={{ background: stageStyle.bg, color: stageStyle.text }}
                >
                  {deal.stage.replace('_', ' ')}
                </span>
                <button
                  onClick={() => linkDeal.mutate({ dealId: deal.id, accountId: null })}
                  className="rounded p-1 transition-colors hover:bg-white/70"
                  aria-label="Unlink deal"
                >
                  <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
                </button>
              </li>
            )
          })}
        </ul>
        <PaginationControls page={pages.deals} totalPages={totalPages} onPageChange={(page) => setPageFor('deals', page)} />
      </>
    )
  }

  const renderQuotesPanel = () => {
    if (quotesQuery.isLoading) return <LinkedPanelLoading />
    if (quotesQuery.isError) return <LinkedPanelError onRetry={() => void quotesQuery.refetch()} />
    const quotes = quotesQuery.data?.data ?? []
    const totalPages = quotesQuery.data?.meta?.total_pages ?? 1
    if (!quotes.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.quotes.empty} ctaHref={activeRoute} ctaLabel="Open Quotes" />
    }

    return (
      <>
        <ul className="space-y-2">
          {quotes.map((quote) => (
            <li key={quote.id} className="rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {quote.title}
                  </p>
                  <div className="mt-1 flex flex-wrap items-center gap-1.5">
                    <span
                      className="rounded-full px-2 py-0.5 text-[11px] font-medium"
                      style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
                    >
                      {accountName}
                    </span>
                    <span className="text-xs" style={{ color: 'var(--text-label)' }}>
                      {quote.deal?.title ?? 'No linked deal'}
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold" style={{ color: 'var(--color-primary)' }}>
                    {formatCurrency(quote.total_cents / 100, quote.currency)}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-label)' }}>
                    {quote.status}
                  </p>
                </div>
              </div>
            </li>
          ))}
        </ul>
        <PaginationControls page={pages.quotes} totalPages={totalPages} onPageChange={(page) => setPageFor('quotes', page)} />
      </>
    )
  }

  const renderTicketsPanel = () => {
    if (ticketsQuery.isLoading) return <LinkedPanelLoading />
    if (ticketsQuery.isError) return <LinkedPanelError onRetry={() => void ticketsQuery.refetch()} />
    const tickets = ticketsQuery.data?.data ?? []
    const totalPages = ticketsQuery.data?.meta?.total_pages ?? 1
    if (!tickets.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.tickets.empty} ctaHref={activeRoute} ctaLabel="Open Tickets" />
    }

    return (
      <>
        <ul className="space-y-2">
          {tickets.map((ticket) => {
            const statusStyle = TICKET_STATUS_STYLES[ticket.status] ?? { bg: 'var(--surface-app)', text: 'var(--text-label)' }
            return (
              <li key={ticket.id} className="flex items-center gap-2 rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
                <div className="min-w-0 flex-1">
                  <Link to={`/tickets/${ticket.id}`} className="block truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {ticket.subject}
                  </Link>
                  <p className="text-xs" style={{ color: 'var(--text-label)' }}>
                    {ticket.contact?.name ?? 'No linked contact'}
                  </p>
                </div>
                <span
                  className="shrink-0 rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase"
                  style={{ background: statusStyle.bg, color: statusStyle.text }}
                >
                  {ticket.status}
                </span>
                <button
                  onClick={() => linkTicket.mutate({ ticketId: ticket.id, accountId: null })}
                  className="rounded p-1 transition-colors hover:bg-white/70"
                  aria-label="Unlink ticket"
                >
                  <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
                </button>
              </li>
            )
          })}
        </ul>
        <PaginationControls page={pages.tickets} totalPages={totalPages} onPageChange={(page) => setPageFor('tickets', page)} />
      </>
    )
  }

  const renderSequencesPanel = () => {
    if (contactsQuery.isLoading || sequencesQuery.isLoading) return <LinkedPanelLoading />
    if (contactsQuery.isError || sequencesQuery.isError) {
      return <LinkedPanelError onRetry={() => void sequencesQuery.refetch()} />
    }
    if (!contactIds.length || !sequencesPageItems.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.sequences.empty} ctaHref={activeRoute} ctaLabel="Open Sequences" />
    }

    return (
      <>
        <ul className="space-y-2">
          {sequencesPageItems.map((sequence) => (
            <li key={sequence.id} className="rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {sequence.name}
                  </p>
                  <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
                    {sequence.description || 'No description'}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
                    {sequence.enrolled_count}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-label)' }}>
                    enrolled
                  </p>
                </div>
              </div>
            </li>
          ))}
        </ul>
        <PaginationControls
          page={pages.sequences}
          totalPages={sequencesTotalPages}
          onPageChange={(page) => setPageFor('sequences', page)}
        />
      </>
    )
  }

  const renderInboxPanel = () => {
    if (contactsQuery.isLoading || inboxQuery.isLoading) return <LinkedPanelLoading />
    if (contactsQuery.isError || inboxQuery.isError) {
      return <LinkedPanelError onRetry={() => void inboxQuery.refetch()} />
    }
    if (!contactIds.length || !inboxPageItems.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.inbox.empty} ctaHref={activeRoute} ctaLabel="Open Inbox" />
    }

    return (
      <>
        <ul className="space-y-2">
          {inboxPageItems.map((thread) => (
            <li key={thread.thread_id} className="rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {thread.subject}
                  </p>
                  <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
                    {thread.participants.join(', ')}
                  </p>
                  <p className="mt-1 truncate text-xs" style={{ color: 'var(--text-secondary)' }}>
                    {thread.snippet}
                  </p>
                </div>
                <div className="shrink-0 text-right">
                  <p className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>
                    {formatRelativeTime(thread.last_message_at)}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-label)' }}>
                    {thread.message_count} messages
                  </p>
                </div>
              </div>
            </li>
          ))}
        </ul>
        <PaginationControls page={pages.inbox} totalPages={inboxTotalPages} onPageChange={(page) => setPageFor('inbox', page)} />
      </>
    )
  }

  const renderActivePanel = () => {
    switch (activeTab) {
      case 'contacts':
        return renderContactsPanel()
      case 'deals':
        return renderDealsPanel()
      case 'quotes':
        return renderQuotesPanel()
      case 'tickets':
        return renderTicketsPanel()
      case 'sequences':
        return renderSequencesPanel()
      case 'inbox':
        return renderInboxPanel()
    }
  }

  const showLinkButton = activeTab === 'contacts' || activeTab === 'deals' || activeTab === 'tickets'
  const openActiveLinkModal = () => {
    if (activeTab === 'contacts' || activeTab === 'deals' || activeTab === 'tickets') {
      setActiveLinkModal(activeTab)
    }
  }
  const closeActiveLinkModal = () => setActiveLinkModal(null)

  const renderActiveLinkModal = () => {
    switch (activeLinkModal) {
      case 'contacts':
        return (
          <EntityLinkModal<Contact>
            open
            onClose={closeActiveLinkModal}
            title="Link Contact"
            placeholder="Search by name or email…"
            search={contactsApi.search}
            onSelect={(contact) => linkContact.mutateAsync({ contactId: contact.id, accountId })}
            renderItem={(contact) => (
              <div className="flex items-center gap-2.5">
                <div
                  className="h-7 w-7 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0"
                  style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
                >
                  {((contact.first_name?.[0] ?? '') + (contact.last_name?.[0] ?? '')).toUpperCase() || '?'}
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                    {contact.first_name} {contact.last_name}
                  </p>
                  <p className="text-xs truncate" style={{ color: 'var(--text-label)' }}>
                    {contact.email ?? 'No email'}
                  </p>
                </div>
              </div>
            )}
            getKey={(contact) => contact.id}
          />
        )
      case 'deals':
        return (
          <EntityLinkModal<Deal>
            open
            onClose={closeActiveLinkModal}
            title="Link Deal"
            placeholder="Search deals by title…"
            search={dealsApi.search}
            onSelect={(deal) => linkDeal.mutateAsync({ dealId: deal.id, accountId })}
            renderItem={(deal) => (
              <div className="flex items-center justify-between gap-2">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                    {deal.title}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
                    {formatCurrency((deal.value_cents ?? 0) / 100)}
                  </p>
                </div>
              </div>
            )}
            getKey={(deal) => deal.id}
          />
        )
      case 'tickets':
        return (
          <EntityLinkModal<TicketType>
            open
            onClose={closeActiveLinkModal}
            title="Link Ticket"
            placeholder="Search tickets by subject…"
            search={async (q) => {
              const result = await ticketsApi.list({ search: q, per_page: 10 })
              return result.data ?? []
            }}
            onSelect={(ticket) => linkTicket.mutateAsync({ ticketId: ticket.id, accountId })}
            mapError={(error, ticket) => {
              const mapped = mapCrmLinkError(error)
              if (mapped.kind !== 'account_contact_mismatch') return null

              const ticketContactId = ticket.contact?.id
              const ticketContactName = ticket.contact?.name ?? 'current contact'
              return {
                message: `This ticket is linked to ${ticketContactName}, whose account does not match this account.`,
                actions: [
                  ...(ticketContactId
                    ? [{
                        label: 'Relink contact to account',
                        action: async () => {
                          await contactsApi.update(ticketContactId, { account_id: accountId })
                          await linkTicket.mutateAsync({ ticketId: ticket.id, accountId })
                        },
                      }]
                    : []),
                  {
                    label: 'Keep ticket account and detach contact',
                    action: async () => {
                      await ticketsApi.patchContact(ticket.id, null)
                      await linkTicket.mutateAsync({ ticketId: ticket.id, accountId })
                    },
                  },
                ],
              }
            }}
            renderItem={(ticket) => (
              <div className="flex items-center justify-between gap-2">
                <p className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                  {ticket.subject}
                </p>
                <span className="text-xs" style={{ color: 'var(--text-label)' }}>
                  {ticket.status}
                </span>
              </div>
            )}
            getKey={(ticket) => ticket.id}
          />
        )
      default:
        return null
    }
  }

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <p
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          Linked Entities
        </p>
        <Link to={activeRoute} className="ml-auto text-sm font-medium" style={{ color: 'var(--color-primary)' }}>
          Open {LINKED_TAB_META[activeTab].label}
        </Link>
        {showLinkButton && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-xs"
            onClick={openActiveLinkModal}
          >
            <Plus className="h-3 w-3" />
            Link
          </Button>
        )}
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        {LINKED_TAB_ORDER.map((tab) => {
          const Icon = LINKED_TAB_META[tab].icon
          const isActive = activeTab === tab
          return (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm transition-colors"
              style={{
                borderColor: isActive ? 'var(--color-primary)' : 'var(--border-default)',
                background: isActive ? 'var(--color-primary-light)' : 'transparent',
                color: isActive ? 'var(--color-primary)' : 'var(--text-secondary)',
              }}
            >
              <Icon className="h-3.5 w-3.5" />
              <span>{LINKED_TAB_META[tab].label}</span>
              <span
                className="rounded-full px-1.5 py-0.5 text-[11px] font-semibold"
                style={{
                  background: isActive ? 'var(--surface-card)' : 'var(--surface-app)',
                  color: isActive ? 'var(--color-primary)' : 'var(--text-label)',
                }}
              >
                {counts[tab]}
              </span>
            </button>
          )
        })}
      </div>

      {renderActivePanel()}
      {renderActiveLinkModal()}
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function AccountDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const { data: account, isLoading, isError } = useAccount(id!)
  const { data: accountContacts } = useAccountContacts(id!)
  const deleteAccount = useDeleteAccount()
  const [relationshipRows, setRelationshipRows] = useState<RelationshipRow[]>([])
  const accountRelationshipSeed = useMemo(
    () => createAccountRelationshipRows(accountContacts),
    [accountContacts]
  )
  const accountRelationshipSeedKey = useMemo(
    () => accountRelationshipSeed.map((row) => `${row.entityId ?? row.id}:${row.label}:${row.meta}`).join('|'),
    [accountRelationshipSeed]
  )

  useEffect(() => {
    if (!id) return
    setRelationshipRows(accountRelationshipSeed)
  }, [id, accountRelationshipSeed, accountRelationshipSeedKey])

  if (isLoading) {
    return (
      <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
        <div className="h-4 w-48 bg-slate-100 animate-pulse rounded mb-5" />
        <div
          className="rounded-xl border p-5 flex items-center gap-4 mb-5"
          style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        >
          <div className="h-[72px] w-[72px] rounded-xl bg-slate-100 animate-pulse shrink-0" />
          <div className="space-y-2 flex-1">
            <div className="h-6 w-56 bg-slate-100 animate-pulse rounded" />
            <div className="h-3.5 w-36 bg-slate-100 animate-pulse rounded" />
          </div>
        </div>
        <div className="flex gap-5">
          <div className="flex-1 space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-32 bg-slate-100 animate-pulse rounded-xl" />
            ))}
          </div>
          <div className="w-[300px] shrink-0 space-y-3">
            <div className="h-48 bg-slate-100 animate-pulse rounded-xl" />
            <div className="h-32 bg-slate-100 animate-pulse rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !account) {
    return (
      <div className="p-6 flex flex-col items-center justify-center py-24 text-center">
        <p className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
          Account not found.
        </p>
        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          This account may have been deleted or you may not have access.
        </p>
        <Link to="/accounts" className="text-sm font-medium" style={{ color: 'var(--color-primary)' }}>
          ← Back to Accounts
        </Link>
      </div>
    )
  }

  const handleDelete = async () => {
    await deleteAccount.mutateAsync(account.id)
    navigate('/accounts')
  }

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 mb-5 text-sm" style={{ color: 'var(--text-label)' }}>
        <Link to="/accounts" className="hover:underline" style={{ color: 'var(--color-primary)' }}>
          Accounts
        </Link>
        <ChevronRight className="h-3.5 w-3.5" />
        <span style={{ color: 'var(--text-primary)' }}>{account.name}</span>
      </nav>

      {/* Header Card */}
      <div
        className="rounded-xl border p-5 mb-5 flex flex-wrap items-center gap-4"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      >
        {/* Avatar */}
        <div
          className="h-[72px] w-[72px] rounded-xl flex items-center justify-center shrink-0 text-2xl font-bold"
          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
        >
          {accountInitials(account.name)}
        </div>

        {/* Identity */}
        <div className="flex-1 min-w-0">
          <h1
            className="font-bold leading-tight"
            style={{ fontSize: 24, color: 'var(--text-primary)', letterSpacing: '-0.01em' }}
          >
            {account.name}
          </h1>
          {account.industry && (
            <p className="mt-0.5 text-sm" style={{ color: 'var(--text-secondary)' }}>
              {account.industry.charAt(0).toUpperCase() + account.industry.slice(1)}
              {account.size ? ` · ${account.size} employees` : ''}
            </p>
          )}
          {account.domain && (
            <a
              href={`https://${account.domain}`}
              target="_blank"
              rel="noreferrer"
              className="mt-0.5 inline-flex items-center gap-1 text-sm font-medium hover:underline"
              style={{ color: 'var(--color-primary)' }}
            >
              {account.domain}
              <ExternalLink className="h-3 w-3" />
            </a>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/search?q=${encodeURIComponent(account.name)}&account_id=${account.id}&account_name=${encodeURIComponent(account.name)}`)}
          >
            <Search className="h-3.5 w-3.5" />
            Search related
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowDeleteConfirm(true)}
            className="text-red-600 border-red-200 hover:bg-red-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Delete
          </Button>
        </div>
      </div>

      {/* Body */}
      <div className="space-y-5">
        <div
          className="rounded-xl border p-5"
          style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        >
          <p
            className="mb-4 text-[11px] font-semibold uppercase tracking-widest"
            style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          >
            Details
          </p>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
            {account.phone && (
              <div>
                <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                  <Phone className="h-3.5 w-3.5" /> Phone
                </dt>
                <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {account.phone}
                </dd>
              </div>
            )}
            {account.website && (
              <div>
                <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                  <Globe className="h-3.5 w-3.5" /> Website
                </dt>
                <dd>
                  <a
                    href={account.website}
                    target="_blank"
                    rel="noreferrer"
                    className="text-sm font-medium hover:underline"
                    style={{ color: 'var(--color-primary)' }}
                  >
                    {account.website}
                  </a>
                </dd>
              </div>
            )}
            {account.address && (
              <div className="sm:col-span-2">
                <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                  <MapPin className="h-3.5 w-3.5" /> Address
                </dt>
                <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {account.address}
                </dd>
              </div>
            )}
            <div>
              <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                <Building2 className="h-3.5 w-3.5" /> Created
              </dt>
              <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                {formatDate(account.created_at)}
              </dd>
            </div>
            <div>
              <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                <Building2 className="h-3.5 w-3.5" /> Last Updated
              </dt>
              <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                {formatRelativeTime(account.updated_at)}
              </dd>
            </div>
          </dl>
        </div>

        <RelationshipEditor
          title="Relationship Roles"
          entityLabel="Contact"
          value={relationshipRows}
          onChange={setRelationshipRows}
          emptyMessage="No contact relationships yet."
          addLabel="Add contact role"
        />

        <UnifiedTimeline entityType="account" entityId={account.id} />

        <LinkedEntitiesSection accountId={account.id} accountName={account.name} />

        <NotesPanel accountId={account.id} />
      </div>

      {/* Delete confirm */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div
            className="rounded-xl border shadow-xl p-6 w-full max-w-sm"
            style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
          >
            <h2 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
              Delete Account
            </h2>
            <p className="text-sm mb-5" style={{ color: 'var(--text-secondary)' }}>
              Are you sure you want to delete <strong>{account.name}</strong>? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowDeleteConfirm(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                disabled={deleteAccount.isPending}
                onClick={handleDelete}
                className="bg-red-600 hover:bg-red-700 text-white border-red-600"
              >
                {deleteAccount.isPending ? 'Deleting…' : 'Delete'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
