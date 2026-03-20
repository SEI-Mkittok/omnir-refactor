import React, { useState, useCallback } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Download, ShieldCheck, LogIn, FileOutput } from 'lucide-react'
import { useAuditLog } from '@/hooks/useAuditLog'
import { downloadAuditLogCsv } from '@/api/auditLog'
import { Table, type Column } from '@/components/ui/Table'
import { Button } from '@/components/ui/Button'
import { FilterBar } from '@/components/ui/FilterBar'
import { AuditLogDetailPanel } from '@/components/omnir/AuditLogDetailPanel'
import { formatDate } from '@/lib/utils'
import type { AuditLog, AuditAction, AuditEntityType } from '@/api/types'

const ENTITY_TYPE_OPTIONS = [
  { label: 'Contact', value: 'contact' },
  { label: 'Account', value: 'account' },
  { label: 'Deal', value: 'deal' },
  { label: 'Lead', value: 'lead' },
  { label: 'User', value: 'user' },
  { label: 'View', value: 'view' },
]

const ACTION_OPTIONS = [
  { label: 'Created', value: 'created' },
  { label: 'Updated', value: 'updated' },
  { label: 'Deleted', value: 'deleted' },
  { label: 'Converted', value: 'converted' },
  { label: 'Login', value: 'login' },
  { label: 'Export', value: 'export' },
]

const ACTION_COLORS: Record<AuditAction, string> = {
  created: 'bg-green-100 text-green-700',
  updated: 'bg-blue-100 text-blue-700',
  deleted: 'bg-red-100 text-red-700',
  converted: 'bg-purple-100 text-purple-700',
  login: 'bg-slate-100 text-slate-700',
  export: 'bg-amber-100 text-amber-700',
}

const ACTION_ICONS: Partial<Record<AuditAction, React.ReactNode>> = {
  login: <LogIn className="h-3 w-3" />,
  export: <FileOutput className="h-3 w-3" />,
}

export function AuditLogPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  // Derive filter state from URL query params (persisted in URL)
  const entityType = searchParams.get('entityType') ?? ''
  const action = searchParams.get('action') ?? ''
  const from = searchParams.get('from') ?? ''
  const to = searchParams.get('to') ?? ''

  const [page, setPage] = useState(1)
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [selectedId, setSelectedId] = useState<string | null>(null)

  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const setParam = useCallback(
    (key: string, value: string) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev)
        if (value) {
          next.set(key, value)
        } else {
          next.delete(key)
        }
        return next
      })
      setPage(1)
    },
    [setSearchParams]
  )

  const { data, isLoading } = useAuditLog({
    entityType: (entityType as AuditEntityType) || undefined,
    action: (action as AuditAction) || undefined,
    from: from || undefined,
    to: to || undefined,
    page,
    limit: 50,
  })

  const entries = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  const columns: Column<AuditLog>[] = [
    {
      key: 'created_at',
      header: 'Timestamp',
      sortable: true,
      render: (e) => (
        <span className="text-sm text-slate-600 whitespace-nowrap">{formatDate(e.created_at)}</span>
      ),
    },
    {
      key: 'action',
      header: 'Action',
      sortable: true,
      render: (e) => (
        <span
          className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium capitalize ${ACTION_COLORS[e.action] ?? 'bg-slate-100 text-slate-700'}`}
        >
          {ACTION_ICONS[e.action]}
          {e.action}
        </span>
      ),
    },
    {
      key: 'entity_type',
      header: 'Entity Type',
      sortable: true,
      render: (e) => (
        <span className="text-sm text-slate-700 capitalize">{e.entity_type}</span>
      ),
    },
    {
      key: 'entity_name',
      header: 'Entity',
      hideOnMobile: true,
      render: (e) => (
        <span className="text-sm text-slate-600">{e.entity_name ?? e.entity_id ?? '—'}</span>
      ),
    },
    {
      key: 'actor',
      header: 'Actor',
      render: (e) => (
        <span className="text-sm text-slate-600 font-mono text-xs">
          {e.user_id ?? e.agent_id ?? <span className="italic text-slate-400">System</span>}
        </span>
      ),
    },
    {
      key: 'ip_address',
      header: 'IP Address',
      hideOnMobile: true,
      render: (e) => (
        <span className="text-sm text-slate-500 font-mono text-xs">{e.ip_address ?? '—'}</span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Audit Log</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total events` : 'Loading…'}
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() =>
            downloadAuditLogCsv({
              entityType: (entityType as AuditEntityType) || undefined,
              action: (action as AuditAction) || undefined,
              from: from || undefined,
              to: to || undefined,
            })
          }
        >
          <Download className="h-4 w-4" />
          Export CSV
        </Button>
      </div>

      {/* Date range filters */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2">
          <label className="text-sm text-slate-600 whitespace-nowrap">From</label>
          <input
            type="date"
            value={from}
            onChange={(e) => setParam('from', e.target.value ? `${e.target.value}T00:00:00Z` : '')}
            className="h-9 rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-indigo-500 text-slate-700"
          />
        </div>
        <div className="flex items-center gap-2">
          <label className="text-sm text-slate-600 whitespace-nowrap">To</label>
          <input
            type="date"
            value={to ? to.slice(0, 10) : ''}
            onChange={(e) => setParam('to', e.target.value ? `${e.target.value}T23:59:59Z` : '')}
            className="h-9 rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-indigo-500 text-slate-700"
          />
        </div>
      </div>

      {/* Entity type + action filters */}
      <FilterBar
        searchValue=""
        onSearchChange={() => {}}
        searchPlaceholder="Search…"
        filters={[
          {
            label: 'Entity Type',
            value: entityType,
            options: ENTITY_TYPE_OPTIONS,
            onChange: (v) => setParam('entityType', v),
          },
          {
            label: 'Action',
            value: action,
            options: ACTION_OPTIONS,
            onChange: (v) => setParam('action', v),
          },
        ]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={entries}
          isLoading={isLoading}
          sortBy={sortBy}
          sortDir={sortDir}
          onSort={handleSort}
          onRowClick={(e) => setSelectedId(e.id)}
          emptyIcon={ShieldCheck}
          emptyTitle="No audit events found"
          emptyDescription="Try adjusting your filters or date range."
          keyExtractor={(e) => e.id}
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

      {/* Detail drawer */}
      <AuditLogDetailPanel entryId={selectedId} onClose={() => setSelectedId(null)} />
    </div>
  )
}
