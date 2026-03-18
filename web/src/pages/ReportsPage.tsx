import { useState, useMemo } from 'react'
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
  Legend,
} from 'recharts'
import { Calendar } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import {
  useTicketReport,
  useLeadReport,
  useContactReport,
  useDealReport,
} from '@/hooks/useReports'
import { formatCurrency } from '@/lib/utils'
import type { DealStage } from '@/api/types'

// ---- Constants ----

const DEAL_STAGE_LABELS: Record<DealStage, string> = {
  lead: 'Lead',
  qualified: 'Qualified',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  closed_won: 'Closed Won',
  closed_lost: 'Closed Lost',
}

const STAGE_COLORS: Record<DealStage, string> = {
  lead: '#94a3b8',
  qualified: '#60a5fa',
  proposal: '#818cf8',
  negotiation: '#fbbf24',
  closed_won: '#22c55e',
  closed_lost: '#f87171',
}

const FUNNEL_COLORS = ['#6366f1', '#818cf8', '#a5b4fc', '#c7d2fe']

// ---- Helpers ----

function isoDate(d: Date) {
  return d.toISOString().slice(0, 10)
}

function addDays(d: Date, n: number) {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

// ---- Sub-components ----

function SkeletonCard() {
  return (
    <Card>
      <CardContent className="pt-5">
        <div className="h-4 w-24 animate-pulse rounded bg-slate-200" />
        <div className="mt-2 h-8 w-20 animate-pulse rounded bg-slate-200" />
      </CardContent>
    </Card>
  )
}

function SkeletonChart() {
  return (
    <Card>
      <CardHeader>
        <div className="h-5 w-40 animate-pulse rounded bg-slate-200" />
      </CardHeader>
      <CardContent>
        <div className="flex h-[220px] items-center justify-center">
          <Spinner />
        </div>
      </CardContent>
    </Card>
  )
}

interface SummaryCardProps {
  label: string
  value: string | number
  sub?: string
  color: string
  loading?: boolean
}

function SummaryCard({ label, value, sub, color, loading }: SummaryCardProps) {
  if (loading) return <SkeletonCard />
  return (
    <Card>
      <CardContent className="pt-5">
        <p className="text-sm font-medium text-slate-500">{label}</p>
        <p className={`mt-1 text-3xl font-bold ${color}`}>{value}</p>
        {sub && <p className="mt-1 text-xs text-slate-400">{sub}</p>}
      </CardContent>
    </Card>
  )
}

// ---- Date range picker ----

interface DateRange {
  from: string
  to: string
}

function getDefaultRange(): DateRange {
  const to = new Date()
  const from = addDays(to, -29)
  return { from: isoDate(from), to: isoDate(to) }
}

interface DateRangePickerProps {
  value: DateRange
  onChange: (r: DateRange) => void
}

function DateRangePicker({ value, onChange }: DateRangePickerProps) {
  return (
    <div className="flex flex-wrap items-center gap-2 text-sm">
      <Calendar className="h-4 w-4 text-slate-400" />
      <input
        type="date"
        value={value.from}
        max={value.to}
        onChange={(e) => onChange({ ...value, from: e.target.value })}
        className="rounded border border-slate-200 bg-white px-2 py-1 text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
      <span className="text-slate-400">–</span>
      <input
        type="date"
        value={value.to}
        min={value.from}
        max={isoDate(new Date())}
        onChange={(e) => onChange({ ...value, to: e.target.value })}
        className="rounded border border-slate-200 bg-white px-2 py-1 text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
    </div>
  )
}

// ---- Page ----

export function ReportsPage() {
  const [range, setRange] = useState<DateRange>(getDefaultRange)

  const params = useMemo(
    () => ({ from: range.from, to: range.to }),
    [range.from, range.to],
  )

  const tickets = useTicketReport(params)
  const leads = useLeadReport(params)
  const contacts = useContactReport(params)
  const deals = useDealReport(params)

  const ticketData = tickets.data
  const leadData = leads.data
  const contactData = contacts.data
  const dealData = deals.data

  const dealsByStage = useMemo(
    () =>
      (dealData?.by_stage ?? []).map((d) => ({
        ...d,
        label: DEAL_STAGE_LABELS[d.stage] ?? d.stage,
        value_dollars: d.total_value_cents / 100,
      })),
    [dealData],
  )

  const anyError =
    (tickets.isError && !tickets.isPlaceholderData) ||
    (leads.isError && !leads.isPlaceholderData) ||
    (contacts.isError && !contacts.isPlaceholderData) ||
    (deals.isError && !deals.isPlaceholderData)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Reports</h1>
          <p className="mt-1 text-sm text-slate-500">Key metrics and analytics</p>
        </div>
        <DateRangePicker value={range} onChange={setRange} />
      </div>

      {anyError && (
        <div className="rounded-md bg-red-50 p-4 text-sm text-red-700">
          Some report data could not be loaded. Showing available data.
        </div>
      )}

      {/* Summary cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <SummaryCard
          label="Open Tickets"
          value={ticketData?.total_open ?? '—'}
          color="bg-indigo-50 text-indigo-600"
          loading={tickets.isLoading && !tickets.isPlaceholderData}
        />
        <SummaryCard
          label="Avg Resolution"
          value={
            ticketData?.avg_resolution_hours != null
              ? `${ticketData.avg_resolution_hours.toFixed(1)}h`
              : '—'
          }
          color="bg-sky-50 text-sky-600"
          loading={tickets.isLoading && !tickets.isPlaceholderData}
        />
        <SummaryCard
          label="SLA Breach Rate"
          value={
            ticketData?.breach_rate != null
              ? `${(ticketData.breach_rate * 100).toFixed(1)}%`
              : '—'
          }
          sub={
            ticketData?.breach_rate != null && ticketData.breach_rate > 0.15
              ? '⚠ Above target'
              : undefined
          }
          color={
            ticketData?.breach_rate != null && ticketData.breach_rate > 0.15
              ? 'text-red-600'
              : 'bg-emerald-50 text-emerald-600'
          }
          loading={tickets.isLoading && !tickets.isPlaceholderData}
        />
        <SummaryCard
          label="New Contacts"
          value={contactData?.new_count ?? '—'}
          color="bg-violet-50 text-violet-600"
          loading={contacts.isLoading && !contacts.isPlaceholderData}
        />
        <SummaryCard
          label="Pipeline Value"
          value={
            dealData?.pipeline_value_cents != null
              ? formatCurrency(dealData.pipeline_value_cents / 100)
              : '—'
          }
          color="bg-green-50 text-green-600"
          loading={deals.isLoading && !deals.isPlaceholderData}
        />
      </div>

      {/* Row 1: Tickets over time + Deal pipeline by stage */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Tickets over time */}
        {tickets.isLoading && !tickets.isPlaceholderData ? (
          <SkeletonChart />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Tickets Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              {(ticketData?.over_time.length ?? 0) === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No ticket data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <AreaChart
                    data={ticketData?.over_time}
                    margin={{ top: 4, right: 8, left: 0, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient id="ticketGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#6366f1" stopOpacity={0.18} />
                        <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <XAxis
                      dataKey="date"
                      tick={{ fontSize: 10 }}
                      tickFormatter={(v: string) => v.slice(5)}
                      interval="preserveStartEnd"
                    />
                    <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                    <Tooltip
                      formatter={(value) => [value, 'Tickets']}
                      labelStyle={{ fontWeight: 600 }}
                    />
                    <Area
                      type="monotone"
                      dataKey="count"
                      stroke="#6366f1"
                      strokeWidth={2}
                      fill="url(#ticketGrad)"
                      dot={false}
                      activeDot={{ r: 4 }}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}

        {/* Deal pipeline by stage */}
        {deals.isLoading && !deals.isPlaceholderData ? (
          <SkeletonChart />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Deal Pipeline by Stage</CardTitle>
            </CardHeader>
            <CardContent>
              {dealsByStage.length === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No deal data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart
                    data={dealsByStage}
                    margin={{ top: 4, right: 8, left: 0, bottom: 0 }}
                  >
                    <XAxis dataKey="label" tick={{ fontSize: 11 }} />
                    <YAxis
                      tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`}
                      tick={{ fontSize: 11 }}
                    />
                    <Tooltip
                      formatter={(value) => [formatCurrency(Number(value)), 'Value']}
                      labelStyle={{ fontWeight: 600 }}
                    />
                    <Bar dataKey="value_dollars" radius={[4, 4, 0, 0]}>
                      {dealsByStage.map((entry) => (
                        <Cell
                          key={entry.stage}
                          fill={STAGE_COLORS[entry.stage] ?? '#94a3b8'}
                        />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Row 2: Lead conversion funnel + New contacts over time */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Lead conversion funnel */}
        {leads.isLoading && !leads.isPlaceholderData ? (
          <SkeletonChart />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>
                Lead Conversion Funnel
                {leadData?.conversion_rate != null && (
                  <span className="ml-2 text-sm font-normal text-slate-500">
                    {(leadData.conversion_rate * 100).toFixed(1)}% conversion rate
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {(leadData?.funnel.length ?? 0) === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No lead data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart
                    data={leadData?.funnel}
                    layout="vertical"
                    margin={{ top: 4, right: 24, left: 64, bottom: 0 }}
                    barSize={28}
                  >
                    <XAxis type="number" allowDecimals={false} tick={{ fontSize: 12 }} />
                    <YAxis
                      type="category"
                      dataKey="label"
                      tick={{ fontSize: 12 }}
                      width={60}
                    />
                    <Tooltip
                      formatter={(value) => [value, 'Leads']}
                      labelStyle={{ fontWeight: 600 }}
                    />
                    <Legend
                      formatter={() => 'Leads'}
                      wrapperStyle={{ fontSize: 12 }}
                    />
                    <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                      {leadData?.funnel.map((_, i) => (
                        <Cell
                          key={i}
                          fill={FUNNEL_COLORS[i % FUNNEL_COLORS.length]}
                        />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}

        {/* New contacts over time */}
        {contacts.isLoading && !contacts.isPlaceholderData ? (
          <SkeletonChart />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>New Contacts (Monthly)</CardTitle>
            </CardHeader>
            <CardContent>
              {(contactData?.over_time.length ?? 0) === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">No contact data yet</p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <AreaChart
                    data={contactData?.over_time}
                    margin={{ top: 4, right: 8, left: 0, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient id="contactGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.18} />
                        <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <XAxis dataKey="month" tick={{ fontSize: 11 }} />
                    <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                    <Tooltip
                      formatter={(value) => [value, 'Contacts']}
                      labelStyle={{ fontWeight: 600 }}
                    />
                    <Area
                      type="monotone"
                      dataKey="count"
                      stroke="#8b5cf6"
                      strokeWidth={2}
                      fill="url(#contactGrad)"
                      dot={{ r: 3 }}
                      activeDot={{ r: 5 }}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
