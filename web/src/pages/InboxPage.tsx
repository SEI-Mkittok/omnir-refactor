import { useState, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import {
  Mail, Plus, Search, RefreshCw, Archive, MoreHorizontal,
  ChevronDown, Send, X, Paperclip as PaperclipIcon,
  Bold, Italic, Underline, User, ExternalLink, Inbox,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useEmailAccounts,
  useInboxThreads,
  useInboxThread,
  useEmailTemplates,
  useSendEmail,
  useMarkThreadRead,
  useConnectEmailAccount,
} from '@/hooks/useInbox'
import { useAccountContacts } from '@/hooks/useAccounts'
import { inboxApi } from '@/api/inbox'
import type { Contact, InboxThread, InboxMessage, EmailAccount, EmailTemplate } from '@/api/types'

// ── Helpers ──────────────────────────────────────────────────────────────────

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days}d ago`
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

function initials(name?: string, addr?: string): string {
  const src = name || addr || '?'
  return src.split(/[\s@]/).filter(Boolean).map((w) => w[0]).join('').slice(0, 2).toUpperCase()
}

function senderLabel(msg: InboxMessage): string {
  return msg.from_name || msg.from_addr
}

// ── Empty state — no accounts connected ──────────────────────────────────────

function ConnectAccountsEmptyState({ onConnect }: { onConnect: (provider: 'gmail' | 'outlook') => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-8 px-6 py-16">
      <div className="text-center">
        <h1 className="text-[36px] font-bold text-[#1A1D23]">Email.</h1>
        <p className="mt-1 text-[14px] text-[#6B7280]">Connect an account to get started.</p>
      </div>

      <div className="flex w-full max-w-xl gap-4">
        {/* Gmail card */}
        <button
          onClick={() => onConnect('gmail')}
          className="flex flex-1 flex-col items-center gap-3 rounded-lg border border-[#E5E7EB] bg-white px-6 py-8 text-left transition-shadow hover:shadow-md active:scale-[0.99]"
        >
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-red-50">
            <svg viewBox="0 0 24 24" className="h-7 w-7" aria-hidden="true">
              <path d="M22 6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6z" fill="#fff" stroke="#E5E7EB"/>
              <path d="M22 6l-10 7L2 6" fill="none" stroke="#EA4335" strokeWidth="2"/>
            </svg>
          </div>
          <div>
            <p className="font-semibold text-[#1A1D23]">Connect Gmail</p>
            <p className="mt-0.5 text-sm text-[#6B7280]">Sync your Google Workspace inbox</p>
          </div>
          <span className="mt-auto flex items-center gap-1 text-sm font-medium text-[#1B3A4B]">
            Connect <ChevronDown className="h-4 w-4 -rotate-90" />
          </span>
        </button>

        {/* Outlook card */}
        <button
          onClick={() => onConnect('outlook')}
          className="flex flex-1 flex-col items-center gap-3 rounded-lg border border-[#E5E7EB] bg-white px-6 py-8 text-left transition-shadow hover:shadow-md active:scale-[0.99]"
        >
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-50">
            <svg viewBox="0 0 24 24" className="h-7 w-7" aria-hidden="true">
              <rect x="2" y="4" width="20" height="16" rx="2" fill="#0078D4"/>
              <path d="M2 8l10 7 10-7" stroke="#fff" strokeWidth="1.5" fill="none"/>
            </svg>
          </div>
          <div>
            <p className="font-semibold text-[#1A1D23]">Connect Outlook</p>
            <p className="mt-0.5 text-sm text-[#6B7280]">Sync your Microsoft 365 inbox</p>
          </div>
          <span className="mt-auto flex items-center gap-1 text-sm font-medium text-[#1B3A4B]">
            Connect <ChevronDown className="h-4 w-4 -rotate-90" />
          </span>
        </button>
      </div>

      <p className="max-w-sm text-center text-[14px] text-[#6B7280]">
        Your emails stay private — PraestOS only reads messages linked to your CRM contacts.
      </p>
    </div>
  )
}

// ── Thread list row ───────────────────────────────────────────────────────────

function ThreadRow({
  thread,
  account,
  isSelected,
  onClick,
}: {
  thread: InboxThread
  account?: EmailAccount
  isSelected: boolean
  onClick: () => void
}) {
  const sender = thread.participants.find((p) => !account || !p.includes(account.email_address.split('@')[1] ?? '')) ?? thread.participants[0]

  return (
    <button
      role="option"
      aria-selected={isSelected}
      onClick={onClick}
      className={cn(
        'flex w-full items-start gap-3 px-4 py-3 text-left transition-colors border-b border-[#F3F4F6] hover:bg-[#F7F8FA]',
        isSelected && 'bg-[#E8EDF2]',
        thread.unread && 'bg-white'
      )}
      style={{ minHeight: 72 }}
    >
      {/* Unread dot */}
      <div className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center">
        {thread.unread && (
          <span className="block h-2.5 w-2.5 rounded-full bg-[#3B82F6]" aria-label="Unread" />
        )}
      </div>

      {/* Avatar */}
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#E8EDF2] text-[11px] font-bold text-[#1B3A4B]">
        {initials(undefined, sender)}
      </div>

      {/* Content */}
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <span className={cn('truncate text-sm', thread.unread ? 'font-semibold text-[#1A1D23]' : 'text-[#374151]')}>
            {sender}
          </span>
          <span className="shrink-0 text-[11px] text-[#9CA3AF]">{relativeTime(thread.last_message_at)}</span>
        </div>
        <p className={cn('truncate text-[13px]', thread.unread ? 'font-medium text-[#374151]' : 'text-[#6B7280]')}>
          {thread.subject}
        </p>
        <p className="truncate text-[12px] text-[#9CA3AF]">{thread.snippet}</p>
        {account && (
          <span className="mt-0.5 inline-block rounded-full bg-[#F3F4F6] px-2 py-0.5 text-[10px] font-medium text-[#6B7280]">
            {account.email_address}
          </span>
        )}
      </div>
    </button>
  )
}

// ── Message bubble ────────────────────────────────────────────────────────────

function MessageBubble({
  message,
  onReply,
}: {
  message: InboxMessage
  onReply: () => void
}) {
  const [collapsed, setCollapsed] = useState(false)
  const isOutbound = message.direction === 'outbound'

  return (
    <article
      aria-label={`Message from ${senderLabel(message)} at ${relativeTime(message.sent_at)}`}
      className={cn(
        'rounded-lg border border-[#E5E7EB] p-4',
        isOutbound ? 'bg-[#F0F4F8] ml-8' : 'bg-white'
      )}
    >
      {/* Header */}
      <button
        className="flex w-full items-center justify-between gap-2 text-left"
        onClick={() => setCollapsed((v) => !v)}
        aria-expanded={!collapsed}
      >
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-[#E8EDF2] text-[10px] font-bold text-[#1B3A4B]">
            {initials(message.from_name, message.from_addr)}
          </div>
          <div>
            <span className="text-sm font-semibold text-[#1A1D23]">{senderLabel(message)}</span>
            <span className="ml-2 text-[11px] text-[#9CA3AF]">{relativeTime(message.sent_at)}</span>
          </div>
        </div>
        <ChevronDown className={cn('h-4 w-4 text-[#9CA3AF] transition-transform', collapsed && 'rotate-180')} />
      </button>

      {!collapsed && (
        <>
          <p className="mt-1 text-[11px] text-[#9CA3AF]">To: {message.to_addrs.join(', ')}</p>
          <div className="mt-3 text-sm text-[#374151] whitespace-pre-wrap">{message.body_text}</div>
          <div className="mt-3 flex gap-3">
            <button
              onClick={onReply}
              className="text-[12px] font-medium text-[#1B3A4B] hover:underline"
            >
              Reply
            </button>
            <button className="text-[12px] font-medium text-[#6B7280] hover:underline">
              Forward
            </button>
          </div>
        </>
      )}
    </article>
  )
}

// ── Reply composer ────────────────────────────────────────────────────────────

interface ReplyComposerProps {
  thread: InboxThread
  accounts: EmailAccount[]
  onSend: (body: string, fromAccountId: string, cc?: string, bcc?: string) => Promise<void>
  onDiscard: () => void
}

function ReplyComposer({ thread, accounts, onSend, onDiscard }: ReplyComposerProps) {
  const [body, setBody] = useState('')
  const [fromAccountId, setFromAccountId] = useState(accounts[0]?.id ?? '')
  const [cc, setCc] = useState('')
  const [bcc, setBcc] = useState('')
  const [showCc, setShowCc] = useState(false)
  const [showBcc, setShowBcc] = useState(false)
  const [sending, setSending] = useState(false)
  const toAddr = thread.participants.find((p) => {
    const acc = accounts.find((a) => a.id === fromAccountId)
    return !acc || p !== acc.email_address
  }) ?? thread.participants[0]

  async function handleSend() {
    if (!body.trim()) return
    setSending(true)
    try {
      await onSend(body, fromAccountId, showCc ? cc : undefined, showBcc ? bcc : undefined)
      setBody('')
    } finally {
      setSending(false)
    }
  }

  return (
    <div
      aria-label="Reply composer"
      className="rounded-lg border border-[#E5E7EB] bg-white shadow-sm"
    >
      {/* To row */}
      <div className="flex items-center gap-2 border-b border-[#F3F4F6] px-4 py-2 text-sm">
        <span className="text-[#9CA3AF]">To:</span>
        <span className="font-medium text-[#374151]">{toAddr}</span>
        <div className="ml-auto flex gap-2">
          {!showCc && (
            <button onClick={() => setShowCc(true)} className="text-[12px] text-[#6B7280] hover:text-[#1B3A4B]">
              + CC
            </button>
          )}
          {!showBcc && (
            <button onClick={() => setShowBcc(true)} className="text-[12px] text-[#6B7280] hover:text-[#1B3A4B]">
              + BCC
            </button>
          )}
        </div>
      </div>

      {/* CC row */}
      {showCc && (
        <div className="flex items-center gap-2 border-b border-[#F3F4F6] px-4 py-2">
          <span className="text-[12px] text-[#9CA3AF]">CC:</span>
          <input
            className="flex-1 text-sm outline-none placeholder:text-[#9CA3AF]"
            placeholder="cc@example.com"
            value={cc}
            onChange={(e) => setCc(e.target.value)}
            aria-label="CC recipients"
          />
        </div>
      )}

      {/* BCC row */}
      {showBcc && (
        <div className="flex items-center gap-2 border-b border-[#F3F4F6] px-4 py-2">
          <span className="text-[12px] text-[#9CA3AF]">BCC:</span>
          <input
            className="flex-1 text-sm outline-none placeholder:text-[#9CA3AF]"
            placeholder="bcc@example.com"
            value={bcc}
            onChange={(e) => setBcc(e.target.value)}
            aria-label="BCC recipients"
          />
        </div>
      )}

      {/* From selector */}
      {accounts.length > 1 && (
        <div className="flex items-center gap-2 border-b border-[#F3F4F6] px-4 py-2">
          <span className="text-[12px] text-[#9CA3AF]">From:</span>
          <select
            aria-label="Send from account"
            value={fromAccountId}
            onChange={(e) => setFromAccountId(e.target.value)}
            className="flex-1 text-sm outline-none bg-transparent"
          >
            {accounts.map((a) => (
              <option key={a.id} value={a.id}>{a.email_address}</option>
            ))}
          </select>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex items-center gap-2 border-b border-[#F3F4F6] px-4 py-2">
        <button aria-label="Bold" className="rounded p-1 hover:bg-[#F3F4F6]"><Bold className="h-3.5 w-3.5 text-[#6B7280]" /></button>
        <button aria-label="Italic" className="rounded p-1 hover:bg-[#F3F4F6]"><Italic className="h-3.5 w-3.5 text-[#6B7280]" /></button>
        <button aria-label="Underline" className="rounded p-1 hover:bg-[#F3F4F6]"><Underline className="h-3.5 w-3.5 text-[#6B7280]" /></button>
        <div className="h-4 w-px bg-[#E5E7EB]" />
        <button aria-label="Insert template" className="flex items-center gap-1 rounded px-2 py-1 text-[11px] font-medium text-[#6B7280] hover:bg-[#F3F4F6]">
          <Mail className="h-3.5 w-3.5" /> Templates
        </button>
      </div>

      {/* Body */}
      <textarea
        aria-label="Reply body"
        className="w-full resize-none px-4 py-3 text-sm outline-none placeholder:text-[#9CA3AF]"
        rows={4}
        placeholder="Write a reply…"
        value={body}
        onChange={(e) => setBody(e.target.value)}
        disabled={sending}
      />

      {/* Actions */}
      <div className="flex items-center gap-3 border-t border-[#F3F4F6] px-4 py-3">
        <button
          onClick={handleSend}
          disabled={!body.trim() || sending}
          className="flex items-center gap-2 rounded-md bg-[#1B3A4B] px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-[#152E3C] disabled:opacity-50"
        >
          {sending ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          {sending ? 'Sending…' : 'Send'}
        </button>
        <button
          onClick={onDiscard}
          disabled={sending}
          className="text-sm text-[#6B7280] hover:text-[#374151]"
        >
          Discard
        </button>
      </div>
    </div>
  )
}

// ── Template picker popover ───────────────────────────────────────────────────

function TemplatePickerPopover({
  onSelect,
  onClose,
}: {
  onSelect: (t: EmailTemplate) => void
  onClose: () => void
}) {
  const [q, setQ] = useState('')
  const { data: templates = [] } = useEmailTemplates(q || undefined)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose()
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [onClose])

  return (
    <div
      ref={ref}
      role="listbox"
      aria-label="Email templates"
      className="absolute bottom-full left-0 mb-2 w-80 rounded-lg border border-[#E5E7EB] bg-white shadow-lg z-50"
    >
      <div className="p-3 border-b border-[#F3F4F6]">
        <div className="flex items-center gap-2 rounded-md border border-[#E5E7EB] px-3 py-1.5">
          <Search className="h-3.5 w-3.5 text-[#9CA3AF]" />
          <input
            autoFocus
            className="flex-1 text-sm outline-none"
            placeholder="Search templates…"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
      </div>
      <div className="max-h-[280px] overflow-y-auto py-1">
        {templates.length === 0 && (
          <p className="px-4 py-3 text-sm text-[#9CA3AF]">No templates found.</p>
        )}
        {templates.map((t) => (
          <button
            key={t.id}
            role="option"
            aria-selected={false}
            onClick={() => { onSelect(t); onClose() }}
            className="flex w-full flex-col gap-0.5 px-4 py-2.5 text-left hover:bg-[#F7F8FA]"
          >
            <span className="text-sm font-medium text-[#1A1D23]">{t.name}</span>
            <span className="truncate text-[12px] text-[#6B7280]">
              {t.body_html.replace(/<[^>]+>/g, '').slice(0, 80)}
            </span>
          </button>
        ))}
      </div>
    </div>
  )
}

// ── Compose modal (new email) ─────────────────────────────────────────────────

interface ComposeModalProps {
  accounts: EmailAccount[]
  onSend: (payload: {
    accountId: string
    to: string[]
    cc?: string[]
    bcc?: string[]
    subject: string
    body: string
  }) => Promise<void>
  onClose: () => void
}

function ComposeModal({ accounts, onSend, onClose }: ComposeModalProps) {
  const [fromAccountId, setFromAccountId] = useState(accounts[0]?.id ?? '')
  const [to, setTo] = useState('')
  const [cc, setCc] = useState('')
  const [bcc, setBcc] = useState('')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [showCc, setShowCc] = useState(false)
  const [showBcc, setShowBcc] = useState(false)
  const [sending, setSending] = useState(false)
  const [showTemplates, setShowTemplates] = useState(false)

  async function handleSend() {
    if (!to.trim() || !subject.trim() || !body.trim()) return
    setSending(true)
    try {
      await onSend({
        accountId: fromAccountId,
        to: to.split(',').map((e) => e.trim()).filter(Boolean),
        cc: showCc && cc ? cc.split(',').map((e) => e.trim()).filter(Boolean) : undefined,
        bcc: showBcc && bcc ? bcc.split(',').map((e) => e.trim()).filter(Boolean) : undefined,
        subject,
        body,
      })
      onClose()
    } finally {
      setSending(false)
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') onClose()
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onKeyDown={handleKeyDown}
      role="dialog"
      aria-modal="true"
      aria-label="New Email"
    >
      <div className="w-full max-w-[680px] rounded-xl bg-white shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-[#F3F4F6] px-6 py-4">
          <h2 className="text-base font-semibold text-[#1A1D23]">New Email</h2>
          <button onClick={onClose} aria-label="Close compose" className="rounded p-1 hover:bg-[#F3F4F6]">
            <X className="h-4 w-4 text-[#6B7280]" />
          </button>
        </div>

        <div className="px-6 py-4 space-y-2">
          {/* From */}
          <div className="flex items-center gap-3">
            <span className="w-14 shrink-0 text-[13px] text-[#9CA3AF]">From</span>
            <select
              aria-label="Send from account"
              value={fromAccountId}
              onChange={(e) => setFromAccountId(e.target.value)}
              className="flex-1 rounded-md border border-[#E5E7EB] px-3 py-1.5 text-sm outline-none focus:border-[#1B3A4B]"
            >
              {accounts.map((a) => <option key={a.id} value={a.id}>{a.email_address}</option>)}
            </select>
          </div>

          {/* To */}
          <div className="flex items-center gap-3">
            <span className="w-14 shrink-0 text-[13px] text-[#9CA3AF]">To</span>
            <input
              aria-label="To: recipients"
              className="flex-1 rounded-md border border-[#E5E7EB] px-3 py-1.5 text-sm outline-none focus:border-[#1B3A4B]"
              placeholder="recipient@example.com"
              value={to}
              onChange={(e) => setTo(e.target.value)}
            />
            <div className="flex gap-2 shrink-0">
              {!showCc && <button onClick={() => setShowCc(true)} className="text-[12px] text-[#6B7280] hover:text-[#1B3A4B]">CC</button>}
              {!showBcc && <button onClick={() => setShowBcc(true)} className="text-[12px] text-[#6B7280] hover:text-[#1B3A4B]">BCC</button>}
            </div>
          </div>

          {showCc && (
            <div className="flex items-center gap-3">
              <span className="w-14 shrink-0 text-[13px] text-[#9CA3AF]">CC</span>
              <input
                aria-label="CC recipients"
                className="flex-1 rounded-md border border-[#E5E7EB] px-3 py-1.5 text-sm outline-none focus:border-[#1B3A4B]"
                placeholder="cc@example.com"
                value={cc}
                onChange={(e) => setCc(e.target.value)}
              />
            </div>
          )}

          {showBcc && (
            <div className="flex items-center gap-3">
              <span className="w-14 shrink-0 text-[13px] text-[#9CA3AF]">BCC</span>
              <input
                aria-label="BCC recipients"
                className="flex-1 rounded-md border border-[#E5E7EB] px-3 py-1.5 text-sm outline-none focus:border-[#1B3A4B]"
                placeholder="bcc@example.com"
                value={bcc}
                onChange={(e) => setBcc(e.target.value)}
              />
            </div>
          )}

          {/* Subject */}
          <div className="flex items-center gap-3">
            <span className="w-14 shrink-0 text-[13px] text-[#9CA3AF]">Subject</span>
            <input
              className="flex-1 rounded-md border border-[#E5E7EB] px-3 py-1.5 text-sm outline-none focus:border-[#1B3A4B]"
              placeholder="Subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
            />
          </div>
        </div>

        {/* Divider */}
        <div className="border-t border-[#F3F4F6]" />

        {/* Toolbar */}
        <div className="flex items-center gap-2 px-6 py-2">
          <button aria-label="Bold" className="rounded p-1 hover:bg-[#F3F4F6]"><Bold className="h-3.5 w-3.5 text-[#6B7280]" /></button>
          <button aria-label="Italic" className="rounded p-1 hover:bg-[#F3F4F6]"><Italic className="h-3.5 w-3.5 text-[#6B7280]" /></button>
          <button aria-label="Underline" className="rounded p-1 hover:bg-[#F3F4F6]"><Underline className="h-3.5 w-3.5 text-[#6B7280]" /></button>
          <div className="h-4 w-px bg-[#E5E7EB]" />
          <button aria-label="Attach file" className="rounded p-1 hover:bg-[#F3F4F6]"><PaperclipIcon className="h-3.5 w-3.5 text-[#6B7280]" /></button>
          <div className="relative">
            <button
              aria-label="Insert template"
              onClick={() => setShowTemplates((v) => !v)}
              className="flex items-center gap-1 rounded px-2 py-1 text-[11px] font-medium text-[#6B7280] hover:bg-[#F3F4F6]"
            >
              <Mail className="h-3.5 w-3.5" /> Templates
            </button>
            {showTemplates && (
              <TemplatePickerPopover
                onSelect={(t) => setBody(t.body_html.replace(/<[^>]+>/g, ''))}
                onClose={() => setShowTemplates(false)}
              />
            )}
          </div>
        </div>

        {/* Body */}
        <div className="px-6 pb-4">
          <textarea
            className="w-full resize-none rounded-md border border-[#E5E7EB] px-3 py-2 text-sm outline-none focus:border-[#1B3A4B] placeholder:text-[#9CA3AF]"
            rows={8}
            placeholder="Email body…"
            value={body}
            onChange={(e) => setBody(e.target.value)}
            disabled={sending}
          />
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-[#F3F4F6] px-6 py-4">
          <button
            onClick={handleSend}
            disabled={!to.trim() || !subject.trim() || !body.trim() || sending}
            className="flex items-center gap-2 rounded-md bg-[#1B3A4B] px-5 py-2 text-sm font-semibold text-white transition-colors hover:bg-[#152E3C] disabled:opacity-50"
          >
            {sending ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
            {sending ? 'Sending…' : 'Send'}
          </button>
          <div className="flex gap-3">
            <button className="text-sm text-[#6B7280] hover:text-[#374151]">Save Draft</button>
            <button onClick={onClose} className="text-sm text-[#6B7280] hover:text-[#374151]">Discard</button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Thread detail (right panel) ───────────────────────────────────────────────

function ThreadDetailPanel({
  threadId,
  accounts,
}: {
  threadId: string
  accounts: EmailAccount[]
}) {
  const { data: thread, isLoading } = useInboxThread(threadId)
  const { mutateAsync: sendEmail } = useSendEmail()
  const [showReply, setShowReply] = useState(false)

  const messages = thread?.messages ?? []

  const COLLAPSE_THRESHOLD = 5
  const [showAll, setShowAll] = useState(false)
  const visibleMessages = messages.length > COLLAPSE_THRESHOLD && !showAll
    ? [...messages.slice(0, 2), ...messages.slice(-3)]
    : messages
  const hiddenCount = messages.length - 5

  async function handleSend(body: string, fromAccountId: string, cc?: string, bcc?: string) {
    if (!thread) return
    const fromAccount = accounts.find((a) => a.id === fromAccountId)
    const replyTo = thread.participants.find((p) => p !== fromAccount?.email_address) ?? thread.participants[0]
    await sendEmail({
      connection_id: fromAccountId,
      to: [replyTo],
      cc: cc ? [cc] : undefined,
      bcc: bcc ? [bcc] : undefined,
      subject: thread.subject.startsWith('Re:') ? thread.subject : `Re: ${thread.subject}`,
      body_html: `<p>${body.replace(/\n/g, '<br/>')}</p>`,
      thread_id: thread.thread_id,
    })
    setShowReply(false)
  }

  if (isLoading) {
    return (
      <div className="flex flex-1 flex-col gap-3 p-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-24 animate-pulse rounded-lg bg-[#F3F4F6]" />
        ))}
      </div>
    )
  }

  if (!thread) return null

  return (
    <div className="flex flex-1 overflow-hidden">
      {/* Thread messages + reply */}
      <div className="flex flex-1 flex-col overflow-y-auto">
        {/* Thread header */}
        <div className="flex shrink-0 items-center justify-between border-b border-[#F3F4F6] px-4 py-3">
          <h2 className="truncate text-sm font-semibold text-[#1A1D23]">{thread.subject}</h2>
          <div className="flex shrink-0 items-center gap-1">
            <button aria-label="Archive" className="rounded p-1.5 hover:bg-[#F3F4F6]">
              <Archive className="h-4 w-4 text-[#6B7280]" />
            </button>
            <button aria-label="More actions" className="rounded p-1.5 hover:bg-[#F3F4F6]">
              <MoreHorizontal className="h-4 w-4 text-[#6B7280]" />
            </button>
          </div>
        </div>

        <div className="flex-1 space-y-3 overflow-y-auto p-4">
          {visibleMessages.map((msg, i) => {
            const isCollapsePivot = messages.length > COLLAPSE_THRESHOLD && !showAll && i === 1
            return (
              <div key={msg.id}>
                <MessageBubble message={msg} onReply={() => setShowReply(true)} />
                {isCollapsePivot && hiddenCount > 0 && (
                  <button
                    onClick={() => setShowAll(true)}
                    className="my-2 w-full text-center text-[12px] font-medium text-[#1B3A4B] hover:underline"
                    aria-live="polite"
                  >
                    Show {hiddenCount} more message{hiddenCount > 1 ? 's' : ''}
                  </button>
                )}
              </div>
            )
          })}

          {/* Reply composer */}
          {showReply && (
            <ReplyComposer
              thread={thread}
              accounts={accounts}
              onSend={handleSend}
              onDiscard={() => setShowReply(false)}
            />
          )}
          {!showReply && (
            <button
              onClick={() => setShowReply(true)}
              className="w-full rounded-lg border border-[#E5E7EB] px-4 py-3 text-left text-sm text-[#9CA3AF] transition-colors hover:bg-[#F7F8FA]"
            >
              Reply to {senderLabel(messages[messages.length - 1] ?? { from_addr: '', from_name: undefined, direction: 'inbound', id: '', org_id: '', connection_id: '', thread_id: '', message_id: '', to_addrs: [], subject: '', body_text: '', snippet: '', has_attachments: false, sent_at: '', created_at: '' })}…
            </button>
          )}
        </div>
      </div>

      {/* Contact panel (260px) */}
      {thread.contact_id && (
        <div
          className="hidden w-[260px] shrink-0 border-l border-[#F3F4F6] overflow-y-auto xl:flex flex-col"
          aria-label="Contact panel"
        >
          <div className="p-4 space-y-4">
            {/* Avatar + name */}
            <div className="flex flex-col items-center gap-2 pt-4 text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[#E8EDF2] text-base font-bold text-[#1B3A4B]">
                <User className="h-7 w-7" />
              </div>
              <div>
                <p className="font-semibold text-[#1A1D23]">
                  {messages.find((m) => m.direction === 'inbound')?.from_name ?? 'Contact'}
                </p>
                <p className="text-[12px] text-[#6B7280]">
                  {messages.find((m) => m.direction === 'inbound')?.from_addr}
                </p>
              </div>
            </div>

            <button className="flex w-full items-center justify-center gap-1.5 rounded-md border border-[#E5E7EB] px-3 py-2 text-sm font-medium text-[#1B3A4B] transition-colors hover:bg-[#F7F8FA]">
              <ExternalLink className="h-3.5 w-3.5" /> View Contact
            </button>

            <div className="rounded-lg border border-[#F3F4F6] bg-[#F7F8FA] p-3 space-y-2">
              <div className="flex justify-between text-[12px]">
                <span className="text-[#9CA3AF]">Open Deals</span>
                <span className="font-medium text-[#374151]">2</span>
              </div>
              <div className="flex justify-between text-[12px]">
                <span className="text-[#9CA3AF]">Last activity</span>
                <span className="font-medium text-[#374151]">3 days ago</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* No contact found */}
      {!thread.contact_id && (
        <div className="hidden w-[260px] shrink-0 border-l border-[#F3F4F6] xl:flex flex-col items-center justify-start p-4 pt-8 gap-3">
          <User className="h-10 w-10 text-[#D1D5DB]" />
          <p className="text-center text-[12px] text-[#9CA3AF]">
            No CRM contact found for this sender.
          </p>
          <button className="flex items-center gap-1 rounded-md border border-[#E5E7EB] px-3 py-2 text-[12px] font-medium text-[#1B3A4B] hover:bg-[#F7F8FA]">
            <Plus className="h-3.5 w-3.5" /> Create Contact
          </button>
        </div>
      )}
    </div>
  )
}

// ── Main page ────────────────────────────────────────────────────────────────

export function InboxPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const accountId = searchParams.get('account_id') ?? ''
  const accountName = searchParams.get('account_name') ?? ''
  const contactId = searchParams.get('contact_id') ?? ''
  const contactName = searchParams.get('contact_name') ?? ''
  const { data: accounts = [], isLoading: loadingAccounts } = useEmailAccounts()
  const contactsQuery = useAccountContacts(accountId)
  const [filterAccountId, setFilterAccountId] = useState<string | null>(null)
  const [unreadOnly, setUnreadOnly] = useState(false)
  const [selectedThreadId, setSelectedThreadId] = useState<string | null>(null)
  const [showCompose, setShowCompose] = useState(false)

  const { data: threadsPage, isLoading: loadingThreadsBase } = useInboxThreads({
    connection_id: filterAccountId ?? undefined,
    contact_id: contactId || undefined,
    unread_only: unreadOnly || undefined,
  })

  const accountThreadsQuery = useQuery({
    queryKey: ['inbox', 'account-threads', accountId, contactsQuery.data, filterAccountId, unreadOnly],
    enabled: !!accountId && !contactId && !contactsQuery.isLoading,
    staleTime: 30_000,
    queryFn: async () => {
      const contactIds = (contactsQuery.data ?? []).map((contact: Contact) => contact.id)
      if (contactIds.length === 0) return [] as InboxThread[]

      const byThreadId = new Map<string, InboxThread>()
      await Promise.all(
        contactIds.map(async (contactId: string) => {
          const response = await inboxApi.listThreads({
            contact_id: contactId,
            connection_id: filterAccountId ?? undefined,
            unread_only: unreadOnly || undefined,
            page: 1,
            limit: 50,
          })
          for (const thread of response.data ?? []) {
            const existing = byThreadId.get(thread.thread_id)
            if (!existing || new Date(thread.last_message_at) > new Date(existing.last_message_at)) {
              byThreadId.set(thread.thread_id, thread)
            }
          }
        })
      )

      return Array.from(byThreadId.values()).sort(
        (a, b) => new Date(b.last_message_at).getTime() - new Date(a.last_message_at).getTime()
      )
    },
  })

  const threads = accountId && !contactId ? accountThreadsQuery.data ?? [] : threadsPage?.data ?? []
  const loadingThreads = accountId && !contactId ? contactsQuery.isLoading || accountThreadsQuery.isLoading : loadingThreadsBase

  const { mutateAsync: sendEmail } = useSendEmail()
  const { mutate: markRead } = useMarkThreadRead()
  const { mutateAsync: connectAccount } = useConnectEmailAccount()

  const unreadCount = threads.filter((t) => t.unread).length

  useEffect(() => {
    if (selectedThreadId && !threads.some((thread) => thread.thread_id === selectedThreadId)) {
      setSelectedThreadId(null)
    }
  }, [selectedThreadId, threads])

  function handleSelectThread(t: InboxThread) {
    setSelectedThreadId(t.thread_id)
    if (t.unread) markRead(t.thread_id)
  }

  async function handleConnect(provider: 'gmail' | 'outlook') {
    const { redirect_url } = await connectAccount({ provider, redirect_uri: window.location.href })
    window.open(redirect_url, '_blank', 'width=600,height=700')
  }

  async function handleComposeSend(payload: {
    accountId: string; to: string[]; cc?: string[]; bcc?: string[]
    subject: string; body: string
  }) {
    await sendEmail({
      connection_id: payload.accountId,
      to: payload.to,
      cc: payload.cc,
      bcc: payload.bcc,
      subject: payload.subject,
      body_html: `<p>${payload.body.replace(/\n/g, '<br/>')}</p>`,
    })
  }

  if (loadingAccounts) {
    return (
      <div className="flex flex-1 items-center justify-center py-24">
        <RefreshCw className="h-6 w-6 animate-spin text-[#1B3A4B]" />
      </div>
    )
  }

  // No accounts — show connect screen
  if (accounts.length === 0) {
    return <ConnectAccountsEmptyState onConnect={handleConnect} />
  }

  return (
    <div className="flex h-[calc(100vh-var(--topbar-height))] flex-col overflow-hidden -m-4 lg:-m-6">
      {/* Page header */}
      <div className="flex shrink-0 items-center justify-between border-b border-[#E5E7EB] bg-white px-4 py-3 lg:px-6">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-[#1A1D23]">Email.</h1>
          {unreadCount > 0 && (
            <span className="rounded-full bg-[#EF4444] px-2 py-0.5 text-[11px] font-semibold text-white">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button aria-label="Search emails" className="rounded-md p-2 hover:bg-[#F3F4F6]">
            <Search className="h-4 w-4 text-[#6B7280]" />
          </button>
          <button
            onClick={() => setShowCompose(true)}
            className="flex items-center gap-2 rounded-md bg-[#1B3A4B] px-3 py-2 text-sm font-semibold text-white transition-colors hover:bg-[#152E3C]"
          >
            <Plus className="h-4 w-4" /> Compose
          </button>
        </div>
      </div>

      {(accountId || contactId) && (
        <div className="flex shrink-0 items-center gap-3 border-b border-[#F3F4F6] bg-[#F8FAFC] px-4 py-2 lg:px-6">
          <span className="text-sm font-medium text-[#374151]">
            {contactId
              ? contactName
                ? `Filtered to ${contactName}`
                : 'Contact filter active'
              : accountName
                ? `Filtered to ${accountName}`
                : 'Account filter active'}
          </span>
          <button
            onClick={() =>
              setSearchParams((prev) => {
                const next = new URLSearchParams(prev)
                next.delete('account_id')
                next.delete('account_name')
                next.delete('contact_id')
                next.delete('contact_name')
                return next
              })
            }
            className="text-[12px] font-medium text-[#1B3A4B] hover:underline"
          >
            Clear
          </button>
        </div>
      )}

      {/* Account filter tabs */}
      <nav
        aria-label="Email accounts filter"
        className="flex shrink-0 items-center gap-2 border-b border-[#F3F4F6] bg-white px-4 py-2 lg:px-6"
      >
        <button
          onClick={() => setFilterAccountId(null)}
          className={cn(
            'rounded-full px-3 py-1 text-[12px] font-semibold transition-colors',
            filterAccountId === null
              ? 'bg-[#1B3A4B] text-white'
              : 'bg-[#F3F4F6] text-[#6B7280] hover:bg-[#E5E7EB]'
          )}
        >
          All
        </button>
        {accounts.map((a) => (
          <button
            key={a.id}
            onClick={() => setFilterAccountId(a.id)}
            className={cn(
              'rounded-full px-3 py-1 text-[12px] font-semibold transition-colors',
              filterAccountId === a.id
                ? 'bg-[#1B3A4B] text-white'
                : 'bg-[#F3F4F6] text-[#6B7280] hover:bg-[#E5E7EB]'
            )}
          >
            {a.email_address}
          </button>
        ))}
        <label className="ml-auto flex items-center gap-1.5 text-[12px] text-[#6B7280] cursor-pointer select-none">
          <input
            type="checkbox"
            checked={unreadOnly}
            onChange={(e) => setUnreadOnly(e.target.checked)}
            className="accent-[#1B3A4B]"
          />
          Unread only
        </label>
      </nav>

      {/* Two-column layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: thread list (380px) */}
        <div
          role="listbox"
          aria-label="Email threads"
          className={cn(
            'flex flex-col overflow-y-auto border-r border-[#E5E7EB] bg-white',
            selectedThreadId ? 'hidden md:flex md:w-[380px] md:shrink-0' : 'flex w-full md:w-[380px] md:shrink-0'
          )}
        >
          {loadingThreads && (
            <div className="space-y-3 p-4">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="h-[72px] animate-pulse rounded-md bg-[#F3F4F6]" />
              ))}
            </div>
          )}

          {!loadingThreads && threads.length === 0 && (
            <div className="flex flex-1 flex-col items-center justify-center gap-3 py-16">
              <Inbox className="h-10 w-10 text-[#D1D5DB]" />
              <p className="text-sm text-[#9CA3AF]">
                {accountId
                  ? unreadOnly
                    ? 'No unread emails for this account.'
                    : 'No inbox threads matched this account yet.'
                  : contactId
                    ? unreadOnly
                      ? 'No unread emails for this contact.'
                      : 'No inbox threads matched this contact yet.'
                  : unreadOnly
                    ? 'No unread emails.'
                    : 'No emails from this account yet.'}
              </p>
              {(filterAccountId || unreadOnly) && (
                <button
                  onClick={() => { setFilterAccountId(null); setUnreadOnly(false) }}
                  className="text-[12px] font-medium text-[#1B3A4B] hover:underline"
                >
                  Clear filter
                </button>
              )}
            </div>
          )}

          {threads.map((t) => (
            <ThreadRow
              key={t.thread_id}
              thread={t}
              account={accounts.find((a) => a.id === t.connection_id)}
              isSelected={selectedThreadId === t.thread_id}
              onClick={() => handleSelectThread(t)}
            />
          ))}
        </div>

        {/* Right: thread detail */}
        {selectedThreadId ? (
          <ThreadDetailPanel threadId={selectedThreadId} accounts={accounts} />
        ) : (
          <div className="hidden flex-1 items-center justify-center md:flex">
            <div className="text-center">
              <Mail className="mx-auto h-10 w-10 text-[#D1D5DB]" />
              <p className="mt-2 text-sm text-[#9CA3AF]">Select an email to read</p>
            </div>
          </div>
        )}
      </div>

      {/* Compose modal */}
      {showCompose && (
        <ComposeModal
          accounts={accounts}
          onSend={handleComposeSend}
          onClose={() => setShowCompose(false)}
        />
      )}
    </div>
  )
}
