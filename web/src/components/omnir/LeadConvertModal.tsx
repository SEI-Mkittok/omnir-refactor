import { useState } from 'react'
import { Loader2, ArrowRightLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
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
  const navigate = useNavigate()
  const convertLead = useConvertLead()

  const leadName = `${lead.first_name} ${lead.last_name}`
  const [error, setError] = useState('')
  const isPending = convertLead.isPending

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const result = await convertLead.mutateAsync({ id: lead.id, payload: {} })
      const contactId = result.contact.id
      onConverted?.(contactId)
      onClose()
      navigate(`/contacts?openId=${contactId}`)
    } catch {
      setError('Lead conversion failed. Please try again.')
    }
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-5 w-5 text-[var(--color-primary)]" />
            Convert Lead to Contact
          </DialogTitle>
          <DialogDescription>
            <strong>{leadName}</strong> will become a contact. This cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Lead summary */}
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

          <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
            Conversion creates:
            <ul className="mt-2 list-disc pl-5">
              <li>Contact</li>
              <li>Account</li>
              <li>Deal</li>
            </ul>
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Convert
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
