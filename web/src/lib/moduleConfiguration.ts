import type {
  CustomFieldDefinition,
  CustomFieldEntityType,
  ModuleLayout,
  ModuleLayoutField,
  ModuleLayoutFieldSource,
} from '@/api/types'

export const MODULE_ENTITIES: { value: CustomFieldEntityType; label: string; plural: string }[] = [
  { value: 'lead', label: 'Lead', plural: 'Leads' },
  { value: 'contact', label: 'Contact', plural: 'Contacts' },
  { value: 'account', label: 'Account', plural: 'Accounts' },
  { value: 'deal', label: 'Deal', plural: 'Deals' },
  { value: 'ticket', label: 'Ticket', plural: 'Tickets' },
]

export const CARDINALITY_LABELS = {
  one_to_one: 'One-to-one',
  many_to_one: 'Many-to-one',
  one_to_many: 'One-to-many',
  many_to_many: 'Many-to-many',
} as const

export function findLayoutField(
  layout: ModuleLayout | undefined,
  source: ModuleLayoutFieldSource,
  fieldKey: string,
): ModuleLayoutField | undefined {
  for (const block of layout?.blocks ?? []) {
    const field = block.fields.find((item) => item.source === source && item.field_key === fieldKey)
    if (field) return field
  }
  return undefined
}

export function isLayoutFieldVisible(
  layout: ModuleLayout | undefined,
  source: ModuleLayoutFieldSource,
  fieldKey: string,
  mode: 'quick_create' | 'detail' = 'detail',
): boolean {
  const field = findLayoutField(layout, source, fieldKey)
  if (!field) return !layout
  if (!field.visible) return false
  if (mode === 'quick_create' && !field.quick_create && !field.required) return false
  return true
}

export function layoutFieldLabel(
  layout: ModuleLayout | undefined,
  source: ModuleLayoutFieldSource,
  fieldKey: string,
  fallback: string,
): string {
  return findLayoutField(layout, source, fieldKey)?.label || fallback
}

export function isLayoutFieldRequired(
  layout: ModuleLayout | undefined,
  source: ModuleLayoutFieldSource,
  fieldKey: string,
): boolean {
  return Boolean(findLayoutField(layout, source, fieldKey)?.required)
}

export function applyLayoutToCustomFields(
  fields: CustomFieldDefinition[],
  layout: ModuleLayout | undefined,
  mode: 'quick_create' | 'detail' = 'detail',
): CustomFieldDefinition[] {
  const withLayout = fields
    .map((field) => {
      const layoutField = findLayoutField(layout, 'custom', field.name)
      if (!layoutField) return { field, order: field.order_idx ?? 0 }
      if (!layoutField.visible) return null
      if (mode === 'quick_create' && !layoutField.quick_create && !layoutField.required) return null
      return {
        field: {
          ...field,
          label: layoutField.label || field.label,
          required: Boolean(layoutField.required || field.required),
        },
        order: layoutField.order,
      }
    })
    .filter((item): item is { field: CustomFieldDefinition; order: number } => Boolean(item))

  return withLayout.sort((a, b) => a.order - b.order).map((item) => item.field)
}

export function visibleLayoutBlocks(layout: ModuleLayout | undefined): ModuleLayout['blocks'] {
  return (layout?.blocks ?? [])
    .map((block) => ({ ...block, fields: block.fields.filter((field) => field.visible) }))
    .filter((block) => block.fields.length > 0)
}
