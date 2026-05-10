import { useState } from 'react'
import { Link, useParams, useSearchParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Search, FileText, ChevronRight } from 'lucide-react'
import { usePublicKbArticles } from '@/hooks/useKB'
import { useDebounce } from '@/hooks/useDebounce'

export function HelpSearchPage() {
  const { orgSlug = '' } = useParams<{ orgSlug: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const initialQ = searchParams.get('q') ?? ''

  const [searchInput, setSearchInput] = useState(initialQ)
  const debouncedQ = useDebounce(searchInput, 300)

  const { data: results = [], isLoading } = usePublicKbArticles(orgSlug, debouncedQ || undefined)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    navigate(`/help/${orgSlug}/search?q=${encodeURIComponent(searchInput.trim())}`, { replace: true })
  }

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      {/* Top bar */}
      <div className="bg-[#1B3A4B] text-white px-4 py-4">
        <div className="mx-auto max-w-3xl flex items-center gap-3">
          <Link
            to={`/help/${orgSlug}`}
            className="flex items-center gap-1.5 text-white/70 hover:text-white text-sm transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Help Center
          </Link>
          <ChevronRight className="h-4 w-4 text-white/40" />
          <span className="text-white text-sm font-medium">Search</span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl px-4 py-10 space-y-6">
        <form onSubmit={handleSearch} className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-400" />
          <input
            autoFocus
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search articles…"
            className="w-full rounded-xl border border-slate-300 bg-white pl-12 pr-4 py-4 text-slate-900 text-base shadow-sm placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
          />
        </form>

        {debouncedQ && (
          <p className="text-sm text-slate-500">
            {isLoading
              ? 'Searching…'
              : `${results.length} result${results.length !== 1 ? 's' : ''} for "${debouncedQ}"`}
          </p>
        )}

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 rounded-xl bg-slate-200 animate-pulse" />
            ))}
          </div>
        ) : results.length === 0 && debouncedQ ? (
          <div className="rounded-xl border border-slate-200 bg-white p-10 text-center">
            <Search className="mx-auto h-8 w-8 text-slate-300 mb-3" />
            <p className="text-sm font-medium text-slate-700 mb-1">No results found</p>
            <p className="text-sm text-slate-500">Try different keywords or browse the categories.</p>
            <Link
              to={`/help/${orgSlug}`}
              className="inline-block mt-4 text-sm text-[var(--color-primary)] hover:underline"
            >
              Browse Help Center
            </Link>
          </div>
        ) : results.length > 0 ? (
          <div className="rounded-xl border border-slate-200 bg-white divide-y divide-slate-100 overflow-hidden">
            {results.map((article) => (
              <Link
                key={article.id}
                to={`/help/${orgSlug}/a/${article.slug}`}
                className="flex flex-col sm:flex-row sm:items-center justify-between px-5 py-4 hover:bg-slate-50 transition-colors group gap-2"
              >
                <div className="flex items-start gap-3 min-w-0">
                  <FileText className="h-4 w-4 text-slate-400 shrink-0 mt-0.5" />
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-slate-800 group-hover:text-[var(--color-primary)] transition-colors">
                      {article.title}
                    </p>
                    {article.excerpt && (
                      <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{article.excerpt}</p>
                    )}
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-300 group-hover:text-[var(--color-primary)] shrink-0 ml-auto" />
              </Link>
            ))}
          </div>
        ) : null}
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Powered by Omnir CRM
      </div>
    </div>
  )
}
