import { useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus, UserRound, Loader2 } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useLeads, useCreateLead } from '@/hooks/useLeads'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { formatDate } from '@/lib/utils'
import { leadStatusBadgeVariant, leadStatusLabel } from '@/components/omnir/LeadDetailPanel'
import { LeadDetailPanel } from '@/components/omnir/LeadDetailPanel'
import type { Lead, LeadStatus, CreateLeadRequest } from '@/api/types'

// ── Create lead form ─────────────────────────────────────────────────────────

const INITIAL_FORM: CreateLeadRequest = {
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  company: '',
  lead_source: '',
  status: 'new',
}

interface LeadFormProps {
  open: boolean
  onClose: () => void
}

function LeadForm({ open, onClose }: LeadFormProps) {
  const [form, setForm] = useState<CreateLeadRequest>(INITIAL_FORM)
  const [errors, setErrors] = useState<Partial<Record<keyof CreateLeadRequest, string>>>({})
  const createLead = useCreateLead()

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
    })
    setForm(INITIAL_FORM)
    setErrors({})
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) { setForm(INITIAL_FORM); setErrors({}); onClose() }
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
              <label className="mb-1 block text-xs font-medium text-slate-700">
                First name <span className="text-red-500">*</span>
              </label>
              <Input value={form.first_name} onChange={set('first_name')} placeholder="Jane" aria-invalid={!!errors.first_name} />
              {errors.first_name && <p className="mt-0.5 text-xs text-red-500">{errors.first_name}</p>}
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">
                Last name <span className="text-red-500">*</span>
              </label>
              <Input value={form.last_name} onChange={set('last_name')} placeholder="Smith" aria-invalid={!!errors.last_name} />
              {errors.last_name && <p className="mt-0.5 text-xs text-red-500">{errors.last_name}</p>}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              Email <span className="text-red-500">*</span>
            </label>
            <Input type="email" value={form.email} onChange={set('email')} placeholder="jane@example.com" aria-invalid={!!errors.email} />
            {errors.email && <p className="mt-0.5 text-xs text-red-500">{errors.email}</p>}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Company</label>
              <Input value={form.company ?? ''} onChange={set('company')} placeholder="Acme Inc." />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Source</label>
              <Input value={form.lead_source ?? ''} onChange={set('lead_source')} placeholder="Website, Referral…" />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Phone</label>
              <Input type="tel" value={form.phone ?? ''} onChange={set('phone')} placeholder="+1 555 000 0000" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Status</label>
              <select
                value={form.status}
                onChange={set('status')}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
              >
                <option value="new">New</option>
                <option value="contacted">Contacted</option>
                <option value="qualified">Qualified</option>
                <option value="unqualified">Unqualified</option>
              </select>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={createLead.isPending}>
              {createLead.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Create Lead
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Leads list options ────────────────────────────────────────────────────────

const STATUS_OPTIONS = [
  { label: 'New', value: 'new' },
  { label: 'Contacted', value: 'contacted' },
  { label: 'Qualified', value: 'qualified' },
  { label: 'Unqualified', value: 'unqualified' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
  { label: 'Name A–Z', value: 'last_name:asc' },
  { label: 'Name Z–A', value: 'last_name:desc' },
]

// ── Main page ─────────────────────────────────────────────────────────────────

export function LeadsPage() {
  const [searchParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(searchParams.get('openId'))
  const [showForm, setShowForm] = useState(false)

  const debouncedSearch = useDebounce(search, 300)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const { data, isLoading } = useLeads({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    status: (status as LeadStatus) || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const leads = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  const columns: Column<Lead>[] = [
    {
      key: 'last_name',
      header: 'Name',
      sortable: true,
      render: (l) => (
        <span className="font-medium text-slate-900">
          {l.first_name} {l.last_name}
        </span>
      ),
    },
    {
      key: 'email',
      header: 'Email',
      render: (l) => <span className="text-slate-600">{l.email}</span>,
    },
    {
      key: 'company',
      header: 'Company',
      hideOnMobile: true,
      render: (l) => <span className="text-slate-600">{l.company ?? '—'}</span>,
    },
    {
      key: 'lead_source',
      header: 'Source',
      hideOnMobile: true,
      render: (l) => <span className="text-slate-500 text-xs">{l.lead_source ?? '—'}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (l) => (
        <Badge variant={leadStatusBadgeVariant[l.status]}>{leadStatusLabel[l.status]}</Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      hideOnMobile: true,
      render: (l) => <span className="text-slate-500 text-xs">{formatDate(l.created_at)}</span>,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Leads</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total` : 'Loading…'}
          </p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4" />
          Add Lead
        </Button>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search leads…"
        filters={[
          {
            label: 'Status',
            value: status,
            options: STATUS_OPTIONS,
            onChange: (v) => { setStatus(v); setPage(1) },
          },
          {
            label: 'Sort',
            value: sortKey,
            options: SORT_OPTIONS,
            onChange: (v) => { setSortKey(v); setPage(1) },
          },
        ]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={leads}
          isLoading={isLoading}
          sortBy={sortBy}
          sortDir={sortDir}
          onSort={handleSort}
          onRowClick={(l) => setSelectedId(l.id)}
          emptyIcon={UserRound}
          emptyTitle="No leads found"
          emptyDescription="Try adjusting your search or filters, or add a new lead."
          keyExtractor={(l) => l.id}
        />

        {/* Pagination */}
        {meta && meta.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3">
            <p className="text-sm text-slate-500">
              Page {meta.page} of {meta.total_pages}
            </p>
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
      </div>

      {/* Detail panel */}
      {selectedId && (
        <LeadDetailPanel leadId={selectedId} onClose={() => setSelectedId(null)} />
      )}

      {/* Create form */}
      <LeadForm open={showForm} onClose={() => setShowForm(false)} />
    </div>
  )
}
