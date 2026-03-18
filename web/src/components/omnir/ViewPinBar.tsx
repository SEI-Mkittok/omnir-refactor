import { useState, useEffect } from 'react'
import {
  DndContext,
  DragEndEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  closestCenter,
} from '@dnd-kit/core'
import {
  SortableContext,
  horizontalListSortingStrategy,
  useSortable,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Plus, Settings2, AlertCircle, RefreshCw } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { Button } from '@/components/ui/Button'
import { SaveViewModal } from './SaveViewModal'
import { ViewManagerPanel } from './ViewManagerPanel'
import { useViews, usePinView } from '@/hooks/useViews'
import type { SavedView, ViewEntityType, ViewFilters } from '@/api/types'

// ── Sortable tab item ──────────────────────────────────────────────────────────

interface SortableTabProps {
  view: SavedView
  isActive: boolean
  onClick: () => void
}

function SortableTab({ view, isActive, onClick }: SortableTabProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: view.id,
  })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    cursor: isDragging ? 'grabbing' : 'grab',
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <TabsTrigger
        value={view.id}
        data-state={isActive ? 'active' : 'inactive'}
        onClick={onClick}
        className="select-none"
      >
        {view.name}
        {view.is_shared && (
          <span className="ml-1 text-xs text-slate-400" title="Shared">
            👥
          </span>
        )}
      </TabsTrigger>
    </div>
  )
}

// ── ViewPinBar ─────────────────────────────────────────────────────────────────

interface ViewPinBarProps {
  entityType: ViewEntityType
  activeViewId: string | null
  hasUnsavedChanges: boolean
  currentFilters: ViewFilters
  onSelectView: (view: SavedView) => void
  onClearView: () => void
  onViewSaved: (view: SavedView) => void
  onUpdateView?: (viewId: string) => void
}

export function ViewPinBar({
  entityType,
  activeViewId,
  hasUnsavedChanges,
  currentFilters,
  onSelectView,
  onClearView,
  onViewSaved,
  onUpdateView,
}: ViewPinBarProps) {
  const [showSaveModal, setShowSaveModal] = useState(false)
  const [showManager, setShowManager] = useState(false)

  const { data: allViews = [] } = useViews(entityType)
  const pinView = usePinView()

  // Only show pinned views in the bar, sorted by pin_order
  const pinnedViews = allViews
    .filter((v) => v.is_pinned)
    .sort((a, b) => a.pin_order - b.pin_order)

  const [localOrder, setLocalOrder] = useState<string[]>([])

  // Sync localOrder when pinned views change (useEffect to avoid render-phase setState)
  const serverOrder = pinnedViews.map((v) => v.id).join(',')
  useEffect(() => {
    setLocalOrder(pinnedViews.map((v) => v.id))
  }, [serverOrder]) // eslint-disable-line react-hooks/exhaustive-deps

  const orderedViews =
    localOrder.length > 0
      ? localOrder
          .map((id) => pinnedViews.find((v) => v.id === id))
          .filter(Boolean) as SavedView[]
      : pinnedViews

  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 250, tolerance: 5 } })
  )

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = orderedViews.findIndex((v) => v.id === active.id)
    const newIndex = orderedViews.findIndex((v) => v.id === over.id)
    const reordered = arrayMove(orderedViews, oldIndex, newIndex)
    setLocalOrder(reordered.map((v) => v.id))

    // Persist new order to backend
    reordered.forEach((view, i) => {
      if (view.pin_order !== i) {
        pinView.mutate({ id: view.id, pinOrder: i })
      }
    })
  }

  if (pinnedViews.length === 0 && !hasUnsavedChanges) {
    // Show minimal bar with just save button when no pinned views
    return (
      <div className="flex items-center gap-2 border-b border-slate-200 px-1 pb-0">
        <button
          onClick={() => setShowSaveModal(true)}
          className="flex items-center gap-1 px-3 py-2 text-sm text-slate-500 hover:text-slate-700"
        >
          <Plus className="h-3.5 w-3.5" />
          Save view
        </button>
        {allViews.length > 0 && (
          <button
            onClick={() => setShowManager(true)}
            className="rounded p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
            title="Manage views"
          >
            <Settings2 className="h-4 w-4" />
          </button>
        )}
        <SaveViewModal
          open={showSaveModal}
          onClose={() => setShowSaveModal(false)}
          entityType={entityType}
          currentFilters={currentFilters}
          onSaved={onViewSaved}
        />
        <ViewManagerPanel
          open={showManager}
          onClose={() => setShowManager(false)}
          entityType={entityType}
          activeViewId={activeViewId}
          onSelectView={onSelectView}
        />
      </div>
    )
  }

  return (
    <>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <Tabs
          value={activeViewId ?? ''}
          onValueChange={(id) => {
            const view = orderedViews.find((v) => v.id === id)
            if (view) onSelectView(view)
          }}
        >
          <TabsList className="flex-wrap gap-0 border-b-0 px-0">
            <SortableContext
              items={orderedViews.map((v) => v.id)}
              strategy={horizontalListSortingStrategy}
            >
              {orderedViews.map((view) => (
                <SortableTab
                  key={view.id}
                  view={view}
                  isActive={activeViewId === view.id}
                  onClick={() => {
                    if (activeViewId === view.id) {
                      onClearView()
                    } else {
                      onSelectView(view)
                    }
                  }}
                />
              ))}
            </SortableContext>

            {/* Save view button */}
            <button
              onClick={() => setShowSaveModal(true)}
              className="flex items-center gap-1 border-b-2 border-transparent px-3 py-2 text-sm text-slate-400 hover:text-slate-600"
            >
              <Plus className="h-3.5 w-3.5" />
              Save view
            </button>

            {/* Gear icon */}
            <button
              onClick={() => setShowManager(true)}
              className="ml-auto rounded p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
              title="Manage views"
            >
              <Settings2 className="h-4 w-4" />
            </button>
          </TabsList>
        </Tabs>
      </DndContext>

      {/* Unsaved changes indicator */}
      {hasUnsavedChanges && (
        <div className="flex items-center gap-2 rounded-md bg-amber-50 px-3 py-1.5 text-xs text-amber-700 ring-1 ring-amber-200">
          <AlertCircle className="h-3.5 w-3.5 shrink-0" />
          <span>Unsaved filter changes</span>
          <button
            onClick={() => setShowSaveModal(true)}
            className="ml-1 font-medium underline underline-offset-2 hover:text-amber-800"
          >
            Save as new
          </button>
          {onUpdateView && activeViewId && (
            <button
              onClick={() => onUpdateView(activeViewId)}
              className="flex items-center gap-0.5 font-medium underline underline-offset-2 hover:text-amber-800"
            >
              <RefreshCw className="h-3 w-3" />
              Update view
            </button>
          )}
        </div>
      )}

      <SaveViewModal
        open={showSaveModal}
        onClose={() => setShowSaveModal(false)}
        entityType={entityType}
        currentFilters={currentFilters}
        onSaved={onViewSaved}
      />
      <ViewManagerPanel
        open={showManager}
        onClose={() => setShowManager(false)}
        entityType={entityType}
        activeViewId={activeViewId}
        onSelectView={onSelectView}
      />
    </>
  )
}
