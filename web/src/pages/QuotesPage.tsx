import { useState } from 'react'
import { Plus, FileText } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { Table, type Column } from '@/components/ui/Table'
import { SidePanel } from '@/components/ui/SidePanel'
import { QuoteBuilder, QuoteStatusBadge } from '@/components/omnir/QuoteBuilder'
import { useQuotes, useDeleteQuote } from '@/hooks/useQuotes'
import { formatCurrency, formatDate } from '@/lib/utils'
import type { Quote, QuoteStatus } from '@/api/types'

const STATUS_OPTIONS: { label: string; value: QuoteStatus | '' }[] = [
  { label: 'All statuses', value: '' },
  { label: 'Draft', value: 'draft' },
  { label: 'Sent', value: 'sent' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' },
  { label: 'Expired', value: 'expired' },
]

function QuoteDetail({ quote, onClose }: { quote: Quote; onClose: () => void }) {
  const deleteQuote = useDeleteQuote()
  const [showEdit, setShowEdit] = useState(false)

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
          <div className="flex items-center gap-3">
            <QuoteStatusBadge status={quote.status} />
            <span className="text-2xl font-bold text-indigo-600">
              {formatCurrency(quote.total_cents / 100, quote.currency)}
            </span>
          </div>

          <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
            <dt className="font-medium text-slate-500">Currency</dt>
            <dd className="text-slate-900">{quote.currency}</dd>

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
        </div>
      </SidePanel>

      {showEdit && (
        <QuoteBuilder
          quote={quote}
          onClose={() => setShowEdit(false)}
        />
      )}
    </>
  )
}

export function QuotesPage() {
  const [statusFilter, setStatusFilter] = useState<QuoteStatus | ''>('')
  const [search, setSearch] = useState('')
  const [selectedQuoteId, setSelectedQuoteId] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const { data, isLoading } = useQuotes({
    q: search || undefined,
    status: statusFilter || undefined,
    limit: 50,
  })

  const quotes = data?.data ?? []

  const selectedQuote = selectedQuoteId ? quotes.find((q) => q.id === selectedQuoteId) : null

  const columns: Column<Quote>[] = [
    {
      key: 'title',
      header: 'Quote',
      render: (q) => (
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-slate-400 shrink-0" />
          <span className="font-medium text-slate-900">{q.title}</span>
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
        <span className="font-semibold text-indigo-600">
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
          <p className="text-sm text-slate-500 mt-0.5">{data?.total ?? 0} total</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4 mr-1.5" /> New quote
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search quotes…"
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm w-56 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as QuoteStatus | '')}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        >
          {STATUS_OPTIONS.map((o) => (
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
        <QuoteDetail quote={selectedQuote} onClose={() => setSelectedQuoteId(null)} />
      )}

      {/* Create quote modal */}
      {showCreate && (
        <QuoteBuilder onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}
