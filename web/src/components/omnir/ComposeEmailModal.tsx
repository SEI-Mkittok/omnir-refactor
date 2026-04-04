import { useState } from 'react'
import { AlertCircle, Loader2, Send } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useSendEmail } from '@/hooks/useEmails'
import { useToast } from '@/components/ui/Toast'

interface ComposeEmailModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  contactId: string
  /** Pre-filled To address */
  toEmail?: string
  /** Pre-filled thread_id for replies */
  threadId?: string
  /** Pre-filled subject for replies */
  replySubject?: string
}

export function ComposeEmailModal({
  open,
  onOpenChange,
  contactId,
  toEmail = '',
  threadId = '',
  replySubject = '',
}: ComposeEmailModalProps) {
  const [to, setTo] = useState(toEmail)
  const [subject, setSubject] = useState(replySubject)
  const [body, setBody] = useState('')
  const [sendError, setSendError] = useState<string | null>(null)
  const sendEmail = useSendEmail()
  const { toast } = useToast()

  // Reset form when modal opens with new props
  const handleOpenChange = (value: boolean) => {
    if (!value) {
      setTo(toEmail)
      setSubject(replySubject)
      setBody('')
      setSendError(null)
    }
    onOpenChange(value)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!to.trim() || !subject.trim() || !body.trim()) return
    setSendError(null)
    try {
      await sendEmail.mutateAsync({
        contact_id: contactId,
        to: to.trim(),
        subject: subject.trim(),
        body: body.trim(),
        thread_id: threadId || undefined,
      })
      setBody('')
      onOpenChange(false)
      toast({ title: 'Email sent successfully' })
    } catch (err) {
      const msg =
        err instanceof Error ? err.message : 'Failed to send email. Please try again.'
      setSendError(msg)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Compose Email</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-3">
          {sendError && (
            <div className="flex items-start gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
              <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
              <span>{sendError}</span>
            </div>
          )}
          <div>
            <label className="block text-xs font-medium text-slate-500 mb-1">To</label>
            <Input
              type="email"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              placeholder="recipient@example.com"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-500 mb-1">Subject</label>
            <Input
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="Subject"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-500 mb-1">Body</label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Write your message…"
              rows={6}
              required
              className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm placeholder:text-slate-400 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={sendEmail.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={!to.trim() || !subject.trim() || !body.trim() || sendEmail.isPending}
            >
              {sendEmail.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              Send
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
