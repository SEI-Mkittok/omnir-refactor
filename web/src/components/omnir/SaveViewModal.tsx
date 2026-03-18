import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useCreateView } from '@/hooks/useViews'
import type { ViewEntityType, ViewFilters } from '@/api/types'

interface SaveViewModalProps {
  open: boolean
  onClose: () => void
  entityType: ViewEntityType
  currentFilters: ViewFilters
  onSaved?: (viewId: string) => void
}

export function SaveViewModal({
  open,
  onClose,
  entityType,
  currentFilters,
  onSaved,
}: SaveViewModalProps) {
  const [name, setName] = useState('')
  const [isShared, setIsShared] = useState(false)
  const [isPinned, setIsPinned] = useState(true)
  const createView = useCreateView()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    const view = await createView.mutateAsync({
      entity_type: entityType,
      name: name.trim(),
      filters: currentFilters,
      is_shared: isShared,
      is_pinned: isPinned,
    })
    setName('')
    setIsShared(false)
    setIsPinned(true)
    onSaved?.(view.id)
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setName('')
      setIsShared(false)
      setIsPinned(true)
      onClose()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Save current filters as view</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              View name <span className="text-red-500">*</span>
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Hot leads this month"
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-700">
              <input
                type="checkbox"
                checked={isShared}
                onChange={(e) => setIsShared(e.target.checked)}
                className="rounded border-slate-300 text-indigo-600"
              />
              Share with team
            </label>
            <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-700">
              <input
                type="checkbox"
                checked={isPinned}
                onChange={(e) => setIsPinned(e.target.checked)}
                className="rounded border-slate-300 text-indigo-600"
              />
              Pin to view bar
            </label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim() || createView.isPending}>
              {createView.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Save view
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
