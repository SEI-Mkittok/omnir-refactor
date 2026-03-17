import { Users, Building2, TrendingUp, DollarSign } from 'lucide-react'
import { useContacts } from '@/hooks/useContacts'
import { useAccounts } from '@/hooks/useAccounts'
import { useDeals } from '@/hooks/useDeals'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { formatCurrency } from '@/lib/utils'
import { Spinner } from '@/components/ui/Spinner'

interface StatCardProps {
  title: string
  value: string | number
  icon: React.ReactNode
  description?: string
  color: string
}

function StatCard({ title, value, icon, description, color }: StatCardProps) {
  return (
    <Card>
      <CardContent className="pt-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">{title}</p>
            <p className="mt-1 text-3xl font-bold text-slate-900">{value}</p>
            {description && (
              <p className="mt-1 text-xs text-slate-500">{description}</p>
            )}
          </div>
          <div className={`flex h-12 w-12 items-center justify-center rounded-xl ${color}`}>
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function DashboardPage() {
  const contacts = useContacts({ per_page: 1 })
  const accounts = useAccounts({ per_page: 1 })
  const deals = useDeals({ per_page: 100 })

  const totalDealValue = deals.data?.data.reduce((sum, d) => sum + d.value, 0) ?? 0
  const openDeals = deals.data?.data.filter(
    (d) => d.stage !== 'closed_won' && d.stage !== 'closed_lost'
  ).length ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>
        <p className="mt-1 text-sm text-slate-500">Your CRM overview at a glance</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Contacts"
          value={contacts.isLoading ? '—' : (contacts.data?.meta.total ?? 0)}
          icon={<Users className="h-6 w-6 text-blue-600" />}
          description="All contacts"
          color="bg-blue-50"
        />
        <StatCard
          title="Accounts"
          value={accounts.isLoading ? '—' : (accounts.data?.meta.total ?? 0)}
          icon={<Building2 className="h-6 w-6 text-purple-600" />}
          description="Active accounts"
          color="bg-purple-50"
        />
        <StatCard
          title="Open Deals"
          value={deals.isLoading ? '—' : openDeals}
          icon={<TrendingUp className="h-6 w-6 text-indigo-600" />}
          description="Deals in pipeline"
          color="bg-indigo-50"
        />
        <StatCard
          title="Pipeline Value"
          value={deals.isLoading ? '—' : formatCurrency(totalDealValue)}
          icon={<DollarSign className="h-6 w-6 text-green-600" />}
          description="Total deal value"
          color="bg-green-50"
        />
      </div>

      {/* Deal stage breakdown */}
      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Deals by Stage</CardTitle>
          </CardHeader>
          <CardContent>
            {deals.isLoading ? (
              <div className="flex justify-center py-8"><Spinner /></div>
            ) : (
              <div className="space-y-3">
                {(['lead', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost'] as const).map(
                  (stage) => {
                    const count = deals.data?.data.filter((d) => d.stage === stage).length ?? 0
                    const total = deals.data?.data.length ?? 1
                    const pct = Math.round((count / total) * 100)
                    const labels: Record<string, string> = {
                      lead: 'Lead',
                      qualified: 'Qualified',
                      proposal: 'Proposal',
                      negotiation: 'Negotiation',
                      closed_won: 'Closed Won',
                      closed_lost: 'Closed Lost',
                    }
                    const colors: Record<string, string> = {
                      lead: 'bg-slate-400',
                      qualified: 'bg-blue-400',
                      proposal: 'bg-indigo-400',
                      negotiation: 'bg-yellow-400',
                      closed_won: 'bg-green-500',
                      closed_lost: 'bg-red-400',
                    }
                    return (
                      <div key={stage}>
                        <div className="flex items-center justify-between text-sm mb-1">
                          <span className="text-slate-700">{labels[stage]}</span>
                          <span className="text-slate-500">{count}</span>
                        </div>
                        <div className="h-1.5 w-full rounded-full bg-slate-100">
                          <div
                            className={`h-1.5 rounded-full ${colors[stage]} transition-all`}
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                      </div>
                    )
                  }
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Contact Stages</CardTitle>
          </CardHeader>
          <CardContent>
            {contacts.isLoading ? (
              <div className="flex justify-center py-8"><Spinner /></div>
            ) : (
              <div className="space-y-3 text-sm text-slate-600">
                <p className="text-slate-500">
                  Total contacts: <span className="font-semibold text-slate-900">{contacts.data?.meta.total ?? 0}</span>
                </p>
                <p className="text-slate-500">
                  Total accounts: <span className="font-semibold text-slate-900">{accounts.data?.meta.total ?? 0}</span>
                </p>
                <p className="text-slate-500">
                  Pipeline value: <span className="font-semibold text-slate-900">{formatCurrency(totalDealValue)}</span>
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
