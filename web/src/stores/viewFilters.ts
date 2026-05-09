import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { ViewEntityType, ViewFilters } from '@/api/types'

// Per-entity filter state
export type EntityFilters = Record<string, string | number | undefined>

interface ViewFiltersState {
  // Active view ID per entity type (null = no saved view active)
  activeViewId: Record<ViewEntityType, string | null>
  // Current filter values per entity type
  filters: Record<ViewEntityType, EntityFilters>
  // Whether current filters differ from the active view's saved filters
  hasUnsavedChanges: Record<ViewEntityType, boolean>

  setActiveView: (entity: ViewEntityType, viewId: string | null, savedFilters?: ViewFilters) => void
  setFilters: (entity: ViewEntityType, filters: EntityFilters, savedFilters?: ViewFilters) => void
  resetFilters: (entity: ViewEntityType) => void
  markSaved: (entity: ViewEntityType) => void
}

const DEFAULT_FILTERS: Record<ViewEntityType, EntityFilters> = {
  contacts: {},
  accounts: {},
  deals: {},
  leads: {},
  tickets: {},
  quotes: {},
}

const DEFAULT_ACTIVE_VIEWS: Record<ViewEntityType, string | null> = {
  contacts: null,
  accounts: null,
  deals: null,
  leads: null,
  tickets: null,
  quotes: null,
}

const DEFAULT_UNSAVED: Record<ViewEntityType, boolean> = {
  contacts: false,
  accounts: false,
  deals: false,
  leads: false,
  tickets: false,
  quotes: false,
}

function filtersMatch(a: EntityFilters, b: ViewFilters): boolean {
  const aKeys = Object.keys(a).filter((k) => a[k] !== undefined && a[k] !== '')
  const bKeys = Object.keys(b).filter((k) => b[k] !== undefined && b[k] !== '')
  if (aKeys.length !== bKeys.length) return false
  return aKeys.every((k) => String(a[k]) === String(b[k]))
}

export const useViewFiltersStore = create<ViewFiltersState>()(
  persist(
    (set) => ({
      activeViewId: { ...DEFAULT_ACTIVE_VIEWS },
      filters: { ...DEFAULT_FILTERS },
      hasUnsavedChanges: { ...DEFAULT_UNSAVED },

      setActiveView: (entity, viewId, savedFilters) =>
        set((state) => ({
          activeViewId: { ...state.activeViewId, [entity]: viewId },
          filters: {
            ...state.filters,
            [entity]: savedFilters ? ({ ...savedFilters } as EntityFilters) : state.filters[entity],
          },
          hasUnsavedChanges: { ...state.hasUnsavedChanges, [entity]: false },
        })),

      setFilters: (entity, filters, savedFilters) =>
        set((state) => {
          const activeViewId = state.activeViewId[entity]
          const hasUnsaved = activeViewId !== null && savedFilters
            ? !filtersMatch(filters, savedFilters)
            : false
          return {
            filters: { ...state.filters, [entity]: filters },
            hasUnsavedChanges: { ...state.hasUnsavedChanges, [entity]: hasUnsaved },
          }
        }),

      resetFilters: (entity) =>
        set((state) => ({
          filters: { ...state.filters, [entity]: {} },
          activeViewId: { ...state.activeViewId, [entity]: null },
          hasUnsavedChanges: { ...state.hasUnsavedChanges, [entity]: false },
        })),

      markSaved: (entity) =>
        set((state) => ({
          hasUnsavedChanges: { ...state.hasUnsavedChanges, [entity]: false },
        })),
    }),
    { name: 'omnir-view-filters' }
  )
)
