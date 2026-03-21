import { useState } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, Building2, TrendingUp, UserPlus, Ticket,
  BarChart3, Settings, LogOut, ChevronLeft, ChevronRight, Plus,
  UserCog, SlidersHorizontal, KeyRound, Clock, ShieldCheck, X,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/DropdownMenu'

// ── Nav structure ──────────────────────────────────────────────────────────────

const NAV_GROUPS = [
  {
    label: 'MAIN HUB',
    id: 'nav-group-main',
    items: [
      { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
    ],
  },
  {
    label: 'SALES OPERATIONS',
    id: 'nav-group-sales',
    items: [
      { to: '/deals',    icon: TrendingUp, label: 'Pipeline' },
      { to: '/contacts', icon: Users,      label: 'Contacts' },
      { to: '/accounts', icon: Building2,  label: 'Accounts' },
      { to: '/leads',    icon: UserPlus,   label: 'Leads' },
    ],
  },
  {
    label: 'SERVICE',
    id: 'nav-group-service',
    items: [
      { to: '/tickets', icon: Ticket, label: 'Tickets' },
    ],
  },
  {
    label: 'INSIGHTS',
    id: 'nav-group-insights',
    items: [
      { to: '/reports', icon: BarChart3, label: 'Reports' },
    ],
  },
]

const ADMIN_ITEMS = [
  { to: '/users',                icon: UserCog,          label: 'Users' },
  { to: '/settings/custom-fields', icon: SlidersHorizontal, label: 'Custom Fields' },
  { to: '/api-keys',             icon: KeyRound,         label: 'API Keys' },
  { to: '/settings/sla',         icon: Clock,            label: 'SLA Policies' },
  { to: '/admin/audit',          icon: ShieldCheck,      label: 'Audit Log' },
]

const NEW_ENTRY_OPTIONS = [
  { label: 'New Lead',    to: '/leads?new=1' },
  { label: 'New Contact', to: '/contacts?new=1' },
  { label: 'New Ticket',  to: '/tickets?new=1' },
  { label: 'New Account', to: '/accounts?new=1' },
]

// ── Component ──────────────────────────────────────────────────────────────────

interface SidebarProps {
  /** Mobile: whether the drawer is open */
  mobileOpen?: boolean
  onMobileClose?: () => void
}

export function Sidebar({ mobileOpen, onMobileClose }: SidebarProps) {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const { sidebarCollapsed, toggleSidebar } = useUIStore()
  const [newEntryOpen, setNewEntryOpen] = useState(false)

  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin'

  async function handleLogout() {
    const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'
    await fetch(`${AUTH_BASE}/auth/logout`, { method: 'POST', credentials: 'include' }).catch(() => {})
    logout()
    navigate('/login')
  }

  function handleNewEntry(to: string) {
    setNewEntryOpen(false)
    onMobileClose?.()
    navigate(to)
  }

  const collapsed = sidebarCollapsed

  return (
    <>
      {/* Mobile scrim */}
      {mobileOpen !== undefined && (
        <div
          className={cn(
            'fixed inset-0 bg-black/32 transition-opacity duration-280 md:hidden',
            mobileOpen
              ? 'opacity-100 pointer-events-auto'
              : 'opacity-0 pointer-events-none'
          )}
          style={{ zIndex: 'var(--z-overlay)' as unknown as number }}
          onClick={onMobileClose}
          aria-hidden="true"
        />
      )}

      {/* Sidebar panel */}
      <aside
        aria-label="Main navigation"
        className={cn(
          'fixed left-0 top-0 h-screen flex flex-col',
          'bg-[var(--surface-sidebar)] border-r border-[var(--border-subtle)]',
          'transition-all duration-200 ease-in-out',
          // desktop width driven by collapse state
          'hidden md:flex',
          collapsed ? 'w-[60px]' : 'w-[210px]',
        )}
        style={{ zIndex: 'var(--z-sidebar)' as unknown as number }}
      >
        <SidebarInner
          collapsed={collapsed}
          isAdmin={isAdmin}
          newEntryOpen={newEntryOpen}
          setNewEntryOpen={setNewEntryOpen}
          onNewEntry={handleNewEntry}
          onLogout={handleLogout}
          onToggleCollapse={toggleSidebar}
          onClose={undefined}
        />
      </aside>

      {/* Mobile drawer — always full-width 280px, slides in */}
      <aside
        aria-label="Main navigation"
        id="mobile-drawer"
        className={cn(
          'fixed left-0 top-0 h-screen flex flex-col md:hidden',
          'bg-[var(--surface-sidebar)] border-r border-[var(--border-subtle)]',
          'w-[280px] transition-transform duration-[280ms]',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
        )}
        style={{
          zIndex: 'var(--z-modal)' as unknown as number,
          transitionTimingFunction: 'cubic-bezier(0.32, 0, 0.15, 1)',
        }}
      >
        <SidebarInner
          collapsed={false}
          isAdmin={isAdmin}
          newEntryOpen={newEntryOpen}
          setNewEntryOpen={setNewEntryOpen}
          onNewEntry={handleNewEntry}
          onLogout={handleLogout}
          onToggleCollapse={undefined}
          onClose={onMobileClose}
        />
      </aside>
    </>
  )
}

// ── Inner content (shared between desktop + mobile) ────────────────────────────

interface InnerProps {
  collapsed: boolean
  isAdmin: boolean
  newEntryOpen: boolean
  setNewEntryOpen: (v: boolean) => void
  onNewEntry: (to: string) => void
  onLogout: () => void
  onToggleCollapse?: () => void
  onClose?: () => void
}

function SidebarInner({
  collapsed,
  isAdmin,
  newEntryOpen,
  setNewEntryOpen,
  onNewEntry,
  onLogout,
  onToggleCollapse,
  onClose,
}: InnerProps) {
  return (
    <div className="flex h-full flex-col">
      {/* ── Header ── */}
      <div
        className="flex items-center justify-between border-b border-[var(--border-subtle)] px-4"
        style={{ height: 'var(--topbar-height)' }}
      >
        {!collapsed && (
          <div className="flex items-center gap-2 min-w-0">
            {/* Logo mark */}
            <div
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
              style={{ background: 'var(--color-primary)' }}
              aria-hidden="true"
            >
              <span className="text-sm font-bold text-white select-none">P</span>
            </div>
            <div className="min-w-0">
              <p
                className="font-bold leading-tight truncate"
                style={{
                  fontSize: '16px',
                  color: 'var(--text-primary)',
                  letterSpacing: 'var(--letter-spacing-tight)',
                }}
              >
                PraestOS
              </p>
              <p
                className="uppercase tracking-[0.08em] font-semibold"
                style={{ fontSize: '9px', color: 'var(--text-label)' }}
              >
                CRM ENTERPRISE
              </p>
            </div>
          </div>
        )}
        {collapsed && (
          <div
            className="mx-auto flex h-8 w-8 items-center justify-center rounded-lg"
            style={{ background: 'var(--color-primary)' }}
          >
            <span className="text-sm font-bold text-white select-none">P</span>
          </div>
        )}
        {/* Mobile close button */}
        {onClose && (
          <button
            onClick={onClose}
            className="ml-auto rounded-md p-1 text-[var(--text-label)] hover:text-[var(--text-primary)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]"
            aria-label="Close navigation menu"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* ── Nav groups ── */}
      <nav
        className="flex-1 overflow-y-auto py-3 px-2"
        role="navigation"
        aria-label="Main navigation"
      >
        {NAV_GROUPS.map((group) => (
          <div key={group.id} role="group" aria-labelledby={group.id}>
            {!collapsed && (
              <p
                id={group.id}
                className="px-3 pb-1 pt-4 uppercase tracking-[0.05em] font-semibold"
                style={{ fontSize: '11px', color: 'var(--text-label)' }}
                aria-hidden="true"
              >
                {group.label}
              </p>
            )}
            {group.items.map((item) => (
              <NavItem
                key={item.to}
                {...item}
                collapsed={collapsed}
              />
            ))}
          </div>
        ))}

        {/* Admin group */}
        {isAdmin && (
          <div role="group" aria-labelledby="nav-group-admin">
            {!collapsed && (
              <p
                id="nav-group-admin"
                className="px-3 pb-1 pt-4 uppercase tracking-[0.05em] font-semibold"
                style={{ fontSize: '11px', color: 'var(--text-label)' }}
                aria-hidden="true"
              >
                ADMIN
              </p>
            )}
            {ADMIN_ITEMS.map((item) => (
              <NavItem key={item.to} {...item} collapsed={collapsed} />
            ))}
          </div>
        )}

        {/* Divider */}
        <div
          className="my-2 mx-3"
          style={{ height: '1px', background: 'var(--border-subtle)' }}
          aria-hidden="true"
        />

        {/* Settings + Sign Out */}
        <NavItem to="/settings" icon={Settings} label="Settings" collapsed={collapsed} />
        <button
          onClick={onLogout}
          className={cn(
            'group flex w-full items-center rounded-[var(--radius-sm)] transition-colors duration-120',
            'text-[var(--text-secondary)] hover:bg-[var(--surface-app)] hover:text-[var(--color-danger)]',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]',
            collapsed ? 'justify-center px-1.5 py-1.5' : 'gap-2.5 px-3 py-1.5',
          )}
          style={{ height: '36px', fontSize: '14px', fontWeight: 500 }}
        >
          <LogOut className="h-[18px] w-[18px] shrink-0" aria-hidden="true" />
          {!collapsed && <span>Sign Out</span>}
          {collapsed && <span className="sr-only">Sign Out</span>}
        </button>
      </nav>

      {/* ── Collapse toggle (desktop only) ── */}
      {onToggleCollapse && (
        <button
          onClick={onToggleCollapse}
          className={cn(
            'flex items-center justify-center border-t border-[var(--border-subtle)]',
            'text-[var(--text-label)] hover:text-[var(--text-primary)] transition-colors',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]',
          )}
          style={{ height: '40px' }}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed
            ? <ChevronRight className="h-4 w-4" />
            : <ChevronLeft className="h-4 w-4" />
          }
        </button>
      )}

      {/* ── +New Entry CTA ── */}
      <div
        className="border-t border-[var(--border-subtle)] p-3"
        style={{ background: 'var(--surface-sidebar)' }}
      >
        <DropdownMenu open={newEntryOpen} onOpenChange={setNewEntryOpen}>
          <DropdownMenuTrigger asChild>
            <button
              className={cn(
                'flex w-full items-center justify-center gap-2 rounded-[var(--radius-md)]',
                'font-semibold text-white transition-colors active:scale-[0.98]',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary)] focus-visible:ring-offset-3',
              )}
              style={{
                height: '40px',
                fontSize: '14px',
                background: 'var(--color-primary)',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--color-primary-hover)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'var(--color-primary)'
              }}
              aria-haspopup="true"
              aria-expanded={newEntryOpen}
              aria-label="Create new entry"
            >
              <Plus className="h-4 w-4 shrink-0" aria-hidden="true" />
              {!collapsed && <span>+ New Entry</span>}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            side="top"
            align="start"
            className="w-48"
            style={{ boxShadow: 'var(--shadow-popover)' }}
          >
            {NEW_ENTRY_OPTIONS.map((opt) => (
              <DropdownMenuItem
                key={opt.to}
                onClick={() => onNewEntry(opt.to)}
                style={{ fontSize: '14px', color: 'var(--text-primary)' }}
              >
                {opt.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}

// ── NavItem ────────────────────────────────────────────────────────────────────

function NavItem({
  to,
  icon: Icon,
  label,
  collapsed,
  onClick,
}: {
  to: string
  icon: React.ElementType
  label: string
  collapsed: boolean
  onClick?: () => void
}) {
  return (
    <NavLink
      to={to}
      onClick={onClick}
      aria-label={collapsed ? label : undefined}
      className={({ isActive }) =>
        cn(
          'flex w-full items-center rounded-[var(--radius-sm)] transition-colors duration-120',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]',
          collapsed ? 'justify-center px-1.5 py-1.5' : 'gap-2.5 px-3 py-1.5',
          isActive
            ? 'bg-[var(--color-primary-light)] text-[var(--color-primary)] font-semibold shadow-[inset_3px_0_0_var(--color-primary)]'
            : 'text-[var(--text-secondary)] hover:bg-[var(--color-primary-light)] hover:text-[var(--text-primary)] font-medium',
        )
      }
      style={{ height: '36px', fontSize: '14px' }}
    >
      {({ isActive }) => (
        <>
          <Icon
            className="h-[18px] w-[18px] shrink-0"
            aria-hidden="true"
            style={{ color: isActive ? 'var(--color-primary)' : 'currentColor' }}
          />
          {!collapsed && <span>{label}</span>}
        </>
      )}
    </NavLink>
  )
}
