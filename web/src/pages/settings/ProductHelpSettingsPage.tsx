import { ExternalLink, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import {
  useProductHelpArticles,
  useProductHelpCategories,
  useProductHelpSyncNow,
  useProductHelpSyncStatus,
} from '@/hooks/useProductHelp'

function formatDate(iso?: string) {
  if (!iso) return 'Never'
  return new Date(iso).toLocaleString()
}

export function ProductHelpSettingsPage() {
  const { data: syncStatus, isLoading: statusLoading } = useProductHelpSyncStatus()
  const { data: categories = [] } = useProductHelpCategories()
  const { data: articles = [] } = useProductHelpArticles()
  const syncNow = useProductHelpSyncNow()
  const latest = syncStatus?.latest ?? null

  const statusTone = latest?.status === 'failed'
    ? 'bg-red-50 text-red-700 border-red-200'
    : latest?.status === 'succeeded'
      ? 'bg-green-50 text-green-700 border-green-200'
      : 'bg-slate-50 text-slate-600 border-slate-200'

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Product Help</h1>
          <p className="mt-1 text-sm text-[#6B7280]">
            Sync global Omnir product guides from the public GitHub Wiki. This is separate from tenant Knowledge Base content.
          </p>
        </div>
        <Button onClick={() => syncNow.mutate()} disabled={syncNow.isPending}>
          <RefreshCw className={syncNow.isPending ? 'mr-2 h-4 w-4 animate-spin' : 'mr-2 h-4 w-4'} />
          {syncNow.isPending ? 'Syncing...' : 'Sync now'}
        </Button>
      </div>

      <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-sm font-semibold text-slate-900">Wiki sync status</h2>
            <p className="mt-1 text-sm text-slate-500">
              Last run: {statusLoading ? 'Loading...' : formatDate(latest?.finished_at ?? latest?.started_at)}
            </p>
          </div>
          <span className={`inline-flex w-fit rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide ${statusTone}`}>
            {latest?.status ?? 'not synced'}
          </span>
        </div>
        {latest?.message && <p className="mt-4 text-sm text-slate-600">{latest.message}</p>}
        {syncNow.data?.error && <p className="mt-4 text-sm text-red-600">{syncNow.data.error}</p>}
        {syncNow.isError && <p className="mt-4 text-sm text-red-600">Unable to sync product help. Check API logs for details.</p>}
        <div className="mt-5 grid gap-3 sm:grid-cols-3">
          <div className="rounded-lg bg-slate-50 p-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Categories</p>
            <p className="mt-1 text-xl font-bold text-slate-900">{latest?.categories_count ?? categories.length}</p>
          </div>
          <div className="rounded-lg bg-slate-50 p-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Articles</p>
            <p className="mt-1 text-xl font-bold text-slate-900">{latest?.articles_count ?? articles.length}</p>
          </div>
          <div className="rounded-lg bg-slate-50 p-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Source</p>
            <a
              href="https://github.com/SEI-Mkittok/omnir-refactor/wiki"
              target="_blank"
              rel="noreferrer"
              className="mt-1 inline-flex items-center gap-1 text-sm font-semibold text-[var(--color-primary)] hover:underline"
            >
              GitHub Wiki
              <ExternalLink className="h-3.5 w-3.5" />
            </a>
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">Published product-help articles</h2>
          <p className="mt-1 text-sm text-slate-500">Edit article copy in GitHub Wiki, then run sync.</p>
        </div>
        {articles.length === 0 ? (
          <p className="px-5 py-8 text-sm text-slate-500">No product-help articles have been synced yet.</p>
        ) : (
          <div className="divide-y divide-slate-100">
            {articles.map((article) => (
              <div key={article.id} className="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-slate-900">{article.title}</p>
                  <p className="text-xs text-slate-500">{article.category_name || 'Uncategorized'} · {article.source_path}</p>
                </div>
                <div className="flex shrink-0 items-center gap-3">
                  <a href={`/product-help/a/${article.slug}`} className="text-sm text-[var(--color-primary)] hover:underline">
                    View
                  </a>
                  {article.edit_url && (
                    <a
                      href={article.edit_url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-[var(--color-primary)] hover:underline"
                    >
                      Edit
                      <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
