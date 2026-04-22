import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import {
  UserPlus,
  CalendarPlus,
  Mail,
  FileText,
  StickyNote,
  BarChart2,
  CheckSquare,
  AlertCircle,
  User,
} from 'lucide-react'
import { ChevronDown, Check } from 'lucide-react'
import * as RadixSelect from '@radix-ui/react-select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { useDealReport, useTicketReport, useReportsSummary } from '@/hooks/useReports'
import { useActivities, useUpdateActivity } from '@/hooks/useActivities'
import { formatCompactCurrency } from '@/lib/utils'
import { cn } from '@/lib/utils'

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatMonth(ym: string): string {
  const [year, month] = ym.split('-')
  const d = new Date(parseInt(year), parseInt(month) - 1, 1)
  return d.toLocaleString('default', { month: 'short' })
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

// ─── KPI Card ─────────────────────────────────────────────────────────────────

interface KPICardProps {
  label: string
  value: string
  isLoading: boolean
  isError: boolean
  onRetry: () => void
}

function KPICard({ label, value, isLoading, isError, onRetry }: KPICardProps) {
  return (
    <article
      aria-label={`${label}: ${value}`}
      className="rounded-lg bg-[var(--surface-card)] border border-[var(--border-default)] px-6 py-5"
    >
      {isError ? (
        <div className="flex items-center gap-2 text-[var(--color-danger)] text-sm">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>Failed to load.</span>
          <button
            onClick={onRetry}
            className="flex items-center gap-1 underline hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
          >
            Retry ↺
          </button>
        </div>
      ) : isLoading ? (
        <div aria-busy="true" className="space-y-2">
          <div aria-hidden="true" className="h-3 w-24 rounded bg-[var(--border-subtle)] animate-pulse" />
          <div aria-hidden="true" className="h-9 w-32 rounded bg-[var(--border-subtle)] animate-pulse" />
        </div>
      ) : (
        <>
          <p className="text-[11px] font-semibold uppercase tracking-[0.05em] text-[var(--text-label)]">
            {label}
          </p>
          <p
            className="mt-1 text-[36px] font-bold leading-none text-[var(--text-primary)]"
            style={{ letterSpacing: 'var(--letter-spacing-tight)' }}
          >
            {value}
          </p>
        </>
      )}
    </article>
  )
}

// ─── Simple Select ────────────────────────────────────────────────────────────

interface SimpleSelectProps {
  value: string
  onValueChange: (v: string) => void
  options: { value: string; label: string }[]
}

function SimpleSelect({ value, onValueChange, options }: SimpleSelectProps) {
  return (
    <RadixSelect.Root value={value} onValueChange={onValueChange}>
      <RadixSelect.Trigger
        className="inline-flex items-center gap-1.5 rounded-md border border-[var(--border-default)] bg-white px-3 py-1.5 text-[13px] text-[var(--text-primary)] hover:border-[var(--color-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] transition-colors"
        aria-label="Chart period"
      >
        <RadixSelect.Value />
        <RadixSelect.Icon>
          <ChevronDown className="h-3.5 w-3.5 text-[var(--text-label)]" />
        </RadixSelect.Icon>
      </RadixSelect.Trigger>
      <RadixSelect.Portal>
        <RadixSelect.Content
          className="z-50 rounded-md border border-[var(--border-default)] bg-white shadow-md overflow-hidden"
          position="popper"
          sideOffset={4}
        >
          <RadixSelect.Viewport className="p-1">
            {options.map((opt) => (
              <RadixSelect.Item
                key={opt.value}
                value={opt.value}
                className="relative flex cursor-pointer select-none items-center gap-2 rounded px-3 py-1.5 text-[13px] text-[var(--text-primary)] outline-none hover:bg-[var(--color-primary-light)] data-[state=checked]:font-medium focus:bg-[var(--color-primary-light)]"
              >
                <RadixSelect.ItemText>{opt.label}</RadixSelect.ItemText>
                <RadixSelect.ItemIndicator className="absolute right-2">
                  <Check className="h-3 w-3" />
                </RadixSelect.ItemIndicator>
              </RadixSelect.Item>
            ))}
          </RadixSelect.Viewport>
        </RadixSelect.Content>
      </RadixSelect.Portal>
    </RadixSelect.Root>
  )
}

// ─── Activity Bar Chart ───────────────────────────────────────────────────────

const PERIOD_OPTIONS = [
  { value: '3', label: 'Last 3 months' },
  { value: '6', label: 'Last 6 months' },
  { value: '12', label: 'Last 12 months' },
]

function ActivityBarChart() {
  const [period, setPeriod] = useState('6')
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null)

  const from = (() => {
    const d = new Date()
    d.setMonth(d.getMonth() - parseInt(period))
    d.setDate(1)
    return d.toISOString().slice(0, 10)
  })()

  const { data, isLoading, isError, refetch, isFetching } = useReportsSummary({ from })

  const chartData = (() => {
    if (!data?.contacts_monthly) return []
    const sorted = [...data.contacts_monthly].sort((a, b) =>
      a.month.localeCompare(b.month)
    )
    return sorted.map((m) => ({
      month: formatMonth(m.month),
      value: Math.max(m.count, 0),
    }))
  })()

  return (
    <Card className="relative">
      {isFetching && !isLoading && (
        <span className="absolute top-3 right-3 text-[11px] text-[var(--text-label)]">
          Updating…
        </span>
      )}
      <CardHeader className="flex-row items-center justify-between pb-2">
        <CardTitle>Activity Overview</CardTitle>
        <SimpleSelect value={period} onValueChange={setPeriod} options={PERIOD_OPTIONS} />
      </CardHeader>
      <CardContent>
        {isError ? (
          <div className="flex items-center gap-2 text-[var(--color-danger)] text-sm py-4">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>⚠ Failed to load.</span>
            <button
              onClick={() => refetch()}
              className="underline hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
            >
              Retry ↺
            </button>
          </div>
        ) : isLoading ? (
          <div aria-busy="true" className="flex items-end gap-3 h-[240px] px-2">
            {[...Array(6)].map((_, i) => (
              <div
                key={i}
                aria-hidden="true"
                className="flex-1 rounded-t bg-[var(--border-subtle)] animate-pulse"
                style={{ height: `${30 + i * 25}%` }}
              />
            ))}
          </div>
        ) : chartData.length === 0 ? (
          <p className="py-8 text-center text-[13px] text-[var(--text-secondary)]">No data yet</p>
        ) : (
          <>
            <div role="img" aria-label="Monthly activity bar chart" className="w-full">
              <ResponsiveContainer width="100%" height={240} minWidth={0}>
                <BarChart data={chartData} margin={{ top: 4, right: 4, left: -24, bottom: 0 }}>
                  <XAxis
                    dataKey="month"
                    tick={{ fontSize: 11, fill: 'var(--text-label)' }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={{ fontSize: 11, fill: 'var(--text-label)' }}
                    axisLine={false}
                    tickLine={false}
                    allowDecimals={false}
                  />
                  <Tooltip
                    cursor={false}
                    contentStyle={{
                      border: '1px solid var(--border-default)',
                      borderRadius: 6,
                      fontSize: 12,
                      color: 'var(--text-primary)',
                      boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
                    }}
                    labelStyle={{ fontWeight: 600 }}
                  />
                  <Bar
                    dataKey="value"
                    radius={[4, 4, 0, 0]}
                    minPointSize={2}
                    isAnimationActive
                    animationBegin={0}
                    animationDuration={600}
                    animationEasing="ease-out"
                    onMouseEnter={(_, index) => setHoveredIdx(index)}
                    onMouseLeave={() => setHoveredIdx(null)}
                  >
                    {chartData.map((_, index) => (
                      <Cell
                        key={index}
                        fill={hoveredIdx === index ? '#1B3A4B' : '#7C8DB0'}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            {/* Screen-reader table */}
            <table className="sr-only">
              <caption>Monthly activity data</caption>
              <thead>
                <tr><th>Month</th><th>Count</th></tr>
              </thead>
              <tbody>
                {chartData.map((d) => (
                  <tr key={d.month}><td>{d.month}</td><td>{d.value}</td></tr>
                ))}
              </tbody>
            </table>
          </>
        )}
      </CardContent>
    </Card>
  )
}

// ─── Quick Actions ────────────────────────────────────────────────────────────

function QuickActionsCard() {
  const navigate = useNavigate()
  const [comingSoon, setComingSoon] = useState(false)

  const actions = [
    {
      icon: <UserPlus className="h-5 w-5" />,
      label: 'Add Lead',
      onClick: () => navigate('/leads'),
    },
    {
      icon: <CalendarPlus className="h-5 w-5" />,
      label: 'Schedule',
      onClick: () => navigate('/calendar'),
    },
    {
      icon: <Mail className="h-5 w-5" />,
      label: 'Email',
      onClick: () => navigate('/tickets'),
    },
    {
      icon: <FileText className="h-5 w-5" />,
      label: 'Invoice',
      soon: true,
      onClick: () => setComingSoon(true),
    },
  ]

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle>Quick Actions</CardTitle>
      </CardHeader>
      <CardContent>
        {comingSoon && (
          <div className="mb-3 rounded-md bg-[var(--color-primary-light)] px-3 py-2 text-[12px] text-[var(--text-primary)]">
            Invoicing is coming soon.
          </div>
        )}
        <div className="grid grid-cols-2 gap-2">
          {actions.map((action) => (
            <button
              key={action.label}
              onClick={action.onClick}
              aria-label={action.label}
              className="relative flex flex-col items-center gap-1.5 rounded-lg border border-[var(--border-default)] bg-[var(--surface-card)] px-3 py-4 text-[13px] text-[var(--text-primary)] hover:bg-[var(--surface-app)] hover:border-[var(--color-primary)] transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            >
              <span className="text-[var(--color-primary)]">{action.icon}</span>
              <span>{action.label}</span>
              {action.soon && (
                <span className="absolute top-1.5 right-1.5 rounded-full bg-[var(--color-warning)] px-1.5 py-0.5 text-[9px] font-semibold text-white leading-none">
                  soon
                </span>
              )}
            </button>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

// ─── Tasks Checklist ──────────────────────────────────────────────────────────

function TasksCard() {
  const { data, isLoading, isError, refetch, isFetching } = useActivities({
    type: 'task',
    per_page: 8,
    sort_by: 'due_date',
    sort_dir: 'asc',
  })
  const updateActivity = useUpdateActivity()
  const tasks = data?.data ?? []

  return (
    <Card className="relative">
      {isFetching && !isLoading && (
        <span className="absolute top-3 right-3 text-[11px] text-[var(--text-label)]">
          Updating…
        </span>
      )}
      <CardHeader className="flex-row items-center justify-between pb-2">
        <CardTitle>My Tasks</CardTitle>
        <button className="text-[13px] text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded">
          + Add task
        </button>
      </CardHeader>
      <CardContent>
        {isError ? (
          <div className="flex items-center gap-2 text-[var(--color-danger)] text-sm py-2">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>⚠ Failed to load.</span>
            <button onClick={() => refetch()} className="underline hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded">
              Retry ↺
            </button>
          </div>
        ) : isLoading ? (
          <div aria-busy="true" className="space-y-3">
            {[...Array(4)].map((_, i) => (
              <div key={i} aria-hidden="true" className="flex items-center gap-3">
                <div className="h-4 w-4 rounded bg-[var(--border-subtle)] animate-pulse shrink-0" />
                <div className="h-3 flex-1 rounded bg-[var(--border-subtle)] animate-pulse" />
              </div>
            ))}
          </div>
        ) : tasks.length === 0 ? (
          <div className="flex flex-col items-center py-8 text-center">
            <CheckSquare className="h-10 w-10 text-[var(--border-default)] mb-2" />
            <p className="text-[13px] text-[var(--text-secondary)]">No tasks yet. Add one above.</p>
          </div>
        ) : (
          <ul className="space-y-2" role="list">
            {tasks.map((task) => {
              const isOverdue =
                !task.completed &&
                task.due_date != null &&
                new Date(task.due_date) < new Date()
              return (
                <li key={task.id} className="flex items-start gap-3 py-1">
                  <input
                    id={`task-${task.id}`}
                    type="checkbox"
                    checked={task.completed ?? false}
                    onChange={(e) => {
                      updateActivity.mutate({
                        id: task.id,
                        payload: { completed: e.target.checked },
                      })
                    }}
                    className="mt-0.5 h-4 w-4 shrink-0 cursor-pointer accent-[var(--color-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
                  />
                  <label
                    htmlFor={`task-${task.id}`}
                    className={cn(
                      'flex-1 cursor-pointer text-[13px] leading-snug transition-colors',
                      task.completed
                        ? 'line-through text-[var(--text-secondary)]'
                        : 'text-[var(--text-primary)]'
                    )}
                  >
                    {task.subject}
                    {task.due_date != null && (
                      <span
                        className={cn(
                          'ml-2 text-[11px]',
                          isOverdue ? 'text-[var(--color-danger)]' : 'text-[var(--text-label)]'
                        )}
                      >
                        {new Date(task.due_date).toLocaleDateString('default', {
                          month: 'short',
                          day: 'numeric',
                        })}
                      </span>
                    )}
                  </label>
                </li>
              )
            })}
          </ul>
        )}
        {!isLoading && !isError && data != null && data.meta.total > 8 && (
          <div className="mt-3 text-center">
            <button className="text-[13px] text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded">
              View all tasks →
            </button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

// ─── Activity Feed ────────────────────────────────────────────────────────────

const ACTIVITY_ICON: Record<string, React.ReactNode> = {
  email: <Mail className="h-5 w-5 text-[var(--color-primary)]" />,
  meeting: <CalendarPlus className="h-5 w-5 text-[var(--color-primary)]" />,
  task: <CheckSquare className="h-5 w-5 text-[var(--color-primary)]" />,
  note: <StickyNote className="h-5 w-5 text-[var(--color-primary)]" />,
}

function ActivityFeedCard() {
  const [perPage, setPerPage] = useState(10)
  const { data, isLoading, isError, refetch, isFetching } = useActivities({
    per_page: perPage,
    sort_by: 'created_at',
    sort_dir: 'desc',
  })
  const items = data?.data ?? []

  return (
    <Card className="relative">
      {isFetching && !isLoading && (
        <span className="absolute top-3 right-3 text-[11px] text-[var(--text-label)]">
          Updating…
        </span>
      )}
      <CardHeader className="pb-2">
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>
      <CardContent>
        {isError ? (
          <div className="flex items-center gap-2 text-[var(--color-danger)] text-sm py-2">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>⚠ Failed to load.</span>
            <button onClick={() => refetch()} className="underline hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded">
              Retry ↺
            </button>
          </div>
        ) : isLoading ? (
          <div aria-busy="true" className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} aria-hidden="true" className="flex items-start gap-3">
                <div className="h-10 w-10 rounded-full bg-[var(--border-subtle)] animate-pulse shrink-0" />
                <div className="flex-1 space-y-2 py-1">
                  <div className="h-3 w-3/4 rounded bg-[var(--border-subtle)] animate-pulse" />
                  <div className="h-3 w-1/4 rounded bg-[var(--border-subtle)] animate-pulse" />
                </div>
              </div>
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center py-8 text-center">
            <BarChart2 className="h-10 w-10 text-[var(--border-default)] mb-2" />
            <p className="text-[13px] text-[var(--text-secondary)]">No activity yet.</p>
          </div>
        ) : (
          <ul role="list" className="space-y-4">
            {items.map((item) => (
              <li key={item.id} className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)]">
                  {ACTIVITY_ICON[item.type] ?? <User className="h-5 w-5 text-[var(--color-primary)]" />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] text-[var(--text-primary)] leading-snug truncate">
                    {item.subject}
                  </p>
                  {item.description != null && (
                    <p className="text-[12px] text-[var(--text-secondary)] truncate mt-0.5">
                      {item.description}
                    </p>
                  )}
                  <p className="text-[11px] text-[var(--text-label)] mt-0.5">
                    {relativeTime(item.created_at)}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
        {!isLoading && !isError && data != null && data.meta.total > perPage && (
          <div className="mt-4 text-center">
            <button
              onClick={() => setPerPage((p) => p + 10)}
              className="text-[13px] text-[var(--color-primary)] hover:underline focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] rounded"
            >
              Load more
            </button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

// ─── Dashboard Page ────────────────────────────────────────────────────────────

export function DashboardPage() {
  const dealReport = useDealReport()
  const ticketReport = useTicketReport()

  const revenue = (() => {
    if (dealReport.data == null) return null
    const wonStage = dealReport.data.by_stage.find((s) => s.stage === 'closed_won')
    return wonStage?.total_value_cents ?? 0
  })()

  const pipelineValue = dealReport.data?.pipeline_value_cents ?? null

  const activeDeals = (() => {
    if (dealReport.data == null) return null
    return dealReport.data.by_stage
      .filter((s) => s.stage !== 'closed_won' && s.stage !== 'closed_lost')
      .reduce((sum, s) => sum + s.count, 0)
  })()

  const openTickets = ticketReport.data?.total_open ?? null

  const kpis: KPICardProps[] = [
    {
      label: 'Total Revenue',
      value: revenue !== null ? formatCompactCurrency(revenue / 100) : '—',
      isLoading: dealReport.isLoading,
      isError: dealReport.isError,
      onRetry: () => dealReport.refetch(),
    },
    {
      label: 'Pipeline Value',
      value: pipelineValue !== null ? formatCompactCurrency(pipelineValue / 100) : '—',
      isLoading: dealReport.isLoading,
      isError: dealReport.isError,
      onRetry: () => dealReport.refetch(),
    },
    {
      label: 'Active Deals',
      value: activeDeals !== null ? String(activeDeals) : '—',
      isLoading: dealReport.isLoading,
      isError: dealReport.isError,
      onRetry: () => dealReport.refetch(),
    },
    {
      label: 'Open Tickets',
      value: openTickets !== null ? String(openTickets) : '—',
      isLoading: ticketReport.isLoading,
      isError: ticketReport.isError,
      onRetry: () => ticketReport.refetch(),
    },
  ]

  return (
    <div className="space-y-6">
      {/* KPI Row */}
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        {kpis.map((kpi) => (
          <KPICard key={kpi.label} {...kpi} />
        ))}
      </div>

      {/* Chart + Quick Actions */}
      <div className="grid gap-4 grid-cols-1 lg:grid-cols-4">
        <div className="lg:col-span-3">
          <ActivityBarChart />
        </div>
        <div className="lg:col-span-1">
          <QuickActionsCard />
        </div>
      </div>

      {/* Tasks + Activity Feed */}
      <div className="grid gap-4 grid-cols-1 md:grid-cols-2">
        <TasksCard />
        <ActivityFeedCard />
      </div>
    </div>
  )
}
