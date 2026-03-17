import { useState, useCallback } from 'react'
import { Plus } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useTickets } from '@/hooks/useTickets'
import { FilterBar } from '@/components/ui/FilterBar'
import { Button } from '@/components/ui/Button'
import { TicketList } from '@/components/omnir/TicketList'
import { TicketDetail } from '@/components/omnir/TicketDetail'
import { TicketForm } from '@/components/omnir/TicketForm'
import { useTicketFilterStore } from '@/stores/ticketFilters'
import type { TicketStatus, TicketPriority } from '@/api/types'

const STATUS_OPTIONS = [
  { label: 'Open', value: 'open' },
  { label: 'Pending', value: 'pending' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' },
]

const PRIORITY_OPTIONS = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Critical', value: 'critical' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
  { label: 'Priority ↑', value: 'priority:asc' },
  { label: 'Priority ↓', value: 'priority:desc' },
  { label: 'Status A–Z', value: 'status:asc' },
]

export function TicketsPage() {
  const { status, priority, search, page, setStatus, setPriority, setSearch, setPage } =
    useTicketFilterStore()

  const [sortKey, setSortKey] = useState('created_at:desc')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)

  const debouncedSearch = useDebounce(search, 300)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const { data, isLoading } = useTickets({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    status: (status as TicketStatus) || undefined,
    priority: (priority as TicketPriority) || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const tickets = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Help Desk</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total` : 'Loading…'}
          </p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4" />
          New Ticket
        </Button>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => setSearch(v)}
        searchPlaceholder="Search tickets…"
        filters={[
          {
            label: 'Status',
            value: status,
            options: STATUS_OPTIONS,
            onChange: (v) => setStatus(v as TicketStatus | ''),
          },
          {
            label: 'Priority',
            value: priority,
            options: PRIORITY_OPTIONS,
            onChange: (v) => setPriority(v as TicketPriority | ''),
          },
          {
            label: 'Sort',
            value: sortKey,
            options: SORT_OPTIONS,
            onChange: (v) => { setSortKey(v) },
          },
        ]}
      />

      {/* Table */}
      <TicketList
        tickets={tickets}
        isLoading={isLoading}
        page={page}
        totalPages={meta?.total_pages ?? 1}
        total={meta?.total ?? 0}
        onPageChange={setPage}
        onRowClick={(t) => setSelectedId(t.id)}
        sortBy={sortBy}
        sortDir={sortDir}
        onSort={handleSort}
      />

      {/* Detail panel */}
      {selectedId && (
        <TicketDetail ticketId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {/* Create form */}
      {showForm && (
        <TicketForm
          onClose={() => setShowForm(false)}
          onCreated={(id) => setSelectedId(id)}
        />
      )}
    </div>
  )
}
