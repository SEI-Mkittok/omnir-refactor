import { useState, useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Plus, Building2, Users, Download, Upload, ExternalLink } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useAccounts } from '@/hooks/useAccounts'
import { useUpdateView } from '@/hooks/useViews'
import { ViewPinBar } from '@/components/omnir/ViewPinBar'
import { AccountForm } from '@/components/omnir/AccountForm'
import { ImportModal } from '@/components/omnir/ImportModal'
import { downloadExportCsv } from '@/api/importExport'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import type { Account, SavedView } from '@/api/types'

// ── Industry Badge ────────────────────────────────────────────────────────────

const INDUSTRY_STYLES: Record<string, { bg: string; text: string }> = {
  aerospace:     { bg: '#EFF6FF', text: '#1D4ED8' },
  fintech:       { bg: '#F0FDF4', text: '#15803D' },
  healthcare:    { bg: '#FFF7ED', text: '#C2410C' },
  retail:        { bg: '#FAF5FF', text: '#7C3AED' },
  technology:    { bg: '#F0F9FF', text: '#0369A1' },
  manufacturing: { bg: '#FEF3C7', text: '#92400E' },
}

function IndustryBadge({ industry }: { industry: string | undefined }) {
  if (!industry) return <span style={{ color: 'var(--text-label)' }}>—</span>
  const key = industry.toLowerCase()
  const style = INDUSTRY_STYLES[key] ?? { bg: 'var(--surface-app)', text: 'var(--text-label)' }
  return (
    <span
      className="inline-flex items-center font-semibold uppercase"
      style={{
        background: style.bg,
        color: style.text,
        fontSize: 11,
        letterSpacing: '0.04em',
        padding: '2px 10px',
        borderRadius: 'var(--radius-pill)',
        border: `1px solid ${style.text}4D`,
      }}
    >
      {industry.toUpperCase()}
    </span>
  )
}

// ── Activity Status Dot ───────────────────────────────────────────────────────

function deriveStatus(account: Account): 'active' | 'pending' | 'archived' {
  const updatedAt = new Date(account.updated_at).getTime()
  const createdAt = new Date(account.created_at).getTime()
  const now = Date.now()
  const dayMs = 86400000
  if (now - updatedAt < 90 * dayMs) {
    if (now - createdAt < 7 * dayMs) return 'pending'
    return 'active'
  }
  return 'archived'
}

const STATUS_STYLES = {
  active:   { dot: '#22C55E', label: 'Active' },
  pending:  { dot: '#F59E0B', label: 'Pending' },
  archived: { dot: '#9CA3AF', label: 'Archived' },
}

function ActivityStatusDot({ account }: { account: Account }) {
  const status = deriveStatus(account)
  const { dot, label } = STATUS_STYLES[status]
  return (
    <span className="flex items-center gap-1.5">
      <span
        className="inline-block rounded-full shrink-0"
        style={{ width: 8, height: 8, background: dot }}
      />
      <span className="text-sm" style={{ color: 'var(--text-primary)' }}>
        {label}
      </span>
    </span>
  )
}

// ── Skeleton row ──────────────────────────────────────────────────────────────

function SkeletonRow() {
  return (
    <tr style={{ height: 60 }}>
      {[3, 2, 2, 1, 1.5, 1.5].map((w, i) => (
        <td key={i} className="px-4 py-3">
          <div
            className="h-3 animate-pulse rounded"
            style={{ width: `${w * 50}px`, background: 'var(--border-subtle)' }}
          />
        </td>
      ))}
    </tr>
  )
}

// ── Filter Select ─────────────────────────────────────────────────────────────

function FilterSelect({
  value,
  onChange,
  options,
  placeholder,
}: {
  value: string
  onChange: (v: string) => void
  options: { label: string; value: string }[]
  placeholder: string
}) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="rounded-md border text-sm font-medium transition-colors"
      style={{
        height: 36,
        padding: '0 12px',
        borderColor: value ? 'var(--color-primary)' : 'var(--border-default)',
        color: value ? 'var(--color-primary)' : 'var(--text-secondary)',
        background: value ? 'var(--color-primary-light)' : 'var(--surface-card)',
        cursor: 'pointer',
        outline: 'none',
      }}
    >
      <option value="">{placeholder}</option>
      {options.map((o) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

const INDUSTRY_OPTIONS = [
  { label: 'Aerospace', value: 'aerospace' },
  { label: 'Fintech', value: 'fintech' },
  { label: 'Healthcare', value: 'healthcare' },
  { label: 'Retail', value: 'retail' },
  { label: 'Technology', value: 'technology' },
  { label: 'Manufacturing', value: 'manufacturing' },
  { label: 'Other', value: 'other' },
]

const STATUS_OPTIONS = [
  { label: 'Active', value: 'active' },
  { label: 'Pending', value: 'pending' },
  { label: 'Archived', value: 'archived' },
]

const TABLE_HEADERS = ['Account', 'Industry', 'Website', 'Contacts', 'Status', 'Created']

export function AccountsPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const [industry, setIndustry] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [showImport, setShowImport] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [activeView, setActiveView] = useState<SavedView | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)

  const updateView = useUpdateView()
  const debouncedSearch = useDebounce(search, 200)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const currentFilters = {
    search: debouncedSearch || undefined,
    industry: industry || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  }

  const applyViewFilters = (view: SavedView) => {
    setActiveView(view)
    setHasUnsavedChanges(false)
    setSearch((view.filters.search as string) ?? '')
    setIndustry((view.filters.industry as string) ?? '')
    setSortKey(view.filters.sort_by ? `${view.filters.sort_by}:${view.filters.sort_dir ?? 'asc'}` : 'created_at:desc')
    setPage(1)
  }

  const markChanged = () => { if (activeView) setHasUnsavedChanges(true) }

  const handleUpdateView = async (viewId: string) => {
    await updateView.mutateAsync({ id: viewId, payload: { filters: currentFilters } })
    setHasUnsavedChanges(false)
  }

  const { data, isLoading, isError, refetch } = useAccounts({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    industry: industry || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const allAccounts = data?.data ?? []
  // Client-side status filter (backend doesn't support it)
  const accounts = statusFilter
    ? allAccounts.filter((a) => deriveStatus(a) === statusFilter)
    : allAccounts
  const meta = data?.meta

  const hasFilters = !!(search || industry || statusFilter)

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
    markChanged()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Opening with ?openId= from other pages
  const _openId = searchParams.get('openId')

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Header */}
      <div className="flex flex-wrap items-start justify-between gap-4 mb-5">
        <h1
          className="font-bold"
          style={{ fontSize: 36, color: 'var(--text-primary)', letterSpacing: '-0.01em', lineHeight: 1.1 }}
        >
          {activeView ? activeView.name : 'Accounts.'}
        </h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() =>
              downloadExportCsv('accounts', {
                ...(debouncedSearch ? { q: debouncedSearch } : {}),
                ...(industry ? { industry } : {}),
                ...(sortBy ? { sort: sortBy, order: sortDir } : {}),
              })
            }
            className="flex items-center gap-1.5 rounded-md border text-sm font-medium transition-colors"
            style={{
              height: 36,
              padding: '0 14px',
              borderColor: 'var(--border-default)',
              color: 'var(--text-secondary)',
              background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <Download className="h-3.5 w-3.5" />
            Export
          </button>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 rounded-md text-sm font-semibold text-white transition-colors"
            style={{
              height: 36,
              padding: '0 16px',
              background: 'var(--color-primary)',
              border: 'none',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary-hover)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary)' }}
          >
            <Plus className="h-4 w-4" />
            New Account
          </button>
        </div>
      </div>

      {/* View pin bar */}
      <ViewPinBar
        entityType="accounts"
        activeViewId={activeView?.id ?? null}
        hasUnsavedChanges={hasUnsavedChanges}
        currentFilters={currentFilters}
        onSelectView={applyViewFilters}
        onClearView={() => { setActiveView(null); setHasUnsavedChanges(false) }}
        onViewSaved={(view) => setActiveView(view)}
        onUpdateView={handleUpdateView}
      />

      {/* Filter bar */}
      <div className="flex flex-wrap items-center gap-2 mb-4">
        <FilterSelect
          value={industry}
          onChange={(v) => { setIndustry(v); setPage(1); markChanged() }}
          options={INDUSTRY_OPTIONS}
          placeholder="All Industries"
        />
        <FilterSelect
          value={statusFilter}
          onChange={(v) => { setStatusFilter(v); setPage(1) }}
          options={STATUS_OPTIONS}
          placeholder="All Statuses"
        />
        {hasFilters && (
          <button
            className="text-sm"
            style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
            onClick={() => { setSearch(''); setIndustry(''); setStatusFilter(''); setPage(1) }}
          >
            Clear filters
          </button>
        )}
        <div className="ml-auto">
          <input
            type="search"
            placeholder="Search accounts…"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); markChanged() }}
            className="rounded-full border text-sm px-4 transition-colors focus:outline-none"
            style={{
              height: 36,
              width: 220,
              borderColor: 'var(--border-default)',
              color: 'var(--text-primary)',
              background: 'var(--surface-card)',
            }}
            onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--border-focus)' }}
            onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)' }}
          />
        </div>
      </div>

      {/* Table */}
      <div
        className="rounded-xl border overflow-hidden"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      >
        {isError ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="text-sm font-medium mb-1" style={{ color: 'var(--text-primary)' }}>
              ⚠ Something went wrong.
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--text-secondary)' }}>
              Couldn&apos;t load accounts. Check your connection and try again.
            </p>
            <button
              onClick={() => refetch()}
              className="text-sm font-medium"
              style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
            >
              Retry
            </button>
          </div>
        ) : (
          <table className="w-full" aria-label="Accounts">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-default)' }}>
                {TABLE_HEADERS.map((h) => (
                  <th
                    key={h}
                    className="px-4 py-3 text-left font-semibold uppercase tracking-widest text-left"
                    style={{
                      fontSize: 11,
                      color: 'var(--text-label)',
                      letterSpacing: 'var(--letter-spacing-label)',
                    }}
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} />)
              ) : accounts.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-16 text-center">
                    <Building2
                      className="mx-auto mb-3"
                      style={{ width: 48, height: 48, color: 'var(--text-label)' }}
                    />
                    <p className="text-sm font-medium mb-1" style={{ color: 'var(--text-primary)' }}>
                      {hasFilters ? 'No accounts match your filters.' : 'No accounts yet.'}
                    </p>
                    <p className="text-xs mb-4" style={{ color: 'var(--text-secondary)' }}>
                      {hasFilters
                        ? ''
                        : 'Add your first account to start tracking company relationships.'}
                    </p>
                    {hasFilters && (
                      <button
                        className="text-sm font-medium"
                        style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
                        onClick={() => { setSearch(''); setIndustry(''); setStatusFilter(''); setPage(1) }}
                      >
                        Clear filters
                      </button>
                    )}
                  </td>
                </tr>
              ) : (
                accounts.map((account) => (
                  <AccountRow
                    key={account.id}
                    account={account}
                    onClick={() => navigate(`/accounts/${account.id}`)}
                  />
                ))
              )}
            </tbody>
          </table>
        )}

        {/* Pagination */}
        {meta && meta.total_pages > 1 && (
          <div
            className="flex flex-wrap items-center justify-between gap-3 px-4 py-3 border-t"
            style={{ borderColor: 'var(--border-default)' }}
          >
            <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
              Showing {(meta.page - 1) * meta.per_page + 1}–
              {Math.min(meta.page * meta.per_page, meta.total)} of {meta.total} accounts
            </p>
            <div className="flex items-center gap-1">
              <PageButton
                label="←"
                disabled={page <= 1}
                active={false}
                onClick={() => setPage((p) => p - 1)}
              />
              {Array.from({ length: Math.min(meta.total_pages, 7) }, (_, i) => {
                const pg = i + 1
                return (
                  <PageButton
                    key={pg}
                    label={String(pg)}
                    disabled={false}
                    active={page === pg}
                    onClick={() => setPage(pg)}
                  />
                )
              })}
              <PageButton
                label="→"
                disabled={page >= meta.total_pages}
                active={false}
                onClick={() => setPage((p) => p + 1)}
              />
            </div>
          </div>
        )}
      </div>

      {showCreate && <AccountForm onClose={() => setShowCreate(false)} />}

      <ImportModal
        open={showImport}
        onClose={() => setShowImport(false)}
        entity="accounts"
        entityLabel="Accounts"
      />
    </div>
  )
}

// ── Table Row ─────────────────────────────────────────────────────────────────

function AccountRow({ account, onClick }: { account: Account; onClick: () => void }) {
  const [hovered, setHovered] = useState(false)
  return (
    <tr
      tabIndex={0}
      style={{
        height: 60,
        cursor: 'pointer',
        background: hovered ? 'var(--surface-app)' : 'transparent',
        transition: 'background 150ms',
      }}
      onClick={onClick}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') onClick() }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      aria-label={`${account.name}, ${account.industry ?? 'No industry'}`}
    >
      {/* Account */}
      <td className="px-4 py-3">
        <div className="flex items-center gap-2.5">
          <div
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md"
            style={{ background: 'var(--color-primary-light)' }}
          >
            <Building2 className="h-4 w-4" style={{ color: 'var(--color-primary)' }} />
          </div>
          <div className="min-w-0">
            <p className="text-sm font-bold truncate" style={{ color: 'var(--text-primary)' }}>
              {account.name}
            </p>
            {account.domain && (
              <p className="text-xs truncate" style={{ color: 'var(--text-label)' }}>
                {account.domain}
              </p>
            )}
          </div>
        </div>
      </td>

      {/* Industry */}
      <td className="px-4 py-3">
        <IndustryBadge industry={account.industry} />
      </td>

      {/* Website */}
      <td className="px-4 py-3 hidden md:table-cell">
        {account.domain ? (
          <a
            href={`https://${account.domain}`}
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1 text-sm hover:underline"
            style={{ color: 'var(--text-secondary)' }}
            onClick={(e) => e.stopPropagation()}
          >
            <span className="truncate max-w-[140px]">{account.domain}</span>
            <ExternalLink className="h-3 w-3 shrink-0" />
          </a>
        ) : (
          <span style={{ color: 'var(--text-label)' }}>—</span>
        )}
      </td>

      {/* Contacts */}
      <td className="px-4 py-3 hidden sm:table-cell">
        <div className="flex items-center gap-1">
          <Users className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
          <span
            className="text-sm font-medium"
            style={{
              color:
                (account.contacts?.length ?? 0) === 0
                  ? 'var(--text-label)'
                  : 'var(--text-primary)',
            }}
          >
            {account.contacts?.length ?? 0}
          </span>
        </div>
      </td>

      {/* Status */}
      <td className="px-4 py-3">
        <ActivityStatusDot account={account} />
      </td>

      {/* Created */}
      <td className="px-4 py-3 hidden lg:table-cell">
        <span
          className="text-sm"
          title={formatDate(account.created_at)}
          style={{ color: 'var(--text-secondary)' }}
        >
          {formatRelativeTime(account.created_at)}
        </span>
      </td>
    </tr>
  )
}

// ── Pagination Button ─────────────────────────────────────────────────────────

function PageButton({
  label,
  disabled,
  active,
  onClick,
}: {
  label: string
  disabled: boolean
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      disabled={disabled}
      onClick={onClick}
      className="flex items-center justify-center rounded-md text-sm font-medium transition-colors"
      style={{
        width: 32,
        height: 32,
        border: '1px solid',
        borderColor: active ? 'var(--color-primary)' : 'var(--border-default)',
        background: active ? 'var(--color-primary)' : 'transparent',
        color: active ? '#FFFFFF' : disabled ? 'var(--text-disabled)' : 'var(--text-secondary)',
        cursor: disabled ? 'not-allowed' : 'pointer',
      }}
    >
      {label}
    </button>
  )
}
