import { useEffect, useState } from 'react'
import { Plus, Save, Share2, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import {
  useACLGroups,
  useACLRoles,
  useGroupMemberCandidates,
  useSaveSharingRules,
  useSharingRules,
} from '@/hooks/useAccessSettings'
import type {
  ACLModule,
  ACLSharingGrant,
  ACLSharingModuleRule,
  ACLSharingRule,
  SharingAccessLevel,
  SharingDefaultMode,
  SharingPrincipalType,
} from '@/api/types'

type TargetPrincipalType = Exclude<SharingPrincipalType, 'all'>

const DEFAULT_LABELS: Record<SharingDefaultMode, string> = {
  private: 'Private',
  public_read: 'Public read',
  public_rw: 'Public read/write',
}

const PRINCIPAL_LABELS: Record<SharingPrincipalType, string> = {
  all: 'All records',
  user: 'User',
  role: 'Role',
  role_subordinates: 'Role + subordinates',
  group: 'Group',
}

function moduleLabel(module: ACLModule) {
  return module.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

interface DraftRule {
  source_type: SharingPrincipalType
  source_id: string
  target_type: TargetPrincipalType
  target_id: string
  access_level: SharingAccessLevel
}

export function appendUniqueSharingRule(
  rules: ACLSharingRule[],
  draft: DraftRule
): ACLSharingRule[] {
  const normalizedSourceID = draft.source_type === 'all' ? undefined : draft.source_id
  const exists = rules.some(
    (rule) =>
      rule.source_type === draft.source_type &&
      (rule.source_id ?? '') === (normalizedSourceID ?? '') &&
      rule.target_type === draft.target_type &&
      rule.target_id === draft.target_id &&
      rule.access_level === draft.access_level
  )
  if (exists) return rules
  return [
    ...rules,
    {
      source_type: draft.source_type,
      source_id: normalizedSourceID,
      target_type: draft.target_type,
      target_id: draft.target_id,
      access_level: draft.access_level,
    },
  ]
}

export function appendUniqueSharingGrant(
  grants: ACLSharingGrant[],
  draft: { grantee_type: 'role' | 'group'; grantee_id: string; access_level: SharingAccessLevel }
): ACLSharingGrant[] {
  const exists = grants.some(
    (grant) =>
      grant.grantee_type === draft.grantee_type &&
      grant.grantee_id === draft.grantee_id &&
      grant.access_level === draft.access_level
  )
  if (exists) return grants
  return [...grants, draft]
}

function normalizeModuleRules(rules: ACLSharingModuleRule[]): ACLSharingModuleRule[] {
  return rules.map((rule) => ({
    ...rule,
    advanced_rules:
      rule.advanced_rules ??
      rule.grants?.map((grant) => ({
        id: grant.id,
        org_id: grant.org_id,
        module: grant.module,
        source_type: 'all' as const,
        target_type: grant.grantee_type,
        target_id: grant.grantee_id,
        access_level: grant.access_level,
      })) ??
      [],
  }))
}

export function SharingRulesPage() {
  const { data, isLoading } = useSharingRules()
  const { data: roles = [] } = useACLRoles()
  const { data: groups = [] } = useACLGroups()
  const { data: usersData } = useGroupMemberCandidates()
  const saveRules = useSaveSharingRules()
  const [rules, setRules] = useState<ACLSharingModuleRule[]>([])
  const [draftRules, setDraftRules] = useState<Record<string, DraftRule>>({})

  const users = usersData?.data ?? []

  useEffect(() => {
    setRules(normalizeModuleRules(data?.rules ?? []))
  }, [data])

  function updateRule(module: ACLModule, patch: Partial<ACLSharingModuleRule>) {
    setRules((current) =>
      current.map((rule) => (rule.module === module ? { ...rule, ...patch } : rule))
    )
  }

  function addRule(module: ACLModule) {
    const draft = draftFor(module)
    if (!draft.target_id) return
    if (draft.source_type !== 'all' && !draft.source_id) return
    const advancedRules = rules.find((rule) => rule.module === module)?.advanced_rules ?? []
    updateRule(module, {
      advanced_rules: appendUniqueSharingRule(advancedRules, draft),
    })
    setDraftRules((current) => ({
      ...current,
      [module]: emptyDraftRule(),
    }))
  }

  function removeRule(module: ACLModule, index: number) {
    const rule = rules.find((candidate) => candidate.module === module)
    if (!rule) return
    updateRule(module, { advanced_rules: rule.advanced_rules.filter((_, i) => i !== index) })
  }

  function nameFor(type: SharingPrincipalType, id?: string) {
    if (type === 'all') return 'all records'
    if (type === 'user') {
      const user = users.find((candidate) => candidate.id === id)
      return user ? `${user.name} (${user.email})` : 'Unknown user'
    }
    if (type === 'role' || type === 'role_subordinates') {
      return roles.find((role) => role.id === id)?.name ?? 'Unknown role'
    }
    return groups.find((group) => group.id === id)?.name ?? 'Unknown group'
  }

  function optionsFor(type: SharingPrincipalType) {
    if (type === 'user') return users.map((user) => ({ id: user.id, name: `${user.name} (${user.email})` }))
    if (type === 'role' || type === 'role_subordinates') return roles
    if (type === 'group') return groups
    return []
  }

  function emptyDraftRule(): DraftRule {
    return {
      source_type: 'all',
      source_id: '',
      target_type: 'role',
      target_id: '',
      access_level: 'read',
    }
  }

  function draftFor(module: ACLModule): DraftRule {
    return draftRules[module] ?? emptyDraftRule()
  }

  function setDraft(module: ACLModule, patch: Partial<DraftRule>) {
    setDraftRules((current) => ({
      ...current,
      [module]: { ...draftFor(module), ...patch },
    }))
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Sharing Rules</h1>
          <p className="mt-1 text-sm text-[#6B7280]">
            Set org-wide defaults and precise source-to-target record access.
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
          const sourceOptions = optionsFor(draft.source_type)
          const targetOptions = optionsFor(draft.target_type)
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
                {rule.advanced_rules.map((advancedRule, index) => (
                  <div key={`${advancedRule.id ?? index}`} className="flex flex-wrap items-center justify-between gap-3 px-3 py-2">
                    <span className="text-sm text-slate-700">
                      Records owned by <strong>{PRINCIPAL_LABELS[advancedRule.source_type]}</strong>{' '}
                      {advancedRule.source_type !== 'all' ? nameFor(advancedRule.source_type, advancedRule.source_id) : ''}
                      {' '}are {advancedRule.access_level === 'write' ? 'read/write' : 'readable'} by{' '}
                      <strong>{PRINCIPAL_LABELS[advancedRule.target_type]}</strong>{' '}
                      {nameFor(advancedRule.target_type, advancedRule.target_id)}
                    </span>
                    <button
                      type="button"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => removeRule(rule.module, index)}
                      aria-label="Remove sharing rule"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                ))}
                {rule.advanced_rules.length === 0 && (
                  <p className="px-3 py-3 text-sm text-slate-500">No advanced rules. Private records stay limited to owners, managers, and org defaults.</p>
                )}
              </div>

              <div className="mt-3 grid gap-3 lg:grid-cols-[160px_1fr_180px_1fr_140px_auto]">
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.source_type}
                  onChange={(e) =>
                    setDraft(rule.module, {
                      source_type: e.target.value as SharingPrincipalType,
                      source_id: '',
                    })
                  }
                >
                  {(['all', 'user', 'role', 'role_subordinates', 'group'] as SharingPrincipalType[]).map((type) => (
                    <option key={type} value={type}>{PRINCIPAL_LABELS[type]}</option>
                  ))}
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm disabled:bg-slate-50"
                  value={draft.source_id}
                  disabled={draft.source_type === 'all'}
                  onChange={(e) => setDraft(rule.module, { source_id: e.target.value })}
                >
                  <option value="">{draft.source_type === 'all' ? 'All owners' : `Select ${PRINCIPAL_LABELS[draft.source_type]}`}</option>
                  {sourceOptions.map((option) => (
                    <option key={option.id} value={option.id}>{option.name}</option>
                  ))}
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.target_type}
                  onChange={(e) =>
                    setDraft(rule.module, {
                      target_type: e.target.value as TargetPrincipalType,
                      target_id: '',
                    })
                  }
                >
                  {(['user', 'role', 'role_subordinates', 'group'] as TargetPrincipalType[]).map((type) => (
                    <option key={type} value={type}>{PRINCIPAL_LABELS[type]}</option>
                  ))}
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.target_id}
                  onChange={(e) => setDraft(rule.module, { target_id: e.target.value })}
                >
                  <option value="">Select {PRINCIPAL_LABELS[draft.target_type]}</option>
                  {targetOptions.map((option) => (
                    <option key={option.id} value={option.id}>{option.name}</option>
                  ))}
                </select>
                <select
                  className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                  value={draft.access_level}
                  onChange={(e) => setDraft(rule.module, { access_level: e.target.value as SharingAccessLevel })}
                >
                  <option value="read">Read</option>
                  <option value="write">Read/write</option>
                </select>
                <Button type="button" variant="outline" onClick={() => addRule(rule.module)}>
                  <Plus className="h-4 w-4" />
                  Add Rule
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
