import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, ChevronRight, Search } from 'lucide-react'
import { usePortalTickets } from '@/hooks/usePortal'
import { useDebounce } from '@/hooks/useDebounce'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { formatDate } from '@/lib/utils'
import type { PortalTicketStatus, TicketPriority } from '@/api/types'

const STATUS_OPTIONS: { label: string; value: PortalTicketStatus | '' }[] = [
  { label: 'All', value: '' },
  { label: 'Open', value: 'open' },
  { label: 'In Progress', value: 'in_progress' },
  { label: 'Pending', value: 'pending' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' },
]

const statusVariant: Record<PortalTicketStatus, 'blue' | 'indigo' | 'yellow' | 'green' | 'gray'> = {
  open: 'blue',
  in_progress: 'indigo',
  pending: 'yellow',
  resolved: 'green',
  closed: 'gray',
}

const statusLabel: Record<PortalTicketStatus, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  pending: 'Pending',
  resolved: 'Resolved',
  closed: 'Closed',
}

const priorityVariant: Record<TicketPriority, 'gray' | 'blue' | 'orange' | 'red'> = {
  low: 'gray',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

export function PortalTicketsPage() {
  const [status, setStatus] = useState<PortalTicketStatus | ''>('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const debouncedSearch = useDebounce(search, 300)

  const { data, isLoading } = usePortalTickets({
    page,
    per_page: 20,
    status: status || undefined,
    search: debouncedSearch || undefined,
    sort_by: 'created_at',
    sort_dir: 'desc',
  })

  const tickets = data?.data ?? []
  const meta = data?.meta

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">My Tickets</h1>
          {meta && (
            <p className="mt-0.5 text-sm text-slate-500">{meta.total} total</p>
          )}
        </div>
        <Button asChild>
          <Link to="/portal/tickets/new">
            <Plus className="h-4 w-4" />
            New Ticket
          </Link>
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 min-w-48">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <Input
            type="search"
            placeholder="Search tickets…"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="pl-9"
          />
        </div>

        {/* Status tabs */}
        <div className="flex flex-wrap gap-1.5">
          {STATUS_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => { setStatus(opt.value as PortalTicketStatus | ''); setPage(1) }}
              className={[
                'rounded-full px-3 py-1 text-sm font-medium transition-colors',
                status === opt.value
                  ? 'bg-indigo-600 text-white'
                  : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50',
              ].join(' ')}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* List */}
      {isLoading ? (
        <div className="flex justify-center py-12">
          <Spinner />
        </div>
      ) : tickets.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-300 bg-white py-16 text-center">
          <p className="text-sm text-slate-500">
            {status || search ? 'No tickets match your filters.' : "You haven't submitted any tickets yet."}
          </p>
          {!status && !search && (
            <Button variant="outline" size="sm" className="mt-4" asChild>
              <Link to="/portal/tickets/new">Submit your first ticket</Link>
            </Button>
          )}
        </div>
      ) : (
        <div className="divide-y divide-slate-100 rounded-xl border border-slate-200 bg-white overflow-hidden">
          {tickets.map((ticket) => (
            <Link
              key={ticket.id}
              to={`/portal/tickets/${ticket.id}`}
              className="flex items-center gap-3 px-4 py-3.5 hover:bg-slate-50 transition-colors"
            >
              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium text-slate-900">{ticket.subject}</p>
                <p className="mt-0.5 text-xs text-slate-400">{formatDate(ticket.created_at)}</p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Badge variant={priorityVariant[ticket.priority]}>{ticket.priority}</Badge>
                <Badge variant={statusVariant[ticket.status]}>{statusLabel[ticket.status]}</Badge>
              </div>
              <ChevronRight className="h-4 w-4 shrink-0 text-slate-300" />
            </Link>
          ))}
        </div>
      )}

      {/* Pagination */}
      {meta && meta.total_pages > 1 && (
        <div className="flex items-center justify-between text-sm text-slate-500">
          <span>
            Page {meta.page} of {meta.total_pages}
          </span>
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
  )
}
