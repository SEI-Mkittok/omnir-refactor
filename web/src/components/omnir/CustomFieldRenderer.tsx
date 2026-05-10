import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { picklistsApi, type PicklistDependency } from '@/api/picklists'
import type { CustomFieldDefinition, CustomFieldValues } from '@/api/types'

// ── Input renderer ───────────────────────────────────────────────────────────

interface CustomFieldInputProps {
  field: CustomFieldDefinition
  value: CustomFieldValues[string]
  options?: string[]
  onChange: (fieldKey: string, value: CustomFieldValues[string]) => void
}

function CustomFieldInput({ field, value, options, onChange }: CustomFieldInputProps) {
  const inputClass =
    'flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]'

  switch (field.field_type) {
    case 'text':
    case 'url':
      return (
        <Input
          type={field.field_type === 'url' ? 'url' : 'text'}
          value={(value as string) ?? ''}
          onChange={(e) => onChange(field.name, e.target.value || null)}
          placeholder={field.label}
        />
      )

    case 'number':
      return (
        <Input
          type="number"
          value={(value as number) ?? ''}
          onChange={(e) =>
            onChange(field.name, e.target.value === '' ? null : Number(e.target.value))
          }
          placeholder={field.label}
        />
      )

    case 'date':
      return (
        <Input
          type="date"
          value={(value as string) ?? ''}
          onChange={(e) => onChange(field.name, e.target.value || null)}
        />
      )

    case 'checkbox':
      return (
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id={`cf-${field.id}`}
            checked={!!(value as boolean)}
            onChange={(e) => onChange(field.name, e.target.checked)}
            className="h-4 w-4 rounded border-slate-300 text-[var(--color-primary)] focus:ring-[var(--border-focus)]"
          />
          <label htmlFor={`cf-${field.id}`} className="text-sm text-slate-700">
            {field.label}
          </label>
        </div>
      )

    case 'select':
      return (
        <select
          className={inputClass}
          value={(value as string) ?? ''}
          onChange={(e) => onChange(field.name, e.target.value || null)}
        >
          <option value="">— Select —</option>
          {(options ?? field.options ?? []).map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
      )

    case 'multiselect': {
      const selected = (value as string[]) ?? []
      return (
        <div className="flex flex-wrap gap-1.5">
          {(options ?? field.options ?? []).map((opt) => {
            const checked = selected.includes(opt)
            return (
              <label
                key={opt}
                className={`flex cursor-pointer items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium border transition-colors ${
                  checked
                    ? 'bg-[var(--color-primary-light)] border-[var(--border-focus)] text-[var(--color-primary)]'
                    : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300'
                }`}
              >
                <input
                  type="checkbox"
                  className="sr-only"
                  checked={checked}
                  onChange={() => {
                    const next = checked
                      ? selected.filter((s) => s !== opt)
                      : [...selected, opt]
                    onChange(field.name, next.length ? next : null)
                  }}
                />
                {opt}
              </label>
            )
          })}
        </div>
      )
    }

    default:
      return null
  }
}

// ── Form section — renders all custom fields for an entity ───────────────────

interface CustomFieldFormSectionProps {
  fields: CustomFieldDefinition[]
  values: CustomFieldValues
  onChange: (values: CustomFieldValues) => void
}

export function CustomFieldFormSection({
  fields,
  values,
  onChange,
}: CustomFieldFormSectionProps) {
  const entityType = fields[0]?.entity_type
  const dependenciesQuery = useQuery({
    queryKey: ['picklist-dependencies', entityType],
    queryFn: () => picklistsApi.listDependencies(entityType),
    enabled: Boolean(entityType),
    staleTime: 30_000,
  })

  const fieldsByID = useMemo(() => {
    const map = new Map<string, CustomFieldDefinition>()
    for (const field of fields) {
      map.set(field.id, field)
    }
    return map
  }, [fields])

  const dependenciesByTargetID = useMemo(() => {
    const map = new Map<string, PicklistDependency[]>()
    for (const dependency of dependenciesQuery.data ?? []) {
      if (!dependency.is_active) continue
      const list = map.get(dependency.target_field_id) ?? []
      list.push(dependency)
      map.set(dependency.target_field_id, list)
    }
    return map
  }, [dependenciesQuery.data])

  const effectiveOptionsByFieldName = useMemo(() => {
    const result = new Map<string, string[]>()

    for (const field of fields) {
      const baseOptions = field.options ?? []
      const deps = dependenciesByTargetID.get(field.id) ?? []
      if (deps.length === 0 || baseOptions.length === 0) {
        result.set(field.name, baseOptions)
        continue
      }

      let constrained: string[] | null = null
      for (const dep of deps) {
        const sourceField = fieldsByID.get(dep.source_field_id)
        if (!sourceField) continue
        const sourceValue = values[sourceField.name]
        if (sourceValue === null || sourceValue === undefined || sourceValue === '') {
          continue
        }

        const allowedSet = new Set<string>()
        if (Array.isArray(sourceValue)) {
          for (const value of sourceValue) {
            const allowed = dep.mapping[value] ?? []
            for (const item of allowed) allowedSet.add(item)
          }
        } else {
          const allowed = dep.mapping[String(sourceValue)] ?? []
          for (const item of allowed) allowedSet.add(item)
        }

        const allowed = baseOptions.filter((opt) => allowedSet.has(opt))
        if (constrained === null) {
          constrained = allowed
        } else {
          const allowedNow = new Set(allowed)
          constrained = constrained.filter((opt) => allowedNow.has(opt))
        }
      }

      result.set(field.name, constrained ?? baseOptions)
    }

    return result
  }, [dependenciesByTargetID, fields, fieldsByID, values])

  useEffect(() => {
    let nextValues: CustomFieldValues | null = null

    for (const field of fields) {
      const options = effectiveOptionsByFieldName.get(field.name) ?? field.options ?? []
      const current = values[field.name]

      if (field.field_type === 'select') {
        if (typeof current === 'string' && current !== '' && !options.includes(current)) {
          if (!nextValues) nextValues = { ...values }
          nextValues[field.name] = null
        }
      }

      if (field.field_type === 'multiselect') {
        if (Array.isArray(current)) {
          const filtered = current.filter((item) => options.includes(item))
          if (filtered.length !== current.length) {
            if (!nextValues) nextValues = { ...values }
            nextValues[field.name] = filtered.length > 0 ? filtered : null
          }
        }
      }
    }

    if (nextValues) {
      onChange(nextValues)
    }
  }, [effectiveOptionsByFieldName, fields, onChange, values])

  const handleChange = (fieldKey: string, value: CustomFieldValues[string]) => {
    onChange({ ...values, [fieldKey]: value })
  }

  if (fields.length === 0) return null

  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Custom Fields</p>
      {fields.map((field) => (
        <div key={field.id}>
          {field.field_type !== 'checkbox' && (
            <label className="mb-1 block text-xs font-medium text-slate-700">
              {field.label}
              {field.required && <span className="ml-0.5 text-red-500">*</span>}
            </label>
          )}
          <CustomFieldInput
            field={field}
            value={values[field.name] ?? null}
            options={effectiveOptionsByFieldName.get(field.name)}
            onChange={handleChange}
          />
        </div>
      ))}
    </div>
  )
}

// ── Display section — renders custom field values in detail views ─────────────

interface CustomFieldDisplaySectionProps {
  fields: CustomFieldDefinition[]
  values: CustomFieldValues | undefined | null
}

export function CustomFieldDisplaySection({
  fields,
  values,
}: CustomFieldDisplaySectionProps) {
  if (fields.length === 0 || !values) return null

  const populated = fields.filter((f) => {
    const v = values[f.name]
    return v !== null && v !== undefined && v !== '' && !(Array.isArray(v) && v.length === 0)
  })

  if (populated.length === 0) return null

  return (
    <div className="space-y-2">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Custom Fields</p>
      {populated.map((field) => {
        const v = values[field.name]
        let displayValue: string
        if (field.field_type === 'checkbox') {
          displayValue = v ? 'Yes' : 'No'
        } else if (field.field_type === 'multiselect' && Array.isArray(v)) {
          displayValue = v.join(', ')
        } else {
          displayValue = String(v)
        }

        return (
          <div key={field.id} className="flex justify-between gap-4 text-sm">
            <span className="text-slate-500">{field.label}</span>
            <span className="font-medium text-slate-800 text-right">{displayValue}</span>
          </div>
        )
      })}
    </div>
  )
}

// ── Editable section — display + inline edit toggle ──────────────────────────

interface CustomFieldEditableSectionProps {
  fields: CustomFieldDefinition[]
  values: CustomFieldValues | undefined | null
  onSave: (values: CustomFieldValues) => Promise<void>
}

export function CustomFieldEditableSection({
  fields,
  values,
  onSave,
}: CustomFieldEditableSectionProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState<CustomFieldValues>({})
  const [saving, setSaving] = useState(false)

  if (fields.length === 0) return null

  const handleEdit = () => {
    setDraft({ ...(values ?? {}) })
    setEditing(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await onSave(draft)
      setEditing(false)
    } finally {
      setSaving(false)
    }
  }

  if (editing) {
    return (
      <div className="rounded-lg border border-slate-200 px-4 py-3 space-y-3">
        <CustomFieldFormSection fields={fields} values={draft} onChange={setDraft} />
        <div className="flex gap-2 pt-1">
          <Button size="sm" onClick={handleSave} disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
          <Button size="sm" variant="outline" onClick={() => setEditing(false)} disabled={saving}>
            Cancel
          </Button>
        </div>
      </div>
    )
  }

  const populated = fields.filter((f) => {
    const v = (values ?? {})[f.name]
    return v !== null && v !== undefined && v !== '' && !(Array.isArray(v) && v.length === 0)
  })

  return (
    <div className="rounded-lg border border-slate-200 px-4 py-3">
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Custom Fields</p>
        <button
          onClick={handleEdit}
          className="text-xs text-[var(--color-primary)] hover:text-[var(--color-primary)] font-medium"
        >
          Edit
        </button>
      </div>
      {populated.length === 0 ? (
        <p className="text-xs text-slate-400">No values set.</p>
      ) : (
        <div className="space-y-2">
          {populated.map((field) => {
            const v = values![field.name]
            let displayValue: string
            if (field.field_type === 'checkbox') {
              displayValue = v ? 'Yes' : 'No'
            } else if (field.field_type === 'multiselect' && Array.isArray(v)) {
              displayValue = v.join(', ')
            } else {
              displayValue = String(v)
            }
            return (
              <div key={field.id} className="flex justify-between gap-4 text-sm">
                <span className="text-slate-500">{field.label}</span>
                <span className="font-medium text-slate-800 text-right">{displayValue}</span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
