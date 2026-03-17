import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Menu, Search, LogOut, User, Users, Building2, TrendingUp } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import { getInitials } from '@/lib/utils'
import { useDebounce } from '@/hooks/useDebounce'
import { searchApi } from '@/api/search'
import { Spinner } from '@/components/ui/Spinner'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '@/components/ui/DropdownMenu'
import type { Contact, Account, Deal } from '@/api/types'

interface TopBarProps {
  onMenuClick: () => void
}

export function TopBar({ onMenuClick }: TopBarProps) {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const [searchValue, setSearchValue] = useState('')
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const debouncedSearch = useDebounce(searchValue, 300)

  const { data, isFetching } = useQuery({
    queryKey: ['search-inline', debouncedSearch],
    queryFn: () => searchApi.search(debouncedSearch),
    enabled: debouncedSearch.trim().length >= 2,
    staleTime: 30_000,
  })

  const contacts = data?.contacts ?? []
  const accounts = data?.accounts ?? []
  const deals = data?.deals ?? []
  const hasResults = contacts.length + accounts.length + deals.length > 0

  // Open dropdown when we have a search value (2+ chars)
  useEffect(() => {
    setDropdownOpen(debouncedSearch.trim().length >= 2)
  }, [debouncedSearch])

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    if (searchValue.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchValue.trim())}`)
      setSearchValue('')
      setDropdownOpen(false)
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      setDropdownOpen(false)
    }
  }

  function navigateTo(path: string) {
    navigate(path)
    setSearchValue('')
    setDropdownOpen(false)
  }

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <header className="flex h-16 items-center justify-between border-b border-slate-200 bg-white px-4 lg:px-6">
      {/* Left: menu button + search */}
      <div className="flex flex-1 items-center gap-4">
        <button
          className="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-900 lg:hidden"
          onClick={onMenuClick}
        >
          <Menu className="h-5 w-5" />
        </button>

        <div ref={containerRef} className="relative flex-1 max-w-md">
          <form onSubmit={handleSearch}>
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400 pointer-events-none" />
            {isFetching && debouncedSearch.length >= 2 && (
              <span className="absolute right-3 top-1/2 -translate-y-1/2">
                <Spinner className="h-3 w-3" />
              </span>
            )}
            <input
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              onFocus={() => {
                if (debouncedSearch.trim().length >= 2) setDropdownOpen(true)
              }}
              onKeyDown={handleKeyDown}
              placeholder="Search contacts, accounts, deals…"
              className="h-9 w-full rounded-md border border-slate-200 bg-slate-50 pl-9 pr-3 text-sm placeholder:text-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-indigo-500"
              autoComplete="off"
            />
          </form>

          {/* Live results dropdown */}
          {dropdownOpen && (
            <div className="absolute top-full mt-1 w-full rounded-md border border-slate-200 bg-white shadow-lg z-50 overflow-hidden">
              {isFetching && !hasResults ? (
                <div className="flex items-center justify-center py-6">
                  <Spinner className="h-4 w-4" />
                </div>
              ) : !hasResults ? (
                <div className="px-4 py-3 text-sm text-slate-400">
                  No results for "{debouncedSearch}"
                </div>
              ) : (
                <div className="max-h-80 overflow-y-auto">
                  {contacts.length > 0 && (
                    <ResultGroup
                      label="Contacts"
                      icon={Users}
                      items={contacts.slice(0, 3).map((c: Contact) => ({
                        id: c.id,
                        primary: `${c.first_name} ${c.last_name}`,
                        secondary: c.email,
                        path: '/contacts',
                      }))}
                      onSelect={navigateTo}
                    />
                  )}
                  {accounts.length > 0 && (
                    <ResultGroup
                      label="Accounts"
                      icon={Building2}
                      items={accounts.slice(0, 3).map((a: Account) => ({
                        id: a.id,
                        primary: a.name,
                        secondary: a.domain ?? a.industry ?? '',
                        path: '/accounts',
                      }))}
                      onSelect={navigateTo}
                    />
                  )}
                  {deals.length > 0 && (
                    <ResultGroup
                      label="Deals"
                      icon={TrendingUp}
                      items={deals.slice(0, 3).map((d: Deal) => ({
                        id: d.id,
                        primary: d.title,
                        secondary: d.stage.replace('_', ' '),
                        path: '/deals',
                      }))}
                      onSelect={navigateTo}
                    />
                  )}
                  <button
                    className="w-full border-t border-slate-100 px-4 py-2.5 text-left text-xs font-medium text-indigo-600 hover:bg-indigo-50 transition-colors"
                    onClick={() => {
                      navigate(`/search?q=${encodeURIComponent(searchValue.trim())}`)
                      setSearchValue('')
                      setDropdownOpen(false)
                    }}
                  >
                    See all results for "{searchValue}"
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Right: user menu */}
      <div className="flex items-center gap-3 ml-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-600 text-xs font-semibold text-white hover:bg-indigo-700 transition-colors">
              {user ? getInitials(user.name) : <User className="h-4 w-4" />}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            {user && (
              <>
                <DropdownMenuLabel>{user.name}</DropdownMenuLabel>
                <DropdownMenuLabel className="font-normal text-slate-500 pt-0">
                  {user.email}
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
              </>
            )}
            <DropdownMenuItem onClick={handleLogout} className="text-red-600 focus:text-red-700">
              <LogOut className="mr-2 h-4 w-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}

// ---- Helpers ----

interface ResultItem {
  id: string
  primary: string
  secondary: string
  path: string
}

function ResultGroup({
  label,
  icon: Icon,
  items,
  onSelect,
}: {
  label: string
  icon: React.ElementType
  items: ResultItem[]
  onSelect: (path: string) => void
}) {
  return (
    <div>
      <div className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-50">
        <Icon className="h-3 w-3 text-slate-400" />
        <span className="text-xs font-semibold text-slate-400 uppercase tracking-wide">{label}</span>
      </div>
      {items.map((item) => (
        <button
          key={item.id}
          onClick={() => onSelect(item.path)}
          className="w-full text-left px-4 py-2 hover:bg-indigo-50 transition-colors"
        >
          <p className="text-sm font-medium text-slate-800">{item.primary}</p>
          {item.secondary && (
            <p className="text-xs text-slate-400 truncate">{item.secondary}</p>
          )}
        </button>
      ))}
    </div>
  )
}
