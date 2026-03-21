import { useState, useEffect, useRef, useCallback } from 'react'
import { Link } from 'react-router-dom'
import {
  User,
  Mail,
  Phone,
  Building2,
  X,
  ArrowRightLeft,
  Search,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { cn } from '@/lib/utils'
import { contactsApi } from '@/api/contacts'
import { useUpdateTicketContact } from '@/hooks/useTickets'
import type { Contact, TicketStatus } from '@/api/types'

// ── Initials avatar ──────────────────────────────────────────────────────────

function InitialsAvatar({ name, size = 'md' }: { name: string; size?: 'sm' | 'md' }) {
  const initials = name
    .split(' ')
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
  return (
    <div
      className={cn(
        'flex items-center justify-center rounded-full bg-indigo-100 text-indigo-700 font-semibold shrink-0',
        size === 'sm' ? 'h-6 w-6 text-xs' : 'h-8 w-8 text-sm'
      )}
    >
      {initials}
    </div>
  )
}

// ── Contact search dropdown ──────────────────────────────────────────────────

interface ContactSearchProps {
  onSelect: (contact: Contact) => void
  onCancel: () => void
}

function ContactSearch({ onSelect, onCancel }: ContactSearchProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Contact[]>([])
  const [searching, setSearching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [highlighted, setHighlighted] = useState(-1)
  const [open, setOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  useEffect(() => {
    if (!query.trim()) {
      setResults([])
      setOpen(false)
      return
    }
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(async () => {
      setSearching(true)
      setError(null)
      try {
        const data = await contactsApi.search(query.trim())
        setResults(data)
        setOpen(true)
        setHighlighted(-1)
      } catch {
        setError('Search failed. Try again.')
        setResults([])
        setOpen(false)
      } finally {
        setSearching(false)
      }
    }, 300)
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [query])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCancel()
        return
      }
      if (!open) return
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setHighlighted((h) => Math.min(h + 1, results.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setHighlighted((h) => Math.max(h - 1, 0))
      } else if (e.key === 'Enter' && highlighted >= 0) {
        e.preventDefault()
        onSelect(results[highlighted])
      }
    },
    [open, results, highlighted, onCancel, onSelect]
  )

  return (
    <div className="relative">
      <div className="relative flex items-center">
        <Search className="absolute left-2.5 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
        <Input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Search contacts…"
          className="pl-8 pr-8 text-sm"
          aria-label="Search contacts"
          role="combobox"
          aria-expanded={open}
          aria-autocomplete="list"
        />
        {searching && (
          <Spinner
            size="sm"
            className="absolute right-2.5"
            aria-label="Searching…"
          />
        )}
      </div>

      {error && <p className="mt-1 text-xs text-red-500">{error}</p>}

      {open && (
        <div
          className="absolute z-10 mt-1 w-full rounded-lg border border-slate-200 bg-white shadow-md max-h-48 overflow-y-auto"
          role="listbox"
          aria-label="Contact search results"
        >
          {results.length === 0 ? (
            <p className="px-3 py-2 text-sm text-slate-400">No contacts found.</p>
          ) : (
            results.map((c, i) => (
              <button
                key={c.id}
                role="option"
                aria-selected={i === highlighted}
                className={cn(
                  'flex w-full items-center gap-2.5 px-3 py-2 text-left hover:bg-slate-50',
                  i === highlighted && 'bg-slate-50'
                )}
                onMouseDown={(e) => {
                  e.preventDefault()
                  onSelect(c)
                }}
              >
                <InitialsAvatar
                  name={`${c.first_name} ${c.last_name}`}
                  size="sm"
                />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-slate-800">
                    {c.first_name} {c.last_name}
                  </p>
                  {c.account && (
                    <p className="truncate text-xs text-slate-500">{c.account.name}</p>
                  )}
                </div>
              </button>
            ))
          )}
          {results.length > 0 && (
            <div className="border-t border-slate-100 px-3 py-2">
              <Link
                to={`/contacts?search=${encodeURIComponent(query)}`}
                className="text-xs text-indigo-600 hover:underline"
                onMouseDown={(e) => e.preventDefault()}
              >
                See all in Contacts →
              </Link>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ── Main ContactSection ──────────────────────────────────────────────────────

interface ContactSectionProps {
  ticketId: string
  contact: Contact | undefined
  ticketStatus: TicketStatus
}

export function ContactSection({ ticketId, contact, ticketStatus }: ContactSectionProps) {
  const [mode, setMode] = useState<'view' | 'search'>('view')
  const [optimisticContact, setOptimisticContact] = useState<Contact | null | undefined>(undefined)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const errorTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const linkBtnRef = useRef<HTMLButtonElement>(null)
  const changeBtnRef = useRef<HTMLButtonElement>(null)
  const liveRef = useRef<HTMLDivElement>(null)

  const { mutateAsync: updateContact } = useUpdateTicketContact()

  const isReadOnly = ticketStatus === 'resolved' || ticketStatus === 'closed'
  const displayed = optimisticContact !== undefined ? optimisticContact : contact

  const showError = useCallback((msg: string) => {
    setError(msg)
    if (errorTimerRef.current) clearTimeout(errorTimerRef.current)
    errorTimerRef.current = setTimeout(() => setError(null), 4000)
  }, [])

  const handleSelect = useCallback(
    async (selected: Contact) => {
      const previous = displayed ?? null
      setOptimisticContact(selected)
      setMode('view')
      setSaving(true)
      try {
        await updateContact({ ticketId, contactId: selected.id })
        if (liveRef.current) {
          liveRef.current.textContent = `${selected.first_name} ${selected.last_name} linked to ticket`
        }
      } catch {
        setOptimisticContact(previous ?? undefined)
        showError('Failed to link contact. Try again.')
      } finally {
        setSaving(false)
        setTimeout(() => changeBtnRef.current?.focus(), 0)
      }
    },
    [displayed, ticketId, updateContact, showError]
  )

  const handleUnlink = useCallback(async () => {
    const previous = displayed ?? null
    setOptimisticContact(null)
    setSaving(true)
    try {
      await updateContact({ ticketId, contactId: null })
    } catch {
      setOptimisticContact(previous ?? undefined)
      showError('Failed to unlink contact. Try again.')
    } finally {
      setSaving(false)
      setTimeout(() => linkBtnRef.current?.focus(), 0)
    }
  }, [displayed, ticketId, updateContact, showError])

  const handleCancel = useCallback(() => {
    setMode('view')
    setTimeout(() => {
      if (displayed) {
        changeBtnRef.current?.focus()
      } else {
        linkBtnRef.current?.focus()
      }
    }, 0)
  }, [displayed])

  return (
    <section aria-label="Contact">
      {/* aria-live region for success announcements */}
      <div ref={liveRef} aria-live="polite" className="sr-only" />

      <div className="flex items-center gap-1.5 mb-3">
        <User className="h-3.5 w-3.5 text-slate-400" />
        <span className="text-[11px] font-medium uppercase tracking-wide text-slate-500">
          Contact
        </span>
      </div>

      {mode === 'search' ? (
        <ContactSearch onSelect={handleSelect} onCancel={handleCancel} />
      ) : displayed ? (
        <div
          className={cn(
            'relative rounded-lg border border-slate-200 p-3 transition-opacity',
            saving && 'opacity-60'
          )}
        >
          {saving && (
            <div className="absolute inset-0 flex items-center justify-center">
              <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
            </div>
          )}

          <div className="flex items-start gap-2.5">
            <InitialsAvatar name={`${displayed.first_name} ${displayed.last_name}`} />
            <div className="min-w-0 flex-1">
              <p
                className="truncate font-semibold text-sm text-slate-800"
                title={`${displayed.first_name} ${displayed.last_name}`}
              >
                {displayed.first_name} {displayed.last_name}
              </p>
              {displayed.email && (
                <div className="flex items-center gap-1.5 mt-0.5">
                  <Mail className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                  <span className="truncate text-xs text-slate-500">{displayed.email}</span>
                </div>
              )}
              {displayed.phone && (
                <div className="flex items-center gap-1.5 mt-0.5">
                  <Phone className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                  <span className="text-xs text-slate-500">{displayed.phone}</span>
                </div>
              )}
              {displayed.account && (
                <div className="flex items-center gap-1.5 mt-0.5">
                  <Building2 className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                  <span className="text-xs text-slate-500">{displayed.account.name}</span>
                </div>
              )}
            </div>

            {!isReadOnly && (
              <div className="flex items-center gap-1 shrink-0">
                <Button
                  ref={changeBtnRef}
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setMode('search')}
                  aria-label="Change linked contact"
                  disabled={saving}
                >
                  <ArrowRightLeft className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleUnlink}
                  aria-label={`Unlink contact ${displayed.first_name} ${displayed.last_name}`}
                  disabled={saving}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2.5">
          <span className="text-sm text-slate-400">No contact linked.</span>
          {!isReadOnly && (
            <Button
              ref={linkBtnRef}
              variant="ghost"
              size="sm"
              onClick={() => setMode('search')}
              aria-label="Link a contact to this ticket"
            >
              + Link contact
            </Button>
          )}
        </div>
      )}

      {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
    </section>
  )
}
