import { Building2, Globe, Users } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/Card'
import type { Account } from '@/api/types'

interface AccountCardProps {
  account: Account
  onClick?: () => void
}

export function AccountCard({ account, onClick }: AccountCardProps) {
  return (
    <Card
      className="cursor-pointer hover:shadow-md transition-shadow"
      onClick={onClick}
    >
      <CardContent className="pt-5">
        <div className="flex items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-slate-100">
            <Building2 className="h-5 w-5 text-slate-600" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate font-semibold text-slate-900">{account.name}</p>
            {account.industry && (
              <p className="mt-0.5 text-xs text-slate-500">{account.industry}</p>
            )}
            <div className="mt-2 space-y-1">
              {account.domain && (
                <div className="flex items-center gap-1.5 text-xs text-slate-500">
                  <Globe className="h-3.5 w-3.5 shrink-0" />
                  <span className="truncate">{account.domain}</span>
                </div>
              )}
              {account.size && (
                <div className="flex items-center gap-1.5 text-xs text-slate-500">
                  <Users className="h-3.5 w-3.5 shrink-0" />
                  <span>{account.size} employees</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
