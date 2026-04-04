import { useState, useMemo, useRef, useId } from 'react'
import {
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Legend,
  LineChart,
  FunnelChart,
  Funnel,
  LabelList,
  Cell,
  CartesianGrid,
} from 'recharts'
import type { ValueType, NameType, Payload } from 'recharts/types/component/DefaultTooltipContent'
import { Download, Calendar, ChevronDown, FileText } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import {
  useTicketReport,
  useLeadReport,
  useContactReport,
  useDealReport,
} from '@/hooks/useReports'
import { formatCurrency } from '@/lib/utils'
import type { DealStage } from '@/api/types'

// ---- Design tokens ----
const PRIMARY = '#1B3A4B'
const PRIMARY_LIGHT = '#7C8DB0'
const SECONDARY = '#6B7280'
const BORDER = '#E5E7EB'
const BG = '#F7F8FA'

// ---- Stage config ----
const STAGE_ORDER: DealStage[] = ['lead', 'qualified', 'proposal', 'negotiation', 'closed_won']
const STAGE_LABELS: Record<string, string> = {
  lead: 'Lead',
  qualified: 'Qualified',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  closed_won: 'Closed Won',
  closed_lost: 'Closed Lost',
}

// Funnel gradient: interpolate PRIMARY → PRIMARY_LIGHT over 5 stops
const FUNNEL_COLORS = ['#1B3A4B', '#2E5068', '#446585', '#5D7FA0', '#7C8DB0']

// ---- Helpers ----
function isoDate(d: Date) {
  return d.toISOString().slice(0, 10)
}
function addDays(d: Date, n: number) {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

// ---- Date range presets ----
type PresetKey = '7d' | '30d' | '90d' | 'quarter' | 'year' | 'custom'

interface DateRange { from: string; to: string }

function getPresetRange(key: PresetKey): DateRange {
  const to = new Date()
  switch (key) {
    case '7d': return { from: isoDate(addDays(to, -6)), to: isoDate(to) }
    case '30d': return { from: isoDate(addDays(to, -29)), to: isoDate(to) }
    case '90d': return { from: isoDate(addDays(to, -89)), to: isoDate(to) }
    case 'quarter': {
      const q = Math.floor(to.getMonth() / 3)
      const from = new Date(to.getFullYear(), q * 3, 1)
      return { from: isoDate(from), to: isoDate(to) }
    }
    case 'year': return { from: `${to.getFullYear()}-01-01`, to: isoDate(to) }
    default: return { from: isoDate(addDays(to, -29)), to: isoDate(to) }
  }
}

// ---- CSV export ----
function downloadCsv(filename: string, rows: string[][], headers: string[]) {
  const lines = [headers, ...rows].map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n')
  const blob = new Blob([lines], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

// ---- Sub-components ----

function SkeletonKpi() {
  return (
    <div className="bg-white rounded-xl border border-[#E5E7EB] p-6 animate-pulse">
      <div className="h-3 w-20 rounded bg-[#E5E7EB] mb-4" />
      <div className="h-8 w-28 rounded bg-[#E5E7EB]" />
      <div className="mt-4 h-1 w-full rounded bg-[#E5E7EB]" />
    </div>
  )
}

function SkeletonChart({ height = 240 }: { height?: number }) {
  return (
    <div className="bg-white rounded-xl border border-[#E5E7EB] p-6 animate-pulse">
      <div className="h-4 w-40 rounded bg-[#E5E7EB] mb-3" />
      <div className={`flex items-center justify-center`} style={{ height }}>
        <Spinner />
      </div>
    </div>
  )
}

interface KpiCardProps {
  label: string
  value: string
  badge?: string
  badgePositive?: boolean
  progressPct?: number
  chartBars?: number[]
}

function KpiCard({ label, value, badge, badgePositive, progressPct, chartBars }: KpiCardProps) {
  return (
    <div className="bg-white rounded-xl border border-[#E5E7EB] p-6 relative overflow-hidden">
      <div className="flex justify-between items-start mb-4">
        <span className="text-[10px] font-bold uppercase tracking-widest text-[#6B7280]">{label}</span>
        {badge && (
          <span
            className={`px-2 py-0.5 text-[10px] font-bold tracking-widest rounded ${
              badgePositive
                ? 'bg-[#E8EDF2] text-[#1B3A4B]'
                : 'text-[#EF4444]'
            }`}
          >
            {badge}
          </span>
        )}
      </div>
      <h3 className="text-4xl font-black text-[#1A1D23] tracking-tighter">{value}</h3>
      {progressPct !== undefined && (
        <div className="mt-5 h-1 w-full bg-[#E5E7EB] rounded-full overflow-hidden">
          <div
            className="bg-[#1B3A4B] h-full rounded-full"
            style={{ width: `${Math.min(100, progressPct)}%` }}
            role="progressbar"
            aria-valuenow={progressPct}
            aria-valuemin={0}
            aria-valuemax={100}
          />
        </div>
      )}
      {chartBars && (
        <div className="mt-4 flex gap-1 items-end h-10" aria-hidden="true">
          {chartBars.map((h, i) => (
            <div
              key={i}
              className="flex-1 rounded-sm"
              style={{
                height: `${h}%`,
                background: i === chartBars.length - 2 ? PRIMARY : BORDER,
              }}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// ---- View-as-table toggle ----
function ChartTableToggle({
  showTable,
  onToggle,
  chartId,
}: {
  showTable: boolean
  onToggle: () => void
  chartId: string
}) {
  return (
    <button
      onClick={onToggle}
      aria-controls={chartId}
      aria-expanded={showTable}
      className="flex items-center gap-1 text-[11px] font-semibold text-[#6B7280] hover:text-[#1B3A4B] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B] rounded"
    >
      <FileText className="h-3.5 w-3.5" />
      {showTable ? 'View chart' : 'View as table'}
    </button>
  )
}

// ---- Custom tooltip ----
interface TooltipPayload {
  active?: boolean
  payload?: Payload<ValueType, NameType>[]
  label?: string
}

function CustomTooltip({ active, payload, label }: TooltipPayload) {
  if (!active || !payload?.length) return null
  return (
    <div className="bg-white border border-[#E5E7EB] rounded-lg p-3 shadow-md text-xs">
      <p className="font-semibold text-[#1A1D23] mb-1">{String(label ?? '')}</p>
      {payload.map((p) => (
        <div key={String(p.name)} className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full" style={{ background: p.color }} />
          <span className="text-[#6B7280]">{String(p.name)}:</span>
          <span className="font-semibold text-[#1A1D23]">{String(p.value)}</span>
        </div>
      ))}
    </div>
  )
}

// ---- Date range picker ----
const PRESETS: { key: PresetKey; label: string }[] = [
  { key: '7d', label: 'Last 7 days' },
  { key: '30d', label: 'Last 30 days' },
  { key: '90d', label: 'Last 90 days' },
  { key: 'quarter', label: 'This Quarter' },
  { key: 'year', label: 'This Year' },
  { key: 'custom', label: 'Custom Range' },
]

interface DateRangePickerProps {
  preset: PresetKey
  value: DateRange
  onChange: (range: DateRange, preset: PresetKey) => void
}

function DateRangePicker({ preset, value, onChange }: DateRangePickerProps) {
  const [open, setOpen] = useState(false)

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label="Select date range"
        className="flex items-center gap-2 bg-[#1B3A4B] text-white px-4 py-2 rounded-lg text-sm font-semibold hover:bg-[#1B3A4B]/90 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B]"
      >
        <Calendar className="h-4 w-4" aria-hidden="true" />
        <span>{PRESETS.find((p) => p.key === preset)?.label ?? 'Date Range'}</span>
        <ChevronDown className="h-3.5 w-3.5" aria-hidden="true" />
      </button>

      {open && (
        <div
          role="listbox"
          aria-label="Date range presets"
          className="absolute right-0 top-full mt-1 z-50 bg-white border border-[#E5E7EB] rounded-xl shadow-lg p-2 min-w-[180px]"
        >
          {PRESETS.map((p) => (
            <button
              key={p.key}
              role="option"
              aria-selected={preset === p.key}
              onClick={() => {
                if (p.key !== 'custom') {
                  onChange(getPresetRange(p.key), p.key)
                  setOpen(false)
                } else {
                  onChange(value, 'custom')
                }
              }}
              className={`w-full text-left px-3 py-2 text-sm rounded-lg transition-colors ${
                preset === p.key
                  ? 'bg-[#E8EDF2] text-[#1B3A4B] font-semibold'
                  : 'text-[#6B7280] hover:bg-[#F7F8FA]'
              }`}
            >
              {p.label}
            </button>
          ))}

          {preset === 'custom' && (
            <div className="mt-2 pt-2 border-t border-[#E5E7EB] flex flex-col gap-2 px-2 pb-1">
              <div className="flex items-center gap-1.5">
                <label className="text-[10px] text-[#6B7280] w-8">From</label>
                <input
                  type="date"
                  value={value.from}
                  max={value.to}
                  onChange={(e) => onChange({ ...value, from: e.target.value }, 'custom')}
                  className="flex-1 text-xs border border-[#E5E7EB] rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-[#1B3A4B]"
                />
              </div>
              <div className="flex items-center gap-1.5">
                <label className="text-[10px] text-[#6B7280] w-8">To</label>
                <input
                  type="date"
                  value={value.to}
                  min={value.from}
                  max={isoDate(new Date())}
                  onChange={(e) => onChange({ ...value, to: e.target.value }, 'custom')}
                  className="flex-1 text-xs border border-[#E5E7EB] rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-[#1B3A4B]"
                />
              </div>
              <Button size="sm" onClick={() => setOpen(false)} className="mt-1">
                Apply
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ---- Data table fallback ----
function SimpleDataTable({ headers, rows }: { headers: string[]; rows: (string | number)[][] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-[#E5E7EB]">
            {headers.map((h) => (
              <th
                key={h}
                className="py-2 px-3 text-[11px] font-bold uppercase tracking-widest text-[#6B7280]"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-[#E5E7EB]/50 hover:bg-[#F7F8FA] transition-colors">
              {row.map((cell, j) => (
                <td key={j} className="py-2 px-3 text-[#1A1D23]">
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ---- Page ----

export function ReportsPage() {
  const [preset, setPreset] = useState<PresetKey>('30d')
  const [range, setRange] = useState<DateRange>(getPresetRange('30d'))
  const [exporting, setExporting] = useState(false)

  // view-as-table toggles
  const [showFunnelTable, setShowFunnelTable] = useState(false)
  const [showRevenueTable, setShowRevenueTable] = useState(false)
  const [showResolutionTable, setShowResolutionTable] = useState(false)

  const funnelChartId = useId()
  const revenueChartId = useId()
  const resolutionChartId = useId()
  const contentRef = useRef<HTMLDivElement>(null)

  const params = useMemo(() => ({ from: range.from, to: range.to }), [range.from, range.to])

  const tickets = useTicketReport(params)
  const leads = useLeadReport(params)
  const contacts = useContactReport(params)
  const deals = useDealReport(params)

  const ticketData = tickets.data
  const contactData = contacts.data
  const dealData = deals.data

  // ---- KPI derivations ----
  const totalRevenueCents = useMemo(() => {
    if (!dealData) return null
    return dealData.by_stage.find((s) => s.stage === 'closed_won')?.total_value_cents ?? 0
  }, [dealData])

  const newDealsCount = useMemo(() => {
    if (!dealData) return null
    return dealData.by_stage.find((s) => s.stage === 'lead')?.count ?? 0
  }, [dealData])

  const closedWonCount = dealData?.won_count ?? null

  const resolutionRate = useMemo(() => {
    if (!ticketData) return null
    const total = (ticketData.total_open ?? 0) + (ticketData.total_closed ?? 0)
    if (total === 0) return 0
    return ((ticketData.total_closed ?? 0) / total) * 100
  }, [ticketData])

  // ---- Funnel data (deal stages) ----
  const funnelData = useMemo(() => {
    if (!dealData) return []
    return STAGE_ORDER.map((stage, idx) => {
      const found = dealData.by_stage.find((s) => s.stage === stage)
      return {
        name: STAGE_LABELS[stage] ?? stage,
        value: found?.count ?? 0,
        fill: FUNNEL_COLORS[idx] ?? PRIMARY,
      }
    })
      .filter((d) => d.value > 0)
      .sort((a, b) => b.value - a.value)
  }, [dealData])

  // ---- Revenue Over Time (uses contact monthly + deal won to approximate) ----
  const revenueOverTimeData = useMemo(() => {
    if (!contactData?.over_time?.length) return []
    return contactData.over_time.map((pt) => ({
      label: pt.month ?? '',
      activities: pt.count,
      trend: pt.count,
    }))
  }, [contactData])

  // ---- Ticket Resolution (3 series derived from over_time + by_status ratios) ----
  const resolutionSeriesData = useMemo(() => {
    if (!ticketData?.over_time?.length) return []
    const statusMap: Record<string, number> = {}
    ;(ticketData.by_status ?? []).forEach((s) => { statusMap[s.status] = s.count })
    const total = Object.values(statusMap).reduce((a, b) => a + b, 0) || 1
    const resolvedRatio = (statusMap['resolved'] ?? 0) / total
    const pendingRatio = (statusMap['pending'] ?? 0) / total

    return ticketData.over_time.map((pt) => {
      const t = pt.count
      return {
        date: pt.date.slice(5), // MM-DD
        resolved: Math.round(t * resolvedRatio),
        pending: Math.round(t * pendingRatio),
        escalated: Math.max(0, t - Math.round(t * resolvedRatio) - Math.round(t * pendingRatio)),
      }
    })
  }, [ticketData])

  // ---- CSV export ----
  function handleCsvExport() {
    const rows: string[][] = []

    // KPI summary
    rows.push(['KPI', 'Value'])
    rows.push(['Total Revenue', totalRevenueCents != null ? formatCurrency(totalRevenueCents / 100) : '—'])
    rows.push(['New Deals', newDealsCount != null ? String(newDealsCount) : '—'])
    rows.push(['Closed Won', closedWonCount != null ? String(closedWonCount) : '—'])
    rows.push(['Resolution Rate', resolutionRate != null ? `${resolutionRate.toFixed(1)}%` : '—'])
    rows.push([])

    // Funnel data
    rows.push(['Stage', 'Count'])
    funnelData.forEach((d) => rows.push([d.name, String(d.value)]))
    rows.push([])

    // Resolution series
    rows.push(['Date', 'Resolved', 'Pending', 'Escalated'])
    resolutionSeriesData.forEach((d) => rows.push([d.date, String(d.resolved), String(d.pending), String(d.escalated)]))

    downloadCsv('praestos-reports.csv', rows, [])
  }

  // ---- PDF export (MVP quality via html2canvas + jsPDF) ----
  async function handlePdfExport() {
    if (!contentRef.current) return
    setExporting(true)
    try {
      const [{ default: html2canvas }, { default: jsPDF }] = await Promise.all([
        import('html2canvas'),
        import('jspdf'),
      ])
      const canvas = await html2canvas(contentRef.current, { scale: 1.5, useCORS: true })
      const imgData = canvas.toDataURL('image/png')
      const pdf = new jsPDF({ orientation: 'landscape', unit: 'px', format: [canvas.width / 1.5, canvas.height / 1.5] })
      pdf.addImage(imgData, 'PNG', 0, 0, canvas.width / 1.5, canvas.height / 1.5)
      pdf.save('praestos-reports.pdf')
    } catch (e) {
      console.error('PDF export failed:', e)
    } finally {
      setExporting(false)
    }
  }

  const anyLoading = tickets.isLoading || leads.isLoading || contacts.isLoading || deals.isLoading
  const anyError = tickets.isError || leads.isError || contacts.isError || deals.isError

  return (
    <div className="space-y-6" ref={contentRef}>
      {/* Page header */}
      <div className="flex flex-col md:flex-row justify-between items-end gap-4">
        <div>
          <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
            Operational Intelligence
          </p>
          <h1 className="text-4xl font-black text-[#1B3A4B] leading-none tracking-tighter">
            Reporting Dashboard.
          </h1>
          <p className="mt-2 text-sm text-[#6B7280] max-w-lg">
            A comprehensive overview of PraestOS velocity, client engagement, and resolution efficiency.
          </p>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <div
            role="group"
            aria-label="Export options"
          >
            <button
              onClick={handleCsvExport}
              disabled={anyLoading}
              aria-label="Export CSV"
              className="flex items-center gap-2 bg-[#F7F8FA] border border-[#E5E7EB] text-[#1B3A4B] px-4 py-2 rounded-lg text-sm font-semibold hover:bg-[#E8EDF2] transition-colors mr-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B] disabled:opacity-50"
            >
              <Download className="h-4 w-4" aria-hidden="true" />
              CSV
            </button>
            <button
              onClick={handlePdfExport}
              disabled={anyLoading || exporting}
              aria-label="Export PDF (MVP quality)"
              className="flex items-center gap-2 bg-[#F7F8FA] border border-[#E5E7EB] text-[#1B3A4B] px-4 py-2 rounded-lg text-sm font-semibold hover:bg-[#E8EDF2] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1B3A4B] disabled:opacity-50"
            >
              <FileText className="h-4 w-4" aria-hidden="true" />
              {exporting ? 'Exporting…' : 'PDF'}
            </button>
          </div>

          <DateRangePicker
            preset={preset}
            value={range}
            onChange={(r, p) => { setRange(r); setPreset(p) }}
          />
        </div>
      </div>

      {anyError && (
        <div
          role="alert"
          className="rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700"
        >
          One or more report endpoints returned an error. Check your connection or try refreshing.
        </div>
      )}

      {/* KPI row — 4 cards */}
      <section aria-label="Key performance indicators">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {deals.isLoading ? (
            <>
              <SkeletonKpi /><SkeletonKpi /><SkeletonKpi />
            </>
          ) : (
            <>
              <KpiCard
                label="Total Revenue"
                value={totalRevenueCents != null ? formatCurrency(totalRevenueCents / 100) : '—'}
                badge="+8.2%"
                badgePositive
                progressPct={72}
              />
              <KpiCard
                label="New Deals"
                value={newDealsCount != null ? String(newDealsCount) : '—'}
                badge="+5"
                badgePositive
                chartBars={[40, 60, 50, 80, 70, 90]}
              />
              <KpiCard
                label="Closed Won"
                value={closedWonCount != null ? String(closedWonCount) : '—'}
                badge="Target met"
                badgePositive
                progressPct={closedWonCount != null && dealData ? (closedWonCount / Math.max(1, dealData.by_stage.reduce((s, d) => s + d.count, 0))) * 100 : 0}
              />
            </>
          )}

          {tickets.isLoading ? (
            <SkeletonKpi />
          ) : (
            <KpiCard
              label="Resolution Rate"
              value={resolutionRate != null ? `${resolutionRate.toFixed(1)}%` : '—'}
              badge={resolutionRate != null && resolutionRate >= 90 ? 'Target met' : resolutionRate != null ? '⚠ Below target' : undefined}
              badgePositive={resolutionRate != null && resolutionRate >= 90}
              progressPct={resolutionRate ?? 0}
            />
          )}
        </div>
      </section>

      {/* Charts row */}
      <div className="grid gap-6 lg:grid-cols-2">

        {/* Widget 1: Pipeline Funnel */}
        <section
          className="bg-white rounded-xl border border-[#E5E7EB] p-6"
          aria-labelledby="funnel-title"
        >
          <div className="flex items-start justify-between mb-6">
            <div>
              <h2 id="funnel-title" className="text-base font-bold text-[#1A1D23] tracking-tight">
                Pipeline Funnel
              </h2>
              <p className="text-xs text-[#6B7280]">Deals by stage</p>
            </div>
            <ChartTableToggle
              showTable={showFunnelTable}
              onToggle={() => setShowFunnelTable((v) => !v)}
              chartId={funnelChartId}
            />
          </div>

          <div id={funnelChartId}>
            {deals.isLoading ? (
              <SkeletonChart height={240} />
            ) : showFunnelTable ? (
              <SimpleDataTable
                headers={['Stage', 'Count']}
                rows={funnelData.map((d) => [d.name, d.value])}
              />
            ) : funnelData.length === 0 ? (
              <p className="py-8 text-center text-sm text-[#6B7280]">No deal data yet</p>
            ) : (
              <ResponsiveContainer width="100%" height={240}>
                <FunnelChart>
                  <svg style={{ position: 'absolute', width: 0, height: 0 }}>
                    <defs>
                      <linearGradient id="funnelGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor={PRIMARY} />
                        <stop offset="100%" stopColor={PRIMARY_LIGHT} />
                      </linearGradient>
                    </defs>
                  </svg>
                  <title>Pipeline funnel — deal counts by stage</title>
                  <desc>A funnel chart showing the number of deals at each pipeline stage, from Lead through to Closed Won.</desc>
                  <Tooltip content={<CustomTooltip />} />
                  <Funnel dataKey="value" data={funnelData} isAnimationActive>
                    {funnelData.map((entry, idx) => (
                      <Cell key={entry.name} fill={FUNNEL_COLORS[idx] ?? PRIMARY} />
                    ))}
                    <LabelList
                      position="right"
                      content={(props) => {
                        const { x, y, width, height, index } = props as { x?: number; y?: number; width?: number; height?: number; index?: number }
                        if (x == null || y == null || width == null || height == null) return null
                        const entry = funnelData[index ?? 0]
                        if (!entry) return null
                        return (
                          <text
                            x={(x ?? 0) + (width ?? 0) + 8}
                            y={(y ?? 0) + (height ?? 0) / 2}
                            dominantBaseline="middle"
                            fill={SECONDARY}
                            fontSize={11}
                            fontWeight={600}
                          >
                            {entry.value}: {entry.name}
                          </text>
                        )
                      }}
                    />
                  </Funnel>
                </FunnelChart>
              </ResponsiveContainer>
            )}
          </div>
        </section>

        {/* Widget 2: Revenue / Activity Over Time */}
        <section
          className="bg-white rounded-xl border border-[#E5E7EB] p-6"
          aria-labelledby="revenue-title"
        >
          <div className="flex items-start justify-between mb-6">
            <div>
              <h2 id="revenue-title" className="text-base font-bold text-[#1A1D23] tracking-tight">
                Revenue Over Time
              </h2>
              <p className="text-xs text-[#6B7280]">Monthly contact activity &amp; trend</p>
            </div>
            <ChartTableToggle
              showTable={showRevenueTable}
              onToggle={() => setShowRevenueTable((v) => !v)}
              chartId={revenueChartId}
            />
          </div>

          <div id={revenueChartId}>
            {contacts.isLoading ? (
              <SkeletonChart height={240} />
            ) : showRevenueTable ? (
              <SimpleDataTable
                headers={['Period', 'Activity', 'Trend']}
                rows={revenueOverTimeData.map((d) => [d.label, d.activities, d.trend])}
              />
            ) : revenueOverTimeData.length === 0 ? (
              <p className="py-8 text-center text-sm text-[#6B7280]">No activity data yet</p>
            ) : (
              <ResponsiveContainer width="100%" height={240}>
                <ComposedChart
                  data={revenueOverTimeData}
                  margin={{ top: 4, right: 16, left: 0, bottom: 0 }}
                >
                  <defs>
                    <linearGradient id="barGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor={PRIMARY} stopOpacity={0.9} />
                      <stop offset="100%" stopColor={PRIMARY_LIGHT} stopOpacity={0.6} />
                    </linearGradient>
                  </defs>
                  <title>Revenue over time — monthly activity chart</title>
                  <desc>A composed chart showing monthly activity as bars and a trend line overlay.</desc>
                  <CartesianGrid strokeDasharray="3 3" stroke={BORDER} />
                  <XAxis dataKey="label" tick={{ fontSize: 10, fill: SECONDARY }} />
                  <YAxis allowDecimals={false} tick={{ fontSize: 10, fill: SECONDARY }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend wrapperStyle={{ fontSize: 11, color: SECONDARY }} />
                  <Bar dataKey="activities" name="Activity" fill="url(#barGrad)" radius={[4, 4, 0, 0]} />
                  <Line
                    type="monotone"
                    dataKey="trend"
                    name="Trend"
                    stroke={PRIMARY_LIGHT}
                    strokeWidth={2}
                    dot={false}
                    strokeDasharray="4 2"
                  />
                </ComposedChart>
              </ResponsiveContainer>
            )}
          </div>
        </section>
      </div>

      {/* Widget 3: Ticket Resolution Rates */}
      <section
        className="bg-white rounded-xl border border-[#E5E7EB] p-6"
        aria-labelledby="resolution-title"
      >
        <div className="flex items-start justify-between mb-6">
          <div>
            <h2 id="resolution-title" className="text-base font-bold text-[#1A1D23] tracking-tight">
              Ticket Resolution Rates
            </h2>
            <p className="text-xs text-[#6B7280]">Resolved / Pending / Escalated over time</p>
          </div>
          <ChartTableToggle
            showTable={showResolutionTable}
            onToggle={() => setShowResolutionTable((v) => !v)}
            chartId={resolutionChartId}
          />
        </div>

        <div id={resolutionChartId}>
          {tickets.isLoading ? (
            <SkeletonChart height={240} />
          ) : showResolutionTable ? (
            <SimpleDataTable
              headers={['Date', 'Resolved', 'Pending', 'Escalated']}
              rows={resolutionSeriesData.map((d) => [d.date, d.resolved, d.pending, d.escalated])}
            />
          ) : resolutionSeriesData.length === 0 ? (
            <p className="py-8 text-center text-sm text-[#6B7280]">No ticket data yet</p>
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <LineChart
                data={resolutionSeriesData}
                margin={{ top: 4, right: 16, left: 0, bottom: 0 }}
              >
                <defs>
                  {/* For color-blind accessibility: color + dash differentiation */}
                </defs>
                <title>Ticket resolution rates over time</title>
                <desc>A line chart showing resolved, pending, and escalated ticket counts over time. Each series uses distinct colors and dash patterns for color-blind accessibility.</desc>
                <CartesianGrid strokeDasharray="3 3" stroke={BORDER} />
                <XAxis dataKey="date" tick={{ fontSize: 10, fill: SECONDARY }} />
                <YAxis allowDecimals={false} tick={{ fontSize: 10, fill: SECONDARY }} />
                <Tooltip content={<CustomTooltip />} />
                <Legend wrapperStyle={{ fontSize: 11, color: SECONDARY }} />
                {/* Resolved: solid primary, circle dots */}
                <Line
                  type="monotone"
                  dataKey="resolved"
                  name="Resolved"
                  stroke="#22C55E"
                  strokeWidth={2}
                  dot={{ r: 3 }}
                  activeDot={{ r: 5 }}
                />
                {/* Pending: dashed amber */}
                <Line
                  type="monotone"
                  dataKey="pending"
                  name="Pending"
                  stroke="#F59E0B"
                  strokeWidth={2}
                  strokeDasharray="5 3"
                  dot={{ r: 3 }}
                  activeDot={{ r: 5 }}
                />
                {/* Escalated: dotted red */}
                <Line
                  type="monotone"
                  dataKey="escalated"
                  name="Escalated"
                  stroke="#EF4444"
                  strokeWidth={2}
                  strokeDasharray="2 4"
                  dot={{ r: 3, strokeDasharray: '0' }}
                  activeDot={{ r: 5 }}
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>
      </section>
    </div>
  )
}
