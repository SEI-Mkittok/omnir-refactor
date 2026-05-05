import { useMemo, useRef } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useVirtualizer } from '@tanstack/react-virtual'
import { Activity, FileText, Mail, Paperclip, Quote, Ticket, Workflow } from 'lucide-react'
import { useActivities, useContactActivities, useDealActivities } from '@/hooks/useActivities'
import { useContactNotes } from '@/hooks/useContacts'
import { useContactEmails } from '@/hooks/useEmails'
import { useTickets } from '@/hooks/useTickets'
import { useDealQuotes, useQuotes } from '@/hooks/useQuotes'
import { useEntityAttachments } from '@/hooks/useAttachments'
import { useAccountContacts, useAccountNotes } from '@/hooks/useAccounts'
import { useDealNotes } from '@/hooks/useDeals'
import { sequencesApi } from '@/api/sequences'
import { inboxApi } from '@/api/inbox'
import { Spinner } from '@/components/ui/Spinner'
import { cn, formatDate, formatRelativeTime } from '@/lib/utils'
import type {
  EntityAttachment,
  InboxThread,
  Note,
  Quote as QuoteType,
  SequenceEnrollment,
  Ticket as TicketType,
} from '@/api/types'

type TimelineEventType = 'activity' | 'note' | 'email' | 'ticket' | 'quote' | 'sequence' | 'attachment'

interface TimelineEvent {
  id: string
  type: TimelineEventType
  title: string
  description?: string
  date: string
  user?: string
  linkedEntity?: string
  href?: string
}

interface UnifiedTimelineProps {
  entityType: 'contact' | 'account' | 'deal'
  entityId: string
  contactId?: string
  accountId?: string
}

const EVENT_META: Record<TimelineEventType, { label: string; Icon: React.ElementType; color: string; bg: string }> = {
  activity: { label: 'Activity', Icon: Activity, color: 'text-blue-600', bg: 'bg-blue-50' },
  note: { label: 'Note', Icon: FileText, color: 'text-slate-600', bg: 'bg-slate-100' },
  email: { label: 'Email', Icon: Mail, color: 'text-indigo-600', bg: 'bg-indigo-50' },
  ticket: { label: 'Ticket', Icon: Ticket, color: 'text-orange-600', bg: 'bg-orange-50' },
  quote: { label: 'Quote', Icon: Quote, color: 'text-emerald-600', bg: 'bg-emerald-50' },
  sequence: { label: 'Sequence', Icon: Workflow, color: 'text-fuchsia-600', bg: 'bg-fuchsia-50' },
  attachment: { label: 'Attachment', Icon: Paperclip, color: 'text-cyan-600', bg: 'bg-cyan-50' },
}

const ALL_EVENT_TYPES = Object.keys(EVENT_META) as TimelineEventType[]

function readCsvParam(params: URLSearchParams, key: string): string[] {
  const value = params.get(key)
  if (!value) return []
  return value.split(',').map((v) => v.trim()).filter(Boolean)
}

export function UnifiedTimeline({ entityType, entityId, contactId, accountId }: UnifiedTimelineProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const parentRef = useRef<HTMLDivElement>(null)

  const selectedTypes = useMemo(() => {
    const values = readCsvParam(searchParams, 'tl_types').filter((v): v is TimelineEventType =>
      ALL_EVENT_TYPES.includes(v as TimelineEventType)
    )
    return values.length ? values : ALL_EVENT_TYPES
  }, [searchParams])

  const fromDate = searchParams.get('tl_from') ?? ''
  const toDate = searchParams.get('tl_to') ?? ''
  const userFilter = searchParams.get('tl_user') ?? ''
  const linkedEntityFilter = searchParams.get('tl_entity') ?? ''

  const resolvedContactId = entityType === 'contact' ? entityId : contactId
  const resolvedAccountId = entityType === 'account' ? entityId : accountId

  const contactActivitiesQuery = useContactActivities(entityType === 'contact' ? entityId : '')
  const accountActivitiesQuery = useActivities(
    entityType === 'account'
      ? { account_id: entityId, per_page: 200, sort_by: 'created_at', sort_dir: 'desc' }
      : undefined
  )
  const dealActivitiesQuery = useDealActivities(entityType === 'deal' ? entityId : '')

  const contactNotesQuery = useContactNotes(entityType === 'contact' ? entityId : '')
  const accountNotesQuery = useAccountNotes(entityType === 'account' ? entityId : '')
  const dealNotesQuery = useDealNotes(entityType === 'deal' ? entityId : '')

  const contactEmailsQuery = useContactEmails(resolvedContactId ?? '')

  const accountContactsQuery = useAccountContacts(entityType === 'account' ? entityId : '')
  const accountInboxThreadsQuery = useQuery({
    queryKey: ['timeline', 'account-inbox', entityId, accountContactsQuery.data?.map((c) => c.id) ?? []],
    enabled: entityType === 'account' && (accountContactsQuery.data?.length ?? 0) > 0,
    staleTime: 30_000,
    queryFn: async () => {
      const threadMap = new Map<string, InboxThread>()
      await Promise.all(
        (accountContactsQuery.data ?? []).map(async (contact) => {
          const result = await inboxApi.listThreads({ contact_id: contact.id, page: 1, limit: 50 })
          for (const thread of result.data ?? []) {
            const existing = threadMap.get(thread.thread_id)
            if (!existing || new Date(thread.last_message_at) > new Date(existing.last_message_at)) {
              threadMap.set(thread.thread_id, thread)
            }
          }
        })
      )
      return Array.from(threadMap.values())
    },
  })

  const ticketsQuery = useTickets({
    per_page: 200,
    sort_by: 'created_at',
    sort_dir: 'desc',
    ...(entityType === 'contact' ? { contact_id: entityId } : {}),
    ...(entityType === 'account' ? { account_id: entityId } : {}),
    ...(entityType === 'deal' && resolvedContactId ? { contact_id: resolvedContactId } : {}),
    ...(entityType === 'deal' && !resolvedContactId && resolvedAccountId ? { account_id: resolvedAccountId } : {}),
  })

  const contactQuotesQuery = useQuotes(entityType === 'contact' ? { contact_id: entityId, limit: 200 } : undefined)
  const accountQuotesQuery = useQuotes(entityType === 'account' ? { account_id: entityId, limit: 200 } : undefined)
  const dealQuotesQuery = useDealQuotes(entityType === 'deal' ? entityId : '')

  const attachmentEntityType = entityType === 'deal' ? 'deal' : entityType === 'account' ? 'account' : 'contact'
  const attachmentsQuery = useEntityAttachments(attachmentEntityType, entityId)

  const sequenceContactIds = useMemo(() => {
    if (entityType === 'contact' && entityId) return [entityId]
    if (entityType === 'deal' && resolvedContactId) return [resolvedContactId]
    if (entityType === 'account') return (accountContactsQuery.data ?? []).map((contact) => contact.id)
    return []
  }, [entityType, entityId, resolvedContactId, accountContactsQuery.data])

  const sequencesQuery = useQuery({
    queryKey: ['timeline', 'sequences', entityType, entityId, sequenceContactIds],
    enabled: sequenceContactIds.length > 0,
    staleTime: 30_000,
    queryFn: async () => {
      const sequences = await sequencesApi.list({ page: 1, limit: 200 })
      const contactIdSet = new Set(sequenceContactIds)
      const matched: { sequenceName: string; enrollment: SequenceEnrollment }[] = []

      await Promise.all(
        (sequences.data ?? []).map(async (sequence) => {
          const enrollments = await sequencesApi.listEnrollments(sequence.id)
          for (const enrollment of enrollments.data ?? []) {
            if (contactIdSet.has(enrollment.contact_id)) {
              matched.push({ sequenceName: sequence.name, enrollment })
            }
          }
        })
      )

      return matched
    },
  })

  const events = useMemo(() => {
    const items: TimelineEvent[] = []

    const activityItems =
      entityType === 'contact'
        ? contactActivitiesQuery.data?.data ?? []
        : entityType === 'account'
          ? accountActivitiesQuery.data?.data ?? []
          : dealActivitiesQuery.data?.data ?? []

    for (const activity of activityItems) {
      const entity = activity.deal?.title ?? activity.account?.name ?? 'Record'
      items.push({
        id: `activity-${activity.id}`,
        type: 'activity',
        title: activity.subject,
        description: activity.description,
        date: activity.created_at,
        user: activity.owner?.name,
        linkedEntity: entity,
      })
    }

    const notes =
      entityType === 'contact'
        ? (contactNotesQuery.data ?? [])
        : entityType === 'account'
          ? (accountNotesQuery.data ?? [])
          : (dealNotesQuery.data ?? [])

    for (const note of notes as Note[]) {
      items.push({
        id: `note-${note.id}`,
        type: 'note',
        title: 'Note added',
        description: note.content,
        date: note.created_at,
        user: note.owner?.name,
        linkedEntity: entityType,
      })
    }

    if (entityType === 'account') {
      for (const thread of accountInboxThreadsQuery.data ?? []) {
        items.push({
          id: `email-thread-${thread.thread_id}`,
          type: 'email',
          title: thread.subject || '(no subject)',
          description: thread.snippet,
          date: thread.last_message_at,
          linkedEntity: `Thread ${thread.thread_id.slice(0, 8)}`,
          href: '/inbox',
        })
      }
    } else {
      for (const email of contactEmailsQuery.data?.data ?? []) {
        const isOutbound = email.direction === 'outbound'
        items.push({
          id: `email-${email.id}`,
          type: 'email',
          title: email.subject || '(no subject)',
          description: `${isOutbound ? 'To' : 'From'} ${isOutbound ? email.to_addr : email.from_addr}`,
          date: email.sent_at,
          linkedEntity: email.thread_id ? `Thread ${email.thread_id.slice(0, 8)}` : 'Email thread',
        })
      }
    }

    for (const ticket of (ticketsQuery.data?.data ?? []) as TicketType[]) {
      items.push({
        id: `ticket-${ticket.id}`,
        type: 'ticket',
        title: ticket.subject,
        description: `${ticket.status} · ${ticket.priority}`,
        date: ticket.updated_at || ticket.created_at,
        user: ticket.assignee?.name,
        linkedEntity: `Ticket ${ticket.id.slice(0, 8)}`,
        href: `/tickets/${ticket.id}`,
      })
    }

    const quotes =
      entityType === 'contact'
        ? (contactQuotesQuery.data?.data ?? [])
        : entityType === 'account'
          ? (accountQuotesQuery.data?.data ?? [])
          : (dealQuotesQuery.data?.data ?? [])

    for (const quote of quotes as QuoteType[]) {
      items.push({
        id: `quote-${quote.id}`,
        type: 'quote',
        title: quote.title,
        description: quote.status,
        date: quote.updated_at || quote.created_at,
        linkedEntity: quote.deal?.title ?? 'Quote',
        href:
          entityType === 'deal'
            ? `/deals?view=list`
            : entityType === 'account'
              ? `/quotes?account_id=${entityId}`
              : `/quotes?contact_id=${entityId}`,
      })
    }

    for (const attachment of (attachmentsQuery.data ?? []) as EntityAttachment[]) {
      items.push({
        id: `attachment-${attachment.id}`,
        type: 'attachment',
        title: attachment.filename,
        description: attachment.content_type,
        date: attachment.created_at,
        linkedEntity: `${entityType} attachment`,
      })
    }

    for (const sequenceEvent of sequencesQuery.data ?? []) {
      items.push({
        id: `sequence-${sequenceEvent.enrollment.id}`,
        type: 'sequence',
        title: `Enrolled in ${sequenceEvent.sequenceName}`,
        description: sequenceEvent.enrollment.status,
        date: sequenceEvent.enrollment.enrolled_at,
        linkedEntity: sequenceEvent.sequenceName,
        href: '/sequences',
      })
    }

    return items.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
  }, [
    entityType,
    entityId,
    contactActivitiesQuery.data,
    accountActivitiesQuery.data,
    dealActivitiesQuery.data,
    contactNotesQuery.data,
    accountNotesQuery.data,
    dealNotesQuery.data,
    contactEmailsQuery.data,
    accountInboxThreadsQuery.data,
    ticketsQuery.data,
    contactQuotesQuery.data,
    accountQuotesQuery.data,
    dealQuotesQuery.data,
    attachmentsQuery.data,
    sequencesQuery.data,
  ])

  const linkedEntities = useMemo(
    () => Array.from(new Set(events.map((event) => event.linkedEntity).filter(Boolean))) as string[],
    [events]
  )

  const users = useMemo(
    () => Array.from(new Set(events.map((event) => event.user).filter(Boolean))) as string[],
    [events]
  )

  const filteredEvents = useMemo(() => {
    return events.filter((event) => {
      if (!selectedTypes.includes(event.type)) return false
      if (userFilter && event.user !== userFilter) return false
      if (linkedEntityFilter && event.linkedEntity !== linkedEntityFilter) return false

      const eventTime = new Date(event.date).getTime()
      if (fromDate) {
        const fromTime = new Date(`${fromDate}T00:00:00`).getTime()
        if (eventTime < fromTime) return false
      }
      if (toDate) {
        const toTime = new Date(`${toDate}T23:59:59`).getTime()
        if (eventTime > toTime) return false
      }

      return true
    })
  }, [events, selectedTypes, userFilter, linkedEntityFilter, fromDate, toDate])

  const virtualizer = useVirtualizer({
    count: filteredEvents.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 86,
    overscan: 8,
  })

  const isLoading =
    (entityType === 'contact' ? contactActivitiesQuery.isLoading : entityType === 'account' ? accountActivitiesQuery.isLoading : dealActivitiesQuery.isLoading) ||
    (entityType === 'contact' ? contactNotesQuery.isLoading : entityType === 'account' ? accountNotesQuery.isLoading : dealNotesQuery.isLoading) ||
    (entityType === 'account' ? accountInboxThreadsQuery.isLoading : contactEmailsQuery.isLoading) ||
    ticketsQuery.isLoading ||
    (entityType === 'contact' ? contactQuotesQuery.isLoading : entityType === 'account' ? accountQuotesQuery.isLoading : dealQuotesQuery.isLoading) ||
    attachmentsQuery.isLoading ||
    (sequenceContactIds.length > 0 && sequencesQuery.isLoading)

  const updateParam = (key: string, value?: string) => {
    const next = new URLSearchParams(searchParams)
    if (!value) next.delete(key)
    else next.set(key, value)
    setSearchParams(next, { replace: true })
  }

  const toggleType = (type: TimelineEventType) => {
    const current = readCsvParam(searchParams, 'tl_types').filter((v): v is TimelineEventType =>
      ALL_EVENT_TYPES.includes(v as TimelineEventType)
    )

    const base = current.length ? current : ALL_EVENT_TYPES
    const nextSet = new Set(base)
    if (nextSet.has(type)) nextSet.delete(type)
    else nextSet.add(type)

    if (nextSet.size === ALL_EVENT_TYPES.length || nextSet.size === 0) {
      updateParam('tl_types', '')
      return
    }

    updateParam('tl_types', Array.from(nextSet).join(','))
  }

  return (
    <div className="rounded-xl border p-5" style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}>
      <div className="mb-4 flex items-center justify-between gap-3">
        <p className="text-[11px] font-semibold uppercase tracking-widest" style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}>
          Timeline
        </p>
        <span className="text-xs" style={{ color: 'var(--text-label)' }}>
          {filteredEvents.length} of {events.length}
        </span>
      </div>

      <div className="mb-3 flex flex-wrap gap-2">
        {ALL_EVENT_TYPES.map((type) => {
          const meta = EVENT_META[type]
          const isActive = selectedTypes.includes(type)
          return (
            <button
              key={type}
              onClick={() => toggleType(type)}
              className={cn(
                'rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
                isActive ? 'border-slate-700 bg-slate-700 text-white' : 'border-slate-200 bg-white text-slate-600'
              )}
            >
              {meta.label}
            </button>
          )
        })}
      </div>

      <div className="mb-4 grid grid-cols-1 gap-2 md:grid-cols-4">
        <input type="date" value={fromDate} onChange={(e) => updateParam('tl_from', e.target.value)} className="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm" aria-label="From date" />
        <input type="date" value={toDate} onChange={(e) => updateParam('tl_to', e.target.value)} className="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm" aria-label="To date" />
        <select value={userFilter} onChange={(e) => updateParam('tl_user', e.target.value)} className="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm" aria-label="Filter by user">
          <option value="">All users</option>
          {users.map((user) => <option key={user} value={user}>{user}</option>)}
        </select>
        <select value={linkedEntityFilter} onChange={(e) => updateParam('tl_entity', e.target.value)} className="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm" aria-label="Filter by linked entity">
          <option value="">All linked entities</option>
          {linkedEntities.map((entity) => <option key={entity} value={entity}>{entity}</option>)}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-8"><Spinner /></div>
      ) : filteredEvents.length === 0 ? (
        <p className="py-8 text-sm text-slate-500">No timeline events match your filters.</p>
      ) : (
        <div ref={parentRef} className="max-h-[640px] overflow-auto rounded-lg border border-slate-100">
          <div style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
            {virtualizer.getVirtualItems().map((virtualRow) => {
              const event = filteredEvents[virtualRow.index]
              const meta = EVENT_META[event.type]
              const Icon = meta.Icon
              return (
                <div key={event.id} className="absolute left-0 top-0 w-full border-b border-slate-100 px-3 py-2" style={{ transform: `translateY(${virtualRow.start}px)` }}>
                  <div className="flex items-start gap-3">
                    <div className={cn('mt-0.5 flex h-7 w-7 items-center justify-center rounded-full', meta.bg)}>
                      <Icon className={cn('h-3.5 w-3.5', meta.color)} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        {event.href ? <Link to={event.href} className="truncate text-sm font-medium text-slate-900 hover:underline">{event.title}</Link> : <p className="truncate text-sm font-medium text-slate-900">{event.title}</p>}
                        <span className="text-xs text-slate-400" title={formatDate(event.date)}>{formatRelativeTime(event.date)}</span>
                      </div>
                      <p className="mt-0.5 text-xs text-slate-500 line-clamp-2">{event.description || meta.label}</p>
                      <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-400">
                        <span>{meta.label}</span>
                        {event.user && <span>By {event.user}</span>}
                        {event.linkedEntity && <span>Linked: {event.linkedEntity}</span>}
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
