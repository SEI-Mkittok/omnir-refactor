import { useState } from 'react'
import { Plus, Trash2, Loader2, GripVertical } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useDealLineItems, useUpsertLineItem, useDeleteLineItem } from '@/hooks/useProducts'
import type { DealLineItem, UpsertLineItemRequest } from '@/api/types'

const fmt = (n: number, currency = 'USD') =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(n)

function LineItemRow({
  item,
  dealId,
  onEdit,
}: {
  item: DealLineItem
  dealId: string
  onEdit: (item: DealLineItem) => void
}) {
  const del = useDeleteLineItem(dealId)
  return (
    <div className="flex items-center gap-2 py-2 border-b last:border-0">
      <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-sm truncate">{item.name}</div>
        <div className="text-xs text-muted-foreground">
          {item.quantity} × {fmt(item.unit_price)}
          {item.discount_pct > 0 && ` − ${item.discount_pct}%`}
        </div>
      </div>
      <div className="text-sm font-medium">{fmt(item.subtotal)}</div>
      <div className="flex gap-1">
        <Button size="sm" variant="ghost" onClick={() => onEdit(item)}>
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536M9 11l6-6 3 3-6 6H9v-3z" />
          </svg>
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            if (confirm(`Remove "${item.name}"?`)) del.mutate(item.id)
          }}
        >
          {del.isPending ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
          ) : (
            <Trash2 className="w-3.5 h-3.5 text-red-500" />
          )}
        </Button>
      </div>
    </div>
  )
}

interface LineItemFormProps {
  dealId: string
  existing?: DealLineItem
  onClose: () => void
}

function LineItemForm({ dealId, existing, onClose }: LineItemFormProps) {
  const [form, setForm] = useState<UpsertLineItemRequest>({
    id: existing?.id,
    name: existing?.name ?? '',
    quantity: existing?.quantity ?? 1,
    unit_price: existing?.unit_price ?? 0,
    discount_pct: existing?.discount_pct ?? 0,
  })
  const upsert = useUpsertLineItem(dealId)

  const set =
    (field: keyof UpsertLineItemRequest) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({
        ...f,
        [field]: ['quantity', 'unit_price', 'discount_pct'].includes(field)
          ? parseFloat(e.target.value) || 0
          : e.target.value,
      }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await upsert.mutateAsync(form)
    onClose()
  }

  return (
    <form onSubmit={handleSubmit} className="border rounded-lg p-3 space-y-3 bg-muted/30">
      <div>
        <label className="block text-xs font-medium mb-1">Description *</label>
        <Input value={form.name} onChange={set('name')} required placeholder="e.g. Enterprise License" />
      </div>
      <div className="grid grid-cols-3 gap-2">
        <div>
          <label className="block text-xs font-medium mb-1">Qty</label>
          <Input type="number" min="0" step="0.01" value={form.quantity} onChange={set('quantity')} required />
        </div>
        <div>
          <label className="block text-xs font-medium mb-1">Unit Price</label>
          <Input type="number" min="0" step="0.01" value={form.unit_price} onChange={set('unit_price')} required />
        </div>
        <div>
          <label className="block text-xs font-medium mb-1">Discount %</label>
          <Input type="number" min="0" max="100" step="0.01" value={form.discount_pct} onChange={set('discount_pct')} />
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" size="sm" variant="outline" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" size="sm" disabled={upsert.isPending}>
          {upsert.isPending && <Loader2 className="w-3.5 h-3.5 mr-1 animate-spin" />}
          {existing ? 'Save' : 'Add'}
        </Button>
      </div>
    </form>
  )
}

interface LineItemsPanelProps {
  dealId: string
}

export function LineItemsPanel({ dealId }: LineItemsPanelProps) {
  const { data: items = [], isLoading } = useDealLineItems(dealId)
  const [showForm, setShowForm] = useState(false)
  const [editingItem, setEditingItem] = useState<DealLineItem | null>(null)

  const total = items.reduce((sum, li) => sum + li.subtotal, 0)

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">Line Items</h3>
        {!showForm && !editingItem && (
          <Button size="sm" variant="ghost" onClick={() => setShowForm(true)}>
            <Plus className="w-3.5 h-3.5 mr-1" />
            Add
          </Button>
        )}
      </div>

      {isLoading ? (
        <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
      ) : (
        <div>
          {items.map((item) =>
            editingItem?.id === item.id ? (
              <LineItemForm
                key={item.id}
                dealId={dealId}
                existing={item}
                onClose={() => setEditingItem(null)}
              />
            ) : (
              <LineItemRow
                key={item.id}
                item={item}
                dealId={dealId}
                onEdit={setEditingItem}
              />
            )
          )}

          {showForm && (
            <LineItemForm dealId={dealId} onClose={() => setShowForm(false)} />
          )}

          {items.length > 0 && (
            <div className="flex justify-between pt-2 text-sm font-semibold border-t">
              <span>Total</span>
              <span>{fmt(total)}</span>
            </div>
          )}

          {items.length === 0 && !showForm && (
            <p className="text-xs text-muted-foreground py-2">No line items yet.</p>
          )}
        </div>
      )}
    </div>
  )
}
