import { Link, useParams } from 'react-router-dom'
import { ChevronRight, ArrowLeft, Eye, Calendar } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { usePublicKbArticle, usePublicKbCategories } from '@/hooks/useKB'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

export function HelpArticlePage() {
  const { orgSlug = '', articleSlug = '' } = useParams<{ orgSlug: string; articleSlug: string }>()

  const { data: article, isLoading } = usePublicKbArticle(orgSlug, articleSlug)
  const { data: categories = [] } = usePublicKbCategories(orgSlug)
  const category = article?.category_id
    ? categories.find((c) => c.id === article.category_id)
    : null

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      {/* Top bar */}
      <div className="bg-[#1B3A4B] text-white px-4 py-4">
        <div className="mx-auto max-w-3xl flex items-center gap-3 flex-wrap">
          <Link
            to={`/help/${orgSlug}`}
            className="flex items-center gap-1.5 text-white/70 hover:text-white text-sm transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Help Center
          </Link>
          {category && (
            <>
              <ChevronRight className="h-4 w-4 text-white/40" />
              <Link
                to={`/help/${orgSlug}/c/${category.slug}`}
                className="text-white/70 hover:text-white text-sm transition-colors"
              >
                {category.name}
              </Link>
            </>
          )}
          <ChevronRight className="h-4 w-4 text-white/40" />
          <span className="text-white text-sm font-medium truncate max-w-[200px]">
            {isLoading ? '…' : article?.title ?? articleSlug}
          </span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl px-4 py-10">
        {isLoading ? (
          <div className="space-y-4 animate-pulse">
            <div className="h-8 bg-slate-200 rounded w-3/4" />
            <div className="h-4 bg-slate-200 rounded w-1/3" />
            <div className="h-4 bg-slate-200 rounded" />
            <div className="h-4 bg-slate-200 rounded" />
            <div className="h-4 bg-slate-200 rounded w-5/6" />
          </div>
        ) : !article ? (
          <div className="text-center py-16">
            <p className="text-slate-500">Article not found.</p>
            <Link to={`/help/${orgSlug}`} className="text-sm text-[var(--color-primary)] hover:underline mt-4 block">
              Back to Help Center
            </Link>
          </div>
        ) : (
          <article className="rounded-xl border border-slate-200 bg-white p-8 sm:p-10">
            {/* Breadcrumb category badge */}
            {category && (
              <Link
                to={`/help/${orgSlug}/c/${category.slug}`}
                className="inline-flex items-center gap-1 rounded-full bg-[var(--color-primary-light)] px-3 py-1 text-xs font-semibold uppercase tracking-wide text-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors mb-4"
              >
                {category.name}
              </Link>
            )}

            <h1 className="text-3xl font-bold text-slate-900 mb-4">{article.title}</h1>

            <div className="flex items-center gap-4 text-xs text-slate-400 mb-8 pb-8 border-b border-slate-100">
              <span className="flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5" />
                Updated {formatDate(article.updated_at)}
              </span>
              <span className="flex items-center gap-1.5">
                <Eye className="h-3.5 w-3.5" />
                {article.view_count.toLocaleString()} view{article.view_count !== 1 ? 's' : ''}
              </span>
            </div>

            {/* Rendered markdown */}
            <div className="prose prose-slate max-w-none prose-headings:font-bold prose-a:text-[var(--color-primary)] prose-a:no-underline hover:prose-a:underline prose-code:bg-slate-100 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-sm prose-pre:bg-slate-900 prose-pre:text-slate-100">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {article.body}
              </ReactMarkdown>
            </div>

            <div className="mt-10 pt-8 border-t border-slate-100 text-center">
              <p className="text-sm text-slate-500">Was this article helpful?</p>
              <div className="flex justify-center gap-3 mt-3">
                <button className="rounded-lg border border-slate-300 px-5 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors">
                  👍 Yes
                </button>
                <button className="rounded-lg border border-slate-300 px-5 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors">
                  👎 No
                </button>
              </div>
            </div>
          </article>
        )}

        {/* Related / back nav */}
        <div className="mt-6 text-center">
          {category ? (
            <Link
              to={`/help/${orgSlug}/c/${category.slug}`}
              className="text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)] transition-colors"
            >
              ← Back to {category.name}
            </Link>
          ) : (
            <Link
              to={`/help/${orgSlug}`}
              className="text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)] transition-colors"
            >
              ← Back to Help Center
            </Link>
          )}
        </div>
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Powered by Omnir CRM
      </div>
    </div>
  )
}
