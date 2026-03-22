import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { usePortalTickets } from '@/hooks/usePortal'
import { useDebounce } from '@/hooks/useDebounce'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
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

  const { data, isLoading, isError } = usePortalTickets({
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
        <h1 className="text-2xl font-bold text-slate-900">My Tickets</h1>
        <Button asChild>
          <Link to="/portal/tickets/new">
            <Plus className="h-4 w-4" />
            New Ticket
          </Link>
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <Input
          type="search"
          placeholder="Search tickets…"
          aria-label="Search tickets"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          className="flex-1 min-w-48"
        />
        <div className="flex flex-wrap gap-1.5" role="group" aria-label="Filter by status">
          {STATUS_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => { setStatus(opt.value as PortalTicketStatus | ''); setPage(1) }}
              className={[
                'rounded-full px-3 py-1 text-sm font-medium transition-colors',
                status === opt.value
                  ? 'bg-[#1B3A4B] text-white'
                  : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50',
              ].join(' ')}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Loading skeletons */}
      {isLoading && (
        <div aria-busy="true" className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-14 rounded-xl bg-slate-100 animate-pulse" aria-hidden="true" />
          ))}
        </div>
      )}

      {/* Error */}
      {isError && (
        <div role="alert" className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          Failed to load tickets. Try refreshing.
        </div>
      )}

      {/* Data */}
      {!isLoading && !isError && (
        <>
          {tickets.length === 0 ? (
            <div className="rounded-xl border border-dashed border-slate-300 bg-white py-16 text-center">
              <p className="text-sm text-slate-500">
                {status || search ? 'No tickets match your filters.' : 'No tickets yet.'}
              </p>
              {!status && !search && (
                <Button variant="outline" size="sm" className="mt-4" asChild>
                  <Link to="/portal/tickets/new">Submit your first ticket</Link>
                </Button>
              )}
            </div>
          ) : (
            <>
              {/* Desktop: Table */}
              <div className="hidden md:block overflow-hidden rounded-xl border border-slate-200 bg-white">
                <table className="w-full text-sm">
                  <caption className="sr-only">My support tickets</caption>
                  <thead className="border-b border-slate-200 bg-slate-50">
                    <tr>
                      <th scope="col" className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Subject</th>
                      <th scope="col" className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Status</th>
                      <th scope="col" className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Priority</th>
                      <th scope="col" className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Created</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {tickets.map((ticket) => (
                      <tr key={ticket.id} role="row" className="hover:bg-slate-50 transition-colors">
                        <td className="px-4 py-3.5 font-medium text-slate-900">
                          <Link
                            to={`/portal/tickets/${ticket.id}`}
                            className="hover:text-[#1B3A4B] hover:underline focus:outline-none focus:underline"
                          >
                            {ticket.subject}
                          </Link>
                        </td>
                        <td className="px-4 py-3.5">
                          <Badge variant={statusVariant[ticket.status]} aria-label={`Status: ${statusLabel[ticket.status]}`}>
                            {statusLabel[ticket.status]}
                          </Badge>
                        </td>
                        <td className="px-4 py-3.5">
                          <Badge variant={priorityVariant[ticket.priority]} aria-label={`Priority: ${ticket.priority}`}>
                            {ticket.priority}
                          </Badge>
                        </td>
                        <td className="px-4 py-3.5 text-slate-500">{formatDate(ticket.created_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Mobile: Card list */}
              <div className="md:hidden space-y-2">
                {tickets.map((ticket) => (
                  <Link
                    key={ticket.id}
                    to={`/portal/tickets/${ticket.id}`}
                    className="block rounded-xl border border-slate-200 bg-white p-4 hover:bg-slate-50 transition-colors"
                  >
                    <p className="font-semibold text-slate-900 truncate">{ticket.subject}</p>
                    <div className="mt-2 flex items-center justify-between gap-2">
                      <div className="flex gap-2">
                        <Badge variant={statusVariant[ticket.status]} aria-label={`Status: ${statusLabel[ticket.status]}`}>
                          {statusLabel[ticket.status]}
                        </Badge>
                        <Badge variant={priorityVariant[ticket.priority]} aria-label={`Priority: ${ticket.priority}`}>
                          {ticket.priority}
                        </Badge>
                      </div>
                      <span className="text-xs text-slate-400 shrink-0">{formatDate(ticket.created_at)}</span>
                    </div>
                  </Link>
                ))}
              </div>
            </>
          )}

          {/* Pagination */}
          {meta && meta.total_pages > 1 && (
            <div className="flex items-center justify-between text-sm text-slate-500">
              <span>Page {meta.page} of {meta.total_pages}</span>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                  Previous
                </Button>
                <Button variant="outline" size="sm" disabled={page >= meta.total_pages} onClick={() => setPage((p) => p + 1)}>
                  Next
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
