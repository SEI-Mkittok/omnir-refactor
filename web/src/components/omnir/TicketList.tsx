import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
  type HeaderGroup,
  type Header,
  type Row,
  type Cell,
} from '@tanstack/react-table'
import { useState } from 'react'
import { ChevronUp, ChevronDown, ChevronsUpDown, Ticket as TicketIcon } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { formatDate, cn } from '@/lib/utils'
import { statusBadgeVariant, statusLabel, priorityBadgeVariant, priorityLabel } from './TicketDetail'
import type { Ticket, SLATrackingStatus } from '@/api/types'

const slaBadgeVariant: Record<SLATrackingStatus, 'green' | 'yellow' | 'red'> = {
  on_track: 'green',
  at_risk: 'yellow',
  breached: 'red',
}

const slaLabel: Record<SLATrackingStatus, string> = {
  on_track: 'On track',
  at_risk: 'At risk',
  breached: 'Breached',
}

interface TicketListProps {
  tickets: Ticket[]
  isLoading: boolean
  page: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
  onRowClick: (ticket: Ticket) => void
  sortBy: string
  sortDir: 'asc' | 'desc'
  onSort: (key: string) => void
}

const columnHelper = createColumnHelper<Ticket>()

function SortIcon({ column, sortBy, sortDir }: { column: string; sortBy: string; sortDir: 'asc' | 'desc' }) {
  if (sortBy !== column) return <ChevronsUpDown className="h-3.5 w-3.5 text-slate-400" />
  return sortDir === 'asc'
    ? <ChevronUp className="h-3.5 w-3.5 text-slate-600" />
    : <ChevronDown className="h-3.5 w-3.5 text-slate-600" />
}

export function TicketList({
  tickets,
  isLoading,
  page,
  totalPages,
  total,
  onPageChange,
  onRowClick,
  sortBy,
  sortDir,
  onSort,
}: TicketListProps) {
  const [sorting, setSorting] = useState<SortingState>([])

  const columns = [
    columnHelper.accessor('subject', {
      header: () => (
        <button
          className="inline-flex items-center gap-1 font-medium hover:text-slate-900"
          onClick={() => onSort('subject')}
        >
          Subject
          <SortIcon column="subject" sortBy={sortBy} sortDir={sortDir} />
        </button>
      ),
      cell: (info) => (
        <span className="font-medium text-slate-900 line-clamp-1">{info.getValue()}</span>
      ),
    }),
    columnHelper.accessor('status', {
      header: () => (
        <button
          className="inline-flex items-center gap-1 font-medium hover:text-slate-900"
          onClick={() => onSort('status')}
        >
          Status
          <SortIcon column="status" sortBy={sortBy} sortDir={sortDir} />
        </button>
      ),
      cell: (info) => {
        const v = info.getValue()
        return <Badge variant={statusBadgeVariant[v]}>{statusLabel[v]}</Badge>
      },
    }),
    columnHelper.accessor('priority', {
      header: () => (
        <button
          className="inline-flex items-center gap-1 font-medium hover:text-slate-900"
          onClick={() => onSort('priority')}
        >
          Priority
          <SortIcon column="priority" sortBy={sortBy} sortDir={sortDir} />
        </button>
      ),
      cell: (info) => {
        const v = info.getValue()
        return <Badge variant={priorityBadgeVariant[v]}>{priorityLabel[v]}</Badge>
      },
    }),
    columnHelper.accessor('sla', {
      id: 'sla',
      header: () => (
        <button
          className="inline-flex items-center gap-1 font-medium hover:text-slate-900"
          onClick={() => onSort('sla_status')}
        >
          SLA
          <SortIcon column="sla_status" sortBy={sortBy} sortDir={sortDir} />
        </button>
      ),
      cell: (info) => {
        const sla = info.getValue()
        if (!sla) return <span className="text-xs text-slate-400">—</span>
        return <Badge variant={slaBadgeVariant[sla.status]}>{slaLabel[sla.status]}</Badge>
      },
      meta: { hideOnMobile: true },
    }),
    columnHelper.accessor('assignee', {
      id: 'assignee',
      header: 'Assignee',
      cell: (info) => {
        const assignee = info.getValue()
        return (
          <span className="text-sm text-slate-600">
            {assignee?.name ?? <span className="text-slate-400">Unassigned</span>}
          </span>
        )
      },
      meta: { hideOnMobile: true },
    }),
    columnHelper.accessor('created_at', {
      header: () => (
        <button
          className="inline-flex items-center gap-1 font-medium hover:text-slate-900"
          onClick={() => onSort('created_at')}
        >
          Created
          <SortIcon column="created_at" sortBy={sortBy} sortDir={sortDir} />
        </button>
      ),
      cell: (info) => (
        <span className="text-xs text-slate-500">{formatDate(info.getValue())}</span>
      ),
      meta: { hideOnMobile: true },
    }),
  ]

  const table = useReactTable({
    data: tickets,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    manualSorting: true, // server-side sorting
    getRowId: (row) => row.id,
  })

  return (
    <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200">
          <thead className="bg-slate-50">
            {table.getHeaderGroups().map((hg: HeaderGroup<Ticket>) => (
              <tr key={hg.id}>
                {hg.headers.map((header: Header<Ticket, unknown>) => (
                  <th
                    key={header.id}
                    className={cn(
                      'px-4 py-3 text-left text-xs text-slate-500',
                      (header.column.columnDef.meta as { hideOnMobile?: boolean } | undefined)?.hideOnMobile && 'hidden sm:table-cell'
                    )}
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="divide-y divide-slate-100">
            {isLoading ? (
              <tr>
                <td colSpan={columns.length} className="py-12 text-center">
                  <div className="flex justify-center">
                    <Spinner className="h-6 w-6 text-[var(--color-primary)]" />
                  </div>
                </td>
              </tr>
            ) : tickets.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="py-12 text-center">
                  <div className="flex flex-col items-center gap-2">
                    <TicketIcon className="h-8 w-8 text-slate-300" />
                    <p className="text-sm font-medium text-slate-500">No tickets found</p>
                    <p className="text-xs text-slate-400">Try adjusting your filters.</p>
                  </div>
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row: Row<Ticket>) => (
                <tr
                  key={row.id}
                  onClick={() => onRowClick(row.original)}
                  className="cursor-pointer transition-colors hover:bg-slate-50"
                >
                  {row.getVisibleCells().map((cell: Cell<Ticket, unknown>) => (
                    <td
                      key={cell.id}
                      className={cn(
                        'px-4 py-3 text-sm',
                        (cell.column.columnDef.meta as { hideOnMobile?: boolean } | undefined)?.hideOnMobile && 'hidden sm:table-cell'
                      )}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 border-t border-slate-200 px-4 py-3">
          <p className="text-sm text-slate-500">
            Page {page} of {totalPages} · {total} total
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
