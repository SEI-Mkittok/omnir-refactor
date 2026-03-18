import { useState, useRef, useCallback } from 'react'
import {
  User,
  Mail,
  Pencil,
  Check,
  X,
  Loader2,
  ArrowRightLeft,
  Minus,
  Plus,
} from 'lucide-react'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { cn, formatDate, formatRelativeTime } from '@/lib/utils'
import { useLead, useUpdateLead, useDeleteLead } from '@/hooks/useLeads'
import { LeadConvertModal } from '@/components/omnir/LeadConvertModal'
import { LeadScoreBadge } from '@/components/omnir/LeadScoreBadge'
import { CustomFieldEditableSection } from '@/components/omnir/CustomFieldRenderer'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import type { Lead, LeadStatus, UpdateLeadRequest, CustomFieldValues } from '@/api/types'

// ── Status badge helpers ─────────────────────────────────────────────────────

export const leadStatusBadgeVariant: Record<LeadStatus, 'default' | 'blue' | 'yellow' | 'green' | 'red' | 'gray' | 'indigo' | 'purple' | 'orange'> = {
  new: 'blue',
  contacted: 'yellow',
  qualified: 'green',
  unqualified: 'red',
  converted: 'gray',
}

export const leadStatusLabel: Record<LeadStatus, string> = {
  new: 'New',
  contacted: 'Contacted',
  qualified: 'Qualified',
  unqualified: 'Unqualified',
  converted: 'Converted',
}

// ── Editable field (reused from ContactDetailPanel pattern) ──────────────────

interface EditableFieldProps {
  label: string
  value: string | undefined | null
  onSave: (val: string) => Promise<void>
  placeholder?: string
  type?: string
}

function EditableField({ label, value, onSave, placeholder = '—', type = 'text' }: EditableFieldProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const startEdit = useCallback(() => {
    setDraft(value ?? '')
    setEditing(true)
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [value])

  const cancel = useCallback(() => { setEditing(false); setDraft('') }, [])

  const commit = useCallback(async () => {
    if (draft === (value ?? '')) { cancel(); return }
    setSaving(true)
    try { await onSave(draft); setEditing(false) } finally { setSaving(false) }
  }, [draft, value, onSave, cancel])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') commit()
    if (e.key === 'Escape') cancel()
  }

  return (
    <div className="group flex items-start gap-1 py-1">
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-slate-400 uppercase tracking-wide">{label}</p>
        {editing ? (
          <div className="mt-1 flex items-center gap-1.5">
            <Input ref={inputRef} type={type} value={draft} onChange={(e) => setDraft(e.target.value)} onKeyDown={handleKeyDown} className="h-7 text-sm" disabled={saving} />
            <button onClick={commit} disabled={saving} className="flex h-6 w-6 items-center justify-center rounded bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50" title="Save">
              {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
            </button>
            <button onClick={cancel} disabled={saving} className="flex h-6 w-6 items-center justify-center rounded border border-slate-200 text-slate-500 hover:bg-slate-50 disabled:opacity-50" title="Cancel">
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ) : (
          <div className="mt-0.5 flex items-center gap-1.5 min-h-[1.5rem]">
            <span className={cn('text-sm', value ? 'text-slate-900' : 'text-slate-400')}>{value || placeholder}</span>
            <button onClick={startEdit} className="invisible group-hover:visible flex h-5 w-5 items-center justify-center rounded text-slate-400 hover:bg-slate-100 hover:text-slate-600" title={`Edit ${label}`}>
              <Pencil className="h-3 w-3" />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Status selector ──────────────────────────────────────────────────────────

const STATUSES: LeadStatus[] = ['new', 'contacted', 'qualified', 'unqualified']

function StatusSelector({ current, onSave }: { current: LeadStatus; onSave: (s: LeadStatus) => Promise<void> }) {
  const [open, setOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const select = async (s: LeadStatus) => {
    if (s === current) { setOpen(false); return }
    setSaving(true)
    try { await onSave(s) } finally { setSaving(false); setOpen(false) }
  }

  return (
    <div className="relative">
      <button onClick={() => setOpen((o) => !o)} disabled={saving} className="flex items-center gap-1.5">
        <Badge variant={leadStatusBadgeVariant[current]} className="cursor-pointer text-sm px-3 py-1">
          {leadStatusLabel[current]}
        </Badge>
        {saving && <Loader2 className="h-3.5 w-3.5 animate-spin text-slate-400" />}
        <span className="text-xs text-slate-400">▾</span>
      </button>
      {open && (
        <div className="absolute left-0 top-full z-10 mt-1 rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
          {STATUSES.map((s) => (
            <button key={s} onClick={() => select(s)} className={cn('flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-slate-50', s === current && 'bg-slate-50 font-semibold')}>
              <Badge variant={leadStatusBadgeVariant[s]} className="pointer-events-none">{leadStatusLabel[s]}</Badge>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

// ── Score stepper widget ─────────────────────────────────────────────────────

interface ScoreStepperProps {
  score: number
  onSave: (delta: number) => Promise<void>
}

function ScoreStepper({ score, onSave }: ScoreStepperProps) {
  const [pending, setPending] = useState<'dec' | 'inc' | null>(null)

  const step = async (delta: number) => {
    const newScore = Math.max(0, Math.min(100, score + delta))
    if (newScore === score) return
    setPending(delta < 0 ? 'dec' : 'inc')
    try {
      await onSave(delta)
    } finally {
      setPending(null)
    }
  }

  return (
    <div className="flex items-center gap-2">
      <LeadScoreBadge score={score} className="text-sm" />
      <div className="flex items-center rounded-md border border-slate-200 overflow-hidden">
        <button
          onClick={() => step(-5)}
          disabled={score <= 0 || pending !== null}
          className="flex h-7 w-7 items-center justify-center bg-white text-slate-500 hover:bg-slate-50 disabled:opacity-40 transition-colors"
          title="Decrease score by 5"
        >
          {pending === 'dec' ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Minus className="h-3.5 w-3.5" />}
        </button>
        <span className="w-px h-5 bg-slate-200" />
        <button
          onClick={() => step(5)}
          disabled={score >= 100 || pending !== null}
          className="flex h-7 w-7 items-center justify-center bg-white text-slate-500 hover:bg-slate-50 disabled:opacity-40 transition-colors"
          title="Increase score by 5"
        >
          {pending === 'inc' ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
        </button>
      </div>
    </div>
  )
}

// ── Main panel ───────────────────────────────────────────────────────────────

interface LeadDetailPanelProps {
  leadId: string
  onClose: () => void
}

export function LeadDetailPanel({ leadId, onClose }: LeadDetailPanelProps) {
  const { data: lead, isLoading } = useLead(leadId)
  const updateLead = useUpdateLead()
  const deleteLead = useDeleteLead()
  const { data: customFields = [] } = useCustomFieldDefinitions('lead')
  const [showConvert, setShowConvert] = useState(false)

  const patch = useCallback(
    async (payload: UpdateLeadRequest) => {
      await updateLead.mutateAsync({ id: leadId, payload })
    },
    [leadId, updateLead]
  )

  // Optimistic-style score update via delta (clamped server-side too)
  const patchScore = useCallback(
    async (delta: number) => {
      if (!lead) return
      const newScore = Math.max(0, Math.min(100, lead.lead_score + delta))
      await patch({ lead_score: newScore })
    },
    [lead, patch]
  )

  if (isLoading) {
    return (
      <SidePanel open title="Lead" onClose={onClose}>
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      </SidePanel>
    )
  }

  if (!lead) return null

  const fullName = `${lead.first_name} ${lead.last_name}`
  const isConverted = !!lead.converted_contact_id

  return (
    <>
      <SidePanel
        open
        title={fullName}
        onClose={onClose}
        width="lg"
        actions={
          <div className="flex items-center gap-2">
            {!isConverted && (
              <Button size="sm" onClick={() => setShowConvert(true)}>
                <ArrowRightLeft className="h-3.5 w-3.5" />
                Convert
              </Button>
            )}
            <Button
              variant="destructive"
              size="sm"
              onClick={async () => {
                if (confirm(`Delete ${fullName}? This cannot be undone.`)) {
                  await deleteLead.mutateAsync(lead.id)
                  onClose()
                }
              }}
            >
              Delete
            </Button>
          </div>
        }
      >
        <div className="space-y-6">
          {/* Converted banner */}
          {isConverted && (
            <div className="flex items-center gap-2 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700">
              <ArrowRightLeft className="h-4 w-4 shrink-0" />
              Converted to contact
            </div>
          )}

          {/* Avatar + status */}
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-violet-100 text-violet-700 text-xl font-bold">
              {lead.first_name[0]}{lead.last_name[0]}
            </div>
            <div className="min-w-0 space-y-1.5">
              <p className="text-lg font-semibold text-slate-900 truncate">{fullName}</p>
              <StatusSelector
                current={lead.status}
                onSave={(status) => patch({ status })}
              />
            </div>
          </div>

          {/* Lead score */}
          <div>
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
              Lead Score
            </h3>
            <ScoreStepper score={lead.lead_score} onSave={patchScore} />
          </div>

          {/* Details */}
          <div>
            <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
              <User className="h-4 w-4" />
              Details
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1">
              <EditableField label="First Name" value={lead.first_name} onSave={(v) => patch({ first_name: v })} />
              <EditableField label="Last Name" value={lead.last_name} onSave={(v) => patch({ last_name: v })} />
              <EditableField label="Company" value={lead.company} onSave={(v) => patch({ company: v })} placeholder="No company" />
              <div className="py-1">
                <p className="text-xs font-medium text-slate-400 uppercase tracking-wide">Source</p>
                <div className="mt-0.5">
                  {lead.lead_source ? (
                    <span className="inline-flex items-center rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-600">
                      {lead.lead_source}
                    </span>
                  ) : (
                    <EditableField label="" value={lead.lead_source} onSave={(v) => patch({ lead_source: v })} placeholder="No source" />
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Contact info */}
          <div>
            <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
              <Mail className="h-4 w-4" />
              Contact
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1">
              <EditableField label="Email" value={lead.email} onSave={(v) => patch({ email: v })} type="email" />
              <EditableField label="Phone" value={lead.phone} onSave={(v) => patch({ phone: v })} placeholder="No phone" type="tel" />
            </div>
          </div>

          {/* Custom fields */}
          <CustomFieldEditableSection
            fields={customFields}
            values={lead.custom_fields as CustomFieldValues | undefined}
            onSave={async (cf) => { await patch({ custom_fields: cf as Record<string, unknown> }) }}
          />

          {/* Metadata */}
          <div className="rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-400">
            <span>Created {formatDate(lead.created_at)}</span>
            {lead.updated_at !== lead.created_at && (
              <span className="ml-3">Updated {formatRelativeTime(lead.updated_at)}</span>
            )}
          </div>

        </div>
      </SidePanel>

      {showConvert && (
        <LeadConvertModal
          lead={lead}
          open={showConvert}
          onClose={() => setShowConvert(false)}
          onConverted={() => onClose()}
        />
      )}
    </>
  )
}
