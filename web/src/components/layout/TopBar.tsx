import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Menu, Search, LogOut, User, Users, Building2, TrendingUp, Ticket } from 'lucide-react'
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
import { TenantSwitcher } from '@/components/layout/TenantSwitcher'
import { NotificationBell } from '@/components/layout/NotificationBell'
import type { Contact, Account, Deal, Ticket as TicketType } from '@/api/types'

interface TopBarProps {
  onMenuClick: () => void
  breadcrumb?: string
}

const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform)
const shortcutLabel = isMac ? '⌘K' : 'Ctrl+K'

export function TopBar({ onMenuClick, breadcrumb }: TopBarProps) {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const [searchValue, setSearchValue] = useState('')
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [focusedIndex, setFocusedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const debouncedSearch = useDebounce(searchValue, 300)

  // Cmd+K / Ctrl+K
  useEffect(() => {
    function handleGlobalKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        inputRef.current?.focus()
        inputRef.current?.select()
      }
    }
    document.addEventListener('keydown', handleGlobalKey)
    return () => document.removeEventListener('keydown', handleGlobalKey)
  }, [])

  const { data, isFetching } = useQuery({
    queryKey: ['search-inline', debouncedSearch],
    queryFn: () => searchApi.search(debouncedSearch),
    enabled: debouncedSearch.trim().length >= 2,
    staleTime: 30_000,
  })

  const contacts = data?.contacts ?? []
  const accounts = data?.accounts ?? []
  const deals    = data?.deals    ?? []
  const tickets  = data?.tickets  ?? []
  const hasResults = contacts.length + accounts.length + deals.length + tickets.length > 0

  const contactItems = contacts.slice(0, 3)
  const accountItems = accounts.slice(0, 3)
  const dealItems    = deals.slice(0, 3)
  const ticketItems  = tickets.slice(0, 3)
  const allItems = [
    ...contactItems.map((c: Contact) => `/contacts?openId=${c.id}`),
    ...accountItems.map((a: Account) => `/accounts?openId=${a.id}`),
    ...dealItems.map((d: Deal) => `/deals?openId=${d.id}`),
    ...ticketItems.map((t: TicketType) => `/tickets?openId=${t.id}`),
  ]

  useEffect(() => {
    setDropdownOpen(debouncedSearch.trim().length >= 2)
    setFocusedIndex(-1)
  }, [debouncedSearch])

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
      setFocusedIndex(-1)
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (!dropdownOpen) setDropdownOpen(true)
      setFocusedIndex((p) => Math.min(p + 1, allItems.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setFocusedIndex((p) => Math.max(p - 1, 0))
    } else if (e.key === 'Enter' && focusedIndex >= 0 && focusedIndex < allItems.length) {
      e.preventDefault()
      navigateTo(allItems[focusedIndex])
    }
  }

  function navigateTo(path: string) {
    navigate(path)
    setSearchValue('')
    setDropdownOpen(false)
    setFocusedIndex(-1)
  }

  async function handleLogout() {
    const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'
    await fetch(`${AUTH_BASE}/auth/logout`, { method: 'POST', credentials: 'include' }).catch(() => {})
    logout()
    navigate('/login')
  }

  const accountOffset = contactItems.length
  const dealOffset    = contactItems.length + accountItems.length
  const ticketOffset  = contactItems.length + accountItems.length + dealItems.length

  return (
    <header
      className="fixed right-0 top-0 flex items-center justify-between border-b px-4 lg:px-5"
      style={{
        left: 'var(--sidebar-current-width)',
        height: 'var(--topbar-height)',
        background: 'var(--surface-card)',
        borderColor: 'var(--border-default)',
        zIndex: 'var(--z-topbar)' as unknown as number,
        transition: 'left 200ms ease',
      }}
    >
      {/* Left: hamburger (mobile) + breadcrumb */}
      <div className="flex items-center gap-3 min-w-0">
        <button
          className="md:hidden rounded-md p-1.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]"
          style={{ color: 'var(--text-secondary)' }}
          onClick={onMenuClick}
          aria-label="Open navigation menu"
          aria-expanded="false"
          aria-controls="mobile-drawer"
        >
          <Menu className="h-5 w-5" />
        </button>

        {breadcrumb && (
          <span
            className="hidden sm:block text-sm font-medium truncate"
            style={{ color: 'var(--text-secondary)' }}
          >
            {breadcrumb}
          </span>
        )}
      </div>

      {/* Centre / Right: search + bells + user */}
      <div className="flex items-center gap-3 ml-4">
        {/* Search pill */}
        <div ref={containerRef} className="relative">
          <form onSubmit={handleSearch}>
            <div
              className="flex items-center gap-2 px-3 cursor-text"
              style={{
                width: '220px',
                height: '36px',
                background: 'var(--surface-app)',
                border: '1px solid var(--border-default)',
                borderRadius: 'var(--radius-pill)',
              }}
              onClick={() => inputRef.current?.focus()}
            >
              <Search
                className="h-3.5 w-3.5 shrink-0 pointer-events-none"
                style={{ color: 'var(--text-label)' }}
                aria-hidden="true"
              />
              {isFetching && debouncedSearch.length >= 2
                ? <Spinner className="h-3 w-3 shrink-0" />
                : null}
              <input
                ref={inputRef}
                value={searchValue}
                onChange={(e) => setSearchValue(e.target.value)}
                onFocus={() => {
                  if (debouncedSearch.trim().length >= 2) setDropdownOpen(true)
                }}
                onKeyDown={handleKeyDown}
                placeholder={`Search… (${shortcutLabel})`}
                className="flex-1 bg-transparent text-sm focus:outline-none min-w-0"
                style={{
                  color: 'var(--text-primary)',
                  '::placeholder': { color: 'var(--text-label)' },
                } as React.CSSProperties}
                autoComplete="off"
              />
            </div>
          </form>

          {/* Results dropdown */}
          {dropdownOpen && (
            <div
              className="absolute right-0 top-full mt-1 w-80 rounded-[var(--radius-md)] border border-[var(--border-default)] bg-[var(--surface-card)] overflow-hidden"
              style={{ boxShadow: 'var(--shadow-popover)', zIndex: 'var(--z-modal)' as unknown as number }}
            >
              {isFetching && !hasResults ? (
                <div className="flex items-center justify-center py-6">
                  <Spinner className="h-4 w-4" />
                </div>
              ) : !hasResults ? (
                <div className="px-4 py-3 text-sm" style={{ color: 'var(--text-label)' }}>
                  No results for "{debouncedSearch}"
                </div>
              ) : (
                <div className="max-h-80 overflow-y-auto">
                  {contactItems.length > 0 && (
                    <ResultGroup label="Contacts" icon={Users} items={contactItems.map((c: Contact) => ({ id: c.id, primary: `${c.first_name} ${c.last_name}`, secondary: c.email, path: `/contacts?openId=${c.id}` }))} offset={0} focusedIndex={focusedIndex} onSelect={navigateTo} />
                  )}
                  {accountItems.length > 0 && (
                    <ResultGroup label="Accounts" icon={Building2} items={accountItems.map((a: Account) => ({ id: a.id, primary: a.name, secondary: a.domain ?? a.industry ?? '', path: `/accounts?openId=${a.id}` }))} offset={accountOffset} focusedIndex={focusedIndex} onSelect={navigateTo} />
                  )}
                  {dealItems.length > 0 && (
                    <ResultGroup label="Deals" icon={TrendingUp} items={dealItems.map((d: Deal) => ({ id: d.id, primary: d.title, secondary: d.stage.replace('_', ' '), path: `/deals?openId=${d.id}` }))} offset={dealOffset} focusedIndex={focusedIndex} onSelect={navigateTo} />
                  )}
                  {ticketItems.length > 0 && (
                    <ResultGroup label="Tickets" icon={Ticket} items={ticketItems.map((t: TicketType) => ({ id: t.id, primary: t.subject, secondary: `${t.status} · ${t.priority}`, path: `/tickets?openId=${t.id}` }))} offset={ticketOffset} focusedIndex={focusedIndex} onSelect={navigateTo} />
                  )}
                  <button
                    className="w-full border-t px-4 py-2.5 text-left text-xs font-medium transition-colors hover:bg-[var(--surface-app)]"
                    style={{ borderColor: 'var(--border-subtle)', color: 'var(--color-primary)' }}
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

        <NotificationBell />
        <TenantSwitcher />

        {/* User avatar */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              className="flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--border-focus)]"
              style={{ background: 'var(--color-primary)', color: 'var(--text-on-dark)' }}
              onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--color-primary-hover)' }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--color-primary)' }}
            >
              {user ? getInitials(user.name) : <User className="h-4 w-4" />}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            {user && (
              <>
                <DropdownMenuLabel>{user.name}</DropdownMenuLabel>
                <DropdownMenuLabel className="font-normal pt-0" style={{ color: 'var(--text-secondary)' }}>
                  {user.email}
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
              </>
            )}
            <DropdownMenuItem onClick={handleLogout} style={{ color: 'var(--color-danger)' }}>
              <LogOut className="mr-2 h-4 w-4" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}

// ── Result group helper ────────────────────────────────────────────────────────

interface ResultItem { id: string; primary: string; secondary: string; path: string }

function ResultGroup({ label, icon: Icon, items, offset, focusedIndex, onSelect }: {
  label: string; icon: React.ElementType; items: ResultItem[]
  offset: number; focusedIndex: number; onSelect: (p: string) => void
}) {
  return (
    <div>
      <div
        className="flex items-center gap-1.5 px-3 py-1.5"
        style={{ background: 'var(--surface-app)' }}
      >
        <Icon className="h-3 w-3" style={{ color: 'var(--text-label)' }} aria-hidden="true" />
        <span
          className="text-xs font-semibold uppercase tracking-wide"
          style={{ color: 'var(--text-label)' }}
        >
          {label}
        </span>
      </div>
      {items.map((item, i) => (
        <button
          key={item.id}
          onClick={() => onSelect(item.path)}
          className="w-full text-left px-4 py-2 transition-colors"
          style={{
            background: offset + i === focusedIndex ? 'var(--color-primary-light)' : undefined,
          }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--color-primary-light)' }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = offset + i === focusedIndex ? 'var(--color-primary-light)' : ''
          }}
        >
          <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{item.primary}</p>
          {item.secondary && <p className="text-xs truncate" style={{ color: 'var(--text-label)' }}>{item.secondary}</p>}
        </button>
      ))}
    </div>
  )
}
