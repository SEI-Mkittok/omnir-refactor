import { useState } from 'react'
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  DragOverlay,
  DragStartEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  closestCorners,
} from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { dealsApi } from '@/api/deals'
import { dealKeys } from '@/hooks/useDeals'
import { DealCard, SortableDealCard } from './DealCard'
import type { Deal, DealStage } from '@/api/types'

const STAGES: { key: DealStage; label: string; color: string }[] = [
  { key: 'lead', label: 'Lead', color: 'bg-slate-100' },
  { key: 'qualified', label: 'Qualified', color: 'bg-blue-50' },
  { key: 'proposal', label: 'Proposal', color: 'bg-indigo-50' },
  { key: 'negotiation', label: 'Negotiation', color: 'bg-yellow-50' },
  { key: 'closed_won', label: 'Closed Won', color: 'bg-green-50' },
  { key: 'closed_lost', label: 'Closed Lost', color: 'bg-red-50' },
]

interface KanbanBoardProps {
  deals: Deal[]
}

export function KanbanBoard({ deals }: KanbanBoardProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeDeal, setActiveDeal] = useState<Deal | null>(null)
  const [localDeals, setLocalDeals] = useState<Deal[]>(deals)

  // Sync when parent data changes
  if (JSON.stringify(deals.map((d) => d.id)) !== JSON.stringify(localDeals.map((d) => d.id))) {
    setLocalDeals(deals)
  }

  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 250, tolerance: 5 } })
  )

  function getDealsByStage(stage: DealStage) {
    return localDeals.filter((d) => d.stage === stage)
  }

  function handleDragStart(event: DragStartEvent) {
    const deal = localDeals.find((d) => d.id === event.active.id)
    setActiveDeal(deal ?? null)
  }

  function handleDragOver(event: DragOverEvent) {
    const { active, over } = event
    if (!over) return

    const activeId = active.id as string
    const overId = over.id as string

    // overId might be a stage column id or a deal id
    const targetStage = STAGES.find((s) => s.key === overId)?.key

    if (targetStage) {
      setLocalDeals((prev) =>
        prev.map((d) => (d.id === activeId ? { ...d, stage: targetStage } : d))
      )
      return
    }

    // overId is a deal id — find its stage
    const overDeal = localDeals.find((d) => d.id === overId)
    if (!overDeal) return
    if (overDeal.stage === localDeals.find((d) => d.id === activeId)?.stage) return

    setLocalDeals((prev) =>
      prev.map((d) => (d.id === activeId ? { ...d, stage: overDeal.stage } : d))
    )
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    setActiveDeal(null)

    if (!over) return

    const activeId = active.id as string
    const deal = localDeals.find((d) => d.id === activeId)
    const originalDeal = deals.find((d) => d.id === activeId)

    if (!deal || !originalDeal) return
    if (deal.stage === originalDeal.stage) return

    // Persist the new stage
    try {
      await dealsApi.update(activeId, { stage: deal.stage })
      queryClient.invalidateQueries({ queryKey: dealKeys.lists() })
    } catch {
      // Rollback on error
      setLocalDeals((prev) =>
        prev.map((d) => (d.id === activeId ? { ...d, stage: originalDeal.stage } : d))
      )
    }
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-4">
        {STAGES.map((stage) => {
          const stageDeals = getDealsByStage(stage.key)
          return (
            <div
              key={stage.key}
              id={stage.key}
              className="flex w-72 shrink-0 flex-col rounded-xl border border-slate-200 bg-slate-50"
            >
              {/* Column header */}
              <div className="flex items-center justify-between px-4 py-3 border-b border-slate-200">
                <span className="text-sm font-semibold text-slate-700">{stage.label}</span>
                <span className="rounded-full bg-slate-200 px-2 py-0.5 text-xs font-medium text-slate-600">
                  {stageDeals.length}
                </span>
              </div>

              {/* Cards */}
              <SortableContext
                id={stage.key}
                items={stageDeals.map((d) => d.id)}
                strategy={verticalListSortingStrategy}
              >
                <div className="flex flex-col gap-2 p-3 min-h-[100px]">
                  {stageDeals.map((deal) => (
                    <SortableDealCard
                      key={deal.id}
                      deal={deal}
                      onClick={() => navigate(`/deals/${deal.id}`)}
                    />
                  ))}
                </div>
              </SortableContext>
            </div>
          )
        })}
      </div>

      <DragOverlay>
        {activeDeal && <DealCard deal={activeDeal} />}
      </DragOverlay>
    </DndContext>
  )
}
