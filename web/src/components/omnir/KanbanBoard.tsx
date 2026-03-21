import { useState, useCallback, useRef, useEffect } from 'react'
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  DragOverlay,
  DragStartEvent,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  closestCorners,
  type UniqueIdentifier,
} from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy, sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { dealsApi } from '@/api/deals'
import { dealKeys } from '@/hooks/useDeals'
import { DealCard, SortableDealCard } from './DealCard'
import { formatCurrency } from '@/lib/utils'
import type { Deal, DealStage } from '@/api/types'

const STAGES: { key: DealStage; label: string; dot: string }[] = [
  { key: 'lead', label: 'Lead', dot: 'bg-[#7C8DB0]' },
  { key: 'qualified', label: 'Qualified', dot: 'bg-[#3B82F6]' },
  { key: 'proposal', label: 'Proposal', dot: 'bg-[#1B3A4B]' },
  { key: 'negotiation', label: 'Negotiation', dot: 'bg-[#F59E0B]' },
  { key: 'closed_won', label: 'Closed Won', dot: 'bg-[#22C55E]' },
  { key: 'closed_lost', label: 'Closed Lost', dot: 'bg-[#EF4444]' },
]

const VIRTUAL_THRESHOLD = 15

// Virtualised column for stages with many cards
function VirtualColumn({ deals, onCardClick }: { deals: Deal[]; onCardClick: (id: string) => void }) {
  const parentRef = useRef<HTMLDivElement>(null)
  const virtualizer = useVirtualizer({
    count: deals.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 130,
    overscan: 3,
  })

  return (
    <div ref={parentRef} className="overflow-y-auto flex-1" style={{ maxHeight: 520 }}>
      <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
        {virtualizer.getVirtualItems().map((vItem) => {
          const deal = deals[vItem.index]
          return (
            <div
              key={deal.id}
              style={{ position: 'absolute', top: vItem.start, left: 0, right: 0, paddingBottom: 8 }}
            >
              <SortableDealCard deal={deal} onClick={() => onCardClick(deal.id)} />
            </div>
          )
        })}
      </div>
    </div>
  )
}

interface KanbanBoardProps {
  deals: Deal[]
  onCardClick?: (id: string) => void
}

export function KanbanBoard({ deals, onCardClick }: KanbanBoardProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeDeal, setActiveDeal] = useState<Deal | null>(null)
  const [localDeals, setLocalDeals] = useState<Deal[]>(deals)

  // Keyboard DnD state
  const [keyboardActiveId, setKeyboardActiveId] = useState<UniqueIdentifier | null>(null)
  const announceRef = useRef<HTMLDivElement>(null)

  // Sync when parent data changes
  const dealIds = deals.map((d) => d.id).join(',')
  const localIds = localDeals.map((d) => d.id).join(',')
  if (dealIds !== localIds) {
    setLocalDeals(deals)
  }

  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 250, tolerance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  function getDealsByStage(stage: DealStage) {
    return localDeals.filter((d) => d.stage === stage)
  }

  function getTotalValue(stage: DealStage): number {
    return getDealsByStage(stage).reduce((sum, d) => sum + (d.value ?? 0), 0)
  }

  function handleDragStart(event: DragStartEvent) {
    const deal = localDeals.find((d) => d.id === event.active.id)
    setActiveDeal(deal ?? null)
    setKeyboardActiveId(event.active.id)
    announce(`Picked up ${deal?.title ?? 'deal'}`)
  }

  function handleDragOver(event: DragOverEvent) {
    const { active, over } = event
    if (!over) return

    const activeId = active.id as string
    const overId = over.id as string

    const targetStage = STAGES.find((s) => s.key === overId)?.key
    if (targetStage) {
      setLocalDeals((prev) =>
        prev.map((d) => (d.id === activeId ? { ...d, stage: targetStage } : d))
      )
      announce(`Moved to ${STAGES.find((s) => s.key === targetStage)?.label}`)
      return
    }

    const overDeal = localDeals.find((d) => d.id === overId)
    if (!overDeal) return
    const activeStage = localDeals.find((d) => d.id === activeId)?.stage
    if (overDeal.stage === activeStage) return

    setLocalDeals((prev) =>
      prev.map((d) => (d.id === activeId ? { ...d, stage: overDeal.stage } : d))
    )
    announce(`Moved to ${STAGES.find((s) => s.key === overDeal.stage)?.label}`)
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    setActiveDeal(null)
    setKeyboardActiveId(null)

    if (!over) {
      announce('Dropped — no change')
      return
    }

    const activeId = active.id as string
    const deal = localDeals.find((d) => d.id === activeId)
    const originalDeal = deals.find((d) => d.id === activeId)

    if (!deal || !originalDeal) return
    if (deal.stage === originalDeal.stage) {
      announce('Dropped — no stage change')
      return
    }

    announce(`Dropped in ${STAGES.find((s) => s.key === deal.stage)?.label}`)
    try {
      await dealsApi.update(activeId, { stage: deal.stage })
      queryClient.invalidateQueries({ queryKey: dealKeys.lists() })
    } catch {
      setLocalDeals((prev) =>
        prev.map((d) => (d.id === activeId ? { ...d, stage: originalDeal.stage } : d))
      )
    }
  }

  const announce = useCallback((msg: string) => {
    if (announceRef.current) {
      announceRef.current.textContent = msg
    }
  }, [])

  return (
    <>
      {/* Screen-reader live region for DnD announcements */}
      <div
        ref={announceRef}
        role="status"
        aria-live="assertive"
        aria-atomic="true"
        className="sr-only"
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
        accessibility={{
          announcements: {
            onDragStart: ({ active }) => `Picked up deal ${active.id}. Use arrow keys to move between stages, Space to drop.`,
            onDragOver: ({ active, over }) => over ? `Deal ${active.id} is over ${over.id}` : `Deal ${active.id} is no longer over a drop target`,
            onDragEnd: ({ active, over }) => over ? `Deal ${active.id} dropped into ${over.id}` : `Deal ${active.id} dropped`,
            onDragCancel: ({ active }) => `Drag cancelled. Deal ${active.id} returned to original position`,
          },
        }}
      >
        <div
          className="flex gap-5 overflow-x-auto pb-4"
          role="region"
          aria-label="Pipeline board — use Space on a card to pick it up, arrow keys to move stages, Space again to drop"
        >
          {STAGES.map((stage) => {
            const stageDeals = getDealsByStage(stage.key)
            const totalValue = getTotalValue(stage.key)
            const useVirtual = stageDeals.length > VIRTUAL_THRESHOLD

            return (
              <div
                key={stage.key}
                id={stage.key}
                className="flex shrink-0 flex-col rounded-xl border border-[#E5E7EB] bg-[#F7F8FA]"
                style={{ minWidth: 240, width: 260 }}
                role="group"
                aria-label={`${stage.label} — ${stageDeals.length} deals`}
              >
                {/* Column header */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-[#E5E7EB]">
                  <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full shrink-0 ${stage.dot}`} aria-hidden="true" />
                    <span className="text-xs font-bold uppercase tracking-wider text-[#1A1D23]">
                      {stage.label}
                    </span>
                    <span
                      className="rounded-full bg-[#E5E7EB] px-2 py-0.5 text-[10px] font-bold text-[#6B7280]"
                      aria-label={`${stageDeals.length} deals`}
                    >
                      {stageDeals.length}
                    </span>
                  </div>
                  {totalValue > 0 && (
                    <span className="text-[11px] font-semibold text-[#6B7280]">
                      {formatCurrency(totalValue)}
                    </span>
                  )}
                </div>

                {/* Cards */}
                <SortableContext
                  id={stage.key}
                  items={stageDeals.map((d) => d.id)}
                  strategy={verticalListSortingStrategy}
                >
                  {useVirtual ? (
                    <VirtualColumn
                      deals={stageDeals}
                      onCardClick={(id) => onCardClick ? onCardClick(id) : navigate(`/deals/${id}`)}
                    />
                  ) : (
                    <div className="flex flex-col gap-2 p-3 min-h-[80px]">
                      {stageDeals.map((deal) => (
                        <SortableDealCard
                          key={deal.id}
                          deal={deal}
                          onClick={() => onCardClick ? onCardClick(deal.id) : navigate(`/deals/${deal.id}`)}
                          isKeyboardActive={keyboardActiveId === deal.id}
                        />
                      ))}
                    </div>
                  )}
                </SortableContext>
              </div>
            )
          })}
        </div>

        <DragOverlay>
          {activeDeal && <DealCard deal={activeDeal} />}
        </DragOverlay>
      </DndContext>
    </>
  )
}
