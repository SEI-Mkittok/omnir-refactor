import { useState } from 'react'
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
import { Input } from '@/components/ui/Input'
import { useConvertLead } from '@/hooks/useLeads'
import type { Lead } from '@/api/types'

interface LeadConvertModalProps {
  lead: Lead
  open: boolean
  onClose: () => void
  onConverted?: (contactId: string) => void
}

export function LeadConvertModal({ lead, open, onClose, onConverted }: LeadConvertModalProps) {
  const [createDeal, setCreateDeal] = useState(false)
  const [dealTitle, setDealTitle] = useState('')
  const [dealValue, setDealValue] = useState('')
  const convertLead = useConvertLead()

  const leadName = `${lead.first_name} ${lead.last_name}`

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const result = await convertLead.mutateAsync({
      id: lead.id,
      payload: {
        create_deal: createDeal,
        deal_title: createDeal && dealTitle ? dealTitle : undefined,
        deal_value: createDeal && dealValue ? parseFloat(dealValue) : undefined,
      },
    })
    onConverted?.(result.contact.id)
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setCreateDeal(false)
      setDealTitle('')
      setDealValue('')
      onClose()
    }
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

          {/* Optional deal creation */}
          <div className="flex items-center gap-2">
            <input
              id="create-deal"
              type="checkbox"
              checked={createDeal}
              onChange={(e) => setCreateDeal(e.target.checked)}
              className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
            />
            <label htmlFor="create-deal" className="text-sm font-medium text-slate-700">
              Also create a deal
            </label>
          </div>

          {createDeal && (
            <div className="space-y-3 rounded-lg border border-indigo-100 bg-indigo-50 p-3">
              <div>
                <label className="mb-1 block text-xs font-medium text-slate-700">Deal title</label>
                <Input
                  value={dealTitle}
                  onChange={(e) => setDealTitle(e.target.value)}
                  placeholder={`Deal with ${leadName}`}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-slate-700">
                  Value (optional)
                </label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={dealValue}
                  onChange={(e) => setDealValue(e.target.value)}
                  placeholder="0.00"
                />
              </div>
            </div>
          )}

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
