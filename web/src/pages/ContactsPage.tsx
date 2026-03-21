import { useState, useCallback } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Plus, Users, Upload, Download } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useContacts } from '@/hooks/useContacts'
import { useUpdateView } from '@/hooks/useViews'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { formatDate } from '@/lib/utils'
import { stageBadgeVariant, stageLabel } from '@/components/omnir/ContactCard'
import { ContactDetailPanel } from '@/components/omnir/ContactDetailPanel'
import { ContactForm } from '@/components/omnir/ContactForm'
import { ViewPinBar } from '@/components/omnir/ViewPinBar'
import { ImportModal } from '@/components/omnir/ImportModal'
import { downloadExportCsv } from '@/api/importExport'
import type { Contact, ContactStage, SavedView } from '@/api/types'

const STAGE_OPTIONS = [
  { label: 'Lead', value: 'lead' },
  { label: 'Prospect', value: 'prospect' },
  { label: 'Customer', value: 'customer' },
  { label: 'Churned', value: 'churned' },
]

const SORT_OPTIONS = [
  { label: 'Name A–Z', value: 'last_name:asc' },
  { label: 'Name Z–A', value: 'last_name:desc' },
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
]

// ---- Main page ----

export function ContactsPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [search, setSearch] = useState(searchParams.get('search') ?? '')
  const [stage, setStage] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(searchParams.get('openId'))
  const [showForm, setShowForm] = useState(false)
  const [showImport, setShowImport] = useState(false)
  const [activeView, setActiveView] = useState<SavedView | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)

  const updateView = useUpdateView()
  const debouncedSearch = useDebounce(search, 300)

  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const currentFilters = {
    search: debouncedSearch || undefined,
    stage: stage || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  }

  const applyViewFilters = (view: SavedView) => {
    setActiveView(view)
    setHasUnsavedChanges(false)
    setSearch((view.filters.search as string) ?? '')
    setStage((view.filters.stage as string) ?? '')
    setSortKey(view.filters.sort_by ? `${view.filters.sort_by}:${view.filters.sort_dir ?? 'asc'}` : 'created_at:desc')
    setPage(1)
  }

  const markChanged = () => { if (activeView) setHasUnsavedChanges(true) }

  const handleUpdateView = async (viewId: string) => {
    await updateView.mutateAsync({ id: viewId, payload: { filters: currentFilters } })
    setHasUnsavedChanges(false)
  }

  const { data, isLoading } = useContacts({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    stage: (stage as ContactStage) || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const contacts = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback(
    (key: string) => {
      setSortKey((prev) => {
        const [prevKey, prevDir] = prev.split(':')
        if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
        return `${key}:asc`
      })
    },
    []
  )

  const columns: Column<Contact>[] = [
    {
      key: 'last_name',
      header: 'Name',
      sortable: true,
      render: (c) => (
        <span className="font-medium text-slate-900">
          {c.first_name} {c.last_name}
        </span>
      ),
    },
    {
      key: 'email',
      header: 'Email',
      render: (c) => <span className="text-slate-600">{c.email}</span>,
    },
    {
      key: 'title',
      header: 'Title',
      hideOnMobile: true,
      render: (c) => <span className="text-slate-600">{c.title ?? '—'}</span>,
    },
    {
      key: 'account',
      header: 'Account',
      hideOnMobile: true,
      render: (c) => <span className="text-slate-600">{c.account?.name ?? '—'}</span>,
    },
    {
      key: 'stage',
      header: 'Stage',
      sortable: true,
      render: (c) => (
        <Badge variant={stageBadgeVariant[c.stage]}>{stageLabel[c.stage]}</Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      hideOnMobile: true,
      render: (c) => <span className="text-slate-500 text-xs">{formatDate(c.created_at)}</span>,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">
            {activeView ? activeView.name : 'Contacts'}
          </h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total` : 'Loading…'}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setShowImport(true)}>
            <Upload className="h-4 w-4" />
            Import
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              downloadExportCsv('contacts', {
                ...(debouncedSearch ? { q: debouncedSearch } : {}),
                ...(stage ? { stage } : {}),
                ...(sortBy ? { sort: sortBy, order: sortDir } : {}),
              })
            }
          >
            <Download className="h-4 w-4" />
            Export CSV
          </Button>
          <Button onClick={() => setShowForm(true)}>
            <Plus className="h-4 w-4" />
            Add Contact
          </Button>
        </div>
      </div>

      {/* View pin bar */}
      <ViewPinBar
        entityType="contacts"
        activeViewId={activeView?.id ?? null}
        hasUnsavedChanges={hasUnsavedChanges}
        currentFilters={currentFilters}
        onSelectView={applyViewFilters}
        onClearView={() => { setActiveView(null); setHasUnsavedChanges(false) }}
        onViewSaved={(view) => setActiveView(view)}
        onUpdateView={handleUpdateView}
      />

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1); markChanged() }}
        searchPlaceholder="Search contacts…"
        filters={[
          {
            label: 'Stage',
            value: stage,
            options: STAGE_OPTIONS,
            onChange: (v) => { setStage(v); setPage(1); markChanged() },
          },
          {
            label: 'Sort',
            value: sortKey,
            options: SORT_OPTIONS,
            onChange: (v) => { setSortKey(v); setPage(1); markChanged() },
          },
        ]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={contacts}
          isLoading={isLoading}
          sortBy={sortBy}
          sortDir={sortDir}
          onSort={handleSort}
          onRowClick={(c) => navigate(`/contacts/${c.id}`)}
          emptyIcon={Users}
          emptyTitle="No contacts found"
          emptyDescription="Try adjusting your search or filters."
          keyExtractor={(c) => c.id}
        />

        {/* Pagination */}
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

      {/* Detail panel */}
      {selectedId && (
        <ContactDetailPanel contactId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {/* Create form */}
      <ContactForm open={showForm} onClose={() => setShowForm(false)} />

      {/* Import modal */}
      <ImportModal
        open={showImport}
        onClose={() => setShowImport(false)}
        entity="contacts"
        entityLabel="Contacts"
      />
    </div>
  )
}
