import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Send } from 'lucide-react'
import { usePortalTicket, usePortalTicketComments, useAddPortalComment } from '@/hooks/usePortal'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
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

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-2 py-2.5 border-b border-slate-100 last:border-0">
      <span className="shrink-0 text-xs font-medium text-slate-500 w-20">{label}</span>
      <div className="flex-1 text-right text-sm text-slate-800">{children}</div>
    </div>
  )
}

export function PortalTicketDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data: ticket, isLoading: ticketLoading } = usePortalTicket(id ?? '')
  const { data: comments = [], isLoading: commentsLoading } = usePortalTicketComments(id ?? '')
  const { mutateAsync: addComment, isPending: commenting } = useAddPortalComment()

  const [body, setBody] = useState('')

  async function handleComment(e: React.FormEvent) {
    e.preventDefault()
    if (!body.trim() || !id) return
    await addComment({ ticketId: id, body: body.trim() })
    setBody('')
  }

  if (ticketLoading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner />
      </div>
    )
  }

  if (!ticket) {
    return (
      <div className="py-16 text-center">
        <p className="text-sm text-slate-500">Ticket not found.</p>
        <Button variant="outline" size="sm" className="mt-4" asChild>
          <Link to="/portal/tickets">Back to tickets</Link>
        </Button>
      </div>
    )
  }

  const isClosed = ticket.status === 'closed' || ticket.status === 'resolved'

  return (
    <div className="space-y-6">
      {/* Back link */}
      <Link
        to="/portal/tickets"
        className="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to tickets
      </Link>

      {/* Ticket header */}
      <div>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="text-xl font-bold text-slate-900 flex-1">{ticket.subject}</h1>
          <div className="flex items-center gap-2 shrink-0">
            <Badge variant={priorityVariant[ticket.priority]}>{ticket.priority}</Badge>
            <Badge variant={statusVariant[ticket.status]}>{statusLabel[ticket.status]}</Badge>
          </div>
        </div>

        {ticket.description && (
          <p className="mt-2 whitespace-pre-wrap text-sm text-slate-600">{ticket.description}</p>
        )}
      </div>

      {/* Metadata card */}
      <div className="rounded-xl border border-slate-200 bg-white px-4 py-1 shadow-sm">
        <Field label="Status">
          <Badge variant={statusVariant[ticket.status]}>{statusLabel[ticket.status]}</Badge>
        </Field>
        <Field label="Priority">
          <Badge variant={priorityVariant[ticket.priority]}>{ticket.priority}</Badge>
        </Field>
        <Field label="Opened">
          <span className="text-slate-500">{formatDate(ticket.created_at)}</span>
        </Field>
        <Field label="Updated">
          <span className="text-slate-500">{formatDate(ticket.updated_at)}</span>
        </Field>
      </div>

      {/* Comments */}
      <div className="space-y-4">
        <h2 className="text-sm font-semibold text-slate-700">
          Replies {comments.length > 0 && `(${comments.length})`}
        </h2>

        {commentsLoading ? (
          <div className="flex justify-center py-6">
            <Spinner size="sm" />
          </div>
        ) : comments.length === 0 ? (
          <p className="text-sm text-slate-400">No replies yet.</p>
        ) : (
          <div className="space-y-3">
            {comments.filter((c) => !c.is_internal).map((c) => (
              <div
                key={c.id}
                className="rounded-lg border border-slate-200 bg-white p-3 space-y-1"
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="text-sm font-medium text-slate-700">
                    {c.author_id ? 'Support Team' : 'You'}
                  </span>
                  <span className="text-xs text-slate-400">{formatDate(c.created_at)}</span>
                </div>
                <p className="whitespace-pre-wrap text-sm text-slate-700">{c.body}</p>
              </div>
            ))}
          </div>
        )}

        {/* Reply form */}
        {!isClosed && (
          <form onSubmit={handleComment} className="space-y-2">
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Write a reply…"
              rows={3}
              className="w-full resize-none rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            />
            <div className="flex justify-end">
              <Button type="submit" size="sm" disabled={!body.trim() || commenting}>
                <Send className="h-3.5 w-3.5" />
                {commenting ? 'Sending…' : 'Send Reply'}
              </Button>
            </div>
          </form>
        )}

        {isClosed && (
          <p className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500">
            This ticket is {statusLabel[ticket.status].toLowerCase()}. Replies are disabled.
          </p>
        )}
      </div>
    </div>
  )
}
