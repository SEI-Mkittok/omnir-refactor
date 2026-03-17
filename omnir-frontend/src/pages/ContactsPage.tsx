import { useState, useCallback } from 'react'
import { Plus, Users } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useContacts, useContact, useDeleteContact } from '@/hooks/useContacts'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { formatDate } from '@/lib/utils'
import { stageBadgeVariant, stageLabel } from '@/components/omnir/ContactCard'
import type { Contact, ContactStage } from '@/api/types'

const STAGE_OPTIONS = [
  { label: 'Lead', value: 'lead' },
  { label: 'Prospect', value: 'prospect' },
  { label: 'Customer', value: 'customer' },
  { label: 'Churned', value: 'churned' },
]

const SORT_OPTIONS = [
  { label: 'Name A–Z', value: 'last_name:asc' },
  { label: 'Name Z–A', value: 'last_name:desc' },
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
]

// ---- Contact detail panel ----

function ContactDetail({ contactId, onClose }: { contactId: string; onClose: () => void }) {
  const { data: contact, isLoading } = useContact(contactId)
  const deleteContact = useDeleteContact()

  if (isLoading) {
    return (
      <SidePanel open title="Contact" onClose={onClose}>
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      </SidePanel>
    )
  }

  if (!contact) return null

  return (
    <SidePanel
      open
      title={`${contact.first_name} ${contact.last_name}`}
      onClose={onClose}
      width="md"
      actions={
        <Button
          variant="destructive"
          size="sm"
          onClick={async () => {
            if (confirm('Delete this contact?')) {
              await deleteContact.mutateAsync(contact.id)
              onClose()
            }
          }}
        >
          Delete
        </Button>
      }
    >
      <div className="space-y-5">
        {/* Stage badge */}
        <div>
          <Badge variant={stageBadgeVariant[contact.stage]} className="text-sm px-3 py-1">
            {stageLabel[contact.stage]}
          </Badge>
        </div>

        {/* Core fields */}
        <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
          {contact.title && (
            <>
              <dt className="font-medium text-slate-500">Title</dt>
              <dd className="text-slate-900">{contact.title}</dd>
            </>
          )}
          {contact.department && (
            <>
              <dt className="font-medium text-slate-500">Department</dt>
              <dd className="text-slate-900">{contact.department}</dd>
            </>
          )}
          <dt className="font-medium text-slate-500">Email</dt>
          <dd className="text-slate-900 break-all">{contact.email}</dd>

          {contact.phone && (
            <>
              <dt className="font-medium text-slate-500">Phone</dt>
              <dd className="text-slate-900">{contact.phone}</dd>
            </>
          )}
          {contact.account?.name && (
            <>
              <dt className="font-medium text-slate-500">Account</dt>
              <dd className="text-slate-900">{contact.account.name}</dd>
            </>
          )}
          <dt className="font-medium text-slate-500">Created</dt>
          <dd className="text-slate-900">{formatDate(contact.created_at)}</dd>
        </dl>

        {/* Tags */}
        {contact.tags && contact.tags.length > 0 && (
          <div>
            <p className="mb-2 text-sm font-medium text-slate-500">Tags</p>
            <div className="flex flex-wrap gap-1.5">
              {contact.tags.map((tag) => (
                <Badge key={tag} variant="default">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Linked deals */}
        {contact.deals && contact.deals.length > 0 && (
          <div>
            <p className="mb-2 text-sm font-medium text-slate-500">Deals</p>
            <ul className="space-y-1.5">
              {contact.deals.map((d) => (
                <li key={d.id} className="rounded-md border border-slate-200 px-3 py-2 text-sm">
                  <span className="font-medium text-slate-900">{d.title}</span>
                  <span className="ml-2 text-slate-500">{d.stage}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </SidePanel>
  )
}

// ---- Main page ----

export function ContactsPage() {
  const [search, setSearch] = useState('')
  const [stage, setStage] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(null)

  const debouncedSearch = useDebounce(search, 300)

  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const { data, isLoading } = useContacts({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    stage: (stage as ContactStage) || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const contacts = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback(
    (key: string) => {
      setSortKey((prev) => {
        const [prevKey, prevDir] = prev.split(':')
        if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
        return `${key}:asc`
      })
    },
    []
  )

  const columns: Column<Contact>[] = [
    {
      key: 'last_name',
      header: 'Name',
      sortable: true,
      render: (c) => (
        <span className="font-medium text-slate-900">
          {c.first_name} {c.last_name}
        </span>
      ),
    },
    {
      key: 'email',
      header: 'Email',
      render: (c) => <span className="text-slate-600">{c.email}</span>,
    },
    {
      key: 'title',
      header: 'Title',
      render: (c) => <span className="text-slate-600">{c.title ?? '—'}</span>,
    },
    {
      key: 'account',
      header: 'Account',
      render: (c) => <span className="text-slate-600">{c.account?.name ?? '—'}</span>,
    },
    {
      key: 'stage',
      header: 'Stage',
      sortable: true,
      render: (c) => (
        <Badge variant={stageBadgeVariant[c.stage]}>{stageLabel[c.stage]}</Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (c) => <span className="text-slate-500 text-xs">{formatDate(c.created_at)}</span>,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Contacts</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total` : 'Loading…'}
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4" />
          Add Contact
        </Button>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search contacts…"
        filters={[
          {
            label: 'Stage',
            value: stage,
            options: STAGE_OPTIONS,
            onChange: (v) => { setStage(v); setPage(1) },
          },
          {
            label: 'Sort',
            value: sortKey,
            options: SORT_OPTIONS,
            onChange: (v) => { setSortKey(v); setPage(1) },
          },
        ]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={contacts}
          isLoading={isLoading}
          sortBy={sortBy}
          sortDir={sortDir}
          onSort={handleSort}
          onRowClick={(c) => setSelectedId(c.id)}
          emptyIcon={Users}
          emptyTitle="No contacts found"
          emptyDescription="Try adjusting your search or filters."
          keyExtractor={(c) => c.id}
        />

        {/* Pagination */}
        {meta && meta.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3">
            <p className="text-sm text-slate-500">
              Page {meta.page} of {meta.total_pages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= meta.total_pages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Detail panel */}
      {selectedId && (
        <ContactDetail contactId={selectedId} onClose={() => setSelectedId(null)} />
      )}
    </div>
  )
}
