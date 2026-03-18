import { useState } from 'react'
import { Pencil, Trash2, Pin, PinOff, Users, Loader2, Check, X } from 'lucide-react'
import { SidePanel } from '@/components/ui/SidePanel'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useViews, useUpdateView, useDeleteView } from '@/hooks/useViews'
import type { SavedView, ViewEntityType } from '@/api/types'

interface ViewManagerPanelProps {
  open: boolean
  onClose: () => void
  entityType: ViewEntityType
  activeViewId: string | null
  onSelectView: (view: SavedView) => void
}

interface ViewRowProps {
  view: SavedView
  isActive: boolean
  onSelect: () => void
  onDelete: () => void
  onUpdate: (id: string, updates: { name?: string; is_shared?: boolean; is_pinned?: boolean }) => void
}

function ViewRow({ view, isActive, onSelect, onDelete, onUpdate }: ViewRowProps) {
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState(view.name)

  const handleSaveName = () => {
    if (editName.trim() && editName.trim() !== view.name) {
      onUpdate(view.id, { name: editName.trim() })
    }
    setEditing(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSaveName()
    if (e.key === 'Escape') { setEditName(view.name); setEditing(false) }
  }

  return (
    <div
      className={`flex items-center gap-2 rounded-lg px-3 py-2 transition-colors ${
        isActive ? 'bg-indigo-50 ring-1 ring-indigo-200' : 'hover:bg-slate-50'
      }`}
    >
      <button
        className="flex-1 text-left"
        onClick={onSelect}
      >
        {editing ? (
          <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
            <Input
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              onKeyDown={handleKeyDown}
              className="h-7 py-0 text-sm"
              autoFocus
            />
            <button onClick={handleSaveName} className="text-green-600 hover:text-green-700">
              <Check className="h-3.5 w-3.5" />
            </button>
            <button onClick={() => { setEditName(view.name); setEditing(false) }} className="text-slate-400 hover:text-slate-600">
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <span className="text-sm font-medium text-slate-800">{view.name}</span>
            {view.is_shared && (
              <span title="Shared with team">
                <Users className="h-3.5 w-3.5 text-slate-400" />
              </span>
            )}
          </div>
        )}
      </button>

      <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100 [.hover-row:hover_&]:opacity-100">
        <button
          title={view.is_pinned ? 'Unpin' : 'Pin to bar'}
          onClick={() => onUpdate(view.id, { is_pinned: !view.is_pinned })}
          className="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
        >
          {view.is_pinned ? <PinOff className="h-3.5 w-3.5" /> : <Pin className="h-3.5 w-3.5" />}
        </button>
        <button
          title="Rename"
          onClick={() => setEditing(true)}
          className="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
        >
          <Pencil className="h-3.5 w-3.5" />
        </button>
        <button
          title="Delete"
          onClick={onDelete}
          className="rounded p-1 text-slate-400 hover:bg-red-50 hover:text-red-500"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  )
}

export function ViewManagerPanel({
  open,
  onClose,
  entityType,
  activeViewId,
  onSelectView,
}: ViewManagerPanelProps) {
  const { data: views = [], isLoading } = useViews(entityType)
  const updateView = useUpdateView()
  const deleteView = useDeleteView()

  const [deletingId, setDeletingId] = useState<string | null>(null)

  const handleDelete = async (view: SavedView) => {
    setDeletingId(view.id)
    try {
      await deleteView.mutateAsync({ id: view.id, entityType: view.entity_type })
    } finally {
      setDeletingId(null)
    }
  }

  const handleUpdate = async (
    id: string,
    updates: { name?: string; is_shared?: boolean; is_pinned?: boolean }
  ) => {
    const view = views.find((v) => v.id === id)
    if (!view) return
    await updateView.mutateAsync({ id, payload: updates })
  }

  const myViews = views.filter((v) => !v.is_shared)
  const sharedViews = views.filter((v) => v.is_shared)

  return (
    <SidePanel open={open} onClose={onClose} title="Manage views" width="sm">
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      ) : views.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-sm text-slate-500">No saved views yet.</p>
          <p className="mt-1 text-xs text-slate-400">
            Use "Save current filters" to create your first view.
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          {myViews.length > 0 && (
            <section>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
                My views
              </h3>
              <div className="group space-y-1">
                {myViews.map((view) => (
                  <div key={view.id} className="hover-row">
                    {deletingId === view.id ? (
                      <div className="flex items-center gap-2 px-3 py-2 text-sm text-slate-400">
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        Deleting…
                      </div>
                    ) : (
                      <ViewRow
                        view={view}
                        isActive={activeViewId === view.id}
                        onSelect={() => { onSelectView(view); onClose() }}
                        onDelete={() => handleDelete(view)}
                        onUpdate={handleUpdate}
                      />
                    )}
                  </div>
                ))}
              </div>
            </section>
          )}

          {sharedViews.length > 0 && (
            <section>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
                Team views
              </h3>
              <div className="group space-y-1">
                {sharedViews.map((view) => (
                  <div key={view.id} className="hover-row">
                    {deletingId === view.id ? (
                      <div className="flex items-center gap-2 px-3 py-2 text-sm text-slate-400">
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        Deleting…
                      </div>
                    ) : (
                      <ViewRow
                        view={view}
                        isActive={activeViewId === view.id}
                        onSelect={() => { onSelectView(view); onClose() }}
                        onDelete={() => handleDelete(view)}
                        onUpdate={handleUpdate}
                      />
                    )}
                  </div>
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </SidePanel>
  )
}
