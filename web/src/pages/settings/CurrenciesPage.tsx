import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { currenciesApi, type OrgCurrencyInput } from '@/api/currencies'

interface CurrencyDraft extends OrgCurrencyInput {
  is_default: boolean
}

function toDrafts(
  rows: { code: string; display_name: string; symbol: string; decimal_places: number; is_active: boolean; is_default: boolean }[],
  defaultCode: string
): CurrencyDraft[] {
  if (rows.length === 0) {
    return [{ code: 'USD', display_name: 'US Dollar', symbol: '$', decimal_places: 2, is_active: true, is_default: true }]
  }
  return rows.map((row) => ({
    code: row.code,
    display_name: row.display_name,
    symbol: row.symbol,
    decimal_places: row.decimal_places,
    is_active: row.is_active,
    is_default: row.code.toUpperCase() === defaultCode.toUpperCase(),
  }))
}

export function CurrenciesPage() {
  const qc = useQueryClient()
  const [drafts, setDrafts] = useState<CurrencyDraft[]>([])
  const [message, setMessage] = useState('')

  const query = useQuery({
    queryKey: ['settings', 'currencies'],
    queryFn: currenciesApi.get,
  })

  useEffect(() => {
    if (!query.data) {
      return
    }
    setDrafts(toDrafts(query.data.currencies, query.data.default_code))
    setMessage('')
  }, [query.data])

  const mutation = useMutation({
    mutationFn: currenciesApi.update,
    onSuccess: () => {
      setMessage('Currency settings saved.')
      qc.invalidateQueries({ queryKey: ['settings', 'currencies'] })
    },
    onError: () => setMessage('Unable to save currency settings.'),
  })

  function addRow() {
    setDrafts((current) => [
      ...current,
      {
        code: '',
        display_name: '',
        symbol: '',
        decimal_places: 2,
        is_active: true,
        is_default: current.length === 0,
      },
    ])
  }

  function removeRow(index: number) {
    setDrafts((current) => {
      const next = current.filter((_, i) => i !== index)
      if (next.length === 0) {
        return [{ code: 'USD', display_name: 'US Dollar', symbol: '$', decimal_places: 2, is_active: true, is_default: true }]
      }
      if (!next.some((item) => item.is_default)) {
        next[0].is_default = true
      }
      return [...next]
    })
  }

  function updateRow(index: number, patch: Partial<CurrencyDraft>) {
    setDrafts((current) => current.map((row, i) => (i === index ? { ...row, ...patch } : row)))
    setMessage('')
  }

  function markDefault(index: number) {
    setDrafts((current) =>
      current.map((row, i) => ({
        ...row,
        is_default: i === index,
        is_active: i === index ? true : row.is_active,
      }))
    )
    setMessage('')
  }

  function handleSave() {
    const normalized = drafts
      .map((row) => ({
        ...row,
        code: row.code.trim().toUpperCase(),
        display_name: row.display_name.trim(),
        symbol: row.symbol.trim(),
      }))
      .filter((row) => row.code !== '')

    if (normalized.length === 0) {
      setMessage('Add at least one currency code.')
      return
    }

    const defaultCurrency = normalized.find((row) => row.is_default) ?? normalized[0]
    mutation.mutate({
      default_code: defaultCurrency.code,
      currencies: normalized.map(({ is_default, ...input }) => input),
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Currencies</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Configure enabled currencies and select the organization default used across pricing forms.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        {query.isError && (
          <p role="alert" className="mb-4 text-sm text-red-600">
            Failed to load currencies.
          </p>
        )}

        <div className="space-y-3">
          {drafts.map((row, index) => (
            <div key={`${index}-${row.code}`} className="grid gap-2 lg:grid-cols-[80px_180px_90px_120px_120px_auto]">
              <Input
                value={row.code}
                placeholder="USD"
                maxLength={10}
                onChange={(e) => updateRow(index, { code: e.target.value })}
                disabled={query.isLoading || mutation.isPending}
              />
              <Input
                value={row.display_name}
                placeholder="US Dollar"
                onChange={(e) => updateRow(index, { display_name: e.target.value })}
                disabled={query.isLoading || mutation.isPending}
              />
              <Input
                value={row.symbol}
                placeholder="$"
                onChange={(e) => updateRow(index, { symbol: e.target.value })}
                disabled={query.isLoading || mutation.isPending}
              />
              <Input
                type="number"
                min={0}
                max={6}
                value={row.decimal_places}
                onChange={(e) => updateRow(index, { decimal_places: Number(e.target.value || 2) })}
                disabled={query.isLoading || mutation.isPending}
              />
              <div className="flex items-center gap-3 text-xs text-slate-600">
                <label className="inline-flex items-center gap-1">
                  <input
                    type="checkbox"
                    checked={row.is_active}
                    onChange={(e) => updateRow(index, { is_active: e.target.checked })}
                    disabled={mutation.isPending}
                  />
                  Active
                </label>
                <label className="inline-flex items-center gap-1">
                  <input
                    type="radio"
                    name="default-currency"
                    checked={row.is_default}
                    onChange={() => markDefault(index)}
                    disabled={mutation.isPending}
                  />
                  Default
                </label>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => removeRow(index)}
                disabled={mutation.isPending}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <Button type="button" variant="outline" onClick={addRow} disabled={mutation.isPending}>
            <Plus className="h-4 w-4" />
            Add Currency
          </Button>
          <div className="flex items-center gap-3">
            {message && (
              <p className={mutation.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>
                {message}
              </p>
            )}
            <Button onClick={handleSave} disabled={query.isLoading || mutation.isPending}>
              <Save className="h-4 w-4" />
              {mutation.isPending ? 'Saving...' : 'Save Currencies'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
