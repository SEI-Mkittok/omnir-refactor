import { useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Users, Building2, TrendingUp, DollarSign, AlertCircle } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Cell,
  ResponsiveContainer,
  PieChart,
  Pie,
  Legend,
} from 'recharts'
import { useContacts } from '@/hooks/useContacts'
import { useAccounts } from '@/hooks/useAccounts'
import { useDeals } from '@/hooks/useDeals'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { formatCurrency, formatDate } from '@/lib/utils'
import { Spinner } from '@/components/ui/Spinner'
import {
  getPipelineFunnel,
  getRevenueProjection,
  getConversionRates,
  getActivitySummary,
} from '@/api/reports'

// ---- Date range helpers ----

function toISODate(d: Date) {
  return d.toISOString().split('T')[0]
}

function defaultRange(): { from: string; to: string } {
  const to = new Date()
  const from = new Date()
  from.setDate(from.getDate() - 90)
  return { from: toISODate(from), to: toISODate(to) }
}

// ---- Stat card ----

interface StatCardProps {
  title: string
  value: string | number
  icon: React.ReactNode
  description?: string
  color: string
}

function StatCard({ title, value, icon, description, color }: StatCardProps) {
  return (
    <Card>
      <CardContent className="pt-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">{title}</p>
            <p className="mt-1 text-3xl font-bold text-slate-900">{value}</p>
            {description && (
              <p className="mt-1 text-xs text-slate-500">{description}</p>
            )}
          </div>
          <div className={`flex h-12 w-12 items-center justify-center rounded-xl ${color}`}>
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

// ---- Widget wrapper ----

function Widget({
  title,
  isLoading,
  error,
  children,
}: {
  title: string
  isLoading: boolean
  error: boolean
  children: React.ReactNode
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Spinner className="h-5 w-5 text-slate-400" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-slate-400">
            <AlertCircle className="h-6 w-6" />
            <p className="text-sm">Failed to load data</p>
          </div>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  )
}

// ---- Chart colours ----

const STAGE_COLORS: Record<string, string> = {
  lead: '#94a3b8',
  qualified: '#60a5fa',
  proposal: '#818cf8',
  negotiation: '#fbbf24',
  closed_won: '#22c55e',
  closed_lost: '#f87171',
}

const ACTIVITY_COLORS: Record<string, string> = {
  call: '#6366f1',
  email: '#3b82f6',
  meeting: '#10b981',
  task: '#f59e0b',
  note: '#94a3b8',
}

function conversionColor(rate: number) {
  if (rate >= 0.5) return '#22c55e'
  if (rate >= 0.25) return '#fbbf24'
  return '#f87171'
}

// ---- Main page ----

export function DashboardPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  const defaults = defaultRange()
  const from = searchParams.get('from') ?? defaults.from
  const to = searchParams.get('to') ?? defaults.to

  const rangeParams = { from, to }

  const setDateRange = useCallback(
    (newFrom: string, newTo: string) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev)
        next.set('from', newFrom)
        next.set('to', newTo)
        return next
      })
    },
    [setSearchParams]
  )

  // Top-level stat queries
  const contacts = useContacts({ per_page: 1 })
  const accounts = useAccounts({ per_page: 1 })
  const deals = useDeals({ per_page: 100 })

  const totalDealValue = deals.data?.data.reduce((sum, d) => sum + d.value, 0) ?? 0
  const openDeals = deals.data?.data.filter(
    (d) => d.stage !== 'closed_won' && d.stage !== 'closed_lost'
  ).length ?? 0

  // Widget queries
  const funnelQ = useQuery({
    queryKey: ['report-pipeline-funnel', from, to],
    queryFn: () => getPipelineFunnel(rangeParams),
    staleTime: 60_000,
  })

  const revenueQ = useQuery({
    queryKey: ['report-revenue-projection', from, to],
    queryFn: () => getRevenueProjection(rangeParams),
    staleTime: 60_000,
  })

  const conversionQ = useQuery({
    queryKey: ['report-conversion-rates', from, to],
    queryFn: () => getConversionRates(rangeParams),
    staleTime: 60_000,
  })

  const activityQ = useQuery({
    queryKey: ['report-activity-summary', from, to],
    queryFn: () => getActivitySummary(rangeParams),
    staleTime: 60_000,
  })

  // ---- Render ----

  return (
    <div className="space-y-6">
      {/* Header + date range */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>
          <p className="mt-1 text-sm text-slate-500">Your CRM overview at a glance</p>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <label className="text-slate-500 shrink-0">From</label>
          <input
            type="date"
            value={from}
            max={to}
            onChange={(e) => setDateRange(e.target.value, to)}
            className="rounded-md border border-slate-200 px-2 py-1 text-sm text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          <label className="text-slate-500 shrink-0">To</label>
          <input
            type="date"
            value={to}
            min={from}
            onChange={(e) => setDateRange(from, e.target.value)}
            className="rounded-md border border-slate-200 px-2 py-1 text-sm text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Contacts"
          value={contacts.isLoading ? '—' : (contacts.data?.meta.total ?? 0)}
          icon={<Users className="h-6 w-6 text-blue-600" />}
          description="All contacts"
          color="bg-blue-50"
        />
        <StatCard
          title="Accounts"
          value={accounts.isLoading ? '—' : (accounts.data?.meta.total ?? 0)}
          icon={<Building2 className="h-6 w-6 text-purple-600" />}
          description="Active accounts"
          color="bg-purple-50"
        />
        <StatCard
          title="Open Deals"
          value={deals.isLoading ? '—' : openDeals}
          icon={<TrendingUp className="h-6 w-6 text-indigo-600" />}
          description="Deals in pipeline"
          color="bg-indigo-50"
        />
        <StatCard
          title="Pipeline Value"
          value={deals.isLoading ? '—' : formatCurrency(totalDealValue)}
          icon={<DollarSign className="h-6 w-6 text-green-600" />}
          description="Total deal value"
          color="bg-green-50"
        />
      </div>

      {/* Charts row 1: funnel + revenue */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* Widget 1: Pipeline Funnel */}
        <Widget
          title="Pipeline Funnel"
          isLoading={funnelQ.isLoading}
          error={!!funnelQ.error}
        >
          {funnelQ.data && (
            <FunnelWidget stages={funnelQ.data.stages} />
          )}
        </Widget>

        {/* Widget 2: Revenue Projection */}
        <Widget
          title="Revenue Projection"
          isLoading={revenueQ.isLoading}
          error={!!revenueQ.error}
        >
          {revenueQ.data && (
            <RevenueWidget months={revenueQ.data.months} />
          )}
        </Widget>
      </div>

      {/* Charts row 2: conversion + activity */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* Widget 3: Conversion Rates */}
        <Widget
          title="Stage Conversion Rates"
          isLoading={conversionQ.isLoading}
          error={!!conversionQ.error}
        >
          {conversionQ.data && (
            <ConversionWidget stages={conversionQ.data.stages} />
          )}
        </Widget>

        {/* Widget 4: Activity Summary */}
        <Widget
          title={activityQ.data ? `Activity Summary — ${activityQ.data.period_label}` : 'Activity Summary'}
          isLoading={activityQ.isLoading}
          error={!!activityQ.error}
        >
          {activityQ.data && (
            <ActivityWidget activities={activityQ.data.activities} />
          )}
        </Widget>
      </div>
    </div>
  )
}

// ---- Widget sub-components ----

import type { PipelineFunnelStage, RevenueProjectionMonth, ConversionRateStage, ActivitySummaryItem } from '@/api/types'

type FunnelView = 'count' | 'value'

function FunnelWidget({ stages }: { stages: PipelineFunnelStage[] }) {
  const [view, setView] = React.useState<FunnelView>('count')

  const chartData = stages.map((s) => ({
    label: s.label,
    stage: s.stage,
    value: view === 'count' ? s.count : s.value_cents / 100,
  }))

  return (
    <div className="space-y-3">
      <div className="flex gap-1 rounded-lg border border-slate-200 bg-slate-50 p-1 w-fit text-xs">
        {(['count', 'value'] as FunnelView[]).map((v) => (
          <button
            key={v}
            onClick={() => setView(v)}
            className={`rounded px-3 py-1 font-medium capitalize transition-colors ${
              view === v ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            {v === 'count' ? 'Count' : 'Value'}
          </button>
        ))}
      </div>
      <ResponsiveContainer width="100%" height={220}>
        <BarChart data={chartData} layout="vertical" margin={{ left: 8, right: 24, top: 0, bottom: 0 }}>
          <XAxis
            type="number"
            tickFormatter={view === 'value' ? (v) => `$${(v / 1000).toFixed(0)}k` : undefined}
            tick={{ fontSize: 11 }}
          />
          <YAxis type="category" dataKey="label" width={90} tick={{ fontSize: 11 }} />
          <Tooltip
            formatter={(v) =>
              v != null ? (view === 'value' ? formatCurrency(Number(v)) : String(v)) : ''
            }
          />
          <Bar dataKey="value" radius={[0, 4, 4, 0]}>
            {chartData.map((entry) => (
              <Cell
                key={entry.stage}
                fill={STAGE_COLORS[entry.stage] ?? '#818cf8'}
              />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

function RevenueWidget({ months }: { months: RevenueProjectionMonth[] }) {
  const chartData = months.map((m) => ({
    month: formatDate(m.month + '-01', { month: 'short', year: '2-digit' }),
    projected: m.projected_revenue_cents / 100,
    weighted: m.weighted_value_cents / 100,
    deals: m.deal_count,
  }))

  return (
    <ResponsiveContainer width="100%" height={240}>
      <BarChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 0 }}>
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} tick={{ fontSize: 11 }} />
        <Tooltip
          formatter={(v, name) => [
            v != null ? formatCurrency(Number(v)) : '',
            name === 'projected' ? 'Projected Revenue' : 'Weighted Value',
          ]}
          labelFormatter={(label) => `Month: ${label}`}
        />
        <Bar dataKey="projected" name="projected" fill="#6366f1" radius={[4, 4, 0, 0]} />
        <Bar dataKey="weighted" name="weighted" fill="#a5b4fc" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  )
}

function ConversionWidget({ stages }: { stages: ConversionRateStage[] }) {
  if (stages.length === 0) {
    return <p className="py-8 text-center text-sm text-slate-400">No conversion data available</p>
  }

  return (
    <div className="space-y-3">
      {stages.map((s) => {
        const pct = Math.round(s.conversion_rate * 100)
        const color = conversionColor(s.conversion_rate)
        return (
          <div key={`${s.from_stage}-${s.to_stage}`}>
            <div className="flex items-center justify-between text-sm mb-1">
              <span className="text-slate-600">
                {s.from_label} → {s.to_label}
              </span>
              <span className="font-semibold" style={{ color }}>{pct}%</span>
            </div>
            <div className="h-2 w-full rounded-full bg-slate-100">
              <div
                className="h-2 rounded-full transition-all"
                style={{ width: `${pct}%`, backgroundColor: color }}
              />
            </div>
          </div>
        )
      })}
      <div className="mt-3 flex gap-4 text-xs text-slate-500">
        <span className="flex items-center gap-1"><span className="inline-block h-2 w-2 rounded-full bg-green-500" /> ≥50%</span>
        <span className="flex items-center gap-1"><span className="inline-block h-2 w-2 rounded-full bg-yellow-400" /> 25–50%</span>
        <span className="flex items-center gap-1"><span className="inline-block h-2 w-2 rounded-full bg-red-400" /> &lt;25%</span>
      </div>
    </div>
  )
}

function ActivityWidget({ activities }: { activities: ActivitySummaryItem[] }) {
  const chartData = activities
    .filter((a) => a.count > 0)
    .map((a) => ({
      name: a.type.charAt(0).toUpperCase() + a.type.slice(1),
      type: a.type,
      value: a.count,
    }))

  if (chartData.length === 0) {
    return <p className="py-8 text-center text-sm text-slate-400">No activity data for this period</p>
  }

  return (
    <ResponsiveContainer width="100%" height={220}>
      <PieChart>
        <Pie
          data={chartData}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          outerRadius={80}
          innerRadius={48}
          paddingAngle={2}
        >
          {chartData.map((entry) => (
            <Cell
              key={entry.type}
              fill={ACTIVITY_COLORS[entry.type] ?? '#94a3b8'}
            />
          ))}
        </Pie>
        <Tooltip formatter={(v, name) => [v, name]} />
        <Legend iconType="circle" iconSize={8} wrapperStyle={{ fontSize: '12px' }} />
      </PieChart>
    </ResponsiveContainer>
  )
}

// React import needed for useState inside module
import React from 'react'
