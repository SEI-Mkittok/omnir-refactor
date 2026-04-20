import { useState } from 'react'
import { Lock, Globe, Send } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import { Button } from '@/components/ui/Button'
import { useTicketComments, useAddTicketComment } from '@/hooks/useTickets'
import type { TicketComment } from '@/api/types'
import { cn } from '@/lib/utils'

interface CommentThreadProps {
  ticketId: string
}

function CommentBubble({ comment }: { comment: TicketComment }) {
  return (
    <div
      className={cn(
        'rounded-lg border p-3 space-y-1',
        comment.is_internal
          ? 'border-amber-200 bg-amber-50'
          : 'border-slate-200 bg-white'
      )}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5">
          <span className="text-sm font-medium text-slate-800">
            {comment.author?.name ?? 'Agent'}
          </span>
          {comment.is_internal && (
            <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700">
              <Lock className="h-3 w-3" />
              Internal
            </span>
          )}
        </div>
        <span className="text-xs text-slate-400">{formatDate(comment.created_at)}</span>
      </div>
      <p className="whitespace-pre-wrap text-sm text-slate-700">{comment.body}</p>
    </div>
  )
}

export function CommentThread({ ticketId }: CommentThreadProps) {
  const { data: commentsData, isLoading } = useTicketComments(ticketId)
  const { mutateAsync: addComment, isPending } = useAddTicketComment()
  const comments = Array.isArray(commentsData) ? commentsData : []

  const [body, setBody] = useState('')
  const [isInternal, setIsInternal] = useState(false)
  const [showInternalOnly, setShowInternalOnly] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!body.trim()) return
    await addComment({ ticketId, payload: { body: body.trim(), is_internal: isInternal } })
    setBody('')
  }

  const visible = showInternalOnly
    ? comments.filter((c) => c.is_internal)
    : comments

  return (
    <div className="space-y-4">
      {/* Header + filter toggle */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-700">
          Comments {comments.length > 0 && `(${comments.length})`}
        </h3>
        <button
          type="button"
          onClick={() => setShowInternalOnly((v) => !v)}
          className={cn(
            'inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium transition-colors',
            showInternalOnly
              ? 'bg-amber-100 text-amber-700'
              : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
          )}
        >
          <Lock className="h-3 w-3" />
          {showInternalOnly ? 'Internal only' : 'All comments'}
        </button>
      </div>

      {/* Comment list */}
      {isLoading ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : visible.length === 0 ? (
        <p className="text-sm text-slate-400">No comments yet.</p>
      ) : (
        <div className="space-y-2">
          {visible.map((c) => (
            <CommentBubble key={c.id} comment={c} />
          ))}
        </div>
      )}

      {/* Compose */}
      <form onSubmit={handleSubmit} className="space-y-2">
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Write a reply…"
          rows={3}
          className="w-full resize-none rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        />
        <div className="flex items-center justify-between">
          {/* Internal toggle */}
          <button
            type="button"
            onClick={() => setIsInternal((v) => !v)}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors',
              isInternal
                ? 'bg-amber-100 text-amber-700 ring-1 ring-amber-300'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            )}
          >
            {isInternal ? (
              <>
                <Lock className="h-3 w-3" />
                Internal note
              </>
            ) : (
              <>
                <Globe className="h-3 w-3" />
                Public reply
              </>
            )}
          </button>

          <Button type="submit" size="sm" disabled={!body.trim() || isPending}>
            <Send className="h-3.5 w-3.5" />
            {isPending ? 'Sending…' : 'Send'}
          </Button>
        </div>
      </form>
    </div>
  )
}
