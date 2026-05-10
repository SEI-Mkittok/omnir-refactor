import React from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, Building2, TrendingUp, UserPlus, Ticket,
  BarChart3, Settings, LogOut, ChevronLeft, ChevronRight, Plus,
  UserCog, SlidersHorizontal, KeyRound, Clock, ShieldCheck, CreditCard,
  BookOpen, FileText, Mail, Zap, CalendarDays, Rocket, Inbox, Plug, X, Hash, Send, PanelTopOpen, MenuSquare,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { updateOnboarding } from '@/api/onboarding'
import { useMenuConfigSettings } from '@/hooks/useAdminSettings'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/DropdownMenu'

const NAV_GROUPS = [
  {
    label: 'MAIN HUB',
    items: [
      { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
    ],
  },
  {
    label: 'SALES OPERATIONS',
    items: [
      { to: '/deals', icon: TrendingUp, label: 'Pipeline' },
      { to: '/contacts', icon: Users, label: 'Contacts' },
      { to: '/accounts', icon: Building2, label: 'Accounts' },
      { to: '/leads', icon: UserPlus, label: 'Leads' },
      { to: '/quotes', icon: FileText, label: 'Quotes' },
      { to: '/sequences', icon: Mail, label: 'Sequences' },
      { to: '/inbox', icon: Inbox, label: 'Email' },
    ],
  },
  {
    label: 'SERVICE',
    items: [
      { to: '/tickets', icon: Ticket, label: 'Tickets' },
      { to: '/kb', icon: BookOpen, label: 'Knowledge Base' },
    ],
  },
  {
    label: 'INSIGHTS',
    items: [
      { to: '/reports', icon: BarChart3, label: 'Reports' },
      { to: '/dashboards', icon: LayoutDashboard, label: 'Dashboards' },
    ],
  },
  {
    label: 'TOOLS',
    items: [
      { to: '/automations', icon: Zap, label: 'Automations' },
      { to: '/calendar', icon: CalendarDays, label: 'Calendar' },
    ],
  },
]

const ADMIN_GROUP = {
  label: 'ADMIN',
  items: [
    { to: '/users', icon: UserCog, label: 'Users' },
    { to: '/settings/custom-fields', icon: SlidersHorizontal, label: 'Custom Fields' },
    { to: '/settings/numbering', icon: Hash, label: 'Numbering' },
    { to: '/settings/company', icon: Building2, label: 'Company Profile' },
    { to: '/settings/portal', icon: PanelTopOpen, label: 'Portal Config' },
    { to: '/settings/outgoing-server', icon: Send, label: 'Outgoing Server' },
    { to: '/settings/config-editor', icon: SlidersHorizontal, label: 'Config Editor' },
    { to: '/settings/menu', icon: MenuSquare, label: 'Menu Config' },
    { to: '/api-keys', icon: KeyRound, label: 'API Keys' },
    { to: '/settings/sla', icon: Clock, label: 'SLA Policies' },
    { to: '/admin/audit', icon: ShieldCheck, label: 'Audit Log' },
    { to: '/settings/billing', icon: CreditCard, label: 'Billing' },
    { to: '/settings/integrations', icon: Plug, label: 'Integrations' },
  ],
}

const ROUTE_MENU_KEYS: Record<string, string> = {
  '/dashboard': 'dashboard',
  '/deals': 'deals',
  '/contacts': 'contacts',
  '/accounts': 'accounts',
  '/leads': 'leads',
  '/quotes': 'quotes',
  '/sequences': 'sequences',
  '/inbox': 'inbox',
  '/tickets': 'tickets',
  '/kb': 'kb',
  '/reports': 'reports',
  '/dashboards': 'dashboards',
  '/automations': 'automations',
  '/calendar': 'calendar',
}

interface SidebarProps {
  mobileOpen?: boolean
  open?: boolean
  onClose?: () => void
  onMobileClose?: () => void
}

export function Sidebar({ mobileOpen, open, onClose, onMobileClose }: SidebarProps) {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const activeOrg = useAuthStore((s) => s.activeOrg)
  const { sidebarCollapsed, toggleSidebar, onboardingDismissed, setOnboardingDismissed } = useUIStore()
  const collapsed = sidebarCollapsed
  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin'
  const { data: menuSettings } = useMenuConfigSettings()

  async function handleDismissOnboarding(e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    setOnboardingDismissed(true)
    await updateOnboarding({ dismissed: true }).catch(() => {})
  }

  // Support both prop naming conventions
  const isOpen = mobileOpen ?? open
  const handleClose = onMobileClose ?? onClose

  const menuConfig = menuSettings?.menu_config ?? {}
  const groups = (isAdmin ? [...NAV_GROUPS, ADMIN_GROUP] : NAV_GROUPS).map((group) => ({
    ...group,
    items: group.items.filter((item) => {
      const key = ROUTE_MENU_KEYS[item.to]
      if (!key) return true
      return menuConfig[key] ?? true
    }),
  }))

  function NavItem({ to, icon: Icon, label }: { to: string; icon: React.ElementType; label: string }) {
    return (
      <NavLink
        to={to}
        onClick={handleClose}
        className={({ isActive }) =>
          cn(
            'flex items-center gap-2.5 rounded px-3 py-1.5 text-sm font-medium transition-colors duration-100',
            isActive
              ? 'bg-[#E8EDF2] text-[#1B3A4B] font-semibold shadow-[inset_3px_0_0_#1B3A4B]'
              : 'text-[#6B7280] hover:bg-[#E8EDF2] hover:text-[#1A1D23]',
            collapsed && 'justify-center px-2'
          )
        }
      >
        <Icon className="h-[18px] w-[18px] shrink-0" aria-hidden="true" />
        {!collapsed && <span>{label}</span>}
      </NavLink>
    )
  }

  return (
    <>
      {/* Mobile scrim */}
      {isOpen !== undefined && (
        <div
          className={cn(
            'fixed inset-0 z-30 bg-black/40 lg:hidden transition-opacity duration-200',
            isOpen ? 'opacity-100' : 'pointer-events-none opacity-0'
          )}
          onClick={handleClose}
          aria-hidden="true"
        />
      )}

      {/* Sidebar */}
      <aside
        aria-label="Main navigation"
        className={cn(
          'fixed left-0 top-0 z-40 flex h-full flex-col transition-all duration-200',
          'border-r border-[#E5E7EB]',
          collapsed ? 'w-[60px]' : 'w-[210px]',
          isOpen !== undefined
            ? isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
            : undefined
        )}
        style={{ background: 'var(--surface-sidebar)' }}
      >
        {/* Header */}
        <div
          className={cn(
            'flex h-16 shrink-0 items-center border-b border-[#F0F1F3]',
            collapsed ? 'justify-center px-2' : 'gap-2.5 px-4'
          )}
        >
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[#1B3A4B] text-white font-bold text-sm select-none">
            {(activeOrg?.name ?? 'O').charAt(0).toUpperCase()}
          </div>
          {!collapsed && (
            <div className="flex flex-col leading-tight min-w-0">
              <span className="text-[16px] font-bold text-[#1A1D23] tracking-tight truncate">{activeOrg?.name ?? 'Omnir'}</span>
              <span className="text-[9px] font-semibold text-[#7C8DB0] uppercase tracking-[0.08em]">CRM ENTERPRISE</span>
            </div>
          )}
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-4" role="navigation">
          {groups.map((group) => (
            <div key={group.label} role="group" aria-labelledby={`nav-group-${group.label}`}>
              {!collapsed && (
                <span
                  id={`nav-group-${group.label}`}
                  className="block px-3 pb-1 text-[11px] font-semibold text-[#7C8DB0] uppercase tracking-[0.05em]"
                  aria-hidden="true"
                >
                  {group.label}
                </span>
              )}
              <div className="space-y-0.5">
                {group.items.map((item) => (
                  <NavItem key={item.to} {...item} />
                ))}
              </div>
            </div>
          ))}
        </nav>

        {/* Footer utilities */}
        <div className="shrink-0 border-t border-[#F0F1F3] px-2 py-2 space-y-0.5">
          {!onboardingDismissed && (
            <div className="relative flex items-center group">
              <NavLink
                to="/settings/onboarding"
                onClick={handleClose}
                className={({ isActive }) =>
                  cn(
                    'flex flex-1 items-center gap-2.5 rounded px-3 py-1.5 text-sm font-medium transition-colors duration-100',
                    isActive
                      ? 'bg-[#E8EDF2] text-[#1B3A4B] font-semibold shadow-[inset_3px_0_0_#1B3A4B]'
                      : 'text-[#6B7280] hover:bg-[#E8EDF2] hover:text-[#1A1D23]',
                    collapsed && 'justify-center px-2'
                  )
                }
              >
                <Rocket className="h-[18px] w-[18px] shrink-0" aria-hidden="true" />
                {!collapsed && <span>Getting Started</span>}
              </NavLink>
              {!collapsed && (
                <button
                  type="button"
                  onClick={handleDismissOnboarding}
                  aria-label="Dismiss Getting Started"
                  className="absolute right-1 hidden group-hover:flex items-center justify-center h-5 w-5 rounded text-[#6B7280] hover:text-[#1A1D23] hover:bg-[#D1D5DB]"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
          )}
          <NavItem to="/settings" icon={Settings} label="Settings" />
          <button
            onClick={async () => { await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' }).catch(() => {}); useAuthStore.getState().logout(); navigate('/login') }}
            className={cn(
              'flex w-full items-center gap-2.5 rounded px-3 py-1.5 text-sm font-medium text-[#6B7280] hover:bg-red-50 hover:text-red-600 transition-colors duration-100',
              collapsed && 'justify-center px-2'
            )}
          >
            <LogOut className="h-[18px] w-[18px] shrink-0" aria-hidden="true" />
            {!collapsed && <span>Sign Out</span>}
          </button>
        </div>

        {/* + New Entry CTA */}
        <div className="shrink-0 border-t border-[#F0F1F3] p-3">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                className={cn(
                  'flex w-full items-center justify-center gap-2 rounded-lg bg-[#1B3A4B] px-3 text-sm font-semibold text-white transition-colors hover:bg-[#152E3C] active:scale-[0.98]',
                  collapsed ? 'h-10 w-10 mx-auto' : 'h-10'
                )}
                aria-haspopup="true"
                aria-label="Create new entry"
              >
                <Plus className="h-4 w-4 shrink-0" />
                {!collapsed && '+ New Entry'}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent side="top" align="start" className="w-44">
              <DropdownMenuItem onClick={() => navigate('/leads?new=1')}>New Lead</DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate('/contacts?new=1')}>New Contact</DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate('/accounts?new=1')}>New Account</DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate('/tickets?new=1')}>New Ticket</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Collapse toggle */}
        <button
          onClick={toggleSidebar}
          className="absolute -right-3 top-20 hidden lg:flex h-6 w-6 items-center justify-center rounded-full border border-[#E5E7EB] bg-white text-[#6B7280] hover:text-[#1B3A4B] shadow-sm"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          aria-expanded={!collapsed}
        >
          {collapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronLeft className="h-3 w-3" />}
        </button>
      </aside>
    </>
  )
}
