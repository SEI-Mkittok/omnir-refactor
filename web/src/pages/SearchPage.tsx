import { useSearchParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Users, Building2, TrendingUp, Search, LifeBuoy } from 'lucide-react'
import { searchApi } from '@/api/search'
import { Spinner } from '@/components/ui/Spinner'
import type { Contact, Account, Deal, Ticket } from '@/api/types'

function ResultSection<T>({
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
      <div className="flex items-center gap-2 mb-3">
        <Icon className="h-4 w-4 text-slate-400" />
        <h2 className="text-sm font-semibold text-slate-500 uppercase tracking-wide">{title}</h2>
        <span className="text-xs text-slate-400">({items.length})</span>
      </div>
      <div className="space-y-1">{items.map((item) => renderItem(item))}</div>
    </section>
  )
}

export function SearchPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const q = searchParams.get('q') ?? ''

  const { data, isLoading, isError } = useQuery({
    queryKey: ['search', q],
    queryFn: () => searchApi.search(q),
    enabled: q.trim().length > 0,
    staleTime: 30_000,
  })

  const contacts = data?.contacts ?? []
  const accounts = data?.accounts ?? []
  const deals = data?.deals ?? []
  const tickets = data?.tickets ?? []
  const totalResults = contacts.length + accounts.length + deals.length + tickets.length

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900 flex items-center gap-2">
          <Search className="h-6 w-6 text-slate-400" />
          Search results
        </h1>
        {q && (
          <p className="mt-1 text-sm text-slate-500">
            {isLoading ? 'Searching…' : `${totalResults} result${totalResults !== 1 ? 's' : ''} for "${q}"`}
          </p>
        )}
      </div>

      {/* States */}
      {!q && (
        <p className="text-slate-400 text-sm">Enter a search query in the bar above to get started.</p>
      )}

      {q && isLoading && (
        <div className="flex items-center justify-center py-16">
          <Spinner />
        </div>
      )}

      {q && isError && (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          Something went wrong while searching. Please try again.
        </div>
      )}

      {q && !isLoading && !isError && totalResults === 0 && (
        <div className="py-16 text-center">
          <Search className="mx-auto h-10 w-10 text-slate-300 mb-3" />
          <p className="text-slate-500 font-medium">No results found for "{q}"</p>
          <p className="text-sm text-slate-400 mt-1">Try a different search term.</p>
        </div>
      )}

      {/* Results */}
      {!isLoading && !isError && totalResults > 0 && (
        <div className="space-y-8">
          <ResultSection<Contact>
            title="Contacts"
            icon={Users}
            items={contacts}
            renderItem={(c) => (
              <button
                key={c.id}
                onClick={() => navigate(`/contacts?openId=${c.id}`)}
                className="w-full text-left flex items-start gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors group"
              >
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary-light)] text-[var(--color-primary)] text-xs font-semibold group-hover:bg-[var(--color-primary-light)]">
                  {c.first_name[0]}{c.last_name[0]}
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-900">{c.first_name} {c.last_name}</p>
                  <p className="text-xs text-slate-500 truncate">{c.email}{c.title ? ` · ${c.title}` : ''}</p>
                </div>
              </button>
            )}
          />

          <ResultSection<Account>
            title="Accounts"
            icon={Building2}
            items={accounts}
            renderItem={(a) => (
              <button
                key={a.id}
                onClick={() => navigate(`/accounts?openId=${a.id}`)}
                className="w-full text-left flex items-start gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors group"
              >
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-blue-600">
                  <Building2 className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-900">{a.name}</p>
                  <p className="text-xs text-slate-500 truncate">
                    {[a.industry, a.domain].filter(Boolean).join(' · ') || 'Account'}
                  </p>
                </div>
              </button>
            )}
          />

          <ResultSection<Deal>
            title="Deals"
            icon={TrendingUp}
            items={deals}
            renderItem={(d) => (
              <button
                key={d.id}
                onClick={() => navigate(`/deals?openId=${d.id}`)}
                className="w-full text-left flex items-start gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors group"
              >
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
                  <TrendingUp className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-900">{d.title}</p>
                  <p className="text-xs text-slate-500 truncate">
                    {d.stage.replace('_', ' ')}{d.value_cents ? ` · $${(d.value_cents / 100).toLocaleString()}` : ''}
                  </p>
                </div>
              </button>
            )}
          />

          <ResultSection<Ticket>
            title="Tickets"
            icon={LifeBuoy}
            items={tickets}
            renderItem={(t) => (
              <button
                key={t.id}
                onClick={() => navigate(`/tickets?openId=${t.id}`)}
                className="w-full text-left flex items-start gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors group"
              >
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-violet-100 text-violet-600">
                  <LifeBuoy className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-900">{t.subject}</p>
                  <p className="text-xs text-slate-500 truncate capitalize">
                    {t.status} · {t.priority}
                  </p>
                </div>
              </button>
            )}
          />
        </div>
      )}
    </div>
  )
}
