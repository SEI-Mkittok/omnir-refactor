import { useState, useRef, useCallback } from 'react'
import {
  User,
  Mail,
  Building2,
  Tag,
  TrendingUp,
  StickyNote,
  Pencil,
  Check,
  X,
  Plus,
  Loader2,
} from 'lucide-react'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { cn, formatDate, formatRelativeTime, formatCurrency } from '@/lib/utils'
import { stageBadgeVariant, stageLabel } from '@/components/omnir/ContactCard'
import {
  useContact,
  useUpdateContact,
  useDeleteContact,
  useContactNotes,
  useAddContactNote,
} from '@/hooks/useContacts'
import type { Contact, ContactStage, UpdateContactRequest } from '@/api/types'

// ── Editable field ──────────────────────────────────────────────────────────

interface EditableFieldProps {
  label: string
  value: string | undefined | null
  onSave: (val: string) => Promise<void>
  placeholder?: string
  type?: string
}

function EditableField({
  label,
  value,
  onSave,
  placeholder = '—',
  type = 'text',
}: EditableFieldProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const startEdit = useCallback(() => {
    setDraft(value ?? '')
    setEditing(true)
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [value])

  const cancel = useCallback(() => {
    setEditing(false)
    setDraft('')
  }, [])

  const commit = useCallback(async () => {
    if (draft === (value ?? '')) {
      cancel()
      return
    }
    setSaving(true)
    try {
      await onSave(draft)
      setEditing(false)
    } finally {
      setSaving(false)
    }
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
            <Input
              ref={inputRef}
              type={type}
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={handleKeyDown}
              className="h-7 text-sm"
              disabled={saving}
            />
            <button
              onClick={commit}
              disabled={saving}
              className="flex h-6 w-6 items-center justify-center rounded bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50"
              title="Save"
            >
              {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
            </button>
            <button
              onClick={cancel}
              disabled={saving}
              className="flex h-6 w-6 items-center justify-center rounded border border-slate-200 text-slate-500 hover:bg-slate-50 disabled:opacity-50"
              title="Cancel"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ) : (
          <div className="mt-0.5 flex items-center gap-1.5 min-h-[1.5rem]">
            <span className={cn('text-sm', value ? 'text-slate-900' : 'text-slate-400')}>
              {value || placeholder}
            </span>
            <button
              onClick={startEdit}
              className="invisible group-hover:visible flex h-5 w-5 items-center justify-center rounded text-slate-400 hover:bg-slate-100 hover:text-slate-600"
              title={`Edit ${label}`}
            >
              <Pencil className="h-3 w-3" />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Stage selector ──────────────────────────────────────────────────────────

const STAGES: ContactStage[] = ['lead', 'prospect', 'customer', 'churned']

interface StageSelectorProps {
  current: ContactStage
  onSave: (stage: ContactStage) => Promise<void>
}

function StageSelector({ current, onSave }: StageSelectorProps) {
  const [open, setOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const select = async (stage: ContactStage) => {
    if (stage === current) { setOpen(false); return }
    setSaving(true)
    try {
      await onSave(stage)
    } finally {
      setSaving(false)
      setOpen(false)
    }
  }

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        disabled={saving}
        className="flex items-center gap-1.5"
      >
        <Badge variant={stageBadgeVariant[current]} className="cursor-pointer text-sm px-3 py-1">
          {stageLabel[current]}
        </Badge>
        {saving && <Loader2 className="h-3.5 w-3.5 animate-spin text-slate-400" />}
        <span className="text-xs text-slate-400">▾</span>
      </button>
      {open && (
        <div className="absolute left-0 top-full z-10 mt-1 rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
          {STAGES.map((s) => (
            <button
              key={s}
              onClick={() => select(s)}
              className={cn(
                'flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-slate-50',
                s === current && 'bg-slate-50 font-semibold'
              )}
            >
              <Badge variant={stageBadgeVariant[s]} className="pointer-events-none">
                {stageLabel[s]}
              </Badge>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

// ── Notes section ───────────────────────────────────────────────────────────

function NotesSection({ contactId }: { contactId: string }) {
  const { data: notes, isLoading } = useContactNotes(contactId)
  const addNote = useAddContactNote()
  const [draft, setDraft] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!draft.trim()) return
    setSubmitting(true)
    try {
      await addNote.mutateAsync({ contactId, payload: { content: draft.trim() } })
      setDraft('')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
        <StickyNote className="h-4 w-4" />
        Notes
      </h3>

      {/* Add note form */}
      <form onSubmit={handleSubmit} className="mb-4">
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Add a note…"
          rows={2}
          className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm placeholder:text-slate-400 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
          onKeyDown={(e) => {
            if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleSubmit(e as unknown as React.FormEvent)
          }}
        />
        <div className="mt-1.5 flex items-center justify-between">
          <span className="text-xs text-slate-400">⌘↵ to save</span>
          <Button type="submit" size="sm" disabled={!draft.trim() || submitting}>
            {submitting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            Add Note
          </Button>
        </div>
      </form>

      {/* Notes list */}
      {isLoading ? (
        <Spinner />
      ) : notes && notes.length > 0 ? (
        <ul className="space-y-2">
          {notes.map((note) => (
            <li key={note.id} className="rounded-lg border border-slate-100 bg-slate-50 px-3 py-2.5">
              <p className="text-sm text-slate-800 whitespace-pre-wrap">{note.content}</p>
              <p className="mt-1 text-xs text-slate-400">{formatRelativeTime(note.created_at)}</p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-slate-400">No notes yet.</p>
      )}
    </div>
  )
}

// ── Main panel ──────────────────────────────────────────────────────────────

interface ContactDetailPanelProps {
  contactId: string
  onClose: () => void
}

export function ContactDetailPanel({ contactId, onClose }: ContactDetailPanelProps) {
  const { data: contact, isLoading } = useContact(contactId)
  const updateContact = useUpdateContact()
  const deleteContact = useDeleteContact()

  const patch = useCallback(
    async (payload: UpdateContactRequest) => {
      await updateContact.mutateAsync({ id: contactId, payload })
    },
    [contactId, updateContact]
  )

  if (isLoading) {
    return (
      <SidePanel open title="Contact" onClose={onClose}>
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      </SidePanel>
    )
  }

  if (!contact) return null

  const fullName = `${contact.first_name} ${contact.last_name}`

  return (
    <SidePanel
      open
      title={fullName}
      onClose={onClose}
      width="lg"
      actions={
        <Button
          variant="destructive"
          size="sm"
          onClick={async () => {
            if (confirm(`Delete ${fullName}? This cannot be undone.`)) {
              await deleteContact.mutateAsync(contact.id)
              onClose()
            }
          }}
        >
          Delete
        </Button>
      }
    >
      <div className="space-y-6">
        {/* Avatar + stage */}
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-indigo-700 text-xl font-bold">
            {contact.first_name[0]}{contact.last_name[0]}
          </div>
          <div className="min-w-0">
            <p className="text-lg font-semibold text-slate-900 truncate">{fullName}</p>
            <div className="mt-1">
              <StageSelector
                current={contact.stage}
                onSave={(stage) => patch({ stage })}
              />
            </div>
          </div>
        </div>

        {/* Core fields — inline editable */}
        <div>
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
            <User className="h-4 w-4" />
            Details
          </h3>
          <div className="grid grid-cols-2 gap-x-6 gap-y-1">
            <EditableField
              label="First Name"
              value={contact.first_name}
              onSave={(v) => patch({ first_name: v })}
            />
            <EditableField
              label="Last Name"
              value={contact.last_name}
              onSave={(v) => patch({ last_name: v })}
            />
            <EditableField
              label="Title"
              value={contact.title}
              onSave={(v) => patch({ title: v })}
              placeholder="No title"
            />
            <EditableField
              label="Department"
              value={contact.department}
              onSave={(v) => patch({ department: v })}
              placeholder="No department"
            />
          </div>
        </div>

        {/* Contact info */}
        <div>
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
            <Mail className="h-4 w-4" />
            Contact
          </h3>
          <div className="grid grid-cols-2 gap-x-6 gap-y-1">
            <EditableField
              label="Email"
              value={contact.email}
              onSave={(v) => patch({ email: v })}
              type="email"
            />
            <EditableField
              label="Phone"
              value={contact.phone}
              onSave={(v) => patch({ phone: v })}
              placeholder="No phone"
              type="tel"
            />
          </div>
        </div>

        {/* Linked account */}
        <div>
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
            <Building2 className="h-4 w-4" />
            Account
          </h3>
          {contact.account ? (
            <div className="flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5">
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-indigo-100">
                <Building2 className="h-4 w-4 text-indigo-600" />
              </div>
              <div className="min-w-0">
                <p className="font-medium text-slate-900 truncate">{contact.account.name}</p>
                {contact.account.industry && (
                  <p className="text-xs text-slate-500 truncate">{contact.account.industry}</p>
                )}
                {contact.account.domain && (
                  <a
                    href={`https://${contact.account.domain}`}
                    target="_blank"
                    rel="noreferrer"
                    className="text-xs text-indigo-600 hover:underline"
                  >
                    {contact.account.domain}
                  </a>
                )}
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-400">No account linked.</p>
          )}
        </div>

        {/* Associated deals */}
        <div>
          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
            <TrendingUp className="h-4 w-4" />
            Deals
            {contact.deals && contact.deals.length > 0 && (
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-normal text-slate-600">
                {contact.deals.length}
              </span>
            )}
          </h3>
          {contact.deals && contact.deals.length > 0 ? (
            <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200">
              {contact.deals.map((deal) => (
                <li key={deal.id} className="flex items-center justify-between px-3 py-2.5 text-sm">
                  <div className="min-w-0">
                    <p className="font-medium text-slate-900 truncate">{deal.title}</p>
                    {deal.account?.name && (
                      <p className="text-xs text-slate-400 truncate">{deal.account.name}</p>
                    )}
                  </div>
                  <div className="ml-3 flex shrink-0 items-center gap-2">
                    {deal.value > 0 && (
                      <span className="font-semibold text-indigo-600">
                        {formatCurrency(deal.value, deal.currency)}
                      </span>
                    )}
                    <Badge variant="default">{deal.stage.replace('_', ' ')}</Badge>
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No deals associated.</p>
          )}
        </div>

        {/* Tags */}
        {contact.tags && contact.tags.length > 0 && (
          <div>
            <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
              <Tag className="h-4 w-4" />
              Tags
            </h3>
            <div className="flex flex-wrap gap-1.5">
              {contact.tags.map((tag) => (
                <Badge key={tag} variant="default">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Metadata */}
        <div className="rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-400">
          <span>Created {formatDate(contact.created_at)}</span>
          {contact.updated_at !== contact.created_at && (
            <span className="ml-3">Updated {formatRelativeTime(contact.updated_at)}</span>
          )}
        </div>

        {/* Notes */}
        <NotesSection contactId={contactId} />
      </div>
    </SidePanel>
  )
}
