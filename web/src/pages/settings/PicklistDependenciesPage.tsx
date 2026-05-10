import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { picklistsApi } from '@/api/picklists'
import type { CustomFieldEntityType } from '@/api/types'

const ENTITY_TYPES: CustomFieldEntityType[] = ['lead', 'contact', 'account', 'deal', 'ticket']

export function PicklistDependenciesPage() {
  const qc = useQueryClient()
  const [entityType, setEntityType] = useState<CustomFieldEntityType>('lead')
  const [sourceFieldID, setSourceFieldID] = useState('')
  const [targetFieldID, setTargetFieldID] = useState('')
  const [mappingJSON, setMappingJSON] = useState('{\n  "source_value": ["target_value"]\n}')
  const [status, setStatus] = useState('')

  const fieldsQuery = useQuery({
    queryKey: ['settings', 'picklist-fields', entityType],
    queryFn: () => picklistsApi.listFields(entityType),
  })

  const dependenciesQuery = useQuery({
    queryKey: ['settings', 'picklist-dependencies', entityType],
    queryFn: () => picklistsApi.listDependencies(entityType),
  })

  const fieldOptions = useMemo(() => fieldsQuery.data?.map((bundle) => bundle.field) ?? [], [fieldsQuery.data])

  const upsertMutation = useMutation({
    mutationFn: async () => {
      const parsed = JSON.parse(mappingJSON) as Record<string, string[]>
      await picklistsApi.upsertDependency({
        entity_type: entityType,
        source_field_id: sourceFieldID,
        target_field_id: targetFieldID,
        mapping: parsed,
        is_active: true,
      })
    },
    onSuccess: async () => {
      setStatus('Dependency saved.')
      await qc.invalidateQueries({ queryKey: ['settings', 'picklist-dependencies', entityType] })
    },
    onError: () => setStatus('Unable to save dependency. Confirm mapping JSON is valid.'),
  })

  const deleteMutation = useMutation({
    mutationFn: picklistsApi.deleteDependency,
    onSuccess: async () => {
      setStatus('Dependency deleted.')
      await qc.invalidateQueries({ queryKey: ['settings', 'picklist-dependencies', entityType] })
    },
    onError: () => setStatus('Unable to delete dependency.'),
  })

  const pending = fieldsQuery.isLoading || dependenciesQuery.isLoading || upsertMutation.isPending || deleteMutation.isPending

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Picklist Dependencies</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Restrict target picklist options based on source field values.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm space-y-4">
        <div className="grid gap-3 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">
              Entity
            </label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={entityType}
              onChange={(e) => {
                setEntityType(e.target.value as CustomFieldEntityType)
                setSourceFieldID('')
                setTargetFieldID('')
                setStatus('')
              }}
              disabled={pending}
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
              Source Field
            </label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={sourceFieldID}
              onChange={(e) => setSourceFieldID(e.target.value)}
              disabled={pending}
            >
              <option value="">Select source</option>
              {fieldOptions.map((field) => (
                <option key={field.id} value={field.id}>
                  {field.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">
              Target Field
            </label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={targetFieldID}
              onChange={(e) => setTargetFieldID(e.target.value)}
              disabled={pending}
            >
              <option value="">Select target</option>
              {fieldOptions
                .filter((field) => field.id !== sourceFieldID)
                .map((field) => (
                  <option key={field.id} value={field.id}>
                    {field.label}
                  </option>
                ))}
            </select>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">
            Mapping (JSON)
          </label>
          <textarea
            className="min-h-[160px] w-full rounded-md border border-slate-200 bg-white px-3 py-2 font-mono text-xs"
            value={mappingJSON}
            onChange={(e) => setMappingJSON(e.target.value)}
            disabled={pending}
          />
        </div>

        <div className="flex justify-end">
          <Button
            onClick={() => upsertMutation.mutate()}
            disabled={pending || !sourceFieldID || !targetFieldID}
          >
            <Save className="h-4 w-4" />
            Save Dependency
          </Button>
        </div>

        <div className="space-y-2">
          <p className="text-sm font-semibold text-slate-700">Configured Dependencies</p>
          {(dependenciesQuery.data ?? []).length === 0 && (
            <p className="text-sm text-slate-500">No dependencies configured for this entity.</p>
          )}
          {(dependenciesQuery.data ?? []).map((dependency) => (
            <div key={dependency.id} className="flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2">
              <div className="text-sm text-slate-700">
                <p className="font-medium">
                  {dependency.source_field_id} {'->'} {dependency.target_field_id}
                </p>
                <pre className="mt-1 overflow-auto text-xs text-slate-500">{JSON.stringify(dependency.mapping, null, 2)}</pre>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => deleteMutation.mutate(dependency.id)}
                disabled={pending}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>

        {status && (
          <p className={status.toLowerCase().includes('unable') ? 'text-sm text-red-600' : 'text-sm text-green-700'}>
            {status}
          </p>
        )}
      </div>
    </div>
  )
}
