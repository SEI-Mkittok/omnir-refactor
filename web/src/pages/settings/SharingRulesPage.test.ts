import { describe, expect, it } from 'vitest'
import { appendUniqueSharingGrant } from './SharingRulesPage'
import type { ACLSharingGrant } from '@/api/types'

describe('appendUniqueSharingGrant', () => {
  it('does not append duplicate grants with the same grantee and access level', () => {
    const existing: ACLSharingGrant[] = [
      { grantee_type: 'role', grantee_id: 'role-1', access_level: 'read' },
    ]

    const result = appendUniqueSharingGrant(existing, {
      grantee_type: 'role',
      grantee_id: 'role-1',
      access_level: 'read',
    })

    expect(result).toBe(existing)
    expect(result).toHaveLength(1)
  })

  it('allows the same grantee to have distinct access levels', () => {
    const existing: ACLSharingGrant[] = [
      { grantee_type: 'role', grantee_id: 'role-1', access_level: 'read' },
    ]

    const result = appendUniqueSharingGrant(existing, {
      grantee_type: 'role',
      grantee_id: 'role-1',
      access_level: 'write',
    })

    expect(result).toHaveLength(2)
    expect(result[1]).toEqual({
      grantee_type: 'role',
      grantee_id: 'role-1',
      access_level: 'write',
    })
  })
})
