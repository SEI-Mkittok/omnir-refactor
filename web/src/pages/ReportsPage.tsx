import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { useReportsSummary } from '@/hooks/useReports'
import { formatCurrency } from '@/lib/utils'
import type { DealStage, ActivityType } from '@/api/types'

const DEAL_STAGE_LABELS: Record<DealStage, string> = {
  lead: 'Lead',
  qualified: 'Qualified',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  closed_won: 'Closed Won',
  closed_lost: 'Closed Lost',
}

const ACTIVITY_TYPE_LABELS: Record<ActivityType, string> = {
  call: 'Call',
  email: 'Email',
  meeting: 'Meeting',
  task: 'Task',
  note: 'Note',
}

const STAGE_COLORS: Record<DealStage, string> = {
  lead: '#94a3b8',
  qualified: '#60a5fa',
  proposal: '#818cf8',
  negotiation: '#fbbf24',
  closed_won: '#22c55e',
  closed_lost: '#f87171',
}

const ACTIVITY_COLORS = ['#6366f1', '#0ea5e9', '#22c55e', '#f59e0b', '#ec4899']

function LoadingCard() {
  return (
    <Card>
      <CardContent className="flex justify-center py-12">
        <Spinner />
      </CardContent>
    </Card>
  )
}

export function ReportsPage() {
  const { data, isLoading, isError } = useReportsSummary()

  const dealsByStage = (data?.deals_by_stage ?? []).map((d) => ({
    ...d,
    label: DEAL_STAGE_LABELS[d.stage] ?? d.stage,
    value_dollars: d.total_value_cents / 100,
  }))

  const contactsMonthly = (data?.contacts_monthly ?? []).map((c) => ({
    ...c,
    label: c.month,
  }))

  const activitiesByType = (data?.activities_by_type ?? []).map((a) => ({
    ...a,
    label: ACTIVITY_TYPE_LABELS[a.type] ?? a.type,
  }))

  const totalPipelineValue = dealsByStage.reduce((sum, d) => sum + d.total_value_cents, 0)
  const totalDeals = dealsByStage.reduce((sum, d) => sum + d.count, 0)
  const wonDeals = dealsByStage.find((d) => d.stage === 'closed_won')?.count ?? 0
  const totalActivities = activitiesByType.reduce((sum, a) => sum + a.count, 0)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Reports</h1>
        <p className="mt-1 text-sm text-slate-500">Key CRM metrics and analytics</p>
      </div>

      {isError && (
        <div className="rounded-md bg-red-50 p-4 text-sm text-red-700">
          Failed to load report data. Please try again.
        </div>
      )}

      {/* Summary cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <SummaryCard label="Total Deals" value={isLoading ? '—' : totalDeals} color="bg-indigo-50 text-indigo-600" />
        <SummaryCard label="Pipeline Value" value={isLoading ? '—' : formatCurrency(totalPipelineValue / 100)} color="bg-green-50 text-green-600" />
        <SummaryCard label="Deals Won" value={isLoading ? '—' : wonDeals} color="bg-emerald-50 text-emerald-600" />
        <SummaryCard label="Total Activities" value={isLoading ? '—' : totalActivities} color="bg-sky-50 text-sky-600" />
      </div>

      {/* Charts row 1 */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Deals by stage — count */}
        {isLoading ? (
          <LoadingCard />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Deals by Stage</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={dealsByStage} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                  <XAxis dataKey="label" tick={{ fontSize: 12 }} />
                  <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                  <Tooltip
                    formatter={(value) => [value, 'Deals']}
                    labelStyle={{ fontWeight: 600 }}
                  />
                  <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                    {dealsByStage.map((entry) => (
                      <Cell key={entry.stage} fill={STAGE_COLORS[entry.stage] ?? '#94a3b8'} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        )}

        {/* Deal value by stage */}
        {isLoading ? (
          <LoadingCard />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Pipeline Value by Stage</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={dealsByStage} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                  <XAxis dataKey="label" tick={{ fontSize: 12 }} />
                  <YAxis tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} tick={{ fontSize: 12 }} />
                  <Tooltip
                    formatter={(value) => [formatCurrency(Number(value)), 'Value']}
                    labelStyle={{ fontWeight: 600 }}
                  />
                  <Bar dataKey="value_dollars" radius={[4, 4, 0, 0]}>
                    {dealsByStage.map((entry) => (
                      <Cell key={entry.stage} fill={STAGE_COLORS[entry.stage] ?? '#94a3b8'} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Charts row 2 */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Contacts created over time */}
        {isLoading ? (
          <LoadingCard />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>New Contacts (Monthly)</CardTitle>
            </CardHeader>
            <CardContent>
              {contactsMonthly.length === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No contact data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart data={contactsMonthly} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                    <XAxis dataKey="label" tick={{ fontSize: 11 }} />
                    <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                    <Tooltip
                      formatter={(value) => [value, 'Contacts']}
                      labelStyle={{ fontWeight: 600 }}
                    />
                    <Line
                      type="monotone"
                      dataKey="count"
                      stroke="#6366f1"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                      activeDot={{ r: 5 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}

        {/* Activities by type */}
        {isLoading ? (
          <LoadingCard />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Activities by Type</CardTitle>
            </CardHeader>
            <CardContent>
              {activitiesByType.length === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No activity data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <PieChart>
                    <Pie
                      data={activitiesByType}
                      dataKey="count"
                      nameKey="label"
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                      label={({ name, percent }) =>
                        `${name} ${((percent ?? 0) * 100).toFixed(0)}%`
                      }
                      labelLine={false}
                    >
                      {activitiesByType.map((_, index) => (
                        <Cell key={index} fill={ACTIVITY_COLORS[index % ACTIVITY_COLORS.length]} />
                      ))}
                    </Pie>
                    <Legend formatter={(value) => <span className="text-sm">{value}</span>} />
                    <Tooltip formatter={(value) => [value, 'Activities']} />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

interface SummaryCardProps {
  label: string
  value: string | number
  color: string
}

function SummaryCard({ label, value, color }: SummaryCardProps) {
  return (
    <Card>
      <CardContent className="pt-5">
        <p className="text-sm font-medium text-slate-500">{label}</p>
        <p className={`mt-1 text-3xl font-bold ${color}`}>{value}</p>
      </CardContent>
    </Card>
  )
}
