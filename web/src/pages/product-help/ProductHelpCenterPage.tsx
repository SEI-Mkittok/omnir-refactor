import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { BookOpen, ChevronRight, FileText, Search } from 'lucide-react'
import { useProductHelpArticles, useProductHelpCategories } from '@/hooks/useProductHelp'

export function ProductHelpCenterPage() {
  const navigate = useNavigate()
  const [searchInput, setSearchInput] = useState('')
  const { data: categories = [], isLoading: catsLoading } = useProductHelpCategories()
  const { data: featuredArticles = [], isLoading: articlesLoading } = useProductHelpArticles()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchInput.trim()) {
      navigate(`/product-help/search?q=${encodeURIComponent(searchInput.trim())}`)
    }
  }

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      <div className="bg-[#1B3A4B] text-white">
        <div className="mx-auto max-w-3xl px-4 py-16 text-center">
          <div className="mb-4 flex justify-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/10">
              <BookOpen className="h-6 w-6 text-white" />
            </div>
          </div>
          <h1 className="mb-2 text-4xl font-bold">Omnir Help</h1>
          <p className="mb-8 text-lg text-white/70">Product guides, workflows, and answers for using Omnir CRM.</p>
          <form onSubmit={handleSearch} className="relative">
            <Search className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Search product help..."
              className="w-full rounded-xl border-0 bg-white py-4 pl-12 pr-24 text-base text-slate-900 shadow-lg placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            />
            <button
              type="submit"
              className="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg bg-[#1B3A4B] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#243f52]"
            >
              Search
            </button>
          </form>
        </div>
      </div>

      <div className="mx-auto max-w-5xl space-y-12 px-4 py-12">
        <section>
          <h2 className="mb-6 text-xl font-bold text-slate-900">Browse by Category</h2>
          {catsLoading ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-28 animate-pulse rounded-xl bg-slate-200" />
              ))}
            </div>
          ) : categories.length === 0 ? (
            <p className="text-sm text-slate-500">No product-help categories have been synced yet.</p>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {categories.map((cat) => (
                <Link
                  key={cat.id}
                  to={`/product-help/c/${cat.slug}`}
                  className="group flex flex-col gap-2 rounded-xl border border-slate-200 bg-white p-5 transition-all hover:border-[var(--color-primary)] hover:shadow-sm"
                >
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-[var(--color-primary-light)]">
                    <FileText className="h-5 w-5 text-[var(--color-primary)]" />
                  </div>
                  <div>
                    <p className="font-semibold text-slate-900 transition-colors group-hover:text-[var(--color-primary)]">{cat.name}</p>
                    <p className="mt-0.5 text-sm text-slate-500">
                      {cat.article_count ?? 0} article{cat.article_count !== 1 ? 's' : ''}
                    </p>
                  </div>
                  <ChevronRight className="mt-auto h-4 w-4 self-end text-slate-300 group-hover:text-[var(--color-primary)]" />
                </Link>
              ))}
            </div>
          )}
        </section>

        <section>
          <h2 className="mb-6 text-xl font-bold text-slate-900">Popular Guides</h2>
          {articlesLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-14 animate-pulse rounded-xl bg-slate-200" />
              ))}
            </div>
          ) : featuredArticles.length === 0 ? (
            <p className="text-sm text-slate-500">No product-help articles have been synced yet.</p>
          ) : (
            <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
              {featuredArticles.slice(0, 10).map((article) => (
                <Link
                  key={article.id}
                  to={`/product-help/a/${article.slug}`}
                  className="group flex items-center justify-between border-b border-slate-100 px-5 py-4 transition-colors last:border-b-0 hover:bg-slate-50"
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <FileText className="h-4 w-4 shrink-0 text-slate-400" />
                    <div className="min-w-0">
                      <span className="block truncate text-sm font-medium text-slate-800 transition-colors group-hover:text-[var(--color-primary)]">
                        {article.title}
                      </span>
                      {article.excerpt && <span className="block truncate text-xs text-slate-500">{article.excerpt}</span>}
                    </div>
                  </div>
                  <ChevronRight className="ml-3 h-4 w-4 shrink-0 text-slate-300 group-hover:text-[var(--color-primary)]" />
                </Link>
              ))}
            </div>
          )}
        </section>
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Omnir Product Help
      </div>
    </div>
  )
}
