import { useState, useCallback } from 'react'
import {
  BookOpen,
  Plus,
  Pencil,
  Trash2,
  Eye,
  EyeOff,
  GripVertical,
  X,
  Search,
  ExternalLink,
} from 'lucide-react'
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import MDEditor from '@uiw/react-md-editor'
import { Button } from '@/components/ui/Button'
import {
  useKbCategories,
  useKbArticles,
  useCreateKbCategory,
  useUpdateKbCategory,
  useDeleteKbCategory,
  useCreateKbArticle,
  useUpdateKbArticle,
  useDeleteKbArticle,
  useKbArticle,
} from '@/hooks/useKB'
import type { KbCategory, KbArticle, KbArticleSummary, KbArticleStatus } from '@/api/types'
import { useAuthStore } from '@/stores/auth'

// ── Sortable Category Row ─────────────────────────────────────────────────────

function SortableCategoryRow({
  category,
  isSelected,
  onSelect,
  onEdit,
  onDelete,
}: {
  category: KbCategory
  isSelected: boolean
  onSelect: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: category.id,
  })

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={`flex items-center gap-2 rounded-lg px-3 py-2 cursor-pointer select-none transition-colors ${
        isDragging ? 'opacity-50 z-50' : ''
      } ${isSelected ? 'bg-indigo-50 border border-indigo-200' : 'hover:bg-slate-50 border border-transparent'}`}
      onClick={onSelect}
    >
      <button
        {...attributes}
        {...listeners}
        className="p-1 text-slate-400 hover:text-slate-600 cursor-grab active:cursor-grabbing"
        onClick={(e) => e.stopPropagation()}
      >
        <GripVertical className="h-4 w-4" />
      </button>
      <span className="flex-1 text-sm font-medium text-slate-800 truncate">{category.name}</span>
      <button
        className="p-1 text-slate-400 hover:text-indigo-600"
        onClick={(e) => { e.stopPropagation(); onEdit() }}
      >
        <Pencil className="h-3.5 w-3.5" />
      </button>
      <button
        className="p-1 text-slate-400 hover:text-red-600"
        onClick={(e) => { e.stopPropagation(); onDelete() }}
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  )
}

// ── Category Form Dialog ──────────────────────────────────────────────────────

function CategoryFormDialog({
  initial,
  onSave,
  onClose,
}: {
  initial?: KbCategory
  onSave: (name: string) => void
  onClose: () => void
}) {
  const [name, setName] = useState(initial?.name ?? '')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative z-10 w-full max-w-sm rounded-xl bg-white shadow-xl p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold text-slate-900">{initial ? 'Edit Category' : 'New Category'}</h3>
          <Button variant="ghost" size="icon" onClick={onClose}><X className="h-4 w-4" /></Button>
        </div>
        <div className="space-y-1">
          <label className="block text-sm font-medium text-slate-700">Name</label>
          <input
            autoFocus
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && name.trim() && onSave(name.trim())}
            placeholder="e.g. Getting Started"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={() => name.trim() && onSave(name.trim())} disabled={!name.trim()}>
            {initial ? 'Save' : 'Create'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// ── Article Editor ────────────────────────────────────────────────────────────

function ArticleEditor({
  article,
  categories,
  onClose,
}: {
  article?: KbArticle | null
  categories: KbCategory[]
  onClose: () => void
}) {
  const [title, setTitle] = useState(article?.title ?? '')
  const [body, setBody] = useState(article?.body ?? '')
  const [categoryId, setCategoryId] = useState<string>(article?.category_id ?? '')
  const [status, setStatus] = useState<KbArticleStatus>(article?.status ?? 'draft')
  const [error, setError] = useState('')

  const { mutateAsync: createArticle, isPending: creating } = useCreateKbArticle()
  const { mutateAsync: updateArticle, isPending: updating } = useUpdateKbArticle()
  const isPending = creating || updating

  const handleSave = async () => {
    if (!title.trim()) { setError('Title is required.'); return }
    setError('')
    const payload = {
      title: title.trim(),
      body,
      status,
      category_id: categoryId || null,
    }
    if (article) {
      await updateArticle({ id: article.id, payload })
    } else {
      await createArticle(payload)
    }
    onClose()
  }

  return (
    <div className="flex flex-col h-full">
      {/* Editor Header */}
      <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4 bg-white">
        <h2 className="text-lg font-semibold text-slate-900">{article ? 'Edit Article' : 'New Article'}</h2>
        <Button variant="ghost" size="icon" onClick={onClose}><X className="h-5 w-5" /></Button>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-4 bg-[#F7F8FA]">
        {/* Title */}
        <div className="space-y-1">
          <label className="block text-xs font-semibold uppercase tracking-wide text-slate-500">Title</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Article title…"
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-base font-medium text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          {error && <p className="text-xs text-red-600">{error}</p>}
        </div>

        {/* Meta row */}
        <div className="flex flex-wrap gap-4">
          <div className="space-y-1 flex-1 min-w-[160px]">
            <label className="block text-xs font-semibold uppercase tracking-wide text-slate-500">Category</label>
            <select
              value={categoryId}
              onChange={(e) => setCategoryId(e.target.value)}
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            >
              <option value="">Uncategorized</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
          <div className="space-y-1 min-w-[140px]">
            <label className="block text-xs font-semibold uppercase tracking-wide text-slate-500">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value as KbArticleStatus)}
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            >
              <option value="draft">Draft</option>
              <option value="published">Published</option>
            </select>
          </div>
        </div>

        {/* Markdown Editor */}
        <div className="space-y-1">
          <label className="block text-xs font-semibold uppercase tracking-wide text-slate-500">Body</label>
          <div data-color-mode="light">
            <MDEditor
              value={body}
              onChange={(val) => setBody(val ?? '')}
              height={420}
              preview="live"
            />
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-end gap-2 border-t border-slate-200 px-6 py-4 bg-white">
        <Button variant="outline" onClick={onClose} disabled={isPending}>Cancel</Button>
        <Button onClick={handleSave} disabled={isPending}>
          {isPending ? 'Saving…' : 'Save Article'}
        </Button>
      </div>
    </div>
  )
}

// ── Article List Table ────────────────────────────────────────────────────────

function ArticleListTable({
  articles,
  categories,
  isLoading,
  onEdit,
  onToggleStatus,
  onDelete,
}: {
  articles: KbArticleSummary[]
  categories: KbCategory[]
  isLoading: boolean
  onEdit: (a: KbArticleSummary) => void
  onToggleStatus: (a: KbArticleSummary) => void
  onDelete: (a: KbArticleSummary) => void
}) {
  const categoryMap = Object.fromEntries(categories.map((c) => [c.id, c.name]))

  if (isLoading) {
    return <div className="p-8 text-center text-sm text-slate-500">Loading articles…</div>
  }

  if (articles.length === 0) {
    return (
      <div className="p-12 text-center">
        <BookOpen className="mx-auto h-10 w-10 text-slate-300 mb-3" />
        <p className="text-sm text-slate-500">No articles yet. Create one to get started.</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200 bg-slate-50">
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-wide text-slate-500">Title</th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-wide text-slate-500">Category</th>
            <th className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-wide text-slate-500">Status</th>
            <th className="px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-wide text-slate-500">Views</th>
            <th className="px-4 py-3" />
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {articles.map((a) => (
            <tr key={a.id} className="hover:bg-slate-50 transition-colors">
              <td className="px-4 py-3">
                <button
                  className="text-left font-medium text-slate-900 hover:text-indigo-600 transition-colors line-clamp-1"
                  onClick={() => onEdit(a)}
                >
                  {a.title}
                </button>
              </td>
              <td className="px-4 py-3 text-slate-500">
                {a.category_id ? (categoryMap[a.category_id] ?? '—') : <span className="italic text-slate-400">Uncategorized</span>}
              </td>
              <td className="px-4 py-3">
                <span
                  className={`inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide ${
                    a.status === 'published'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-amber-100 text-amber-700'
                  }`}
                >
                  {a.status}
                </span>
              </td>
              <td className="px-4 py-3 text-right text-slate-500">{a.view_count.toLocaleString()}</td>
              <td className="px-4 py-3">
                <div className="flex items-center justify-end gap-1">
                  <button
                    title={a.status === 'published' ? 'Unpublish' : 'Publish'}
                    className="p-1.5 rounded text-slate-400 hover:text-indigo-600 hover:bg-indigo-50"
                    onClick={() => onToggleStatus(a)}
                  >
                    {a.status === 'published' ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                  <button
                    title="Edit"
                    className="p-1.5 rounded text-slate-400 hover:text-indigo-600 hover:bg-indigo-50"
                    onClick={() => onEdit(a)}
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    title="Delete"
                    className="p-1.5 rounded text-slate-400 hover:text-red-600 hover:bg-red-50"
                    onClick={() => onDelete(a)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function KnowledgeBasePage() {
  const activeOrg = useAuthStore((s) => s.activeOrg)
  const user = useAuthStore((s) => s.user)

  const { data: categories = [], isLoading: catsLoading } = useKbCategories()
  const [selectedCategoryId, setSelectedCategoryId] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<'' | 'draft' | 'published'>('')
  const [search, setSearch] = useState('')

  const articleParams = {
    ...(selectedCategoryId ? { category_id: selectedCategoryId } : {}),
    ...(statusFilter ? { status: statusFilter as 'draft' | 'published' } : {}),
    ...(search ? { q: search } : {}),
  }
  const { data: articlesData, isLoading: articlesLoading } = useKbArticles(
    Object.keys(articleParams).length ? articleParams : undefined
  )
  const articles = articlesData?.data ?? []

  const { mutateAsync: createCategory } = useCreateKbCategory()
  const { mutateAsync: updateCategory } = useUpdateKbCategory()
  const { mutateAsync: deleteCategory } = useDeleteKbCategory()
  const { mutateAsync: updateArticle } = useUpdateKbArticle()
  const { mutateAsync: deleteArticle } = useDeleteKbArticle()

  // UI state — declared before useKbArticle so they're available in the hook call
  const [showCategoryForm, setShowCategoryForm] = useState(false)
  const [editingCategory, setEditingCategory] = useState<KbCategory | null>(null)
  const [editingArticle, setEditingArticle] = useState<KbArticleSummary | 'new' | null>(null)

  const { data: editingArticleData } = useKbArticle(
    typeof editingArticle === 'object' && editingArticle !== null && 'id' in editingArticle
      ? (editingArticle as KbArticleSummary).id
      : ''
  )

  // DnD sensors
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }))

  const handleCategoryDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
      if (!over || active.id === over.id) return
      const oldIndex = categories.findIndex((c) => c.id === active.id)
      const newIndex = categories.findIndex((c) => c.id === over.id)
      if (oldIndex === -1 || newIndex === -1) return
      const reordered = arrayMove(categories, oldIndex, newIndex)
      reordered.forEach((cat, i) => {
        if (cat.position !== i) {
          updateCategory({ id: cat.id, payload: { position: i } })
        }
      })
    },
    [categories, updateCategory]
  )

  const handleSaveCategory = async (name: string) => {
    if (editingCategory) {
      await updateCategory({ id: editingCategory.id, payload: { name } })
      setEditingCategory(null)
    } else {
      await createCategory({ name })
      setShowCategoryForm(false)
    }
  }

  const handleToggleStatus = async (a: KbArticleSummary) => {
    await updateArticle({
      id: a.id,
      payload: { status: a.status === 'published' ? 'draft' : 'published' },
    })
  }

  const handleDeleteArticle = async (a: KbArticleSummary) => {
    if (confirm(`Delete "${a.title}"? This cannot be undone.`)) {
      await deleteArticle(a.id)
    }
  }

  const handleDeleteCategory = async (c: KbCategory) => {
    if (confirm(`Delete category "${c.name}"? Articles will become uncategorized.`)) {
      await deleteCategory(c.id)
      if (selectedCategoryId === c.id) setSelectedCategoryId(null)
    }
  }

  // Full article data for editing
  const fullArticleForEditor =
    editingArticle === 'new'
      ? null
      : editingArticle && editingArticleData
      ? editingArticleData
      : null

  const orgSlug = activeOrg?.slug ?? user?.email?.split('@')[1] ?? ''

  // If editing, show full-screen editor
  if (editingArticle !== null) {
    return (
      <div className="fixed inset-0 z-40 bg-[#F7F8FA] flex flex-col">
        <ArticleEditor
          article={fullArticleForEditor}
          categories={categories}
          onClose={() => setEditingArticle(null)}
        />
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Knowledge Base.</h1>
          <p className="text-sm text-slate-500 mt-1">Manage help articles and categories.</p>
        </div>
        <div className="flex items-center gap-2">
          {orgSlug && (
            <a
              href={`/help/${orgSlug}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700 hover:bg-slate-50 transition-colors"
            >
              <ExternalLink className="h-4 w-4" />
              View Help Center
            </a>
          )}
          <Button onClick={() => setEditingArticle('new')}>
            <Plus className="h-4 w-4 mr-1.5" />
            New Article
          </Button>
        </div>
      </div>

      <div className="flex gap-6">
        {/* Categories sidebar */}
        <div className="w-56 shrink-0 space-y-2">
          <div className="flex items-center justify-between mb-1">
            <span className="text-[11px] font-semibold uppercase tracking-wide text-slate-500">Categories</span>
            <button
              className="rounded p-0.5 text-slate-400 hover:text-indigo-600"
              onClick={() => setShowCategoryForm(true)}
              title="Add category"
            >
              <Plus className="h-4 w-4" />
            </button>
          </div>

          {/* All articles option */}
          <div
            className={`flex items-center gap-2 rounded-lg px-3 py-2 cursor-pointer text-sm font-medium transition-colors ${
              selectedCategoryId === null
                ? 'bg-indigo-50 border border-indigo-200 text-indigo-800'
                : 'text-slate-700 hover:bg-slate-50 border border-transparent'
            }`}
            onClick={() => setSelectedCategoryId(null)}
          >
            <BookOpen className="h-4 w-4 shrink-0" />
            All Articles
          </div>

          {catsLoading ? (
            <div className="text-xs text-slate-400 px-3 py-2">Loading…</div>
          ) : (
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleCategoryDragEnd}>
              <SortableContext items={categories.map((c) => c.id)} strategy={verticalListSortingStrategy}>
                {categories.map((cat) => (
                  <SortableCategoryRow
                    key={cat.id}
                    category={cat}
                    isSelected={selectedCategoryId === cat.id}
                    onSelect={() => setSelectedCategoryId(cat.id)}
                    onEdit={() => setEditingCategory(cat)}
                    onDelete={() => handleDeleteCategory(cat)}
                  />
                ))}
              </SortableContext>
            </DndContext>
          )}
        </div>

        {/* Article list */}
        <div className="flex-1 rounded-xl border border-slate-200 bg-white overflow-hidden">
          {/* Filters */}
          <div className="flex flex-wrap items-center gap-3 border-b border-slate-200 px-4 py-3">
            <div className="relative flex-1 min-w-[180px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search articles…"
                className="w-full rounded-lg border border-slate-300 pl-9 pr-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as '' | 'draft' | 'published')}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
            >
              <option value="">All statuses</option>
              <option value="published">Published</option>
              <option value="draft">Draft</option>
            </select>
          </div>

          <ArticleListTable
            articles={articles}
            categories={categories}
            isLoading={articlesLoading}
            onEdit={(a) => setEditingArticle(a)}
            onToggleStatus={handleToggleStatus}
            onDelete={handleDeleteArticle}
          />
        </div>
      </div>

      {/* Category dialogs */}
      {showCategoryForm && (
        <CategoryFormDialog
          onSave={handleSaveCategory}
          onClose={() => setShowCategoryForm(false)}
        />
      )}
      {editingCategory && (
        <CategoryFormDialog
          initial={editingCategory}
          onSave={handleSaveCategory}
          onClose={() => setEditingCategory(null)}
        />
      )}
    </div>
  )
}
