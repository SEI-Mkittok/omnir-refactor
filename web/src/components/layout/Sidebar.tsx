import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Users, Building2, TrendingUp, UserCog, X, BarChart2, UserRound, SlidersHorizontal, KeyRound, Clock, LifeBuoy, ShieldCheck, Mail, FileText } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const baseNavItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/contacts', icon: Users, label: 'Contacts' },
  { to: '/leads', icon: UserRound, label: 'Leads' },
  { to: '/accounts', icon: Building2, label: 'Accounts' },
  { to: '/deals', icon: TrendingUp, label: 'Deals' },
  { to: '/tickets', icon: LifeBuoy, label: 'Help Desk' },
  { to: '/quotes', icon: FileText, label: 'Quotes' },
  { to: '/sequences', icon: Mail, label: 'Sequences' },
  { to: '/reports', icon: BarChart2, label: 'Reports' },
]

const adminNavItems = [
  { to: '/users', icon: UserCog, label: 'Users' },
  { to: '/settings/custom-fields', icon: SlidersHorizontal, label: 'Custom Fields' },
  { to: '/api-keys', icon: KeyRound, label: 'API Keys' },
  { to: '/settings/sla', icon: Clock, label: 'SLA Policies' },
  { to: '/admin/audit', icon: ShieldCheck, label: 'Audit Log' },
]

interface SidebarProps {
  open?: boolean
  onClose?: () => void
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const user = useAuthStore((s) => s.user)
  const isAdmin = user?.role === 'admin'
  const navItems = isAdmin ? [...baseNavItems, ...adminNavItems] : baseNavItems
  return (
    <>
      {/* Mobile overlay */}
      {open !== undefined && (
        <div
          className={cn(
            'fixed inset-0 z-30 bg-black/50 lg:hidden transition-opacity duration-200',
            open ? 'opacity-100' : 'pointer-events-none opacity-0'
          )}
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed left-0 top-0 z-40 flex h-full w-64 flex-col bg-slate-900 transition-transform duration-200 lg:static lg:translate-x-0',
          open !== undefined
            ? open
              ? 'translate-x-0'
              : '-translate-x-full'
            : undefined
        )}
      >
        {/* Logo */}
        <div className="flex h-16 items-center justify-between px-6">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
              <TrendingUp className="h-5 w-5 text-white" />
            </div>
            <span className="text-lg font-bold text-white">Omnir</span>
          </div>
          {onClose && (
            <button
              className="rounded-md p-1 text-slate-400 hover:text-white lg:hidden"
              onClick={onClose}
            >
              <X className="h-5 w-5" />
            </button>
          )}
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 px-3 py-4">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              onClick={onClose}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-indigo-700 text-white'
                    : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                )
              }
            >
              <Icon className="h-5 w-5 shrink-0" />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* Bottom branding */}
        <div className="border-t border-slate-800 px-4 py-4">
          <p className="text-xs text-slate-500">Omnir CRM v0.1.0</p>
        </div>
      </aside>
    </>
  )
}
