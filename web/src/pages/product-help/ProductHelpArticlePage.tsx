import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, Calendar, ChevronRight, ExternalLink, Eye } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useProductHelpArticle, useProductHelpCategories } from '@/hooks/useProductHelp'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

export function ProductHelpArticlePage() {
  const { articleSlug = '' } = useParams<{ articleSlug: string }>()
  const { data: article, isLoading } = useProductHelpArticle(articleSlug)
  const { data: categories = [] } = useProductHelpCategories()
  const category = article?.category_slug
    ? categories.find((c) => c.slug === article.category_slug)
    : null

  return (
    <div className="min-h-screen bg-[#F7F8FA]">
      <div className="bg-[#1B3A4B] px-4 py-4 text-white">
        <div className="mx-auto flex max-w-3xl flex-wrap items-center gap-3">
          <Link to="/product-help" className="flex items-center gap-1.5 text-sm text-white/70 transition-colors hover:text-white">
            <ArrowLeft className="h-4 w-4" />
            Omnir Help
          </Link>
          {category && (
            <>
              <ChevronRight className="h-4 w-4 text-white/40" />
              <Link to={`/product-help/c/${category.slug}`} className="text-sm text-white/70 transition-colors hover:text-white">
                {category.name}
              </Link>
            </>
          )}
          <ChevronRight className="h-4 w-4 text-white/40" />
          <span className="max-w-[220px] truncate text-sm font-medium text-white">
            {isLoading ? 'Loading' : article?.title ?? articleSlug}
          </span>
        </div>
      </div>

      <div className="mx-auto max-w-3xl px-4 py-10">
        {isLoading ? (
          <div className="space-y-4 animate-pulse">
            <div className="h-8 w-3/4 rounded bg-slate-200" />
            <div className="h-4 w-1/3 rounded bg-slate-200" />
            <div className="h-4 rounded bg-slate-200" />
            <div className="h-4 rounded bg-slate-200" />
            <div className="h-4 w-5/6 rounded bg-slate-200" />
          </div>
        ) : !article ? (
          <div className="py-16 text-center">
            <p className="text-slate-500">Article not found.</p>
            <Link to="/product-help" className="mt-4 block text-sm text-[var(--color-primary)] hover:underline">
              Back to Omnir Help
            </Link>
          </div>
        ) : (
          <article className="rounded-xl border border-slate-200 bg-white p-8 sm:p-10">
            {category && (
              <Link
                to={`/product-help/c/${category.slug}`}
                className="mb-4 inline-flex items-center gap-1 rounded-full bg-[var(--color-primary-light)] px-3 py-1 text-xs font-semibold uppercase tracking-wide text-[var(--color-primary)] transition-colors hover:bg-[var(--color-primary-light)]"
              >
                {category.name}
              </Link>
            )}

            <div className="flex flex-col gap-4 border-b border-slate-100 pb-8 sm:flex-row sm:items-start sm:justify-between">
              <div className="min-w-0">
                <h1 className="text-3xl font-bold text-slate-900">{article.title}</h1>
                <div className="mt-4 flex flex-wrap items-center gap-4 text-xs text-slate-400">
                  <span className="flex items-center gap-1.5">
                    <Calendar className="h-3.5 w-3.5" />
                    Updated {formatDate(article.updated_at)}
                  </span>
                  <span className="flex items-center gap-1.5">
                    <Eye className="h-3.5 w-3.5" />
                    {article.view_count.toLocaleString()} view{article.view_count !== 1 ? 's' : ''}
                  </span>
                </div>
              </div>
              {article.wiki_url && (
                <a
                  href={article.wiki_url}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex shrink-0 items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50"
                >
                  <ExternalLink className="h-4 w-4" />
                  GitHub Wiki
                </a>
              )}
            </div>

            <div className="prose prose-slate mt-8 max-w-none prose-headings:font-bold prose-a:text-[var(--color-primary)] prose-a:no-underline hover:prose-a:underline prose-code:rounded prose-code:bg-slate-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-sm prose-pre:bg-slate-900 prose-pre:text-slate-100">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {article.body}
              </ReactMarkdown>
            </div>
          </article>
        )}

        <div className="mt-6 text-center">
          {category ? (
            <Link to={`/product-help/c/${category.slug}`} className="text-sm text-[var(--color-primary)] transition-colors hover:text-[var(--color-primary)]">
              Back to {category.name}
            </Link>
          ) : (
            <Link to="/product-help" className="text-sm text-[var(--color-primary)] transition-colors hover:text-[var(--color-primary)]">
              Back to Omnir Help
            </Link>
          )}
        </div>
      </div>

      <div className="border-t border-slate-200 py-8 text-center text-sm text-slate-400">
        Omnir Product Help
      </div>
    </div>
  )
}
