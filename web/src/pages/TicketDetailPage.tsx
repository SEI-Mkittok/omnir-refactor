import { useState, useRef, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Mail,
  Phone,
  MapPin,
  Lock,
  CheckCircle,
  AlertCircle,
  Bold,
  Italic,
  Underline,
  List,
  ListOrdered,
  Quote,
  Link,
  Paperclip as PaperclipIcon,
  X,
} from 'lucide-react'
import * as RadixSelect from '@radix-ui/react-select'
import { ChevronDown, Check } from 'lucide-react'
import {
  useTicket,
  useTicketComments,
  useUpdateTicket,
  useAddTicketComment,
} from '@/hooks/useTickets'
import { useUsers } from '@/hooks/useUsers'
import type { TicketStatus, TicketPriority, TicketComment } from '@/api/types'
import { cn } from '@/lib/utils'

// ─── Helpers ──────────────────────────────────────────────────────────────────

const STATUS_BG: Record<TicketStatus, string> = {
  open: '#0D9488',
  pending: '#F59E0B',
  resolved: '#22C55E',
  closed: '#9CA3AF',
}
const STATUS_LABEL: Record<TicketStatus, string> = {
  open: 'OPEN',
  pending: 'PENDING',
  resolved: 'RESOLVED',
  closed: 'CLOSED',
}
const STATUS_OPTIONS: { value: TicketStatus; label: string }[] = [
  { value: 'open', label: 'Open' },
  { value: 'pending', label: 'Pending' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'closed', label: 'Closed' },
]
const PRIORITY_OPTIONS: { value: TicketPriority; label: string }[] = [
  { value: 'critical', label: 'Critical' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
]

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

function nameInitials(name?: string): string {
  if (!name) return '?'
  return name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase()
}

function contactFullName(first?: string, last?: string): string {
  return [first, last].filter(Boolean).join(' ') || 'Unknown Contact'
}

function contactInitials(first?: string, last?: string): string {
  return ((first?.[0] ?? '') + (last?.[0] ?? '')).toUpperCase() || '?'
}

// ─── Status Badge ──────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: TicketStatus }) {
  return (
    <span
      aria-label={`Status: ${STATUS_LABEL[status]}`}
      className="inline-block rounded-full px-2.5 py-0.5 text-[12px] font-semibold uppercase text-white"
      style={{ background: STATUS_BG[status] }}
    >
      {STATUS_LABEL[status]}
    </span>
  )
}

// ─── Inline Select ─────────────────────────────────────────────────────────────

interface InlineSelectProps<T extends string> {
  value: T
  options: { value: T; label: string }[]
  onValueChange: (v: T) => void
  label: string
}

function InlineSelect<T extends string>({ value, options, onValueChange, label }: InlineSelectProps<T>) {
  const selected = options.find((o) => o.value === value)
  return (
    <RadixSelect.Root value={value} onValueChange={(v) => onValueChange(v as T)}>
      <RadixSelect.Trigger
        aria-label={label}
        className="inline-flex items-center gap-1 rounded px-2 py-0.5 text-[14px] text-[var(--text-primary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors group min-h-[44px] sm:min-h-0"
      >
        <RadixSelect.Value>{selected?.label ?? value}</RadixSelect.Value>
        <RadixSelect.Icon className="opacity-0 group-hover:opacity-100 transition-opacity">
          <ChevronDown className="h-3 w-3 text-[var(--text-label)]" />
        </RadixSelect.Icon>
      </RadixSelect.Trigger>
      <RadixSelect.Portal>
        <RadixSelect.Content
          className="z-50 rounded-md border border-[var(--border-default)] bg-white shadow-md overflow-hidden"
          position="popper"
          sideOffset={4}
        >
          <RadixSelect.Viewport className="p-1">
            {options.map((opt) => (
              <RadixSelect.Item
                key={opt.value}
                value={opt.value}
                className="relative flex cursor-pointer select-none items-center gap-2 rounded px-3 py-1.5 text-[13px] text-[var(--text-primary)] outline-none hover:bg-[var(--color-primary-light)] data-[state=checked]:font-medium focus:bg-[var(--color-primary-light)]"
              >
                <RadixSelect.ItemText>{opt.label}</RadixSelect.ItemText>
                <RadixSelect.ItemIndicator className="absolute right-2">
                  <Check className="h-3 w-3" />
                </RadixSelect.ItemIndicator>
              </RadixSelect.Item>
            ))}
          </RadixSelect.Viewport>
        </RadixSelect.Content>
      </RadixSelect.Portal>
    </RadixSelect.Root>
  )
}

// ─── Thread Message ────────────────────────────────────────────────────────────

function MessageBubble({ comment }: { comment: TicketComment }) {
  const isInternal = comment.is_internal
  const isAgent = !!comment.author

  const bg = isInternal ? '#FEF3C7' : isAgent ? '#EFF6FF' : '#F3F4F6'
  const borderLeft = isInternal ? '3px solid #F59E0B' : 'none'

  return (
    <article
      aria-label={`${isInternal ? 'Internal note: ' : ''}${comment.author?.name ?? 'Client'} said: ${relativeTime(comment.created_at)}`}
      className="mb-3 rounded-md px-4 py-3"
      style={{ background: bg, borderLeft }}
    >
      {isInternal && (
        <div className="mb-2 flex items-center gap-1.5">
          <Lock className="h-3 w-3" style={{ color: '#92400E' }} />
          <span
            className="text-[11px] font-semibold uppercase tracking-[0.04em]"
            style={{ color: '#92400E' }}
          >
            INTERNAL NOTE · {comment.author?.name ?? 'Agent'} · {relativeTime(comment.created_at)}
          </span>
        </div>
      )}
      {!isInternal && (
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-[11px] font-semibold text-[var(--color-primary)]">
              {nameInitials(comment.author?.name)}
            </div>
            <span className="text-[13px] font-medium text-[var(--text-primary)]">
              {comment.author?.name ?? 'Client'}
            </span>
          </div>
          <span
            title={new Date(comment.created_at).toLocaleString()}
            className="text-[12px] text-[var(--text-label)]"
          >
            {relativeTime(comment.created_at)}
          </span>
        </div>
      )}
      <p className="text-[14px] leading-[1.6] text-[var(--text-primary)] whitespace-pre-wrap">
        {comment.body}
      </p>
    </article>
  )
}

// ─── Reply Composer ────────────────────────────────────────────────────────────

interface ReplyComposerProps {
  ticketId: string
  onSubmitted?: () => void
}

function ReplyComposer({ ticketId, onSubmitted }: ReplyComposerProps) {
  const [body, setBody] = useState('')
  const [isInternal, setIsInternal] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const addComment = useAddTicketComment()

  useEffect(() => {
    textareaRef.current?.focus()
  }, [])

  const handleSubmit = async () => {
    if (!body.trim() || submitting) return
    setSubmitting(true)
    try {
      await addComment.mutateAsync({ ticketId, payload: { body: body.trim(), is_internal: isInternal } })
      setBody('')
      setIsInternal(false)
      onSubmitted?.()
    } finally {
      setSubmitting(false)
    }
  }

  const applyFormatting = (tag: string) => {
    const ta = textareaRef.current
    if (!ta) return
    const start = ta.selectionStart
    const end = ta.selectionEnd
    const selected = body.slice(start, end)
    const wrapped = `${tag}${selected}${tag}`
    setBody(body.slice(0, start) + wrapped + body.slice(end))
  }

  const toolbarButtons = [
    { icon: <Bold className="h-4 w-4" />, label: 'Bold', action: () => applyFormatting('**') },
    { icon: <Italic className="h-4 w-4" />, label: 'Italic', action: () => applyFormatting('_') },
    { icon: <Underline className="h-4 w-4" />, label: 'Underline', action: () => applyFormatting('__') },
    null, // separator
    { icon: <ListOrdered className="h-4 w-4" />, label: 'Ordered list', action: () => setBody((b) => b + '\n1. ') },
    { icon: <List className="h-4 w-4" />, label: 'Unordered list', action: () => setBody((b) => b + '\n- ') },
    { icon: <Quote className="h-4 w-4" />, label: 'Quote', action: () => setBody((b) => b + '\n> ') },
    null,
    { icon: <Link className="h-4 w-4" />, label: 'Link', action: () => setBody((b) => b + '[text](url)') },
    { icon: <PaperclipIcon className="h-4 w-4" />, label: 'Attachment', action: () => {} },
  ]

  return (
    <form
      aria-label="Reply to ticket"
      className="sticky bottom-0 border-t border-[var(--border-default)] bg-[var(--surface-card)] p-4"
      style={isInternal ? { background: '#FEF3C7', borderColor: '#F59E0B' } : undefined}
      onSubmit={(e) => { e.preventDefault(); handleSubmit() }}
    >
      {/* Toolbar */}
      <div
        role="toolbar"
        aria-label="Formatting options"
        className="flex items-center gap-0.5 border-b border-[var(--border-subtle)] pb-2 mb-2 h-9"
      >
        {toolbarButtons.map((btn, i) =>
          btn === null ? (
            <div key={`sep-${i}`} className="mx-1 h-5 w-px bg-[var(--border-subtle)]" />
          ) : (
            <button
              key={btn.label}
              type="button"
              aria-label={btn.label}
              aria-pressed={false}
              onClick={btn.action}
              className="flex h-7 w-7 items-center justify-center rounded text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors sm:h-6 sm:w-6"
            >
              {btn.icon}
            </button>
          )
        )}
      </div>

      {/* Textarea */}
      <textarea
        ref={textareaRef}
        aria-label="Reply text"
        aria-describedby="composer-hint"
        placeholder={isInternal ? 'Add an internal note…' : 'Write a reply…'}
        value={body}
        onChange={(e) => setBody(e.target.value)}
        className="w-full resize-y rounded border px-3 py-2.5 text-[14px] text-[var(--text-primary)] placeholder:text-[var(--text-label)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
        style={{
          minHeight: '80px',
          maxHeight: '240px',
          borderColor: isInternal ? '#F59E0B' : 'var(--border-default)',
          background: 'transparent',
        }}
      />
      <p id="composer-hint" className="sr-only">
        Check "Internal Note" to make this visible only to agents.
      </p>

      {/* Footer */}
      <div className="mt-3 flex items-center justify-between gap-3">
        <label className="flex cursor-pointer items-center gap-2 text-[13px] select-none">
          <input
            type="checkbox"
            checked={isInternal}
            onChange={(e) => setIsInternal(e.target.checked)}
            className="accent-[var(--color-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
          />
          <span style={{ color: isInternal ? '#F59E0B' : 'var(--text-secondary)' }}>
            INTERNAL NOTE
          </span>
        </label>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setBody('')}
            className="h-9 rounded-md border border-[var(--border-default)] bg-transparent px-4 text-[13px] text-[var(--text-secondary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
          >
            Save as Draft
          </button>
          <button
            type="submit"
            disabled={!body.trim() || submitting}
            className="h-9 rounded-md bg-[var(--color-primary)] px-5 text-[13px] font-semibold text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
          >
            {submitting ? 'Sending…' : 'Submit Reply'}
          </button>
        </div>
      </div>
    </form>
  )
}

// ─── SLA Status Bar ────────────────────────────────────────────────────────────

function SLASection({ sla }: { sla: NonNullable<import('@/api/types').Ticket['sla']> }) {
  const resolutionDeadlineMs = new Date(sla.resolution_deadline).getTime()
  const now = Date.now()
  const isOverdue = sla.resolution_breached || now > resolutionDeadlineMs

  const timeLeft = (() => {
    if (isOverdue) {
      const over = now - resolutionDeadlineMs
      const hrs = Math.floor(over / 3600000)
      const mins = Math.floor((over % 3600000) / 60000)
      return `Overdue by ${hrs}h ${mins}m`
    }
    const left = resolutionDeadlineMs - now
    const hrs = Math.floor(left / 3600000)
    const mins = Math.floor((left % 3600000) / 60000)
    return `Due in ${hrs}h ${mins}m`
  })()

  // Estimate elapsed percent (can't know exact start without created_at SLA window)
  const progressPct = isOverdue ? 100 : 60

  const barColor = isOverdue
    ? '#EF4444'
    : progressPct >= 95
    ? '#EF4444'
    : progressPct >= 75
    ? '#F59E0B'
    : '#22C55E'

  return (
    <div
      className="rounded-md border p-4 space-y-3"
      style={isOverdue ? { borderColor: '#FCA5A5' } : { borderColor: 'var(--border-default)' }}
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
        SLA STATUS
      </p>

      {/* First response */}
      <div className="flex items-center gap-2 text-[13px]">
        <CheckCircle
          className="h-4 w-4 shrink-0"
          style={{ color: sla.response_breached ? 'var(--color-danger)' : 'var(--color-success)' }}
        />
        <span className={sla.response_breached ? 'text-[var(--color-danger)]' : 'text-[var(--text-secondary)]'}>
          First Response {sla.response_breached ? 'Breached' : 'Met'}
        </span>
      </div>

      {/* Resolution */}
      <div className="space-y-1.5">
        <p className="text-[12px] text-[var(--text-label)]">Resolution Deadline</p>
        <p
          className="text-[13px] font-medium"
          style={{ color: isOverdue ? 'var(--color-danger)' : 'var(--text-primary)' }}
          role="status"
          aria-live="polite"
          aria-label={isOverdue ? `SLA overdue: ${timeLeft}` : `SLA resolution: ${timeLeft}`}
        >
          {timeLeft}
        </p>
        <div className="h-2 w-full rounded-full" style={{ background: 'var(--border-default)' }}>
          <div
            role="progressbar"
            aria-valuenow={progressPct}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`Resolution SLA: ${progressPct}% elapsed`}
            className="h-2 rounded-full transition-all"
            style={{ width: `${progressPct}%`, background: barColor }}
          />
        </div>
      </div>
    </div>
  )
}

// ─── Tag Editor ────────────────────────────────────────────────────────────────

function TagEditor({ tags, onChange }: { tags: string[]; onChange: (tags: string[]) => void }) {
  const [adding, setAdding] = useState(false)
  const [input, setInput] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (adding) inputRef.current?.focus()
  }, [adding])

  const addTag = () => {
    const t = input.trim()
    if (t && !tags.includes(t)) onChange([...tags, t])
    setInput('')
    setAdding(false)
  }

  return (
    <div>
      <p className="mb-2 text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
        TAGS
      </p>
      <div className="flex flex-wrap gap-2">
        {tags.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[12px] font-medium"
            style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
          >
            {tag}
            <button
              aria-label={`Remove tag: ${tag}`}
              onClick={() => onChange(tags.filter((t) => t !== tag))}
              className="flex items-center justify-center rounded-full hover:opacity-70 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            >
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
        {adding ? (
          <input
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') { e.preventDefault(); addTag() }
              if (e.key === 'Escape') { setAdding(false); setInput('') }
            }}
            onBlur={addTag}
            className="w-28 rounded-full border px-3 py-0.5 text-[12px] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            style={{ borderColor: 'var(--border-default)' }}
            placeholder="tag name"
          />
        ) : (
          <button
            onClick={() => setAdding(true)}
            className="inline-flex items-center rounded-full border border-dashed px-2.5 py-0.5 text-[12px] text-[var(--text-label)] hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
            style={{ borderColor: 'var(--text-disabled)' }}
          >
            + ADD TAG
          </button>
        )}
      </div>
    </div>
  )
}

// ─── Right Panel ───────────────────────────────────────────────────────────────

interface RightPanelProps {
  ticketId: string
}

function RightPanel({ ticketId }: RightPanelProps) {
  const { data: ticket } = useTicket(ticketId)
  const updateTicket = useUpdateTicket()
  const { data: usersData } = useUsers({ limit: 100 })
  const [localTags, setLocalTags] = useState<string[]>([])

  useEffect(() => {
    if (ticket?.custom_fields?.tags) {
      setLocalTags(ticket.custom_fields.tags as string[])
    }
  }, [ticket?.id])

  if (!ticket) return null

  const contact = ticket.contact
  const assigneeOptions = [
    ...(usersData?.data.map((u) => ({ value: u.id, label: u.name })) ?? []),
    { value: '', label: 'Unassigned' },
  ]

  const propertyRows: { label: string; content: React.ReactNode }[] = [
    {
      label: 'Assigned To',
      content: (
        <InlineSelect
          value={ticket.assignee?.id ?? ''}
          options={assigneeOptions}
          onValueChange={(v) => updateTicket.mutate({ id: ticketId, payload: { assignee_id: v || undefined } })}
          label="Change assignee"
        />
      ),
    },
    {
      label: 'Status',
      content: (
        <InlineSelect
          value={ticket.status}
          options={STATUS_OPTIONS}
          onValueChange={(v) => updateTicket.mutate({ id: ticketId, payload: { status: v } })}
          label="Change status"
        />
      ),
    },
    {
      label: 'Source',
      content: <span className="text-[14px] text-[var(--text-primary)]">{ticket.source ?? '—'}</span>,
    },
    {
      label: 'Created',
      content: (
        <span className="text-[14px] text-[var(--text-primary)]">
          {new Date(ticket.created_at).toLocaleDateString('default', { month: 'short', day: 'numeric', year: 'numeric' })}
        </span>
      ),
    },
  ]

  return (
    <aside className="w-full lg:w-[320px] lg:flex-shrink-0 space-y-3">
      {/* Contact card */}
      <div className="rounded-md border border-[var(--border-default)] bg-[var(--surface-card)] p-4">
        <p className="mb-3 text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
          CONTACT
        </p>
        {contact ? (
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-[14px] font-semibold text-[var(--color-primary)]">
                {contactInitials(contact.first_name, contact.last_name)}
              </div>
              <div>
                <p className="text-[16px] font-bold text-[var(--text-primary)]">
                  {contactFullName(contact.first_name, contact.last_name)}
                </p>
                {contact.title && (
                  <p className="text-[13px] text-[var(--text-secondary)]">{contact.title}</p>
                )}
              </div>
            </div>
            <div className="space-y-1.5">
              {contact.email && (
                <div className="flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)]">
                  <Mail className="h-3.5 w-3.5 text-[var(--text-label)] shrink-0" />
                  {contact.email}
                </div>
              )}
              {contact.phone && (
                <div className="flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)]">
                  <Phone className="h-3.5 w-3.5 text-[var(--text-label)] shrink-0" />
                  {contact.phone}
                </div>
              )}
              {contact.account?.name && (
                <div className="flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)]">
                  <MapPin className="h-3.5 w-3.5 text-[var(--text-label)] shrink-0" />
                  {contact.account.name}
                </div>
              )}
            </div>
            <a
              href={`/contacts?id=${contact.id}`}
              className="text-[13px] font-medium text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
            >
              View Contact →
            </a>
          </div>
        ) : (
          <p className="text-[13px] text-[var(--text-secondary)]">No contact linked.</p>
        )}
      </div>

      {/* Ticket properties */}
      <div className="rounded-md border border-[var(--border-default)] bg-[var(--surface-card)] p-4">
        <p className="mb-3 text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
          TICKET PROPERTIES
        </p>
        <div className="divide-y divide-[var(--border-subtle)]">
          {propertyRows.map((row, i) => (
            <div
              key={row.label}
              className="flex items-center gap-2 py-1.5"
              style={i === propertyRows.length - 1 ? { borderBottom: 'none' } : undefined}
            >
              <span className="w-24 shrink-0 text-[12px] text-[var(--text-label)]">{row.label}</span>
              {row.content}
            </div>
          ))}
        </div>
      </div>

      {/* SLA */}
      {ticket.sla && <SLASection sla={ticket.sla} />}

      {/* Tags */}
      <div className="rounded-md border border-[var(--border-default)] bg-[var(--surface-card)] p-4">
        <TagEditor tags={localTags} onChange={setLocalTags} />
      </div>
    </aside>
  )
}

// ─── Ticket Detail Page ────────────────────────────────────────────────────────

export function TicketDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const ticketId = id ?? ''

  const { data: ticket, isLoading: ticketLoading, isError: ticketError } = useTicket(ticketId)
  const { data: comments = [], isLoading: commentsLoading } = useTicketComments(ticketId)

  if (ticketError) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-center gap-4">
        <AlertCircle className="h-12 w-12 text-[var(--color-danger)]" />
        <p className="text-[18px] font-semibold text-[var(--text-primary)]">Ticket not found.</p>
        <p className="text-[14px] text-[var(--text-secondary)]">
          This ticket may have been deleted or you may not have permission to view it.
        </p>
        <button
          onClick={() => navigate('/tickets')}
          className="inline-flex items-center gap-2 h-9 rounded-md border border-[var(--border-default)] px-4 text-[13px] text-[var(--text-primary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Tickets
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <button
            onClick={() => navigate('/tickets')}
            className="mb-2 inline-flex items-center gap-1.5 text-[14px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to List
          </button>
          <div className="flex flex-wrap items-center gap-3">
            {ticketLoading ? (
              <div className="h-7 w-64 rounded bg-[var(--border-subtle)] animate-pulse" />
            ) : (
              <h1
                className="text-[28px] font-bold text-[var(--text-primary)] leading-tight"
                style={{ letterSpacing: 'var(--letter-spacing-tight)' }}
              >
                {ticket?.subject}.
              </h1>
            )}
            {ticket && <StatusBadge status={ticket.status} />}
          </div>
        </div>
      </div>

      {/* Body: mobile stacks, desktop side-by-side */}
      <div className="flex flex-col lg:flex-row gap-4">
        {/* Right panel on mobile (accordion-like: just shows at top) */}
        <div className="lg:hidden">
          <RightPanel ticketId={ticketId} />
        </div>

        {/* Thread + Composer */}
        <div className="flex flex-1 min-w-0 flex-col rounded-lg border border-[var(--border-default)] bg-[var(--surface-card)] overflow-hidden">
          {/* Thread scroll area */}
          <div className="flex-1 overflow-y-auto p-4 space-y-0" style={{ minHeight: '300px', maxHeight: '60vh' }}>
            {commentsLoading ? (
              <div aria-busy="true" className="space-y-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} aria-hidden="true" className="rounded-md p-4" style={{ background: 'var(--border-subtle)', opacity: 0.6 }}>
                    <div className="h-3 w-1/3 rounded bg-[var(--border-default)] animate-pulse mb-2" />
                    <div className="h-3 w-3/4 rounded bg-[var(--border-default)] animate-pulse" />
                  </div>
                ))}
              </div>
            ) : comments.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <p className="text-[14px] text-[var(--text-secondary)]">No messages yet.</p>
                <p className="text-[13px] text-[var(--text-label)]">Be the first to reply.</p>
              </div>
            ) : (
              comments.map((c) => <MessageBubble key={c.id} comment={c} />)
            )}
          </div>

          {/* Reply composer */}
          <ReplyComposer ticketId={ticketId} />
        </div>

        {/* Right panel on desktop */}
        <div className="hidden lg:block">
          <RightPanel ticketId={ticketId} />
        </div>
      </div>
    </div>
  )
}
