import { useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useCreateTicket } from '@/hooks/useTickets'
import type { CreateTicketRequest, TicketStatus, TicketPriority } from '@/api/types'

interface TicketFormProps {
  onClose: () => void
  onCreated?: (ticketId: string) => void
}

const STATUS_OPTIONS: { label: string; value: TicketStatus }[] = [
  { label: 'Open', value: 'open' },
  { label: 'Pending', value: 'pending' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' },
]

const PRIORITY_OPTIONS: { label: string; value: TicketPriority }[] = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Critical', value: 'critical' },
]

export function TicketForm({ onClose, onCreated }: TicketFormProps) {
  const { mutateAsync: createTicket, isPending } = useCreateTicket()

  const [subject, setSubject] = useState('')
  const [status, setStatus] = useState<TicketStatus>('open')
  const [priority, setPriority] = useState<TicketPriority>('medium')
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!subject.trim()) {
      setError('Subject is required.')
      return
    }
    setError('')
    const payload: CreateTicketRequest = {
      subject: subject.trim(),
      status,
      priority,
    }
    const ticket = await createTicket(payload)
    onCreated?.(ticket.id)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />

      {/* Modal */}
      <div className="relative z-10 w-full max-w-md rounded-xl bg-white shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">New Ticket</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="space-y-4 p-6">
          {/* Subject */}
          <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="ticket-subject">
              Subject <span className="text-red-500">*</span>
            </label>
            <input
              id="ticket-subject"
              type="text"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="Describe the issue…"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            />
            {error && <p className="text-xs text-red-600">{error}</p>}
          </div>

          {/* Status + Priority (side by side) */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="ticket-status">
                Status
              </label>
              <select
                id="ticket-status"
                value={status}
                onChange={(e) => setStatus(e.target.value as TicketStatus)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                {STATUS_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="ticket-priority">
                Priority
              </label>
              <select
                id="ticket-priority"
                value={priority}
                onChange={(e) => setPriority(e.target.value as TicketPriority)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                {PRIORITY_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create Ticket'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
