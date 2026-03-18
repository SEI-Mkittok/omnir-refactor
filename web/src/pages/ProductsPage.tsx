import { useState } from 'react'
import { Package, Plus, Pencil, PowerOff, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { Table, type Column } from '@/components/ui/Table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { useProducts, useCreateProduct, useUpdateProduct, useDeactivateProduct } from '@/hooks/useProducts'
import { formatDate } from '@/lib/utils'
import type { Product, CreateProductRequest, UpdateProductRequest } from '@/api/types'

const EMPTY_FORM: CreateProductRequest = {
  name: '',
  description: '',
  sku: '',
  price: 0,
  currency: 'USD',
}

function ProductForm({
  open,
  onClose,
  product,
}: {
  open: boolean
  onClose: () => void
  product?: Product
}) {
  const isEdit = !!product
  const [form, setForm] = useState<CreateProductRequest>(
    product
      ? {
          name: product.name,
          description: product.description ?? '',
          sku: product.sku ?? '',
          price: product.price,
          currency: product.currency,
        }
      : EMPTY_FORM
  )
  const createProduct = useCreateProduct()
  const updateProduct = useUpdateProduct()

  const set = (field: keyof CreateProductRequest) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: field === 'price' ? parseFloat(e.target.value) || 0 : e.target.value }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const payload = {
      ...form,
      description: form.description || undefined,
      sku: form.sku || undefined,
    }
    if (isEdit && product) {
      await updateProduct.mutateAsync({ id: product.id, ...payload } as UpdateProductRequest & { id: string })
    } else {
      await createProduct.mutateAsync(payload)
    }
    onClose()
  }

  const isPending = createProduct.isPending || updateProduct.isPending

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Product' : 'New Product'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">Name *</label>
            <Input value={form.name} onChange={set('name')} required />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">SKU</label>
            <Input value={form.sku} onChange={set('sku')} placeholder="e.g. PRD-001" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium mb-1">Price *</label>
              <Input
                type="number"
                min="0"
                step="0.01"
                value={form.price}
                onChange={set('price')}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Currency</label>
              <Input value={form.currency} onChange={set('currency')} placeholder="USD" />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Description</label>
            <Input value={form.description} onChange={set('description')} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="w-4 h-4 mr-1 animate-spin" />}
              {isEdit ? 'Save' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export function ProductsPage() {
  const [showAll, setShowAll] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const { data: products = [], isLoading } = useProducts(!showAll)
  const deactivate = useDeactivateProduct()

  const columns: Column<Product>[] = [
    { key: 'name', header: 'Name', render: (p) => <span className="font-medium">{p.name}</span> },
    { key: 'sku', header: 'SKU', render: (p) => p.sku ?? '—' },
    {
      key: 'price',
      header: 'Price',
      render: (p) =>
        new Intl.NumberFormat('en-US', { style: 'currency', currency: p.currency }).format(p.price),
    },
    {
      key: 'is_active',
      header: 'Status',
      render: (p) => (
        <Badge variant={p.is_active ? 'green' : 'gray'}>
          {p.is_active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    { key: 'created_at', header: 'Created', render: (p) => formatDate(p.created_at) },
    {
      key: 'actions',
      header: '',
      render: (p) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            onClick={(e) => {
              e.stopPropagation()
              setEditing(p)
            }}
          >
            <Pencil className="w-3.5 h-3.5" />
          </Button>
          {p.is_active && (
            <Button
              size="sm"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation()
                if (confirm(`Deactivate "${p.name}"?`)) deactivate.mutate(p.id)
              }}
            >
              <PowerOff className="w-3.5 h-3.5 text-red-500" />
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Package className="w-5 h-5 text-muted-foreground" />
          <h1 className="text-xl font-semibold">Products</h1>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5 text-sm text-muted-foreground cursor-pointer">
            <input
              type="checkbox"
              checked={showAll}
              onChange={(e) => setShowAll(e.target.checked)}
              className="rounded"
            />
            Show inactive
          </label>
          <Button size="sm" onClick={() => setShowForm(true)}>
            <Plus className="w-4 h-4 mr-1" />
            New Product
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <Table columns={columns} data={products} keyExtractor={(p) => p.id}
          emptyTitle="No products found." />
      )}

      {(showForm || editing) && (
        <ProductForm
          open
          product={editing ?? undefined}
          onClose={() => {
            setShowForm(false)
            setEditing(null)
          }}
        />
      )}
    </div>
  )
}
