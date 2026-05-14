import { describe, expect, it } from 'vitest'
import { appendUniqueSharingGrant, appendUniqueSharingRule } from './SharingRulesPage'
import type { ACLSharingGrant, ACLSharingRule } from '@/api/types'

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

describe('appendUniqueSharingRule', () => {
  it('does not append duplicate source-target sharing rules', () => {
    const existing: ACLSharingRule[] = [
      {
        source_type: 'role_subordinates',
        source_id: 'sales-role',
        target_type: 'user',
        target_id: 'manager-user',
        access_level: 'read',
      },
    ]

    const result = appendUniqueSharingRule(existing, {
      source_type: 'role_subordinates',
      source_id: 'sales-role',
      target_type: 'user',
      target_id: 'manager-user',
      access_level: 'read',
    })

    expect(result).toBe(existing)
    expect(result).toHaveLength(1)
  })

  it('normalizes all-record rules without a source id', () => {
    const result = appendUniqueSharingRule([], {
      source_type: 'all',
      source_id: 'ignored-source',
      target_type: 'group',
      target_id: 'ops-group',
      access_level: 'write',
    })

    expect(result).toEqual([
      {
        source_type: 'all',
        source_id: undefined,
        target_type: 'group',
        target_id: 'ops-group',
        access_level: 'write',
      },
    ])
  })
})
