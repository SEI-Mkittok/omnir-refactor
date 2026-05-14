import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, BookOpen, ChevronRight, FileText } from 'lucide-react'
import { useProductHelpArticles, useProductHelpCategories } from '@/hooks/useProductHelp'

export function ProductHelpCategoryPage() {
  const { categorySlug = '' } = useParams<{ categorySlug: string }>()
  const { data: categories = [] } = useProductHelpCategories()
  const category = categories.find((c) => c.slug === categorySlug)
  const { data: articles = [], isLoading } = useProductHelpArticles({ category_slug: categorySlug })

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      <div className="bg-[#1B3A4B] px-4 py-4 text-white">
        <div className="mx-auto flex max-w-3xl items-center gap-3">
          <Link to="/product-help" className="flex items-center gap-1.5 text-sm text-white/70 transition-colors hover:text-white">
            <ArrowLeft className="h-4 w-4" />
            Omnir Help
          </Link>
          <ChevronRight className="h-4 w-4 text-white/40" />
          <span className="text-sm font-medium text-white">{category?.name ?? categorySlug}</span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl space-y-6 px-4 py-10">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--color-primary-light)]">
            <BookOpen className="h-5 w-5 text-[var(--color-primary)]" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-slate-900">{category?.name ?? categorySlug}</h1>
            <p className="text-sm text-slate-500">{articles.length} article{articles.length !== 1 ? 's' : ''}</p>
          </div>
        </div>

        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 animate-pulse rounded-xl bg-slate-200" />
            ))}
          </div>
        ) : articles.length === 0 ? (
          <div className="rounded-xl border border-slate-200 bg-white p-10 text-center">
            <FileText className="mx-auto mb-3 h-8 w-8 text-slate-300" />
            <p className="text-sm text-slate-500">No product-help articles are published in this category yet.</p>
          </div>
        ) : (
          <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
            {articles.map((article) => (
              <Link
                key={article.id}
                to={`/product-help/a/${article.slug}`}
                className="group flex flex-col justify-between gap-2 border-b border-slate-100 px-5 py-4 transition-colors last:border-b-0 hover:bg-slate-50 sm:flex-row sm:items-center"
              >
                <div className="flex min-w-0 items-start gap-3 sm:items-center">
                  <FileText className="mt-0.5 h-4 w-4 shrink-0 text-slate-400 sm:mt-0" />
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-slate-800 transition-colors group-hover:text-[var(--color-primary)]">
                      {article.title}
                    </p>
                    {article.excerpt && <p className="mt-0.5 line-clamp-2 text-xs text-slate-500">{article.excerpt}</p>}
                  </div>
                </div>
                <ChevronRight className="ml-auto h-4 w-4 shrink-0 self-end text-slate-300 group-hover:text-[var(--color-primary)] sm:self-center" />
              </Link>
            ))}
          </div>
        )}

        <div className="text-center">
          <Link to="/product-help" className="text-sm text-[var(--color-primary)] transition-colors hover:text-[var(--color-primary)]">
            Back to all categories
          </Link>
        </div>
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Omnir Product Help
      </div>
    </div>
  )
}
