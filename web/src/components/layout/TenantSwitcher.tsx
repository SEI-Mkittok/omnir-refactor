import { useNavigate } from 'react-router-dom'
import { Building2, ChevronDown, Users } from 'lucide-react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import { orgsApi } from '@/api/orgs'
import type { Org } from '@/api/types'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '@/components/ui/DropdownMenu'
import { Spinner } from '@/components/ui/Spinner'

// Hide the tenant switcher entirely in single-org mode.
const ORG_MODE = import.meta.env.VITE_ORG_MODE ?? 'saas'

export function TenantSwitcher() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  // Only render for super_admin in multi-tenant deployments.
  if (!user || user.role !== 'super_admin' || ORG_MODE === 'single') return null

  return <TenantSwitcherInner currentOrgId={user.org_id} navigate={navigate} />
}

function TenantSwitcherInner({
  currentOrgId,
  navigate,
}: {
  currentOrgId: string
  navigate: ReturnType<typeof useNavigate>
}) {
  const { setUser, user } = useAuthStore()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['orgs'],
    queryFn: () => orgsApi.list(),
    staleTime: 60_000,
  })

  const orgs: Org[] = data?.data ?? []
  const currentOrg = orgs.find((o) => o.id === currentOrgId)
  const triggerLabel = isLoading ? null : currentOrg?.name ?? 'Switch org'

  async function handleSwitch(orgId: string) {
    if (orgId === currentOrgId) return
    try {
      const res = await orgsApi.switchTo(orgId)
      // Merge new org context into the stored user; keep all other fields intact.
      if (user) {
        setUser({ ...user, org_id: res.org.id })
      }
      // Bust all cached queries so data refreshes under the new org context.
      await queryClient.invalidateQueries()
      navigate('/dashboard')
    } catch {
      // Switch failed — leave state unchanged.
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="flex items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors">
          <Building2 className="h-4 w-4 text-slate-400 shrink-0" />
          <span className="max-w-[140px] truncate">
            {isLoading ? <Spinner className="h-3 w-3" /> : triggerLabel}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-slate-400 shrink-0" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>Organizations</DropdownMenuLabel>
        <DropdownMenuSeparator />

        {isLoading ? (
          <div className="flex items-center justify-center py-4">
            <Spinner className="h-4 w-4" />
          </div>
        ) : orgs.length === 0 ? (
          <div className="px-3 py-2 text-sm text-slate-400">No organizations found</div>
        ) : (
          orgs.map((org) => {
            const isCurrent = org.id === currentOrgId
            return (
              <DropdownMenuItem
                key={org.id}
                onClick={() => handleSwitch(org.id)}
                className={isCurrent ? 'bg-indigo-50' : undefined}
              >
                <Building2 className="mr-2 h-4 w-4 shrink-0 text-slate-400" />
                <div className="flex-1 min-w-0">
                  <p className="truncate font-medium">{org.name}</p>
                  <p className="truncate text-xs text-slate-400">{org.slug}</p>
                </div>
                {org.user_count !== undefined && (
                  <span className="ml-2 flex items-center gap-0.5 text-xs text-slate-400 shrink-0">
                    <Users className="h-3 w-3" />
                    {org.user_count}
                  </span>
                )}
              </DropdownMenuItem>
            )
          })
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
