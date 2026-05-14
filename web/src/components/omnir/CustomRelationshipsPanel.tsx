import { useMemo, useState } from 'react'
import { Link2, Plus, Search, Trash2 } from 'lucide-react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accountsApi } from '@/api/accounts'
import { contactsApi } from '@/api/contacts'
import { dealsApi } from '@/api/deals'
import { leadsApi } from '@/api/leads'
import { moduleConfigurationApi } from '@/api/moduleConfiguration'
import { ticketsApi } from '@/api/tickets'
import type {
  Account,
  Contact,
  CRMEntityLink,
  CustomFieldEntityType,
  Deal,
  Lead,
  ModuleRelationshipDefinition,
  Ticket,
} from '@/api/types'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { useRuntimeModuleRelationships } from '@/hooks/useModuleConfiguration'
import { CARDINALITY_LABELS, MODULE_ENTITIES } from '@/lib/moduleConfiguration'

interface CustomRelationshipsPanelProps {
  entityType: CustomFieldEntityType
  entityId: string
  title?: string
}

interface RelationshipCandidate {
  id: string
  label: string
  sublabel?: string
}

const ENTITY_SEARCH_TYPE: Partial<Record<CustomFieldEntityType, string>> = {
  account: 'Accounts',
  contact: 'Contacts',
  lead: 'Leads',
  deal: 'Deals',
  ticket: 'Tickets',
}

function entityLabel(entityType: CustomFieldEntityType): string {
  return MODULE_ENTITIES.find((entity) => entity.value === entityType)?.label ?? entityType
}

function recordLabel(entityType: CustomFieldEntityType, record: Account | Contact | Deal | Lead | Ticket): RelationshipCandidate {
  switch (entityType) {
    case 'account': {
      const account = record as Account
      return { id: account.id, label: account.name, sublabel: account.domain }
    }
    case 'contact': {
      const contact = record as Contact
      return {
        id: contact.id,
        label: `${contact.first_name} ${contact.last_name}`.trim() || contact.email || contact.id,
        sublabel: contact.email,
      }
    }
    case 'lead': {
      const lead = record as Lead
      return {
        id: lead.id,
        label: `${lead.first_name} ${lead.last_name}`.trim() || lead.email || lead.id,
        sublabel: lead.company || lead.email,
      }
    }
    case 'deal': {
      const deal = record as Deal
      return { id: deal.id, label: deal.title, sublabel: deal.stage }
    }
    case 'ticket': {
      const ticket = record as Ticket
      return { id: ticket.id, label: ticket.subject, sublabel: ticket.status }
    }
    default:
      return { id: record.id, label: record.id }
  }
}

async function searchRecords(entityType: CustomFieldEntityType, q: string): Promise<RelationshipCandidate[]> {
  const query = q.trim()
  if (query.length < 2) return []
  switch (entityType) {
    case 'account':
      return (await accountsApi.search(query)).map((record) => recordLabel(entityType, record))
    case 'contact':
      return (await contactsApi.search(query)).map((record) => recordLabel(entityType, record))
    case 'deal':
      return (await dealsApi.search(query)).map((record) => recordLabel(entityType, record))
    case 'lead': {
      const data = await leadsApi.list({ search: query, per_page: 10 })
      return (data.data ?? []).map((record) => recordLabel(entityType, record))
    }
    case 'ticket': {
      const data = await ticketsApi.list({ search: query, per_page: 10 })
      return (data.data ?? []).map((record) => recordLabel(entityType, record))
    }
    default:
      return []
  }
}

function targetEntityFor(definition: ModuleRelationshipDefinition, currentType: CustomFieldEntityType): CustomFieldEntityType | null {
  if (definition.from_entity_type === currentType) return definition.to_entity_type
  if (definition.to_entity_type === currentType) return definition.from_entity_type
  return null
}

function linkOppositeId(
  definition: ModuleRelationshipDefinition,
  currentType: CustomFieldEntityType,
  currentId: string,
  link: Pick<CRMEntityLink, 'from_entity_id' | 'to_entity_id'>,
) {
  if (definition.from_entity_type === currentType && link.from_entity_id === currentId) return link.to_entity_id
  if (definition.to_entity_type === currentType && link.to_entity_id === currentId) return link.from_entity_id
  if (definition.from_entity_type === currentType && definition.to_entity_type !== currentType) return link.to_entity_id
  if (definition.to_entity_type === currentType && definition.from_entity_type !== currentType) return link.from_entity_id
  return link.from_entity_id
}

export function CustomRelationshipsPanel({ entityType, entityId, title = 'Custom relationships' }: CustomRelationshipsPanelProps) {
  const queryClient = useQueryClient()
  const { data: runtimeRelationships = [], isLoading: definitionsLoading } = useRuntimeModuleRelationships(entityType)
  const [activeDefinitionId, setActiveDefinitionId] = useState<string | null>(null)
  const [searchText, setSearchText] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)

  const definitions = useMemo(
    () =>
      runtimeRelationships
        .filter((definition) => (
          definition.id &&
          definition.storage_strategy === 'crm_entity_links' &&
          !definition.system_locked &&
          definition.is_enabled &&
          targetEntityFor(definition, entityType)
        ))
        .sort((a, b) => a.order_idx - b.order_idx),
    [entityType, runtimeRelationships]
  )

  const definitionIds = definitions.map((definition) => definition.id).filter(Boolean).join(',')
  const linksQueryKey = ['module-configuration', 'entity-links', entityType, entityId, definitionIds]
  const { data: linksByDefinition = {}, isLoading: linksLoading } = useQuery({
    queryKey: linksQueryKey,
    enabled: definitions.length > 0,
    queryFn: async () => {
      const entries = await Promise.all(
        definitions.map(async (definition) => {
          const links = await moduleConfigurationApi.listEntityLinks(entityType, entityId, definition.id)
          return [definition.id!, links] as const
        })
      )
      return Object.fromEntries(entries)
    },
  })

  const activeDefinition = definitions.find((definition) => definition.id === activeDefinitionId) ?? null
  const activeTargetType = activeDefinition ? targetEntityFor(activeDefinition, entityType) : null
  const activeSearch = activeDefinitionId ? searchText[activeDefinitionId] ?? '' : ''
  const { data: candidates = [], isFetching: searchLoading } = useQuery({
    queryKey: ['module-configuration', 'relationship-search', activeTargetType, activeSearch],
    enabled: Boolean(activeTargetType && activeSearch.trim().length >= 2),
    queryFn: () => searchRecords(activeTargetType!, activeSearch),
    staleTime: 15_000,
  })

  const createLink = useMutation({
    mutationFn: async ({ definition, candidate }: { definition: ModuleRelationshipDefinition; candidate: RelationshipCandidate }) => {
      const existingLinks = linksByDefinition[definition.id!] ?? []
      const alreadyLinked = existingLinks.some((link) => linkOppositeId(definition, entityType, entityId, link) === candidate.id)
      if (alreadyLinked) {
        throw new Error('This record is already linked through this relationship.')
      }
      const currentIsFrom = definition.from_entity_type === entityType
      return moduleConfigurationApi.createEntityLink({
        relationship_definition_id: definition.id!,
        from_entity_type: definition.from_entity_type,
        from_entity_id: currentIsFrom ? entityId : candidate.id,
        to_entity_type: definition.to_entity_type,
        to_entity_id: currentIsFrom ? candidate.id : entityId,
        metadata: {},
      })
    },
    onSuccess: () => {
      setError(null)
      queryClient.invalidateQueries({ queryKey: ['module-configuration', 'entity-links', entityType, entityId] })
      queryClient.invalidateQueries({ queryKey: linksQueryKey })
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Unable to create relationship link.')
    },
  })

  const deleteLink = useMutation({
    mutationFn: (id: string) => moduleConfigurationApi.deleteEntityLink(id),
    onSuccess: () => {
      setError(null)
      queryClient.invalidateQueries({ queryKey: ['module-configuration', 'entity-links', entityType, entityId] })
      queryClient.invalidateQueries({ queryKey: linksQueryKey })
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Unable to remove relationship link.')
    },
  })

  const loading = definitionsLoading || linksLoading

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Link2 className="h-4 w-4" />
          {title}
        </CardTitle>
        <CardDescription>Admin-configured links that use Bundle 5 relationship definitions.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>}
        {loading ? (
          <div className="flex justify-center py-4"><Spinner /></div>
        ) : definitions.length === 0 ? (
          <p className="text-sm text-slate-500">No enabled custom relationships for this module yet.</p>
        ) : (
          definitions.map((definition) => {
            const targetType = targetEntityFor(definition, entityType)!
            const links = linksByDefinition[definition.id!] ?? []
            const searchValue = searchText[definition.id!] ?? ''
            const isActive = activeDefinitionId === definition.id
            return (
              <div key={definition.id} className="rounded-lg border border-slate-200 p-3">
                <div className="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p className="text-sm font-semibold text-slate-900">{definition.label}</p>
                    <p className="text-xs text-slate-500">
                      {entityLabel(entityType)} to {entityLabel(targetType)} · {CARDINALITY_LABELS[definition.cardinality]}
                    </p>
                  </div>
                  <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600">{links.length} linked</span>
                </div>

                <div className="mt-3 space-y-2">
                  {links.length === 0 ? (
                    <p className="text-xs text-slate-400">No linked {entityLabel(targetType).toLowerCase()} records.</p>
                  ) : (
                    links.map((link) => (
                      <div key={link.id} className="flex items-center justify-between gap-2 rounded-md bg-slate-50 px-2 py-2 text-sm">
                        <span className="min-w-0 truncate text-slate-700">
                          {entityLabel(targetType)} <span className="font-mono text-xs text-slate-500">{linkOppositeId(definition, entityType, entityId, link)}</span>
                        </span>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => deleteLink.mutate(link.id)}
                          disabled={deleteLink.isPending}
                          aria-label="Remove relationship link"
                        >
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      </div>
                    ))
                  )}
                </div>

                <div className="mt-3 space-y-2">
                  <div className="relative">
                    <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-slate-400" />
                    <Input
                      value={searchValue}
                      onFocus={() => setActiveDefinitionId(definition.id!)}
                      onChange={(event) => {
                        setActiveDefinitionId(definition.id!)
                        setSearchText((current) => ({ ...current, [definition.id!]: event.target.value }))
                      }}
                      placeholder={`Search ${ENTITY_SEARCH_TYPE[targetType] ?? targetType} to link`}
                      className="pl-8"
                    />
                  </div>
                  {isActive && searchValue.trim().length >= 2 && (
                    <div className="max-h-48 overflow-y-auto rounded-md border border-slate-200 bg-white">
                      {searchLoading ? (
                        <div className="flex justify-center py-3"><Spinner /></div>
                      ) : candidates.length === 0 ? (
                        <p className="px-3 py-2 text-sm text-slate-500">No matching records.</p>
                      ) : (
                        candidates.map((candidate) => (
                          <button
                            key={candidate.id}
                            type="button"
                            className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm hover:bg-slate-50"
                            onClick={() => createLink.mutate({ definition, candidate })}
                            disabled={createLink.isPending}
                          >
                            <span className="min-w-0">
                              <span className="block truncate font-medium text-slate-800">{candidate.label}</span>
                              {candidate.sublabel && <span className="block truncate text-xs text-slate-500">{candidate.sublabel}</span>}
                            </span>
                            <Plus className="h-4 w-4 text-slate-400" />
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
              </div>
            )
          })
        )}
      </CardContent>
    </Card>
  )
}
