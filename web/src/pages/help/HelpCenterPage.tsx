import { useState } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { BookOpen, Search, ChevronRight, FileText } from 'lucide-react'
import { usePublicKbCategories, usePublicKbArticles } from '@/hooks/useKB'

export function HelpCenterPage() {
  const { orgSlug = '' } = useParams<{ orgSlug: string }>()
  const navigate = useNavigate()
  const [searchInput, setSearchInput] = useState('')

  const { data: categories = [], isLoading: catsLoading } = usePublicKbCategories(orgSlug)
  const { data: featuredArticles = [], isLoading: articlesLoading } = usePublicKbArticles(orgSlug)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchInput.trim()) {
      navigate(`/help/${orgSlug}/search?q=${encodeURIComponent(searchInput.trim())}`)
    }
  }

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      {/* Hero */}
      <div className="bg-[#1B3A4B] text-white">
        <div className="mx-auto max-w-3xl px-4 py-16 text-center">
          <div className="flex justify-center mb-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/10">
              <BookOpen className="h-6 w-6 text-white" />
            </div>
          </div>
          <h1 className="text-4xl font-bold mb-2">Help Center</h1>
          <p className="text-white/70 text-lg mb-8">Find answers to your questions</p>
          <form onSubmit={handleSearch} className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-400" />
            <input
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Search articles…"
              className="w-full rounded-xl border-0 bg-white pl-12 pr-4 py-4 text-slate-900 text-base shadow-lg placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
            <button
              type="submit"
              className="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg bg-[#1B3A4B] px-4 py-2 text-sm font-medium text-white hover:bg-[#243f52] transition-colors"
            >
              Search
            </button>
          </form>
        </div>
      </div>

      <div className="mx-auto max-w-5xl px-4 py-12 space-y-12">
        {/* Categories */}
        <section>
          <h2 className="text-xl font-bold text-slate-900 mb-6">Browse by Category</h2>
          {catsLoading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-28 rounded-xl bg-slate-200 animate-pulse" />
              ))}
            </div>
          ) : categories.length === 0 ? (
            <p className="text-slate-500 text-sm">No categories yet.</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {categories.map((cat) => (
                <Link
                  key={cat.id}
                  to={`/help/${orgSlug}/c/${cat.slug}`}
                  className="group flex flex-col gap-2 rounded-xl border border-slate-200 bg-white p-5 hover:border-indigo-300 hover:shadow-sm transition-all"
                >
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50 group-hover:bg-indigo-100 transition-colors">
                    <FileText className="h-5 w-5 text-indigo-600" />
                  </div>
                  <div>
                    <p className="font-semibold text-slate-900 group-hover:text-indigo-700 transition-colors">{cat.name}</p>
                    <p className="text-sm text-slate-500 mt-0.5">
                      {cat.article_count ?? 0} article{cat.article_count !== 1 ? 's' : ''}
                    </p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-slate-300 group-hover:text-indigo-400 mt-auto self-end" />
                </Link>
              ))}
            </div>
          )}
        </section>

        {/* Featured Articles */}
        <section>
          <h2 className="text-xl font-bold text-slate-900 mb-6">Featured Articles</h2>
          {articlesLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-14 rounded-xl bg-slate-200 animate-pulse" />
              ))}
            </div>
          ) : featuredArticles.length === 0 ? (
            <p className="text-slate-500 text-sm">No articles published yet.</p>
          ) : (
            <div className="rounded-xl border border-slate-200 bg-white divide-y divide-slate-100 overflow-hidden">
              {featuredArticles.slice(0, 8).map((article) => (
                <Link
                  key={article.id}
                  to={`/help/${orgSlug}/a/${article.slug}`}
                  className="flex items-center justify-between px-5 py-4 hover:bg-slate-50 transition-colors group"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <FileText className="h-4 w-4 text-slate-400 shrink-0" />
                    <span className="text-sm font-medium text-slate-800 group-hover:text-indigo-700 transition-colors truncate">
                      {article.title}
                    </span>
                  </div>
                  <ChevronRight className="h-4 w-4 text-slate-300 group-hover:text-indigo-400 shrink-0 ml-3" />
                </Link>
              ))}
            </div>
          )}
        </section>
      </div>

      {/* Footer */}
      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Powered by Omnir CRM
      </div>
    </div>
  )
}
