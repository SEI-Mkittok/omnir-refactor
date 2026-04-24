import { useState } from 'react'
import { X, BookOpen, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useCreateTicket } from '@/hooks/useTickets'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { CustomFieldFormSection } from '@/components/omnir/CustomFieldRenderer'
import { useKbSuggest } from '@/hooks/useKB'
import { useDebounce } from '@/hooks/useDebounce'
import { useAuthStore } from '@/stores/auth'
import type { CreateTicketRequest, TicketStatus, TicketPriority, CustomFieldValues } from '@/api/types'

interface TicketFormProps {
  onClose: () => void
  initialValues?: Partial<CreateTicketRequest>
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

export function TicketForm({ onClose, initialValues, onCreated }: TicketFormProps) {
  const { mutateAsync: createTicket, isPending } = useCreateTicket()
  const { data: customFields = [] } = useCustomFieldDefinitions('ticket')
  const activeOrg = useAuthStore((s) => s.activeOrg)
  const user = useAuthStore((s) => s.user)
  const orgSlug = activeOrg?.slug ?? user?.email?.split('@')[1] ?? ''

  const [subject, setSubject] = useState(initialValues?.subject ?? '')
  const [status, setStatus] = useState<TicketStatus>(initialValues?.status ?? 'open')
  const [priority, setPriority] = useState<TicketPriority>(initialValues?.priority ?? 'medium')
  const [customFieldValues, setCustomFieldValues] = useState<CustomFieldValues>({})
  const [error, setError] = useState('')
  const [deflectionDismissed, setDeflectionDismissed] = useState(false)

  const debouncedSubject = useDebounce(subject, 300)
  const { data: suggestions = [] } = useKbSuggest(debouncedSubject)
  const showDeflection = !deflectionDismissed && suggestions.length > 0 && subject.length >= 3

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!subject.trim()) {
      setError('Subject is required.')
      return
    }
    setError('')
    const payload = {
      subject: subject.trim(),
      status,
      priority,
      ...(initialValues?.contact_id ? { contact_id: initialValues.contact_id } : {}),
      ...(initialValues?.account_id ? { account_id: initialValues.account_id } : {}),
      ...(initialValues?.assignee_id ? { assignee_id: initialValues.assignee_id } : {}),
      ...(Object.keys(customFieldValues).length ? { custom_fields: customFieldValues } : {}),
    } as CreateTicketRequest & { custom_fields?: CustomFieldValues }
    const ticket = await createTicket(payload as CreateTicketRequest)
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
        <div className="flex items-center justify-between border-b border-slate-200 px-4 sm:px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">New Ticket</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="space-y-4 p-4 sm:p-6">
          {/* Subject */}
          <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="ticket-subject">
              Subject <span className="text-red-500">*</span>
            </label>
            <input
              id="ticket-subject"
              type="text"
              value={subject}
              onChange={(e) => { setSubject(e.target.value); setDeflectionDismissed(false) }}
              placeholder="Describe the issue…"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
            />
            {error && <p className="text-xs text-red-600">{error}</p>}
          </div>

          {/* KB Deflection panel */}
          {showDeflection && (
            <div className="rounded-lg border border-[var(--color-primary-light)] bg-[var(--color-primary-light)] p-3 space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  <BookOpen className="h-4 w-4 text-[var(--color-primary)]" />
                  <span className="text-xs font-semibold text-[var(--color-primary)]">Did you find your answer?</span>
                </div>
                <button
                  type="button"
                  className="text-[var(--color-primary)] hover:text-[var(--color-primary)]"
                  onClick={() => setDeflectionDismissed(true)}
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
              <div className="space-y-1">
                {suggestions.map((s) => (
                  <a
                    key={s.id}
                    href={`/help/${orgSlug}/a/${s.slug}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 rounded-md px-2 py-1.5 text-xs text-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors group"
                  >
                    <ChevronRight className="h-3 w-3 text-[var(--color-primary)] shrink-0" />
                    <span className="flex-1 truncate">{s.title}</span>
                  </a>
                ))}
              </div>
              <button
                type="button"
                className="text-xs text-[var(--color-primary)] hover:text-[var(--color-primary)] underline"
                onClick={() => setDeflectionDismissed(true)}
              >
                None of these help — continue creating ticket
              </button>
            </div>
          )}

          {/* Status + Priority (side by side) */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="ticket-status">
                Status
              </label>
              <select
                id="ticket-status"
                value={status}
                onChange={(e) => setStatus(e.target.value as TicketStatus)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
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
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              >
                {PRIORITY_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <CustomFieldFormSection
            fields={customFields}
            values={customFieldValues}
            onChange={setCustomFieldValues}
          />

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
