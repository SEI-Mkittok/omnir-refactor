import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowDown, ArrowUp, Plus, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { picklistsApi, type PicklistFieldBundle, type PicklistValue } from '@/api/picklists'
import type { CustomFieldEntityType } from '@/api/types'

type PicklistValueDraft = PicklistValue

const ENTITY_TYPES: CustomFieldEntityType[] = ['lead', 'contact', 'account', 'deal', 'ticket']

function toDraft(values: PicklistValue[]): PicklistValueDraft[] {
  return values.map((row) => ({ ...row }))
}

export function PicklistsPage() {
  const qc = useQueryClient()
  const [entityType, setEntityType] = useState<CustomFieldEntityType>('lead')
  const [selectedFieldID, setSelectedFieldID] = useState('')
  const [rows, setRows] = useState<PicklistValueDraft[]>([])
  const [remapFrom, setRemapFrom] = useState('')
  const [remapTo, setRemapTo] = useState('')
  const [status, setStatus] = useState('')

  const fieldsQuery = useQuery({
    queryKey: ['settings', 'picklists', entityType],
    queryFn: () => picklistsApi.listFields(entityType),
  })

  const selectedBundle: PicklistFieldBundle | undefined = useMemo(
    () => fieldsQuery.data?.find((bundle) => bundle.field.id === selectedFieldID),
    [fieldsQuery.data, selectedFieldID]
  )

  useEffect(() => {
    if (!fieldsQuery.data || fieldsQuery.data.length === 0) {
      setSelectedFieldID('')
      setRows([])
      return
    }
    if (!selectedFieldID) {
      setSelectedFieldID(fieldsQuery.data[0].field.id)
      return
    }
    const current = fieldsQuery.data.find((bundle) => bundle.field.id === selectedFieldID)
    if (!current) {
      setSelectedFieldID(fieldsQuery.data[0].field.id)
    }
  }, [fieldsQuery.data, selectedFieldID])

  useEffect(() => {
    if (selectedBundle) {
      setRows(toDraft(selectedBundle.values))
      setStatus('')
    }
  }, [selectedBundle])

  const saveMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFieldID) return
      await picklistsApi.upsertValues(
        selectedFieldID,
        rows.map((row, index) => ({
          value: row.value.trim(),
          display_label: row.display_label.trim() || row.value.trim(),
          order_idx: index,
          is_active: row.is_active,
        }))
      )
    },
    onSuccess: async () => {
      setStatus('Picklist values saved.')
      await qc.invalidateQueries({ queryKey: ['settings', 'picklists', entityType] })
    },
    onError: () => setStatus('Unable to save picklist values.'),
  })

  const remapMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFieldID || !remapFrom.trim()) return
      const next = await picklistsApi.remapDelete(selectedFieldID, remapFrom.trim(), remapTo.trim() || undefined)
      return next
    },
    onSuccess: async () => {
      setStatus('Picklist value remapped and deleted.')
      setRemapFrom('')
      setRemapTo('')
      await qc.invalidateQueries({ queryKey: ['settings', 'picklists', entityType] })
    },
    onError: () => setStatus('Unable to remap/delete value.'),
  })

  function updateRow(index: number, patch: Partial<PicklistValueDraft>) {
    setRows((current) => current.map((row, i) => (i === index ? { ...row, ...patch } : row)))
    setStatus('')
  }

  function move(index: number, direction: -1 | 1) {
    setRows((current) => {
      const next = [...current]
      const target = index + direction
      if (target < 0 || target >= next.length) return next
      const [item] = next.splice(index, 1)
      next.splice(target, 0, item)
      return next
    })
    setStatus('')
  }

  function addValue() {
    setRows((current) => [
      ...current,
      {
        id: `new-${Date.now()}`,
        org_id: '',
        custom_field_id: selectedFieldID,
        value: '',
        display_label: '',
        order_idx: current.length,
        is_active: true,
        created_at: '',
        updated_at: '',
      },
    ])
    setStatus('')
  }

  function removeAt(index: number) {
    setRows((current) => current.filter((_, i) => i !== index))
    setStatus('')
  }

  const loading = fieldsQuery.isLoading || saveMutation.isPending || remapMutation.isPending

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Picklists</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Manage picklist values, ordering, and safe remap/delete workflows for custom selectable fields.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm space-y-4">
        <div className="grid gap-3 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">
              Entity
            </label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={entityType}
              onChange={(e) => setEntityType(e.target.value as CustomFieldEntityType)}
              disabled={loading}
            >
              {ENTITY_TYPES.map((entity) => (
                <option key={entity} value={entity}>
                  {entity}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">
              Field
            </label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={selectedFieldID}
              onChange={(e) => setSelectedFieldID(e.target.value)}
              disabled={loading || !fieldsQuery.data?.length}
            >
              {(fieldsQuery.data ?? []).map((bundle) => (
                <option key={bundle.field.id} value={bundle.field.id}>
                  {bundle.field.label} ({bundle.field.name})
                </option>
              ))}
            </select>
          </div>
        </div>

        {!selectedFieldID && !fieldsQuery.isLoading && (
          <p className="text-sm text-slate-500">
            No select/multiselect custom fields found for this entity.
          </p>
        )}

        {selectedFieldID && (
          <div className="space-y-3">
            {rows.map((row, index) => (
              <div key={row.id} className="grid gap-2 lg:grid-cols-[1fr_1fr_130px_auto_auto_auto]">
                <Input
                  value={row.value}
                  placeholder="value_key"
                  onChange={(e) => updateRow(index, { value: e.target.value })}
                  disabled={loading}
                />
                <Input
                  value={row.display_label}
                  placeholder="Display label"
                  onChange={(e) => updateRow(index, { display_label: e.target.value })}
                  disabled={loading}
                />
                <label className="inline-flex items-center gap-2 text-sm text-slate-600">
                  <input
                    type="checkbox"
                    checked={row.is_active}
                    onChange={(e) => updateRow(index, { is_active: e.target.checked })}
                    disabled={loading}
                  />
                  Active
                </label>
                <Button type="button" variant="outline" size="sm" onClick={() => move(index, -1)} disabled={loading || index === 0}>
                  <ArrowUp className="h-4 w-4" />
                </Button>
                <Button type="button" variant="outline" size="sm" onClick={() => move(index, 1)} disabled={loading || index === rows.length - 1}>
                  <ArrowDown className="h-4 w-4" />
                </Button>
                <Button type="button" variant="outline" size="sm" onClick={() => removeAt(index)} disabled={loading}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}

            <div className="flex flex-wrap items-center justify-between gap-2">
              <Button type="button" variant="outline" onClick={addValue} disabled={loading}>
                <Plus className="h-4 w-4" />
                Add Value
              </Button>
              <Button type="button" onClick={() => saveMutation.mutate()} disabled={loading || !selectedFieldID}>
                <Save className="h-4 w-4" />
                {saveMutation.isPending ? 'Saving...' : 'Save Values'}
              </Button>
            </div>

            <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 space-y-2">
              <p className="text-sm font-semibold text-slate-700">Remap Then Delete</p>
              <p className="text-xs text-slate-500">
                If a value is already used, remap records to another value before delete.
              </p>
              <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto]">
                <Input
                  value={remapFrom}
                  placeholder="From value"
                  onChange={(e) => setRemapFrom(e.target.value)}
                  disabled={loading}
                />
                <Input
                  value={remapTo}
                  placeholder="To value (optional if unused)"
                  onChange={(e) => setRemapTo(e.target.value)}
                  disabled={loading}
                />
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => remapMutation.mutate()}
                  disabled={loading || !remapFrom.trim() || !selectedFieldID}
                >
                  Apply
                </Button>
              </div>
            </div>
          </div>
        )}

        {status && (
          <p className={status.toLowerCase().includes('unable') ? 'text-sm text-red-600' : 'text-sm text-green-700'}>
            {status}
          </p>
        )}
      </div>
    </div>
  )
}
