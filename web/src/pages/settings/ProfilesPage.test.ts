import { describe, expect, it } from 'vitest'
import { canSaveProfilePermissions, upsertFieldPermissionOverride } from './ProfilesPage'
import type { ACLProfileFieldPermission } from '@/api/types'

describe('upsertFieldPermissionOverride', () => {
  it('replaces an existing module-field override instead of appending a duplicate', () => {
    const existing: ACLProfileFieldPermission[] = [
      { module: 'contacts', field_name: 'annual_revenue', can_write: false },
    ]

    const result = upsertFieldPermissionOverride(existing, {
      module: 'contacts',
      field_name: 'annual_revenue',
      can_write: true,
    })

    expect(result).toHaveLength(1)
    expect(result[0]).toEqual({
      module: 'contacts',
      field_name: 'annual_revenue',
      can_write: true,
    })
  })

  it('appends distinct module-field overrides', () => {
    const existing: ACLProfileFieldPermission[] = [
      { module: 'contacts', field_name: 'annual_revenue', can_write: false },
    ]

    const result = upsertFieldPermissionOverride(existing, {
      module: 'accounts',
      field_name: 'annual_revenue',
      can_write: false,
    })

    expect(result).toHaveLength(2)
    expect(result[1]).toEqual({
      module: 'accounts',
      field_name: 'annual_revenue',
      can_write: false,
    })
  })

  it('trims the field name before matching and saving', () => {
    const existing: ACLProfileFieldPermission[] = [
      { module: 'deals', field_name: 'amount', can_write: false },
    ]

    const result = upsertFieldPermissionOverride(existing, {
      module: 'deals',
      field_name: ' amount ',
      can_write: true,
    })

    expect(result).toEqual([{ module: 'deals', field_name: 'amount', can_write: true }])
  })
})

describe('canSaveProfilePermissions', () => {
  it('allows saving only when editable permissions belong to the selected profile', () => {
    expect(canSaveProfilePermissions({
      selectedProfileId: 'profile-1',
      editableProfileId: 'profile-1',
      isLoading: false,
      isFetching: false,
      isSaving: false,
    })).toBe(true)

    expect(canSaveProfilePermissions({
      selectedProfileId: 'profile-2',
      editableProfileId: 'profile-1',
      isLoading: false,
      isFetching: false,
      isSaving: false,
    })).toBe(false)
  })

  it('blocks saving while permissions are loading, fetching, or already saving', () => {
    expect(canSaveProfilePermissions({
      selectedProfileId: 'profile-1',
      editableProfileId: 'profile-1',
      isLoading: true,
      isFetching: false,
      isSaving: false,
    })).toBe(false)

    expect(canSaveProfilePermissions({
      selectedProfileId: 'profile-1',
      editableProfileId: 'profile-1',
      isLoading: false,
      isFetching: true,
      isSaving: false,
    })).toBe(false)

    expect(canSaveProfilePermissions({
      selectedProfileId: 'profile-1',
      editableProfileId: 'profile-1',
      isLoading: false,
      isFetching: false,
      isSaving: true,
    })).toBe(false)
  })
})
