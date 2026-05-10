import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ChevronRight,
  Pencil,
  Mail,
  ExternalLink,
  Lock,
  Download,
  ClipboardList,
  Trash2,
  Sparkles,
  Loader2,
  RefreshCw,
  Building2,
  Briefcase,
  FileText,
  Plus,
  Ticket,
  X,
  Zap,
  Search,
} from 'lucide-react'
import { useContact, useDeleteContact, useUpdateContact, useContactNotes, useAddContactNote, useEnrichContact, contactKeys } from '@/hooks/useContacts'
import { accountKeys } from '@/hooks/useAccounts'
import { dealKeys } from '@/hooks/useDeals'
import { ticketKeys } from '@/hooks/useTickets'
import { quoteKeys } from '@/hooks/useQuotes'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { useToast } from '@/components/ui/Toast'
import { formatDate, formatRelativeTime, formatCurrency } from '@/lib/utils'
import { mapCrmLinkError } from '@/lib/crmLinkErrors'
import { accountsApi } from '@/api/accounts'
import { contactsApi } from '@/api/contacts'
import { dealsApi } from '@/api/deals'
import { inboxApi } from '@/api/inbox'
import { quotesApi } from '@/api/quotes'
import { sequencesApi } from '@/api/sequences'
import { ticketsApi } from '@/api/tickets'
import type { Account, Contact, Deal, EmailSequence, EnrichmentResult, PaginatedResponse, Ticket as TicketType, Quote } from '@/api/types'
import { AccountForm } from '@/components/omnir/AccountForm'
import { ComposeEmailModal } from '@/components/omnir/ComposeEmailModal'
import { DealForm } from '@/components/omnir/DealForm'
import { EntityLinkModal } from '@/components/omnir/EntityLinkModal'
import { UnifiedTimeline } from '@/components/omnir/UnifiedTimeline'
import {
  RelationshipEditor,
  type RelationshipRow,
  createRelationshipRow,
  normalizeRelationshipRows,
} from '@/components/omnir/RelationshipEditor'
import { TicketForm } from '@/components/omnir/TicketForm'

// ── Helpers ──────────────────────────────────────────────────────────────────

function contactInitials(firstName: string, lastName: string) {
  return `${firstName[0] ?? ''}${lastName[0] ?? ''}`.toUpperCase()
}

type ContactLinkedAccount = Account
type ContactRecord = Contact & { linked_accounts?: ContactLinkedAccount[] }

function getLinkedAccounts(contact: ContactRecord | null | undefined): ContactLinkedAccount[] {
  if (!contact) return []
  if (contact.linked_accounts?.length) return contact.linked_accounts
  return contact.account ? [contact.account] : []
}

function syncPrimaryAccount(contact: ContactRecord, linkedAccounts: ContactLinkedAccount[]): ContactRecord {
  return {
    ...contact,
    linked_accounts: linkedAccounts,
    account: linkedAccounts[0],
    account_id: linkedAccounts[0]?.id,
  }
}

function createContactRelationshipRows(linkedAccounts: ContactLinkedAccount[]): RelationshipRow[] {
  if (!linkedAccounts.length) return []

  return normalizeRelationshipRows(
    linkedAccounts.map((account, index) =>
      createRelationshipRow({
        id: `contact-relationship-${account.id}`,
        entityId: account.id,
        label: account.name,
        meta: account.industry ?? account.domain ?? '',
        role: index === 0 ? 'primary' : 'billing',
        isPrimary: index === 0,
      })
    )
  )
}

// ── Linked Entities ───────────────────────────────────────────────────────────

type LinkedEntitiesTab = 'accounts' | 'deals' | 'quotes' | 'tickets' | 'sequences' | 'inbox'
type LinkedEntityPageState = Record<LinkedEntitiesTab, number>

const LINKED_ENTITY_PAGE_SIZE = 5
const ASSOCIATION_FETCH_SIZE = 100

const LINKED_TAB_ORDER: LinkedEntitiesTab[] = [
  'accounts',
  'deals',
  'quotes',
  'tickets',
  'sequences',
  'inbox',
]

const LINKED_TAB_META: Record<
  LinkedEntitiesTab,
  { label: string; route: string; icon: typeof Building2; empty: string }
> = {
  accounts: { label: 'Accounts', route: '/accounts', icon: Building2, empty: 'No accounts linked yet.' },
  deals: { label: 'Deals', route: '/deals', icon: Briefcase, empty: 'No deals linked yet.' },
  quotes: { label: 'Quotes', route: '/quotes', icon: FileText, empty: 'No quotes linked yet.' },
  tickets: { label: 'Tickets', route: '/tickets', icon: Ticket, empty: 'No tickets linked yet.' },
  sequences: { label: 'Sequences', route: '/sequences', icon: Zap, empty: 'No sequences linked yet.' },
  inbox: { label: 'Inbox', route: '/inbox', icon: Mail, empty: 'No inbox threads linked yet.' },
}

function buildContactScopedPath(path: string, contactId: string, contactName: string) {
  const params = new URLSearchParams({
    contact_id: contactId,
    contact_name: contactName,
  })
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

function LinkedPanelEmpty({
  message,
  ctaHref,
  ctaLabel,
}: {
  message: string
  ctaHref: string
  ctaLabel: string
}) {
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

function LinkedEntitiesSection({
  contact,
  contactId,
  contactName,
}: {
  contact: ContactRecord
  contactId: string
  contactName: string
}) {
  const [activeTab, setActiveTab] = useState<LinkedEntitiesTab>('accounts')
  const [pages, setPages] = useState<LinkedEntityPageState>({
    accounts: 1,
    deals: 1,
    quotes: 1,
    tickets: 1,
    sequences: 1,
    inbox: 1,
  })
  const [showAccountLinkModal, setShowAccountLinkModal] = useState(false)
  const [showDealLinkModal, setShowDealLinkModal] = useState(false)
  const [showTicketLinkModal, setShowTicketLinkModal] = useState(false)
  const [showCreateAccount, setShowCreateAccount] = useState(false)
  const [showCreateDeal, setShowCreateDeal] = useState(false)
  const [showCreateTicket, setShowCreateTicket] = useState(false)
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const linkedAccounts = useMemo(() => getLinkedAccounts(contact), [contact])
  const primaryLinkedAccount = linkedAccounts[0]
  const dealsParams = useMemo(
    () => ({
      contact_id: contactId,
      page: 1,
      per_page: ASSOCIATION_FETCH_SIZE,
      sort_by: 'created_at' as const,
      sort_dir: 'desc' as const,
    }),
    [contactId]
  )
  const ticketsParams = useMemo(
    () => ({
      contact_id: contactId,
      page: 1,
      per_page: ASSOCIATION_FETCH_SIZE,
      sort_by: 'created_at' as const,
      sort_dir: 'desc' as const,
    }),
    [contactId]
  )

  const dealsQuery = useQuery({
    queryKey: dealKeys.list(dealsParams),
    staleTime: 30_000,
    queryFn: async () => {
      let page = 1
      let total = 0
      let totalPages = 1
      const allDeals: Deal[] = []

      let hasMore = true
      while (hasMore) {
        const response = await dealsApi.list({ ...dealsParams, page })
        const pageDeals = response.data ?? []
        const pageMeta = response.meta
        allDeals.push(...pageDeals)
        total = pageMeta.total ?? allDeals.length
        totalPages = pageMeta.total_pages ?? Math.max(1, Math.ceil(total / dealsParams.per_page))

        hasMore = pageDeals.length > 0 && page < totalPages
        page += 1
      }

      return {
        data: allDeals,
        meta: {
          page: 1,
          per_page: dealsParams.per_page,
          total,
          total_pages: Math.max(1, Math.ceil(total / dealsParams.per_page)),
        },
      } satisfies PaginatedResponse<Deal>
    },
  })

  const ticketsQuery = useQuery({
    queryKey: ticketKeys.list(ticketsParams),
    staleTime: 30_000,
    queryFn: async () => {
      let page = 1
      let total = 0
      let totalPages = 1
      const allTickets: TicketType[] = []

      let hasMore = true
      while (hasMore) {
        const response = await ticketsApi.list({ ...ticketsParams, page })
        const pageTickets = response.data ?? []
        const pageMeta = response.meta
        allTickets.push(...pageTickets)
        total = pageMeta.total ?? allTickets.length
        totalPages = pageMeta.total_pages ?? Math.max(1, Math.ceil(total / ticketsParams.per_page))

        hasMore = pageTickets.length > 0 && page < totalPages
        page += 1
      }

      return {
        data: allTickets,
        meta: {
          page: 1,
          per_page: ticketsParams.per_page,
          total,
          total_pages: Math.max(1, Math.ceil(total / ticketsParams.per_page)),
        },
      } satisfies PaginatedResponse<TicketType>
    },
  })

  const quotesQuery = useQuery({
    queryKey: ['contacts', contactId, 'linked-quotes', dealsQuery.data?.data?.map((deal) => deal.id) ?? []],
    enabled: (dealsQuery.data?.data?.length ?? 0) > 0,
    staleTime: 30_000,
    queryFn: async () => {
      const byId = new Map<string, Quote>()
      await Promise.all(
        (dealsQuery.data?.data ?? []).map(async (deal) => {
          const response = await quotesApi.listByDeal(deal.id)
          for (const quote of response.data ?? []) {
            byId.set(quote.id, quote)
          }
        })
      )
      return Array.from(byId.values()).sort(
        (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
      )
    },
  })

  const sequencesQuery = useQuery({
    queryKey: ['contacts', contactId, 'linked-sequences'],
    staleTime: 30_000,
    queryFn: async () => {
      const result = await sequencesApi.list({ page: 1, limit: 200 })
      const matches: EmailSequence[] = []
      await Promise.all(
        (result.data ?? []).map(async (sequence) => {
          const enrollments = await sequencesApi.listEnrollments(sequence.id)
          if ((enrollments.data ?? []).some((enrollment) => enrollment.contact_id === contactId)) {
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
    queryKey: ['contacts', contactId, 'linked-inbox'],
    staleTime: 30_000,
    queryFn: () => inboxApi.listThreads({ contact_id: contactId, page: 1, limit: ASSOCIATION_FETCH_SIZE }),
  })

  const accountMutation = useMutation({
    mutationFn: async ({ accountId }: { accountId: string | null; account?: Account }) =>
      contactsApi.update(contactId, { account_id: accountId }),
    onMutate: async ({ accountId, account }) => {
      await queryClient.cancelQueries({ queryKey: contactKeys.detail(contactId) })
      const previousContact = queryClient.getQueryData<ContactRecord>(contactKeys.detail(contactId))
      if (previousContact) {
        const currentLinkedAccounts = getLinkedAccounts(previousContact)
        const nextLinkedAccounts = accountId
          ? account
            ? [account]
            : currentLinkedAccounts.filter((linked) => linked.id === accountId)
          : currentLinkedAccounts.filter((linked) => linked.id !== account?.id)
        queryClient.setQueryData<ContactRecord>(
          contactKeys.detail(contactId),
          syncPrimaryAccount(previousContact, nextLinkedAccounts)
        )
      }
      return { previousContact }
    },
    onError: (_error, variables, context) => {
      if (context?.previousContact) {
        queryClient.setQueryData(contactKeys.detail(contactId), context.previousContact)
      }
      toast({
        title: variables.accountId ? 'Could not link account' : 'Could not unlink account',
        description: 'Your changes were rolled back.',
        variant: 'destructive',
      })
    },
    onSuccess: (_result, variables) => {
      toast({
        title: variables.accountId ? 'Account linked' : 'Account unlinked',
      })
    },
    onSettled: async (_result, _error, variables) => {
      await queryClient.invalidateQueries({ queryKey: contactKeys.detail(contactId) })
      await queryClient.invalidateQueries({ queryKey: contactKeys.lists() })
      if (variables.accountId) {
        await queryClient.invalidateQueries({ queryKey: accountKeys.detail(variables.accountId) })
        await queryClient.invalidateQueries({ queryKey: accountKeys.contacts(variables.accountId) })
      }
    },
  })

  const dealMutation = useMutation({
    mutationFn: async ({ dealId, shouldLink }: { dealId: string; shouldLink: boolean; deal: Deal }) =>
      dealsApi.update(dealId, { contact_id: shouldLink ? contactId : null }),
    onMutate: async ({ shouldLink, deal }) => {
      await queryClient.cancelQueries({ queryKey: dealKeys.list(dealsParams) })
      const previousDeals = queryClient.getQueryData<PaginatedResponse<Deal>>(dealKeys.list(dealsParams))
      if (previousDeals) {
        const existingDeals = previousDeals.data ?? []
        const filteredDeals = existingDeals.filter((item) => item.id !== deal.id)
        const nextDeals = shouldLink ? [deal, ...filteredDeals] : filteredDeals
        queryClient.setQueryData<PaginatedResponse<Deal>>(dealKeys.list(dealsParams), {
          ...previousDeals,
          data: nextDeals,
          meta: {
            ...previousDeals.meta,
            total: shouldLink ? previousDeals.meta.total + 1 : Math.max(0, previousDeals.meta.total - 1),
            total_pages: Math.max(
              1,
              Math.ceil(
                (shouldLink ? previousDeals.meta.total + 1 : Math.max(0, previousDeals.meta.total - 1)) /
                  previousDeals.meta.per_page
              )
            ),
          },
        })
      }
      return { previousDeals }
    },
    onError: (_error, variables, context) => {
      if (context?.previousDeals) {
        queryClient.setQueryData(dealKeys.list(dealsParams), context.previousDeals)
      }
      toast({
        title: variables.shouldLink ? 'Could not link deal' : 'Could not unlink deal',
        description: 'Your changes were rolled back.',
        variant: 'destructive',
      })
    },
    onSuccess: (_result, variables) => {
      toast({ title: variables.shouldLink ? 'Deal linked' : 'Deal unlinked' })
    },
    onSettled: async (_result, _error, variables) => {
      await queryClient.invalidateQueries({ queryKey: dealKeys.list(dealsParams) })
      await queryClient.invalidateQueries({ queryKey: dealKeys.detail(variables.dealId) })
      await queryClient.invalidateQueries({ queryKey: ['contacts', contactId, 'linked-quotes'] })
      await queryClient.invalidateQueries({ queryKey: quoteKeys.lists() })
    },
  })

  const ticketMutation = useMutation({
    mutationFn: async ({ ticketId, shouldLink }: { ticketId: string; shouldLink: boolean; ticket: TicketType }) =>
      ticketsApi.patchContact(ticketId, shouldLink ? contactId : null),
    onMutate: async ({ shouldLink, ticket }) => {
      await queryClient.cancelQueries({ queryKey: ticketKeys.list(ticketsParams) })
      const previousTickets = queryClient.getQueryData<PaginatedResponse<TicketType>>(ticketKeys.list(ticketsParams))
      if (previousTickets) {
        const existingTickets = previousTickets.data ?? []
        const filteredTickets = existingTickets.filter((item) => item.id !== ticket.id)
        const nextTickets = shouldLink ? [ticket, ...filteredTickets] : filteredTickets
        queryClient.setQueryData<PaginatedResponse<TicketType>>(ticketKeys.list(ticketsParams), {
          ...previousTickets,
          data: nextTickets,
          meta: {
            ...previousTickets.meta,
            total: shouldLink ? previousTickets.meta.total + 1 : Math.max(0, previousTickets.meta.total - 1),
            total_pages: Math.max(
              1,
              Math.ceil(
                (shouldLink ? previousTickets.meta.total + 1 : Math.max(0, previousTickets.meta.total - 1)) /
                  previousTickets.meta.per_page
              )
            ),
          },
        })
      }
      return { previousTickets }
    },
    onError: (_error, variables, context) => {
      if (context?.previousTickets) {
        queryClient.setQueryData(ticketKeys.list(ticketsParams), context.previousTickets)
      }
      const mappedError = mapCrmLinkError(_error)
      if (mappedError.kind === 'account_contact_mismatch') return
      toast({
        title: variables.shouldLink ? 'Could not link ticket' : 'Could not unlink ticket',
        description: 'Your changes were rolled back.',
        variant: 'destructive',
      })
    },
    onSuccess: (_result, variables) => {
      toast({ title: variables.shouldLink ? 'Ticket linked' : 'Ticket unlinked' })
    },
    onSettled: async (_result, _error, variables) => {
      await queryClient.invalidateQueries({ queryKey: ticketKeys.list(ticketsParams) })
      await queryClient.invalidateQueries({ queryKey: ticketKeys.detail(variables.ticketId) })
    },
  })

  const allDeals = dealsQuery.data?.data ?? []
  const allTickets = ticketsQuery.data?.data ?? []
  const allQuotes = quotesQuery.data ?? []
  const allSequences = sequencesQuery.data ?? []
  const allInboxThreads = inboxQuery.data?.data ?? []

  const setPageFor = (tab: LinkedEntitiesTab, page: number) =>
    setPages((prev) => ({ ...prev, [tab]: page }))

  const paginate = <T,>(items: T[], page: number) => {
    const start = (page - 1) * LINKED_ENTITY_PAGE_SIZE
    return items.slice(start, start + LINKED_ENTITY_PAGE_SIZE)
  }

  const accountsPageItems = paginate(linkedAccounts, pages.accounts)
  const dealsPageItems = paginate(allDeals, pages.deals)
  const quotesPageItems = paginate(allQuotes, pages.quotes)
  const ticketsPageItems = paginate(allTickets, pages.tickets)
  const sequencesPageItems = paginate(allSequences, pages.sequences)
  const inboxPageItems = paginate(allInboxThreads, pages.inbox)

  const counts: Record<LinkedEntitiesTab, number> = {
    accounts: linkedAccounts.length,
    deals: dealsQuery.data?.meta?.total ?? allDeals.length,
    quotes: allQuotes.length,
    tickets: ticketsQuery.data?.meta?.total ?? allTickets.length,
    sequences: allSequences.length,
    inbox: inboxQuery.data?.meta?.total ?? allInboxThreads.length,
  }

  const totalPages = (tab: LinkedEntitiesTab) => Math.max(1, Math.ceil(counts[tab] / LINKED_ENTITY_PAGE_SIZE))
  const activeRoute = buildContactScopedPath(LINKED_TAB_META[activeTab].route, contactId, contactName)
  const showMutableCtas = activeTab === 'accounts' || activeTab === 'deals' || activeTab === 'tickets'
  const showCreateCta = activeTab === 'accounts' || activeTab === 'deals' || activeTab === 'tickets'

  const renderAccountsPanel = () => {
    if (!linkedAccounts.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.accounts.empty} ctaHref={activeRoute} ctaLabel="Open Accounts" />
    }

    return (
      <>
        <ul className="space-y-2">
          {accountsPageItems.map((account) => (
            <li key={account.id} className="flex items-center gap-2 rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <Link to={`/accounts/${account.id}`} className="flex min-w-0 flex-1 items-center gap-3">
                <div
                  className="flex h-8 w-8 items-center justify-center rounded-lg text-[11px] font-bold"
                  style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
                >
                  {account.name.slice(0, 2).toUpperCase()}
                </div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {account.name}
                  </p>
                  <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
                    {account.industry ?? account.domain ?? 'No account details yet'}
                  </p>
                </div>
              </Link>
              <button
                onClick={() => accountMutation.mutate({ accountId: null, account })}
                className="rounded p-1 transition-colors hover:bg-white/70"
                aria-label="Unlink account"
              >
                <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
              </button>
            </li>
          ))}
        </ul>
        <PaginationControls
          page={pages.accounts}
          totalPages={totalPages('accounts')}
          onPageChange={(page) => setPageFor('accounts', page)}
        />
      </>
    )
  }

  const renderDealsPanel = () => {
    if (dealsQuery.isLoading) return <LinkedPanelLoading />
    if (dealsQuery.isError) return <LinkedPanelError onRetry={() => void dealsQuery.refetch()} />
    if (!allDeals.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.deals.empty} ctaHref={activeRoute} ctaLabel="Open Deals" />
    }

    return (
      <>
        <ul className="space-y-2">
          {dealsPageItems.map((deal) => (
            <li key={deal.id} className="flex items-center gap-2 rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <Link to={`/deals`} className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {deal.title}
                </p>
                <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
                  {formatCurrency((deal.value_cents ?? 0) / 100, deal.currency)} · {deal.stage.replace(/_/g, ' ')}
                </p>
              </Link>
              <button
                onClick={() => dealMutation.mutate({ dealId: deal.id, shouldLink: false, deal })}
                className="rounded p-1 transition-colors hover:bg-white/70"
                aria-label="Unlink deal"
              >
                <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
              </button>
            </li>
          ))}
        </ul>
        <PaginationControls page={pages.deals} totalPages={totalPages('deals')} onPageChange={(page) => setPageFor('deals', page)} />
      </>
    )
  }

  const renderQuotesPanel = () => {
    if (dealsQuery.isLoading || quotesQuery.isLoading) return <LinkedPanelLoading />
    if (dealsQuery.isError || quotesQuery.isError) return <LinkedPanelError onRetry={() => void quotesQuery.refetch()} />
    if (!allQuotes.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.quotes.empty} ctaHref={activeRoute} ctaLabel="Open Quotes" />
    }

    return (
      <>
        <ul className="space-y-2">
          {quotesPageItems.map((quote) => {
            const quoteDeal = dealsQuery.data?.data?.find((deal) => deal.id === quote.deal_id)
            const quoteAccountName = quote.account?.name ?? quoteDeal?.account?.name ?? contact.account?.name
            return (
              <li key={quote.id} className="rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
                <div className="flex items-center justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                      {quote.title}
                    </p>
                    <div className="mt-1 flex flex-wrap items-center gap-1.5">
                      {quoteAccountName && (
                        <span
                          className="rounded-full px-2 py-0.5 text-[11px] font-medium"
                          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
                        >
                          {quoteAccountName}
                        </span>
                      )}
                      <span className="text-xs" style={{ color: 'var(--text-label)' }}>
                        {quote.deal?.title ?? quoteDeal?.title ?? 'No linked deal'}
                      </span>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold" style={{ color: 'var(--color-primary)' }}>
                      {formatCurrency(quote.total_cents / 100, quote.currency)}
                    </p>
                    <p className="text-xs uppercase" style={{ color: 'var(--text-label)' }}>
                      {quote.status}
                    </p>
                  </div>
                </div>
              </li>
            )
          })}
        </ul>
        <PaginationControls page={pages.quotes} totalPages={totalPages('quotes')} onPageChange={(page) => setPageFor('quotes', page)} />
      </>
    )
  }

  const renderTicketsPanel = () => {
    if (ticketsQuery.isLoading) return <LinkedPanelLoading />
    if (ticketsQuery.isError) return <LinkedPanelError onRetry={() => void ticketsQuery.refetch()} />
    if (!allTickets.length) {
      return <LinkedPanelEmpty message={LINKED_TAB_META.tickets.empty} ctaHref={activeRoute} ctaLabel="Open Tickets" />
    }

    return (
      <>
        <ul className="space-y-2">
          {ticketsPageItems.map((ticket) => (
            <li key={ticket.id} className="flex items-center gap-2 rounded-lg px-3 py-2" style={{ background: 'var(--surface-app)' }}>
              <Link to={`/tickets/${ticket.id}`} className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {ticket.subject}
                </p>
                <p className="text-xs uppercase" style={{ color: 'var(--text-label)' }}>
                  {ticket.status}
                </p>
              </Link>
              <button
                onClick={() => ticketMutation.mutate({ ticketId: ticket.id, shouldLink: false, ticket })}
                className="rounded p-1 transition-colors hover:bg-white/70"
                aria-label="Unlink ticket"
              >
                <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
              </button>
            </li>
          ))}
        </ul>
        <PaginationControls page={pages.tickets} totalPages={totalPages('tickets')} onPageChange={(page) => setPageFor('tickets', page)} />
      </>
    )
  }

  const renderSequencesPanel = () => {
    if (sequencesQuery.isLoading) return <LinkedPanelLoading />
    if (sequencesQuery.isError) return <LinkedPanelError onRetry={() => void sequencesQuery.refetch()} />
    if (!allSequences.length) {
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
          totalPages={totalPages('sequences')}
          onPageChange={(page) => setPageFor('sequences', page)}
        />
      </>
    )
  }

  const renderInboxPanel = () => {
    if (inboxQuery.isLoading) return <LinkedPanelLoading />
    if (inboxQuery.isError) return <LinkedPanelError onRetry={() => void inboxQuery.refetch()} />
    if (!allInboxThreads.length) {
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
        <PaginationControls page={pages.inbox} totalPages={totalPages('inbox')} onPageChange={(page) => setPageFor('inbox', page)} />
      </>
    )
  }

  const renderActivePanel = () => {
    switch (activeTab) {
      case 'accounts':
        return renderAccountsPanel()
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

  return (
    <section
      className="rounded-xl border p-4"
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
        {showMutableCtas && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-xs"
            onClick={() => {
              if (activeTab === 'accounts') setShowAccountLinkModal(true)
              if (activeTab === 'deals') setShowDealLinkModal(true)
              if (activeTab === 'tickets') setShowTicketLinkModal(true)
            }}
          >
            <Plus className="h-3 w-3" />
            Link existing
          </Button>
        )}
        {showCreateCta && (
          <Button
            variant="outline"
            size="sm"
            className="h-8 px-2 text-xs"
            onClick={() => {
              if (activeTab === 'accounts') setShowCreateAccount(true)
              if (activeTab === 'deals') setShowCreateDeal(true)
              if (activeTab === 'tickets') setShowCreateTicket(true)
            }}
          >
            <Plus className="h-3 w-3" />
            Create + link
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

      <EntityLinkModal<Account>
        open={showAccountLinkModal}
        onClose={() => setShowAccountLinkModal(false)}
        title="Link Account"
        placeholder="Search accounts by name…"
        search={accountsApi.search}
        onSelect={(account) => accountMutation.mutateAsync({ accountId: account.id, account })}
        renderItem={(account) => (
          <div className="min-w-0">
            <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
              {account.name}
            </p>
            <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
              {account.industry ?? account.domain ?? 'No account details'}
            </p>
          </div>
        )}
        getKey={(account) => account.id}
      />

      <EntityLinkModal<Deal>
        open={showDealLinkModal}
        onClose={() => setShowDealLinkModal(false)}
        title="Link Deal"
        placeholder="Search deals by title…"
        search={dealsApi.search}
        onSelect={(deal) => dealMutation.mutateAsync({ dealId: deal.id, shouldLink: true, deal: { ...deal, contact_id: contactId, contact } })}
        renderItem={(deal) => (
          <div className="min-w-0">
            <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
              {deal.title}
            </p>
            <p className="truncate text-xs" style={{ color: 'var(--text-label)' }}>
              {formatCurrency((deal.value_cents ?? 0) / 100, deal.currency)}
            </p>
          </div>
        )}
        getKey={(deal) => deal.id}
      />

      <EntityLinkModal<TicketType>
        open={showTicketLinkModal}
        onClose={() => setShowTicketLinkModal(false)}
        title="Link Ticket"
        placeholder="Search tickets by subject…"
        search={async (q) => {
          const response = await ticketsApi.list({ search: q, per_page: 10 })
          return response.data ?? []
        }}
        onSelect={(ticket) => ticketMutation.mutateAsync({ ticketId: ticket.id, shouldLink: true, ticket })}
        mapError={(error, ticket) => {
          const mapped = mapCrmLinkError(error)
          if (mapped.kind !== 'account_contact_mismatch') return null

          const ticketAccountId = ticket.account?.id
          const ticketAccountName = ticket.account?.name ?? 'this ticket account'
          return {
            message: `This ticket is scoped to ${ticketAccountName}, but this contact is linked to a different account.`,
            actions: [
              ...(ticketAccountId
                ? [{
                    label: 'Relink contact to account',
                    action: async () => {
                      await contactsApi.update(contactId, { account_id: ticketAccountId })
                      await ticketMutation.mutateAsync({ ticketId: ticket.id, shouldLink: true, ticket })
                    },
                  }]
                : []),
              {
                label: 'Keep ticket account and detach contact',
                action: async () => {
                  await ticketsApi.patchContact(ticket.id, null)
                  await ticketMutation.mutateAsync({ ticketId: ticket.id, shouldLink: true, ticket })
                },
              },
            ],
          }
        }}
        renderItem={(ticket) => (
          <div className="flex items-center justify-between gap-2">
            <p className="truncate text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
              {ticket.subject}
            </p>
            <span className="text-xs uppercase" style={{ color: 'var(--text-label)' }}>
              {ticket.status}
            </span>
          </div>
        )}
        getKey={(ticket) => ticket.id}
      />

      {showCreateAccount && (
        <AccountForm
          onClose={() => setShowCreateAccount(false)}
          initialValues={
            contact.email.includes('@')
              ? { domain: contact.email.split('@')[1] }
              : undefined
          }
          onCreated={(account) => {
            void accountMutation.mutateAsync({ accountId: account.id, account })
          }}
        />
      )}

      {showCreateDeal && (
        <DealForm
          onClose={() => setShowCreateDeal(false)}
          initialValues={{
            contact_id: contactId,
            account_id: linkedAccounts.length === 1 ? primaryLinkedAccount?.id : undefined,
          }}
          onCreated={(deal) => {
            queryClient.setQueryData<PaginatedResponse<Deal> | undefined>(
              dealKeys.list(dealsParams),
              (current) =>
                current
                  ? {
                      ...current,
                      data: [deal, ...(current.data ?? []).filter((item) => item.id !== deal.id)],
                      meta: {
                        ...current.meta,
                        total: current.meta.total + 1,
                        total_pages: Math.max(1, Math.ceil((current.meta.total + 1) / current.meta.per_page)),
                      },
                    }
                  : current
            )
            toast({ title: 'Deal created and linked' })
          }}
        />
      )}

      {showCreateTicket && (
        <TicketForm
          onClose={() => setShowCreateTicket(false)}
          initialValues={{
            contact_id: contactId,
            account_id: linkedAccounts.length === 1 ? primaryLinkedAccount?.id : undefined,
          }}
          onCreated={(ticketId) => {
            void queryClient.invalidateQueries({ queryKey: ticketKeys.list(ticketsParams) })
            void queryClient.invalidateQueries({ queryKey: ticketKeys.detail(ticketId) })
            toast({ title: 'Ticket created and linked' })
          }}
        />
      )}
    </section>
  )
}

// ── Private Notes ─────────────────────────────────────────────────────────────

function PrivateNotesPanel({ contactId, contactName }: { contactId: string; contactName: string }) {
  const { data: notes } = useContactNotes(contactId)
  const addNote = useAddContactNote()

  const latestContent = notes?.[0]?.content ?? ''
  const [text, setText] = useState('')
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle')
  const initializedRef = useRef(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Pre-fill with latest note on load (only once)
  useEffect(() => {
    if (!initializedRef.current && notes !== undefined) {
      setText(latestContent)
      initializedRef.current = true
    }
  }, [notes, latestContent])

  const save = useCallback(
    async (content: string) => {
      if (!content.trim()) return
      setSaveStatus('saving')
      try {
        await addNote.mutateAsync({ contactId, payload: { content } })
        setSaveStatus('saved')
        setTimeout(() => setSaveStatus('idle'), 1500)
      } catch {
        setSaveStatus('error')
      }
    },
    [addNote, contactId]
  )

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value)
    setSaveStatus('idle')
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => save(e.target.value), 1000)
  }

  return (
    <section
      className="rounded-xl border p-4"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      aria-labelledby="notes-heading"
    >
      <div className="flex items-center gap-1.5 mb-3">
        <Lock className="h-3 w-3" style={{ color: 'var(--text-label)' }} aria-hidden="true" />
        <p
          id="notes-heading"
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          title="Only visible to your team"
        >
          Private Notes
        </p>
      </div>
      <textarea
        className="w-full rounded border px-3 py-2.5 text-sm resize-y transition-colors focus:outline-none"
        style={{
          minHeight: 100,
          maxHeight: 280,
          borderColor: 'var(--border-default)',
          color: 'var(--text-primary)',
          lineHeight: 1.6,
          fontSize: 14,
        }}
        placeholder="Add a private note about this contact…"
        value={text}
        onChange={handleChange}
        aria-label={`Private notes about ${contactName}`}
        aria-describedby="notes-save-status"
        onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--border-focus)' }}
        onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)' }}
      />
      <div className="flex justify-end mt-1.5">
        <span
          id="notes-save-status"
          aria-live="polite"
          className="text-xs italic"
          style={{
            color:
              saveStatus === 'saving'
                ? 'var(--text-label)'
                : saveStatus === 'saved'
                ? 'var(--color-success)'
                : saveStatus === 'error'
                ? 'var(--color-danger)'
                : 'transparent',
          }}
        >
          {saveStatus === 'saving'
            ? 'Saving…'
            : saveStatus === 'saved'
            ? 'Saved ✓'
            : saveStatus === 'error'
            ? 'Save failed — Retry'
            : '.'}
        </span>
      </div>
    </section>
  )
}

// ── Enrichment Panel ──────────────────────────────────────────────────────────

type EnrichStatus = 'idle' | 'enriched' | 'failed'

function EnrichmentStatusBadge({ status }: { status: EnrichStatus }) {
  if (status === 'idle') {
    return (
      <span
        className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
        style={{ background: '#F1F5F9', color: 'var(--text-label)' }}
      >
        Not enriched
      </span>
    )
  }
  if (status === 'enriched') {
    return (
      <span
        className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
        style={{ background: '#F0FDF4', color: '#16A34A' }}
      >
        <Sparkles className="h-3 w-3" />
        Enriched
      </span>
    )
  }
  return (
    <span
      className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
      style={{ background: '#FEF2F2', color: 'var(--color-danger)' }}
    >
      Failed
    </span>
  )
}

function EnrichmentPanel({ contactId }: { contactId: string }) {
  const enrichContact = useEnrichContact()
  const [status, setStatus] = useState<EnrichStatus>('idle')
  const [result, setResult] = useState<EnrichmentResult | null>(null)

  const handleEnrich = async () => {
    try {
      const data = await enrichContact.mutateAsync(contactId)
      setResult(data)
      setStatus('enriched')
    } catch {
      setStatus('failed')
    }
  }

  return (
    <section
      className="rounded-xl border p-4 mt-4"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      aria-labelledby="enrichment-heading"
    >
      <div className="flex items-center justify-between gap-2 mb-3">
        <div className="flex items-center gap-2">
          <p
            id="enrichment-heading"
            className="text-[11px] font-semibold uppercase tracking-widest"
            style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          >
            Enrichment
          </p>
          <EnrichmentStatusBadge status={status} />
        </div>
        <button
          type="button"
          onClick={handleEnrich}
          disabled={enrichContact.isPending}
          className="flex items-center gap-1.5 px-3 rounded-md text-xs font-medium transition-colors whitespace-nowrap"
          style={{
            height: 28,
            border: '1px solid var(--border-default)',
            background: 'var(--surface-card)',
            color: 'var(--text-primary)',
            cursor: enrichContact.isPending ? 'default' : 'pointer',
            opacity: enrichContact.isPending ? 0.7 : 1,
          }}
          aria-label={status === 'enriched' ? 'Re-enrich contact' : 'Enrich contact'}
        >
          {enrichContact.isPending ? (
            <Loader2 className="h-3 w-3 animate-spin" />
          ) : status === 'enriched' ? (
            <RefreshCw className="h-3 w-3" />
          ) : (
            <Sparkles className="h-3 w-3" />
          )}
          {status === 'enriched' ? 'Re-enrich' : 'Enrich contact'}
        </button>
      </div>

      {status === 'failed' && (
        <p className="text-xs" style={{ color: 'var(--color-danger)' }}>
          Enrichment failed — no data found for this email domain.
        </p>
      )}

      {status === 'enriched' && result?.data && (
        <dl className="space-y-1.5">
          {result.data.company_name && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Company</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.company_name}</dd>
            </div>
          )}
          {result.data.industry && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Industry</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.industry}</dd>
            </div>
          )}
          {result.data.size && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Size</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.size}</dd>
            </div>
          )}
          {result.data.linkedin_url && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>LinkedIn</dt>
              <dd className="text-xs">
                <a
                  href={result.data.linkedin_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 hover:underline"
                  style={{ color: 'var(--color-primary)' }}
                >
                  View profile
                  <ExternalLink className="h-3 w-3" />
                </a>
              </dd>
            </div>
          )}
        </dl>
      )}

      {status === 'idle' && (
        <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
          Click Enrich to auto-populate company data from this contact's email domain.
        </p>
      )}
    </section>
  )
}

// ── Edit Contact Modal ────────────────────────────────────────────────────────

function EditContactModal({
  contact,
  onClose,
}: {
  contact: Contact
  onClose: () => void
}) {
  const updateContact = useUpdateContact()
  const [form, setForm] = useState({
    first_name: contact.first_name,
    last_name: contact.last_name,
    email: contact.email,
    phone: contact.phone ?? '',
    title: contact.title ?? '',
  })

  const set = (field: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [field]: e.target.value }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await updateContact.mutateAsync({
      id: contact.id,
      payload: {
        first_name: form.first_name,
        last_name: form.last_name,
        email: form.email,
        phone: form.phone || undefined,
        title: form.title || undefined,
      },
    })
    onClose()
  }

  return (
    <Dialog open onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Contact</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                First name
              </label>
              <Input value={form.first_name} onChange={set('first_name')} required />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Last name
              </label>
              <Input value={form.last_name} onChange={set('last_name')} required />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
              Email
            </label>
            <Input type="email" value={form.email} onChange={set('email')} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Title
              </label>
              <Input value={form.title} onChange={set('title')} placeholder="Senior Developer" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Phone
              </label>
              <Input type="tel" value={form.phone} onChange={set('phone')} placeholder="+1 555 000 0000" />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateContact.isPending}>
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Delete Confirm Modal ──────────────────────────────────────────────────────

function DeleteConfirmModal({
  name,
  onConfirm,
  onCancel,
  isPending,
}: {
  name: string
  onConfirm: () => void
  onCancel: () => void
  isPending: boolean
}) {
  return (
    <Dialog open onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Delete Contact?</DialogTitle>
        </DialogHeader>
        <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
          This will permanently remove{' '}
          <strong style={{ color: 'var(--text-primary)' }}>{name}</strong> and all associated
          data. This action cannot be undone.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button
            disabled={isPending}
            onClick={onConfirm}
            style={{ background: 'var(--color-danger)', color: '#FFFFFF', border: 'none' }}
          >
            Delete Contact
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function ContactDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [composeOpen, setComposeOpen] = useState(false)

  const { data: contact, isLoading, isError } = useContact(id!)
  const deleteContact = useDeleteContact()
  const [relationshipRows, setRelationshipRows] = useState<RelationshipRow[]>([])
  const seedContactRecord = (contact as ContactRecord | undefined) ?? null
  const seedContactId = seedContactRecord?.id ?? null
  const linkedAccounts = getLinkedAccounts(seedContactRecord)
  const relationshipSeed = useMemo(
    () => createContactRelationshipRows(linkedAccounts),
    [linkedAccounts]
  )
  const relationshipSeedKey = useMemo(
    () => relationshipSeed.map((row) => `${row.entityId ?? row.id}:${row.label}:${row.meta}`).join('|'),
    [relationshipSeed]
  )
  const previousRelationshipSeedKeyRef = useRef<string | null>(null)
  const previousRelationshipContactIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!seedContactId) return

    const isNewContact = previousRelationshipContactIdRef.current !== seedContactId

    setRelationshipRows((currentRows) => {
      const currentRowsKey = currentRows
        .map((row) => `${row.entityId ?? row.id}:${row.label}:${row.meta}`)
        .join('|')
      const canReseed =
        isNewContact ||
        currentRows.length === 0 ||
        currentRowsKey === previousRelationshipSeedKeyRef.current

      return canReseed ? relationshipSeed : currentRows
    })

    previousRelationshipContactIdRef.current = seedContactId
    previousRelationshipSeedKeyRef.current = relationshipSeedKey
  }, [seedContactId, relationshipSeed, relationshipSeedKey])

  if (isLoading) {
    return (
      <div className="p-6 max-w-[var(--content-max-width,1280px)]">
        {/* Breadcrumb skeleton */}
        <div className="h-4 w-48 bg-slate-100 animate-pulse rounded mb-5" />
        {/* Header card skeleton */}
        <div
          className="rounded-xl border p-5 flex items-center gap-4 mb-5"
          style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        >
          <div className="h-[72px] w-[72px] rounded-full bg-slate-100 animate-pulse shrink-0" />
          <div className="space-y-2 flex-1">
            <div className="h-6 w-56 bg-slate-100 animate-pulse rounded" />
            <div className="h-3.5 w-36 bg-slate-100 animate-pulse rounded" />
          </div>
        </div>
        {/* Body skeletons */}
        <div className="flex gap-5">
          <div className="flex-1 space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-16 bg-slate-100 animate-pulse rounded-xl" />
            ))}
          </div>
          <div className="w-[320px] shrink-0 space-y-3">
            <div className="h-32 bg-slate-100 animate-pulse rounded-xl" />
            <div className="h-24 bg-slate-100 animate-pulse rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !contact) {
    return (
      <div className="p-6 flex flex-col items-center justify-center py-24 text-center">
        <p className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
          Contact not found.
        </p>
        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          This contact may have been deleted or you may not have access.
        </p>
        <Link
          to="/contacts"
          className="text-sm font-medium"
          style={{ color: 'var(--color-primary)' }}
        >
          ← Back to Contacts
        </Link>
      </div>
    )
  }

  const contactRecord = contact as ContactRecord
  const primaryRelationship = relationshipRows.find((row) => row.isPrimary) ?? relationshipRows[0]
  const primaryLinkedAccount = linkedAccounts.find((account) => account.id === primaryRelationship?.entityId)
  const fullName = `${contact.first_name} ${contact.last_name}`
  const initials = contactInitials(contact.first_name, contact.last_name)

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 mb-5 text-sm" style={{ color: 'var(--text-label)' }}>
        <Link
          to="/contacts"
          className="hover:underline"
          style={{ color: 'var(--color-primary)' }}
        >
          Contacts
        </Link>
        <ChevronRight className="h-3.5 w-3.5" />
        <span style={{ color: 'var(--text-primary)' }}>{fullName}</span>
      </nav>

      {/* Header Card */}
      <div
        className="rounded-xl border p-5 mb-5 flex flex-wrap items-center gap-4"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      >
        {/* Avatar */}
        <div
          className="h-[72px] w-[72px] rounded-full flex items-center justify-center shrink-0 text-2xl font-bold"
          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
        >
          {initials}
        </div>

        {/* Identity */}
        <div className="flex-1 min-w-0">
          <h1
            className="font-bold leading-tight"
            style={{ fontSize: 24, color: 'var(--text-primary)', letterSpacing: '-0.01em' }}
          >
            {fullName}
          </h1>
          {contact.title && (
            <p className="mt-0.5 text-sm" style={{ color: 'var(--text-secondary)' }}>
              {contact.title}
            </p>
          )}
          {primaryRelationship ? (
            primaryLinkedAccount ? (
              <Link
                to={`/accounts/${primaryLinkedAccount.id}`}
                className="mt-0.5 inline-flex items-center gap-1 text-sm font-medium hover:underline"
                style={{ color: 'var(--color-primary)' }}
                aria-label={`${primaryRelationship.label} (opens account detail)`}
              >
                {primaryRelationship.label}
                <ExternalLink className="h-3 w-3" aria-hidden="true" />
              </Link>
            ) : (
              <p className="mt-0.5 text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                {primaryRelationship.label}
              </p>
            )
          ) : (
            <p className="mt-0.5 text-sm italic" style={{ color: 'var(--text-label)' }}>
              No accounts linked
            </p>
          )}
          {relationshipRows.length > 1 && (
            <p className="mt-1 text-xs font-medium" style={{ color: 'var(--text-label)' }}>
              +{relationshipRows.length - 1} more linked
            </p>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => navigate(`/search?q=${encodeURIComponent(fullName)}&contact_id=${contact.id}&contact_name=${encodeURIComponent(fullName)}`)}
            className="flex items-center gap-1.5 px-4 rounded-md text-sm font-medium transition-colors"
            style={{
              height: 36,
              border: '1px solid var(--border-default)',
              background: 'var(--surface-card)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-app)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-card)' }}
          >
            <Search className="h-3.5 w-3.5" />
            Search related
          </button>
          <button
            onClick={() => setShowEdit(true)}
            className="flex items-center gap-1.5 px-4 rounded-md text-sm font-medium transition-colors"
            style={{
              height: 36,
              border: '1px solid var(--border-default)',
              background: 'var(--surface-card)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-app)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-card)' }}
            aria-label={`Edit contact: ${fullName}`}
          >
            <Pencil className="h-3.5 w-3.5" />
            Edit
          </button>
          <button
            onClick={() => setComposeOpen(true)}
            className="flex items-center gap-1.5 px-4 rounded-md text-sm font-semibold text-white transition-colors"
            style={{
              height: 36,
              background: 'var(--color-primary)',
              border: 'none',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary-hover)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary)' }}
            aria-label={`Send email to ${fullName}`}
          >
            <Mail className="h-3.5 w-3.5" />
            Send Email
          </button>
        </div>
      </div>

      {/* Body — 2-col on desktop, stack on mobile */}
      <div className="flex flex-col md:flex-row gap-5">
        {/* Right column on mobile (linked entities + notes + enrichment) */}
        <div className="md:hidden flex flex-col gap-4">
          <RelationshipEditor
            title="Relationship Roles"
            entityLabel="Account"
            value={relationshipRows}
            onChange={setRelationshipRows}
            emptyMessage="No account relationships yet."
            addLabel="Add account role"
          />
          <LinkedEntitiesSection contact={contactRecord} contactId={id!} contactName={fullName} />
          <PrivateNotesPanel contactId={id!} contactName={fullName} />
          <EnrichmentPanel contactId={id!} />
        </div>

        {/* Left — Timeline */}
        <div className="flex-1 min-w-0">
          <UnifiedTimeline entityType="contact" entityId={id!} />
        </div>

        {/* Right — Desktop only */}
        <div className="hidden md:block w-[320px] shrink-0">
          <RelationshipEditor
            title="Relationship Roles"
            entityLabel="Account"
            value={relationshipRows}
            onChange={setRelationshipRows}
            emptyMessage="No account relationships yet."
            addLabel="Add account role"
          />
          <LinkedEntitiesSection contact={contactRecord} contactId={id!} contactName={fullName} />
          <PrivateNotesPanel contactId={id!} contactName={fullName} />
          <EnrichmentPanel contactId={id!} />
        </div>
      </div>

      {/* Record Footer */}
      <footer
        className="mt-6 pt-4 border-t flex flex-wrap items-center justify-between gap-4"
        style={{ borderColor: 'var(--border-subtle)' }}
      >
        <p className="text-xs" style={{ color: 'var(--text-label)' }}>
          <span title={contact.id}>ID: {contact.id.slice(0, 8)}…</span>
          {' · '}
          Created: {formatDate(contact.created_at)}
          {' · '}
          Modified: {formatDate(contact.updated_at)}
        </p>
        <div className="flex items-center gap-2">
          <button
            onClick={() => window.print()}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--text-secondary)',
              borderColor: 'var(--border-default)',
              background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <Download className="h-3 w-3" />
            Export PDF
          </button>
          <button
            onClick={() => navigate(`/admin/audit?entityType=contact`)}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--text-secondary)',
              borderColor: 'var(--border-default)',
              background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <ClipboardList className="h-3 w-3" />
            Audit Log
          </button>
          <button
            onClick={() => setShowDelete(true)}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--color-danger)',
              borderColor: '#FCA5A5',
              background: 'transparent',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = '#FEF2F2' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent' }}
          >
            <Trash2 className="h-3 w-3" />
            Delete
          </button>
        </div>
      </footer>

      {/* Edit modal */}
      {showEdit && (
        <EditContactModal contact={contact} onClose={() => setShowEdit(false)} />
      )}

      {/* Delete confirm */}
      {showDelete && (
        <DeleteConfirmModal
          name={fullName}
          isPending={deleteContact.isPending}
          onConfirm={async () => {
            await deleteContact.mutateAsync(id!)
            navigate('/contacts')
          }}
          onCancel={() => setShowDelete(false)}
        />
      )}

      <ComposeEmailModal
        open={composeOpen}
        onOpenChange={setComposeOpen}
        contactId={id!}
        toEmail={contact.email ?? ''}
      />
    </div>
  )
}

