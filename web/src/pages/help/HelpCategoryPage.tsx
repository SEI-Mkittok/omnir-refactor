import { Link, useParams } from 'react-router-dom'
import { BookOpen, ChevronRight, FileText, ArrowLeft } from 'lucide-react'
import { usePublicKbCategories, usePublicKbArticles } from '@/hooks/useKB'

export function HelpCategoryPage() {
  const { orgSlug = '', categorySlug = '' } = useParams<{ orgSlug: string; categorySlug: string }>()

  const { data: categories = [] } = usePublicKbCategories(orgSlug)
  const category = categories.find((c) => c.slug === categorySlug)

  const { data: articles = [], isLoading } = usePublicKbArticles(orgSlug)
  const filtered = articles.filter((a) => a.category_id === category?.id)

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
          <span className="text-white text-sm font-medium">{category?.name ?? categorySlug}</span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl px-4 py-10 space-y-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--color-primary-light)]">
            <BookOpen className="h-5 w-5 text-[var(--color-primary)]" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-slate-900">{category?.name ?? categorySlug}</h1>
            <p className="text-sm text-slate-500">{filtered.length} article{filtered.length !== 1 ? 's' : ''}</p>
          </div>
        </div>

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 rounded-xl bg-slate-200 animate-pulse" />
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <div className="rounded-xl border border-slate-200 bg-white p-10 text-center">
            <FileText className="mx-auto h-8 w-8 text-slate-300 mb-3" />
            <p className="text-sm text-slate-500">No articles in this category yet.</p>
          </div>
        ) : (
          <div className="rounded-xl border border-slate-200 bg-white divide-y divide-slate-100 overflow-hidden">
            {filtered.map((article) => (
              <Link
                key={article.id}
                to={`/help/${orgSlug}/a/${article.slug}`}
                className="flex flex-col sm:flex-row sm:items-center justify-between px-5 py-4 hover:bg-slate-50 transition-colors group gap-2"
              >
                <div className="flex items-start sm:items-center gap-3 min-w-0">
                  <FileText className="h-4 w-4 text-slate-400 shrink-0 mt-0.5 sm:mt-0" />
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-slate-800 group-hover:text-[var(--color-primary)] transition-colors">
                      {article.title}
                    </p>
                    {article.excerpt && (
                      <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{article.excerpt}</p>
                    )}
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-300 group-hover:text-[var(--color-primary)] shrink-0 self-end sm:self-center ml-auto" />
              </Link>
            ))}
          </div>
        )}

        <div className="text-center">
          <Link
            to={`/help/${orgSlug}`}
            className="text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)] transition-colors"
          >
            ← Back to all categories
          </Link>
        </div>
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Powered by Omnir CRM
      </div>
    </div>
  )
}
