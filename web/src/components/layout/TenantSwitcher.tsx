import { useNavigate } from 'react-router-dom'
import { Building2, ChevronDown, Plus } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import { orgsApi } from '@/api/orgs'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '@/components/ui/DropdownMenu'
import { Spinner } from '@/components/ui/Spinner'

// Hide the tenant switcher entirely when running in single-org mode.
const ORG_MODE = import.meta.env.VITE_ORG_MODE ?? 'saas'

export function TenantSwitcher() {
  const navigate = useNavigate()
  const { user, setUser } = useAuthStore()

  // Only render for super_admin; hide in single-org deployments.
  if (user?.role !== 'super_admin' || ORG_MODE === 'single') return null

  return <TenantSwitcherInner onSwitched={setUser} navigate={navigate} />
}

function TenantSwitcherInner({
  onSwitched,
  navigate,
}: {
  onSwitched: (user: import('@/api/types').User) => void
  navigate: ReturnType<typeof useNavigate>
}) {
  const { data, isLoading } = useQuery({
    queryKey: ['orgs'],
    queryFn: () => orgsApi.list(),
    staleTime: 60_000,
  })

  const orgs = data?.data ?? []

  async function handleSwitch(orgId: string) {
    try {
      const res = await orgsApi.switchTo(orgId)
      onSwitched(res.user)
      navigate('/dashboard')
    } catch {
      // Switching will be wired once Tyr ships the endpoint.
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="flex items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors">
          <Building2 className="h-4 w-4 text-slate-400" />
          <span className="max-w-[120px] truncate">
            {isLoading ? <Spinner className="h-3 w-3" /> : 'Switch org'}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel>Organizations</DropdownMenuLabel>
        <DropdownMenuSeparator />

        {isLoading ? (
          <div className="flex items-center justify-center py-4">
            <Spinner className="h-4 w-4" />
          </div>
        ) : orgs.length === 0 ? (
          <div className="px-3 py-2 text-sm text-slate-400">No organizations found</div>
        ) : (
          orgs.map((org) => (
            <DropdownMenuItem key={org.id} onClick={() => handleSwitch(org.id)}>
              <Building2 className="mr-2 h-4 w-4 text-slate-400" />
              <span className="truncate">{org.name}</span>
              <span className="ml-auto text-xs text-slate-400">{org.slug}</span>
            </DropdownMenuItem>
          ))
        )}

        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => navigate('/orgs/new')}>
          <Plus className="mr-2 h-4 w-4" />
          New organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
