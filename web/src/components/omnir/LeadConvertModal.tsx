import { Loader2, ArrowRightLeft } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { useConvertLead } from '@/hooks/useLeads'
import type { Lead } from '@/api/types'

interface LeadConvertModalProps {
  lead: Lead
  open: boolean
  onClose: () => void
  onConverted?: (contactId: string) => void
}

export function LeadConvertModal({ lead, open, onClose, onConverted }: LeadConvertModalProps) {
  const convertLead = useConvertLead()

  const leadName = `${lead.first_name} ${lead.last_name}`

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const result = await convertLead.mutateAsync({ id: lead.id, payload: {} })
    onConverted?.(result.contact.id)
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-5 w-5 text-indigo-600" />
            Convert Lead to Contact
          </DialogTitle>
          <DialogDescription>
            <strong>{leadName}</strong> will become a contact. This cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Summary */}
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm">
            <div className="grid grid-cols-2 gap-1 text-slate-600">
              <span className="font-medium text-slate-500">Name</span>
              <span>{leadName}</span>
              <span className="font-medium text-slate-500">Email</span>
              <span>{lead.email}</span>
              {lead.company && (
                <>
                  <span className="font-medium text-slate-500">Company</span>
                  <span>{lead.company}</span>
                </>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={convertLead.isPending}>
              {convertLead.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Convert
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
