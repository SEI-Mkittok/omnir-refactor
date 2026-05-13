import { useMemo, useState } from 'react'
import { GitFork, Lock, Plus, Save, Trash2, Waypoints } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import {
  useDeleteModuleRelationship,
  useModuleRelationships,
  useSaveModuleRelationship,
} from '@/hooks/useModuleConfiguration'
import { CARDINALITY_LABELS, MODULE_ENTITIES } from '@/lib/moduleConfiguration'
import type { CustomFieldEntityType, ModuleRelationshipDefinition, RelationshipCardinality } from '@/api/types'

const CARDINALITIES: RelationshipCardinality[] = ['one_to_one', 'many_to_one', 'one_to_many', 'many_to_many']

const emptyDefinition = (from: CustomFieldEntityType): ModuleRelationshipDefinition => ({
  relationship_key: '',
  from_entity_type: from,
  to_entity_type: 'contact',
  label: '',
  cardinality: 'many_to_many',
  storage_strategy: 'crm_entity_links',
  is_enabled: true,
  system_locked: false,
  order_idx: 100,
  metadata: {},
})

function errorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback
}

export function ModuleRelationshipsPage() {
  const [entityType, setEntityType] = useState<CustomFieldEntityType>('account')
  const { data: relationships = [], isLoading } = useModuleRelationships(entityType)
  const saveRelationship = useSaveModuleRelationship()
  const deleteRelationship = useDeleteModuleRelationship()
  const [drafts, setDrafts] = useState<Record<string, ModuleRelationshipDefinition>>({})
  const [newRelationship, setNewRelationship] = useState<ModuleRelationshipDefinition>(() => emptyDefinition('account'))
  const [status, setStatus] = useState<string | null>(null)

  const selectedEntity = useMemo(
    () => MODULE_ENTITIES.find((item) => item.value === entityType) ?? MODULE_ENTITIES[0],
    [entityType]
  )

  const rows = relationships.map((relationship) => drafts[relationship.id || relationship.relationship_key] ?? relationship)
  const systemCount = rows.filter((row) => row.system_locked).length
  const customCount = rows.length - systemCount

  const setDraft = (relationship: ModuleRelationshipDefinition, patch: Partial<ModuleRelationshipDefinition>) => {
    const key = relationship.id || relationship.relationship_key
    setDrafts((current) => ({ ...current, [key]: { ...relationship, ...current[key], ...patch } }))
  }

  const saveDraft = async (relationship: ModuleRelationshipDefinition) => {
    try {
      const saved = await saveRelationship.mutateAsync(relationship)
      setDrafts((current) => {
        const next = { ...current }
        delete next[relationship.id || relationship.relationship_key]
        return next
      })
      setStatus(`${saved.label} saved.`)
    } catch (error) {
      setStatus(errorMessage(error, 'Relationship could not be saved.'))
    }
  }

  const createCustomRelationship = async () => {
    if (!newRelationship.label.trim()) {
      setStatus('Name the relationship before saving it.')
      return
    }
    try {
      const saved = await saveRelationship.mutateAsync({
        ...newRelationship,
        label: newRelationship.label.trim(),
        storage_strategy: 'crm_entity_links',
        system_locked: false,
        is_enabled: true,
      })
      setNewRelationship(emptyDefinition(entityType))
      setStatus(`${saved.label} created.`)
    } catch (error) {
      setStatus(errorMessage(error, 'Relationship could not be created.'))
    }
  }

  const handleEntityChange = (value: string) => {
    const next = value as CustomFieldEntityType
    setEntityType(next)
    setNewRelationship(emptyDefinition(next))
    setDrafts({})
    setStatus(null)
  }

  return (
    <div className="space-y-6">
      <div className="relative overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-br from-[#17313d] via-[#1f4b5d] to-[#0f172a] p-6 text-white shadow-sm">
        <div className="absolute right-0 top-0 h-full w-1/3 bg-[radial-gradient(circle_at_top_right,rgba(250,204,21,0.25),transparent_55%)]" />
        <div className="relative flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-amber-200">Relationship graph</p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight">Module Relationships</h1>
            <p className="mt-2 max-w-2xl text-sm text-slate-200">
              Configure native CRM links and define safe custom relationships backed by the generic entity-link graph.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="rounded-xl border border-white/15 bg-white/10 px-4 py-3 backdrop-blur">
              <p className="text-xl font-bold">{systemCount}</p>
              <p className="text-xs uppercase tracking-wide text-slate-300">System</p>
            </div>
            <div className="rounded-xl border border-white/15 bg-white/10 px-4 py-3 backdrop-blur">
              <p className="text-xl font-bold">{customCount}</p>
              <p className="text-xs uppercase tracking-wide text-slate-300">Custom</p>
            </div>
          </div>
        </div>
      </div>

      <Tabs value={entityType} onValueChange={handleEntityChange}>
        <TabsList className="overflow-x-auto">
          {MODULE_ENTITIES.map((entity) => (
            <TabsTrigger key={entity.value} value={entity.value}>{entity.plural}</TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {status && <p className="rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-700">{status}</p>}

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_340px]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2"><Waypoints className="h-4 w-4" /> {selectedEntity.label} Relationships</CardTitle>
            <CardDescription>System rows are locked to existing storage. Custom rows use generic CRM entity links.</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex justify-center py-10"><Spinner /></div>
            ) : rows.length === 0 ? (
              <p className="py-8 text-center text-sm text-slate-400">No relationships configured for this module.</p>
            ) : (
              <div className="space-y-3">
                {rows.map((relationship) => (
                  <div key={relationship.id || relationship.relationship_key} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                      <div className="min-w-0 flex-1 space-y-3">
                        <div className="flex flex-wrap items-center gap-2">
                          <input
                            value={relationship.label}
                            onChange={(e) => setDraft(relationship, { label: e.target.value })}
                            className="min-w-[220px] rounded-md border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                            aria-label="Relationship label"
                          />
                          <span className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600">
                            {relationship.system_locked ? <Lock className="h-3 w-3" /> : <GitFork className="h-3 w-3" />}
                            {relationship.system_locked ? 'System' : 'Custom'}
                          </span>
                          <span className="rounded-full bg-cyan-50 px-2.5 py-1 text-xs font-medium text-cyan-700">
                            {CARDINALITY_LABELS[relationship.cardinality]}
                          </span>
                        </div>

                        <div className="grid gap-2 text-xs text-slate-500 sm:grid-cols-3">
                          <span><strong className="text-slate-700">From:</strong> {relationship.from_entity_type}</span>
                          <span><strong className="text-slate-700">To:</strong> {relationship.to_entity_type}</span>
                          <span><strong className="text-slate-700">Storage:</strong> {relationship.storage_strategy}</span>
                        </div>

                        <div className="flex flex-wrap gap-3">
                          <label className="flex items-center gap-2 text-sm text-slate-600">
                            <input
                              type="checkbox"
                              checked={relationship.is_enabled}
                              onChange={(e) => setDraft(relationship, { is_enabled: e.target.checked })}
                            />
                            Enabled
                          </label>
                          <label className="flex items-center gap-2 text-sm text-slate-600">
                            Order
                            <input
                              type="number"
                              value={relationship.order_idx}
                              onChange={(e) => setDraft(relationship, { order_idx: Number(e.target.value) || 0 })}
                              className="h-8 w-20 rounded-md border border-slate-200 px-2 text-sm"
                            />
                          </label>
                        </div>
                      </div>

                      <div className="flex gap-2 lg:justify-end">
                        <Button type="button" variant="outline" onClick={() => saveDraft(relationship)} disabled={saveRelationship.isPending}>
                          <Save className="h-4 w-4" />
                          Save
                        </Button>
                        {!relationship.system_locked && relationship.id && (
                          <Button
                            type="button"
                            variant="outline"
                            className="text-red-600 hover:bg-red-50"
                            onClick={async () => {
                              try {
                                await deleteRelationship.mutateAsync(relationship.id!)
                                setStatus(`${relationship.label} deleted.`)
                              } catch (error) {
                                setStatus(errorMessage(error, 'Relationship is still in use. Disable it or remove its links before deleting.'))
                              }
                            }}
                            disabled={deleteRelationship.isPending}
                          >
                            <Trash2 className="h-4 w-4" />
                            Delete
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="h-max sticky top-4">
          <CardHeader>
            <CardTitle className="flex items-center gap-2"><Plus className="h-4 w-4" /> New Custom Link</CardTitle>
            <CardDescription>Custom links are stored in `crm_entity_links` and enforce cardinality at creation.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">Label</label>
              <input
                value={newRelationship.label}
                onChange={(e) => setNewRelationship((current) => ({ ...current, label: e.target.value }))}
                placeholder="Implementation partner"
                className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">From</label>
                <select
                  value={newRelationship.from_entity_type}
                  onChange={(e) => setNewRelationship((current) => ({ ...current, from_entity_type: e.target.value as CustomFieldEntityType }))}
                  className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
                >
                  {MODULE_ENTITIES.map((entity) => <option key={entity.value} value={entity.value}>{entity.label}</option>)}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">To</label>
                <select
                  value={newRelationship.to_entity_type}
                  onChange={(e) => setNewRelationship((current) => ({ ...current, to_entity_type: e.target.value as CustomFieldEntityType }))}
                  className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
                >
                  {MODULE_ENTITIES.map((entity) => <option key={entity.value} value={entity.value}>{entity.label}</option>)}
                </select>
              </div>
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">Cardinality</label>
              <select
                value={newRelationship.cardinality}
                onChange={(e) => setNewRelationship((current) => ({ ...current, cardinality: e.target.value as RelationshipCardinality }))}
                className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
              >
                {CARDINALITIES.map((cardinality) => (
                  <option key={cardinality} value={cardinality}>{CARDINALITY_LABELS[cardinality]}</option>
                ))}
              </select>
            </div>
            <Button type="button" className="w-full" onClick={createCustomRelationship} disabled={saveRelationship.isPending}>
              <Plus className="h-4 w-4" />
              Create Relationship
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
