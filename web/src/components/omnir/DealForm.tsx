import { useMemo, useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useCreateDeal } from '@/hooks/useDeals'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { useModuleLayout } from '@/hooks/useModuleConfiguration'
import { CustomFieldFormSection } from '@/components/omnir/CustomFieldRenderer'
import { applyLayoutToCustomFields, isLayoutFieldVisible, layoutFieldLabel } from '@/lib/moduleConfiguration'
import type { CreateDealRequest, CustomFieldValues, Deal, DealStage } from '@/api/types'

const STAGE_OPTIONS: { label: string; value: DealStage }[] = [
  { label: 'Lead', value: 'lead' },
  { label: 'Qualified', value: 'qualified' },
  { label: 'Proposal', value: 'proposal' },
  { label: 'Negotiation', value: 'negotiation' },
  { label: 'Closed Won', value: 'closed_won' },
  { label: 'Closed Lost', value: 'closed_lost' },
]

interface DealFormProps {
  onClose: () => void
  initialValues?: Partial<CreateDealRequest>
  onCreated?: (deal: Deal) => void
}

export function DealForm({ onClose, initialValues, onCreated }: DealFormProps) {
  const { mutateAsync: createDeal, isPending } = useCreateDeal()
  const { data: customFields = [] } = useCustomFieldDefinitions('deal', { activeOptionsOnly: true })
  const { data: layout } = useModuleLayout('deal')
  const layoutCustomFields = useMemo(
    () => applyLayoutToCustomFields(customFields, layout, 'quick_create'),
    [customFields, layout]
  )
  const [title, setTitle] = useState(initialValues?.title ?? '')
  const [value, setValue] = useState(
    typeof initialValues?.value_cents === 'number' ? (initialValues.value_cents / 100).toFixed(2) : ''
  )
  const [stage, setStage] = useState<DealStage>(initialValues?.stage ?? 'lead')
  const [closeDate, setCloseDate] = useState(initialValues?.expected_close_date ?? '')
  const [customFieldValues, setCustomFieldValues] = useState<CustomFieldValues>({})
  const [errors, setErrors] = useState<{ title?: string; value?: string }>({})
  const visible = (fieldKey: string) => isLayoutFieldVisible(layout, 'standard', fieldKey, 'quick_create')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const errs: typeof errors = {}
    if (!title.trim()) errs.title = 'Required'
    const numValue = parseFloat(value)
    if (visible('value_cents') && (!value || isNaN(numValue) || numValue < 0)) errs.value = 'Enter a valid amount'
    setErrors(errs)
    if (Object.keys(errs).length > 0) return

    const payload: CreateDealRequest = {
      title: title.trim(),
      value_cents: visible('value_cents') ? Math.round(numValue * 100) : 0,
      stage,
      ...(initialValues?.account_id ? { account_id: initialValues.account_id } : {}),
      ...(initialValues?.contact_id ? { contact_id: initialValues.contact_id } : {}),
      ...(initialValues?.pipeline_id ? { pipeline_id: initialValues.pipeline_id } : {}),
      ...(initialValues?.owner_id ? { owner_id: initialValues.owner_id } : {}),
      ...(initialValues?.currency ? { currency: initialValues.currency } : {}),
      ...(closeDate ? { expected_close_date: closeDate } : {}),
      ...(Object.keys(customFieldValues).length ? { custom_fields: customFieldValues } : {}),
    }
    const deal = await createDeal(payload)
    onCreated?.(deal)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-xl bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-slate-200 px-4 sm:px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">New Deal</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 p-4 sm:p-6">
          <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="deal-title">
              {layoutFieldLabel(layout, 'standard', 'title', 'Deal title')} <span className="text-red-500">*</span>
            </label>
            <input
              id="deal-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Acme Corp — Enterprise plan"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
            />
            {errors.title && <p className="text-xs text-red-600">{errors.title}</p>}
          </div>

          {(visible('value_cents') || visible('stage')) && <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {visible('value_cents') && <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="deal-value">
                Value ($) <span className="text-red-500">*</span>
              </label>
              <input
                id="deal-value"
                type="number"
                min="0"
                step="0.01"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="0.00"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
              {errors.value && <p className="text-xs text-red-600">{errors.value}</p>}
            </div>}

            {visible('stage') && <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="deal-stage">
                Stage
              </label>
              <select
                id="deal-stage"
                value={stage}
                onChange={(e) => setStage(e.target.value as DealStage)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              >
                {STAGE_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </select>
            </div>}
          </div>}

          {visible('expected_close_date') && <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="deal-close-date">
              Expected close date
            </label>
            <input
              id="deal-close-date"
              type="date"
              value={closeDate}
              onChange={(e) => setCloseDate(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
            />
          </div>}

          <CustomFieldFormSection
            fields={layoutCustomFields}
            values={customFieldValues}
            onChange={setCustomFieldValues}
          />

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create Deal'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
