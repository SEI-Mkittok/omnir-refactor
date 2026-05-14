import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft, ChevronRight, FileText, Search } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useProductHelpSearch } from '@/hooks/useProductHelp'

export function ProductHelpSearchPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const initialQ = searchParams.get('q') ?? ''
  const [searchInput, setSearchInput] = useState(initialQ)
  const debouncedQ = useDebounce(searchInput, 300)
  const { data: results = [], isLoading } = useProductHelpSearch(debouncedQ)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    navigate(`/product-help/search?q=${encodeURIComponent(searchInput.trim())}`, { replace: true })
  }

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      <div className="bg-[#1B3A4B] px-4 py-4 text-white">
        <div className="mx-auto flex max-w-3xl items-center gap-3">
          <Link to="/product-help" className="flex items-center gap-1.5 text-sm text-white/70 transition-colors hover:text-white">
            <ArrowLeft className="h-4 w-4" />
            Omnir Help
          </Link>
          <ChevronRight className="h-4 w-4 text-white/40" />
          <span className="text-sm font-medium text-white">Search</span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl space-y-6 px-4 py-10">
        <form onSubmit={handleSearch} className="relative">
          <Search className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
          <input
            autoFocus
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search product help..."
            className="w-full rounded-xl border border-slate-300 bg-white py-4 pl-12 pr-4 text-base text-slate-900 shadow-sm placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
          />
        </form>

        {debouncedQ && (
          <p className="text-sm text-slate-500">
            {isLoading
              ? 'Searching...'
              : `${results.length} result${results.length !== 1 ? 's' : ''} for "${debouncedQ}"`}
          </p>
        )}

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 animate-pulse rounded-xl bg-slate-200" />
            ))}
          </div>
        ) : results.length === 0 && debouncedQ ? (
          <div className="rounded-xl border border-slate-200 bg-white p-10 text-center">
            <Search className="mx-auto mb-3 h-8 w-8 text-slate-300" />
            <p className="mb-1 text-sm font-medium text-slate-700">No results found</p>
            <p className="text-sm text-slate-500">Try different keywords or browse the categories.</p>
            <Link to="/product-help" className="mt-4 inline-block text-sm text-[var(--color-primary)] hover:underline">
              Browse Omnir Help
            </Link>
          </div>
        ) : results.length > 0 ? (
          <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
            {results.map((article) => (
              <Link
                key={article.id}
                to={`/product-help/a/${article.slug}`}
                className="group flex flex-col justify-between gap-2 border-b border-slate-100 px-5 py-4 transition-colors last:border-b-0 hover:bg-slate-50 sm:flex-row sm:items-center"
              >
                <div className="flex min-w-0 items-start gap-3">
                  <FileText className="mt-0.5 h-4 w-4 shrink-0 text-slate-400" />
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-slate-800 transition-colors group-hover:text-[var(--color-primary)]">
                      {article.title}
                    </p>
                    {article.excerpt && <p className="mt-0.5 line-clamp-2 text-xs text-slate-500">{article.excerpt}</p>}
                  </div>
                </div>
                <ChevronRight className="ml-auto h-4 w-4 shrink-0 text-slate-300 group-hover:text-[var(--color-primary)]" />
              </Link>
            ))}
          </div>
        ) : null}
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Omnir Product Help
      </div>
    </div>
  )
}
