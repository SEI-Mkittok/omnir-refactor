import { useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus, Building2, Globe, Users, TrendingUp } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useAccounts, useAccount, useAccountContacts, useAccountDeals, useDeleteAccount } from '@/hooks/useAccounts'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { SidePanel } from '@/components/ui/SidePanel'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { CustomFieldDisplaySection } from '@/components/omnir/CustomFieldRenderer'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { formatDate, formatCurrency } from '@/lib/utils'
import type { Account, CustomFieldValues } from '@/api/types'

const INDUSTRY_OPTIONS = [
  { label: 'Technology', value: 'Technology' },
  { label: 'Finance', value: 'Finance' },
  { label: 'Healthcare', value: 'Healthcare' },
  { label: 'Retail', value: 'Retail' },
  { label: 'Manufacturing', value: 'Manufacturing' },
  { label: 'Education', value: 'Education' },
  { label: 'Other', value: 'Other' },
]

const SORT_OPTIONS = [
  { label: 'Name A–Z', value: 'name:asc' },
  { label: 'Name Z–A', value: 'name:desc' },
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
]

// ---- Account detail panel ----

function AccountDetail({ accountId, onClose }: { accountId: string; onClose: () => void }) {
  const { data: account, isLoading } = useAccount(accountId)
  const { data: contacts, isLoading: contactsLoading } = useAccountContacts(accountId)
  const { data: deals, isLoading: dealsLoading } = useAccountDeals(accountId)
  const deleteAccount = useDeleteAccount()
  const { data: customFields = [] } = useCustomFieldDefinitions('account')

  if (isLoading) {
    return (
      <SidePanel open title="Account" onClose={onClose}>
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      </SidePanel>
    )
  }

  if (!account) return null

  return (
    <SidePanel
      open
      title={account.name}
      onClose={onClose}
      width="lg"
      actions={
        <Button
          variant="destructive"
          size="sm"
          onClick={async () => {
            if (confirm('Delete this account?')) {
              await deleteAccount.mutateAsync(account.id)
              onClose()
            }
          }}
        >
          Delete
        </Button>
      }
    >
      <div className="space-y-6">
        {/* Core fields */}
        <div>
          <h3 className="mb-3 text-sm font-semibold text-slate-700 uppercase tracking-wide">Details</h3>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-3 text-sm">
            {account.industry && (
              <>
                <dt className="font-medium text-slate-500">Industry</dt>
                <dd className="text-slate-900">{account.industry}</dd>
              </>
            )}
            {account.size && (
              <>
                <dt className="font-medium text-slate-500">Size</dt>
                <dd className="text-slate-900">{account.size} employees</dd>
              </>
            )}
            {account.domain && (
              <>
                <dt className="font-medium text-slate-500">Domain</dt>
                <dd>
                  <a
                    href={`https://${account.domain}`}
                    target="_blank"
                    rel="noreferrer"
                    className="text-indigo-600 hover:underline flex items-center gap-1"
                  >
                    <Globe className="h-3.5 w-3.5" />
                    {account.domain}
                  </a>
                </dd>
              </>
            )}
            {account.phone && (
              <>
                <dt className="font-medium text-slate-500">Phone</dt>
                <dd className="text-slate-900">{account.phone}</dd>
              </>
            )}
            {account.address && (
              <>
                <dt className="font-medium text-slate-500">Address</dt>
                <dd className="text-slate-900">{account.address}</dd>
              </>
            )}
            <dt className="font-medium text-slate-500">Created</dt>
            <dd className="text-slate-900">{formatDate(account.created_at)}</dd>
          </dl>
        </div>

        {/* Custom Fields */}
        {customFields.length > 0 && account.custom_fields && (
          <div className="rounded-lg border border-slate-200 px-4 py-3 space-y-2">
            <CustomFieldDisplaySection
              fields={customFields}
              values={account.custom_fields as CustomFieldValues}
            />
          </div>
        )}

        {/* Linked contacts */}
        <div>
          <h3 className="mb-3 text-sm font-semibold text-slate-700 uppercase tracking-wide flex items-center gap-2">
            <Users className="h-4 w-4" />
            Contacts
            {contacts && (
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-normal text-slate-600">
                {contacts.length}
              </span>
            )}
          </h3>
          {contactsLoading ? (
            <Spinner />
          ) : contacts && contacts.length > 0 ? (
            <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200">
              {contacts.map((c) => (
                <li key={c.id} className="px-3 py-2.5 text-sm">
                  <span className="font-medium text-slate-900">
                    {c.first_name} {c.last_name}
                  </span>
                  {c.title && <span className="ml-2 text-slate-500">{c.title}</span>}
                  <div className="text-xs text-slate-400 mt-0.5">{c.email}</div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No contacts linked yet.</p>
          )}
        </div>

        {/* Linked deals */}
        <div>
          <h3 className="mb-3 text-sm font-semibold text-slate-700 uppercase tracking-wide flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Deals
            {deals && (
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-normal text-slate-600">
                {deals.length}
              </span>
            )}
          </h3>
          {dealsLoading ? (
            <Spinner />
          ) : deals && deals.length > 0 ? (
            <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200">
              {deals.map((d) => (
                <li key={d.id} className="flex items-center justify-between px-3 py-2.5 text-sm">
                  <span className="font-medium text-slate-900">{d.title}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-indigo-600 font-semibold">
                      {formatCurrency(d.value, d.currency)}
                    </span>
                    <Badge variant="default">{d.stage}</Badge>
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No deals linked yet.</p>
          )}
        </div>
      </div>
    </SidePanel>
  )
}

// ---- Main page ----

export function AccountsPage() {
  const [searchParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const [industry, setIndustry] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [selectedId, setSelectedId] = useState<string | null>(searchParams.get('openId'))

  const debouncedSearch = useDebounce(search, 300)
  const [sortBy, sortDir] = sortKey.split(':') as [string, 'asc' | 'desc']

  const { data, isLoading } = useAccounts({
    page,
    per_page: 20,
    search: debouncedSearch || undefined,
    industry: industry || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  })

  const accounts = data?.data ?? []
  const meta = data?.meta

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  const columns: Column<Account>[] = [
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (a) => (
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-100">
            <Building2 className="h-4 w-4 text-slate-500" />
          </div>
          <span className="font-medium text-slate-900">{a.name}</span>
        </div>
      ),
    },
    {
      key: 'domain',
      header: 'Domain',
      hideOnMobile: true,
      render: (a) => a.domain ? (
        <span className="flex items-center gap-1 text-slate-600">
          <Globe className="h-3.5 w-3.5" />
          {a.domain}
        </span>
      ) : <span className="text-slate-400">—</span>,
    },
    {
      key: 'industry',
      header: 'Industry',
      sortable: true,
      render: (a) => <span className="text-slate-600">{a.industry ?? '—'}</span>,
    },
    {
      key: 'size',
      header: 'Size',
      hideOnMobile: true,
      render: (a) => a.size ? (
        <span className="flex items-center gap-1 text-slate-600">
          <Users className="h-3.5 w-3.5" />
          {a.size}
        </span>
      ) : <span className="text-slate-400">—</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      hideOnMobile: true,
      render: (a) => <span className="text-slate-500 text-xs">{formatDate(a.created_at)}</span>,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Accounts</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {meta ? `${meta.total} total` : 'Loading…'}
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4" />
          Add Account
        </Button>
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search accounts…"
        filters={[
          {
            label: 'Industry',
            value: industry,
            options: INDUSTRY_OPTIONS,
            onChange: (v) => { setIndustry(v); setPage(1) },
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
          data={accounts}
          isLoading={isLoading}
          sortBy={sortBy}
          sortDir={sortDir}
          onSort={handleSort}
          onRowClick={(a) => setSelectedId(a.id)}
          emptyIcon={Building2}
          emptyTitle="No accounts found"
          emptyDescription="Try adjusting your search or filters."
          keyExtractor={(a) => a.id}
        />

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

      {selectedId && (
        <AccountDetail accountId={selectedId} onClose={() => setSelectedId(null)} />
      )}
    </div>
  )
}
