import { useState } from 'react'
import { Mail, ArrowUpRight, ArrowDownLeft, ChevronDown, ChevronUp, Plus } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { cn, formatRelativeTime } from '@/lib/utils'
import { useContactEmails } from '@/hooks/useEmails'
import { ComposeEmailModal } from '@/components/omnir/ComposeEmailModal'
import type { ContactEmail } from '@/api/types'

// ── Thread grouping ───────────────────────────────────────────────────────────

interface EmailThread {
  threadId: string
  subject: string
  emails: ContactEmail[]
  latestAt: string
}

function groupIntoThreads(emails: ContactEmail[]): EmailThread[] {
  const map = new Map<string, ContactEmail[]>()
  for (const email of emails) {
    const key = email.thread_id || email.id
    const group = map.get(key) ?? []
    group.push(email)
    map.set(key, group)
  }

  const threads: EmailThread[] = []
  map.forEach((msgs, threadId) => {
    const sorted = [...msgs].sort(
      (a, b) => new Date(a.sent_at).getTime() - new Date(b.sent_at).getTime()
    )
    threads.push({
      threadId,
      subject: sorted[0].subject || '(no subject)',
      emails: sorted,
      latestAt: sorted[sorted.length - 1].sent_at,
    })
  })

  return threads.sort(
    (a, b) => new Date(b.latestAt).getTime() - new Date(a.latestAt).getTime()
  )
}

// ── Single email item ─────────────────────────────────────────────────────────

function EmailItem({ email }: { email: ContactEmail }) {
  const [expanded, setExpanded] = useState(false)
  const preview = email.body.length > 140 ? email.body.slice(0, 140) + '…' : email.body
  const isOutbound = email.direction === 'outbound'

  return (
    <div className="py-2.5 first:pt-0">
      <div className="flex items-start gap-2">
        <div
          className={cn(
            'mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full',
            isOutbound ? 'bg-indigo-100 text-indigo-600' : 'bg-slate-100 text-slate-500'
          )}
        >
          {isOutbound ? (
            <ArrowUpRight className="h-3 w-3" />
          ) : (
            <ArrowDownLeft className="h-3 w-3" />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <p className="text-xs text-slate-500 truncate">
              {isOutbound ? `To: ${email.to_addr}` : `From: ${email.from_addr}`}
            </p>
            <span className="shrink-0 text-xs text-slate-400">{formatRelativeTime(email.sent_at)}</span>
          </div>
          <button
            onClick={() => setExpanded((v) => !v)}
            className="mt-0.5 w-full text-left"
          >
            <p className="text-sm text-slate-800 whitespace-pre-wrap">
              {expanded ? email.body : preview}
            </p>
            {email.body.length > 140 && (
              <span className="mt-0.5 flex items-center gap-0.5 text-xs text-indigo-500 hover:text-indigo-700">
                {expanded ? (
                  <><ChevronUp className="h-3 w-3" /> Show less</>
                ) : (
                  <><ChevronDown className="h-3 w-3" /> Show more</>
                )}
              </span>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Thread card ───────────────────────────────────────────────────────────────

interface ThreadCardProps {
  thread: EmailThread
  contactId: string
  contactEmail?: string
}

function ThreadCard({ thread, contactId, contactEmail }: ThreadCardProps) {
  const [open, setOpen] = useState(false)
  const [replyOpen, setReplyOpen] = useState(false)
  const latestEmail = thread.emails[thread.emails.length - 1]
  const replyTo =
    latestEmail.direction === 'inbound' ? latestEmail.from_addr : (contactEmail ?? latestEmail.to_addr)

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      {/* Thread header */}
      <button
        className="flex w-full items-center justify-between gap-3 px-3 py-2.5 text-left hover:bg-slate-50 rounded-lg transition-colors"
        onClick={() => setOpen((v) => !v)}
      >
        <div className="flex items-center gap-2 min-w-0">
          <Mail className="h-4 w-4 shrink-0 text-slate-400" />
          <p className="font-medium text-sm text-slate-900 truncate">{thread.subject}</p>
          <span className="shrink-0 rounded-full bg-slate-100 px-1.5 py-0.5 text-xs text-slate-500">
            {thread.emails.length}
          </span>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <span className="text-xs text-slate-400">{formatRelativeTime(thread.latestAt)}</span>
          {open ? <ChevronUp className="h-4 w-4 text-slate-400" /> : <ChevronDown className="h-4 w-4 text-slate-400" />}
        </div>
      </button>

      {/* Thread messages */}
      {open && (
        <div className="border-t border-slate-100 px-3 divide-y divide-slate-50">
          {thread.emails.map((email) => (
            <EmailItem key={email.id} email={email} />
          ))}
          {/* Reply button */}
          <div className="py-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setReplyOpen(true)}
            >
              Reply
            </Button>
          </div>
        </div>
      )}

      <ComposeEmailModal
        open={replyOpen}
        onOpenChange={setReplyOpen}
        contactId={contactId}
        toEmail={replyTo}
        threadId={thread.threadId}
        replySubject={thread.subject.startsWith('Re:') ? thread.subject : `Re: ${thread.subject}`}
      />
    </div>
  )
}

// ── Main component ────────────────────────────────────────────────────────────

interface EmailTimelineProps {
  contactId: string
  contactEmail?: string
}

export function EmailTimeline({ contactId, contactEmail }: EmailTimelineProps) {
  const { data, isLoading } = useContactEmails(contactId)
  const [composeOpen, setComposeOpen] = useState(false)
  const emails = data?.data ?? []
  const threads = groupIntoThreads(emails)

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h3 className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
          <Mail className="h-4 w-4" />
          Emails
          {threads.length > 0 && (
            <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-normal text-slate-600">
              {threads.length}
            </span>
          )}
        </h3>
        <Button size="sm" variant="outline" onClick={() => setComposeOpen(true)}>
          <Plus className="h-3.5 w-3.5" />
          Compose
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-4">
          <Spinner />
        </div>
      ) : threads.length > 0 ? (
        <div className="space-y-2">
          {threads.map((thread) => (
            <ThreadCard
              key={thread.threadId}
              thread={thread}
              contactId={contactId}
              contactEmail={contactEmail}
            />
          ))}
        </div>
      ) : (
        <p className="text-sm text-slate-400">No emails yet.</p>
      )}

      <ComposeEmailModal
        open={composeOpen}
        onOpenChange={setComposeOpen}
        contactId={contactId}
        toEmail={contactEmail ?? ''}
      />
    </div>
  )
}
