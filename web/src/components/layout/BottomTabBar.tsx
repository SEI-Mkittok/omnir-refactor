import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Users, TrendingUp, LifeBuoy, MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'

const tabs = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Home' },
  { to: '/contacts', icon: Users, label: 'Contacts' },
  { to: '/deals', icon: TrendingUp, label: 'Deals' },
  { to: '/tickets', icon: LifeBuoy, label: 'Tickets' },
  { to: '/reports', icon: MoreHorizontal, label: 'More' },
]

export function BottomTabBar() {
  return (
    <nav className="flex border-t border-slate-200 bg-white safe-area-pb lg:hidden">
      {tabs.map(({ to, icon: Icon, label }) => (
        <NavLink
          key={to}
          to={to}
          className={({ isActive }) =>
            cn(
              'flex flex-1 flex-col items-center justify-center gap-1 py-2 text-xs font-medium transition-colors',
              isActive ? 'text-[#1B3A4B]' : 'text-slate-400'
            )
          }
          style={{ minHeight: '56px' }}
        >
          {({ isActive }) => (
            <>
              <Icon className={cn('h-5 w-5', isActive && 'text-[#1B3A4B]')} />
              <span>{label}</span>
            </>
          )}
        </NavLink>
      ))}
    </nav>
  )
}
