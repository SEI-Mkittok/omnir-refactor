import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ChevronRight,
  Building2,
  Globe,
  Phone,
  MapPin,
  Users,
  Briefcase,
  Trash2,
  ExternalLink,
  Plus,
} from 'lucide-react'
import {
  useAccount,
  useAccountContacts,
  useAccountDeals,
  useAccountNotes,
  useDeleteAccount,
  useAddAccountNote,
} from '@/hooks/useAccounts'

import { Button } from '@/components/ui/Button'
import { formatDate, formatRelativeTime, formatCurrency } from '@/lib/utils'
import type { Contact, Deal, Note } from '@/api/types'

// ── Helpers ───────────────────────────────────────────────────────────────────

function accountInitials(name: string) {
  return name
    .split(' ')
    .map((w) => w[0] ?? '')
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

// ── Contacts List ─────────────────────────────────────────────────────────────

function ContactsList({ accountId }: { accountId: string }) {
  const { data: contacts, isLoading } = useAccountContacts(accountId)

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <div className="flex items-center gap-2 mb-3">
        <Users className="h-4 w-4" style={{ color: 'var(--text-label)' }} />
        <p
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          Contacts
        </p>
        {contacts && (
          <span
            className="ml-auto text-xs font-medium rounded-full px-2 py-0.5"
            style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
          >
            {contacts.length}
          </span>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-10 bg-slate-100 animate-pulse rounded-lg" />
          ))}
        </div>
      ) : !contacts?.length ? (
        <p className="text-sm" style={{ color: 'var(--text-label)' }}>
          No contacts linked.
        </p>
      ) : (
        <ul className="space-y-1">
          {contacts.map((c: Contact) => (
            <li key={c.id}>
              <Link
                to={`/contacts/${c.id}`}
                className="flex items-center gap-2.5 rounded-lg px-2 py-2 hover:bg-[var(--surface-app)] transition-colors"
              >
                <div
                  className="h-7 w-7 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0"
                  style={{
                    background: 'var(--color-primary-light)',
                    color: 'var(--color-primary)',
                  }}
                >
                  {((c.first_name?.[0] ?? '') + (c.last_name?.[0] ?? '')).toUpperCase() || '?'}
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                    {c.first_name} {c.last_name}
                  </p>
                  {c.title && (
                    <p className="text-xs truncate" style={{ color: 'var(--text-label)' }}>
                      {c.title}
                    </p>
                  )}
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

// ── Deals List ────────────────────────────────────────────────────────────────

const STAGE_STYLES: Record<string, { bg: string; text: string }> = {
  prospecting:    { bg: '#EFF6FF', text: '#2563EB' },
  qualification:  { bg: '#F0FDF4', text: '#15803D' },
  proposal:       { bg: '#FAF5FF', text: '#7C3AED' },
  negotiation:    { bg: '#FFF7ED', text: '#C2410C' },
  closed_won:     { bg: '#F0FDF4', text: '#15803D' },
  closed_lost:    { bg: '#FEF2F2', text: '#DC2626' },
}

function DealsList({ accountId }: { accountId: string }) {
  const { data: deals, isLoading } = useAccountDeals(accountId)

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <div className="flex items-center gap-2 mb-3">
        <Briefcase className="h-4 w-4" style={{ color: 'var(--text-label)' }} />
        <p
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          Deals
        </p>
        {deals && (
          <span
            className="ml-auto text-xs font-medium rounded-full px-2 py-0.5"
            style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
          >
            {deals.length}
          </span>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="h-12 bg-slate-100 animate-pulse rounded-lg" />
          ))}
        </div>
      ) : !deals?.length ? (
        <p className="text-sm" style={{ color: 'var(--text-label)' }}>
          No deals linked.
        </p>
      ) : (
        <ul className="space-y-2">
          {deals.map((d: Deal) => {
            const stageStyle = STAGE_STYLES[d.stage] ?? { bg: 'var(--surface-app)', text: 'var(--text-label)' }
            return (
              <li
                key={d.id}
                className="flex items-center justify-between gap-2 rounded-lg px-2 py-2"
                style={{ background: 'var(--surface-app)' }}
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                    {d.title}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>
                    {formatCurrency((d.value_cents ?? 0) / 100)}
                  </p>
                </div>
                <span
                  className="shrink-0 text-[11px] font-semibold uppercase px-2 py-0.5 rounded-full"
                  style={{
                    background: stageStyle.bg,
                    color: stageStyle.text,
                    letterSpacing: '0.04em',
                  }}
                >
                  {d.stage.replace('_', ' ')}
                </span>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

// ── Notes Panel ───────────────────────────────────────────────────────────────

function NotesPanel({ accountId }: { accountId: string }) {
  const { data: notes, isLoading } = useAccountNotes(accountId)
  const addNote = useAddAccountNote()
  const [noteText, setNoteText] = useState('')

  const handleSubmit = async () => {
    const content = noteText.trim()
    if (!content) return
    await addNote.mutateAsync({ accountId, payload: { content } })
    setNoteText('')
  }

  return (
    <div
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
    >
      <p
        className="mb-3 text-[11px] font-semibold uppercase tracking-widest"
        style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
      >
        Notes
      </p>

      {/* Add note */}
      <div className="mb-4">
        <textarea
          value={noteText}
          onChange={(e) => setNoteText(e.target.value)}
          placeholder="Add a note…"
          rows={3}
          className="w-full rounded-lg border text-sm resize-none px-3 py-2 focus:outline-none"
          style={{
            borderColor: 'var(--border-default)',
            color: 'var(--text-primary)',
            background: 'var(--surface-app)',
          }}
          onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--border-focus)' }}
          onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)' }}
        />
        <div className="flex justify-end mt-1.5">
          <Button
            size="sm"
            disabled={!noteText.trim() || addNote.isPending}
            onClick={handleSubmit}
          >
            <Plus className="h-3.5 w-3.5" />
            Add Note
          </Button>
        </div>
      </div>

      {/* Note list */}
      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="h-10 bg-slate-100 animate-pulse rounded" />
          ))}
        </div>
      ) : !notes?.length ? (
        <p className="text-sm" style={{ color: 'var(--text-label)' }}>
          No notes yet.
        </p>
      ) : (
        <ul className="space-y-3">
          {notes.map((n: Note) => (
            <li
              key={n.id}
              className="border-b pb-3 last:border-0 last:pb-0"
              style={{ borderColor: 'var(--border-subtle)' }}
            >
              <p className="text-sm" style={{ color: 'var(--text-primary)', whiteSpace: 'pre-wrap' }}>
                {n.content}
              </p>
              <time
                className="mt-1 block text-xs"
                style={{ color: 'var(--text-label)' }}
                title={formatDate(n.created_at)}
              >
                {formatRelativeTime(n.created_at)}
              </time>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function AccountDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const { data: account, isLoading, isError } = useAccount(id!)
  const deleteAccount = useDeleteAccount()

  if (isLoading) {
    return (
      <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
        <div className="h-4 w-48 bg-slate-100 animate-pulse rounded mb-5" />
        <div
          className="rounded-xl border p-5 flex items-center gap-4 mb-5"
          style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        >
          <div className="h-[72px] w-[72px] rounded-xl bg-slate-100 animate-pulse shrink-0" />
          <div className="space-y-2 flex-1">
            <div className="h-6 w-56 bg-slate-100 animate-pulse rounded" />
            <div className="h-3.5 w-36 bg-slate-100 animate-pulse rounded" />
          </div>
        </div>
        <div className="flex gap-5">
          <div className="flex-1 space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-32 bg-slate-100 animate-pulse rounded-xl" />
            ))}
          </div>
          <div className="w-[300px] shrink-0 space-y-3">
            <div className="h-48 bg-slate-100 animate-pulse rounded-xl" />
            <div className="h-32 bg-slate-100 animate-pulse rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !account) {
    return (
      <div className="p-6 flex flex-col items-center justify-center py-24 text-center">
        <p className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
          Account not found.
        </p>
        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          This account may have been deleted or you may not have access.
        </p>
        <Link to="/accounts" className="text-sm font-medium" style={{ color: 'var(--color-primary)' }}>
          ← Back to Accounts
        </Link>
      </div>
    )
  }

  const handleDelete = async () => {
    await deleteAccount.mutateAsync(account.id)
    navigate('/accounts')
  }

  return (
    <div className="p-6" style={{ maxWidth: 'var(--content-max-width, 1280px)' }}>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 mb-5 text-sm" style={{ color: 'var(--text-label)' }}>
        <Link to="/accounts" className="hover:underline" style={{ color: 'var(--color-primary)' }}>
          Accounts
        </Link>
        <ChevronRight className="h-3.5 w-3.5" />
        <span style={{ color: 'var(--text-primary)' }}>{account.name}</span>
      </nav>

      {/* Header Card */}
      <div
        className="rounded-xl border p-5 mb-5 flex flex-wrap items-center gap-4"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      >
        {/* Avatar */}
        <div
          className="h-[72px] w-[72px] rounded-xl flex items-center justify-center shrink-0 text-2xl font-bold"
          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
        >
          {accountInitials(account.name)}
        </div>

        {/* Identity */}
        <div className="flex-1 min-w-0">
          <h1
            className="font-bold leading-tight"
            style={{ fontSize: 24, color: 'var(--text-primary)', letterSpacing: '-0.01em' }}
          >
            {account.name}
          </h1>
          {account.industry && (
            <p className="mt-0.5 text-sm" style={{ color: 'var(--text-secondary)' }}>
              {account.industry.charAt(0).toUpperCase() + account.industry.slice(1)}
              {account.size ? ` · ${account.size} employees` : ''}
            </p>
          )}
          {account.domain && (
            <a
              href={`https://${account.domain}`}
              target="_blank"
              rel="noreferrer"
              className="mt-0.5 inline-flex items-center gap-1 text-sm font-medium hover:underline"
              style={{ color: 'var(--color-primary)' }}
            >
              {account.domain}
              <ExternalLink className="h-3 w-3" />
            </a>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowDeleteConfirm(true)}
            className="text-red-600 border-red-200 hover:bg-red-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Delete
          </Button>
        </div>
      </div>

      {/* Body */}
      <div className="flex flex-col lg:flex-row gap-5">
        {/* Left column — details + notes */}
        <div className="flex-1 min-w-0 space-y-5">
          {/* Details card */}
          <div
            className="rounded-xl border p-5"
            style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
          >
            <p
              className="mb-4 text-[11px] font-semibold uppercase tracking-widest"
              style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
            >
              Details
            </p>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
              {account.phone && (
                <div>
                  <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                    <Phone className="h-3.5 w-3.5" /> Phone
                  </dt>
                  <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {account.phone}
                  </dd>
                </div>
              )}
              {account.website && (
                <div>
                  <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                    <Globe className="h-3.5 w-3.5" /> Website
                  </dt>
                  <dd>
                    <a
                      href={account.website}
                      target="_blank"
                      rel="noreferrer"
                      className="text-sm font-medium hover:underline"
                      style={{ color: 'var(--color-primary)' }}
                    >
                      {account.website}
                    </a>
                  </dd>
                </div>
              )}
              {account.address && (
                <div className="sm:col-span-2">
                  <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                    <MapPin className="h-3.5 w-3.5" /> Address
                  </dt>
                  <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                    {account.address}
                  </dd>
                </div>
              )}
              <div>
                <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                  <Building2 className="h-3.5 w-3.5" /> Created
                </dt>
                <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {formatDate(account.created_at)}
                </dd>
              </div>
              <div>
                <dt className="flex items-center gap-1.5 text-xs mb-0.5" style={{ color: 'var(--text-label)' }}>
                  <Building2 className="h-3.5 w-3.5" /> Last Updated
                </dt>
                <dd className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
                  {formatRelativeTime(account.updated_at)}
                </dd>
              </div>
            </dl>
          </div>

          {/* Notes */}
          <NotesPanel accountId={account.id} />
        </div>

        {/* Right column — contacts + deals */}
        <div className="w-full lg:w-[300px] shrink-0 space-y-5">
          <ContactsList accountId={account.id} />
          <DealsList accountId={account.id} />
        </div>
      </div>

      {/* Delete confirm */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div
            className="rounded-xl border shadow-xl p-6 w-full max-w-sm"
            style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
          >
            <h2 className="text-base font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
              Delete Account
            </h2>
            <p className="text-sm mb-5" style={{ color: 'var(--text-secondary)' }}>
              Are you sure you want to delete <strong>{account.name}</strong>? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowDeleteConfirm(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                disabled={deleteAccount.isPending}
                onClick={handleDelete}
                className="bg-red-600 hover:bg-red-700 text-white border-red-600"
              >
                {deleteAccount.isPending ? 'Deleting…' : 'Delete'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
