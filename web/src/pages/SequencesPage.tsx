import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Play,
  Pause,
  Archive,
  Trash2,
  ChevronRight,
  Mail,
  Clock,
  Users,
  BarChart2,
  ArrowLeft,
  GripVertical,
  X,
  Check,
} from 'lucide-react'
import { sequencesApi } from '@/api/sequences'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import type {
  EmailSequence,
  SequenceStatus,
  CreateStepRequest,
  StepKind,
  SequenceEnrollment,
} from '@/api/types'

// ---- Helpers ----

function statusVariant(status: SequenceStatus): 'green' | 'blue' | 'yellow' | 'gray' {
  switch (status) {
    case 'active': return 'green'
    case 'draft': return 'blue'
    case 'paused': return 'yellow'
    case 'archived':
    default: return 'gray'
  }
}

function formatPct(v: number) {
  return (v * 100).toFixed(1) + '%'
}

function enrollmentStatusVariant(status: string): 'green' | 'yellow' | 'red' | 'gray' | 'blue' {
  switch (status) {
    case 'active': return 'blue'
    case 'completed': return 'green'
    case 'paused': return 'yellow'
    case 'bounced': return 'red'
    case 'unsubscribed': return 'gray'
    default: return 'gray'
  }
}

// ---- Step editor types ----

type DraftStep = CreateStepRequest & { _key: string }

function makeKey() {
  return Math.random().toString(36).slice(2)
}

function defaultStep(kind: StepKind, position: number): DraftStep {
  if (kind === 'wait') {
    return { _key: makeKey(), kind: 'wait', position, wait_duration_hours: 24 }
  }
  return { _key: makeKey(), kind: 'email', position, subject: '', body: '' }
}

// ---- Sub-components ----

function SequenceStatusBadge({ status }: { status: SequenceStatus }) {
  return <Badge variant={statusVariant(status)}>{status}</Badge>
}

interface StepEditorProps {
  step: DraftStep
  onChange: (updated: DraftStep) => void
  onRemove: () => void
}

function StepEditor({ step, onChange, onRemove }: StepEditorProps) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="mb-3 flex items-center gap-2">
        <GripVertical className="h-4 w-4 shrink-0 cursor-grab text-slate-400" />
        <span className="flex-1 text-xs font-semibold uppercase tracking-wide text-slate-500">
          {step.kind === 'email' ? 'Email Step' : 'Wait Step'}
        </span>
        <button
          onClick={onRemove}
          className="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-red-500"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {step.kind === 'email' ? (
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Subject</label>
            <input
              className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              placeholder="Email subject line..."
              value={step.subject ?? ''}
              onChange={(e) => onChange({ ...step, subject: e.target.value })}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">Body</label>
            <textarea
              className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              rows={5}
              placeholder="Write your email body..."
              value={step.body ?? ''}
              onChange={(e) => onChange({ ...step, body: e.target.value })}
            />
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-3">
          <label className="text-sm text-slate-600">Wait for</label>
          <input
            type="number"
            min={1}
            className="w-20 rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            value={step.wait_duration_hours ?? 24}
            onChange={(e) =>
              onChange({ ...step, wait_duration_hours: Number(e.target.value) })
            }
          />
          <span className="text-sm text-slate-600">hours before next step</span>
        </div>
      )}
    </div>
  )
}

// ---- Analytics view ----

function AnalyticsView({ sequenceId }: { sequenceId: string }) {
  const { data, isLoading } = useQuery({
    queryKey: ['sequence-analytics', sequenceId],
    queryFn: () => sequencesApi.analytics(sequenceId),
  })

  if (isLoading) return <div className="flex justify-center py-8"><Spinner /></div>
  if (!data) return null

  const stats = [
    { label: 'Sent', value: data.sent },
    { label: 'Opened', value: data.opened, pct: data.open_rate },
    { label: 'Clicked', value: data.clicked, pct: data.click_rate },
    { label: 'Completed', value: data.completed },
    { label: 'Bounced', value: data.bounced },
    { label: 'Unsubscribed', value: data.unsubscribed },
  ]

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {stats.map((s) => (
          <Card key={s.label}>
            <CardContent className="pt-4">
              <p className="text-xs text-slate-500">{s.label}</p>
              <p className="text-2xl font-bold text-slate-900">{s.value}</p>
              {s.pct !== undefined && (
                <p className="text-xs text-slate-400">{formatPct(s.pct)}</p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {data.steps && data.steps.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Per-Step Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 text-left text-xs text-slate-500">
                  <th className="pb-2 font-medium">Step</th>
                  <th className="pb-2 font-medium">Sent</th>
                  <th className="pb-2 font-medium">Opened</th>
                  <th className="pb-2 font-medium">Clicked</th>
                </tr>
              </thead>
              <tbody>
                {data.steps.map((s) => (
                  <tr key={s.step_id} className="border-b border-slate-50">
                    <td className="py-2 text-slate-700">Step {s.position + 1}</td>
                    <td className="py-2">{s.sent}</td>
                    <td className="py-2">{s.opened}</td>
                    <td className="py-2">{s.clicked}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

// ---- Enrollment list ----

function EnrollmentList({ sequenceId }: { sequenceId: string }) {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['sequence-enrollments', sequenceId],
    queryFn: () => sequencesApi.listEnrollments(sequenceId),
  })

  const updateStatus = useMutation({
    mutationFn: ({ enrollmentId, status }: { enrollmentId: string; status: string }) =>
      sequencesApi.updateEnrollment(sequenceId, enrollmentId, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sequence-enrollments', sequenceId] }),
  })

  if (isLoading) return <div className="flex justify-center py-8"><Spinner /></div>

  const enrollments: SequenceEnrollment[] = data?.data ?? []

  if (enrollments.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-slate-500">
        No contacts enrolled yet.
      </p>
    )
  }

  return (
    <div className="overflow-auto rounded-lg border border-slate-200">
      <table className="w-full text-sm">
        <thead className="bg-slate-50">
          <tr className="text-left text-xs text-slate-500">
            <th className="px-4 py-3 font-medium">Contact</th>
            <th className="px-4 py-3 font-medium">Status</th>
            <th className="px-4 py-3 font-medium">Step</th>
            <th className="px-4 py-3 font-medium">Enrolled</th>
            <th className="px-4 py-3 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {enrollments.map((e) => (
            <tr key={e.id}>
              <td className="px-4 py-3">
                <p className="font-medium text-slate-800">{e.contact_name}</p>
                <p className="text-xs text-slate-400">{e.contact_email}</p>
              </td>
              <td className="px-4 py-3">
                <Badge variant={enrollmentStatusVariant(e.status)}>{e.status}</Badge>
              </td>
              <td className="px-4 py-3 text-slate-600">{e.current_step}</td>
              <td className="px-4 py-3 text-slate-500">
                {new Date(e.enrolled_at).toLocaleDateString()}
              </td>
              <td className="px-4 py-3">
                {e.status === 'active' && (
                  <button
                    onClick={() =>
                      updateStatus.mutate({ enrollmentId: e.id, status: 'paused' })
                    }
                    className="text-xs text-slate-500 hover:text-yellow-600"
                  >
                    Pause
                  </button>
                )}
                {e.status === 'paused' && (
                  <button
                    onClick={() =>
                      updateStatus.mutate({ enrollmentId: e.id, status: 'active' })
                    }
                    className="text-xs text-slate-500 hover:text-green-600"
                  >
                    Resume
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ---- Builder view ----

type BuilderTab = 'steps' | 'enrollments' | 'analytics'

interface BuilderViewProps {
  sequence: EmailSequence
  onBack: () => void
}

function BuilderView({ sequence, onBack }: BuilderViewProps) {
  const qc = useQueryClient()
  const [tab, setTab] = useState<BuilderTab>('steps')
  const [steps, setSteps] = useState<DraftStep[]>(() =>
    (sequence.steps ?? []).map((s) => ({
      _key: s.id,
      kind: s.kind,
      position: s.position,
      subject: s.subject,
      body: s.body,
      wait_duration_hours: s.wait_duration_hours,
    }))
  )
  const [saving, setSaving] = useState(false)
  const [enrollOpen, setEnrollOpen] = useState(false)
  const [enrollIds, setEnrollIds] = useState('')

  const { data: contacts } = useQuery({
    queryKey: ['contacts-lite'],
    queryFn: () =>
      import('@/api/contacts').then((m) =>
        m.contactsApi.list({ per_page: 200 }).then((r) => r.data)
      ),
    enabled: enrollOpen,
    staleTime: 60_000,
  })

  const updateMutation = useMutation({
    mutationFn: (req: Parameters<typeof sequencesApi.update>[1]) =>
      sequencesApi.update(sequence.id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sequences'] })
      qc.invalidateQueries({ queryKey: ['sequence', sequence.id] })
    },
  })

  const enrollMutation = useMutation({
    mutationFn: (ids: string[]) =>
      sequencesApi.enroll(sequence.id, { contact_ids: ids }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sequence-enrollments', sequence.id] })
      setEnrollOpen(false)
      setEnrollIds('')
    },
  })

  function addStep(kind: StepKind) {
    setSteps((prev) => [...prev, defaultStep(kind, prev.length)])
  }

  function updateStep(key: string, updated: DraftStep) {
    setSteps((prev) => prev.map((s) => (s._key === key ? updated : s)))
  }

  function removeStep(key: string) {
    setSteps((prev) =>
      prev
        .filter((s) => s._key !== key)
        .map((s, i) => ({ ...s, position: i }))
    )
  }

  async function saveSteps() {
    setSaving(true)
    try {
      await updateMutation.mutateAsync({
        steps: steps.map(({ _key: _, ...rest }) => rest),
      })
    } finally {
      setSaving(false)
    }
  }

  async function setStatus(status: SequenceStatus) {
    await updateMutation.mutateAsync({ status })
    qc.invalidateQueries({ queryKey: ['sequences'] })
    onBack()
  }

  function handleEnroll() {
    const ids = enrollIds
      .split(/[\n,]+/)
      .map((s) => s.trim())
      .filter(Boolean)
    enrollMutation.mutate(ids)
  }

  const tabs: { id: BuilderTab; label: string; icon: typeof Mail }[] = [
    { id: 'steps', label: 'Steps', icon: Mail },
    { id: 'enrollments', label: 'Enrollments', icon: Users },
    { id: 'analytics', label: 'Analytics', icon: BarChart2 },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={onBack}
          className="rounded-md p-1 text-slate-500 hover:bg-slate-100"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div className="flex-1">
          <h2 className="text-xl font-semibold text-slate-900">{sequence.name}</h2>
          <p className="text-sm text-slate-500">{sequence.description}</p>
        </div>
        <SequenceStatusBadge status={sequence.status} />
        <div className="flex gap-2">
          {sequence.status !== 'active' && sequence.status !== 'archived' && (
            <Button size="sm" onClick={() => setStatus('active')}>
              <Play className="h-4 w-4" /> Activate
            </Button>
          )}
          {sequence.status === 'active' && (
            <Button size="sm" variant="outline" onClick={() => setStatus('paused')}>
              <Pause className="h-4 w-4" /> Pause
            </Button>
          )}
          {sequence.status !== 'archived' && (
            <Button size="sm" variant="outline" onClick={() => setStatus('archived')}>
              <Archive className="h-4 w-4" /> Archive
            </Button>
          )}
          <Button size="sm" variant="outline" onClick={() => setEnrollOpen(true)}>
            <Users className="h-4 w-4" /> Enroll Contacts
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-slate-200">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={`flex items-center gap-2 border-b-2 px-4 py-2 text-sm font-medium transition-colors ${
              tab === id
                ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            <Icon className="h-4 w-4" />
            {label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === 'steps' && (
        <div className="space-y-3">
          {steps.length === 0 && (
            <p className="py-6 text-center text-sm text-slate-400">
              No steps yet. Add an email or wait step below.
            </p>
          )}
          {steps.map((step) => (
            <StepEditor
              key={step._key}
              step={step}
              onChange={(updated) => updateStep(step._key, updated)}
              onRemove={() => removeStep(step._key)}
            />
          ))}
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={() => addStep('email')}>
              <Mail className="h-4 w-4" /> Add Email Step
            </Button>
            <Button size="sm" variant="outline" onClick={() => addStep('wait')}>
              <Clock className="h-4 w-4" /> Add Wait Step
            </Button>
          </div>
          {steps.length > 0 && (
            <div className="flex justify-end pt-2">
              <Button onClick={saveSteps} disabled={saving}>
                {saving ? <Spinner className="h-4 w-4" /> : <Check className="h-4 w-4" />}
                Save Steps
              </Button>
            </div>
          )}
        </div>
      )}

      {tab === 'enrollments' && <EnrollmentList sequenceId={sequence.id} />}
      {tab === 'analytics' && <AnalyticsView sequenceId={sequence.id} />}

      {/* Enroll modal */}
      {enrollOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">Enroll Contacts</h3>
              <button
                onClick={() => setEnrollOpen(false)}
                className="rounded p-1 text-slate-400 hover:text-slate-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            {contacts && contacts.length > 0 ? (
              <div className="mb-4 max-h-64 overflow-y-auto rounded-lg border border-slate-200">
                {contacts.map((c) => {
                  const included = enrollIds.includes(c.id)
                  return (
                    <label
                      key={c.id}
                      className="flex cursor-pointer items-center gap-3 px-4 py-2 hover:bg-slate-50"
                    >
                      <input
                        type="checkbox"
                        checked={included}
                        onChange={(e) => {
                          setEnrollIds((prev) =>
                            e.target.checked
                              ? prev + (prev ? '\n' : '') + c.id
                              : prev
                                  .split('\n')
                                  .filter((id) => id !== c.id)
                                  .join('\n')
                          )
                        }}
                        className="rounded"
                      />
                      <span className="text-sm">
                        {c.first_name} {c.last_name}
                      </span>
                      <span className="text-xs text-slate-400">{c.email}</span>
                    </label>
                  )
                })}
              </div>
            ) : (
              <p className="mb-4 text-sm text-slate-500">
                Paste contact IDs (one per line or comma-separated):
              </p>
            )}
            {(!contacts || contacts.length === 0) && (
              <textarea
                className="mb-4 w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
                rows={4}
                placeholder="contact-uuid-1&#10;contact-uuid-2"
                value={enrollIds}
                onChange={(e) => setEnrollIds(e.target.value)}
              />
            )}
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setEnrollOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleEnroll}
                disabled={!enrollIds.trim() || enrollMutation.isPending}
              >
                {enrollMutation.isPending ? <Spinner className="h-4 w-4" /> : null}
                Enroll
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ---- Create sequence modal ----

function CreateSequenceModal({
  onClose,
  onCreate,
}: {
  onClose: () => void
  onCreate: (id: string) => void
}) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  const createMutation = useMutation({
    mutationFn: () => sequencesApi.create({ name, description }),
    onSuccess: (seq) => {
      onCreate(seq.id)
    },
  })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-900">New Sequence</h3>
          <button
            onClick={onClose}
            className="rounded p-1 text-slate-400 hover:text-slate-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Name</label>
            <input
              autoFocus
              className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              placeholder="e.g. Onboarding drip"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">
              Description <span className="font-normal text-slate-400">(optional)</span>
            </label>
            <textarea
              className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
              rows={2}
              placeholder="Brief description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() => createMutation.mutate()}
            disabled={!name.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? <Spinner className="h-4 w-4" /> : null}
            Create Sequence
          </Button>
        </div>
      </div>
    </div>
  )
}

// ---- Sequence list ----

function SequenceList({ onOpen }: { onOpen: (seq: EmailSequence) => void }) {
  const qc = useQueryClient()
  const [statusFilter, setStatusFilter] = useState<SequenceStatus | ''>('')
  const [showCreate, setShowCreate] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['sequences', statusFilter],
    queryFn: () =>
      sequencesApi.list(statusFilter ? { status: statusFilter as SequenceStatus } : {}),
    staleTime: 30_000,
  })

  const deleteMutation = useMutation({
    mutationFn: sequencesApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sequences'] }),
  })

  const sequences = data?.data ?? []

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Sequences</h1>
          <p className="text-sm text-slate-500">Automated email drip campaigns</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" /> New Sequence
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        {(['', 'active', 'draft', 'paused', 'archived'] as const).map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              statusFilter === s
                ? 'bg-[var(--color-primary)] text-white'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            }`}
          >
            {s === '' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {/* List */}
      {isLoading ? (
        <div className="flex justify-center py-16">
          <Spinner />
        </div>
      ) : sequences.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-300 py-16 text-center">
          <Mail className="mx-auto mb-3 h-10 w-10 text-slate-300" />
          <p className="text-slate-500">No sequences yet. Create your first one.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {sequences.map((seq) => (
            <div
              key={seq.id}
              className="flex cursor-pointer items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 hover:border-[var(--color-primary)] hover:shadow-sm"
              onClick={() => onOpen(seq)}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <p className="font-semibold text-slate-900 truncate">{seq.name}</p>
                  <SequenceStatusBadge status={seq.status} />
                </div>
                {seq.description && (
                  <p className="text-sm text-slate-500 truncate">{seq.description}</p>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-6 text-sm text-slate-500">
                <div className="text-center">
                  <p className="font-semibold text-slate-800">{seq.enrolled_count}</p>
                  <p className="text-xs">Enrolled</p>
                </div>
                <div className="text-center">
                  <p className="font-semibold text-slate-800">{formatPct(seq.open_rate)}</p>
                  <p className="text-xs">Open rate</p>
                </div>
                <div className="flex gap-1">
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      if (confirm(`Delete "${seq.name}"?`)) deleteMutation.mutate(seq.id)
                    }}
                    className="rounded p-1 text-slate-400 hover:text-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                  <ChevronRight className="h-5 w-5 text-slate-400" />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <CreateSequenceModal
          onClose={() => setShowCreate(false)}
          onCreate={(id) => {
            setShowCreate(false)
            qc.invalidateQueries({ queryKey: ['sequences'] })
            qc.fetchQuery({
              queryKey: ['sequence', id],
              queryFn: () => sequencesApi.get(id),
            }).then((seq) => onOpen(seq))
          }}
        />
      )}
    </div>
  )
}

// ---- Page root ----

export function SequencesPage() {
  const [selected, setSelected] = useState<EmailSequence | null>(null)

  if (selected) {
    return (
      <div className="mx-auto max-w-3xl">
        <BuilderView sequence={selected} onBack={() => setSelected(null)} />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl">
      <SequenceList onOpen={setSelected} />
    </div>
  )
}
