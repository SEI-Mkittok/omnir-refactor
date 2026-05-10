import { useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, AlertCircle } from 'lucide-react'
import { usePortalTicket, usePortalTicketComments, useAddPortalComment } from '@/hooks/usePortal'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { formatDate } from '@/lib/utils'
import type { PortalTicketStatus, TicketPriority } from '@/api/types'

const statusVariant: Record<PortalTicketStatus, 'blue' | 'indigo' | 'yellow' | 'green' | 'gray'> = {
  open: 'blue',
  in_progress: 'indigo',
  pending: 'yellow',
  resolved: 'green',
  closed: 'gray',
}

const statusLabel: Record<PortalTicketStatus, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  pending: 'Pending',
  resolved: 'Resolved',
  closed: 'Closed',
}

const priorityVariant: Record<TicketPriority, 'gray' | 'blue' | 'orange' | 'red'> = {
  low: 'gray',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

function Initials({ name }: { name: string }) {
  const letters = name
    .split(' ')
    .map((w) => w[0]?.toUpperCase() ?? '')
    .slice(0, 2)
    .join('')
  return (
    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-300 text-xs font-semibold text-slate-700">
      {letters}
    </div>
  )
}

export function PortalTicketDetailPage() {
  const { id } = useParams<{ id: string }>()
  const user = useAuthStore((s) => s.user)
  const { data: ticket, isLoading: ticketLoading } = usePortalTicket(id ?? '')
  const { data: comments = [], isLoading: commentsLoading } = usePortalTicketComments(id ?? '')
  const { mutateAsync: addComment, isPending: commenting } = useAddPortalComment()

  const [body, setBody] = useState('')
  const [replyError, setReplyError] = useState<string | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Return focus to textarea after submission
  useEffect(() => {
    if (!commenting && !replyError) {
      textareaRef.current?.focus()
    }
  }, [commenting, replyError])

  async function handleComment(e: React.FormEvent) {
    e.preventDefault()
    if (!body.trim() || !id) return
    setReplyError(null)
    try {
      await addComment({ ticketId: id, body: body.trim() })
      setBody('')
    } catch {
      setReplyError('Failed to send reply. Please try again.')
    }
  }

  if (ticketLoading) {
    return (
      <div aria-busy="true" className="space-y-4">
        {/* Header skeleton */}
        <div className="h-6 w-48 rounded bg-slate-100 animate-pulse" aria-hidden="true" />
        <div className="h-8 w-2/3 rounded bg-slate-100 animate-pulse" aria-hidden="true" />
        {/* Comment skeletons */}
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-20 rounded-lg bg-slate-100 animate-pulse" aria-hidden="true" />
        ))}
      </div>
    )
  }

  if (!ticket) {
    return (
      <div className="py-16 text-center">
        <p className="text-sm text-slate-500">Ticket not found.</p>
        <Button variant="outline" size="sm" className="mt-4" asChild>
          <Link to="/portal/tickets">← My Tickets</Link>
        </Button>
      </div>
    )
  }

  const visibleComments = comments.filter((c) => !c.is_internal)

  return (
    <div className="space-y-6">
      {/* Back link */}
      <Link
        to="/portal/tickets"
        className="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700"
      >
        <ArrowLeft className="h-4 w-4" />
        ← My Tickets
      </Link>

      {/* Ticket header */}
      <div>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h2 className="text-xl font-bold text-slate-900 flex-1">{ticket.subject}</h2>
          <div className="flex items-center gap-2 shrink-0">
            <Badge
              variant={priorityVariant[ticket.priority]}
              aria-label={`Priority: ${ticket.priority}`}
            >
              {ticket.priority}
            </Badge>
            <Badge
              variant={statusVariant[ticket.status]}
              aria-label={`Status: ${statusLabel[ticket.status]}`}
            >
              {statusLabel[ticket.status]}
            </Badge>
          </div>
        </div>
        <p className="mt-1 text-xs text-slate-400">Opened {formatDate(ticket.created_at)}</p>
      </div>

      {/* Comment thread */}
      <div className="space-y-3">
        {commentsLoading ? (
          <div aria-busy="true" className="space-y-3">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-16 rounded-lg bg-slate-100 animate-pulse" aria-hidden="true" />
            ))}
          </div>
        ) : visibleComments.length === 0 ? (
          <p className="text-sm text-slate-400">No replies yet.</p>
        ) : (
          visibleComments.map((c) => {
            const isClient = c.author_id === user?.id
            return (
              <div
                key={c.id}
                className={[
                  'flex gap-3 rounded-lg p-3',
                  isClient ? 'bg-white border border-slate-200' : 'bg-slate-100',
                ].join(' ')}
              >
                <Initials name={isClient ? (user?.name ?? 'You') : 'Support'} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium text-slate-700">
                      {isClient ? (user?.name ?? 'You') : 'Support Team'}
                    </span>
                    <span className="text-xs text-slate-400 shrink-0">{formatDate(c.created_at)}</span>
                  </div>
                  <p className="mt-1 whitespace-pre-wrap text-sm text-slate-700">{c.body}</p>
                </div>
              </div>
            )
          })
        )}
      </div>

      {/* Reply form — always shown; backend enforces closed-ticket restriction */}
      <form onSubmit={handleComment} className="space-y-2">
        <label htmlFor="reply-body" className="block text-sm font-medium text-slate-700">
          Your reply
        </label>
        <textarea
          id="reply-body"
          ref={textareaRef}
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Add a reply…"
          rows={3}
          aria-label="Your reply"
          disabled={commenting}
          className="w-full resize-none rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)] disabled:opacity-50"
        />
        {replyError && (
          <div role="alert" className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {replyError}
          </div>
        )}
        <div className="flex justify-end">
          <Button
            type="submit"
            size="sm"
            aria-label="Send reply"
            disabled={!body.trim() || commenting}
          >
            {commenting ? 'Sending…' : 'Send Reply'}
          </Button>
        </div>
      </form>
    </div>
  )
}
