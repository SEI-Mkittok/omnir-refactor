import { useState, useRef, useEffect } from 'react'
import {
  Plus,
  Zap,
  Play,
  Pause,
  Trash2,
  ChevronRight,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  X,
  Info,
  Braces,
  Search,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import {
  useAutomations,
  useAutomationRuns,
  useCreateAutomation,
  useUpdateAutomation,
  useDeleteAutomation,
  useExecuteAutomation,
} from '@/hooks/useAutomations'
import type {
  Automation,
  AutomationStatus,
  TriggerType,
  ConditionOperator,
  ActionType,
  AutomationTrigger,
  AutomationCondition,
  AutomationAction,
  CreateAutomationRequest,
} from '@/api/types'

// ---- Constants ----

const TRIGGER_LABELS: Record<TriggerType, string> = {
  contact_created: 'Contact Created',
  contact_updated: 'Contact Updated',
  deal_created: 'Deal Created',
  deal_stage_changed: 'Deal Stage Changed',
  ticket_created: 'Ticket Created',
  activity_overdue: 'Activity Overdue',
  manual: 'Manual (Run via API)',
}

const OPERATOR_LABELS: Record<ConditionOperator, string> = {
  equals: 'equals',
  not_equals: 'does not equal',
  contains: 'contains',
  not_contains: 'does not contain',
  greater_than: 'is greater than',
  less_than: 'is less than',
  is_set: 'is set',
  is_not_set: 'is not set',
}

const ACTION_LABELS: Record<ActionType, string> = {
  assign_owner: 'Assign Owner',
  send_email: 'Send Email',
  enroll_in_sequence: 'Enroll in Sequence',
  create_activity: 'Create Activity',
  webhook: 'Call Webhook',
}

const VALUE_LESS_OPERATORS: ConditionOperator[] = ['is_set', 'is_not_set']

// ---- Status badge ----

function statusBadge(status: AutomationStatus) {
  if (status === 'active') return <Badge variant="green">Active</Badge>
  if (status === 'paused') return <Badge variant="yellow">Paused</Badge>
  return <Badge variant="gray">Draft</Badge>
}

function runStatusIcon(status: string) {
  if (status === 'succeeded') return <CheckCircle className="h-4 w-4 text-green-500" />
  if (status === 'failed') return <XCircle className="h-4 w-4 text-red-500" />
  if (status === 'running') return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />
  return <Clock className="h-4 w-4 text-slate-400" />
}

// ---- Trigger Picker ----

function TriggerPicker({
  value,
  onChange,
}: {
  value: AutomationTrigger
  onChange: (t: AutomationTrigger) => void
}) {
  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-slate-700">Trigger</label>
      <select
        className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        value={value.type}
        onChange={(e) => onChange({ type: e.target.value as TriggerType, config: {} })}
      >
        {(Object.keys(TRIGGER_LABELS) as TriggerType[]).map((t) => (
          <option key={t} value={t}>
            {TRIGGER_LABELS[t]}
          </option>
        ))}
      </select>
    </div>
  )
}

// ---- Condition Editor ----

const COMMON_FIELDS = [
  { value: 'name', label: 'Name' },
  { value: 'email', label: 'Email' },
  { value: 'company', label: 'Company' },
  { value: 'stage', label: 'Stage' },
  { value: 'amount', label: 'Amount' },
  { value: 'status', label: 'Status' },
  { value: 'priority', label: 'Priority' },
]

function ConditionRow({
  condition,
  onChange,
  onRemove,
}: {
  condition: AutomationCondition
  onChange: (c: AutomationCondition) => void
  onRemove: () => void
}) {
  const showValue = !VALUE_LESS_OPERATORS.includes(condition.operator)
  return (
    <div className="flex items-center gap-2 rounded-md border border-slate-200 bg-slate-50 p-2">
      <select
        className="flex-1 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        value={condition.field}
        onChange={(e) => onChange({ ...condition, field: e.target.value })}
      >
        {COMMON_FIELDS.map((f) => (
          <option key={f.value} value={f.value}>
            {f.label}
          </option>
        ))}
      </select>
      <select
        className="flex-1 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        value={condition.operator}
        onChange={(e) => onChange({ ...condition, operator: e.target.value as ConditionOperator })}
      >
        {(Object.keys(OPERATOR_LABELS) as ConditionOperator[]).map((op) => (
          <option key={op} value={op}>
            {OPERATOR_LABELS[op]}
          </option>
        ))}
      </select>
      {showValue && (
        <input
          type="text"
          className="flex-1 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="Value"
          value={(condition.value as string) ?? ''}
          onChange={(e) => onChange({ ...condition, value: e.target.value })}
        />
      )}
      <button
        type="button"
        onClick={onRemove}
        className="rounded p-1 text-slate-400 hover:text-red-500"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}

function ConditionEditor({
  conditions,
  onChange,
}: {
  conditions: AutomationCondition[]
  onChange: (c: AutomationCondition[]) => void
}) {
  function addCondition() {
    onChange([...conditions, { field: 'name', operator: 'equals', value: '' }])
  }

  function updateCondition(idx: number, c: AutomationCondition) {
    const next = [...conditions]
    next[idx] = c
    onChange(next)
  }

  function removeCondition(idx: number) {
    onChange(conditions.filter((_, i) => i !== idx))
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-slate-700">Conditions (all must match)</label>
        <Button type="button" variant="outline" size="sm" onClick={addCondition}>
          <Plus className="mr-1 h-3 w-3" /> Add condition
        </Button>
      </div>
      {conditions.length === 0 && (
        <p className="flex items-center gap-1 text-xs text-slate-500">
          <Info className="h-3 w-3" /> No conditions — automation runs on every trigger.
        </p>
      )}
      {conditions.map((c, i) => (
        <ConditionRow
          key={i}
          condition={c}
          onChange={(updated) => updateCondition(i, updated)}
          onRemove={() => removeCondition(i)}
        />
      ))}
    </div>
  )
}

// ---- Variable Picker ----

const AUTOMATION_VARIABLES: { module: string; key: string; fields: { key: string; label: string }[] }[] = [
  {
    module: 'Contact',
    key: 'contact',
    fields: [
      { key: 'email', label: 'Email' },
      { key: 'first_name', label: 'First Name' },
      { key: 'last_name', label: 'Last Name' },
      { key: 'name', label: 'Full Name' },
      { key: 'phone', label: 'Phone' },
      { key: 'company', label: 'Company' },
    ],
  },
  {
    module: 'Deal',
    key: 'deal',
    fields: [
      { key: 'name', label: 'Name' },
      { key: 'value', label: 'Value' },
      { key: 'stage', label: 'Stage' },
      { key: 'owner', label: 'Owner' },
      { key: 'close_date', label: 'Close Date' },
    ],
  },
  {
    module: 'Ticket',
    key: 'ticket',
    fields: [
      { key: 'number', label: 'Number' },
      { key: 'title', label: 'Title' },
      { key: 'status', label: 'Status' },
      { key: 'priority', label: 'Priority' },
    ],
  },
  {
    module: 'Account',
    key: 'account',
    fields: [
      { key: 'name', label: 'Name' },
      { key: 'industry', label: 'Industry' },
      { key: 'website', label: 'Website' },
      { key: 'phone', label: 'Phone' },
    ],
  },
  {
    module: 'Activity',
    key: 'activity',
    fields: [
      { key: 'type', label: 'Type' },
      { key: 'description', label: 'Description' },
      { key: 'date', label: 'Date' },
    ],
  },
]

interface VariablePickerProps {
  onInsert: (variable: string) => void
}

function VariablePicker({ onInsert }: VariablePickerProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    searchRef.current?.focus()
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  const q = search.toLowerCase()
  const filtered = AUTOMATION_VARIABLES
    .map((group) => ({
      ...group,
      fields: group.fields.filter(
        (f) =>
          !q ||
          f.label.toLowerCase().includes(q) ||
          f.key.includes(q) ||
          group.module.toLowerCase().includes(q)
      ),
    }))
    .filter((g) => g.fields.length > 0)

  function handleSelect(moduleKey: string, fieldKey: string) {
    onInsert(`{{${moduleKey}.${fieldKey}}}`)
    setOpen(false)
    setSearch('')
  }

  return (
    <div ref={containerRef} className="relative shrink-0">
      <button
        type="button"
        title="Insert variable"
        onClick={() => setOpen((v) => !v)}
        className="flex h-7 w-7 items-center justify-center rounded border border-slate-300 bg-white text-slate-500 hover:bg-slate-50 hover:text-[var(--color-primary)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
      >
        <Braces className="h-3.5 w-3.5" />
      </button>

      {open && (
        <div className="absolute right-0 top-full z-50 mt-1 w-56 rounded-lg border border-slate-200 bg-white shadow-lg">
          {/* Search */}
          <div className="flex items-center gap-1.5 border-b border-slate-100 px-2 py-1.5">
            <Search className="h-3.5 w-3.5 shrink-0 text-slate-400" />
            <input
              ref={searchRef}
              type="text"
              placeholder="Search variables…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-transparent text-xs text-slate-700 placeholder-slate-400 focus:outline-none"
            />
          </div>

          {/* Variable groups */}
          <div className="max-h-60 overflow-y-auto py-1">
            {filtered.length === 0 && (
              <p className="px-3 py-2 text-xs text-slate-400">No variables match.</p>
            )}
            {filtered.map((group) => (
              <div key={group.key}>
                <p className="px-3 pb-0.5 pt-2 text-[10px] font-semibold uppercase tracking-wider text-slate-400">
                  {group.module}
                </p>
                {group.fields.map((field) => (
                  <button
                    key={field.key}
                    type="button"
                    onMouseDown={(e) => {
                      e.preventDefault() // keep focus in target input
                      handleSelect(group.key, field.key)
                    }}
                    className="flex w-full items-center justify-between px-3 py-1.5 text-left text-xs text-slate-700 hover:bg-slate-50"
                  >
                    <span>{field.label}</span>
                    <span className="ml-2 rounded bg-slate-100 px-1 font-mono text-[10px] text-slate-500">
                      {`{{${group.key}.${field.key}}}`}
                    </span>
                  </button>
                ))}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// VariableInput / VariableTextarea — text fields with an inline variable picker

interface VariableInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  value: string
  onValueChange: (v: string) => void
}

function VariableInput({ value, onValueChange, className, ...rest }: VariableInputProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  function handleInsert(variable: string) {
    const el = inputRef.current
    const start = el?.selectionStart ?? value.length
    const end = el?.selectionEnd ?? value.length
    const next = value.slice(0, start) + variable + value.slice(end)
    onValueChange(next)
    requestAnimationFrame(() => {
      el?.focus()
      el?.setSelectionRange(start + variable.length, start + variable.length)
    })
  }

  return (
    <div className="flex items-center gap-1">
      <input
        ref={inputRef}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        className={className}
        {...rest}
      />
      <VariablePicker onInsert={handleInsert} />
    </div>
  )
}

interface VariableTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  value: string
  onValueChange: (v: string) => void
}

function VariableTextarea({ value, onValueChange, className, ...rest }: VariableTextareaProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  function handleInsert(variable: string) {
    const el = textareaRef.current
    const start = el?.selectionStart ?? value.length
    const end = el?.selectionEnd ?? value.length
    const next = value.slice(0, start) + variable + value.slice(end)
    onValueChange(next)
    requestAnimationFrame(() => {
      el?.focus()
      el?.setSelectionRange(start + variable.length, start + variable.length)
    })
  }

  return (
    <div className="flex items-start gap-1">
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        className={className}
        {...rest}
      />
      <VariablePicker onInsert={handleInsert} />
    </div>
  )
}

// ---- Action Configurator ----

function ActionConfigFields({
  action,
  onChange,
}: {
  action: AutomationAction
  onChange: (a: AutomationAction) => void
}) {
  function setConfig(key: string, val: string) {
    onChange({ ...action, config: { ...action.config, [key]: val } })
  }

  const cfg = action.config as Record<string, string>

  if (action.type === 'send_email') {
    const inputCls = 'flex-1 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]'
    return (
      <div className="space-y-2 pl-4">
        <VariableInput
          type="text"
          className={inputCls}
          placeholder="To (e.g. {{contact.email}})"
          value={cfg.to ?? ''}
          onValueChange={(v) => setConfig('to', v)}
        />
        <VariableInput
          type="text"
          className={inputCls}
          placeholder="Subject"
          value={cfg.subject ?? ''}
          onValueChange={(v) => setConfig('subject', v)}
        />
        <VariableTextarea
          className={inputCls}
          placeholder="Body"
          rows={3}
          value={cfg.body ?? ''}
          onValueChange={(v) => setConfig('body', v)}
        />
      </div>
    )
  }

  if (action.type === 'assign_owner') {
    return (
      <div className="pl-4">
        <input
          type="text"
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="Owner user ID"
          value={cfg.owner_id ?? ''}
          onChange={(e) => setConfig('owner_id', e.target.value)}
        />
      </div>
    )
  }

  if (action.type === 'enroll_in_sequence') {
    return (
      <div className="pl-4">
        <input
          type="text"
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="Sequence ID"
          value={cfg.sequence_id ?? ''}
          onChange={(e) => setConfig('sequence_id', e.target.value)}
        />
      </div>
    )
  }

  if (action.type === 'create_activity') {
    return (
      <div className="space-y-2 pl-4">
        <input
          type="text"
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="Activity subject"
          value={cfg.subject ?? cfg.title ?? ''}
          onChange={(e) => setConfig('subject', e.target.value)}
        />
        <input
          type="number"
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="Due in (days)"
          value={cfg.due_in_days ?? ''}
          onChange={(e) => setConfig('due_in_days', e.target.value)}
        />
      </div>
    )
  }

  if (action.type === 'webhook') {
    return (
      <div className="space-y-2 pl-4">
        <input
          type="url"
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          placeholder="URL"
          value={cfg.url ?? ''}
          onChange={(e) => setConfig('url', e.target.value)}
        />
        <select
          className="w-full rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          value={cfg.method ?? 'POST'}
          onChange={(e) => setConfig('method', e.target.value)}
        >
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="GET">GET</option>
        </select>
      </div>
    )
  }

  return null
}

function ActionRow({
  action,
  idx,
  onChange,
  onRemove,
}: {
  action: AutomationAction
  idx: number
  onChange: (a: AutomationAction) => void
  onRemove: () => void
}) {
  return (
    <div className="rounded-md border border-slate-200 bg-slate-50 p-3 space-y-2">
      <div className="flex items-center gap-2">
        <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-xs font-bold text-[var(--color-primary)]">
          {idx + 1}
        </span>
        <select
          className="flex-1 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
          value={action.type}
          onChange={(e) =>
            onChange({ type: e.target.value as ActionType, config: {} })
          }
        >
          {(Object.keys(ACTION_LABELS) as ActionType[]).map((t) => (
            <option key={t} value={t}>
              {ACTION_LABELS[t]}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={onRemove}
          className="rounded p-1 text-slate-400 hover:text-red-500"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
      <ActionConfigFields action={action} onChange={onChange} />
    </div>
  )
}

function ActionConfigurator({
  actions,
  onChange,
}: {
  actions: AutomationAction[]
  onChange: (a: AutomationAction[]) => void
}) {
  function addAction() {
    onChange([...actions, { type: 'send_email', config: {} }])
  }

  function updateAction(idx: number, a: AutomationAction) {
    const next = [...actions]
    next[idx] = a
    onChange(next)
  }

  function removeAction(idx: number) {
    onChange(actions.filter((_, i) => i !== idx))
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-slate-700">Actions</label>
        <Button type="button" variant="outline" size="sm" onClick={addAction}>
          <Plus className="mr-1 h-3 w-3" /> Add action
        </Button>
      </div>
      {actions.length === 0 && (
        <p className="text-xs text-red-500">At least one action is required.</p>
      )}
      {actions.map((a, i) => (
        <ActionRow
          key={i}
          action={a}
          idx={i}
          onChange={(updated) => updateAction(i, updated)}
          onRemove={() => removeAction(i)}
        />
      ))}
    </div>
  )
}

// ---- Builder Modal ----

interface BuilderState {
  name: string
  description: string
  trigger: AutomationTrigger
  conditions: AutomationCondition[]
  actions: AutomationAction[]
}

function defaultBuilderState(): BuilderState {
  return {
    name: '',
    description: '',
    trigger: { type: 'contact_created', config: {} },
    conditions: [],
    actions: [{ type: 'send_email', config: {} }],
  }
}

function fromAutomation(a: Automation): BuilderState {
  return {
    name: a.name,
    description: a.description,
    trigger: a.trigger,
    conditions: a.conditions,
    actions: a.actions,
  }
}

interface BuilderModalProps {
  initial?: Automation | null
  onClose: () => void
  onSave: (s: BuilderState) => void
  saving: boolean
}

function BuilderModal({ initial, onClose, onSave, saving }: BuilderModalProps) {
  const [state, setState] = useState<BuilderState>(
    initial ? fromAutomation(initial) : defaultBuilderState()
  )

  function update<K extends keyof BuilderState>(key: K, val: BuilderState[K]) {
    setState((s) => ({ ...s, [key]: val }))
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onSave(state)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="relative flex max-h-[90vh] w-full max-w-2xl flex-col rounded-xl bg-white shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b px-6 py-4">
          <div className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-[var(--color-primary)]" />
            <h2 className="text-lg font-semibold text-slate-900">
              {initial ? 'Edit Automation' : 'New Automation'}
            </h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 text-slate-400 hover:text-slate-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Body */}
        <form id="automation-form" onSubmit={handleSubmit} className="flex-1 overflow-y-auto px-6 py-4 space-y-5">
          {/* Name + Description */}
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-slate-700">Name *</label>
              <input
                type="text"
                required
                className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                placeholder="e.g. Welcome new contacts"
                value={state.name}
                onChange={(e) => update('name', e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700">Description</label>
              <input
                type="text"
                className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                placeholder="What does this automation do?"
                value={state.description}
                onChange={(e) => update('description', e.target.value)}
              />
            </div>
          </div>

          <hr className="border-slate-200" />

          {/* Trigger */}
          <TriggerPicker value={state.trigger} onChange={(t) => update('trigger', t)} />

          <hr className="border-slate-200" />

          {/* Conditions */}
          <ConditionEditor
            conditions={state.conditions}
            onChange={(c) => update('conditions', c)}
          />

          <hr className="border-slate-200" />

          {/* Actions */}
          <ActionConfigurator actions={state.actions} onChange={(a) => update('actions', a)} />
        </form>

        {/* Footer */}
        <div className="flex justify-end gap-2 border-t px-6 py-4">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" form="automation-form" disabled={saving}>
            {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {initial ? 'Save changes' : 'Create automation'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// ---- Execution History Panel ----

function ExecutionHistoryPanel({ automationId }: { automationId: string }) {
  const { data, isLoading } = useAutomationRuns(automationId)
  const runs = data?.data ?? []

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
      </div>
    )
  }

  if (runs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-slate-500">
        <Clock className="mb-2 h-8 w-8 text-slate-300" />
        <p className="text-sm">No executions yet</p>
      </div>
    )
  }

  return (
    <div className="divide-y divide-slate-100">
      {runs.map((run) => (
        <div key={run.id} className="flex items-center gap-3 px-4 py-3 text-sm">
          {runStatusIcon(run.status)}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium capitalize text-slate-700">{run.status}</span>
              {run.entity_type && (
                <span className="text-xs text-slate-500">· {run.entity_type}</span>
              )}
            </div>
            {run.error_message && (
              <p className="mt-0.5 truncate text-xs text-red-500">{run.error_message}</p>
            )}
          </div>
          <span className="shrink-0 text-xs text-slate-400">
            {new Date(run.created_at).toLocaleString()}
          </span>
        </div>
      ))}
    </div>
  )
}

// ---- Detail Panel ----

function AutomationDetailPanel({
  automation,
  onClose,
  onEdit,
}: {
  automation: Automation
  onClose: () => void
  onEdit: () => void
}) {
  const update = useUpdateAutomation()
  const execute = useExecuteAutomation()

  function toggleStatus() {
    const next: Record<string, 'active' | 'paused'> = { active: 'paused', paused: 'active', draft: 'active' }
    const newStatus = next[automation.status] ?? 'active'
    update.mutate({ id: automation.id, payload: { status: newStatus } })
  }

  function runManual() {
    execute.mutate({ id: automation.id, payload: { entity_type: 'manual', data: {} } })
  }

  return (
    <div className="fixed inset-y-0 right-0 z-40 flex w-full flex-col bg-white shadow-xl sm:w-[480px]">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-6 py-4">
        <div className="flex items-center gap-2 min-w-0">
          <Zap className="h-5 w-5 shrink-0 text-[var(--color-primary)]" />
          <h2 className="truncate text-base font-semibold text-slate-900">{automation.name}</h2>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="ml-2 rounded-md p-1 text-slate-400 hover:text-slate-600"
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {/* Status + actions */}
        <div className="flex items-center gap-3 px-6 py-4 border-b">
          {statusBadge(automation.status)}
          <span className="text-sm text-slate-500">{automation.run_count} runs total</span>
          <div className="ml-auto flex gap-2">
            <Button size="sm" variant="outline" onClick={runManual} disabled={execute.isPending}>
              <Play className="mr-1 h-3 w-3" /> Run
            </Button>
            <Button size="sm" variant="outline" onClick={toggleStatus} disabled={update.isPending}>
              {automation.status === 'active' ? (
                <>
                  <Pause className="mr-1 h-3 w-3" /> Pause
                </>
              ) : (
                <>
                  <Play className="mr-1 h-3 w-3" /> Activate
                </>
              )}
            </Button>
            <Button size="sm" variant="outline" onClick={onEdit}>
              Edit
            </Button>
          </div>
        </div>

        {/* Trigger */}
        <div className="px-6 py-4 border-b">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Trigger</p>
          <p className="mt-1 text-sm text-slate-800">
            {TRIGGER_LABELS[automation.trigger.type] ?? automation.trigger.type}
          </p>
        </div>

        {/* Conditions */}
        <div className="px-6 py-4 border-b">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Conditions</p>
          {automation.conditions.length === 0 ? (
            <p className="mt-1 text-sm text-slate-400">None — runs on every trigger</p>
          ) : (
            <ul className="mt-2 space-y-1">
              {automation.conditions.map((c, i) => (
                <li key={i} className="text-sm text-slate-700">
                  <span className="font-medium">{c.field}</span>{' '}
                  <span className="text-slate-500">{OPERATOR_LABELS[c.operator]}</span>{' '}
                  {c.value !== undefined && c.value !== '' && (
                    <span className="font-medium">"{String(c.value)}"</span>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Actions */}
        <div className="px-6 py-4 border-b">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Actions</p>
          <ol className="mt-2 space-y-2">
            {automation.actions.map((a, i) => (
              <li key={i} className="flex items-center gap-2 text-sm text-slate-700">
                <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-xs font-bold text-[var(--color-primary)]">
                  {i + 1}
                </span>
                {ACTION_LABELS[a.type] ?? a.type}
              </li>
            ))}
          </ol>
        </div>

        {/* Execution History */}
        <div>
          <div className="px-6 py-3 border-b">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Execution History
            </p>
          </div>
          <ExecutionHistoryPanel automationId={automation.id} />
        </div>
      </div>
    </div>
  )
}

// ---- Main Page ----

export function AutomationsPage() {
  const [statusFilter, setStatusFilter] = useState<AutomationStatus | ''>('')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [showBuilder, setShowBuilder] = useState(false)
  const [editingAutomation, setEditingAutomation] = useState<Automation | null>(null)

  const { data, isLoading } = useAutomations(
    statusFilter ? { status: statusFilter } : undefined
  )
  const automations = data?.data ?? []

  const createMutation = useCreateAutomation()
  const updateMutation = useUpdateAutomation()
  const deleteMutation = useDeleteAutomation()

  const selectedAutomation = automations.find((a) => a.id === selectedId) ?? null

  function openCreate() {
    setEditingAutomation(null)
    setShowBuilder(true)
  }

  function openEdit(a: Automation) {
    setEditingAutomation(a)
    setShowBuilder(true)
  }

  async function handleSave(state: BuilderState) {
    if (editingAutomation) {
      await updateMutation.mutateAsync({
        id: editingAutomation.id,
        payload: {
          name: state.name,
          description: state.description,
          trigger: state.trigger,
          conditions: state.conditions,
          actions: state.actions,
        },
      })
    } else {
      const req: CreateAutomationRequest = {
        name: state.name,
        description: state.description,
        trigger: state.trigger,
        conditions: state.conditions,
        actions: state.actions,
      }
      await createMutation.mutateAsync(req)
    }
    setShowBuilder(false)
    setEditingAutomation(null)
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this automation?')) return
    await deleteMutation.mutateAsync(id)
    if (selectedId === id) setSelectedId(null)
  }

  const isSaving = createMutation.isPending || updateMutation.isPending

  return (
    <div className="flex h-full gap-0">
      {/* Main list */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-1 pb-4">
          <div>
            <h1 className="text-xl font-bold text-slate-900">Automations</h1>
            <p className="mt-0.5 text-sm text-slate-500">
              Build trigger-based workflows to automate your CRM.
            </p>
          </div>
          <Button onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            New Automation
          </Button>
        </div>

        {/* Filter bar */}
        <div className="mb-4 flex gap-2">
          {(['', 'active', 'paused', 'draft'] as const).map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s as AutomationStatus | '')}
              className={`rounded-full px-3 py-1 text-sm font-medium transition-colors ${
                statusFilter === s
                  ? 'bg-[var(--color-primary)] text-white'
                  : 'bg-white text-slate-600 border border-slate-300 hover:bg-slate-50'
              }`}
            >
              {s === '' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>

        {/* List */}
        {isLoading ? (
          <div className="flex flex-1 items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-slate-400" />
          </div>
        ) : automations.length === 0 ? (
          <div className="flex flex-1 flex-col items-center justify-center text-center">
            <Zap className="mb-3 h-12 w-12 text-slate-200" />
            <p className="text-base font-medium text-slate-600">No automations yet</p>
            <p className="mt-1 text-sm text-slate-400">
              Create your first automation to start saving time.
            </p>
            <Button className="mt-4" onClick={openCreate}>
              <Plus className="mr-2 h-4 w-4" />
              New Automation
            </Button>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto rounded-lg border border-slate-200 bg-white divide-y divide-slate-100">
            {automations.map((a) => (
              <div
                key={a.id}
                className={`flex cursor-pointer items-center gap-4 px-4 py-3 hover:bg-slate-50 transition-colors ${
                  selectedId === a.id ? 'bg-[var(--color-primary-light)]' : ''
                }`}
                onClick={() => setSelectedId(a.id === selectedId ? null : a.id)}
              >
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--color-primary-light)]">
                  <Zap className="h-5 w-5 text-[var(--color-primary)]" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="truncate font-medium text-slate-900">{a.name}</p>
                  <p className="truncate text-xs text-slate-500">
                    {TRIGGER_LABELS[a.trigger.type]} · {a.run_count} runs
                  </p>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  {statusBadge(a.status)}
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(a.id)
                    }}
                    className="rounded p-1 text-slate-300 hover:text-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                  <ChevronRight
                    className={`h-4 w-4 text-slate-300 transition-transform ${
                      selectedId === a.id ? 'rotate-90' : ''
                    }`}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Detail panel */}
      {selectedAutomation && (
        <>
          <div
            className="fixed inset-0 z-30 lg:hidden"
            onClick={() => setSelectedId(null)}
          />
          <AutomationDetailPanel
            automation={selectedAutomation}
            onClose={() => setSelectedId(null)}
            onEdit={() => openEdit(selectedAutomation)}
          />
        </>
      )}

      {/* Builder modal */}
      {showBuilder && (
        <BuilderModal
          initial={editingAutomation}
          onClose={() => {
            setShowBuilder(false)
            setEditingAutomation(null)
          }}
          onSave={handleSave}
          saving={isSaving}
        />
      )}
    </div>
  )
}
