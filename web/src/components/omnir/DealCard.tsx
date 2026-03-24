import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Calendar } from 'lucide-react'
import { formatCurrency, formatDate } from '@/lib/utils'
import type { Deal } from '@/api/types'
import { cn } from '@/lib/utils'

interface DealCardProps {
  deal: Deal
  onClick?: () => void
  isDragging?: boolean
  /** keyboard-dnd: this card is "picked up" */
  isKeyboardActive?: boolean
}

function dealAgeDays(deal: Deal): number {
  const created = new Date(deal.created_at)
  return Math.floor((Date.now() - created.getTime()) / 86_400_000)
}

export function DealCard({ deal, onClick, isDragging, isKeyboardActive }: DealCardProps) {
  const age = dealAgeDays(deal)

  return (
    <div
      className={cn(
        'rounded-lg border bg-white p-4 cursor-pointer transition-all',
        'hover:border-[#1B3A4B]/30 hover:shadow-md',
        isDragging
          ? 'opacity-50 shadow-lg border-[#1B3A4B]/40'
          : 'border-[#E5E7EB] shadow-sm',
        isKeyboardActive && 'ring-2 ring-[#1B3A4B] border-[#1B3A4B]/40'
      )}
      onClick={onClick}
    >
      {/* Category tag row */}
      <div className="flex items-start justify-between mb-2">
        {deal.account?.name ? (
          <span className="text-[10px] font-bold text-[#7C8DB0] uppercase tracking-tight px-1.5 py-0.5 bg-[#E8EDF2] rounded">
            {deal.account.name.slice(0, 14)}
          </span>
        ) : (
          <span className="text-[10px] font-bold text-[#7C8DB0] uppercase tracking-tight px-1.5 py-0.5 bg-[#E8EDF2] rounded">
            Deal
          </span>
        )}
        {age > 14 && (
          <span className="text-[9px] font-semibold text-[#F59E0B] bg-amber-50 px-1.5 py-0.5 rounded">
            {age}d
          </span>
        )}
      </div>

      {/* Title */}
      <h4 className="text-sm font-semibold text-[#1A1D23] leading-snug mb-1">{deal.title}</h4>

      {/* Contact */}
      {deal.contact && (
        <p className="text-xs text-[#6B7280] mb-3">
          {deal.contact.first_name} {deal.contact.last_name}
        </p>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between pt-3 border-t border-[#E5E7EB]">
        <span className="text-sm font-bold text-[#1B3A4B]">
          {formatCurrency(deal.value_cents / 100, deal.currency)}
        </span>
        {deal.expected_close_date ? (
          <div className="flex items-center gap-1 text-[10px] text-[#6B7280]">
            <Calendar className="h-3 w-3" />
            <span>{formatDate(deal.expected_close_date)}</span>
          </div>
        ) : (
          <span className="text-[10px] text-[#6B7280]">{age}d old</span>
        )}
      </div>
    </div>
  )
}

interface SortableDealCardProps {
  deal: Deal
  onClick?: () => void
  isKeyboardActive?: boolean
}

export function SortableDealCard({ deal, onClick, isKeyboardActive }: SortableDealCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: deal.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div ref={setNodeRef} style={style}>
      <div {...attributes} {...listeners}>
        <DealCard
          deal={deal}
          onClick={onClick}
          isDragging={isDragging}
          isKeyboardActive={isKeyboardActive}
        />
      </div>
    </div>
  )
}
