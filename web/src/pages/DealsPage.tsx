import { useState, useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus, LayoutGrid, List, X, TrendingUp } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useDeals, useDeal, useDeleteDeal, useUpdateDeal } from '@/hooks/useDeals'
import { KanbanBoard } from '@/components/omnir/KanbanBoard'
import { ActivityTimeline } from '@/components/omnir/ActivityTimeline'
import { CustomFieldEditableSection } from '@/components/omnir/CustomFieldRenderer'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Table, type Column } from '@/components/ui/Table'
import { AttachmentsPanel } from '@/components/omnir/AttachmentsPanel'
import { DealForm } from '@/components/omnir/DealForm'
import { QuoteBuilder, QuoteStatusBadge } from '@/components/omnir/QuoteBuilder'
import { useDealQuotes } from '@/hooks/useQuotes'
import { formatDate, formatCurrency } from '@/lib/utils'
import type { Deal, DealStage, CustomFieldValues } from '@/api/types'
import { FileText } from 'lucide-react'

// ---- Constants ----

const STAGE_OPTIONS = [
  { label: 'Lead', value: 'lead' },
  { label: 'Qualified', value: 'qualified' },
  { label: 'Proposal', value: 'proposal' },
  { label: 'Negotiation', value: 'negotiation' },
  { label: 'Closed Won', value: 'closed_won' },
  { label: 'Closed Lost', value: 'closed_lost' },
]

const PRIORITY_OPTIONS = [
  { label: 'High', value: 'high' },
  { label: 'Medium', value: 'medium' },
  { label: 'Low', value: 'low' },
]

const stageBadge: Record<DealStage, 'blue' | 'indigo' | 'purple' | 'yellow' | 'green' | 'red'> = {
  lead: 'blue',
  qualified: 'indigo',
  proposal: 'purple',
  negotiation: 'yellow',
  closed_won: 'green',
  closed_lost: 'red',
}

const stageLabel: Record<DealStage, string> = {
  lead: 'Lead',
  qualified: 'Qualified',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  closed_won: 'Closed Won',
  closed_lost: 'Closed Lost',
}

// ---- Active filter chips ----

interface FilterChip {
  label: string
  key: string
}

// ---- Deal detail panel ----

function DealDetail({ dealId, onClose }: { dealId: string; onClose: () => void }) {
  const { data: deal, isLoading } = useDeal(dealId)
  const deleteDeal = useDeleteDeal()
  const updateDeal = useUpdateDeal()
  const { data: customFields = [] } = useCustomFieldDefinitions('deal')

  if (isLoading) {
    return (
      <SidePanel open title="Deal" onClose={onClose}>
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      </SidePanel>
    )
  }

  if (!deal) return null

  return (
    <SidePanel
      open
      title={deal.title}
      onClose={onClose}
      width="lg"
      actions={
        <Button
          variant="destructive"
          size="sm"
          onClick={async () => {
            if (confirm('Delete this deal?')) {
              await deleteDeal.mutateAsync(deal.id)
              onClose()
            }
          }}
        >
          Delete
        </Button>
      }
    >
      <div className="space-y-5">
        <div className="flex items-center gap-3">
          <Badge variant={stageBadge[deal.stage]} className="text-sm px-3 py-1">
            {stageLabel[deal.stage]}
          </Badge>
          <span className="text-2xl font-bold text-[#1B3A4B]">
            {formatCurrency(deal.value_cents / 100, deal.currency)}
          </span>
        </div>

        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-3 text-sm">
          {deal.probability !== undefined && (
            <>
              <dt className="font-medium text-[#6B7280]">Probability</dt>
              <dd className="text-[#1A1D23]">{deal.probability}%</dd>
            </>
          )}
          {deal.expected_close_date && (
            <>
              <dt className="font-medium text-[#6B7280]">Close Date</dt>
              <dd className="text-[#1A1D23]">{formatDate(deal.expected_close_date)}</dd>
            </>
          )}
          {deal.account?.name && (
            <>
              <dt className="font-medium text-[#6B7280]">Account</dt>
              <dd className="text-[#1A1D23]">{deal.account.name}</dd>
            </>
          )}
          {deal.contact && (
            <>
              <dt className="font-medium text-[#6B7280]">Contact</dt>
              <dd className="text-[#1A1D23]">
                {deal.contact.first_name} {deal.contact.last_name}
              </dd>
            </>
          )}
          <dt className="font-medium text-[#6B7280]">Created</dt>
          <dd className="text-[#1A1D23]">{formatDate(deal.created_at)}</dd>
        </dl>

        {deal.tags && deal.tags.length > 0 && (
          <div>
            <p className="mb-2 text-sm font-medium text-[#6B7280]">Tags</p>
            <div className="flex flex-wrap gap-1.5">
              {deal.tags.map((tag) => (
                <Badge key={tag} variant="default">{tag}</Badge>
              ))}
            </div>
          </div>
        )}

        <CustomFieldEditableSection
          fields={customFields}
          values={deal.custom_fields as CustomFieldValues | undefined}
          onSave={async (cf) => {
            await updateDeal.mutateAsync({ id: dealId, payload: { custom_fields: cf as Record<string, unknown> } })
          }}
        />

        <ActivityTimeline dealId={deal.id} />
        <DealQuotesSection dealId={deal.id} />
        <AttachmentsPanel entityType="deal" entityId={deal.id} />
      </div>
    </SidePanel>
  )
}

function DealQuotesSection({ dealId }: { dealId: string }) {
  const { data, isLoading } = useDealQuotes(dealId)
  const quotes = data?.data ?? []
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <p className="text-sm font-semibold text-[#1A1D23]">Quotes</p>
        <Button variant="outline" size="sm" onClick={() => setShowCreate(true)}>
          <Plus className="h-3.5 w-3.5 mr-1" /> New quote
        </Button>
      </div>
      {isLoading ? (
        <Spinner size="sm" />
      ) : quotes.length === 0 ? (
        <p className="text-sm text-[#6B7280]">No quotes yet.</p>
      ) : (
        <div className="space-y-1.5">
          {quotes.map((q) => (
            <div
              key={q.id}
              className="flex items-center justify-between rounded-lg border border-[#E5E7EB] px-3 py-2"
            >
              <div className="flex items-center gap-2 min-w-0">
                <FileText className="h-3.5 w-3.5 text-[#6B7280] shrink-0" />
                <span className="text-sm font-medium text-[#1A1D23] truncate">{q.title}</span>
                <QuoteStatusBadge status={q.status} />
              </div>
              <span className="text-sm font-semibold text-[#1B3A4B] ml-2 shrink-0">
                {formatCurrency(q.total_cents / 100, q.currency)}
              </span>
            </div>
          ))}
        </div>
      )}
      {showCreate && <QuoteBuilder dealId={dealId} onClose={() => setShowCreate(false)} />}
    </div>
  )
}

// ---- Stats overlay ----

interface PipelineStats {
  pipelineValue: number
  winRate: number
  avgAgeDays: number
}

function StatsOverlay({ stats }: { stats: PipelineStats }) {
  return (
    <div
      className="fixed bottom-8 right-8 bg-white/90 backdrop-blur-xl p-5 rounded-xl border border-[#E5E7EB] shadow-lg flex gap-6 z-30"
      aria-label="Pipeline statistics"
    >
      <div className="flex flex-col">
        <span className="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest">Pipeline Value</span>
        <span className="text-xl font-extrabold text-[#1B3A4B]">{formatCurrency(stats.pipelineValue)}</span>
      </div>
      <div className="w-px bg-[#E5E7EB]" />
      <div className="flex flex-col">
        <span className="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest">Win Rate</span>
        <span className="text-xl font-extrabold text-[#1B3A4B]">{stats.winRate.toFixed(1)}%</span>
      </div>
      <div className="w-px bg-[#E5E7EB]" />
      <div className="flex flex-col">
        <span className="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest">Avg Age</span>
        <span className="text-xl font-extrabold text-[#1B3A4B]">{stats.avgAgeDays}d</span>
      </div>
    </div>
  )
}

// ---- Filter chip ----

function FilterChipBadge({ chip, onRemove }: { chip: FilterChip; onRemove: (key: string) => void }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-[#E8EDF2] px-2.5 py-1 text-xs font-medium text-[#1B3A4B]">
      {chip.label}
      <button
        onClick={() => onRemove(chip.key)}
        className="ml-0.5 hover:text-[#EF4444] transition-colors"
        aria-label={`Remove filter: ${chip.label}`}
      >
        <X className="h-3 w-3" />
      </button>
    </span>
  )
}

// ---- Main page ----

type ViewMode = 'board' | 'list'

export function DealsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const viewMode = (searchParams.get('view') as ViewMode) ?? 'board'
  const [search, setSearch] = useState('')
  const [stage, setStage] = useState('')
  const [priority, setPriority] = useState('')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [sortKey, setSortKey] = useState('created_at:desc')

  const debouncedSearch = useDebounce(search, 300)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  function setViewMode(mode: ViewMode) {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.set('view', mode)
      return next
    })
  }

  // Active filter chips
  const activeChips: FilterChip[] = useMemo(() => {
    const chips: FilterChip[] = []
    if (debouncedSearch) chips.push({ key: 'search', label: `"${debouncedSearch}"` })
    if (stage) chips.push({ key: 'stage', label: STAGE_OPTIONS.find((o) => o.value === stage)?.label ?? stage })
    if (priority) chips.push({ key: 'priority', label: PRIORITY_OPTIONS.find((o) => o.value === priority)?.label ?? priority })
    return chips
  }, [debouncedSearch, stage, priority])

  function removeChip(key: string) {
    if (key === 'search') setSearch('')
    if (key === 'stage') setStage('')
    if (key === 'priority') setPriority('')
    setPage(1)
  }

  function clearAll() {
    setSearch('')
    setStage('')
    setPriority('')
    setPage(1)
  }

  // Kanban: all deals (no pagination)
  const kanbanQuery = useDeals({
    per_page: 500,
    search: debouncedSearch || undefined,
    stage: (stage as DealStage) || undefined,
  })

  // List: paginated
  const listQuery = useDeals({
    page,
    per_page: 25,
    search: debouncedSearch || undefined,
    stage: (stage as DealStage) || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const kanbanDeals = kanbanQuery.data?.data ?? []
  const listDeals = listQuery.data?.data ?? []
  const meta = listQuery.data?.meta
  const totalDeals = (viewMode === 'board' ? kanbanQuery.data?.meta : listQuery.data?.meta)?.total

  // Pipeline stats
  const pipelineStats: PipelineStats = useMemo(() => {
    const all = kanbanDeals
    const activeDeals = all.filter((d) => d.stage !== 'closed_won' && d.stage !== 'closed_lost')
    const wonDeals = all.filter((d) => d.stage === 'closed_won')
    const closedDeals = all.filter((d) => d.stage === 'closed_won' || d.stage === 'closed_lost')
    const pipelineValue = activeDeals.reduce((s, d) => s + ((d.value_cents ?? 0) / 100), 0)
    const winRate = closedDeals.length > 0 ? (wonDeals.length / closedDeals.length) * 100 : 0
    const now = Date.now()
    const avgAge = all.length > 0
      ? Math.round(all.reduce((s, d) => s + (now - new Date(d.created_at).getTime()) / 86_400_000, 0) / all.length)
      : 0
    return { pipelineValue, winRate, avgAgeDays: avgAge }
  }, [kanbanDeals])

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  const columns: Column<Deal>[] = [
    {
      key: 'title',
      header: 'Title',
      sortable: true,
      render: (d) => <span className="font-medium text-[#1A1D23]">{d.title}</span>,
    },
    {
      key: 'value',
      header: 'Value',
      sortable: true,
      render: (d) => (
        <span className="font-semibold text-[#1B3A4B]">
          {formatCurrency(d.value_cents / 100, d.currency)}
        </span>
      ),
    },
    {
      key: 'stage',
      header: 'Stage',
      sortable: true,
      render: (d) => (
        <Badge variant={stageBadge[d.stage]}>{stageLabel[d.stage]}</Badge>
      ),
    },
    {
      key: 'account',
      header: 'Account',
      hideOnMobile: true,
      render: (d) => <span className="text-[#6B7280]">{d.account?.name ?? '—'}</span>,
    },
    {
      key: 'contact',
      header: 'Contact',
      hideOnMobile: true,
      render: (d) =>
        d.contact ? (
          <span className="text-[#6B7280]">
            {d.contact.first_name} {d.contact.last_name}
          </span>
        ) : (
          <span className="text-[#6B7280]">—</span>
        ),
    },
    {
      key: 'expected_close_date',
      header: 'Close Date',
      sortable: true,
      hideOnMobile: true,
      render: (d) => <span className="text-[#6B7280] text-xs">{d.expected_close_date ? formatDate(d.expected_close_date) : '—'}</span>,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
            Sales Operations
          </p>
          <h1 className="text-3xl font-extrabold text-[#1B3A4B] tracking-tight leading-none">
            Pipeline.
          </h1>
          <p className="mt-1 text-sm text-[#6B7280]">
            {totalDeals !== undefined ? `${totalDeals} total` : 'Loading…'}
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          {/* View toggle */}
          <div
            className="flex rounded-lg border border-[#E5E7EB] bg-white overflow-hidden"
            role="group"
            aria-label="View mode"
          >
            <button
              onClick={() => setViewMode('board')}
              aria-pressed={viewMode === 'board'}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm font-semibold transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B] ${
                viewMode === 'board'
                  ? 'bg-[#1B3A4B] text-white'
                  : 'text-[#6B7280] hover:bg-[#F7F8FA]'
              }`}
            >
              <LayoutGrid className="h-4 w-4" aria-hidden="true" />
              Board
            </button>
            <button
              onClick={() => setViewMode('list')}
              aria-pressed={viewMode === 'list'}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm font-semibold transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B] ${
                viewMode === 'list'
                  ? 'bg-[#1B3A4B] text-white'
                  : 'text-[#6B7280] hover:bg-[#F7F8FA]'
              }`}
            >
              <List className="h-4 w-4" aria-hidden="true" />
              List
            </button>
          </div>

          <Button
            onClick={() => setShowCreate(true)}
            className="bg-[#1B3A4B] hover:bg-[#1B3A4B]/90 text-white"
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            New Deal
          </Button>
        </div>
      </div>

      {/* Filter bar */}
      <div className="flex flex-wrap items-center gap-2">
        <input
          type="search"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          placeholder="Search pipeline…"
          aria-label="Search pipeline"
          className="rounded-lg border border-[#E5E7EB] bg-white px-3 py-1.5 text-sm text-[#1A1D23] placeholder:text-[#6B7280] focus:outline-none focus:ring-2 focus:ring-[#1B3A4B] w-48"
        />

        <select
          value={stage}
          onChange={(e) => { setStage(e.target.value); setPage(1) }}
          aria-label="Filter by stage"
          className="rounded-lg border border-[#E5E7EB] bg-white px-3 py-1.5 text-sm text-[#6B7280] focus:outline-none focus:ring-2 focus:ring-[#1B3A4B]"
        >
          <option value="">All Stages</option>
          {STAGE_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>

        <select
          value={priority}
          onChange={(e) => { setPriority(e.target.value); setPage(1) }}
          aria-label="Filter by priority"
          className="rounded-lg border border-[#E5E7EB] bg-white px-3 py-1.5 text-sm text-[#6B7280] focus:outline-none focus:ring-2 focus:ring-[#1B3A4B]"
        >
          <option value="">All Priorities</option>
          {PRIORITY_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>

        {/* Active filter chips */}
        {activeChips.map((chip) => (
          <FilterChipBadge key={chip.key} chip={chip} onRemove={removeChip} />
        ))}

        {activeChips.length > 1 && (
          <button
            onClick={clearAll}
            className="text-xs font-medium text-[#6B7280] hover:text-[#EF4444] transition-colors"
          >
            Clear all
          </button>
        )}
      </div>

      {/* Content */}
      {viewMode === 'board' ? (
        <div>
          {kanbanQuery.isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Spinner size="lg" />
            </div>
          ) : (
            <KanbanBoard deals={kanbanDeals} onCardClick={(id) => setSelectedId(id)} />
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-[#E5E7EB] bg-white overflow-hidden shadow-sm">
          <Table
            columns={columns}
            data={listDeals}
            isLoading={listQuery.isLoading}
            onSort={handleSort}
            onRowClick={(d) => setSelectedId(d.id)}
            emptyIcon={TrendingUp}
            emptyTitle="No deals found"
            emptyDescription="Try adjusting your filters."
            keyExtractor={(d) => d.id}
          />

          {meta && meta.total_pages > 1 && (
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 border-t border-[#E5E7EB] px-4 py-3">
              <p className="text-sm text-[#6B7280]">
                Page {meta.page} of {meta.total_pages}
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= meta.total_pages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Stats overlay (board view) */}
      {viewMode === 'board' && !kanbanQuery.isLoading && (
        <StatsOverlay stats={pipelineStats} />
      )}

      {selectedId && (
        <DealDetail dealId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {showCreate && <DealForm onClose={() => setShowCreate(false)} />}
    </div>
  )
}
