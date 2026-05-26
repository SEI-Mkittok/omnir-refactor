import { useState } from 'react'
import { Eye, MailCheck, Play, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import {
  useCreateMailConverterRule,
  useDeleteMailConverterRule,
  useMailConverterRules,
  usePreviewMailConverterRule,
  useScanMailConverterRule,
  useUpdateMailConverterRule,
} from '@/hooks/useMailConverter'
import type { MailConverterAction, MailConverterCondition, MailConverterRule } from '@/api/types'

const DEFAULT_CONDITION: MailConverterCondition = { field: 'subject', operator: 'contains', value: '' }
const DEFAULT_ACTION: MailConverterAction = { type: 'create_ticket', config: {} }

function statusBadge(status: string) {
  return status === 'active' ? <Badge variant="green">Active</Badge> : <Badge variant="gray">Inactive</Badge>
}

export function MailConverterPage() {
  const [draft, setDraft] = useState<Partial<MailConverterRule>>({
    name: '',
    status: 'inactive',
    conditions: [DEFAULT_CONDITION],
    actions: [DEFAULT_ACTION],
  })
  const [previewRule, setPreviewRule] = useState<MailConverterRule | null>(null)
  const rules = useMailConverterRules()
  const createRule = useCreateMailConverterRule()
  const updateRule = useUpdateMailConverterRule()
  const deleteRule = useDeleteMailConverterRule()
  const preview = usePreviewMailConverterRule()
  const scan = useScanMailConverterRule()
  const rows = rules.data?.data ?? []

  function updateCondition(index: number, condition: MailConverterCondition) {
    const conditions = [...(draft.conditions ?? [])]
    conditions[index] = condition
    setDraft({ ...draft, conditions })
  }

  function updateAction(index: number, action: MailConverterAction) {
    const actions = [...(draft.actions ?? [])]
    actions[index] = action
    setDraft({ ...draft, actions })
  }

  async function saveDraft() {
    await createRule.mutateAsync(draft)
    setDraft({ name: '', status: 'inactive', conditions: [DEFAULT_CONDITION], actions: [DEFAULT_ACTION] })
  }

  async function runPreview(rule: MailConverterRule) {
    setPreviewRule(rule)
    await preview.mutateAsync(rule.id)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Mail Converter</h1>
        <p className="mt-1 text-sm text-[#6B7280]">Convert synced inbox messages into CRM records without adding another mail sync path.</p>
      </div>

      <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-4 flex items-center gap-2">
          <MailCheck className="h-4 w-4 text-[#1B3A4B]" />
          <h2 className="text-base font-semibold text-slate-900">Rule builder</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-3">
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Rule name" value={draft.name ?? ''} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
          <select className="rounded-md border border-slate-200 px-3 py-2 text-sm" value={draft.status} onChange={(e) => setDraft({ ...draft, status: e.target.value as MailConverterRule['status'] })}>
            <option value="inactive">Inactive</option>
            <option value="active">Active</option>
          </select>
          <Button onClick={saveDraft} disabled={createRule.isPending || !draft.name}>Create rule</Button>
        </div>

        <div className="mt-4 space-y-2">
          {(draft.conditions ?? []).map((condition, index) => (
            <div key={index} className="grid gap-2 rounded-md bg-slate-50 p-2 md:grid-cols-[1fr_1fr_1.5fr_auto]">
              <select className="rounded border border-slate-200 px-2 py-1.5 text-sm" value={condition.field} onChange={(e) => updateCondition(index, { ...condition, field: e.target.value })}>
                <option value="from">From</option>
                <option value="to">To</option>
                <option value="subject">Subject</option>
                <option value="body">Body</option>
              </select>
              <select className="rounded border border-slate-200 px-2 py-1.5 text-sm" value={condition.operator} onChange={(e) => updateCondition(index, { ...condition, operator: e.target.value as MailConverterCondition['operator'] })}>
                <option value="contains">contains</option>
                <option value="not_contains">does not contain</option>
                <option value="equals">equals</option>
                <option value="starts_with">starts with</option>
                <option value="ends_with">ends with</option>
                <option value="is_set">is set</option>
              </select>
              <input className="rounded border border-slate-200 px-2 py-1.5 text-sm" placeholder="Match text" value={condition.value ?? ''} onChange={(e) => updateCondition(index, { ...condition, value: e.target.value })} />
              <button className="rounded p-1 text-slate-400 hover:text-red-500" onClick={() => setDraft({ ...draft, conditions: (draft.conditions ?? []).filter((_, i) => i !== index) })}><Trash2 className="h-4 w-4" /></button>
            </div>
          ))}
          <Button variant="outline" size="sm" onClick={() => setDraft({ ...draft, conditions: [...(draft.conditions ?? []), DEFAULT_CONDITION] })}>
            <Plus className="h-3.5 w-3.5" /> Condition
          </Button>
        </div>

        <div className="mt-4 space-y-2">
          {(draft.actions ?? []).map((action, index) => (
            <div key={index} className="grid gap-2 rounded-md bg-slate-50 p-2 md:grid-cols-[1fr_1fr_auto]">
              <select className="rounded border border-slate-200 px-2 py-1.5 text-sm" value={action.type} onChange={(e) => updateAction(index, { ...action, type: e.target.value as MailConverterAction['type'] })}>
                <option value="create_ticket">Create ticket</option>
                <option value="create_lead">Create lead</option>
                <option value="create_contact">Create contact</option>
                <option value="update_contact">Update contact</option>
                <option value="create_activity">Create activity</option>
              </select>
              <input className="rounded border border-slate-200 px-2 py-1.5 text-sm" placeholder="Owner ID or subject override" onChange={(e) => updateAction(index, { ...action, config: { ...action.config, owner_id: e.target.value } })} />
              <button className="rounded p-1 text-slate-400 hover:text-red-500" onClick={() => setDraft({ ...draft, actions: (draft.actions ?? []).filter((_, i) => i !== index) })}><Trash2 className="h-4 w-4" /></button>
            </div>
          ))}
          <Button variant="outline" size="sm" onClick={() => setDraft({ ...draft, actions: [...(draft.actions ?? []), DEFAULT_ACTION] })}>
            <Plus className="h-3.5 w-3.5" /> Action
          </Button>
        </div>
      </section>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-5 py-4">
          <h2 className="text-base font-semibold text-slate-900">Rules</h2>
        </div>
        {rows.length === 0 ? (
          <div className="px-5 py-10 text-sm text-slate-500">No mail converter rules yet.</div>
        ) : (
          rows.map((rule) => (
            <div key={rule.id} className="flex flex-col gap-3 border-b border-slate-100 px-5 py-4 md:flex-row md:items-center md:justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <p className="font-medium text-slate-900">{rule.name}</p>
                  {statusBadge(rule.status)}
                </div>
                <p className="mt-1 text-xs text-slate-500">{rule.conditions.length} conditions · {rule.actions.length} actions · last run {rule.last_run_at ? new Date(rule.last_run_at).toLocaleString() : 'never'}</p>
              </div>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" size="sm" onClick={() => runPreview(rule)}><Eye className="h-3.5 w-3.5" /> Preview</Button>
                <Button variant="outline" size="sm" onClick={() => scan.mutate(rule.id)} disabled={scan.isPending || rule.status !== 'active'}><Play className="h-3.5 w-3.5" /> Scan</Button>
                <Button variant="outline" size="sm" onClick={() => updateRule.mutate({ id: rule.id, payload: { status: rule.status === 'active' ? 'inactive' : 'active' } })}>{rule.status === 'active' ? 'Disable' : 'Enable'}</Button>
                <Button variant="ghost" size="sm" onClick={() => deleteRule.mutate(rule.id)}><Trash2 className="h-3.5 w-3.5" /></Button>
              </div>
            </div>
          ))
        )}
      </section>

      {(previewRule || preview.data || scan.data) && (
        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          {previewRule && <h2 className="text-base font-semibold text-slate-900">Preview: {previewRule.name}</h2>}
          {preview.data && (
            <div className="mt-4 divide-y divide-slate-100">
              <p className="pb-3 text-sm text-slate-500">{preview.data.total} recent matches</p>
              {preview.data.matches.slice(0, 8).map((msg) => (
                <div key={msg.id} className="py-3 text-sm">
                  <p className="font-medium text-slate-900">{msg.subject || '(no subject)'}</p>
                  <p className="text-xs text-slate-500">{msg.from_addr} · {new Date(msg.sent_at).toLocaleString()}</p>
                </div>
              ))}
            </div>
          )}
          {scan.data && (
            <div className="mt-4 rounded-md bg-slate-50 p-3 text-sm text-slate-700">
              Scan {scan.data.status}: {scan.data.processed_count} processed, {scan.data.skipped_count} skipped.
            </div>
          )}
        </section>
      )}
    </div>
  )
}
