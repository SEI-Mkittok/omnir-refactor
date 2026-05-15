import { useState } from 'react'
import axios from 'axios'
import { Plus, Pencil, Trash2, SlidersHorizontal } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/Dialog'
import {
  useCustomFieldDefinitions,
  useCreateCustomFieldDefinition,
  useUpdateCustomFieldDefinition,
  useDeleteCustomFieldDefinition,
} from '@/hooks/useCustomFields'
import type {
  CustomFieldDefinition,
  CustomFieldEntityType,
  CustomFieldType,
  CreateCustomFieldDefinitionRequest,
  UpdateCustomFieldDefinitionRequest,
} from '@/api/types'

const ENTITY_TABS: { value: CustomFieldEntityType; label: string }[] = [
  { value: 'ticket', label: 'Tickets' },
  { value: 'contact', label: 'Contacts' },
  { value: 'lead', label: 'Leads' },
  { value: 'deal', label: 'Deals' },
  { value: 'account', label: 'Accounts' },
  { value: 'quote', label: 'Quotes' },
  { value: 'kb_article', label: 'KB Articles' },
]

const FIELD_TYPE_OPTIONS: { value: CustomFieldType; label: string }[] = [
  { value: 'text', label: 'Text' },
  { value: 'number', label: 'Number' },
  { value: 'date', label: 'Date' },
  { value: 'url', label: 'URL' },
  { value: 'select', label: 'Single select' },
  { value: 'multiselect', label: 'Multi-select' },
  { value: 'checkbox', label: 'Checkbox' },
]

const fieldTypeBadge: Record<CustomFieldType, 'default' | 'blue' | 'green' | 'yellow' | 'indigo' | 'purple' | 'orange'> = {
  text: 'default',
  number: 'blue',
  date: 'green',
  url: 'indigo',
  select: 'yellow',
  multiselect: 'orange',
  checkbox: 'purple',
}

const selectClass =
  'flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]'

function getApiErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const message = err.response?.data?.error
    if (typeof message === 'string' && message.trim()) {
      return message
    }
  }
  return fallback
}

// ── Options editor (for select / multiselect) ─────────────────────────────────

interface OptionsEditorProps {
  options: string[]
  onChange: (options: string[]) => void
}

function OptionsEditor({ options, onChange }: OptionsEditorProps) {
  const [draft, setDraft] = useState('')

  const add = () => {
    const v = draft.trim()
    if (!v || options.includes(v)) return
    onChange([...options, v])
    setDraft('')
  }

  const remove = (opt: string) => onChange(options.filter((o) => o !== opt))

  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <Input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Add option…"
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              add()
            }
          }}
        />
        <Button type="button" variant="outline" size="sm" onClick={add}>
          Add
        </Button>
      </div>
      <div className="flex flex-wrap gap-1.5">
        {options.map((opt) => (
          <span
            key={opt}
            className="flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-700"
          >
            {opt}
            <button
              type="button"
              onClick={() => remove(opt)}
              className="ml-0.5 text-slate-400 hover:text-red-500"
            >
              ×
            </button>
          </span>
        ))}
      </div>
    </div>
  )
}

// ── Create dialog ─────────────────────────────────────────────────────────────

interface CreateFieldDialogProps {
  entityType: CustomFieldEntityType
  open: boolean
  onClose: () => void
}

function CreateFieldDialog({ entityType, open, onClose }: CreateFieldDialogProps) {
  const { mutate: create, isPending } = useCreateCustomFieldDefinition()
  const [form, setForm] = useState<CreateCustomFieldDefinitionRequest>({
    entity_type: entityType,
    name: '',
    label: '',
    field_type: 'text',
    options: [],
    required: false,
  })
  const [error, setError] = useState('')

  const needsOptions = form.field_type === 'select' || form.field_type === 'multiselect'

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.label.trim()) { setError('Label is required.'); return }
    if (needsOptions && (form.options ?? []).length === 0) {
      setError('Add at least one option.')
      return
    }
    const payload: CreateCustomFieldDefinitionRequest = {
      ...form,
      entity_type: entityType,
      name: form.label.toLowerCase().replace(/\s+/g, '_').replace(/[^a-z0-9_]/g, ''),
      options: needsOptions ? form.options : undefined,
    }
    create(payload, {
      onSuccess: () => {
        setForm({ entity_type: entityType, name: '', label: '', field_type: 'text', options: [], required: false })
        onClose()
      },
      onError: (err) => setError(getApiErrorMessage(err, 'Failed to create field.')),
    })
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add Custom Field</DialogTitle>
          <DialogDescription>
            Create a reusable field for {entityType} records. The label is used to generate the stored field key.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-1">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Label</label>
            <Input
              value={form.label}
              onChange={(e) => setForm((f) => ({ ...f, label: e.target.value }))}
              placeholder="e.g. Contract Value"
              autoFocus
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Field type</label>
            <select
              className={selectClass}
              value={form.field_type}
              onChange={(e) => setForm((f) => ({ ...f, field_type: e.target.value as CustomFieldType, options: [] }))}
            >
              {FIELD_TYPE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
          {needsOptions && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Options</label>
              <OptionsEditor
                options={form.options ?? []}
                onChange={(opts) => setForm((f) => ({ ...f, options: opts }))}
              />
            </div>
          )}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="cf-required"
              checked={form.required ?? false}
              onChange={(e) => setForm((f) => ({ ...f, required: e.target.checked }))}
              className="h-4 w-4 rounded border-slate-300 text-[#1B3A4B] focus:ring-[var(--border-focus)]"
            />
            <label htmlFor="cf-required" className="text-sm text-slate-700">Required</label>
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create Field'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Edit dialog ───────────────────────────────────────────────────────────────

interface EditFieldDialogProps {
  field: CustomFieldDefinition | null
  onClose: () => void
}

function EditFieldDialog({ field, onClose }: EditFieldDialogProps) {
  const { mutate: update, isPending } = useUpdateCustomFieldDefinition()
  const [form, setForm] = useState<UpdateCustomFieldDefinitionRequest>({})
  const [error, setError] = useState('')

  const effectiveType = (form.field_type ?? field?.field_type) as CustomFieldType | undefined
  const needsOptions = effectiveType === 'select' || effectiveType === 'multiselect'

  const handleOpen = () => {
    if (field) {
      setForm({ label: field.label, field_type: field.field_type, options: field.options ?? [], required: field.required })
      setError('')
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!field) return
    setError('')
    if (!form.label?.trim()) { setError('Label is required.'); return }
    if (needsOptions && (form.options ?? []).length === 0) {
      setError('Add at least one option.')
      return
    }
    update(
      { id: field.id, payload: { ...form, options: needsOptions ? form.options : undefined } },
      { onSuccess: onClose, onError: (err) => setError(getApiErrorMessage(err, 'Failed to update field.')) }
    )
  }

  return (
    <Dialog open={!!field} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-md" onOpenAutoFocus={handleOpen}>
        <DialogHeader>
          <DialogTitle>Edit Custom Field</DialogTitle>
          <DialogDescription>
            Update how this field appears on records. Changes affect all {field?.entity_type ?? 'selected'} entries using it.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-1">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Label</label>
            <Input
              value={form.label ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, label: e.target.value }))}
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Field type</label>
            <select
              className={selectClass}
              value={form.field_type ?? field?.field_type ?? 'text'}
              onChange={(e) =>
                setForm((f) => ({ ...f, field_type: e.target.value as CustomFieldType, options: [] }))
              }
            >
              {FIELD_TYPE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
          {needsOptions && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Options</label>
              <OptionsEditor
                options={form.options ?? []}
                onChange={(opts) => setForm((f) => ({ ...f, options: opts }))}
              />
            </div>
          )}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="cf-edit-required"
              checked={form.required ?? false}
              onChange={(e) => setForm((f) => ({ ...f, required: e.target.checked }))}
              className="h-4 w-4 rounded border-slate-300 text-[#1B3A4B] focus:ring-[var(--border-focus)]"
            />
            <label htmlFor="cf-edit-required" className="text-sm text-slate-700">Required</label>
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Saving…' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export function CustomFieldsPage() {
  const [activeTab, setActiveTab] = useState<CustomFieldEntityType>('ticket')
  const [showCreate, setShowCreate] = useState(false)
  const [editField, setEditField] = useState<CustomFieldDefinition | null>(null)

  const { data: fields = [], isLoading } = useCustomFieldDefinitions(activeTab)
  const { mutate: deleteField } = useDeleteCustomFieldDefinition()

  const handleDelete = (f: CustomFieldDefinition) => {
    if (!confirm(`Delete field "${f.label}"? All stored values will be lost.`)) return
    deleteField(f.id)
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Custom Fields</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            Define extra fields for tickets, contacts, leads, deals, accounts, quotes, and KB articles.
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" />
          Add Field
        </Button>
      </div>

      {/* Entity tabs */}
      <div className="flex gap-1 border-b border-slate-200">
        {ENTITY_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setActiveTab(tab.value)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab.value
                ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Field list */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        {isLoading ? (
          <div className="py-12 text-center text-sm text-slate-400">Loading…</div>
        ) : fields.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-14 text-slate-400">
            <SlidersHorizontal className="h-8 w-8 opacity-40" />
            <p className="text-sm font-medium">No custom fields yet</p>
            <p className="text-xs">Click "Add Field" to create one.</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 bg-slate-50 text-left">
                <th className="px-4 py-2.5 font-medium text-slate-600">Label</th>
                <th className="px-4 py-2.5 font-medium text-slate-600">Internal name</th>
                <th className="px-4 py-2.5 font-medium text-slate-600">Type</th>
                <th className="px-4 py-2.5 font-medium text-slate-600">Options</th>
                <th className="px-4 py-2.5 font-medium text-slate-600">Required</th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {fields.map((f) => (
                <tr key={f.id} className="hover:bg-slate-50/50">
                  <td className="px-4 py-3 font-medium text-slate-900">{f.label}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-500">{f.name}</td>
                  <td className="px-4 py-3">
                    <Badge variant={fieldTypeBadge[f.field_type]}>
                      {FIELD_TYPE_OPTIONS.find((o) => o.value === f.field_type)?.label ?? f.field_type}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-slate-500">
                    {f.options?.length ? f.options.join(', ') : '—'}
                  </td>
                  <td className="px-4 py-3 text-slate-500">{f.required ? 'Yes' : 'No'}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setEditField(f)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-red-600 hover:bg-red-50 hover:border-red-300"
                        onClick={() => handleDelete(f)}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                        Delete
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Dialogs */}
      <CreateFieldDialog
        key={activeTab}
        entityType={activeTab}
        open={showCreate}
        onClose={() => setShowCreate(false)}
      />
      <EditFieldDialog
        field={editField}
        onClose={() => setEditField(null)}
      />
    </div>
  )
}
