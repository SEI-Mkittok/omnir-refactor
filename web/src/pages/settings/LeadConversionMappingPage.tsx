import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  leadConversionMappingApi,
  type LeadConversionMapping,
  type LeadConversionTarget,
} from '@/api/leadConversionMapping'

const TARGETS: LeadConversionTarget[] = ['contact', 'account', 'deal']

function blankMapping(): LeadConversionMapping {
  return {
    lead_field: '',
    target_entity: 'contact',
    target_field: '',
    is_active: true,
  }
}

export function LeadConversionMappingPage() {
  const qc = useQueryClient()
  const [rows, setRows] = useState<LeadConversionMapping[]>([])
  const [status, setStatus] = useState('')

  const query = useQuery({
    queryKey: ['settings', 'lead-conversion-mappings'],
    queryFn: leadConversionMappingApi.list,
  })

  useEffect(() => {
    if (!query.data) return
    setRows(query.data.length > 0 ? query.data : [blankMapping()])
    setStatus('')
  }, [query.data])

  const mutation = useMutation({
    mutationFn: async () => {
      const payload = rows
        .map((row) => ({
          lead_field: row.lead_field.trim(),
          target_entity: row.target_entity,
          target_field: row.target_field.trim(),
          is_active: row.is_active,
        }))
        .filter((row) => row.lead_field !== '' && row.target_field !== '')
      return leadConversionMappingApi.replace(payload)
    },
    onSuccess: async () => {
      setStatus('Lead conversion mappings saved.')
      await qc.invalidateQueries({ queryKey: ['settings', 'lead-conversion-mappings'] })
    },
    onError: () => setStatus('Unable to save mappings. Verify field compatibility and try again.'),
  })

  function addRow() {
    setRows((current) => [...current, blankMapping()])
    setStatus('')
  }

  function updateRow(index: number, patch: Partial<LeadConversionMapping>) {
    setRows((current) => current.map((row, i) => (i === index ? { ...row, ...patch } : row)))
    setStatus('')
  }

  function removeRow(index: number) {
    setRows((current) => {
      const next = current.filter((_, i) => i !== index)
      return next.length > 0 ? next : [blankMapping()]
    })
    setStatus('')
  }

  const loading = query.isLoading || mutation.isPending

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Lead Conversion Mapping</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Map lead fields to contact/account/deal fields used during conversion.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm space-y-3">
        {query.isError && (
          <p className="text-sm text-red-600">Failed to load mappings.</p>
        )}

        {rows.map((row, index) => (
          <div key={`${row.id ?? 'new'}-${index}`} className="grid gap-2 lg:grid-cols-[1fr_140px_1fr_110px_auto]">
            <Input
              value={row.lead_field}
              placeholder="lead_field or custom:field_name"
              onChange={(e) => updateRow(index, { lead_field: e.target.value })}
              disabled={loading}
            />
            <select
              className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={row.target_entity}
              onChange={(e) => updateRow(index, { target_entity: e.target.value as LeadConversionTarget })}
              disabled={loading}
            >
              {TARGETS.map((target) => (
                <option key={target} value={target}>
                  {target}
                </option>
              ))}
            </select>
            <Input
              value={row.target_field}
              placeholder="target_field or custom:field_name"
              onChange={(e) => updateRow(index, { target_field: e.target.value })}
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
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => removeRow(index)}
              disabled={loading}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}

        <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
          <Button type="button" variant="outline" onClick={addRow} disabled={loading}>
            <Plus className="h-4 w-4" />
            Add Mapping
          </Button>
          <Button type="button" onClick={() => mutation.mutate()} disabled={loading}>
            <Save className="h-4 w-4" />
            {mutation.isPending ? 'Saving...' : 'Save Mappings'}
          </Button>
        </div>

        {status && (
          <p className={status.toLowerCase().includes('unable') ? 'text-sm text-red-600' : 'text-sm text-green-700'}>
            {status}
          </p>
        )}

        <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs text-slate-600">
          Supported defaults:
          <br />
          lead fields: first_name, last_name, email, phone, company, lead_source, lead_score, status, custom:*
          <br />
          contact fields: first_name, last_name, email, phone, lead_source, custom:*
          <br />
          account fields: name, domain, custom:*
          <br />
          deal fields: title, value_cents, currency, custom:*
        </div>
      </div>
    </div>
  )
}
