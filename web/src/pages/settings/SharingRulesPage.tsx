import { useEffect, useState } from 'react'
import { Plus, Save, Share2, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import {
  useACLGroups,
  useACLRoles,
  useSaveSharingRules,
  useSharingRules,
} from '@/hooks/useAccessSettings'
import type {
  ACLModule,
  ACLSharingGrant,
  ACLSharingModuleRule,
  SharingAccessLevel,
  SharingDefaultMode,
  SharingGranteeType,
} from '@/api/types'

const DEFAULT_LABELS: Record<SharingDefaultMode, string> = {
  private: 'Private',
  public_read: 'Public read',
  public_rw: 'Public read/write',
}

function moduleLabel(module: ACLModule) {
  return module.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

interface DraftGrant {
  grantee_type: SharingGranteeType
  grantee_id: string
  access_level: SharingAccessLevel
}

export function appendUniqueSharingGrant(
  grants: ACLSharingGrant[],
  draft: DraftGrant
): ACLSharingGrant[] {
  const exists = grants.some(
    (grant) =>
      grant.grantee_type === draft.grantee_type &&
      grant.grantee_id === draft.grantee_id &&
      grant.access_level === draft.access_level
  )
  if (exists) return grants
  return [
    ...grants,
    {
      grantee_type: draft.grantee_type,
      grantee_id: draft.grantee_id,
      access_level: draft.access_level,
    },
  ]
}

export function SharingRulesPage() {
  const { data, isLoading } = useSharingRules()
  const { data: roles = [] } = useACLRoles()
  const { data: groups = [] } = useACLGroups()
  const saveRules = useSaveSharingRules()
  const [rules, setRules] = useState<ACLSharingModuleRule[]>([])
  const [draftGrants, setDraftGrants] = useState<Record<string, DraftGrant>>({})

  useEffect(() => {
    setRules(data?.rules ?? [])
  }, [data])

  function updateRule(module: ACLModule, patch: Partial<ACLSharingModuleRule>) {
    setRules((current) =>
      current.map((rule) => (rule.module === module ? { ...rule, ...patch } : rule))
    )
  }

  function addGrant(module: ACLModule) {
    const draft = draftGrants[module]
    if (!draft?.grantee_id) return
    const grants = rules.find((rule) => rule.module === module)?.grants ?? []
    updateRule(module, {
      grants: appendUniqueSharingGrant(grants, draft),
    })
    setDraftGrants((current) => ({
      ...current,
      [module]: { grantee_type: 'role', grantee_id: '', access_level: 'read' },
    }))
  }

  function removeGrant(module: ACLModule, index: number) {
    const rule = rules.find((candidate) => candidate.module === module)
    if (!rule) return
    updateRule(module, { grants: rule.grants.filter((_, i) => i !== index) })
  }

  function granteeName(grant: ACLSharingGrant) {
    if (grant.grantee_type === 'role') {
      return roles.find((role) => role.id === grant.grantee_id)?.name ?? 'Unknown role'
    }
    return groups.find((group) => group.id === grant.grantee_id)?.name ?? 'Unknown group'
  }

  function draftFor(module: ACLModule): DraftGrant {
    return draftGrants[module] ?? { grantee_type: 'role', grantee_id: '', access_level: 'read' }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Sharing Rules</h1>
          <p className="mt-1 text-sm text-[#6B7280]">
            Set org-wide defaults and grant all-record read/write access by role or group.
          </p>
        </div>
        <Button onClick={() => saveRules.mutate({ rules })} disabled={saveRules.isPending || isLoading}>
          <Save className="h-4 w-4" />
          {saveRules.isPending ? 'Saving...' : 'Save Rules'}
        </Button>
      </div>

      <div className="space-y-4">
        {rules.map((rule) => {
          const draft = draftFor(rule.module)
          const granteeOptions = draft.grantee_type === 'role' ? roles : groups
          return (
            <section key={rule.module} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <Share2 className="h-4 w-4 text-slate-500" />
                  <h2 className="text-sm font-semibold text-slate-900">{moduleLabel(rule.module)}</h2>
                </div>
                <label className="flex items-center gap-2 text-sm text-slate-600">
                  Org default
                  <select
                    className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                    value={rule.mode}
                    onChange={(e) => updateRule(rule.module, { mode: e.target.value as SharingDefaultMode })}
                  >
                    {Object.entries(DEFAULT_LABELS).map(([value, label]) => (
                      <option key={value} value={value}>{label}</option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="mt-4 divide-y divide-slate-100 rounded-lg border border-slate-100">
                {rule.grants.map((grant, index) => (
                  <div key={`${grant.grantee_type}-${grant.grantee_id}-${index}`} className="flex flex-wrap items-center justify-between gap-3 px-3 py-2">
                    <span className="text-sm text-slate-700">
                      <strong className="capitalize">{grant.grantee_type}</strong> {granteeName(grant)} can {grant.access_level}
                    </span>
                    <button
                      type="button"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => removeGrant(rule.module, index)}
                      aria-label="Remove grant"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                ))}
                {rule.grants.length === 0 && (
                  <p className="px-3 py-3 text-sm text-slate-500">No explicit grants. Only owners/assignees can see private records.</p>
                )}
              </div>

              <div className="mt-3 grid gap-3 md:grid-cols-[140px_1fr_140px_auto]">
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.grantee_type}
                  onChange={(e) =>
                    setDraftGrants((current) => ({
                      ...current,
                      [rule.module]: { ...draft, grantee_type: e.target.value as SharingGranteeType, grantee_id: '' },
                    }))
                  }
                >
                  <option value="role">Role</option>
                  <option value="group">Group</option>
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.grantee_id}
                  onChange={(e) =>
                    setDraftGrants((current) => ({
                      ...current,
                      [rule.module]: { ...draft, grantee_id: e.target.value },
                    }))
                  }
                >
                  <option value="">Select {draft.grantee_type}</option>
                  {granteeOptions.map((option) => (
                    <option key={option.id} value={option.id}>{option.name}</option>
                  ))}
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.access_level}
                  onChange={(e) =>
                    setDraftGrants((current) => ({
                      ...current,
                      [rule.module]: { ...draft, access_level: e.target.value as SharingAccessLevel },
                    }))
                  }
                >
                  <option value="read">Read</option>
                  <option value="write">Write</option>
                </select>
                <Button type="button" variant="outline" onClick={() => addGrant(rule.module)}>
                  <Plus className="h-4 w-4" />
                  Grant
                </Button>
              </div>
            </section>
          )
        })}
        {!isLoading && rules.length === 0 && (
          <div className="rounded-xl border border-slate-200 bg-white px-4 py-8 text-center text-sm text-slate-500 shadow-sm">
            No sharing modules configured yet.
          </div>
        )}
      </div>
    </div>
  )
}
