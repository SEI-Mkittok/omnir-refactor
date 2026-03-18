import { useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus, TrendingUp, LayoutGrid, List, Download } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useDeals, useDeal, useDeleteDeal, useUpdateDeal } from '@/hooks/useDeals'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { KanbanBoard } from '@/components/omnir/KanbanBoard'
import { ActivityTimeline } from '@/components/omnir/ActivityTimeline'
import { CustomFieldEditableSection } from '@/components/omnir/CustomFieldRenderer'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { formatDate, formatCurrency } from '@/lib/utils'
import { downloadExportCsv } from '@/api/importExport'
import { AttachmentsPanel } from '@/components/omnir/AttachmentsPanel'
import { DealForm } from '@/components/omnir/DealForm'
import type { Deal, DealStage, CustomFieldValues } from '@/api/types'

const STAGE_OPTIONS = [
  { label: 'Lead', value: 'lead' },
  { label: 'Qualified', value: 'qualified' },
  { label: 'Proposal', value: 'proposal' },
  { label: 'Negotiation', value: 'negotiation' },
  { label: 'Closed Won', value: 'closed_won' },
  { label: 'Closed Lost', value: 'closed_lost' },
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
        {/* Stage + value */}
        <div className="flex items-center gap-3">
          <Badge variant={stageBadge[deal.stage]} className="text-sm px-3 py-1">
            {stageLabel[deal.stage]}
          </Badge>
          <span className="text-2xl font-bold text-indigo-600">
            {formatCurrency(deal.value, deal.currency)}
          </span>
        </div>

        {/* Details */}
        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-3 text-sm">
          {deal.probability !== undefined && (
            <>
              <dt className="font-medium text-slate-500">Probability</dt>
              <dd className="text-slate-900">{deal.probability}%</dd>
            </>
          )}
          {deal.close_date && (
            <>
              <dt className="font-medium text-slate-500">Close Date</dt>
              <dd className="text-slate-900">{formatDate(deal.close_date)}</dd>
            </>
          )}
          {deal.account?.name && (
            <>
              <dt className="font-medium text-slate-500">Account</dt>
              <dd className="text-slate-900">{deal.account.name}</dd>
            </>
          )}
          {deal.contact && (
            <>
              <dt className="font-medium text-slate-500">Contact</dt>
              <dd className="text-slate-900">
                {deal.contact.first_name} {deal.contact.last_name}
              </dd>
            </>
          )}
          {deal.pipeline?.name && (
            <>
              <dt className="font-medium text-slate-500">Pipeline</dt>
              <dd className="text-slate-900">{deal.pipeline.name}</dd>
            </>
          )}
          <dt className="font-medium text-slate-500">Created</dt>
          <dd className="text-slate-900">{formatDate(deal.created_at)}</dd>
        </dl>

        {/* Tags */}
        {deal.tags && deal.tags.length > 0 && (
          <div>
            <p className="mb-2 text-sm font-medium text-slate-500">Tags</p>
            <div className="flex flex-wrap gap-1.5">
              {deal.tags.map((tag) => (
                <Badge key={tag} variant="default">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Custom Fields */}
        <CustomFieldEditableSection
          fields={customFields}
          values={deal.custom_fields as CustomFieldValues | undefined}
          onSave={async (cf) => { await updateDeal.mutateAsync({ id: dealId, payload: { custom_fields: cf as Record<string, unknown> } }) }}
        />

        {/* Activity Timeline */}
        <ActivityTimeline dealId={deal.id} />

        {/* Attachments */}
        <AttachmentsPanel entityType="deal" entityId={deal.id} />
      </div>
    </SidePanel>
  )
}

// ---- Main page ----

type ViewMode = 'kanban' | 'list'

export function DealsPage() {
  const [searchParams] = useSearchParams()
  const [viewMode, setViewMode] = useState<ViewMode>('kanban')
  const [search, setSearch] = useState('')
  const [stage, setStage] = useState('')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(searchParams.get('openId'))
  const [showCreate, setShowCreate] = useState(false)

  const debouncedSearch = useDebounce(search, 300)

  // Kanban: fetch all (no pagination)
  const kanbanQuery = useDeals({
    per_page: 500,
    search: debouncedSearch || undefined,
    stage: (stage as DealStage) || undefined,
  })

  // List: paginated
  const listQuery = useDeals({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    stage: (stage as DealStage) || undefined,
    sort_by: 'created_at',
    sort_dir: 'desc',
  })

  const kanbanDeals = kanbanQuery.data?.data ?? []
  const listDeals = listQuery.data?.data ?? []
  const meta = listQuery.data?.meta

  const totalDeals = (viewMode === 'kanban' ? kanbanQuery.data?.meta : listQuery.data?.meta)?.total

  const handleSort = useCallback((_key: string) => {
    // future: column sort in list mode
  }, [])

  const columns: Column<Deal>[] = [
    {
      key: 'title',
      header: 'Title',
      sortable: true,
      render: (d) => <span className="font-medium text-slate-900">{d.title}</span>,
    },
    {
      key: 'value',
      header: 'Value',
      sortable: true,
      render: (d) => (
        <span className="font-semibold text-indigo-600">
          {formatCurrency(d.value, d.currency)}
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
      render: (d) => <span className="text-slate-600">{d.account?.name ?? '—'}</span>,
    },
    {
      key: 'probability',
      header: 'Probability',
      hideOnMobile: true,
      render: (d) =>
        d.probability !== undefined ? (
          <span className="text-slate-600">{d.probability}%</span>
        ) : (
          <span className="text-slate-400">—</span>
        ),
    },
    {
      key: 'close_date',
      header: 'Close Date',
      sortable: true,
      hideOnMobile: true,
      render: (d) => (
        <span className="text-slate-500 text-xs">{formatDate(d.close_date)}</span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Deals</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {totalDeals !== undefined ? `${totalDeals} total` : 'Loading…'}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {/* View toggle */}
          <div className="flex rounded-md border border-slate-200 bg-white overflow-hidden">
            <button
              onClick={() => setViewMode('kanban')}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium transition-colors ${
                viewMode === 'kanban'
                  ? 'bg-indigo-600 text-white'
                  : 'text-slate-600 hover:bg-slate-50'
              }`}
            >
              <LayoutGrid className="h-4 w-4" />
              Board
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium transition-colors ${
                viewMode === 'list'
                  ? 'bg-indigo-600 text-white'
                  : 'text-slate-600 hover:bg-slate-50'
              }`}
            >
              <List className="h-4 w-4" />
              List
            </button>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              downloadExportCsv('deals', {
                ...(debouncedSearch ? { q: debouncedSearch } : {}),
                ...(stage ? { stage } : {}),
              })
            }
          >
            <Download className="h-4 w-4" />
            Export CSV
          </Button>
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" />
            Add Deal
          </Button>
        </div>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search deals…"
        filters={[
          {
            label: 'Stage',
            value: stage,
            options: STAGE_OPTIONS,
            onChange: (v) => { setStage(v); setPage(1) },
          },
        ]}
      />

      {/* Content */}
      {viewMode === 'kanban' ? (
        <div>
          {kanbanQuery.isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Spinner size="lg" />
            </div>
          ) : (
            <KanbanBoard deals={kanbanDeals} />
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
          <Table
            columns={columns}
            data={listDeals}
            isLoading={listQuery.isLoading}
            onSort={handleSort}
            onRowClick={(d) => setSelectedId(d.id)}
            emptyIcon={TrendingUp}
            emptyTitle="No deals found"
            emptyDescription="Try adjusting your search or stage filter."
            keyExtractor={(d) => d.id}
          />

          {meta && meta.total_pages > 1 && (
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 border-t border-slate-200 px-4 py-3">
              <p className="text-sm text-slate-500">
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

      {selectedId && (
        <DealDetail dealId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {showCreate && <DealForm onClose={() => setShowCreate(false)} />}
    </div>
  )
}
