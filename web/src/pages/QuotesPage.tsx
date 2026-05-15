import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Plus, FileText } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { Table, type Column } from '@/components/ui/Table'
import { SidePanel } from '@/components/ui/SidePanel'
import { CustomFieldDisplaySection } from '@/components/omnir/CustomFieldRenderer'
import { QuoteBuilder, QuoteStatusBadge } from '@/components/omnir/QuoteBuilder'
import { ViewPinBar } from '@/components/omnir/ViewPinBar'
import { useAccounts } from '@/hooks/useAccounts'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { useQuotes, useDeleteQuote } from '@/hooks/useQuotes'
import { useUpdateView } from '@/hooks/useViews'
import { formatCurrency, formatDate } from '@/lib/utils'
import { cleanCurrentFilters, pickViewFilters, sortKeyFromFilters, stringFilter } from '@/lib/savedViewFilters'
import type { Quote, QuoteStatus, SavedView, CustomFieldValues } from '@/api/types'

const STATUS_OPTIONS: { label: string; value: QuoteStatus | '' }[] = [
  { label: 'All statuses', value: '' },
  { label: 'Draft', value: 'draft' },
  { label: 'Sent', value: 'sent' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' },
  { label: 'Expired', value: 'expired' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
  { label: 'Title A-Z', value: 'title:asc' },
  { label: 'Total High-Low', value: 'total_cents:desc' },
  { label: 'Valid Until', value: 'valid_until:asc' },
]

const QUOTE_VIEW_FILTER_KEYS = ['q', 'search', 'status', 'account_id', 'contact_id', 'deal_id', 'sort_by', 'sort_dir'] as const

function AccountChip({ name }: { name?: string }) {
  if (!name) return null
  return <Badge variant="indigo" className="gap-1">{name}</Badge>
}

function QuoteDetail({
  quote,
  accountName,
  lockedAccountId,
  lockedAccountName,
  onClose,
}: {
  quote: Quote
  accountName?: string
  lockedAccountId?: string
  lockedAccountName?: string
  onClose: () => void
}) {
  const deleteQuote = useDeleteQuote()
  const [showEdit, setShowEdit] = useState(false)
  const { data: customFieldDefs = [] } = useCustomFieldDefinitions('quote', { activeOptionsOnly: true })
  const displayAccountName = quote.account?.name ?? accountName

  return (
    <>
      <SidePanel open title={quote.title} onClose={onClose} width="lg"
        actions={
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setShowEdit(true)}>Edit</Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={async () => {
                if (confirm('Delete this quote?')) {
                  await deleteQuote.mutateAsync(quote.id)
                  onClose()
                }
              }}
            >
              Delete
            </Button>
          </div>
        }
      >
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-3">
            <QuoteStatusBadge status={quote.status} />
            <AccountChip name={displayAccountName} />
            <span className="text-2xl font-bold text-[var(--color-primary)]">
              {formatCurrency(quote.total_cents / 100, quote.currency)}
            </span>
          </div>

          <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
            <dt className="font-medium text-slate-500">Currency</dt>
            <dd className="text-slate-900">{quote.currency}</dd>

            {displayAccountName && (
              <>
                <dt className="font-medium text-slate-500">Account</dt>
                <dd className="text-slate-900">{displayAccountName}</dd>
              </>
            )}

            {quote.valid_until && (
              <>
                <dt className="font-medium text-slate-500">Valid until</dt>
                <dd className="text-slate-900">{formatDate(quote.valid_until)}</dd>
              </>
            )}
            {quote.sent_at && (
              <>
                <dt className="font-medium text-slate-500">Sent at</dt>
                <dd className="text-slate-900">{formatDate(quote.sent_at)}</dd>
              </>
            )}
            {quote.approved_at && (
              <>
                <dt className="font-medium text-slate-500">Approved at</dt>
                <dd className="text-green-600">{formatDate(quote.approved_at)}</dd>
              </>
            )}
            {quote.rejected_at && (
              <>
                <dt className="font-medium text-slate-500">Rejected at</dt>
                <dd className="text-red-600">{formatDate(quote.rejected_at)}</dd>
              </>
            )}
            <dt className="font-medium text-slate-500">Created</dt>
            <dd className="text-slate-900">{formatDate(quote.created_at)}</dd>
          </dl>

          {/* Line items */}
          {quote.line_items?.length > 0 && (
            <div>
              <p className="mb-2 text-sm font-semibold text-slate-700">Line Items</p>
              <div className="space-y-2">
                {quote.line_items.map((li) => (
                  <div
                    key={li.id}
                    className="flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 text-sm"
                  >
                    <div>
                      <p className="font-medium text-slate-900">{li.product_name}</p>
                      <p className="text-slate-500 text-xs">
                        {li.quantity} × {formatCurrency(li.unit_price_cents / 100, quote.currency)}
                        {li.discount_pct > 0 && ` − ${li.discount_pct}%`}
                      </p>
                    </div>
                    <span className="font-semibold text-slate-900">
                      {formatCurrency(li.total_cents / 100, quote.currency)}
                    </span>
                  </div>
                ))}
              </div>
              <div className="mt-3 flex justify-end">
                <div className="text-right">
                  <div className="flex gap-8 font-bold text-slate-900 text-sm border-t border-slate-200 pt-2">
                    <span>Total</span>
                    <span>{formatCurrency(quote.total_cents / 100, quote.currency)}</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {quote.notes && (
            <div>
              <p className="mb-1 text-sm font-medium text-slate-500">Notes</p>
              <p className="text-sm text-slate-900 whitespace-pre-line">{quote.notes}</p>
            </div>
          )}

          <CustomFieldDisplaySection
            fields={customFieldDefs}
            values={quote.custom_fields as CustomFieldValues | undefined}
          />
        </div>
      </SidePanel>

      {showEdit && (
        <QuoteBuilder
          quote={quote}
          accountId={lockedAccountId}
          accountName={lockedAccountName}
          lockAccount={!!lockedAccountId}
          onClose={() => setShowEdit(false)}
        />
      )}
    </>
  )
}

export function QuotesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const accountId = searchParams.get('account_id') ?? ''
  const accountName = searchParams.get('account_name') ?? ''
  const dealId = searchParams.get('deal_id') ?? ''
  const dealName = searchParams.get('deal_name') ?? ''
  const contactId = searchParams.get('contact_id') ?? ''
  const contactName = searchParams.get('contact_name') ?? ''
  const [statusFilter, setStatusFilter] = useState<QuoteStatus | ''>('')
  const [search, setSearch] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [selectedQuoteId, setSelectedQuoteId] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [activeView, setActiveView] = useState<SavedView | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const updateView = useUpdateView()
  const { data: accountsResult, isLoading: accountsLoading } = useAccounts({ per_page: 200 })
  const accounts = useMemo(() => accountsResult?.data ?? [], [accountsResult?.data])
  const accountNameById = useMemo(
    () => new Map(accounts.map((account) => [account.id, account.name])),
    [accounts]
  )

  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']
  const currentFilters = cleanCurrentFilters({
    search: search || undefined,
    q: search || undefined,
    status: statusFilter || undefined,
    account_id: accountId || undefined,
    contact_id: contactId || undefined,
    deal_id: dealId || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })
  const markChanged = () => { if (activeView) setHasUnsavedChanges(true) }

  const applyViewFilters = (view: SavedView) => {
    const filters = pickViewFilters(view.filters, QUOTE_VIEW_FILTER_KEYS)
    setActiveView(view)
    setHasUnsavedChanges(false)
    setSearch(stringFilter(filters, 'search') || stringFilter(filters, 'q'))
    setStatusFilter(stringFilter(filters, 'status') as QuoteStatus | '')
    setSortKey(sortKeyFromFilters(filters, 'created_at:desc'))
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      const nextAccountId = stringFilter(filters, 'account_id')
      const nextContactId = stringFilter(filters, 'contact_id')
      const nextDealId = stringFilter(filters, 'deal_id')
      nextAccountId ? next.set('account_id', nextAccountId) : next.delete('account_id')
      nextContactId ? next.set('contact_id', nextContactId) : next.delete('contact_id')
      nextDealId ? next.set('deal_id', nextDealId) : next.delete('deal_id')
      next.delete('account_name')
      next.delete('contact_name')
      next.delete('deal_name')
      return next
    })
  }

  const handleUpdateView = async (viewId: string) => {
    await updateView.mutateAsync({ id: viewId, payload: { filters: currentFilters } })
    setHasUnsavedChanges(false)
  }

  const { data, isLoading } = useQuotes({
    q: search || undefined,
    status: statusFilter || undefined,
    account_id: accountId || undefined,
    deal_id: dealId || undefined,
    contact_id: contactId || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
    limit: 50,
  })

  const quotes = data?.data ?? []
  const activeContextLabel = useMemo(() => {
    if (dealId) return dealName ? `Deal: ${dealName}` : 'Deal filter active'
    if (contactId) return contactName ? `Contact: ${contactName}` : 'Contact filter active'
    if (accountId) return accountName ? `Account: ${accountName}` : 'Account filter active'
    return null
  }, [accountId, accountName, contactId, contactName, dealId, dealName])

  const selectedQuote = selectedQuoteId ? quotes.find((q) => q.id === selectedQuoteId) : null
  const selectedAccountName = accountId ? (accountNameById.get(accountId) ?? accountName) : ''
  const getQuoteAccountName = (quote: Quote) =>
    quote.account?.name ??
    (quote.account_id ? accountNameById.get(quote.account_id) : undefined) ??
    (quote.account_id === accountId ? accountName : undefined)

  const columns: Column<Quote>[] = [
    {
      key: 'title',
      header: 'Quote',
      render: (q) => (
        <div className="flex items-start gap-2">
          <FileText className="mt-0.5 h-4 w-4 text-slate-400 shrink-0" />
          <div className="min-w-0 space-y-1">
            <span className="block truncate font-medium text-slate-900">{q.title}</span>
            <AccountChip name={getQuoteAccountName(q)} />
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (q) => <QuoteStatusBadge status={q.status} />,
    },
    {
      key: 'total_cents',
      header: 'Total',
      render: (q) => (
        <span className="font-semibold text-[var(--color-primary)]">
          {formatCurrency(q.total_cents / 100, q.currency)}
        </span>
      ),
    },
    {
      key: 'valid_until',
      header: 'Valid until',
      render: (q) => q.valid_until ? formatDate(q.valid_until) : '—',
    },
    {
      key: 'created_at',
      header: 'Created',
      render: (q) => formatDate(q.created_at),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Quotes</h1>
          <p className="text-sm text-slate-500 mt-0.5">{data?.meta?.total ?? 0} total</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4 mr-1.5" /> New quote
        </Button>
      </div>

      <ViewPinBar
        entityType="quotes"
        activeViewId={activeView?.id ?? null}
        hasUnsavedChanges={hasUnsavedChanges}
        currentFilters={currentFilters}
        onSelectView={applyViewFilters}
        onClearView={() => { setActiveView(null); setHasUnsavedChanges(false) }}
        onViewSaved={(view) => setActiveView(view)}
        onUpdateView={handleUpdateView}
      />

      {activeContextLabel && (
        <div className="flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm">
          <span className="font-medium text-slate-700">{activeContextLabel}</span>
          <button
            onClick={() =>
              setSearchParams((prev) => {
                const next = new URLSearchParams(prev)
                next.delete('account_id')
                next.delete('account_name')
                next.delete('deal_id')
                next.delete('deal_name')
                next.delete('contact_id')
                next.delete('contact_name')
                return next
              })
            }
            className="text-slate-500 hover:text-slate-900"
          >
            Clear
          </button>
          <Link to="/accounts" className="ml-auto text-[var(--color-primary)] hover:underline">
            Browse accounts
          </Link>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <input
          type="text"
          value={search}
          onChange={(e) => { setSearch(e.target.value); markChanged() }}
          placeholder="Search quotes…"
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm w-56 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        />
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value as QuoteStatus | ''); markChanged() }}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <select
          value={accountId}
          onChange={(e) => {
            const nextAccountId = e.target.value
            setSearchParams((prev) => {
              const next = new URLSearchParams(prev)
              if (nextAccountId) {
                next.set('account_id', nextAccountId)
                next.set('account_name', accountNameById.get(nextAccountId) ?? '')
              } else {
                next.delete('account_id')
                next.delete('account_name')
              }
              return next
            })
            markChanged()
          }}
          disabled={accountsLoading}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        >
          <option value="">All accounts</option>
          {accounts.map((account) => (
            <option key={account.id} value={account.id}>{account.name}</option>
          ))}
        </select>
        <select
          value={sortKey}
          onChange={(e) => { setSortKey(e.target.value); markChanged() }}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        >
          {SORT_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="flex justify-center py-12">
          <Spinner size="lg" />
        </div>
      ) : (
        <Table<Quote>
          columns={columns}
          data={quotes}
          keyExtractor={(q) => q.id}
          onRowClick={(q) => setSelectedQuoteId(q.id)}
          emptyTitle="No quotes yet"
          emptyDescription="Create one to get started."
        />
      )}

      {/* Detail panel */}
      {selectedQuote && (
        <QuoteDetail
          quote={selectedQuote}
          accountName={getQuoteAccountName(selectedQuote)}
          lockedAccountId={accountId || undefined}
          lockedAccountName={selectedAccountName || undefined}
          onClose={() => setSelectedQuoteId(null)}
        />
      )}

      {/* Create quote modal */}
      {showCreate && (
        <QuoteBuilder
          accountId={accountId || undefined}
          accountName={selectedAccountName || undefined}
          lockAccount={!!accountId}
          dealId={dealId || undefined}
          contactId={contactId || undefined}
          onClose={() => setShowCreate(false)}
        />
      )}
    </div>
  )
}
