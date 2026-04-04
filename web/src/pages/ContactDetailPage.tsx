import { useState, useRef, useEffect, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ChevronRight,
  Pencil,
  Mail,
  Phone,
  Calendar,
  ExternalLink,
  Lock,
  Download,
  ClipboardList,
  Trash2,
  Activity as ActivityIcon,
  Sparkles,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { useContact, useDeleteContact, useUpdateContact, useContactNotes, useAddContactNote, useEnrichContact } from '@/hooks/useContacts'
import { useContactActivities } from '@/hooks/useActivities'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { formatDate, formatRelativeTime, formatCurrency, getInitials } from '@/lib/utils'
import type { Activity, ActivityType, Contact, Deal, EnrichmentResult } from '@/api/types'
// Note: 'Activity' from lucide-react aliased to ActivityIcon above to avoid collision

// ── Helpers ──────────────────────────────────────────────────────────────────

function contactInitials(firstName: string, lastName: string) {
  return `${firstName[0] ?? ''}${lastName[0] ?? ''}`.toUpperCase()
}

// ── Activity Timeline ─────────────────────────────────────────────────────────

type ActivityTab = 'all' | 'call' | 'email' | 'meeting'

const TABS: { id: ActivityTab; label: string; icon: React.ReactNode; apiType?: ActivityType }[] = [
  { id: 'all', label: 'All', icon: null },
  { id: 'call', label: 'Call', icon: <Phone className="h-3.5 w-3.5" />, apiType: 'call' },
  { id: 'email', label: 'Email', icon: <Mail className="h-3.5 w-3.5" />, apiType: 'email' },
  { id: 'meeting', label: 'Meeting', icon: <Calendar className="h-3.5 w-3.5" />, apiType: 'meeting' },
]

const ACTIVITY_ICON_STYLES: Record<string, { bg: string; color: string }> = {
  call:    { bg: '#F0FDF4', color: '#16A34A' },
  email:   { bg: '#EFF6FF', color: '#2563EB' },
  meeting: { bg: '#FAF5FF', color: '#7C3AED' },
  task:    { bg: '#FFF7ED', color: '#C2410C' },
  note:    { bg: 'var(--color-primary-light)', color: 'var(--color-primary)' },
}

function ActivityIconCircle({ type }: { type: ActivityType }) {
  const style = ACTIVITY_ICON_STYLES[type] ?? ACTIVITY_ICON_STYLES.note
  const Icon = type === 'call' ? Phone : type === 'email' ? Mail : Calendar
  return (
    <div
      className="h-8 w-8 shrink-0 flex items-center justify-center rounded-full"
      style={{ background: style.bg }}
    >
      <Icon className="h-4 w-4" style={{ color: style.color }} />
    </div>
  )
}

function ActivityEntry({ activity }: { activity: Activity }) {
  return (
    <article
      className="flex gap-3 py-3 border-b last:border-0"
      style={{ borderColor: 'var(--border-subtle)' }}
      aria-label={`${activity.type} — ${activity.subject}`}
    >
      <ActivityIconCircle type={activity.type} />
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
            {activity.subject}
          </p>
          <time
            className="shrink-0 text-xs"
            style={{ color: 'var(--text-label)' }}
            title={formatDate(activity.created_at)}
          >
            {formatRelativeTime(activity.created_at)}
          </time>
        </div>
        {activity.description && (
          <p
            className="mt-0.5 text-xs line-clamp-2"
            style={{ color: 'var(--text-secondary)' }}
          >
            {activity.description}
          </p>
        )}
        {activity.owner && (
          <div className="mt-1.5 flex items-center gap-1.5">
            <div
              className="h-5 w-5 rounded-full flex items-center justify-center text-[9px] font-bold"
              style={{
                background: 'var(--color-primary-light)',
                color: 'var(--color-primary)',
              }}
            >
              {getInitials(activity.owner.name)}
            </div>
            <span className="text-xs" style={{ color: 'var(--text-label)' }}>
              {activity.owner.name}
            </span>
          </div>
        )}
      </div>
    </article>
  )
}

interface ActivityTimelineProps {
  contactId: string
}

function ActivityTimeline({ contactId }: ActivityTimelineProps) {
  const [activeTab, setActiveTab] = useState<ActivityTab>('all')
  const [visibleCount, setVisibleCount] = useState(20)
  const { data, isLoading } = useContactActivities(contactId)

  const activities = data?.data ?? []
  const filtered =
    activeTab === 'all'
      ? activities
      : activities.filter((a) => a.type === activeTab)
  const visible = filtered.slice(0, visibleCount)

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <p
        className="mb-3 text-[11px] font-semibold uppercase tracking-widest"
        style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
      >
        Activity
      </p>

      {/* Tabs */}
      <div
        role="tablist"
        aria-label="Filter activity by type"
        className="flex gap-1 mb-4 pb-3 border-b overflow-x-auto"
        style={{ borderColor: 'var(--border-subtle)' }}
      >
        {TABS.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            onClick={() => { setActiveTab(tab.id); setVisibleCount(20) }}
            className="flex items-center gap-1.5 px-3 rounded transition-colors shrink-0"
            style={{
              height: 32,
              fontSize: 13,
              fontWeight: activeTab === tab.id ? 600 : 500,
              color: activeTab === tab.id ? 'var(--color-primary)' : 'var(--text-secondary)',
              background: activeTab === tab.id ? 'var(--color-primary-light)' : 'transparent',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Entries */}
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="flex gap-3 py-3">
              <div className="h-8 w-8 rounded-full bg-slate-100 animate-pulse shrink-0" />
              <div className="flex-1 space-y-1.5">
                <div className="h-3 bg-slate-100 animate-pulse rounded w-3/4" />
                <div className="h-3 bg-slate-100 animate-pulse rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      ) : visible.length === 0 ? (
        <div
          className="flex flex-col items-center justify-center py-10 text-center"
          style={{ minHeight: 160 }}
        >
          <ActivityIcon className="h-10 w-10 mb-3" style={{ color: 'var(--text-label)' }} />
          <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
            No activity yet.
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--text-secondary)' }}>
            Log a call, send an email, or schedule a meeting.
          </p>
        </div>
      ) : (
        <>
          <div>
            {visible.map((a) => (
              <ActivityEntry key={a.id} activity={a} />
            ))}
          </div>
          {filtered.length > visibleCount && (
            <button
              className="mt-3 text-sm font-medium"
              style={{ color: 'var(--color-primary)', background: 'none', border: 'none', cursor: 'pointer' }}
              onClick={() => setVisibleCount((c) => c + 20)}
            >
              Load more
            </button>
          )}
        </>
      )}
    </div>
  )
}

// ── Opportunities Panel ───────────────────────────────────────────────────────

const DEAL_STAGE_ORDER: Record<string, number> = {
  lead: 1,
  qualified: 2,
  proposal: 3,
  negotiation: 4,
  closed_won: 5,
  closed_lost: 6,
}
const TOTAL_STAGES = 5

function PipelineProgress({ deal }: { deal: Deal }) {
  const stageIndex = DEAL_STAGE_ORDER[deal.stage] ?? 0
  const pct = deal.stage === 'closed_won' ? 100 : (stageIndex / TOTAL_STAGES) * 100
  const fillColor =
    deal.stage === 'closed_won'
      ? 'var(--color-success)'
      : deal.probability != null && deal.probability < 30
      ? 'var(--color-warning)'
      : 'var(--color-primary)'
  return (
    <div
      className="mt-2 rounded-full overflow-hidden"
      style={{ height: 6, background: 'var(--border-default)' }}
    >
      <div
        className="h-full rounded-full transition-all"
        style={{ width: `${pct}%`, background: fillColor }}
      />
    </div>
  )
}

function OpportunitiesPanel({ deals: allDeals }: { deals: Deal[] }) {
  const deals = allDeals.filter((d) => d.stage !== 'closed_lost')
  const shown = deals.slice(0, 3)
  const extra = deals.length - 3

  return (
    <div
      className="rounded-xl border p-4 mb-4"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <div className="flex items-center gap-2 mb-3">
        <p
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          Open Opportunities
        </p>
        {deals.length > 0 && (
          <span
            className="rounded-full px-1.5 py-0.5 text-[11px] font-semibold"
            style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
          >
            {deals.length}
          </span>
        )}
      </div>

      {shown.length === 0 ? (
        <div>
          <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
            No open opportunities.
          </p>
        </div>
      ) : (
        <>
          {shown.map((deal) => (
            <div
              key={deal.id}
              className="rounded-md border p-3 mb-2 cursor-pointer transition-all"
              style={{
                background: 'var(--surface-app)',
                borderColor: 'var(--border-subtle)',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--border-default)';
                (e.currentTarget as HTMLElement).style.background = 'var(--surface-card)'
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--border-subtle)';
                (e.currentTarget as HTMLElement).style.background = 'var(--surface-app)'
              }}
            >
              <p className="text-sm font-semibold mb-1" style={{ color: 'var(--text-primary)' }}>
                {deal.title}
              </p>
              <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
                {formatCurrency(deal.value_cents / 100, deal.currency)} · {deal.stage.replace(/_/g, ' ')}
              </p>
              <PipelineProgress deal={deal} />
            </div>
          ))}
          {extra > 0 && (
            <p className="text-xs mt-1" style={{ color: 'var(--color-primary)' }}>
              View all {deals.length} opportunities →
            </p>
          )}
        </>
      )}
    </div>
  )
}

// ── Private Notes ─────────────────────────────────────────────────────────────

function PrivateNotesPanel({ contactId, contactName }: { contactId: string; contactName: string }) {
  const { data: notes } = useContactNotes(contactId)
  const addNote = useAddContactNote()

  const latestContent = notes?.[0]?.content ?? ''
  const [text, setText] = useState('')
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle')
  const initializedRef = useRef(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Pre-fill with latest note on load (only once)
  useEffect(() => {
    if (!initializedRef.current && notes !== undefined) {
      setText(latestContent)
      initializedRef.current = true
    }
  }, [notes, latestContent])

  const save = useCallback(
    async (content: string) => {
      if (!content.trim()) return
      setSaveStatus('saving')
      try {
        await addNote.mutateAsync({ contactId, payload: { content } })
        setSaveStatus('saved')
        setTimeout(() => setSaveStatus('idle'), 1500)
      } catch {
        setSaveStatus('error')
      }
    },
    [addNote, contactId]
  )

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value)
    setSaveStatus('idle')
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => save(e.target.value), 1000)
  }

  return (
    <section
      className="rounded-xl border p-4"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      aria-labelledby="notes-heading"
    >
      <div className="flex items-center gap-1.5 mb-3">
        <Lock className="h-3 w-3" style={{ color: 'var(--text-label)' }} aria-hidden="true" />
        <p
          id="notes-heading"
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          title="Only visible to your team"
        >
          Private Notes
        </p>
      </div>
      <textarea
        className="w-full rounded border px-3 py-2.5 text-sm resize-y transition-colors focus:outline-none"
        style={{
          minHeight: 100,
          maxHeight: 280,
          borderColor: 'var(--border-default)',
          color: 'var(--text-primary)',
          lineHeight: 1.6,
          fontSize: 14,
        }}
        placeholder="Add a private note about this contact…"
        value={text}
        onChange={handleChange}
        aria-label={`Private notes about ${contactName}`}
        aria-describedby="notes-save-status"
        onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--border-focus)' }}
        onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)' }}
      />
      <div className="flex justify-end mt-1.5">
        <span
          id="notes-save-status"
          aria-live="polite"
          className="text-xs italic"
          style={{
            color:
              saveStatus === 'saving'
                ? 'var(--text-label)'
                : saveStatus === 'saved'
                ? 'var(--color-success)'
                : saveStatus === 'error'
                ? 'var(--color-danger)'
                : 'transparent',
          }}
        >
          {saveStatus === 'saving'
            ? 'Saving…'
            : saveStatus === 'saved'
            ? 'Saved ✓'
            : saveStatus === 'error'
            ? 'Save failed — Retry'
            : '.'}
        </span>
      </div>
    </section>
  )
}

// ── Enrichment Panel ──────────────────────────────────────────────────────────

type EnrichStatus = 'idle' | 'enriched' | 'failed'

function EnrichmentStatusBadge({ status }: { status: EnrichStatus }) {
  if (status === 'idle') {
    return (
      <span
        className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
        style={{ background: '#F1F5F9', color: 'var(--text-label)' }}
      >
        Not enriched
      </span>
    )
  }
  if (status === 'enriched') {
    return (
      <span
        className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
        style={{ background: '#F0FDF4', color: '#16A34A' }}
      >
        <Sparkles className="h-3 w-3" />
        Enriched
      </span>
    )
  }
  return (
    <span
      className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold"
      style={{ background: '#FEF2F2', color: 'var(--color-danger)' }}
    >
      Failed
    </span>
  )
}

function EnrichmentPanel({ contactId }: { contactId: string }) {
  const enrichContact = useEnrichContact()
  const [status, setStatus] = useState<EnrichStatus>('idle')
  const [result, setResult] = useState<EnrichmentResult | null>(null)

  const handleEnrich = async () => {
    try {
      const data = await enrichContact.mutateAsync(contactId)
      setResult(data)
      setStatus('enriched')
    } catch {
      setStatus('failed')
    }
  }

  return (
    <section
      className="rounded-xl border p-4 mt-4"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      aria-labelledby="enrichment-heading"
    >
      <div className="flex items-center justify-between gap-2 mb-3">
        <div className="flex items-center gap-2">
          <p
            id="enrichment-heading"
            className="text-[11px] font-semibold uppercase tracking-widest"
            style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          >
            Enrichment
          </p>
          <EnrichmentStatusBadge status={status} />
        </div>
        <button
          type="button"
          onClick={handleEnrich}
          disabled={enrichContact.isPending}
          className="flex items-center gap-1.5 px-3 rounded-md text-xs font-medium transition-colors whitespace-nowrap"
          style={{
            height: 28,
            border: '1px solid var(--border-default)',
            background: 'var(--surface-card)',
            color: 'var(--text-primary)',
            cursor: enrichContact.isPending ? 'default' : 'pointer',
            opacity: enrichContact.isPending ? 0.7 : 1,
          }}
          aria-label={status === 'enriched' ? 'Re-enrich contact' : 'Enrich contact'}
        >
          {enrichContact.isPending ? (
            <Loader2 className="h-3 w-3 animate-spin" />
          ) : status === 'enriched' ? (
            <RefreshCw className="h-3 w-3" />
          ) : (
            <Sparkles className="h-3 w-3" />
          )}
          {status === 'enriched' ? 'Re-enrich' : 'Enrich contact'}
        </button>
      </div>

      {status === 'failed' && (
        <p className="text-xs" style={{ color: 'var(--color-danger)' }}>
          Enrichment failed — no data found for this email domain.
        </p>
      )}

      {status === 'enriched' && result?.data && (
        <dl className="space-y-1.5">
          {result.data.company_name && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Company</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.company_name}</dd>
            </div>
          )}
          {result.data.industry && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Industry</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.industry}</dd>
            </div>
          )}
          {result.data.size && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>Size</dt>
              <dd className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>{result.data.size}</dd>
            </div>
          )}
          {result.data.linkedin_url && (
            <div className="flex gap-2">
              <dt className="text-xs w-20 shrink-0" style={{ color: 'var(--text-label)' }}>LinkedIn</dt>
              <dd className="text-xs">
                <a
                  href={result.data.linkedin_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 hover:underline"
                  style={{ color: 'var(--color-primary)' }}
                >
                  View profile
                  <ExternalLink className="h-3 w-3" />
                </a>
              </dd>
            </div>
          )}
        </dl>
      )}

      {status === 'idle' && (
        <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
          Click Enrich to auto-populate company data from this contact's email domain.
        </p>
      )}
    </section>
  )
}

// ── Edit Contact Modal ────────────────────────────────────────────────────────

function EditContactModal({
  contact,
  onClose,
}: {
  contact: Contact
  onClose: () => void
}) {
  const updateContact = useUpdateContact()
  const [form, setForm] = useState({
    first_name: contact.first_name,
    last_name: contact.last_name,
    email: contact.email,
    phone: contact.phone ?? '',
    title: contact.title ?? '',
  })

  const set = (field: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [field]: e.target.value }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await updateContact.mutateAsync({
      id: contact.id,
      payload: {
        first_name: form.first_name,
        last_name: form.last_name,
        email: form.email,
        phone: form.phone || undefined,
        title: form.title || undefined,
      },
    })
    onClose()
  }

  return (
    <Dialog open onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Contact</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                First name
              </label>
              <Input value={form.first_name} onChange={set('first_name')} required />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Last name
              </label>
              <Input value={form.last_name} onChange={set('last_name')} required />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
              Email
            </label>
            <Input type="email" value={form.email} onChange={set('email')} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Title
              </label>
              <Input value={form.title} onChange={set('title')} placeholder="Senior Developer" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
                Phone
              </label>
              <Input type="tel" value={form.phone} onChange={set('phone')} placeholder="+1 555 000 0000" />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateContact.isPending}>
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Delete Confirm Modal ──────────────────────────────────────────────────────

function DeleteConfirmModal({
  name,
  onConfirm,
  onCancel,
  isPending,
}: {
  name: string
  onConfirm: () => void
  onCancel: () => void
  isPending: boolean
}) {
  return (
    <Dialog open onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Delete Contact?</DialogTitle>
        </DialogHeader>
        <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
          This will permanently remove{' '}
          <strong style={{ color: 'var(--text-primary)' }}>{name}</strong> and all associated
          data. This action cannot be undone.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button
            disabled={isPending}
            onClick={onConfirm}
            style={{ background: 'var(--color-danger)', color: '#FFFFFF', border: 'none' }}
          >
            Delete Contact
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function ContactDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)

  const { data: contact, isLoading, isError } = useContact(id!)
  const deleteContact = useDeleteContact()

  if (isLoading) {
    return (
      <div className="p-6 max-w-[var(--content-max-width,1280px)]">
        {/* Breadcrumb skeleton */}
        <div className="h-4 w-48 bg-slate-100 animate-pulse rounded mb-5" />
        {/* Header card skeleton */}
        <div
          className="rounded-xl border p-5 flex items-center gap-4 mb-5"
          style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        >
          <div className="h-[72px] w-[72px] rounded-full bg-slate-100 animate-pulse shrink-0" />
          <div className="space-y-2 flex-1">
            <div className="h-6 w-56 bg-slate-100 animate-pulse rounded" />
            <div className="h-3.5 w-36 bg-slate-100 animate-pulse rounded" />
          </div>
        </div>
        {/* Body skeletons */}
        <div className="flex gap-5">
          <div className="flex-1 space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-16 bg-slate-100 animate-pulse rounded-xl" />
            ))}
          </div>
          <div className="w-[320px] shrink-0 space-y-3">
            <div className="h-32 bg-slate-100 animate-pulse rounded-xl" />
            <div className="h-24 bg-slate-100 animate-pulse rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !contact) {
    return (
      <div className="p-6 flex flex-col items-center justify-center py-24 text-center">
        <p className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
          Contact not found.
        </p>
        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          This contact may have been deleted or you may not have access.
        </p>
        <Link
          to="/contacts"
          className="text-sm font-medium"
          style={{ color: 'var(--color-primary)' }}
        >
          ← Back to Contacts
        </Link>
      </div>
    )
  }

  const fullName = `${contact.first_name} ${contact.last_name}`
  const initials = contactInitials(contact.first_name, contact.last_name)

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 mb-5 text-sm" style={{ color: 'var(--text-label)' }}>
        <Link
          to="/contacts"
          className="hover:underline"
          style={{ color: 'var(--color-primary)' }}
        >
          Contacts
        </Link>
        <ChevronRight className="h-3.5 w-3.5" />
        <span style={{ color: 'var(--text-primary)' }}>{fullName}</span>
      </nav>

      {/* Header Card */}
      <div
        className="rounded-xl border p-5 mb-5 flex flex-wrap items-center gap-4"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      >
        {/* Avatar */}
        <div
          className="h-[72px] w-[72px] rounded-full flex items-center justify-center shrink-0 text-2xl font-bold"
          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
        >
          {initials}
        </div>

        {/* Identity */}
        <div className="flex-1 min-w-0">
          <h1
            className="font-bold leading-tight"
            style={{ fontSize: 24, color: 'var(--text-primary)', letterSpacing: '-0.01em' }}
          >
            {fullName}
          </h1>
          {contact.title && (
            <p className="mt-0.5 text-sm" style={{ color: 'var(--text-secondary)' }}>
              {contact.title}
            </p>
          )}
          {contact.account ? (
            <Link
              to={`/accounts/${contact.account.id}`}
              className="mt-0.5 inline-flex items-center gap-1 text-sm font-medium hover:underline"
              style={{ color: 'var(--color-primary)' }}
              aria-label={`${contact.account.name} (opens account detail)`}
            >
              {contact.account.name}
              <ExternalLink className="h-3 w-3" aria-hidden="true" />
            </Link>
          ) : (
            <p className="mt-0.5 text-sm italic" style={{ color: 'var(--text-label)' }}>
              No company linked
            </p>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => setShowEdit(true)}
            className="flex items-center gap-1.5 px-4 rounded-md text-sm font-medium transition-colors"
            style={{
              height: 36,
              border: '1px solid var(--border-default)',
              background: 'var(--surface-card)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-app)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--surface-card)' }}
            aria-label={`Edit contact: ${fullName}`}
          >
            <Pencil className="h-3.5 w-3.5" />
            Edit
          </button>
          <button
            onClick={() => {
              window.location.href = `mailto:${contact.email}`
            }}
            className="flex items-center gap-1.5 px-4 rounded-md text-sm font-semibold text-white transition-colors"
            style={{
              height: 36,
              background: 'var(--color-primary)',
              border: 'none',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary-hover)' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--color-primary)' }}
            aria-label={`Send email to ${fullName}`}
          >
            <Mail className="h-3.5 w-3.5" />
            Send Email
          </button>
        </div>
      </div>

      {/* Body — 2-col on desktop, stack on mobile */}
      <div className="flex flex-col md:flex-row gap-5">
        {/* Right column on mobile (opportunities + notes + enrichment) */}
        <div className="md:hidden flex flex-col gap-4">
          <OpportunitiesPanel deals={contact.deals ?? []} />
          <PrivateNotesPanel contactId={id!} contactName={fullName} />
          <EnrichmentPanel contactId={id!} />
        </div>

        {/* Left — Timeline */}
        <div className="flex-1 min-w-0">
          <ActivityTimeline contactId={id!} />
        </div>

        {/* Right — Desktop only */}
        <div className="hidden md:block w-[320px] shrink-0">
          <OpportunitiesPanel deals={contact.deals ?? []} />
          <PrivateNotesPanel contactId={id!} contactName={fullName} />
          <EnrichmentPanel contactId={id!} />
        </div>
      </div>

      {/* Record Footer */}
      <footer
        className="mt-6 pt-4 border-t flex flex-wrap items-center justify-between gap-4"
        style={{ borderColor: 'var(--border-subtle)' }}
      >
        <p className="text-xs" style={{ color: 'var(--text-label)' }}>
          <span title={contact.id}>ID: {contact.id.slice(0, 8)}…</span>
          {' · '}
          Created: {formatDate(contact.created_at)}
          {' · '}
          Modified: {formatDate(contact.updated_at)}
        </p>
        <div className="flex items-center gap-2">
          <button
            onClick={() => window.print()}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--text-secondary)',
              borderColor: 'var(--border-default)',
              background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <Download className="h-3 w-3" />
            Export PDF
          </button>
          <button
            onClick={() => navigate(`/admin/audit?entityType=contact`)}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--text-secondary)',
              borderColor: 'var(--border-default)',
              background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <ClipboardList className="h-3 w-3" />
            Audit Log
          </button>
          <button
            onClick={() => setShowDelete(true)}
            className="flex items-center gap-1.5 px-3 rounded-md text-xs border transition-colors"
            style={{
              height: 32,
              color: 'var(--color-danger)',
              borderColor: '#FCA5A5',
              background: 'transparent',
              cursor: 'pointer',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = '#FEF2F2' }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent' }}
          >
            <Trash2 className="h-3 w-3" />
            Delete
          </button>
        </div>
      </footer>

      {/* Edit modal */}
      {showEdit && (
        <EditContactModal contact={contact} onClose={() => setShowEdit(false)} />
      )}

      {/* Delete confirm */}
      {showDelete && (
        <DeleteConfirmModal
          name={fullName}
          isPending={deleteContact.isPending}
          onConfirm={async () => {
            await deleteContact.mutateAsync(id!)
            navigate('/contacts')
          }}
          onCancel={() => setShowDelete(false)}
        />
      )}
    </div>
  )
}
