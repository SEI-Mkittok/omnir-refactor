import { useState } from 'react'
import { Webhook, Plus, ChevronDown, ChevronRight, CheckCircle, XCircle, Clock } from 'lucide-react'
import {
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useDeleteWebhook,
  useTestWebhook,
  useWebhookDeliveries,
} from '@/hooks/useWebhooks'
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
import type { Webhook as WebhookType, WebhookEvent, WebhookDelivery, CreateWebhookRequest, UpdateWebhookRequest } from '@/api/types'

const ALL_EVENTS: WebhookEvent[] = [
  'contact.created',
  'contact.updated',
  'deal.created',
  'deal.updated',
  'deal.stage_changed',
  'deal.deleted',
  'activity.created',
]

// ---- Create / Edit Dialog ----

interface WebhookFormDialogProps {
  open: boolean
  webhook?: WebhookType | null
  onClose: () => void
}

function WebhookFormDialog({ open, webhook, onClose }: WebhookFormDialogProps) {
  const isEdit = !!webhook
  const { mutate: create, isPending: isCreating } = useCreateWebhook()
  const { mutate: update, isPending: isUpdating } = useUpdateWebhook()
  const isPending = isCreating || isUpdating

  const [url, setUrl] = useState('')
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [error, setError] = useState('')

  const handleOpen = () => {
    setUrl(webhook?.url ?? '')
    setEvents(webhook?.events ?? [])
    setError('')
  }

  const toggleEvent = (e: WebhookEvent) =>
    setEvents((prev) => prev.includes(e) ? prev.filter((x) => x !== e) : [...prev, e])

  const handleSubmit = (ev: React.FormEvent) => {
    ev.preventDefault()
    setError('')
    if (!url.trim()) { setError('URL is required.'); return }
    if (!/^https?:\/\//.test(url)) { setError('URL must start with http:// or https://.'); return }
    if (events.length === 0) { setError('Select at least one event.'); return }

    if (isEdit && webhook) {
      const payload: UpdateWebhookRequest = { url, events }
      update({ id: webhook.id, payload }, { onSuccess: onClose, onError: () => setError('Failed to update webhook.') })
    } else {
      const payload: CreateWebhookRequest = { url, events }
      create(payload, { onSuccess: onClose, onError: () => setError('Failed to create webhook.') })
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent onOpenAutoFocus={handleOpen}>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Webhook' : 'Add Webhook'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Endpoint URL</label>
            <Input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/webhooks"
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">Events</label>
            <div className="grid grid-cols-2 gap-1.5">
              {ALL_EVENTS.map((e) => (
                <label key={e} className="flex items-center gap-2 cursor-pointer text-sm text-slate-700">
                  <input
                    type="checkbox"
                    checked={events.includes(e)}
                    onChange={() => toggleEvent(e)}
                    className="rounded border-slate-300 text-[var(--color-primary)] focus:ring-[var(--border-focus)]"
                  />
                  {e}
                </label>
              ))}
            </div>
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Saving…' : isEdit ? 'Save Changes' : 'Add Webhook'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ---- Delivery Log (expanded row) ----

function DeliveryLog({ webhookId }: { webhookId: string }) {
  const { data: deliveries, isLoading } = useWebhookDeliveries(webhookId)

  if (isLoading) return <p className="px-4 py-3 text-sm text-slate-500">Loading deliveries…</p>
  if (!deliveries?.length) return <p className="px-4 py-3 text-sm text-slate-500">No delivery attempts yet.</p>

  return (
    <div className="divide-y divide-slate-100">
      {deliveries.map((d: WebhookDelivery) => (
        <div key={d.id} className="flex items-center gap-3 px-4 py-2 text-sm">
          {d.status === 'delivered'
            ? <CheckCircle className="h-4 w-4 shrink-0 text-green-500" />
            : d.status === 'failed'
            ? <XCircle className="h-4 w-4 shrink-0 text-red-500" />
            : <Clock className="h-4 w-4 shrink-0 text-amber-500" />}
          <span className="w-36 shrink-0 text-slate-500 text-xs">{formatDate(d.created_at)}</span>
          <Badge variant={d.status === 'delivered' ? 'default' : d.status === 'failed' ? 'red' : 'orange'}>
            {d.status}
          </Badge>
          <span className="text-slate-600 truncate">{d.event}</span>
          {d.last_error && (
            <span className="ml-auto text-xs text-red-500 truncate max-w-xs" title={d.last_error}>
              {d.last_error}
            </span>
          )}
        </div>
      ))}
    </div>
  )
}

// ---- Main page ----

export function WebhooksPage() {
  const [showCreate, setShowCreate] = useState(false)
  const [editWebhook, setEditWebhook] = useState<WebhookType | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const { data: hooks, isLoading } = useWebhooks()
  const { mutate: deleteWebhook } = useDeleteWebhook()
  const { mutate: updateWebhook } = useUpdateWebhook()
  const { mutate: testWebhook, isPending: isTesting } = useTestWebhook()

  const webhooks = hooks ?? []

  const handleDelete = (w: WebhookType) => {
    if (!confirm(`Delete webhook for "${w.url}"? This cannot be undone.`)) return
    deleteWebhook(w.id)
  }

  const handleToggleActive = (w: WebhookType) => {
    updateWebhook({ id: w.id, payload: { active: !w.active } })
  }

  const handleTest = (w: WebhookType) => {
    testWebhook(w.id, {
      onSuccess: () => alert(`Test event queued for ${w.url}. Check the delivery log for results.`),
      onError: () => alert('Failed to send test event.'),
    })
  }

  const columns: Column<WebhookType>[] = [
    {
      key: 'url',
      header: 'Endpoint URL',
      render: (w) => (
        <div>
          <p className="font-medium text-slate-900 truncate max-w-xs">{w.url}</p>
          <p className="text-xs text-slate-400">{w.events.length} event{w.events.length !== 1 ? 's' : ''}</p>
        </div>
      ),
    },
    {
      key: 'events',
      header: 'Events',
      hideOnMobile: true,
      render: (w) => (
        <div className="flex flex-wrap gap-1">
          {w.events.slice(0, 3).map((e) => (
            <Badge key={e} variant="gray" className="text-xs">{e}</Badge>
          ))}
          {w.events.length > 3 && (
            <Badge variant="gray" className="text-xs">+{w.events.length - 3}</Badge>
          )}
        </div>
      ),
    },
    {
      key: 'active',
      header: 'Status',
      render: (w) => (
        <Badge variant={w.active ? 'default' : 'gray'}>
          {w.active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Added',
      hideOnMobile: true,
      render: (w) => <span className="text-xs text-slate-500">{formatDate(w.created_at)}</span>,
    },
    {
      key: 'id',
      header: '',
      render: (w) => (
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="outline"
            size="sm"
            onClick={(e) => { e.stopPropagation(); handleTest(w) }}
            disabled={isTesting}
            title="Send test event"
          >
            Test
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={(e) => { e.stopPropagation(); handleToggleActive(w) }}
          >
            {w.active ? 'Disable' : 'Enable'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={(e) => { e.stopPropagation(); setEditWebhook(w) }}
          >
            Edit
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="text-red-600 hover:bg-red-50 hover:border-red-300"
            onClick={(e) => { e.stopPropagation(); handleDelete(w) }}
          >
            Delete
          </Button>
          <Button
            variant="outline"
            size="sm"
            title="Delivery log"
            onClick={(e) => { e.stopPropagation(); setExpandedId(expandedId === w.id ? null : w.id) }}
          >
            {expandedId === w.id ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
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
          <h1 className="text-2xl font-bold text-slate-900">Webhooks</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${webhooks.length} registered endpoint${webhooks.length !== 1 ? 's' : ''}`}
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" />
          Add Webhook
        </Button>
      </div>

      {/* Table + expandable delivery log */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={webhooks}
          isLoading={isLoading}
          emptyIcon={Webhook}
          emptyTitle="No webhooks yet"
          emptyDescription="Add a webhook endpoint to receive real-time event notifications."
          keyExtractor={(w) => w.id}
          onRowClick={(w) => setExpandedId(expandedId === w.id ? null : w.id)}
        />

        {/* Expandable delivery log rows */}
        {webhooks.map((w) =>
          expandedId === w.id ? (
            <div key={`deliveries-${w.id}`} className="border-t border-slate-200 bg-slate-50">
              <p className="px-4 pt-3 pb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">
                Recent Deliveries
              </p>
              <DeliveryLog webhookId={w.id} />
            </div>
          ) : null
        )}
      </div>

      {/* Dialogs */}
      <WebhookFormDialog open={showCreate} onClose={() => setShowCreate(false)} />
      <WebhookFormDialog
        open={!!editWebhook}
        webhook={editWebhook}
        onClose={() => setEditWebhook(null)}
      />
    </div>
  )
}
