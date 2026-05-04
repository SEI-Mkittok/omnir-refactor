import { useState, useCallback, useRef } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { Plus, Search, X, Inbox, AlertCircle, ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'
import { TicketForm } from '@/components/omnir/TicketForm'
import * as RadixSelect from '@radix-ui/react-select'
import { ChevronDown as ChevronDownIcon, Check } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useTickets } from '@/hooks/useTickets'
import { useTicketReport } from '@/hooks/useReports'
import { useTicketFilterStore } from '@/stores/ticketFilters'
import type { Ticket, TicketStatus, TicketPriority } from '@/api/types'
import { cn } from '@/lib/utils'

// ─── Design helpers ────────────────────────────────────────────────────────────

const STATUS_BG: Record<TicketStatus, string> = {
  open: '#0D9488',
  pending: '#F59E0B',
  resolved: '#22C55E',
  closed: '#9CA3AF',
}
const STATUS_LABEL: Record<TicketStatus, string> = {
  open: 'OPEN',
  pending: 'PENDING',
  resolved: 'RESOLVED',
  closed: 'CLOSED',
}
const PRIORITY_DOT: Record<TicketPriority, string> = {
  critical: '#EF4444',
  high: '#F97316',
  medium: '#3B82F6',
  low: '#9CA3AF',
}
const PRIORITY_LABEL: Record<TicketPriority, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
}

function StatusBadge({ status }: { status: TicketStatus }) {
  return (
    <span
      aria-label={`Status: ${STATUS_LABEL[status]}`}
      className="inline-block rounded-full px-2.5 py-0.5 text-[12px] font-semibold uppercase text-white"
      style={{ background: STATUS_BG[status] }}
    >
      {STATUS_LABEL[status]}
    </span>
  )
}

function PriorityDot({ priority }: { priority: TicketPriority }) {
  return (
    <span aria-label={`Priority: ${PRIORITY_LABEL[priority]}`} className="inline-flex items-center gap-1.5">
      <span
        aria-hidden="true"
        className="inline-block h-2 w-2 rounded-full"
        style={{ background: PRIORITY_DOT[priority] }}
      />
      <span className="text-[13px] text-[var(--text-primary)]">{PRIORITY_LABEL[priority]}</span>
    </span>
  )
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

/** Initials from a full name string (e.g. "Alex Wu" → "AW") */
function nameInitials(name?: string): string {
  if (!name) return '?'
  return name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase()
}

function formatHours(h: number | null): string {
  if (h === null) return '—'
  if (h < 1) return `${Math.round(h * 60)}m`
  const hrs = Math.floor(h)
  const mins = Math.round((h - hrs) * 60)
  return mins > 0 ? `${hrs}h ${mins}m` : `${hrs}h`
}

// ─── Filter Select ─────────────────────────────────────────────────────────────

interface FilterSelectProps {
  id: string
  label: string
  value: string
  options: { value: string; label: string }[]
  onValueChange: (v: string) => void
  onClear: () => void
}

function FilterSelect({ id, label, value, options, onValueChange, onClear }: FilterSelectProps) {
  const selected = options.find((o) => o.value === value)
  return (
    <div className="relative flex items-center">
      <label htmlFor={id} className="sr-only">{`Filter by ${label}`}</label>
      <RadixSelect.Root value={value || '__all__'} onValueChange={(v) => onValueChange(v === '__all__' ? '' : v)}>
        <RadixSelect.Trigger
          id={id}
          aria-label={`Filter by ${label}`}
          className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[var(--border-default)] bg-[var(--surface-card)] px-3 text-[13px] text-[var(--text-primary)] hover:border-[var(--color-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
        >
          <RadixSelect.Value>{selected ? selected.label : label}</RadixSelect.Value>
          <RadixSelect.Icon>
            <ChevronDownIcon className="h-3.5 w-3.5 text-[var(--text-label)]" />
          </RadixSelect.Icon>
        </RadixSelect.Trigger>
        <RadixSelect.Portal>
          <RadixSelect.Content
            className="z-50 rounded-md border border-[var(--border-default)] bg-white shadow-md overflow-hidden"
            position="popper"
            sideOffset={4}
          >
            <RadixSelect.Viewport className="p-1">
              <RadixSelect.Item
                value="__all__"
                className="relative flex cursor-pointer select-none items-center gap-2 rounded px-3 py-1.5 text-[13px] text-[var(--text-secondary)] outline-none hover:bg-[var(--color-primary-light)] focus:bg-[var(--color-primary-light)]"
              >
                <RadixSelect.ItemText>{label === 'All Statuses' ? 'All Statuses' : label === 'All Priorities' ? 'All Priorities' : `All ${label}`}</RadixSelect.ItemText>
              </RadixSelect.Item>
              {options.map((opt) => (
                <RadixSelect.Item
                  key={opt.value}
                  value={opt.value}
                  className="relative flex cursor-pointer select-none items-center gap-2 rounded px-3 py-1.5 text-[13px] text-[var(--text-primary)] outline-none hover:bg-[var(--color-primary-light)] data-[state=checked]:font-medium focus:bg-[var(--color-primary-light)]"
                >
                  <RadixSelect.ItemText>{opt.label}</RadixSelect.ItemText>
                  <RadixSelect.ItemIndicator className="absolute right-2">
                    <Check className="h-3 w-3" />
                  </RadixSelect.ItemIndicator>
                </RadixSelect.Item>
              ))}
            </RadixSelect.Viewport>
          </RadixSelect.Content>
        </RadixSelect.Portal>
      </RadixSelect.Root>
      {value && (
        <button
          onClick={onClear}
          aria-label={`Clear ${label} filter`}
          className="ml-1 rounded p-0.5 text-[var(--text-label)] hover:text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  )
}

// ─── Metrics Row ───────────────────────────────────────────────────────────────

function MetricsRow() {
  const { data, isLoading } = useTicketReport()

  const resolutionRate =
    data != null
      ? data.total_closed + (data.total_open ?? 0) > 0
        ? Math.round((data.total_closed / (data.total_closed + (data.total_open ?? 0))) * 100)
        : null
      : null
  const slaCompliance =
    data != null ? Math.round((1 - data.breach_rate) * 100) : null

  const metrics = [
    {
      label: 'AVG RESPONSE TIME',
      value: isLoading ? null : formatHours(data?.avg_resolution_hours ?? null),
    },
    {
      label: 'RESOLUTION RATE',
      value: isLoading ? null : resolutionRate !== null ? `${resolutionRate}%` : '—',
    },
    {
      label: 'SLA COMPLIANCE',
      value: isLoading ? null : slaCompliance !== null ? `${slaCompliance}%` : '—',
    },
  ]

  return (
    <dl className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      {metrics.map((m) => (
        <div
          key={m.label}
          className="rounded-lg bg-[var(--surface-card)] border border-[var(--border-default)] px-6 py-5"
        >
          {isLoading ? (
            <div aria-busy="true" className="space-y-2">
              <div aria-hidden="true" className="h-3 w-28 rounded bg-[var(--border-subtle)] animate-pulse" />
              <div aria-hidden="true" className="h-9 w-24 rounded bg-[var(--border-subtle)] animate-pulse" />
            </div>
          ) : (
            <>
              <dt className="text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
                {m.label}
              </dt>
              <dd
                className="mt-1 text-[36px] font-bold leading-none text-[var(--text-primary)]"
                style={{ letterSpacing: 'var(--letter-spacing-tight)' }}
              >
                {m.value ?? '—'}
              </dd>
            </>
          )}
        </div>
      ))}
    </dl>
  )
}

// ─── Tickets Table ─────────────────────────────────────────────────────────────

type SortKey = 'subject' | 'status' | 'priority' | 'created_at'

interface SortIconProps { col: SortKey; sortBy: string; sortDir: 'asc' | 'desc' }
function SortIconWidget({ col, sortBy, sortDir }: SortIconProps) {
  if (sortBy !== col) return <ChevronsUpDown className="h-3 w-3 text-[var(--text-label)]" />
  return sortDir === 'asc'
    ? <ChevronUp className="h-3 w-3 text-[var(--color-primary)]" />
    : <ChevronDown className="h-3 w-3 text-[var(--color-primary)]" />
}

interface TicketsTableProps {
  tickets: Ticket[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  page: number
  totalPages: number
  total: number
  onPageChange: (p: number) => void
  onRowClick: (t: Ticket) => void
  sortBy: string
  sortDir: 'asc' | 'desc'
  onSort: (key: string) => void
  hasActiveFilters: boolean
  onClearFilters: () => void
}

function TicketsTable({
  tickets,
  isLoading,
  isError,
  onRetry,
  page,
  totalPages,
  total,
  onPageChange,
  onRowClick,
  sortBy,
  sortDir,
  onSort,
  hasActiveFilters,
  onClearFilters,
}: TicketsTableProps) {
  const perPage = 20
  const start = (page - 1) * perPage + 1
  const end = Math.min(page * perPage, total)

  const headers: { key: SortKey | string; label: string; sortable?: boolean; width?: string; hidden?: boolean }[] = [
    { key: 'subject', label: 'Subject', sortable: true },
    { key: 'contact', label: 'Contact', hidden: false },
    { key: 'status', label: 'Status', sortable: true, width: '120px' },
    { key: 'priority', label: 'Priority', sortable: true, width: '120px' },
    { key: 'assignee', label: 'Assigned To', hidden: true },
    { key: 'created_at', label: 'Created', sortable: true, width: '140px', hidden: true },
  ]

  return (
    <div className="rounded-lg border border-[var(--border-default)] bg-[var(--surface-card)] overflow-hidden">
      <div className="overflow-x-auto">
        <table aria-label="Tickets" className="min-w-full">
          <thead>
            <tr style={{ background: 'var(--surface-app)', borderBottom: '1px solid var(--border-default)', height: '40px' }}>
              {headers.map((h) => (
                <th
                  key={h.key}
                  scope="col"
                  aria-sort={sortBy === h.key ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}
                  className={cn(
                    'px-4 text-left text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]',
                    h.hidden && 'hidden lg:table-cell'
                  )}
                  style={h.width ? { width: h.width } : undefined}
                >
                  {h.sortable ? (
                    <button
                      onClick={() => onSort(h.key)}
                      className="inline-flex items-center gap-1 hover:text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
                    >
                      {h.label}
                      <SortIconWidget col={h.key as SortKey} sortBy={sortBy} sortDir={sortDir} />
                    </button>
                  ) : (
                    h.label
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isError ? (
              <tr>
                <td colSpan={headers.length} className="px-4 py-12 text-center">
                  <div className="flex flex-col items-center gap-3">
                    <AlertCircle className="h-8 w-8 text-[var(--color-danger)]" />
                    <p className="text-[14px] font-medium text-[var(--text-primary)]">Something went wrong.</p>
                    <p className="text-[13px] text-[var(--text-secondary)]">Couldn't load tickets. Check your connection and try again.</p>
                    <button
                      onClick={onRetry}
                      className="mt-1 h-9 rounded-md border border-[var(--border-default)] px-4 text-[13px] text-[var(--text-primary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
                    >
                      Retry
                    </button>
                  </div>
                </td>
              </tr>
            ) : isLoading ? (
              [...Array(8)].map((_, i) => (
                <tr key={i} style={{ height: '60px', borderBottom: '1px solid var(--border-subtle)' }}>
                  {[60, 30, 12, 12, 20, 15].map((w, j) => (
                    <td key={j} className={cn('px-4', j >= 4 && 'hidden lg:table-cell')}>
                      <div
                        aria-hidden="true"
                        className="h-3 rounded animate-pulse"
                        style={{ width: `${w}%`, background: 'var(--border-subtle)' }}
                      />
                    </td>
                  ))}
                </tr>
              ))
            ) : tickets.length === 0 ? (
              <tr>
                <td colSpan={headers.length} style={{ minHeight: '320px' }}>
                  <div className="flex flex-col items-center justify-center py-20 gap-3">
                    <Inbox className="h-12 w-12 text-[var(--text-label)]" />
                    <p className="text-[14px] font-medium text-[var(--text-primary)]">
                      {hasActiveFilters ? 'No tickets found.' : 'Your queue is clear.'}
                    </p>
                    <p className="text-[13px] text-[var(--text-secondary)]">
                      {hasActiveFilters ? 'No tickets match your current filters.' : 'No open tickets.'}
                    </p>
                    {hasActiveFilters && (
                      <div className="flex gap-3 mt-1">
                        <button
                          onClick={onClearFilters}
                          className="text-[13px] text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
                        >
                          Clear filters
                        </button>
                      </div>
                    )}
                  </div>
                </td>
              </tr>
            ) : (
              tickets.map((ticket) => (
                <tr
                  key={ticket.id}
                  tabIndex={0}
                  aria-label={`Ticket #${ticket.id.slice(-4)}: ${ticket.subject}, ${STATUS_LABEL[ticket.status]}, ${PRIORITY_LABEL[ticket.priority]} priority`}
                  onClick={() => onRowClick(ticket)}
                  onKeyDown={(e) => { if (e.key === 'Enter') onRowClick(ticket) }}
                  className="cursor-pointer transition-colors focus:outline-none focus:bg-[var(--color-primary-light)]"
                  style={{ height: '60px', borderBottom: '1px solid var(--border-subtle)' }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--surface-app)')}
                  onMouseLeave={(e) => (e.currentTarget.style.background = '')}
                >
                  {/* Subject */}
                  <td className="px-4">
                    <div>
                      <p className="text-[14px] text-[var(--text-primary)] truncate max-w-xs">{ticket.subject}</p>
                      <p className="text-[12px] text-[var(--text-label)]">#{ticket.id.slice(-4)}</p>
                    </div>
                  </td>
                  {/* Contact */}
                  <td className="px-4">
                    {ticket.contact ? (
                      <Link
                        to={`/contacts/${ticket.contact.id}`}
                        onClick={(e) => e.stopPropagation()}
                        className="flex items-center gap-2 hover:underline"
                      >
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-[11px] font-semibold text-[var(--color-primary)]">
                          {nameInitials(ticket.contact.name)}
                        </div>
                        <span className="text-[14px] text-[var(--text-primary)] truncate max-w-[120px]">
                          {ticket.contact.name || 'Unknown Contact'}
                        </span>
                      </Link>
                    ) : (
                      <span className="text-[13px] text-[var(--text-label)]">—</span>
                    )}
                  </td>
                  {/* Status */}
                  <td className="px-4"><StatusBadge status={ticket.status} /></td>
                  {/* Priority */}
                  <td className="px-4"><PriorityDot priority={ticket.priority} /></td>
                  {/* Assigned To */}
                  <td className="hidden lg:table-cell px-4">
                    {ticket.assignee ? (
                      <div className="flex items-center gap-2">
                        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-[10px] font-semibold text-[var(--color-primary)]">
                          {nameInitials(ticket.assignee.name)}
                        </div>
                        <span className="text-[13px] text-[var(--text-primary)]">{ticket.assignee.name}</span>
                      </div>
                    ) : (
                      <span className="text-[13px] italic text-[var(--text-label)]">Unassigned</span>
                    )}
                  </td>
                  {/* Created */}
                  <td className="hidden lg:table-cell px-4">
                    <span title={new Date(ticket.created_at).toLocaleString()} className="text-[13px] text-[var(--text-secondary)]">
                      {relativeTime(ticket.created_at)}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 border-t border-[var(--border-default)] px-4 py-3">
        <p className="text-[13px] text-[var(--text-secondary)]">
          {total > 0
            ? `Showing ${start}–${end} of ${total} ticket${total !== 1 ? 's' : ''}`
            : 'No tickets'}
        </p>
        {totalPages > 1 && (
          <div className="flex items-center gap-1">
            <button
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1}
              className="flex h-8 w-8 items-center justify-center rounded text-[13px] text-[var(--text-secondary)] hover:bg-[var(--surface-app)] disabled:opacity-40 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              aria-label="Previous page"
            >
              ←
            </button>
            {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
              const p = i + 1
              return (
                <button
                  key={p}
                  onClick={() => onPageChange(p)}
                  aria-current={p === page ? 'page' : undefined}
                  className={cn(
                    'flex h-8 min-w-[32px] items-center justify-center rounded px-1 text-[13px] transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]',
                    p === page
                      ? 'bg-[var(--color-primary)] text-white font-medium'
                      : 'text-[var(--text-secondary)] hover:bg-[var(--surface-app)]'
                  )}
                >
                  {p}
                </button>
              )
            })}
            <button
              onClick={() => onPageChange(page + 1)}
              disabled={page >= totalPages}
              className="flex h-8 w-8 items-center justify-center rounded text-[13px] text-[var(--text-secondary)] hover:bg-[var(--surface-app)] disabled:opacity-40 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              aria-label="Next page"
            >
              →
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Tickets Page ──────────────────────────────────────────────────────────────

const STATUS_OPTIONS = [
  { label: 'Open', value: 'open' },
  { label: 'Pending', value: 'pending' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' },
]

const PRIORITY_OPTIONS = [
  { label: 'Critical', value: 'critical' },
  { label: 'High', value: 'high' },
  { label: 'Medium', value: 'medium' },
  { label: 'Low', value: 'low' },
]

export function TicketsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const accountId = searchParams.get('account_id') ?? ''
  const accountName = searchParams.get('account_name') ?? ''
  const contactId = searchParams.get('contact_id') ?? ''
  const contactName = searchParams.get('contact_name') ?? ''
  const { status, priority, search, page, setStatus, setPriority, setSearch, setPage, reset } =
    useTicketFilterStore()
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [showNewTicket, setShowNewTicket] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)

  const debouncedSearch = useDebounce(search, 300)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const hasActiveFilters = !!(status || priority || search)

  const { data, isLoading, isError, refetch } = useTickets({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    status: (status as TicketStatus) || undefined,
    priority: (priority as TicketPriority) || undefined,
    contact_id: contactId || undefined,
    account_id: accountId || undefined,
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
    <div className="space-y-5">
      {/* Page header */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <h1
          className="text-[36px] font-bold text-[var(--text-primary)] leading-tight"
          style={{ letterSpacing: 'var(--letter-spacing-tight)' }}
        >
          Tickets.
        </h1>
        <button
          onClick={() => setShowNewTicket(true)}
          className="inline-flex h-9 items-center gap-2 rounded-md bg-[var(--color-primary)] px-4 text-[14px] font-medium text-white hover:bg-[var(--color-primary-hover)] transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
        >
          <Plus className="h-4 w-4" />
          New Ticket
        </button>
      </div>

      {/* Metrics row */}
      <MetricsRow />

      {(accountId || contactId) && (
        <div className="flex items-center gap-3 rounded-lg border border-[var(--border-default)] bg-[var(--surface-card)] px-4 py-3">
          <span className="text-[14px] font-medium text-[var(--text-primary)]">
            {contactId
              ? contactName
                ? `Filtered to ${contactName}`
                : 'Contact filter active'
              : accountName
                ? `Filtered to ${accountName}`
                : 'Account filter active'}
          </span>
          <button
            onClick={() =>
              setSearchParams((prev) => {
                const next = new URLSearchParams(prev)
                next.delete('account_id')
                next.delete('account_name')
                next.delete('contact_id')
                next.delete('contact_name')
                return next
              })
            }
            className="text-[13px] text-[var(--color-primary)] hover:underline"
          >
            Clear
          </button>
        </div>
      )}

      {/* Filter bar */}
      <div className="flex flex-wrap items-center gap-2 py-1">
        <FilterSelect
          id="filter-status"
          label="All Statuses"
          value={status}
          options={STATUS_OPTIONS}
          onValueChange={(v) => setStatus(v as TicketStatus | '')}
          onClear={() => setStatus('')}
        />
        <FilterSelect
          id="filter-priority"
          label="All Priorities"
          value={priority}
          options={PRIORITY_OPTIONS}
          onValueChange={(v) => setPriority(v as TicketPriority | '')}
          onClear={() => setPriority('')}
        />

        {/* Search */}
        <div className="relative ml-auto">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-label)]" />
          <input
            ref={searchRef}
            type="text"
            placeholder="Search…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            aria-label="Search tickets"
            className="h-9 w-[220px] rounded-full border border-[var(--border-default)] bg-[var(--surface-card)] pl-9 pr-8 text-[13px] text-[var(--text-primary)] placeholder:text-[var(--text-label)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              aria-label="Clear search"
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[var(--text-label)] hover:text-[var(--text-primary)] focus:outline-none"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {hasActiveFilters && (
          <button
            onClick={() => { reset() }}
            className="ml-1 text-[14px] text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
          >
            Clear filters
          </button>
        )}
      </div>

      {/* Table */}
      <TicketsTable
        tickets={tickets}
        isLoading={isLoading}
        isError={isError}
        onRetry={() => refetch()}
        page={page}
        totalPages={meta?.total_pages ?? 1}
        total={meta?.total ?? 0}
        onPageChange={setPage}
        onRowClick={(t) => navigate(`/tickets/${t.id}`)}
        sortBy={sortBy}
        sortDir={sortDir}
        onSort={handleSort}
        hasActiveFilters={hasActiveFilters}
        onClearFilters={() => reset()}
      />

      {showNewTicket && (
        <TicketForm
          onClose={() => setShowNewTicket(false)}
          initialValues={{
            contact_id: contactId || undefined,
            account_id: accountId || undefined,
          }}
          onCreated={(ticketId) => navigate(`/tickets/${ticketId}`)}
        />
      )}
    </div>
  )
}
