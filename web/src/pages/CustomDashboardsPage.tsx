import { useState, useEffect, useId } from 'react'
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  useSortable,
  rectSortingStrategy,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import {
  BarChart2, LineChart, PieChart, Activity, TrendingUp, Users,
  Mail, Ticket, Plus, Trash2, GripVertical, Save, Edit2, X,
  Calendar, Clock, ChevronDown, MoreVertical, Play,
} from 'lucide-react'
import {
  ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip,
  LineChart as RLineChart, Line, PieChart as RPieChart, Pie, Cell,
  CartesianGrid,
} from 'recharts'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import {
  Dialog, DialogPortal, DialogOverlay, DialogContent,
  DialogClose,
} from '@/components/ui/Dialog'
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem,
} from '@/components/ui/DropdownMenu'
import {
  useDashboards, useCreateDashboard, useUpdateDashboard, useDeleteDashboard,
  useRunDashboard, useSchedules, useCreateSchedule, useUpdateSchedule, useDeleteSchedule,
} from '@/hooks/useDashboards'
import type { Widget } from '@/api/dashboards'
import type { ScheduledReport } from '@/api/dashboards'
import { cn } from '@/lib/utils'

// ── Design tokens ────────────────────────────────────────────────────────────

const PRIMARY = '#1B3A4B'
const BORDER = '#E5E7EB'
const BG = '#F7F8FA'

// ── Widget catalogue ─────────────────────────────────────────────────────────

interface WidgetMeta {
  type: string
  label: string
  description: string
  icon: React.ElementType
  defaultSize: { w: number; h: number }
}

const WIDGET_CATALOGUE: WidgetMeta[] = [
  { type: 'deals_by_stage', label: 'Deals by Stage', description: 'Pipeline funnel breakdown', icon: BarChart2, defaultSize: { w: 2, h: 2 } },
  { type: 'contacts_monthly', label: 'Contacts Monthly', description: 'New contacts over time', icon: LineChart, defaultSize: { w: 2, h: 2 } },
  { type: 'activities_by_type', label: 'Activities by Type', description: 'Activity distribution', icon: Activity, defaultSize: { w: 2, h: 2 } },
  { type: 'ticket_metrics', label: 'Ticket Metrics', description: 'Support KPIs', icon: Ticket, defaultSize: { w: 1, h: 1 } },
  { type: 'deal_metrics', label: 'Deal Metrics', description: 'Revenue & pipeline KPIs', icon: TrendingUp, defaultSize: { w: 1, h: 1 } },
  { type: 'contact_metrics', label: 'Contact Metrics', description: 'Contact growth KPIs', icon: Users, defaultSize: { w: 1, h: 1 } },
  { type: 'lead_metrics', label: 'Lead Metrics', description: 'Lead funnel KPIs', icon: Activity, defaultSize: { w: 1, h: 1 } },
  { type: 'revenue_projection', label: 'Revenue Projection', description: '3-month revenue forecast', icon: TrendingUp, defaultSize: { w: 2, h: 2 } },
  { type: 'activity_summary', label: 'Activity Summary', description: 'Team activity rollup', icon: Activity, defaultSize: { w: 2, h: 1 } },
  { type: 'pipeline_funnel', label: 'Pipeline Funnel', description: 'Stage-by-stage funnel', icon: BarChart2, defaultSize: { w: 2, h: 2 } },
  { type: 'conversion_rates', label: 'Conversion Rates', description: 'Lead → deal conversion', icon: PieChart, defaultSize: { w: 2, h: 2 } },
]

const PIE_COLORS = ['#1B3A4B', '#2E5068', '#446585', '#5D7FA0', '#7C8DB0', '#A0B3C0']

function widgetMeta(type: string): WidgetMeta {
  return WIDGET_CATALOGUE.find((w) => w.type === type) ?? {
    type,
    label: type.replace(/_/g, ' '),
    description: '',
    icon: BarChart2,
    defaultSize: { w: 1, h: 1 },
  }
}

// ── Schedule helpers ─────────────────────────────────────────────────────────

const FREQ_OPTIONS = [
  { label: 'Daily at 8 AM', value: '0 8 * * *' },
  { label: 'Weekly (Mon 8 AM)', value: '0 8 * * 1' },
  { label: 'Monthly (1st, 8 AM)', value: '0 8 1 * *' },
]

function cronLabel(cron: string): string {
  return FREQ_OPTIONS.find((o) => o.value === cron)?.label ?? cron
}

// ── Widget data renderer ──────────────────────────────────────────────────────

function WidgetChart({ type, data }: { type: string; data: unknown }) {
  if (!data || typeof data !== 'object') {
    return <div className="text-xs text-gray-400 flex items-center justify-center h-full">No data</div>
  }

  const arr = Array.isArray(data) ? data : []

  if (type === 'deals_by_stage' || type === 'pipeline_funnel') {
    return (
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={arr} margin={{ top: 4, right: 8, left: -10, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={BORDER} />
          <XAxis dataKey="stage" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <Tooltip contentStyle={{ fontSize: 11 }} />
          <Bar dataKey="count" fill={PRIMARY} radius={[3, 3, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    )
  }

  if (type === 'contacts_monthly') {
    return (
      <ResponsiveContainer width="100%" height="100%">
        <RLineChart data={arr} margin={{ top: 4, right: 8, left: -10, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={BORDER} />
          <XAxis dataKey="month" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <Tooltip contentStyle={{ fontSize: 11 }} />
          <Line type="monotone" dataKey="count" stroke={PRIMARY} strokeWidth={2} dot={false} />
        </RLineChart>
      </ResponsiveContainer>
    )
  }

  if (type === 'activities_by_type' || type === 'conversion_rates') {
    return (
      <ResponsiveContainer width="100%" height="100%">
        <RPieChart>
          <Pie data={arr} dataKey="count" nameKey="type" cx="50%" cy="50%" outerRadius="70%">
            {arr.map((_: unknown, i: number) => (
              <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
            ))}
          </Pie>
          <Tooltip contentStyle={{ fontSize: 11 }} />
        </RPieChart>
      </ResponsiveContainer>
    )
  }

  if (type === 'revenue_projection') {
    return (
      <ResponsiveContainer width="100%" height="100%">
        <RLineChart data={arr} margin={{ top: 4, right: 8, left: -10, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={BORDER} />
          <XAxis dataKey="month" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <Tooltip contentStyle={{ fontSize: 11 }} />
          <Line type="monotone" dataKey="projected" stroke={PRIMARY} strokeWidth={2} dot={false} />
        </RLineChart>
      </ResponsiveContainer>
    )
  }

  // Metrics / summary widgets — render as stat grid
  if (typeof data === 'object' && !Array.isArray(data)) {
    const entries = Object.entries(data as Record<string, unknown>).slice(0, 4)
    return (
      <div className="grid grid-cols-2 gap-2 h-full content-start">
        {entries.map(([k, v]) => (
          <div key={k} className="rounded-md border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-2">
            <p className="text-[10px] font-semibold uppercase tracking-wide text-[#7C8DB0] truncate">{k.replace(/_/g, ' ')}</p>
            <p className="text-lg font-bold text-[#1A1D23] truncate">{String(v)}</p>
          </div>
        ))}
      </div>
    )
  }

  return <div className="text-xs text-gray-400 flex items-center justify-center h-full">Unsupported data format</div>
}

// ── Sortable Widget Card ──────────────────────────────────────────────────────

interface SortableWidgetCardProps {
  widget: Widget & { _id: string }
  runData?: unknown
  onRemove: (id: string) => void
}

function SortableWidgetCard({ widget, runData, onRemove }: SortableWidgetCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: widget._id,
  })
  const meta = widgetMeta(widget.type)
  const Icon = meta.icon

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={cn(
        'flex flex-col rounded-lg border border-[#E5E7EB] bg-white overflow-hidden',
        'min-h-[200px]',
        isDragging && 'shadow-xl opacity-80 z-50'
      )}
    >
      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-[#F0F1F3]">
        <button
          {...listeners}
          {...attributes}
          className="cursor-grab active:cursor-grabbing text-[#B0B8C1] hover:text-[#6B7280] touch-none"
          aria-label="Drag to reorder"
        >
          <GripVertical className="h-4 w-4" />
        </button>
        <Icon className="h-4 w-4 text-[#7C8DB0]" />
        <span className="text-sm font-semibold text-[#1A1D23] flex-1 truncate">{meta.label}</span>
        <button
          onClick={() => onRemove(widget._id)}
          className="text-[#B0B8C1] hover:text-red-500 transition-colors"
          aria-label="Remove widget"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
      {/* Content */}
      <div className="flex-1 p-3 min-h-0" style={{ height: 180 }}>
        {runData !== undefined ? (
          <WidgetChart type={widget.type} data={runData} />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-1 text-[#B0B8C1]">
            <Icon className="h-8 w-8 opacity-30" />
            <span className="text-xs">{meta.description}</span>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Widget Picker Modal ───────────────────────────────────────────────────────

interface WidgetPickerProps {
  open: boolean
  onClose: () => void
  onAdd: (type: string) => void
}

function WidgetPicker({ open, onClose, onAdd }: WidgetPickerProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogPortal>
        <DialogOverlay />
        <DialogContent className="max-w-lg w-full p-0 overflow-hidden">
          <div className="flex items-center justify-between px-5 py-4 border-b border-[#E5E7EB]">
            <h2 className="text-base font-semibold text-[#1A1D23]">Add Widget</h2>
            <DialogClose asChild>
              <button className="text-[#6B7280] hover:text-[#1A1D23]">
                <X className="h-4 w-4" />
              </button>
            </DialogClose>
          </div>
          <div className="grid grid-cols-2 gap-3 p-5 max-h-[60vh] overflow-y-auto">
            {WIDGET_CATALOGUE.map((meta) => {
              const Icon = meta.icon
              return (
                <button
                  key={meta.type}
                  onClick={() => { onAdd(meta.type); onClose() }}
                  className="flex items-start gap-3 rounded-lg border border-[#E5E7EB] p-3 text-left hover:border-[#1B3A4B] hover:bg-[#F7F8FA] transition-colors"
                >
                  <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[#F0F1F3]">
                    <Icon className="h-4 w-4 text-[#1B3A4B]" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-[#1A1D23]">{meta.label}</p>
                    <p className="text-xs text-[#6B7280] mt-0.5">{meta.description}</p>
                  </div>
                </button>
              )
            })}
          </div>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  )
}

// ── Schedule Delivery Modal ───────────────────────────────────────────────────

interface ScheduleModalProps {
  open: boolean
  onClose: () => void
  dashboardId: string
  existingSchedule?: ScheduledReport | null
}

function ScheduleModal({ open, onClose, dashboardId, existingSchedule }: ScheduleModalProps) {
  const [frequency, setFrequency] = useState(existingSchedule?.schedule ?? FREQ_OPTIONS[0].value)
  const [recipientsRaw, setRecipientsRaw] = useState(
    existingSchedule?.recipients.join(', ') ?? ''
  )
  const [freqOpen, setFreqOpen] = useState(false)

  const createSchedule = useCreateSchedule()
  const updateSchedule = useUpdateSchedule()
  const deleteSchedule = useDeleteSchedule()

  useEffect(() => {
    if (open) {
      setFrequency(existingSchedule?.schedule ?? FREQ_OPTIONS[0].value)
      setRecipientsRaw(existingSchedule?.recipients.join(', ') ?? '')
    }
  }, [open, existingSchedule])

  function handleSave() {
    const recipients = recipientsRaw
      .split(/[\s,]+/)
      .map((r) => r.trim())
      .filter(Boolean)
    if (!recipients.length) return

    if (existingSchedule) {
      updateSchedule.mutate(
        { id: existingSchedule.id, patch: { schedule: frequency, recipients } },
        { onSuccess: onClose }
      )
    } else {
      createSchedule.mutate(
        { dashboard_id: dashboardId, schedule: frequency, recipients },
        { onSuccess: onClose }
      )
    }
  }

  function handleDelete() {
    if (existingSchedule) {
      deleteSchedule.mutate(existingSchedule.id, { onSuccess: onClose })
    }
  }

  const saving = createSchedule.isPending || updateSchedule.isPending

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogPortal>
        <DialogOverlay />
        <DialogContent className="max-w-md w-full p-0 overflow-hidden">
          <div className="flex items-center justify-between px-5 py-4 border-b border-[#E5E7EB]">
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4 text-[#1B3A4B]" />
              <h2 className="text-base font-semibold text-[#1A1D23]">Schedule Delivery</h2>
            </div>
          </div>

          <div className="px-5 py-4 space-y-4">
            {/* Frequency picker */}
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-[#7C8DB0] mb-1.5">
                Frequency
              </label>
              <DropdownMenu open={freqOpen} onOpenChange={setFreqOpen}>
                <DropdownMenuTrigger asChild>
                  <button className="flex w-full items-center justify-between rounded-md border border-[#E5E7EB] bg-white px-3 py-2 text-sm text-[#1A1D23] hover:border-[#1B3A4B] transition-colors">
                    <span className="flex items-center gap-2">
                      <Clock className="h-3.5 w-3.5 text-[#7C8DB0]" />
                      {cronLabel(frequency)}
                    </span>
                    <ChevronDown className="h-3.5 w-3.5 text-[#6B7280]" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-56">
                  {FREQ_OPTIONS.map((opt) => (
                    <DropdownMenuItem
                      key={opt.value}
                      onClick={() => { setFrequency(opt.value); setFreqOpen(false) }}
                    >
                      {opt.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Recipients */}
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-[#7C8DB0] mb-1.5">
                Recipients
              </label>
              <textarea
                value={recipientsRaw}
                onChange={(e) => setRecipientsRaw(e.target.value)}
                rows={3}
                placeholder="alice@example.com, bob@example.com"
                className="flex w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-slate-400 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#1B3A4B] resize-none"
              />
              <p className="mt-1 text-[10px] text-[#6B7280]">Separate addresses with commas or spaces.</p>
            </div>
          </div>

          <div className="flex items-center justify-between px-5 py-3 border-t border-[#E5E7EB] bg-[#F7F8FA]">
            {existingSchedule ? (
              <button
                onClick={handleDelete}
                disabled={deleteSchedule.isPending}
                className="text-xs text-red-500 hover:text-red-700 font-medium"
              >
                Remove schedule
              </button>
            ) : <div />}
            <div className="flex gap-2">
              <Button variant="ghost" size="sm" onClick={onClose}>Cancel</Button>
              <Button
                size="sm"
                onClick={handleSave}
                disabled={saving || !recipientsRaw.trim()}
                style={{ background: PRIMARY, color: '#fff' }}
              >
                {saving && <Spinner className="mr-1.5 h-3 w-3" />}
                {existingSchedule ? 'Update' : 'Save'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

let _widgetCounter = 0
function newWidgetId() { return `widget-${++_widgetCounter}` }

type LocalWidget = Widget & { _id: string }

export function CustomDashboardsPage() {
  const { data: dashboards, isLoading: loadingList } = useDashboards()
  const { data: schedules } = useSchedules()
  const createDash = useCreateDashboard()
  const updateDash = useUpdateDashboard()
  const deleteDash = useDeleteDashboard()

  const [activeDashId, setActiveDashId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState(false)
  const [draftName, setDraftName] = useState('')
  const [widgets, setWidgets] = useState<LocalWidget[]>([])
  const [dirty, setDirty] = useState(false)
  const [showPicker, setShowPicker] = useState(false)
  const [showSchedule, setShowSchedule] = useState(false)
  const [showNewModal, setShowNewModal] = useState(false)
  const [newDashName, setNewDashName] = useState('')

  const activeDash = dashboards?.find((d) => d.id === activeDashId) ?? null

  const { data: runResult } = useRunDashboard(activeDashId)

  const scheduleForActive = schedules?.find((s) => s.dashboard_id === activeDashId) ?? null

  // When a dashboard is selected, load its widgets into local state
  useEffect(() => {
    if (activeDash) {
      setDraftName(activeDash.name)
      setWidgets(
        activeDash.widgets.map((w) => ({ ...w, _id: newWidgetId() }))
      )
      setDirty(false)
    }
  }, [activeDash?.id])

  // Auto-select first dashboard
  useEffect(() => {
    if (!activeDashId && dashboards?.length) {
      setActiveDashId(dashboards[0].id)
    }
  }, [dashboards])

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }))

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id) return
    setWidgets((prev) => {
      const oldIndex = prev.findIndex((w) => w._id === active.id)
      const newIndex = prev.findIndex((w) => w._id === over.id)
      return arrayMove(prev, oldIndex, newIndex).map((w, i) => ({
        ...w,
        position: { ...w.position, x: i % 2, y: Math.floor(i / 2) },
      }))
    })
    setDirty(true)
  }

  function addWidget(type: string) {
    const meta = widgetMeta(type)
    const count = widgets.length
    setWidgets((prev) => [
      ...prev,
      {
        _id: newWidgetId(),
        type,
        query_params: {},
        position: { x: count % 2, y: Math.floor(count / 2), ...meta.defaultSize },
      },
    ])
    setDirty(true)
  }

  function removeWidget(id: string) {
    setWidgets((prev) => prev.filter((w) => w._id !== id))
    setDirty(true)
  }

  function handleSave() {
    if (!activeDashId) return
    const patch: { name?: string; widgets?: Widget[] } = {}
    if (draftName !== activeDash?.name) patch.name = draftName
    patch.widgets = widgets.map(({ _id: _, ...w }) => w)
    updateDash.mutate(
      { id: activeDashId, patch },
      { onSuccess: () => setDirty(false) }
    )
  }

  function handleCreateDashboard() {
    if (!newDashName.trim()) return
    createDash.mutate(
      { name: newDashName.trim(), widgets: [] },
      {
        onSuccess: (d) => {
          setActiveDashId(d.id)
          setShowNewModal(false)
          setNewDashName('')
        },
      }
    )
  }

  function handleDeleteDash() {
    if (!activeDashId) return
    deleteDash.mutate(activeDashId, {
      onSuccess: () => {
        setActiveDashId(dashboards?.filter((d) => d.id !== activeDashId)[0]?.id ?? null)
      },
    })
  }

  function getRunData(widgetType: string): unknown {
    if (!runResult) return undefined
    const found = runResult.widgets.find((w) => w.widget.type === widgetType)
    return found?.data
  }

  const nameInputId = useId()

  return (
    <div className="flex h-full min-h-0" style={{ background: BG }}>
      {/* ── Dashboard list sidebar ── */}
      <aside className="w-56 shrink-0 border-r border-[#E5E7EB] bg-white flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b border-[#E5E7EB]">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-[#7C8DB0]">Dashboards</h2>
          <button
            onClick={() => setShowNewModal(true)}
            className="flex items-center justify-center h-6 w-6 rounded-md bg-[#F7F8FA] hover:bg-[#E8EDF2] text-[#6B7280] hover:text-[#1B3A4B] transition-colors"
            aria-label="New dashboard"
          >
            <Plus className="h-3.5 w-3.5" />
          </button>
        </div>
        <nav className="flex-1 overflow-y-auto py-1">
          {loadingList ? (
            <div className="flex justify-center py-6"><Spinner /></div>
          ) : !dashboards?.length ? (
            <p className="px-4 py-4 text-xs text-[#6B7280]">No dashboards yet.</p>
          ) : (
            dashboards.map((d) => (
              <button
                key={d.id}
                onClick={() => setActiveDashId(d.id)}
                className={cn(
                  'w-full text-left px-4 py-2 text-sm font-medium transition-colors truncate',
                  d.id === activeDashId
                    ? 'bg-[#E8EDF2] text-[#1B3A4B] font-semibold shadow-[inset_3px_0_0_#1B3A4B]'
                    : 'text-[#6B7280] hover:bg-[#F7F8FA] hover:text-[#1A1D23]'
                )}
              >
                {d.name}
              </button>
            ))
          )}
        </nav>
      </aside>

      {/* ── Main canvas ── */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Toolbar */}
        <div className="flex items-center gap-3 px-5 py-3 border-b border-[#E5E7EB] bg-white shrink-0">
          {/* Editable name */}
          {activeDash ? (
            editingName ? (
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <Input
                  id={nameInputId}
                  value={draftName}
                  onChange={(e) => { setDraftName(e.target.value); setDirty(true) }}
                  onBlur={() => setEditingName(false)}
                  onKeyDown={(e) => { if (e.key === 'Enter') setEditingName(false) }}
                  autoFocus
                  className="h-8 max-w-xs text-sm font-semibold"
                />
              </div>
            ) : (
              <button
                className="flex items-center gap-1.5 group flex-1 min-w-0"
                onClick={() => setEditingName(true)}
                aria-label="Rename dashboard"
              >
                <h1 className="text-base font-bold text-[#1A1D23] truncate">{draftName}</h1>
                <Edit2 className="h-3.5 w-3.5 text-[#B0B8C1] group-hover:text-[#6B7280] shrink-0" />
              </button>
            )
          ) : (
            <h1 className="text-base font-bold text-[#1A1D23] flex-1">Custom Dashboards</h1>
          )}

          <div className="flex items-center gap-2 shrink-0">
            {activeDash && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowSchedule(true)}
                  className="flex items-center gap-1.5 text-xs"
                >
                  <Calendar className="h-3.5 w-3.5" />
                  {scheduleForActive ? 'Edit Schedule' : 'Schedule'}
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowPicker(true)}
                  className="flex items-center gap-1.5 text-xs"
                >
                  <Plus className="h-3.5 w-3.5" />
                  Add Widget
                </Button>

                {dirty && (
                  <Button
                    size="sm"
                    onClick={handleSave}
                    disabled={updateDash.isPending}
                    className="flex items-center gap-1.5 text-xs"
                    style={{ background: PRIMARY, color: '#fff' }}
                  >
                    {updateDash.isPending ? <Spinner className="h-3 w-3" /> : <Save className="h-3.5 w-3.5" />}
                    Save
                  </Button>
                )}

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button className="flex items-center justify-center h-8 w-8 rounded-md hover:bg-[#F7F8FA] text-[#6B7280] hover:text-[#1A1D23] transition-colors">
                      <MoreVertical className="h-4 w-4" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem
                      className="text-red-600 focus:text-red-600 focus:bg-red-50"
                      onClick={handleDeleteDash}
                    >
                      <Trash2 className="mr-2 h-3.5 w-3.5" />
                      Delete Dashboard
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </>
            )}
          </div>
        </div>

        {/* Schedule badge */}
        {scheduleForActive && (
          <div className="flex items-center gap-2 px-5 py-1.5 bg-[#EFF6FF] border-b border-[#DBEAFE] text-xs text-[#1D4ED8]">
            <Clock className="h-3.5 w-3.5" />
            Scheduled: {cronLabel(scheduleForActive.schedule)} → {scheduleForActive.recipients.join(', ')}
          </div>
        )}

        {/* Canvas */}
        <div className="flex-1 overflow-auto p-5">
          {!activeDash ? (
            <div className="flex flex-col items-center justify-center h-full gap-4 text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-[#E8EDF2]">
                <BarChart2 className="h-7 w-7 text-[#1B3A4B]" />
              </div>
              <div>
                <p className="text-base font-semibold text-[#1A1D23]">No dashboard selected</p>
                <p className="text-sm text-[#6B7280] mt-1">Create a new dashboard or select one from the list.</p>
              </div>
              <Button
                onClick={() => setShowNewModal(true)}
                style={{ background: PRIMARY, color: '#fff' }}
                className="flex items-center gap-2"
              >
                <Plus className="h-4 w-4" />
                New Dashboard
              </Button>
            </div>
          ) : widgets.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full gap-4 text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-[#E8EDF2]">
                <Plus className="h-7 w-7 text-[#1B3A4B]" />
              </div>
              <div>
                <p className="text-base font-semibold text-[#1A1D23]">Empty dashboard</p>
                <p className="text-sm text-[#6B7280] mt-1">Add widgets to start building your view.</p>
              </div>
              <Button
                onClick={() => setShowPicker(true)}
                style={{ background: PRIMARY, color: '#fff' }}
                className="flex items-center gap-2"
              >
                <Plus className="h-4 w-4" />
                Add Widget
              </Button>
            </div>
          ) : (
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext items={widgets.map((w) => w._id)} strategy={rectSortingStrategy}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 auto-rows-auto">
                  {widgets.map((w) => (
                    <SortableWidgetCard
                      key={w._id}
                      widget={w}
                      runData={runResult ? getRunData(w.type) : undefined}
                      onRemove={removeWidget}
                    />
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          )}
        </div>
      </div>

      {/* ── Modals ── */}
      <WidgetPicker open={showPicker} onClose={() => setShowPicker(false)} onAdd={addWidget} />

      {activeDashId && (
        <ScheduleModal
          open={showSchedule}
          onClose={() => setShowSchedule(false)}
          dashboardId={activeDashId}
          existingSchedule={scheduleForActive}
        />
      )}

      {/* New Dashboard modal */}
      <Dialog open={showNewModal} onOpenChange={(v) => !v && setShowNewModal(false)}>
        <DialogPortal>
          <DialogOverlay />
          <DialogContent className="max-w-sm w-full p-0 overflow-hidden">
            <div className="flex items-center justify-between px-5 py-4 border-b border-[#E5E7EB]">
              <h2 className="text-base font-semibold text-[#1A1D23]">New Dashboard</h2>
              <DialogClose asChild>
                <button className="text-[#6B7280] hover:text-[#1A1D23]">
                  <X className="h-4 w-4" />
                </button>
              </DialogClose>
            </div>
            <div className="px-5 py-4">
              <label className="block text-xs font-semibold uppercase tracking-wide text-[#7C8DB0] mb-1.5">
                Dashboard Name
              </label>
              <Input
                value={newDashName}
                onChange={(e) => setNewDashName(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleCreateDashboard() }}
                placeholder="e.g. Sales Overview"
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-2 px-5 py-3 border-t border-[#E5E7EB] bg-[#F7F8FA]">
              <Button variant="ghost" size="sm" onClick={() => setShowNewModal(false)}>Cancel</Button>
              <Button
                size="sm"
                onClick={handleCreateDashboard}
                disabled={!newDashName.trim() || createDash.isPending}
                style={{ background: PRIMARY, color: '#fff' }}
              >
                {createDash.isPending && <Spinner className="mr-1.5 h-3 w-3" />}
                Create
              </Button>
            </div>
          </DialogContent>
        </DialogPortal>
      </Dialog>
    </div>
  )
}
