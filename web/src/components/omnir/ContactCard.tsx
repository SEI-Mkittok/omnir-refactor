import { User, Mail, Phone, Building2 } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Card, CardContent } from '@/components/ui/Card'
import type { Contact, ContactStage } from '@/api/types'

const stageBadgeVariant: Record<ContactStage, 'blue' | 'yellow' | 'green' | 'gray'> = {
  lead: 'blue',
  prospect: 'yellow',
  customer: 'green',
  churned: 'gray',
}

const stageLabel: Record<ContactStage, string> = {
  lead: 'Lead',
  prospect: 'Prospect',
  customer: 'Customer',
  churned: 'Churned',
}

interface ContactCardProps {
  contact: Contact
  onClick?: () => void
}

export function ContactCard({ contact, onClick }: ContactCardProps) {
  return (
    <Card
      className="cursor-pointer hover:shadow-md transition-shadow"
      onClick={onClick}
    >
      <CardContent className="pt-5">
        <div className="flex items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-100">
            <User className="h-5 w-5 text-indigo-600" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center justify-between gap-2">
              <p className="truncate font-semibold text-slate-900">
                {contact.first_name} {contact.last_name}
              </p>
              <Badge variant={stageBadgeVariant[contact.stage]}>
                {stageLabel[contact.stage]}
              </Badge>
            </div>
            {contact.title && (
              <p className="mt-0.5 truncate text-xs text-slate-500">{contact.title}</p>
            )}
            <div className="mt-2 space-y-1">
              {contact.email && (
                <div className="flex items-center gap-1.5 text-xs text-slate-500">
                  <Mail className="h-3.5 w-3.5 shrink-0" />
                  <span className="truncate">{contact.email}</span>
                </div>
              )}
              {contact.phone && (
                <div className="flex items-center gap-1.5 text-xs text-slate-500">
                  <Phone className="h-3.5 w-3.5 shrink-0" />
                  <span>{contact.phone}</span>
                </div>
              )}
              {contact.account?.name && (
                <div className="flex items-center gap-1.5 text-xs text-slate-500">
                  <Building2 className="h-3.5 w-3.5 shrink-0" />
                  <span className="truncate">{contact.account.name}</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export { stageBadgeVariant, stageLabel }
