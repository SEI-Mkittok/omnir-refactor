import { describe, expect, it } from 'vitest'
import { canAccessAdminModule, canAccessAdminPath, isWorkspaceAdmin } from './access'
import type { User } from '@/api/types'

function makeUser(patch: Partial<User> = {}): User {
  return {
    id: 'user-1',
    org_id: 'org-1',
    email: 'user@acme.test',
    name: 'User',
    role: 'agent',
    created_at: '2026-05-11T00:00:00Z',
    updated_at: '2026-05-11T00:00:00Z',
    ...patch,
  }
}

describe('access helpers', () => {
  it('treats users:admin as workspace admin for the Users route', () => {
    const user = makeUser({
      profile_name: 'Delegated People Admin',
      permissions: {
        users: { admin: true },
      },
    })

    expect(isWorkspaceAdmin(user)).toBe(true)
    expect(canAccessAdminModule(user, 'users')).toBe(true)
    expect(canAccessAdminPath(user, '/users')).toBe(true)
  })

  it('treats settings:admin as workspace admin for access-settings routes', () => {
    const user = makeUser({
      profile_name: 'Access Steward',
      permissions: {
        settings: { admin: true },
      },
    })

    expect(isWorkspaceAdmin(user)).toBe(true)
    expect(canAccessAdminPath(user, '/settings/roles')).toBe(true)
    expect(canAccessAdminPath(user, '/settings/profiles')).toBe(true)
    expect(canAccessAdminPath(user, '/settings/sharing-rules')).toBe(true)
    expect(canAccessAdminPath(user, '/settings/groups')).toBe(true)
  })

  it('does not let unrelated admin permissions open other admin modules', () => {
    const user = makeUser({
      permissions: {
        users: { admin: true },
      },
    })

    expect(canAccessAdminPath(user, '/settings/roles')).toBe(false)
    expect(canAccessAdminModule(user, 'billing')).toBe(false)
  })
})
