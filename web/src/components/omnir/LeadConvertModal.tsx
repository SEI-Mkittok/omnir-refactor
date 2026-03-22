import { useState } from 'react'
import { Loader2, ArrowRightLeft, UserCheck, TrendingUp } from 'lucide-react'
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
import { Input } from '@/components/ui/Input'
import { cn } from '@/lib/utils'
import { useConvertLead } from '@/hooks/useLeads'
import { useCreateDeal } from '@/hooks/useDeals'
import type { Lead } from '@/api/types'

type Mode = 'contact' | 'contact_and_deal'

interface LeadConvertModalProps {
  lead: Lead
  open: boolean
  onClose: () => void
  onConverted?: (contactId: string) => void
}

export function LeadConvertModal({ lead, open, onClose, onConverted }: LeadConvertModalProps) {
  const navigate = useNavigate()
  const convertLead = useConvertLead()
  const createDeal = useCreateDeal()

  const [mode, setMode] = useState<Mode>('contact')
  const [dealTitle, setDealTitle] = useState(`${lead.first_name} ${lead.last_name} — Deal`)
  const [dealValue, setDealValue] = useState('')

  const leadName = `${lead.first_name} ${lead.last_name}`
  const isPending = convertLead.isPending || createDeal.isPending

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const result = await convertLead.mutateAsync({ id: lead.id, payload: {} })
    const contactId = result.contact.id

    if (mode === 'contact_and_deal' && dealTitle.trim()) {
      await createDeal.mutateAsync({
        title: dealTitle.trim(),
        value_cents: dealValue ? Math.round(parseFloat(dealValue) * 100) : 0,
        contact_id: contactId,
      })
    }

    onConverted?.(contactId)
    onClose()
    navigate(`/contacts?openId=${contactId}`)
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

          {/* Mode selector */}
          <div className="space-y-2">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Conversion type</p>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setMode('contact')}
                className={cn(
                  'flex flex-col items-start gap-1.5 rounded-lg border-2 p-3 text-left transition-colors',
                  mode === 'contact'
                    ? 'border-[var(--border-focus)] bg-[var(--color-primary-light)]'
                    : 'border-slate-200 hover:border-slate-300',
                )}
              >
                <UserCheck className={cn('h-4 w-4', mode === 'contact' ? 'text-[var(--color-primary)]' : 'text-slate-400')} />
                <span className="text-sm font-medium text-slate-900">Contact only</span>
                <span className="text-xs text-slate-500">Create contact, no deal</span>
              </button>
              <button
                type="button"
                onClick={() => setMode('contact_and_deal')}
                className={cn(
                  'flex flex-col items-start gap-1.5 rounded-lg border-2 p-3 text-left transition-colors',
                  mode === 'contact_and_deal'
                    ? 'border-[var(--border-focus)] bg-[var(--color-primary-light)]'
                    : 'border-slate-200 hover:border-slate-300',
                )}
              >
                <TrendingUp className={cn('h-4 w-4', mode === 'contact_and_deal' ? 'text-[var(--color-primary)]' : 'text-slate-400')} />
                <span className="text-sm font-medium text-slate-900">Contact + Deal</span>
                <span className="text-xs text-slate-500">Create contact and deal</span>
              </button>
            </div>
          </div>

          {/* Deal fields (shown only in contact_and_deal mode) */}
          {mode === 'contact_and_deal' && (
            <div className="space-y-3 rounded-lg border border-slate-200 bg-slate-50 p-3">
              <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Deal details</p>
              <div>
                <label className="mb-1 block text-xs font-medium text-slate-700">Deal title</label>
                <Input
                  value={dealTitle}
                  onChange={(e) => setDealTitle(e.target.value)}
                  placeholder="Enter deal title"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-slate-700">Value ($)</label>
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
