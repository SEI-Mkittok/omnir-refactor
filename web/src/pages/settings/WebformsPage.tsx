import { useMemo, useState } from 'react'
import { Eye, Plus, Trash2, Webhook } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import {
  useCreateWebform,
  useDeleteWebform,
  usePreviewWebform,
  useUpdateWebform,
  useWebforms,
} from '@/hooks/useWebforms'
import type { Webform, WebformField, WebformTargetModule } from '@/api/types'

const DEFAULT_FIELDS: WebformField[] = [
  { key: 'first_name', label: 'First name', type: 'text', required: true, target_field: 'first_name' },
  { key: 'last_name', label: 'Last name', type: 'text', required: true, target_field: 'last_name' },
  { key: 'email', label: 'Email', type: 'email', required: true, target_field: 'email' },
]

function statusBadge(status: string) {
  return status === 'active' ? <Badge variant="green">Active</Badge> : <Badge variant="gray">Inactive</Badge>
}

export function WebformsPage() {
  const [draft, setDraft] = useState<Partial<Webform>>({
    name: '',
    status: 'inactive',
    target_module: 'lead',
    success_message: 'Thanks. Your submission has been received.',
    spam_trap_field: 'website',
    fields: DEFAULT_FIELDS,
  })
  const [previewForm, setPreviewForm] = useState<Webform | null>(null)
  const [previewPayload, setPreviewPayload] = useState('{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}')
  const forms = useWebforms()
  const createForm = useCreateWebform()
  const updateForm = useUpdateWebform()
  const deleteForm = useDeleteWebform()
  const preview = usePreviewWebform()
  const rows = forms.data?.data ?? []
  const activeCount = useMemo(() => rows.filter((f) => f.status === 'active').length, [rows])

  function addField() {
    const fields = draft.fields ?? []
    setDraft({
      ...draft,
      fields: [...fields, { key: '', label: '', type: 'text', required: false, target_field: '' }],
    })
  }

  function updateField(index: number, field: WebformField) {
    const fields = [...(draft.fields ?? [])]
    fields[index] = field
    setDraft({ ...draft, fields })
  }

  async function saveDraft() {
    await createForm.mutateAsync(draft)
    setDraft({
      name: '',
      status: 'inactive',
      target_module: 'lead',
      success_message: 'Thanks. Your submission has been received.',
      spam_trap_field: 'website',
      fields: DEFAULT_FIELDS,
    })
  }

  async function runPreview() {
    if (!previewForm) return
    let payload: Record<string, unknown> = {}
    try {
      payload = JSON.parse(previewPayload)
    } catch {
      payload = {}
    }
    await preview.mutateAsync({ id: previewForm.id, payload })
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Webforms</h1>
          <p className="mt-1 text-sm text-[#6B7280]">Build public intake forms that create leads, contacts, or tickets.</p>
        </div>
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div className="rounded-md border border-slate-200 bg-white px-3 py-2">
            <span className="block text-xs text-slate-500">Forms</span>
            <span className="font-semibold text-slate-900">{rows.length}</span>
          </div>
          <div className="rounded-md border border-slate-200 bg-white px-3 py-2">
            <span className="block text-xs text-slate-500">Active</span>
            <span className="font-semibold text-slate-900">{activeCount}</span>
          </div>
        </div>
      </div>

      <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-4 flex items-center gap-2">
          <Webhook className="h-4 w-4 text-[#1B3A4B]" />
          <h2 className="text-base font-semibold text-slate-900">Builder</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-4">
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Form name" value={draft.name ?? ''} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
          <select className="rounded-md border border-slate-200 px-3 py-2 text-sm" value={draft.target_module} onChange={(e) => setDraft({ ...draft, target_module: e.target.value as WebformTargetModule })}>
            <option value="lead">Lead</option>
            <option value="contact">Contact</option>
            <option value="ticket">Ticket</option>
          </select>
          <select className="rounded-md border border-slate-200 px-3 py-2 text-sm" value={draft.status} onChange={(e) => setDraft({ ...draft, status: e.target.value as Webform['status'] })}>
            <option value="inactive">Inactive</option>
            <option value="active">Active</option>
          </select>
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Return URL" value={draft.return_url ?? ''} onChange={(e) => setDraft({ ...draft, return_url: e.target.value })} />
        </div>
        <div className="mt-3 grid gap-3 md:grid-cols-2">
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Success message" value={draft.success_message ?? ''} onChange={(e) => setDraft({ ...draft, success_message: e.target.value })} />
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Spam trap field" value={draft.spam_trap_field ?? ''} onChange={(e) => setDraft({ ...draft, spam_trap_field: e.target.value })} />
        </div>
        <div className="mt-4 space-y-2">
          {(draft.fields ?? []).map((field, index) => (
            <div key={index} className="grid gap-2 rounded-md bg-slate-50 p-2 md:grid-cols-[1fr_1fr_0.8fr_1fr_auto_auto]">
              <input className="rounded border border-slate-200 px-2 py-1.5 text-sm" placeholder="key" value={field.key} onChange={(e) => updateField(index, { ...field, key: e.target.value })} />
              <input className="rounded border border-slate-200 px-2 py-1.5 text-sm" placeholder="label" value={field.label} onChange={(e) => updateField(index, { ...field, label: e.target.value })} />
              <select className="rounded border border-slate-200 px-2 py-1.5 text-sm" value={field.type} onChange={(e) => updateField(index, { ...field, type: e.target.value })}>
                <option value="text">Text</option>
                <option value="email">Email</option>
                <option value="textarea">Textarea</option>
                <option value="select">Select</option>
              </select>
              <input className="rounded border border-slate-200 px-2 py-1.5 text-sm" placeholder="target field" value={field.target_field ?? ''} onChange={(e) => updateField(index, { ...field, target_field: e.target.value })} />
              <label className="flex items-center gap-2 text-sm text-slate-600">
                <input type="checkbox" checked={Boolean(field.required)} onChange={(e) => updateField(index, { ...field, required: e.target.checked })} />
                Required
              </label>
              <button className="rounded p-1 text-slate-400 hover:text-red-500" onClick={() => setDraft({ ...draft, fields: (draft.fields ?? []).filter((_, i) => i !== index) })}>
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
        <div className="mt-4 flex gap-2">
          <Button variant="outline" onClick={addField}><Plus className="h-4 w-4" /> Field</Button>
          <Button onClick={saveDraft} disabled={createForm.isPending || !draft.name}>Create form</Button>
        </div>
      </section>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-5 py-4">
          <h2 className="text-base font-semibold text-slate-900">Forms</h2>
        </div>
        {rows.length === 0 ? (
          <div className="px-5 py-10 text-sm text-slate-500">No webforms yet.</div>
        ) : (
          rows.map((form) => (
            <div key={form.id} className="flex flex-col gap-3 border-b border-slate-100 px-5 py-4 md:flex-row md:items-center md:justify-between">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <p className="font-medium text-slate-900">{form.name}</p>
                  {statusBadge(form.status)}
                  <Badge variant="blue">{form.target_module}</Badge>
                </div>
                <p className="mt-1 truncate text-xs text-slate-500">/api/webforms/{form.public_id}</p>
              </div>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" size="sm" onClick={() => setPreviewForm(form)}><Eye className="h-3.5 w-3.5" /> Preview</Button>
                <Button variant="outline" size="sm" onClick={() => updateForm.mutate({ id: form.id, payload: { status: form.status === 'active' ? 'inactive' : 'active' } })}>
                  {form.status === 'active' ? 'Disable' : 'Enable'}
                </Button>
                <Button variant="ghost" size="sm" onClick={() => deleteForm.mutate(form.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
              </div>
            </div>
          ))
        )}
      </section>

      {previewForm && (
        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="text-base font-semibold text-slate-900">Preview mapping: {previewForm.name}</h2>
          <textarea className="mt-3 h-28 w-full rounded-md border border-slate-200 px-3 py-2 text-sm font-mono" value={previewPayload} onChange={(e) => setPreviewPayload(e.target.value)} />
          <div className="mt-3 flex gap-2">
            <Button onClick={runPreview} disabled={preview.isPending}>Preview</Button>
            <Button variant="ghost" onClick={() => setPreviewForm(null)}>Close</Button>
          </div>
          {preview.data && (
            <pre className="mt-4 overflow-auto rounded-md bg-slate-950 p-3 text-xs text-slate-100">{JSON.stringify(preview.data, null, 2)}</pre>
          )}
        </section>
      )}
    </div>
  )
}
