import type { SavedView, ViewFilters } from '@/api/types'

type Scalar = string | number | boolean

export function normalizeViewFilters(value: unknown): ViewFilters {
  if (!value || Array.isArray(value) || typeof value !== 'object') return {}

  const normalized: ViewFilters = {}
  for (const [key, raw] of Object.entries(value as Record<string, unknown>)) {
    if (raw === undefined || raw === null || raw === '') continue
    if (typeof raw === 'string' || typeof raw === 'number' || typeof raw === 'boolean') {
      normalized[key] = raw
    }
  }
  return normalized
}

export function pickViewFilters(filters: unknown, allowedKeys: readonly string[]): ViewFilters {
  const normalized = normalizeViewFilters(filters)
  const allowed = new Set(allowedKeys)
  return Object.fromEntries(Object.entries(normalized).filter(([key]) => allowed.has(key))) as ViewFilters
}

export function stringFilter(filters: ViewFilters, key: string): string {
  const value = filters[key]
  return typeof value === 'string' ? value : value == null ? '' : String(value)
}

export function sortKeyFromFilters(filters: ViewFilters, fallback: string): string {
  const sortBy = stringFilter(filters, 'sort_by')
  if (!sortBy) return fallback
  const sortDir = stringFilter(filters, 'sort_dir') === 'asc' ? 'asc' : 'desc'
  return `${sortBy}:${sortDir}`
}

export function normalizeSavedView(view: SavedView): SavedView {
  return {
    ...view,
    filters: normalizeViewFilters(view.filters),
    pin_order: view.pin_order ?? view.pinned_order ?? 0,
  }
}

export function cleanCurrentFilters(filters: Record<string, Scalar | undefined>): ViewFilters {
  return normalizeViewFilters(filters)
}
