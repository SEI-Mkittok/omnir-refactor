import { describe, expect, it } from 'vitest'
import { applyPlatformRoleChange } from './userAssignments'

describe('applyPlatformRoleChange', () => {
  it('clears ACL assignments when the platform role changes', () => {
    const result = applyPlatformRoleChange(
      {
        name: 'Avery Admin',
        email: 'avery@omnir.test',
        role: 'agent',
        role_id: 'agent-role-id',
        profile_id: 'sales-profile-id',
      },
      'admin'
    )

    expect(result).toEqual({
      name: 'Avery Admin',
      email: 'avery@omnir.test',
      role: 'admin',
    })
    expect(result).not.toHaveProperty('role_id')
    expect(result).not.toHaveProperty('profile_id')
  })

  it('preserves ACL assignments when the platform role is unchanged', () => {
    const form = {
      role: 'agent' as const,
      role_id: 'agent-role-id',
      profile_id: 'sales-profile-id',
    }

    expect(applyPlatformRoleChange(form, 'agent')).toBe(form)
  })
})
