import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Building2, Calendar, GripVertical } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { formatCurrency, formatDate } from '@/lib/utils'
import type { Deal } from '@/api/types'
import { cn } from '@/lib/utils'

interface DealCardProps {
  deal: Deal
  onClick?: () => void
  isDragging?: boolean
}

export function DealCard({ deal, onClick, isDragging }: DealCardProps) {
  const probabilityColor =
    deal.probability !== undefined
      ? deal.probability >= 70
        ? 'green'
        : deal.probability >= 40
          ? 'yellow'
          : 'red'
      : 'gray'

  return (
    <div
      className={cn(
        'rounded-lg border border-slate-200 bg-white p-3 shadow-sm cursor-pointer hover:shadow-md transition-shadow',
        isDragging && 'opacity-50 shadow-lg'
      )}
      onClick={onClick}
    >
      <p className="font-medium text-slate-900 text-sm leading-snug">{deal.title}</p>

      <div className="mt-2 flex items-center justify-between">
        <span className="text-base font-bold text-indigo-600">
          {formatCurrency(deal.value, deal.currency)}
        </span>
        {deal.probability !== undefined && (
          <Badge variant={probabilityColor as 'green' | 'yellow' | 'red' | 'gray'}>
            {deal.probability}%
          </Badge>
        )}
      </div>

      <div className="mt-2 space-y-1">
        {deal.account?.name && (
          <div className="flex items-center gap-1.5 text-xs text-slate-500">
            <Building2 className="h-3 w-3 shrink-0" />
            <span className="truncate">{deal.account.name}</span>
          </div>
        )}
        {deal.close_date && (
          <div className="flex items-center gap-1.5 text-xs text-slate-500">
            <Calendar className="h-3 w-3 shrink-0" />
            <span>{formatDate(deal.close_date)}</span>
          </div>
        )}
      </div>
    </div>
  )
}

interface SortableDealCardProps {
  deal: Deal
  onClick?: () => void
}

export function SortableDealCard({ deal, onClick }: SortableDealCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: deal.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div ref={setNodeRef} style={style} className="relative group">
      <div
        {...attributes}
        {...listeners}
        className="absolute left-1 top-1/2 -translate-y-1/2 cursor-grab opacity-0 group-hover:opacity-50 active:cursor-grabbing z-10"
      >
        <GripVertical className="h-4 w-4 text-slate-400" />
      </div>
      <DealCard deal={deal} onClick={onClick} isDragging={isDragging} />
    </div>
  )
}
