import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ChevronLeft,
  ChevronRight,
  Plus,
  Calendar,
  Settings,
  CheckCircle2,
  XCircle,
  Link2,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs'
import {
  Dialog,
  DialogContent,
  DialogHeader,
} from '@/components/ui/Dialog'
import * as RadixDialog from '@radix-ui/react-dialog'
import { Input } from '@/components/ui/Input'
import { cn } from '@/lib/utils'
import { useActivities, useCreateActivity } from '@/hooks/useActivities'
import type { CreateActivityRequest } from '@/api/activities'
import type { ActivityType, Activity, CreatableActivityType } from '@/api/types'
import { calendarApi } from '@/api/calendar'

// ---- Helpers ----

const DAYS_OF_WEEK = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const ACTIVITY_TYPE_COLORS: Record<ActivityType, string> = {
  call: 'bg-blue-100 text-blue-800 border-blue-200',
  email: 'bg-teal-100 text-teal-800 border-teal-200',
  meeting: 'bg-green-100 text-green-800 border-green-200',
  task: 'bg-orange-100 text-orange-800 border-orange-200',
  note: 'bg-slate-100 text-slate-700 border-slate-200',
}

function toLocalDateKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function toLocalDateTimeValue(date: Date): string {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function toISOOrUndefined(localDateTime: string): string | undefined {
  if (!localDateTime) return undefined
  const parsed = new Date(localDateTime)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

function isToday(date: Date): boolean {
  const today = new Date()
  return (
    date.getFullYear() === today.getFullYear() &&
    date.getMonth() === today.getMonth() &&
    date.getDate() === today.getDate()
  )
}

// ---- Activity chip ----

function ActivityChip({ activity }: { activity: Activity }) {
  return (
    <div
      className={cn(
        'truncate rounded border px-1 py-0.5 text-xs font-medium',
        ACTIVITY_TYPE_COLORS[activity.type]
      )}
      title={activity.subject}
    >
      {activity.subject}
    </div>
  )
}

// ---- New Activity Dialog ----

interface NewActivityDialogProps {
  defaultDate: string | null
  onClose: () => void
}

function NewActivityDialog({ defaultDate, onClose }: NewActivityDialogProps) {
  const createActivity = useCreateActivity()
  const defaultStart = defaultDate ? `${defaultDate}T09:00` : toLocalDateTimeValue(new Date())
  const defaultEnd = defaultDate
    ? `${defaultDate}T10:00`
    : toLocalDateTimeValue(new Date(Date.now() + 60 * 60 * 1000))
  const [form, setForm] = useState<{
    type: CreatableActivityType
    subject: string
    description: string
    start_at: string
    end_at: string
  }>({
    type: 'task',
    subject: '',
    description: '',
    start_at: defaultStart,
    end_at: defaultEnd,
  })
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.subject.trim()) {
      setError('Subject is required')
      return
    }
    const startAt = toISOOrUndefined(form.start_at)
    if (!startAt) {
      setError('Start time is required')
      return
    }
    const endAt = toISOOrUndefined(form.end_at)
    if (endAt && new Date(endAt).getTime() < new Date(startAt).getTime()) {
      setError('End time must be after start time')
      return
    }
    const payload: CreateActivityRequest = {
      type: form.type,
      subject: form.subject.trim(),
      description: form.description.trim() || undefined,
      due_date: startAt,
      start_at: startAt,
      end_at: endAt,
    }
    try {
      await createActivity.mutateAsync(payload)
      onClose()
    } catch {
      setError('Failed to create activity')
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <RadixDialog.Title className="text-base font-semibold text-slate-900">
            New Activity
          </RadixDialog.Title>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">Type</label>
            <select
              className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
              value={form.type}
              onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as CreatableActivityType }))}
            >
              <option value="task">Task</option>
              <option value="meeting">Meeting</option>
              <option value="call">Call</option>
              <option value="email">Email</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">Subject</label>
            <Input
              value={form.subject}
              onChange={(e) => setForm((f) => ({ ...f, subject: e.target.value }))}
              placeholder="Activity subject"
              autoFocus
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">Start time</label>
            <Input
              type="datetime-local"
              value={form.start_at}
              onChange={(e) => setForm((f) => ({ ...f, start_at: e.target.value }))}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              End time <span className="text-slate-400">(optional)</span>
            </label>
            <Input
              type="datetime-local"
              value={form.end_at}
              onChange={(e) => setForm((f) => ({ ...f, end_at: e.target.value }))}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              Description <span className="text-slate-400">(optional)</span>
            </label>
            <textarea
              className="flex w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)] resize-none"
              rows={3}
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              placeholder="Optional notes"
            />
          </div>
          {error && <p className="text-xs text-red-600">{error}</p>}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={createActivity.isPending}>
              {createActivity.isPending ? 'Creating…' : 'Create Activity'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ---- Month Grid ----

interface MonthGridProps {
  year: number
  month: number
  activityMap: Map<string, Activity[]>
  onDayClick: (dateKey: string) => void
}

function MonthGrid({ year, month, activityMap, onDayClick }: MonthGridProps) {
  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)
  const startPad = firstDay.getDay()
  const totalCells = startPad + lastDay.getDate()
  const rows = Math.ceil(totalCells / 7)

  const cells: (Date | null)[] = []
  for (let i = 0; i < startPad; i++) cells.push(null)
  for (let d = 1; d <= lastDay.getDate(); d++) cells.push(new Date(year, month, d))
  while (cells.length < rows * 7) cells.push(null)

  return (
    <div className="rounded-xl border border-slate-200 bg-white overflow-hidden">
      {/* Day headers */}
      <div className="grid grid-cols-7 border-b border-slate-200 bg-slate-50">
        {DAYS_OF_WEEK.map((d) => (
          <div key={d} className="py-2 text-center text-xs font-semibold text-slate-500">
            {d}
          </div>
        ))}
      </div>

      {/* Weeks */}
      <div className="grid grid-cols-7 divide-x divide-y divide-slate-100">
        {cells.map((date, i) => {
          if (!date) {
            return <div key={`pad-${i}`} className="min-h-[90px] bg-slate-50/50" />
          }
          const key = toLocalDateKey(date)
          const events = activityMap.get(key) ?? []
          const today = isToday(date)
          return (
            <div
              key={key}
              className={cn(
                'min-h-[90px] p-1.5 cursor-pointer hover:bg-[var(--color-primary-light)]/50 transition-colors',
                today && 'bg-[var(--color-primary-light)]'
              )}
              onClick={() => onDayClick(key)}
            >
              <div className="flex items-center justify-between mb-1">
                <span
                  className={cn(
                    'inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium',
                    today
                      ? 'bg-[var(--color-primary)] text-white'
                      : 'text-slate-700'
                  )}
                >
                  {date.getDate()}
                </span>
                {events.length > 0 && (
                  <span className="text-[10px] text-slate-400">{events.length}</span>
                )}
              </div>
              <div className="space-y-0.5">
                {events.slice(0, 3).map((a) => (
                  <ActivityChip key={a.id} activity={a} />
                ))}
                {events.length > 3 && (
                  <div className="text-[10px] text-slate-400 pl-1">+{events.length - 3} more</div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ---- Week Grid ----

interface WeekGridProps {
  weekStart: Date
  activityMap: Map<string, Activity[]>
  onDayClick: (dateKey: string) => void
}

function WeekGrid({ weekStart, activityMap, onDayClick }: WeekGridProps) {
  const days: Date[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(weekStart)
    d.setDate(d.getDate() + i)
    days.push(d)
  }

  return (
    <div className="rounded-xl border border-slate-200 bg-white overflow-hidden">
      <div className="grid grid-cols-7 divide-x divide-slate-200">
        {days.map((date) => {
          const key = toLocalDateKey(date)
          const events = activityMap.get(key) ?? []
          const today = isToday(date)
          return (
            <div
              key={key}
              className={cn(
                'min-h-[300px] cursor-pointer hover:bg-[var(--color-primary-light)]/50 transition-colors',
                today && 'bg-[var(--color-primary-light)]'
              )}
              onClick={() => onDayClick(key)}
            >
              {/* Day header */}
              <div
                className={cn(
                  'border-b border-slate-200 px-2 py-2 text-center',
                  today && 'border-[var(--color-primary-light)]'
                )}
              >
                <p className="text-[11px] font-medium text-slate-500 uppercase">
                  {DAYS_OF_WEEK[date.getDay()]}
                </p>
                <span
                  className={cn(
                    'inline-flex h-7 w-7 items-center justify-center rounded-full text-sm font-semibold mt-0.5',
                    today ? 'bg-[var(--color-primary)] text-white' : 'text-slate-800'
                  )}
                >
                  {date.getDate()}
                </span>
              </div>

              {/* Events */}
              <div className="p-1.5 space-y-1">
                {events.map((a) => (
                  <ActivityChip key={a.id} activity={a} />
                ))}
                {events.length === 0 && (
                  <p className="text-[10px] text-slate-300 text-center mt-4">—</p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ---- Calendar Settings Tab ----

function CalendarSettingsTab() {
  const queryClient = useQueryClient()
  const [syncing, setSyncing] = useState(false)
  const [lastSync, setLastSync] = useState<string | null>(null)

  const { data: rawConnections, isLoading } = useQuery({
    queryKey: ['calendar-connections'],
    queryFn: calendarApi.listConnections,
  })
  const connections = Array.isArray(rawConnections) ? rawConnections : []

  const disconnectMutation = useMutation({
    mutationFn: (id: string) => calendarApi.disconnect(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['calendar-connections'] }),
  })

  const googleConnection = connections.find((c) => c.provider === 'google')
  const microsoftConnection = connections.find((c) => c.provider === 'microsoft')

  function handleConnect(provider: 'google' | 'microsoft') {
    const path = provider === 'google'
      ? '/api/v1/calendar/auth/google'
      : '/api/v1/calendar/auth/microsoft'
    window.location.href = path
  }

  async function handleSync() {
    setSyncing(true)
    try {
      await calendarApi.triggerSync()
      setLastSync(new Date().toLocaleString())
    } finally {
      setSyncing(false)
    }
  }

  if (isLoading) {
    return <div className="h-24 animate-pulse rounded-xl border border-slate-200 bg-white" />
  }

  return (
    <div className="space-y-6 max-w-xl">
      <div>
        <h2 className="text-base font-semibold text-slate-900">Connected Calendar Accounts</h2>
        <p className="mt-1 text-sm text-slate-500">
          Sync external calendar events alongside your CRM activities. Connecting an account
          will import events and keep them in sync automatically.
        </p>
      </div>

      {/* Connection cards */}
      <div className="space-y-3">
        {/* Google */}
        <div className="flex items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-red-50 border border-red-100">
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84C6.71 7.31 9.14 5.38 12 5.38z" fill="#EA4335"/>
              </svg>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-900">Google Calendar</p>
              {googleConnection ? (
                <div className="flex items-center gap-1.5 mt-0.5">
                  <CheckCircle2 className="h-3 w-3 text-green-500" />
                  <span className="text-xs text-green-600">Connected</span>
                  {lastSync && (
                    <span className="text-xs text-slate-400">· Last synced {lastSync}</span>
                  )}
                </div>
              ) : (
                <p className="text-xs text-slate-400">Not connected</p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            {googleConnection && (
              <Button variant="outline" size="sm" onClick={handleSync} disabled={syncing}>
                <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', syncing && 'animate-spin')} />
                {syncing ? 'Syncing…' : 'Sync'}
              </Button>
            )}
            {googleConnection ? (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => disconnectMutation.mutate(googleConnection.id)}
                disabled={disconnectMutation.isPending}
              >
                <XCircle className="h-3.5 w-3.5 mr-1.5" />
                Disconnect
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={() => handleConnect('google')}
                disabled={!!microsoftConnection}
              >
                <Link2 className="h-3.5 w-3.5 mr-1.5" />
                Connect
              </Button>
            )}
          </div>
        </div>

        {/* Microsoft */}
        <div className="flex items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-50 border border-blue-100">
              <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none">
                <path d="M7 4H2v16h5V4z" fill="#0078D4"/>
                <path d="M14 4H9v16h5V4z" fill="#0078D4" opacity=".7"/>
                <path d="M22 7h-6v10h6V7z" fill="#0078D4" opacity=".5"/>
              </svg>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-900">Microsoft Outlook</p>
              {microsoftConnection ? (
                <div className="flex items-center gap-1.5 mt-0.5">
                  <CheckCircle2 className="h-3 w-3 text-green-500" />
                  <span className="text-xs text-green-600">Connected</span>
                  {lastSync && (
                    <span className="text-xs text-slate-400">· Last synced {lastSync}</span>
                  )}
                </div>
              ) : (
                <p className="text-xs text-slate-400">Not connected</p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            {microsoftConnection && (
              <Button variant="outline" size="sm" onClick={handleSync} disabled={syncing}>
                <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', syncing && 'animate-spin')} />
                {syncing ? 'Syncing…' : 'Sync'}
              </Button>
            )}
            {microsoftConnection ? (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => disconnectMutation.mutate(microsoftConnection.id)}
                disabled={disconnectMutation.isPending}
              >
                <XCircle className="h-3.5 w-3.5 mr-1.5" />
                Disconnect
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={() => handleConnect('microsoft')}
                disabled={!!googleConnection}
              >
                <Link2 className="h-3.5 w-3.5 mr-1.5" />
                Connect
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Sync notice */}
      <div className="rounded-lg border border-amber-100 bg-amber-50 px-4 py-3 text-sm text-amber-800">
        <strong>Note:</strong> Calendar sync requires OAuth credentials to be configured by your administrator.
        Connect buttons will initiate the OAuth flow when credentials are set up.
      </div>
    </div>
  )
}

// ---- Main CalendarPage ----

export function CalendarPage() {
  const today = new Date()
  const [viewMode, setViewMode] = useState<'month' | 'week'>('month')
  const [currentDate, setCurrentDate] = useState(today)
  const [newActivityDate, setNewActivityDate] = useState<string | null>(null)

  // For month view: derive year/month from currentDate
  const year = currentDate.getFullYear()
  const month = currentDate.getMonth()

  // For week view: find the Sunday of the current week
  const weekStart = useMemo(() => {
    const d = new Date(currentDate)
    d.setDate(d.getDate() - d.getDay())
    d.setHours(0, 0, 0, 0)
    return d
  }, [currentDate])

  // Load activities (no server-side date range — fetch all and filter in UI)
  const { data: activitiesData, isLoading } = useActivities({ per_page: 500 })

  // Build a map: dateKey -> Activity[]
  const activityMap = useMemo(() => {
    const map = new Map<string, Activity[]>()
    for (const a of activitiesData?.data ?? []) {
      const anchorDate = a.start_at ?? a.due_date
      if (!anchorDate) continue
      // Parse the ISO date string to a local date key
      const d = new Date(anchorDate)
      const key = toLocalDateKey(d)
      const existing = map.get(key) ?? []
      existing.push(a)
      map.set(key, existing)
    }
    return map
  }, [activitiesData])

  function navigatePrev() {
    if (viewMode === 'month') {
      setCurrentDate(new Date(year, month - 1, 1))
    } else {
      const d = new Date(weekStart)
      d.setDate(d.getDate() - 7)
      setCurrentDate(d)
    }
  }

  function navigateNext() {
    if (viewMode === 'month') {
      setCurrentDate(new Date(year, month + 1, 1))
    } else {
      const d = new Date(weekStart)
      d.setDate(d.getDate() + 7)
      setCurrentDate(d)
    }
  }

  function navigateToday() {
    setCurrentDate(new Date())
  }

  function getHeaderLabel() {
    if (viewMode === 'month') {
      return `${MONTH_NAMES[month]} ${year}`
    }
    const weekEnd = new Date(weekStart)
    weekEnd.setDate(weekEnd.getDate() + 6)
    const startLabel = weekStart.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    const endLabel = weekEnd.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    return `${startLabel} – ${endLabel}`
  }

  return (
    <div className="space-y-4">
      <Tabs defaultValue="calendar">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold text-slate-900">Calendar</h1>
          <TabsList>
            <TabsTrigger value="calendar">
              <Calendar className="h-4 w-4" />
              Calendar
            </TabsTrigger>
            <TabsTrigger value="settings">
              <Settings className="h-4 w-4" />
              Settings
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Calendar tab */}
        <TabsContent value="calendar" className="pt-4">
          {/* Toolbar */}
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={navigateToday}>
                Today
              </Button>
              <button
                onClick={navigatePrev}
                className="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <button
                onClick={navigateNext}
                className="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
              <h2 className="text-sm font-semibold text-slate-800 min-w-[180px]">
                {getHeaderLabel()}
              </h2>
            </div>

            <div className="flex items-center gap-2">
              {/* View toggle */}
              <div className="flex rounded-lg border border-slate-200 overflow-hidden text-xs">
                <button
                  onClick={() => setViewMode('month')}
                  className={cn(
                    'px-3 py-1.5 font-medium transition-colors',
                    viewMode === 'month'
                      ? 'bg-[var(--color-primary)] text-white'
                      : 'bg-white text-slate-600 hover:bg-slate-50'
                  )}
                >
                  Month
                </button>
                <button
                  onClick={() => setViewMode('week')}
                  className={cn(
                    'px-3 py-1.5 font-medium transition-colors border-l border-slate-200',
                    viewMode === 'week'
                      ? 'bg-[var(--color-primary)] text-white'
                      : 'bg-white text-slate-600 hover:bg-slate-50'
                  )}
                >
                  Week
                </button>
              </div>

              <Button size="sm" onClick={() => setNewActivityDate(toLocalDateKey(today))}>
                <Plus className="h-4 w-4 mr-1" />
                New Activity
              </Button>
            </div>
          </div>

          {/* Legend */}
          <div className="mb-3 flex items-center gap-3">
            {(Object.keys(ACTIVITY_TYPE_COLORS) as ActivityType[]).map((type) => (
              <div key={type} className="flex items-center gap-1">
                <span
                  className={cn(
                    'inline-block h-2.5 w-2.5 rounded-sm border',
                    ACTIVITY_TYPE_COLORS[type]
                  )}
                />
                <span className="text-xs text-slate-500 capitalize">{type}</span>
              </div>
            ))}
            <Badge variant="gray" className="ml-auto text-xs">
              Click any day to add an activity
            </Badge>
          </div>

          {isLoading ? (
            <div className="flex h-64 items-center justify-center text-slate-400 text-sm">
              Loading activities…
            </div>
          ) : viewMode === 'month' ? (
            <MonthGrid
              year={year}
              month={month}
              activityMap={activityMap}
              onDayClick={(key) => setNewActivityDate(key)}
            />
          ) : (
            <WeekGrid
              weekStart={weekStart}
              activityMap={activityMap}
              onDayClick={(key) => setNewActivityDate(key)}
            />
          )}
        </TabsContent>

        {/* Settings tab */}
        <TabsContent value="settings" className="pt-6">
          <CalendarSettingsTab />
        </TabsContent>
      </Tabs>

      {/* New activity dialog */}
      {newActivityDate !== null && (
        <NewActivityDialog
          defaultDate={newActivityDate}
          onClose={() => setNewActivityDate(null)}
        />
      )}
    </div>
  )
}
