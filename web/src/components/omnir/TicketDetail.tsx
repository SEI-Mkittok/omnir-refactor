import { useState } from 'react'
import { Trash2, Paperclip, Clock } from 'lucide-react'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { FileUpload } from '@/components/ui/FileUpload'
import { CommentThread } from './CommentThread'
import { CustomFieldDisplaySection } from './CustomFieldRenderer'
import { formatDate } from '@/lib/utils'
import {
  useTicket,
  useUpdateTicket,
  useDeleteTicket,
  useTicketAttachments,
  useUploadTicketAttachment,
} from '@/hooks/useTickets'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import type { TicketStatus, TicketPriority, CustomFieldValues, TicketSLA, SLATrackingStatus } from '@/api/types'
import { cn } from '@/lib/utils'

// ---- SLA helpers ----

const slaStatusVariant: Record<SLATrackingStatus, 'green' | 'yellow' | 'red'> = {
  on_track: 'green',
  at_risk: 'yellow',
  breached: 'red',
}

const slaStatusLabel: Record<SLATrackingStatus, string> = {
  on_track: 'On track',
  at_risk: 'At risk',
  breached: 'Breached',
}

function formatDeadline(iso: string): string {
  const date = new Date(iso)
  return date.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function SLAProgressBar({ deadline, breached }: { deadline: string; breached: boolean }) {
  // Show how much time has been used relative to the deadline from created_at
  // Since we only have the deadline, we show how far past/before we are vs now
  const now = Date.now()
  const deadlineMs = new Date(deadline).getTime()
  const isOver = now > deadlineMs

  if (breached || isOver) {
    return <div className="h-1.5 rounded-full bg-red-500 w-full" />
  }

  return <div className="h-1.5 rounded-full bg-green-500 w-3/4" />
}

function SLAWidget({ sla }: { sla: TicketSLA }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <Clock className="h-3.5 w-3.5 text-slate-500" />
          <span className="text-xs font-medium text-slate-700">{sla.policy_name}</span>
        </div>
        <Badge variant={slaStatusVariant[sla.status]}>{slaStatusLabel[sla.status]}</Badge>
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500">First response</span>
          <span className={cn('font-medium', sla.response_breached ? 'text-red-600' : 'text-slate-700')}>
            {sla.response_breached ? 'Breached · ' : ''}{formatDeadline(sla.response_deadline)}
          </span>
        </div>
        <SLAProgressBar deadline={sla.response_deadline} breached={sla.response_breached} />
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500">Resolution</span>
          <span className={cn('font-medium', sla.resolution_breached ? 'text-red-600' : 'text-slate-700')}>
            {sla.resolution_breached ? 'Breached · ' : ''}{formatDeadline(sla.resolution_deadline)}
          </span>
        </div>
        <SLAProgressBar deadline={sla.resolution_deadline} breached={sla.resolution_breached} />
      </div>
    </div>
  )
}

// ---- Badge helpers ----

export const statusBadgeVariant: Record<TicketStatus, 'blue' | 'yellow' | 'green' | 'gray'> = {
  open: 'blue',
  pending: 'yellow',
  resolved: 'green',
  closed: 'gray',
}

export const statusLabel: Record<TicketStatus, string> = {
  open: 'Open',
  pending: 'Pending',
  resolved: 'Resolved',
  closed: 'Closed',
}

export const priorityBadgeVariant: Record<TicketPriority, 'gray' | 'blue' | 'orange' | 'red'> = {
  low: 'gray',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

export const priorityLabel: Record<TicketPriority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  critical: 'Critical',
}

// ---- Sub-components ----

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-2 py-2">
      <span className="shrink-0 text-xs font-medium text-slate-500 w-24">{label}</span>
      <div className="flex-1 text-right text-sm text-slate-800">{children}</div>
    </div>
  )
}

function AttachmentsSection({ ticketId }: { ticketId: string }) {
  const { data: attachments = [] } = useTicketAttachments(ticketId)
  const { mutateAsync: upload, isPending } = useUploadTicketAttachment()
  const [showUpload, setShowUpload] = useState(false)

  const handleFiles = async (files: File[]) => {
    for (const file of files) {
      await upload({ ticketId, file })
    }
    setShowUpload(false)
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-700">
          Attachments {attachments.length > 0 && `(${attachments.length})`}
        </h3>
        <Button variant="ghost" size="sm" onClick={() => setShowUpload((v) => !v)}>
          <Paperclip className="h-3.5 w-3.5" />
          Attach
        </Button>
      </div>

      {showUpload && (
        <FileUpload
          onFiles={handleFiles}
          disabled={isPending}
          className="mt-2"
        />
      )}

      {attachments.length > 0 && (
        <ul className="space-y-1">
          {attachments.map((a) => (
            <li
              key={a.id}
              className="flex items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2"
            >
              <Paperclip className="h-3.5 w-3.5 shrink-0 text-slate-400" />
              <span className="truncate text-sm text-slate-700">{a.filename}</span>
              {a.size_bytes != null && (
                <span className="ml-auto shrink-0 text-xs text-slate-400">
                  {(a.size_bytes / 1024).toFixed(0)} KB
                </span>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

// ---- Main ----

interface TicketDetailProps {
  ticketId: string
  onClose: () => void
}

export function TicketDetail({ ticketId, onClose }: TicketDetailProps) {
  const { data: ticket, isLoading } = useTicket(ticketId)
  const { mutateAsync: updateTicket } = useUpdateTicket()
  const { mutateAsync: deleteTicket } = useDeleteTicket()
  const { data: customFields = [] } = useCustomFieldDefinitions('ticket')

  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleStatusChange = (status: TicketStatus) => {
    updateTicket({ id: ticketId, payload: { status } })
  }

  const handlePriorityChange = (priority: TicketPriority) => {
    updateTicket({ id: ticketId, payload: { priority } })
  }

  const handleDelete = async () => {
    await deleteTicket(ticketId)
    onClose()
  }

  return (
    <SidePanel
      open
      onClose={onClose}
      title={ticket?.subject ?? 'Ticket'}
      width="lg"
      actions={
        <Button
          variant="ghost"
          size="icon"
          className="text-red-500 hover:text-red-600"
          onClick={() => setConfirmDelete(true)}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      }
    >
      {isLoading || !ticket ? (
        <div className="flex items-center justify-center py-12">
          <span className="text-sm text-slate-500">Loading…</span>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Metadata */}
          <div className="rounded-lg border border-slate-200 divide-y divide-slate-100">
            <Field label="Status">
              <select
                value={ticket.status}
                onChange={(e) => handleStatusChange(e.target.value as TicketStatus)}
                className="rounded border border-transparent bg-transparent text-sm font-medium focus:border-slate-300 focus:outline-none focus:ring-0"
              >
                {(['open', 'pending', 'resolved', 'closed'] as TicketStatus[]).map((s) => (
                  <option key={s} value={s}>
                    {statusLabel[s]}
                  </option>
                ))}
              </select>
            </Field>
            <Field label="Priority">
              <select
                value={ticket.priority}
                onChange={(e) => handlePriorityChange(e.target.value as TicketPriority)}
                className="rounded border border-transparent bg-transparent text-sm font-medium focus:border-slate-300 focus:outline-none focus:ring-0"
              >
                {(['low', 'medium', 'high', 'critical'] as TicketPriority[]).map((p) => (
                  <option key={p} value={p}>
                    {priorityLabel[p]}
                  </option>
                ))}
              </select>
            </Field>
            {ticket.assignee && (
              <Field label="Assigned to">
                <span>{ticket.assignee.name}</span>
              </Field>
            )}
            {ticket.contact && (
              <Field label="Contact">
                <span>
                  {ticket.contact.first_name} {ticket.contact.last_name}
                </span>
              </Field>
            )}
            {ticket.account && (
              <Field label="Account">
                <span>{ticket.account.name}</span>
              </Field>
            )}
            {ticket.source && (
              <Field label="Source">
                <span className="capitalize">{ticket.source}</span>
              </Field>
            )}
            <Field label="Created">
              <span className="text-slate-500">{formatDate(ticket.created_at)}</span>
            </Field>
            <Field label="Updated">
              <span className="text-slate-500">{formatDate(ticket.updated_at)}</span>
            </Field>
          </div>

          {/* Custom fields */}
          {customFields.length > 0 && ticket.custom_fields && (
            <div className="rounded-lg border border-slate-200 px-4 py-3 space-y-2">
              <CustomFieldDisplaySection
                fields={customFields}
                values={ticket.custom_fields as CustomFieldValues}
              />
            </div>
          )}

          {/* SLA Widget */}
          {ticket.sla && <SLAWidget sla={ticket.sla} />}

          {/* Attachments */}
          <AttachmentsSection ticketId={ticketId} />

          {/* Divider */}
          <hr className="border-slate-200" />

          {/* Comments */}
          <CommentThread ticketId={ticketId} />
        </div>
      )}

      {/* Delete confirmation */}
      {confirmDelete && (
        <div className="fixed inset-0 z-60 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/40" onClick={() => setConfirmDelete(false)} />
          <div className="relative z-10 w-full max-w-sm rounded-xl bg-white p-6 shadow-xl space-y-4">
            <h3 className="text-base font-semibold text-slate-900">Delete ticket?</h3>
            <p className="text-sm text-slate-600">
              This action cannot be undone. All comments and attachments will be removed.
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setConfirmDelete(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={handleDelete}>
                Delete
              </Button>
            </div>
          </div>
        </div>
      )}
    </SidePanel>
  )
}
