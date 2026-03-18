import { useState } from 'react'
import { Loader2, AlertCircle } from 'lucide-react'
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
import type { ViewEntityType, ViewFilters, SavedView } from '@/api/types'

interface SaveViewModalProps {
  open: boolean
  onClose: () => void
  entityType: ViewEntityType
  currentFilters: ViewFilters
  onSaved?: (view: SavedView) => void
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
  const [error, setError] = useState<string | null>(null)
  const createView = useCreateView()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setError(null)
    try {
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
      onSaved?.(view)
      onClose()
    } catch {
      setError('Failed to save view. Please try again.')
    }
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setName('')
      setIsShared(false)
      setIsPinned(true)
      setError(null)
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

          {error && (
            <div className="flex items-center gap-1.5 rounded-md bg-red-50 px-3 py-2 text-xs text-red-700 ring-1 ring-red-200">
              <AlertCircle className="h-3.5 w-3.5 shrink-0" />
              {error}
            </div>
          )}

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
