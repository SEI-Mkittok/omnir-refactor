import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus, UserPlus, UserCheck, Upload } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useLeads, useCreateLead } from '@/hooks/useLeads'
import { useUpdateView } from '@/hooks/useViews'
import { ViewPinBar } from '@/components/omnir/ViewPinBar'
import { ImportModal } from '@/components/omnir/ImportModal'
import { LeadDetailPanel } from '@/components/omnir/LeadDetailPanel'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { CustomFieldFormSection } from '@/components/omnir/CustomFieldRenderer'
import type { Lead, LeadStatus, CreateLeadRequest, CustomFieldValues, SavedView } from '@/api/types'

// ── Lead Score Indicator ──────────────────────────────────────────────────────

function scoreHeat(score: number): 'hot' | 'warm' | 'cold' {
  if (score >= 80) return 'hot'
  if (score >= 50) return 'warm'
  return 'cold'
}

const SCORE_FILL: Record<string, string> = {
  hot:  'var(--color-success)',
  warm: 'var(--color-warning)',
  cold: 'var(--color-neutral)',
}

function LeadScoreIndicator({ score }: { score: number | null | undefined }) {
  if (score == null || score === 0) {
    return (
      <span className="text-sm" style={{ color: 'var(--text-label)' }}>
        —
      </span>
    )
  }
  const heat = scoreHeat(score)
  const heatLabel = heat === 'hot' ? 'Hot' : heat === 'warm' ? 'Warm' : 'Cold'
  return (
    <span className="flex items-center gap-2">
      <span aria-hidden="true" className="flex items-center gap-2">
        <span
          className="inline-block rounded-full overflow-hidden"
          style={{ width: 60, height: 6, background: 'var(--border-default)' }}
        >
          <span
            className="block h-full rounded-full"
            style={{ width: `${score}%`, background: SCORE_FILL[heat] }}
          />
        </span>
        <span className="text-sm font-semibold" style={{ color: 'var(--text-primary)', minWidth: 28 }}>
          {score}
        </span>
      </span>
      <span className="sr-only">
        Lead score: {score} out of 100 ({heatLabel})
      </span>
    </span>
  )
}

// ── Conversion Status Badge ───────────────────────────────────────────────────

const STATUS_BADGE_STYLES: Record<LeadStatus | 'unqualified', { bg: string; text: string; label: string }> = {
  new:         { bg: '#EFF6FF', text: '#1D4ED8', label: 'NEW' },
  contacted:   { bg: '#F0FDF4', text: '#15803D', label: 'CONTACTED' },
  qualified:   { bg: '#0D9488', text: '#FFFFFF', label: 'QUALIFIED' },
  converted:   { bg: '#22C55E', text: '#FFFFFF', label: 'CONVERTED' },
  unqualified: { bg: '#9CA3AF', text: '#FFFFFF', label: 'DEAD' },
}

function ConversionStatusBadge({ status, lead: _lead }: { status: LeadStatus; lead: Lead }) {
  const s = STATUS_BADGE_STYLES[status] ?? STATUS_BADGE_STYLES.unqualified
  return (
    <span
      className="inline-flex items-center gap-1 font-semibold uppercase"
      style={{
        background: s.bg,
        color: s.text,
        fontSize: 11,
        letterSpacing: '0.04em',
        padding: '2px 10px',
        borderRadius: 'var(--radius-pill)',
      }}
      aria-label={`Status: ${s.label}`}
    >
      {status === 'converted' && (
        <UserCheck className="h-3 w-3 shrink-0" style={{ color: 'var(--color-success)' }} />
      )}
      {s.label}
    </span>
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

// ── Skeleton Row ──────────────────────────────────────────────────────────────

function SkeletonRow() {
  return (
    <tr style={{ height: 60 }}>
      {[3, 2, 1.5, 2, 2, 1.5].map((w, i) => (
        <td key={i} className="px-4 py-3">
          <div
            className="h-3 animate-pulse rounded"
            style={{ width: `${w * 45}px`, background: 'var(--border-subtle)' }}
          />
        </td>
      ))}
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
      className="flex items-center justify-center rounded-md text-sm font-medium"
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

// ── Create Lead Form ──────────────────────────────────────────────────────────

const INITIAL_FORM: CreateLeadRequest = {
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  company: '',
  lead_source: '',
  status: 'new',
}

function LeadForm({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [form, setForm] = useState<CreateLeadRequest>(INITIAL_FORM)
  const [errors, setErrors] = useState<Partial<Record<keyof CreateLeadRequest, string>>>({})
  const [customFieldValues, setCustomFieldValues] = useState<CustomFieldValues>({})
  const createLead = useCreateLead()
  const { data: customFields = [] } = useCustomFieldDefinitions('lead')

  const set = (field: keyof CreateLeadRequest) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }))

  const validate = () => {
    const errs: typeof errors = {}
    if (!form.first_name.trim()) errs.first_name = 'Required'
    if (!form.last_name.trim()) errs.last_name = 'Required'
    if (!form.email.trim()) errs.email = 'Required'
    else if (!/\S+@\S+\.\S+/.test(form.email)) errs.email = 'Invalid email'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    await createLead.mutateAsync({
      ...form,
      phone: form.phone || undefined,
      company: form.company || undefined,
      lead_source: form.lead_source || undefined,
      ...(Object.keys(customFieldValues).length ? { custom_fields: customFieldValues } : {}),
    } as CreateLeadRequest & { custom_fields?: CustomFieldValues })
    setForm(INITIAL_FORM)
    setErrors({})
    setCustomFieldValues({})
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) { setForm(INITIAL_FORM); setErrors({}); setCustomFieldValues({}); onClose() }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>New Lead</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                First name <span className="text-red-500">*</span>
              </label>
              <Input value={form.first_name} onChange={set('first_name')} placeholder="Jane" aria-invalid={!!errors.first_name} />
              {errors.first_name && <p className="mt-0.5 text-xs text-red-500">{errors.first_name}</p>}
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Last name <span className="text-red-500">*</span>
              </label>
              <Input value={form.last_name} onChange={set('last_name')} placeholder="Smith" aria-invalid={!!errors.last_name} />
              {errors.last_name && <p className="mt-0.5 text-xs text-red-500">{errors.last_name}</p>}
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
              Email <span className="text-red-500">*</span>
            </label>
            <Input type="email" value={form.email} onChange={set('email')} placeholder="jane@example.com" aria-invalid={!!errors.email} />
            {errors.email && <p className="mt-0.5 text-xs text-red-500">{errors.email}</p>}
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Company</label>
              <Input value={form.company ?? ''} onChange={set('company')} placeholder="Acme Inc." />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Source</label>
              <Input value={form.lead_source ?? ''} onChange={set('lead_source')} placeholder="Website, Referral…" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Phone</label>
              <Input type="tel" value={form.phone ?? ''} onChange={set('phone')} placeholder="+1 555 000 0000" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Status</label>
              <select
                value={form.status}
                onChange={set('status')}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
              >
                <option value="new">New</option>
                <option value="contacted">Contacted</option>
                <option value="qualified">Qualified</option>
                <option value="unqualified">Unqualified</option>
              </select>
            </div>
          </div>
          <CustomFieldFormSection
            fields={customFields}
            values={customFieldValues}
            onChange={setCustomFieldValues}
          />
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={createLead.isPending}>
              Create Lead
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Score Options ─────────────────────────────────────────────────────────────

const STATUS_OPTIONS = [
  { label: 'New', value: 'new' },
  { label: 'Contacted', value: 'contacted' },
  { label: 'Qualified', value: 'qualified' },
  { label: 'Converted', value: 'converted' },
  { label: 'Dead', value: 'unqualified' },
]

const SCORE_RANGE_OPTIONS = [
  { label: 'Hot (80–100)', value: '80:100' },
  { label: 'Warm (50–79)', value: '50:79' },
  { label: 'Cold (0–49)', value: '0:49' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
  { label: 'Name A–Z', value: 'last_name:asc' },
  { label: 'Score High–Low', value: 'lead_score:desc' },
]

const TABLE_HEADERS = ['Lead', 'Company', 'Score', 'Status', 'Assigned To', 'Created']

// ── Main Page ─────────────────────────────────────────────────────────────────

export function LeadsPage() {
  const [searchParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [scoreRange, setScoreRange] = useState('')
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

  const [scoreMin, scoreMax] = scoreRange
    ? scoreRange.split(':').map(Number)
    : [undefined, undefined]

  const currentFilters = {
    search: debouncedSearch || undefined,
    status: status || undefined,
    score_range: scoreRange || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  }

  const applyViewFilters = (view: SavedView) => {
    setActiveView(view)
    setHasUnsavedChanges(false)
    setSearch((view.filters.search as string) ?? '')
    setStatus((view.filters.status as string) ?? '')
    setScoreRange((view.filters.score_range as string) ?? '')
    setSortKey(view.filters.sort_by ? `${view.filters.sort_by}:${view.filters.sort_dir ?? 'asc'}` : 'created_at:desc')
    setPage(1)
  }

  const markChanged = () => { if (activeView) setHasUnsavedChanges(true) }

  const handleUpdateView = async (viewId: string) => {
    await updateView.mutateAsync({ id: viewId, payload: { filters: currentFilters } })
    setHasUnsavedChanges(false)
  }

  const { data, isLoading, isError, refetch } = useLeads({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    status: (status as LeadStatus) || undefined,
    score_min: scoreMin,
    score_max: scoreMax,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const leads = data?.data ?? []
  const meta = data?.meta
  const hasFilters = !!(search || status || scoreRange)

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Header */}
      <div className="flex flex-wrap items-start justify-between gap-4 mb-5">
        <h1
          className="font-bold"
          style={{ fontSize: 36, color: 'var(--text-primary)', letterSpacing: '-0.01em', lineHeight: 1.1 }}
        >
          {activeView ? activeView.name : 'Leads.'}
        </h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowImport(true)}
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
            <Upload className="h-3.5 w-3.5" />
            Import
          </button>
          <button
            onClick={() => setShowForm(true)}
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
            New Lead
          </button>
        </div>
      </div>

      {/* View pin bar */}
      <ViewPinBar
        entityType="leads"
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
          value={status}
          onChange={(v) => { setStatus(v); setPage(1); markChanged() }}
          options={STATUS_OPTIONS}
          placeholder="All Statuses"
        />
        <FilterSelect
          value={scoreRange}
          onChange={(v) => { setScoreRange(v); setPage(1); markChanged() }}
          options={SCORE_RANGE_OPTIONS}
          placeholder="All Scores"
        />
        <FilterSelect
          value={sortKey}
          onChange={(v) => { setSortKey(v); setPage(1); markChanged() }}
          options={SORT_OPTIONS}
          placeholder="Sort"
        />
        {hasFilters && (
          <button
            className="text-sm"
            style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
            onClick={() => { setSearch(''); setStatus(''); setScoreRange(''); setPage(1) }}
          >
            Clear filters
          </button>
        )}
        <div className="ml-auto">
          <input
            type="search"
            placeholder="Search leads…"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); markChanged() }}
            className="rounded-full border text-sm px-4 focus:outline-none"
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
              Couldn&apos;t load leads. Check your connection and try again.
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
          <table className="w-full" aria-label="Leads">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-default)' }}>
                {TABLE_HEADERS.map((h) => (
                  <th
                    key={h}
                    className="px-4 py-3 text-left font-semibold uppercase"
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
              ) : leads.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-16 text-center">
                    <UserPlus
                      className="mx-auto mb-3"
                      style={{ width: 48, height: 48, color: 'var(--text-label)' }}
                    />
                    <p className="text-sm font-medium mb-1" style={{ color: 'var(--text-primary)' }}>
                      {hasFilters ? 'No leads match your filters.' : 'No leads yet.'}
                    </p>
                    <p className="text-xs mb-4" style={{ color: 'var(--text-secondary)' }}>
                      {hasFilters
                        ? ''
                        : 'Import leads or add them manually to start building your pipeline.'}
                    </p>
                    {hasFilters ? (
                      <button
                        className="text-sm font-medium"
                        style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
                        onClick={() => { setSearch(''); setStatus(''); setScoreRange(''); setPage(1) }}
                      >
                        Clear filters
                      </button>
                    ) : (
                      <button
                        onClick={() => setShowForm(true)}
                        className="text-sm font-medium"
                        style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
                      >
                        + New Lead
                      </button>
                    )}
                  </td>
                </tr>
              ) : (
                leads.map((lead) => (
                  <LeadRow
                    key={lead.id}
                    lead={lead}
                    onClick={() => setSelectedId(lead.id)}
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
              {Math.min(meta.page * meta.per_page, meta.total)} of {meta.total} leads
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

      {/* Lead detail panel */}
      {selectedId && (
        <LeadDetailPanel leadId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {/* Create form */}
      <LeadForm open={showForm} onClose={() => setShowForm(false)} />

      {/* Import modal */}
      <ImportModal
        open={showImport}
        onClose={() => setShowImport(false)}
        entity="leads"
        entityLabel="Leads"
      />
    </div>
  )
}

// ── Lead Row ──────────────────────────────────────────────────────────────────

function LeadRow({ lead, onClick }: { lead: Lead; onClick: () => void }) {
  const [hovered, setHovered] = useState(false)
  const isConverted = lead.status === 'converted'
  const dimStyle: React.CSSProperties = isConverted ? { opacity: 0.6 } : {}

  return (
    <tr
      tabIndex={0}
      style={{
        height: 60,
        cursor: 'pointer',
        background: hovered ? 'var(--surface-app)' : 'transparent',
        transition: 'background 150ms',
        ...dimStyle,
      }}
      onClick={onClick}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') onClick() }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* Lead name + email */}
      <td className="px-4 py-3">
        <div className="flex items-center gap-2.5">
          <div
            className="h-8 w-8 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
            style={{
              background: 'var(--color-primary-light)',
              color: 'var(--color-primary)',
            }}
          >
            {(lead.first_name[0] ?? '') + (lead.last_name[0] ?? '')}
          </div>
          <div className="min-w-0">
            <p
              className="text-sm font-bold truncate"
              style={{ color: isConverted ? 'var(--text-secondary)' : 'var(--text-primary)' }}
            >
              {lead.first_name} {lead.last_name}
            </p>
            <p className="text-xs truncate" style={{ color: 'var(--text-label)' }}>
              {lead.email}
            </p>
          </div>
        </div>
      </td>

      {/* Company */}
      <td className="px-4 py-3 hidden md:table-cell">
        <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>
          {lead.company ?? '—'}
        </span>
      </td>

      {/* Score */}
      <td className="px-4 py-3">
        <LeadScoreIndicator score={lead.lead_score} />
      </td>

      {/* Status */}
      <td className="px-4 py-3">
        <ConversionStatusBadge status={lead.status} lead={lead} />
      </td>

      {/* Assigned To */}
      <td className="px-4 py-3 hidden lg:table-cell">
        {lead.owner ? (
          <div className="flex items-center gap-1.5">
            <div
              className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
              style={{
                background: 'var(--color-primary-light)',
                color: 'var(--color-primary)',
              }}
            >
              {lead.owner.name
                .split(' ')
                .map((n) => n[0])
                .join('')
                .toUpperCase()
                .slice(0, 2)}
            </div>
            <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>
              {lead.owner.name}
            </span>
          </div>
        ) : (
          <span className="text-sm" style={{ color: 'var(--text-label)' }}>
            Unassigned
          </span>
        )}
      </td>

      {/* Created */}
      <td className="px-4 py-3 hidden lg:table-cell">
        <span
          className="text-sm"
          title={formatDate(lead.created_at)}
          style={{ color: 'var(--text-secondary)' }}
        >
          {formatRelativeTime(lead.created_at)}
        </span>
      </td>
    </tr>
  )
}
