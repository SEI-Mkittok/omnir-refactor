import { useMemo, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Bookmark, Building2, LifeBuoy, Search, TrendingUp, Users, X } from 'lucide-react'
import { searchApi } from '@/api/search'
import { accountsApi } from '@/api/accounts'
import { contactsApi } from '@/api/contacts'
import { Spinner } from '@/components/ui/Spinner'
import { Input } from '@/components/ui/Input'
import type {
  Account,
  Contact,
  Deal,
  SearchEntityType,
  SearchFilters,
  SearchRelationship,
  SearchResult,
  Ticket,
} from '@/api/types'

type SearchItem = NonNullable<SearchResult[keyof SearchResult]>[number]
type ScopedPickerKind = 'account' | 'contact'

interface SearchPreset {
  id: string
  name: string
  filters: Omit<SearchFilters, 'q'>
  accountName?: string
  contactName?: string
}

const PRESET_STORAGE_KEY = 'omnir-search-filter-presets'

const ENTITY_OPTIONS: Array<{ value: SearchEntityType | ''; label: string; icon: React.ElementType }> = [
  { value: '', label: 'All', icon: Search },
  { value: 'contacts', label: 'Contacts', icon: Users },
  { value: 'accounts', label: 'Accounts', icon: Building2 },
  { value: 'deals', label: 'Deals', icon: TrendingUp },
  { value: 'tickets', label: 'Tickets', icon: LifeBuoy },
]

const RELATIONSHIP_OPTIONS = [
  { value: '', label: 'Any relationship' },
  { value: 'primary', label: 'Primary' },
  { value: 'linked', label: 'Linked' },
  { value: 'requester', label: 'Requester' },
  { value: 'parent', label: 'Parent' },
  { value: 'subsidiary', label: 'Subsidiary' },
  { value: 'partner', label: 'Partner' },
  { value: 'reseller', label: 'Reseller' },
  { value: 'vendor', label: 'Vendor' },
]

function readPresets(): SearchPreset[] {
  if (typeof window === 'undefined') return []
  try {
    return JSON.parse(window.localStorage.getItem(PRESET_STORAGE_KEY) ?? '[]')
  } catch {
    return []
  }
}

function writePresets(presets: SearchPreset[]) {
  window.localStorage.setItem(PRESET_STORAGE_KEY, JSON.stringify(presets))
}

function resultCount(data: SearchResult | undefined) {
  return (data?.contacts?.length ?? 0) + (data?.accounts?.length ?? 0) + (data?.deals?.length ?? 0) + (data?.tickets?.length ?? 0)
}

function relationshipLabel(value?: string) {
  if (!value) return ''
  return value.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

function relationshipKey(item: SearchItem) {
  const account = 'related_account' in item ? item.related_account : undefined
  const contact = 'related_contact' in item ? item.related_contact : undefined
  const related = account ?? contact
  return related ? `${related.entity_type}:${related.id}:${item.relationship_type ?? 'linked'}` : 'unscoped'
}

function relationshipTitle(item: SearchItem, scopeLabel: string) {
  const account = 'related_account' in item ? item.related_account : undefined
  const contact = 'related_contact' in item ? item.related_contact : undefined
  const related = account ?? contact
  if (!related) return 'Other matches'
  const type = relationshipLabel(item.relationship_type)
  return type ? `${type} related to ${scopeLabel}` : `Related to ${scopeLabel}`
}

function updateParam(params: URLSearchParams, key: string, value?: string) {
  if (value) params.set(key, value)
  else params.delete(key)
}

function ScopedPicker({
  kind,
  selectedId,
  selectedName,
  onSelect,
  onClear,
}: {
  kind: ScopedPickerKind
  selectedId?: string
  selectedName?: string
  onSelect: (id: string, name: string) => void
  onClear: () => void
}) {
  const [term, setTerm] = useState('')
  const query = useQuery<Array<Account | Contact>>({
    queryKey: ['search-scope-picker', kind, term],
    queryFn: async () => kind === 'account' ? accountsApi.search(term) : contactsApi.search(term),
    enabled: term.trim().length >= 2,
    staleTime: 30_000,
  })
  const items = query.data ?? []

  if (selectedId) {
    return (
      <FilterPill label={`${kind === 'account' ? 'Account' : 'Contact'}: ${selectedName ?? selectedId}`} onClear={onClear} />
    )
  }

  return (
    <div className="relative min-w-[220px]">
      <Input
        value={term}
        onChange={(e) => setTerm(e.target.value)}
        placeholder={`Filter by ${kind}`}
        className="h-9"
      />
      {term.trim().length >= 2 && (
        <div className="absolute left-0 right-0 top-full z-20 mt-1 max-h-56 overflow-y-auto rounded-md border border-slate-200 bg-white shadow-lg">
          {query.isFetching ? (
            <div className="flex items-center justify-center py-4"><Spinner className="h-4 w-4" /></div>
          ) : items.length === 0 ? (
            <div className="px-3 py-2 text-sm text-slate-500">No matches</div>
          ) : (
            items.slice(0, 8).map((item) => {
              const id = item.id
              const name = kind === 'account'
                ? (item as Account).name
                : `${(item as Contact).first_name} ${(item as Contact).last_name}`.trim()
              return (
                <button
                  key={id}
                  type="button"
                  onClick={() => {
                    onSelect(id, name)
                    setTerm('')
                  }}
                  className="block w-full px-3 py-2 text-left text-sm hover:bg-slate-50"
                >
                  {name}
                </button>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}

function FilterPill({ label, onClear }: { label: string; onClear: () => void }) {
  return (
    <span className="inline-flex h-8 items-center gap-1 rounded-full border border-slate-200 bg-white px-3 text-sm text-slate-700">
      {label}
      <button type="button" onClick={onClear} aria-label={`Clear ${label}`} className="rounded-full p-0.5 hover:bg-slate-100">
        <X className="h-3 w-3" />
      </button>
    </span>
  )
}

function ResultSection<T extends SearchItem>({
  title,
  icon: Icon,
  items,
  renderItem,
}: {
  title: string
  icon: React.ElementType
  items: T[]
  renderItem: (item: T) => React.ReactNode
}) {
  if (!items.length) return null
  return (
    <section>
      <div className="mb-3 flex items-center gap-2">
        <Icon className="h-4 w-4 text-slate-400" />
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">{title}</h2>
        <span className="text-xs text-slate-400">({items.length})</span>
      </div>
      <div className="space-y-1">{items.map((item) => renderItem(item))}</div>
    </section>
  )
}

function RelationshipMeta({
  relationshipType,
  account,
  contact,
}: {
  relationshipType?: string
  account?: SearchRelationship
  contact?: SearchRelationship
}) {
  const parts = [relationshipLabel(relationshipType), account?.name, contact?.name].filter(Boolean)
  if (!parts.length) return null
  return <p className="mt-1 text-xs text-slate-500">{parts.join(' · ')}</p>
}

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const [presets, setPresets] = useState<SearchPreset[]>(readPresets)
  const [presetName, setPresetName] = useState('')

  const q = searchParams.get('q') ?? ''
  const entityType = (searchParams.get('entity_type') ?? '') as SearchEntityType | ''
  const accountId = searchParams.get('account_id') ?? ''
  const accountName = searchParams.get('account_name') ?? ''
  const contactId = searchParams.get('contact_id') ?? ''
  const contactName = searchParams.get('contact_name') ?? ''
  const relationshipType = searchParams.get('relationship_type') ?? ''
  const scopeLabel = accountName || contactName || 'this record'
  const hasScopedSearch = !!(accountId || contactId)

  const filters: SearchFilters = useMemo(() => ({
    q,
    entity_type: entityType || undefined,
    account_id: accountId || undefined,
    contact_id: contactId || undefined,
    relationship_type: relationshipType || undefined,
  }), [accountId, contactId, entityType, q, relationshipType])

  const { data, isLoading, isError } = useQuery({
    queryKey: ['search', filters],
    queryFn: () => searchApi.search(filters),
    enabled: q.trim().length > 0,
    staleTime: 30_000,
  })

  const contacts = data?.contacts ?? []
  const accounts = data?.accounts ?? []
  const deals = data?.deals ?? []
  const tickets = data?.tickets ?? []
  const totalResults = resultCount(data)

  const applyFilters = (patch: Partial<SearchFilters> & { accountName?: string; contactName?: string }) => {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      updateParam(next, 'entity_type', patch.entity_type)
      updateParam(next, 'relationship_type', patch.relationship_type)
      updateParam(next, 'account_id', patch.account_id)
      updateParam(next, 'account_name', patch.accountName)
      updateParam(next, 'contact_id', patch.contact_id)
      updateParam(next, 'contact_name', patch.contactName)
      return next
    })
  }

  const clearScopedEntity = (kind: ScopedPickerKind) => {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      next.delete(`${kind}_id`)
      next.delete(`${kind}_name`)
      return next
    })
  }

  const savePreset = () => {
    const name = presetName.trim()
    if (!name) return
    const next = [
      ...presets.filter((preset) => preset.name !== name),
      {
        id: crypto.randomUUID(),
        name,
        filters: {
          entity_type: entityType || undefined,
          account_id: accountId || undefined,
          contact_id: contactId || undefined,
          relationship_type: relationshipType || undefined,
        },
        accountName,
        contactName,
      },
    ]
    setPresets(next)
    writePresets(next)
    setPresetName('')
  }

  const renderGroupedResults = () => {
    if (!hasScopedSearch) return null
    const allItems: SearchItem[] = [...contacts, ...accounts, ...deals, ...tickets]
    const groups = new Map<string, SearchItem[]>()
    allItems.forEach((item) => {
      const key = relationshipKey(item)
      groups.set(key, [...(groups.get(key) ?? []), item])
    })
    return Array.from(groups.values()).map((items) => (
      <section key={relationshipKey(items[0])} className="rounded-lg border border-slate-200 bg-white p-4">
        <h2 className="mb-3 text-sm font-semibold text-slate-700">{relationshipTitle(items[0], scopeLabel)}</h2>
        <div className="space-y-2">
          {items.map((item) => (
            <button
              key={`${'subject' in item ? 'ticket' : 'title' in item ? 'deal' : 'first_name' in item ? 'contact' : 'account'}-${item.id}`}
              type="button"
              onClick={() => {
                if ('first_name' in item) navigate(`/contacts?openId=${item.id}`)
                else if ('name' in item) navigate(`/accounts?openId=${item.id}`)
                else if ('subject' in item) navigate(`/tickets?openId=${item.id}`)
                else navigate(`/deals?openId=${item.id}`)
              }}
              className="block w-full rounded-md border border-slate-100 px-3 py-2 text-left hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]"
            >
              <p className="text-sm font-medium text-slate-900">
                {'first_name' in item ? `${item.first_name} ${item.last_name}` : 'name' in item ? item.name : 'subject' in item ? item.subject : item.title}
              </p>
              <RelationshipMeta relationshipType={item.relationship_type} account={'related_account' in item ? item.related_account : undefined} contact={'related_contact' in item ? item.related_contact : undefined} />
            </button>
          ))}
        </div>
      </section>
    ))
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <h1 className="flex items-center gap-2 text-2xl font-bold text-slate-900">
          <Search className="h-6 w-6 text-slate-400" />
          Search results
        </h1>
        {q && (
          <p className="mt-1 text-sm text-slate-500">
            {isLoading ? 'Searching...' : `${totalResults} result${totalResults !== 1 ? 's' : ''} for "${q}"`}
          </p>
        )}
      </div>

      <div className="space-y-3 rounded-lg border border-slate-200 bg-slate-50 p-3">
        <div className="flex flex-wrap gap-2">
          {ENTITY_OPTIONS.map(({ value, label, icon: Icon }) => (
            <button
              key={label}
              type="button"
              onClick={() => applyFilters({ entity_type: value || undefined })}
              className={`inline-flex h-9 items-center gap-1.5 rounded-md border px-3 text-sm font-medium ${entityType === value ? 'border-[var(--color-primary)] bg-white text-[var(--color-primary)]' : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-100'}`}
            >
              <Icon className="h-4 w-4" />
              {label}
            </button>
          ))}
          <select
            value={relationshipType}
            onChange={(e) => applyFilters({ relationship_type: e.target.value || undefined })}
            className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700"
          >
            {RELATIONSHIP_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>{option.label}</option>
            ))}
          </select>
          <ScopedPicker
            kind="account"
            selectedId={accountId}
            selectedName={accountName}
            onSelect={(id, name) => applyFilters({ account_id: id, accountName: name, contact_id: undefined, contactName: undefined })}
            onClear={() => clearScopedEntity('account')}
          />
          <ScopedPicker
            kind="contact"
            selectedId={contactId}
            selectedName={contactName}
            onSelect={(id, name) => applyFilters({ contact_id: id, contactName: name, account_id: undefined, accountName: undefined })}
            onClear={() => clearScopedEntity('contact')}
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {relationshipType && <FilterPill label={`Relationship: ${relationshipLabel(relationshipType)}`} onClear={() => applyFilters({ relationship_type: undefined })} />}
          {entityType && <FilterPill label={`Type: ${relationshipLabel(entityType)}`} onClear={() => applyFilters({ entity_type: undefined })} />}
          <div className="ml-auto flex items-center gap-2">
            {presets.map((preset) => (
              <button
                key={preset.id}
                type="button"
                onClick={() => applyFilters({ ...preset.filters, accountName: preset.accountName, contactName: preset.contactName })}
                className="inline-flex h-8 items-center gap-1 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 hover:bg-slate-100"
              >
                <Bookmark className="h-3.5 w-3.5" />
                {preset.name}
              </button>
            ))}
            <Input value={presetName} onChange={(e) => setPresetName(e.target.value)} placeholder="Preset name" className="h-8 w-32 text-xs" />
            <button type="button" onClick={savePreset} className="h-8 rounded-md bg-slate-900 px-3 text-xs font-medium text-white disabled:opacity-40" disabled={!presetName.trim()}>
              Save preset
            </button>
          </div>
        </div>
      </div>

      {!q && <p className="text-sm text-slate-400">Enter a search query in the bar above to get started.</p>}
      {q && isLoading && <div className="flex justify-center py-16"><Spinner /></div>}
      {q && isError && <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">Something went wrong while searching. Please try again.</div>}
      {q && !isLoading && !isError && totalResults === 0 && (
        <div className="py-16 text-center">
          <Search className="mx-auto mb-3 h-10 w-10 text-slate-300" />
          <p className="font-medium text-slate-500">No results found for "{q}"</p>
          <p className="mt-1 text-sm text-slate-400">Try a different search term or remove a filter.</p>
        </div>
      )}

      {!isLoading && !isError && totalResults > 0 && (
        <div className="space-y-8">
          {hasScopedSearch ? renderGroupedResults() : (
            <>
              <ResultSection title="Contacts" icon={Users} items={contacts} renderItem={(c) => (
                <button key={c.id} onClick={() => navigate(`/contacts?openId=${c.id}`)} className="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-left hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]">
                  <p className="text-sm font-medium text-slate-900">{c.first_name} {c.last_name}</p>
                  <p className="text-xs text-slate-500">{c.email}</p>
                  <RelationshipMeta relationshipType={c.relationship_type} account={c.related_account} />
                </button>
              )} />
              <ResultSection title="Accounts" icon={Building2} items={accounts} renderItem={(a) => (
                <button key={a.id} onClick={() => navigate(`/accounts?openId=${a.id}`)} className="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-left hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]">
                  <p className="text-sm font-medium text-slate-900">{a.name}</p>
                  <p className="text-xs text-slate-500">{[a.industry, a.domain].filter(Boolean).join(' · ') || 'Account'}</p>
                  <RelationshipMeta relationshipType={a.relationship_type} account={a.related_account} contact={a.related_contact} />
                </button>
              )} />
              <ResultSection title="Deals" icon={TrendingUp} items={deals} renderItem={(d) => (
                <button key={d.id} onClick={() => navigate(`/deals?openId=${d.id}`)} className="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-left hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]">
                  <p className="text-sm font-medium text-slate-900">{d.title}</p>
                  <p className="text-xs text-slate-500">{d.stage.replace('_', ' ')}</p>
                  <RelationshipMeta relationshipType={d.relationship_type} account={d.related_account} contact={d.related_contact} />
                </button>
              )} />
              <ResultSection title="Tickets" icon={LifeBuoy} items={tickets} renderItem={(t) => (
                <button key={t.id} onClick={() => navigate(`/tickets?openId=${t.id}`)} className="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-left hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]">
                  <p className="text-sm font-medium text-slate-900">{t.subject}</p>
                  <p className="text-xs capitalize text-slate-500">{t.status} · {t.priority}</p>
                  <RelationshipMeta relationshipType={t.relationship_type} account={t.related_account} contact={t.related_contact} />
                </button>
              )} />
            </>
          )}
        </div>
      )}
    </div>
  )
}
