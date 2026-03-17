import { useState } from 'react'
import {
  Phone,
  Mail,
  Users,
  CheckSquare,
  FileText,
  Plus,
  Check,
  Loader2,
  Clock,
  ChevronDown,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { Input } from '@/components/ui/Input'
import { cn, formatRelativeTime, formatDate } from '@/lib/utils'
import { useContactActivities, useDealActivities, useCreateActivity, useUpdateActivity } from '@/hooks/useActivities'
import type { Activity, ActivityType } from '@/api/types'

// ── Activity type config ─────────────────────────────────────────────────────

const typeConfig: Record<
  ActivityType,
  { label: string; Icon: React.ElementType; color: string; badgeVariant: 'blue' | 'purple' | 'green' | 'yellow' | 'gray' }
> = {
  call: { label: 'Call', Icon: Phone, color: 'text-blue-500', badgeVariant: 'blue' },
  email: { label: 'Email', Icon: Mail, color: 'text-purple-500', badgeVariant: 'purple' },
  meeting: { label: 'Meeting', Icon: Users, color: 'text-green-500', badgeVariant: 'green' },
  task: { label: 'Task', Icon: CheckSquare, color: 'text-yellow-500', badgeVariant: 'yellow' },
  note: { label: 'Note', Icon: FileText, color: 'text-slate-500', badgeVariant: 'gray' },
}

const ALL_TYPES: ActivityType[] = ['call', 'email', 'meeting', 'task', 'note']

// ── Quick-add form ───────────────────────────────────────────────────────────

interface QuickAddFormProps {
  contactId?: string
  dealId?: string
  onSuccess: () => void
}

function QuickAddForm({ contactId, dealId, onSuccess }: QuickAddFormProps) {
  const [open, setOpen] = useState(false)
  const [type, setType] = useState<ActivityType>('call')
  const [subject, setSubject] = useState('')
  const [description, setDescription] = useState('')
  const [dueDate, setDueDate] = useState('')
  const createActivity = useCreateActivity()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!subject.trim()) return
    await createActivity.mutateAsync({
      type,
      subject: subject.trim(),
      description: description.trim() || undefined,
      due_date: dueDate || undefined,
      contact_id: contactId,
      deal_id: dealId,
    })
    setSubject('')
    setDescription('')
    setDueDate('')
    setOpen(false)
    onSuccess()
  }

  if (!open) {
    return (
      <Button size="sm" variant="outline" onClick={() => setOpen(true)}>
        <Plus className="h-3.5 w-3.5" />
        Log Activity
      </Button>
    )
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-lg border border-indigo-200 bg-indigo-50 p-3 space-y-3"
    >
      {/* Type selector */}
      <div className="flex flex-wrap gap-1.5">
        {ALL_TYPES.map((t) => {
          const { label, Icon } = typeConfig[t]
          return (
            <button
              key={t}
              type="button"
              onClick={() => setType(t)}
              className={cn(
                'flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
                type === t
                  ? 'border-indigo-600 bg-indigo-600 text-white'
                  : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
              )}
            >
              <Icon className="h-3 w-3" />
              {label}
            </button>
          )
        })}
      </div>

      {/* Subject */}
      <Input
        placeholder="Subject…"
        value={subject}
        onChange={(e) => setSubject(e.target.value)}
        className="h-8 text-sm"
        autoFocus
        required
      />

      {/* Description */}
      <textarea
        placeholder="Description (optional)…"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        rows={2}
        className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-1.5 text-sm shadow-sm placeholder:text-slate-400 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
      />

      {/* Due date */}
      <div className="flex items-center gap-2">
        <Clock className="h-3.5 w-3.5 text-slate-400 shrink-0" />
        <input
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          className="flex-1 rounded-md border border-slate-200 bg-white px-2.5 py-1 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
        />
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={() => setOpen(false)}
          className="text-sm text-slate-500 hover:text-slate-700"
        >
          Cancel
        </button>
        <Button type="submit" size="sm" disabled={!subject.trim() || createActivity.isPending}>
          {createActivity.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Check className="h-3.5 w-3.5" />
          )}
          Save
        </Button>
      </div>
    </form>
  )
}

// ── Single activity item ─────────────────────────────────────────────────────

interface ActivityItemProps {
  activity: Activity
  isLast: boolean
}

function ActivityItem({ activity, isLast }: ActivityItemProps) {
  const { Icon, color, badgeVariant, label } = typeConfig[activity.type]
  const updateActivity = useUpdateActivity()
  const [expanded, setExpanded] = useState(false)

  const toggleComplete = async () => {
    await updateActivity.mutateAsync({
      id: activity.id,
      payload: { completed: !activity.completed },
    })
  }

  return (
    <div className="flex gap-3">
      {/* Timeline line + icon */}
      <div className="flex flex-col items-center">
        <div
          className={cn(
            'flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2',
            activity.completed
              ? 'border-green-200 bg-green-50'
              : 'border-slate-200 bg-white'
          )}
        >
          <Icon className={cn('h-3.5 w-3.5', activity.completed ? 'text-green-500' : color)} />
        </div>
        {!isLast && <div className="mt-1 w-px flex-1 bg-slate-100" />}
      </div>

      {/* Content */}
      <div className={cn('min-w-0 flex-1 pb-4', isLast && 'pb-0')}>
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-1.5 mb-0.5">
              <Badge variant={badgeVariant} className="text-xs px-1.5 py-0">
                {label}
              </Badge>
              {activity.due_date && (
                <span className="flex items-center gap-1 text-xs text-slate-400">
                  <Clock className="h-3 w-3" />
                  {formatDate(activity.due_date)}
                </span>
              )}
            </div>
            <p
              className={cn(
                'text-sm font-medium',
                activity.completed ? 'text-slate-400 line-through' : 'text-slate-900'
              )}
            >
              {activity.subject}
            </p>
            {activity.description && (
              <>
                <p
                  className={cn(
                    'mt-0.5 text-xs text-slate-500',
                    !expanded && 'line-clamp-2'
                  )}
                >
                  {activity.description}
                </p>
                {activity.description.length > 120 && (
                  <button
                    onClick={() => setExpanded((v) => !v)}
                    className="flex items-center gap-0.5 text-xs text-indigo-500 hover:text-indigo-700 mt-0.5"
                  >
                    <ChevronDown
                      className={cn('h-3 w-3 transition-transform', expanded && 'rotate-180')}
                    />
                    {expanded ? 'Less' : 'More'}
                  </button>
                )}
              </>
            )}
          </div>

          {/* Complete toggle (only for task/call/meeting) */}
          {activity.type !== 'note' && (
            <button
              onClick={toggleComplete}
              disabled={updateActivity.isPending}
              title={activity.completed ? 'Mark incomplete' : 'Mark complete'}
              className={cn(
                'flex h-5 w-5 shrink-0 items-center justify-center rounded border transition-colors mt-0.5',
                activity.completed
                  ? 'border-green-300 bg-green-50 text-green-600 hover:bg-green-100'
                  : 'border-slate-300 bg-white text-transparent hover:border-green-400 hover:text-green-500'
              )}
            >
              <Check className="h-3 w-3" />
            </button>
          )}
        </div>

        <p className="mt-1 text-xs text-slate-400">{formatRelativeTime(activity.created_at)}</p>
      </div>
    </div>
  )
}

// ── Main timeline component ──────────────────────────────────────────────────

interface ActivityTimelineProps {
  contactId?: string
  dealId?: string
}

export function ActivityTimeline({ contactId, dealId }: ActivityTimelineProps) {
  const [typeFilter, setTypeFilter] = useState<ActivityType | ''>('')

  const contactQuery = useContactActivities(contactId ?? '')
  const dealQuery = useDealActivities(dealId ?? '')

  const query = contactId ? contactQuery : dealQuery
  const { data, isLoading, refetch } = query

  const activities = (data?.data ?? []).filter(
    (a) => !typeFilter || a.type === typeFilter
  )

  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-2">
        <h3 className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-slate-700">
          <Clock className="h-4 w-4" />
          Activity
          {data && data.meta.total > 0 && (
            <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-normal text-slate-600">
              {data.meta.total}
            </span>
          )}
        </h3>
        <QuickAddForm
          contactId={contactId}
          dealId={dealId}
          onSuccess={() => refetch()}
        />
      </div>

      {/* Type filter pills */}
      <div className="mb-4 flex flex-wrap gap-1.5">
        <button
          onClick={() => setTypeFilter('')}
          className={cn(
            'rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors',
            typeFilter === ''
              ? 'border-slate-700 bg-slate-700 text-white'
              : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
          )}
        >
          All
        </button>
        {ALL_TYPES.map((t) => {
          const { label, Icon } = typeConfig[t]
          return (
            <button
              key={t}
              onClick={() => setTypeFilter(typeFilter === t ? '' : t)}
              className={cn(
                'flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors',
                typeFilter === t
                  ? 'border-slate-700 bg-slate-700 text-white'
                  : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
              )}
            >
              <Icon className="h-3 w-3" />
              {label}
            </button>
          )
        })}
      </div>

      {/* Timeline */}
      {isLoading ? (
        <div className="flex justify-center py-6">
          <Spinner />
        </div>
      ) : activities.length > 0 ? (
        <div>
          {activities.map((activity, index) => (
            <ActivityItem
              key={activity.id}
              activity={activity}
              isLast={index === activities.length - 1}
            />
          ))}
        </div>
      ) : (
        <p className="text-sm text-slate-400">
          {typeFilter ? `No ${typeConfig[typeFilter].label.toLowerCase()} activities yet.` : 'No activities yet.'}
        </p>
      )}
    </div>
  )
}
