import { useEffect, useMemo, useState } from 'react'
import { ArrowDown, ArrowUp, Eye, EyeOff, GripVertical, LayoutTemplate, Plus, RotateCcw, Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import {
  useAdminModuleLayout,
  useResetModuleLayout,
  useSaveModuleLayout,
} from '@/hooks/useModuleConfiguration'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { MODULE_ENTITIES } from '@/lib/moduleConfiguration'
import type { CustomFieldEntityType, ModuleLayout, ModuleLayoutBlock, ModuleLayoutField } from '@/api/types'

function cloneLayout(layout: ModuleLayout): ModuleLayout {
  return {
    ...layout,
    blocks: layout.blocks.map((block) => ({
      ...block,
      fields: block.fields.map((field) => ({ ...field })),
    })),
  }
}

function reorder<T>(items: T[], from: number, to: number): T[] {
  const next = [...items]
  const [item] = next.splice(from, 1)
  next.splice(to, 0, item)
  return next
}

function normalizeOrders(blocks: ModuleLayoutBlock[]): ModuleLayoutBlock[] {
  return blocks.map((block, blockIndex) => ({
    ...block,
    order: blockIndex,
    fields: block.fields.map((field, fieldIndex) => ({ ...field, order: fieldIndex })),
  }))
}

const HARD_REQUIRED_FIELDS: Record<CustomFieldEntityType, Set<string>> = {
  account: new Set(['name']),
  contact: new Set(['first_name', 'last_name']),
  lead: new Set(['first_name', 'last_name']),
  deal: new Set(['title', 'pipeline_id']),
  ticket: new Set(['subject']),
}

const CREATE_FORM_REQUIRED_STANDARD_FIELDS: Record<CustomFieldEntityType, Set<string>> = {
  account: new Set(['name', 'domain', 'industry', 'size']),
  contact: new Set(['first_name', 'last_name', 'email', 'phone', 'account_id', 'stage']),
  lead: new Set(['first_name', 'last_name', 'email', 'phone', 'company', 'lead_source', 'status']),
  deal: new Set(['title', 'value_cents', 'stage', 'expected_close_date']),
  ticket: new Set(['subject', 'status', 'priority']),
}

function isLockedRequiredField(entityType: CustomFieldEntityType, field: ModuleLayoutField, requiredCustomFields: Set<string>): boolean {
  if (field.source === 'standard') {
    return HARD_REQUIRED_FIELDS[entityType]?.has(field.field_key) ?? false
  }
  return requiredCustomFields.has(field.field_key)
}

function isHardRequiredStandardField(entityType: CustomFieldEntityType, field: ModuleLayoutField): boolean {
  return field.source === 'standard' && (HARD_REQUIRED_FIELDS[entityType]?.has(field.field_key) ?? false)
}

function canRequireFieldInCreateForm(entityType: CustomFieldEntityType, field: ModuleLayoutField): boolean {
  if (field.source === 'custom') {
    return true
  }
  if (isHardRequiredStandardField(entityType, field)) {
    return true
  }
  return CREATE_FORM_REQUIRED_STANDARD_FIELDS[entityType]?.has(field.field_key) ?? false
}

function sanitizeUnsupportedRequiredFields(
  blocks: ModuleLayoutBlock[],
  entityType: CustomFieldEntityType
): ModuleLayoutBlock[] {
  return blocks.map((block) => ({
    ...block,
    fields: block.fields.map((field) =>
      field.required && !canRequireFieldInCreateForm(entityType, field)
        ? { ...field, required: false }
        : field
    ),
  }))
}

export function ModuleLayoutsPage() {
  const [entityType, setEntityType] = useState<CustomFieldEntityType>('lead')
  const { data: layout, isLoading } = useAdminModuleLayout(entityType)
  const { data: customFieldDefs = [] } = useCustomFieldDefinitions(entityType)
  const saveLayout = useSaveModuleLayout(entityType)
  const resetLayout = useResetModuleLayout(entityType)
  const [draft, setDraft] = useState<ModuleLayout | null>(null)
  const [status, setStatus] = useState<string | null>(null)

  useEffect(() => {
    if (layout) {
      const cloned = cloneLayout(layout)
      setDraft({ ...cloned, blocks: sanitizeUnsupportedRequiredFields(cloned.blocks, entityType) })
      setStatus(null)
    }
  }, [layout, entityType])

  const selectedEntity = useMemo(
    () => MODULE_ENTITIES.find((item) => item.value === entityType) ?? MODULE_ENTITIES[0],
    [entityType]
  )
  const requiredCustomFields = useMemo(
    () => new Set(customFieldDefs.filter((field) => field.required).map((field) => field.name)),
    [customFieldDefs]
  )

  const updateField = (blockIndex: number, fieldIndex: number, patch: Partial<ModuleLayoutField>) => {
    setDraft((current) => {
      if (!current) return current
      const next = cloneLayout(current)
      const updatedField = {
        ...next.blocks[blockIndex].fields[fieldIndex],
        ...patch,
      }
      next.blocks[blockIndex].fields[fieldIndex] = canRequireFieldInCreateForm(entityType, updatedField)
        ? updatedField
        : { ...updatedField, required: false }
      return next
    })
  }

  const moveField = (blockIndex: number, fieldIndex: number, direction: -1 | 1) => {
    setDraft((current) => {
      if (!current) return current
      const next = cloneLayout(current)
      const target = fieldIndex + direction
      if (target < 0 || target >= next.blocks[blockIndex].fields.length) return current
      next.blocks[blockIndex].fields = reorder(next.blocks[blockIndex].fields, fieldIndex, target)
      next.blocks = normalizeOrders(next.blocks)
      return next
    })
  }

  const moveFieldToBlock = (blockIndex: number, fieldIndex: number, targetBlockId: string) => {
    setDraft((current) => {
      if (!current) return current
      const next = cloneLayout(current)
      const targetBlockIndex = next.blocks.findIndex((block) => block.id === targetBlockId)
      if (targetBlockIndex < 0 || targetBlockIndex === blockIndex) return current
      const [field] = next.blocks[blockIndex].fields.splice(fieldIndex, 1)
      next.blocks[targetBlockIndex].fields.push(field)
      next.blocks = normalizeOrders(next.blocks)
      return next
    })
  }

  const addBlock = () => {
    setDraft((current) => {
      if (!current) return current
      const next = cloneLayout(current)
      const index = next.blocks.length + 1
      next.blocks.push({ id: `section_${Date.now()}`, label: `Section ${index}`, order: next.blocks.length, fields: [] })
      return next
    })
  }

  const renameBlock = (blockIndex: number, label: string) => {
    setDraft((current) => {
      if (!current) return current
      const next = cloneLayout(current)
      next.blocks[blockIndex].label = label
      return next
    })
  }

  const handleSave = async () => {
    if (!draft) return
    const saved = await saveLayout.mutateAsync({
      blocks: normalizeOrders(sanitizeUnsupportedRequiredFields(draft.blocks, entityType)),
    })
    setDraft(cloneLayout(saved))
    setStatus('Layout saved.')
  }

  const handleReset = async () => {
    const reset = await resetLayout.mutateAsync()
    setDraft(cloneLayout(reset))
    setStatus('Layout reset to defaults.')
  }

  return (
    <div className="space-y-6">
      <div className="relative overflow-hidden rounded-2xl border border-slate-200 bg-[#111827] p-6 text-white shadow-sm">
        <div className="absolute -right-16 -top-16 h-48 w-48 rounded-full bg-cyan-400/20 blur-3xl" />
        <div className="absolute bottom-0 left-1/2 h-px w-1/2 bg-gradient-to-r from-transparent via-cyan-200/60 to-transparent" />
        <div className="relative flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-200">CRM Admin Configuration</p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight">Module Layouts</h1>
            <p className="mt-2 max-w-2xl text-sm text-slate-300">
              Shape how records appear in quick-create forms and detail views without touching the underlying custom-field schema.
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleReset} disabled={resetLayout.isPending || !draft}>
              <RotateCcw className="h-4 w-4" />
              Reset
            </Button>
            <Button onClick={handleSave} disabled={saveLayout.isPending || !draft}>
              <Save className="h-4 w-4" />
              {saveLayout.isPending ? 'Saving…' : 'Save Layout'}
            </Button>
          </div>
        </div>
      </div>

      <Tabs value={entityType} onValueChange={(value) => setEntityType(value as CustomFieldEntityType)}>
        <TabsList className="overflow-x-auto">
          {MODULE_ENTITIES.map((entity) => (
            <TabsTrigger key={entity.value} value={entity.value}>{entity.plural}</TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {status && <p className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{status}</p>}

      {isLoading || !draft ? (
        <Card><CardContent className="flex items-center justify-center py-12"><Spinner /></CardContent></Card>
      ) : (
        <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_280px]">
          <div className="space-y-4">
            {draft.blocks.map((block, blockIndex) => (
              <Card key={block.id} className="overflow-hidden">
                <CardHeader className="border-b border-slate-100 bg-slate-50/70">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-center gap-2">
                      <GripVertical className="h-4 w-4 text-slate-400" />
                      <input
                        value={block.label}
                        onChange={(e) => renameBlock(blockIndex, e.target.value)}
                        className="rounded-md border border-transparent bg-transparent px-2 py-1 text-base font-semibold text-slate-900 focus:border-slate-300 focus:bg-white focus:outline-none"
                        aria-label="Section label"
                      />
                    </div>
                    <span className="text-xs font-medium uppercase tracking-wide text-slate-400">
                      {block.fields.length} fields
                    </span>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  {block.fields.length === 0 ? (
                    <div className="px-5 py-8 text-sm text-slate-400">Drop fields into this section using the section menu.</div>
                  ) : (
                    <div className="divide-y divide-slate-100">
                      {block.fields.map((field, fieldIndex) => {
                        const lockedRequired = isLockedRequiredField(entityType, field, requiredCustomFields)
                        const requiredUnsupported = !canRequireFieldInCreateForm(entityType, field)
                        return (
                        <div key={`${field.source}:${field.field_key}`} className="grid gap-3 px-4 py-3 lg:grid-cols-[minmax(180px,1fr)_130px_1.3fr_80px] lg:items-center">
                          <div>
                            <div className="flex items-center gap-2">
                              {field.visible ? <Eye className="h-4 w-4 text-emerald-500" /> : <EyeOff className="h-4 w-4 text-slate-400" />}
                              <input
                                value={field.label}
                                onChange={(e) => updateField(blockIndex, fieldIndex, { label: e.target.value })}
                                className="w-full rounded-md border border-slate-200 px-2 py-1 text-sm font-medium text-slate-900 focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                                aria-label={`${field.field_key} label`}
                              />
                            </div>
                            <p className="mt-1 text-xs text-slate-400">
                              {field.source}.{field.field_key}
                              {lockedRequired && <span className="ml-2 rounded-full bg-amber-50 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-amber-700">Locked required</span>}
                            </p>
                          </div>

                          <select
                            value={block.id}
                            onChange={(e) => moveFieldToBlock(blockIndex, fieldIndex, e.target.value)}
                            className="h-9 rounded-md border border-slate-200 bg-white px-2 text-sm"
                            aria-label="Move to section"
                          >
                            {draft.blocks.map((candidate) => (
                              <option key={candidate.id} value={candidate.id}>{candidate.label}</option>
                            ))}
                          </select>

                          <div className="space-y-1">
                            <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
                              {([
                                ['visible', 'Visible'],
                                ['required', 'Required'],
                                ['quick_create', 'Quick create'],
                                ['mass_edit', 'Mass edit'],
                                ['header', 'Header'],
                                ['key_field', 'Key field'],
                              ] as const).map(([key, label]) => (
                                <label key={key} className="flex items-center gap-1.5 rounded-md border border-slate-200 px-2 py-1 text-xs text-slate-600">
                                  <input
                                    type="checkbox"
                                    checked={Boolean(field[key])}
                                    disabled={
                                      (lockedRequired && (key === 'visible' || key === 'required' || key === 'quick_create')) ||
                                      (key === 'required' && requiredUnsupported)
                                    }
                                    aria-label={`${field.field_key} ${label}`}
                                    onChange={(e) => updateField(blockIndex, fieldIndex, { [key]: e.target.checked })}
                                  />
                                  {label}
                                </label>
                              ))}
                            </div>
                            {requiredUnsupported && (
                              <span className="text-[11px] font-medium text-amber-700">
                                Required locked: not available in quick create
                              </span>
                            )}
                          </div>

                          <div className="flex gap-1 lg:justify-end">
                            <Button type="button" variant="outline" size="icon" onClick={() => moveField(blockIndex, fieldIndex, -1)} disabled={fieldIndex === 0}>
                              <ArrowUp className="h-4 w-4" />
                            </Button>
                            <Button type="button" variant="outline" size="icon" onClick={() => moveField(blockIndex, fieldIndex, 1)} disabled={fieldIndex === block.fields.length - 1}>
                              <ArrowDown className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                        )
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
            <Button type="button" variant="outline" onClick={addBlock}>
              <Plus className="h-4 w-4" />
              Add Section
            </Button>
          </div>

          <Card className="h-max sticky top-4">
            <CardHeader>
              <CardTitle className="flex items-center gap-2"><LayoutTemplate className="h-4 w-4" /> {selectedEntity.label} Contract</CardTitle>
              <CardDescription>Backend validation keeps required system fields visible and present in quick create.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-slate-600">
              <div className="rounded-lg bg-slate-50 p-3">
                <p className="font-semibold text-slate-900">Fields in layout</p>
                <p>{draft.blocks.reduce((sum, block) => sum + block.fields.length, 0)} total fields across {draft.blocks.length} sections.</p>
              </div>
              <p>Hidden fields are removed from the UI, but existing API contracts remain backward compatible.</p>
              <p>Required layout fields are enforced when new records are created.</p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
