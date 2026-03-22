import { useState } from 'react'
import { Plus, Clock } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useSLAPolicies, useCreateSLAPolicy, useUpdateSLAPolicy, useDeleteSLAPolicy } from '@/hooks/useSLA'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/Dialog'
import { formatDate } from '@/lib/utils'
import type { SLAPolicy, SLAPriorityFilter, CreateSLAPolicyRequest, UpdateSLAPolicyRequest } from '@/api/types'

const PRIORITY_FILTER_OPTIONS = [
  { label: 'All Priorities', value: 'all' },
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Critical', value: 'critical' },
]

const priorityFilterLabel: Record<SLAPriorityFilter, string> = {
  all: 'All',
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  critical: 'Critical',
}

const priorityFilterVariant: Record<SLAPriorityFilter, 'default' | 'blue' | 'orange' | 'red' | 'gray'> = {
  all: 'default',
  low: 'gray',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

function formatMinutes(minutes: number): string {
  if (minutes < 60) return `${minutes}m`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

// ---- Create / Edit Dialog ----

interface SLAPolicyFormDialogProps {
  open: boolean
  policy?: SLAPolicy | null
  onClose: () => void
}

function SLAPolicyFormDialog({ open, policy, onClose }: SLAPolicyFormDialogProps) {
  const isEdit = !!policy
  const { mutate: create, isPending: isCreating } = useCreateSLAPolicy()
  const { mutate: update, isPending: isUpdating } = useUpdateSLAPolicy()
  const isPending = isCreating || isUpdating

  const [form, setForm] = useState<CreateSLAPolicyRequest>({
    name: '',
    response_time_minutes: 60,
    resolution_time_minutes: 480,
    priority_filter: 'all',
  })
  const [error, setError] = useState('')

  const handleOpen = () => {
    if (policy) {
      setForm({
        name: policy.name,
        response_time_minutes: policy.response_time_minutes,
        resolution_time_minutes: policy.resolution_time_minutes,
        priority_filter: policy.priority_filter,
      })
    } else {
      setForm({ name: '', response_time_minutes: 60, resolution_time_minutes: 480, priority_filter: 'all' })
    }
    setError('')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.name.trim()) {
      setError('Policy name is required.')
      return
    }
    if (form.response_time_minutes <= 0 || form.resolution_time_minutes <= 0) {
      setError('Response and resolution times must be positive.')
      return
    }

    if (isEdit && policy) {
      const payload: UpdateSLAPolicyRequest = {
        name: form.name,
        response_time_minutes: form.response_time_minutes,
        resolution_time_minutes: form.resolution_time_minutes,
        priority_filter: form.priority_filter,
      }
      update({ id: policy.id, payload }, { onSuccess: onClose, onError: () => setError('Failed to update policy.') })
    } else {
      create(form, { onSuccess: onClose, onError: () => setError('Failed to create policy.') })
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent onOpenAutoFocus={handleOpen}>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit SLA Policy' : 'Add SLA Policy'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Policy name</label>
            <Input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="e.g. Standard SLA"
              autoFocus
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Response time (minutes)</label>
              <Input
                type="number"
                min={1}
                value={form.response_time_minutes}
                onChange={(e) => setForm((f) => ({ ...f, response_time_minutes: Number(e.target.value) }))}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Resolution time (minutes)</label>
              <Input
                type="number"
                min={1}
                value={form.resolution_time_minutes}
                onChange={(e) => setForm((f) => ({ ...f, resolution_time_minutes: Number(e.target.value) }))}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Applies to priority</label>
            <select
              className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
              value={form.priority_filter}
              onChange={(e) => setForm((f) => ({ ...f, priority_filter: e.target.value as SLAPriorityFilter }))}
            >
              {PRIORITY_FILTER_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Saving…' : isEdit ? 'Save Changes' : 'Create Policy'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ---- Main page ----

export function SLASettingsPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)
  const [editPolicy, setEditPolicy] = useState<SLAPolicy | null>(null)

  const debouncedSearch = useDebounce(search, 300)

  const { data, isLoading } = useSLAPolicies({
    page,
    per_page: 20,
    q: debouncedSearch || undefined,
  })

  const { mutate: deletePolicy } = useDeleteSLAPolicy()

  const policies = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const totalPages = data?.meta?.total_pages ?? 1

  const handleDelete = (p: SLAPolicy) => {
    if (!confirm(`Delete SLA policy "${p.name}"? This cannot be undone.`)) return
    deletePolicy(p.id)
  }

  const columns: Column<SLAPolicy>[] = [
    {
      key: 'name',
      header: 'Name',
      render: (p) => <span className="font-medium text-slate-900">{p.name}</span>,
    },
    {
      key: 'response_time_minutes',
      header: 'Response Time',
      render: (p) => (
        <span className="text-slate-600">{formatMinutes(p.response_time_minutes)}</span>
      ),
    },
    {
      key: 'resolution_time_minutes',
      header: 'Resolution Time',
      hideOnMobile: true,
      render: (p) => (
        <span className="text-slate-600">{formatMinutes(p.resolution_time_minutes)}</span>
      ),
    },
    {
      key: 'priority_filter',
      header: 'Priority Filter',
      render: (p) => (
        <Badge variant={priorityFilterVariant[p.priority_filter]}>
          {priorityFilterLabel[p.priority_filter]}
        </Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      hideOnMobile: true,
      render: (p) => <span className="text-xs text-slate-500">{formatDate(p.created_at)}</span>,
    },
    {
      key: 'id',
      header: '',
      render: (p) => (
        <div className="flex items-center justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={(e) => { e.stopPropagation(); setEditPolicy(p) }}
          >
            Edit
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="text-red-600 hover:bg-red-50 hover:border-red-300"
            onClick={(e) => { e.stopPropagation(); handleDelete(p) }}
          >
            Delete
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">SLA Policies</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${total} ${total === 1 ? 'policy' : 'policies'}`}
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" />
          Add Policy
        </Button>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search policies…"
        filters={[]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={policies}
          isLoading={isLoading}
          emptyIcon={Clock}
          emptyTitle="No SLA policies yet"
          emptyDescription="Add a policy to track response and resolution times."
          keyExtractor={(p) => p.id}
        />

        {totalPages > 1 && (
          <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3">
            <p className="text-sm text-slate-500">
              Page {page} of {totalPages}
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
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Dialogs */}
      <SLAPolicyFormDialog open={showCreate} onClose={() => setShowCreate(false)} />
      <SLAPolicyFormDialog
        open={!!editPolicy}
        policy={editPolicy}
        onClose={() => setEditPolicy(null)}
      />
    </div>
  )
}
