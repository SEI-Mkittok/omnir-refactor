import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  CreditCard,
  Download,
  AlertTriangle,
  ChevronDown,
  ExternalLink,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import {
  billingApi,
  usageStatus,
  usagePercent,
  formatPrice,
  type Invoice,
  type InvoiceStatus,
} from '@/api/billing'

// ── Usage Meter ──────────────────────────────────────────────────────────────

function UsageMeter({
  label,
  used,
  limit,
}: {
  label: string
  used: number
  limit: number | null
}) {
  const status = usageStatus(used, limit)
  const pct = usagePercent(used, limit)

  const trackColor = '#E5E7EB'
  const fillColor =
    status === 'critical' ? '#EF4444' : status === 'warning' ? '#F59E0B' : '#1B3A4B'

  const displayLimit = limit === null ? 'unlimited' : limit.toLocaleString()
  const displayUsed = used.toLocaleString()

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>
          {label}
        </span>
        <span className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
          {displayUsed}
          <span className="font-normal text-xs ml-1" style={{ color: 'var(--text-secondary)' }}>
            / {displayLimit}
          </span>
        </span>
      </div>
      {limit !== null ? (
        <div
          className="h-2 w-full rounded-full overflow-hidden"
          style={{ background: trackColor }}
          role="meter"
          aria-label={label}
          aria-valuenow={used}
          aria-valuemin={0}
          aria-valuemax={limit}
        >
          <div
            className="h-full rounded-full transition-all duration-300"
            style={{ width: `${pct}%`, background: fillColor }}
          />
        </div>
      ) : (
        <div
          className="h-2 w-full rounded-full"
          style={{ background: trackColor }}
          role="meter"
          aria-label={`${label} — unlimited`}
          aria-valuenow={0}
          aria-valuemin={0}
          aria-valuemax={100}
        />
      )}
      {status === 'critical' && (
        <p className="text-xs font-medium" style={{ color: '#EF4444' }}>
          Limit reached — upgrade to continue.
        </p>
      )}
    </div>
  )
}

// ── Invoice Status Badge ─────────────────────────────────────────────────────

function InvoiceStatusBadge({ status }: { status: InvoiceStatus }) {
  const styles: Record<InvoiceStatus, { label: string; bg: string; text: string }> = {
    paid:    { label: 'PAID',    bg: '#DCFCE7', text: '#15803D' },
    failed:  { label: 'FAILED',  bg: '#FEE2E2', text: '#DC2626' },
    pending: { label: 'PENDING', bg: '#FEF3C7', text: '#D97706' },
  }
  const s = styles[status]
  return (
    <span
      className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold uppercase"
      style={{ background: s.bg, color: s.text, letterSpacing: '0.04em' }}
    >
      {s.label}
    </span>
  )
}

// ── Cancel Subscription Dialog ───────────────────────────────────────────────

function CancelDialog({
  periodEnd,
  onConfirm,
  onClose,
  loading,
}: {
  periodEnd: string
  onConfirm: () => void
  onClose: () => void
  loading: boolean
}) {
  const date = new Date(periodEnd).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })

  return (
    <div
      className="fixed inset-0 z-[var(--z-modal)] flex items-center justify-center p-4"
      style={{ background: 'var(--surface-overlay)' }}
    >
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="cancel-dialog-title"
        className="w-full max-w-md rounded-xl p-6 shadow-[var(--shadow-modal)]"
        style={{ background: 'var(--surface-card)' }}
      >
        <div className="mb-4 flex items-center gap-3">
          <div
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
            style={{ background: '#FEE2E2' }}
          >
            <AlertTriangle className="h-5 w-5" style={{ color: '#EF4444' }} />
          </div>
          <h2
            id="cancel-dialog-title"
            className="text-base font-semibold"
            style={{ color: 'var(--text-primary)' }}
          >
            Cancel subscription?
          </h2>
        </div>

        <p className="mb-6 text-sm" style={{ color: 'var(--text-secondary)' }}>
          You'll lose access to Pro features on <strong>{date}</strong>. Your data will be retained
          but certain features will be disabled.
        </p>

        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Button
            type="button"
            onClick={onConfirm}
            disabled={loading}
            className="border-none text-white"
            style={{ background: '#EF4444' }}
          >
            {loading ? 'Cancelling…' : 'Cancel Subscription'}
          </Button>
          {/* Pre-focus the safe action */}
          <Button
            type="button"
            variant="outline"
            autoFocus
            onClick={onClose}
            style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
          >
            Keep My Plan
          </Button>
        </div>
      </div>
    </div>
  )
}

// ── Main Page ────────────────────────────────────────────────────────────────

export function BillingSettingsPage() {
  const qc = useQueryClient()
  const [cancelOpen, setCancelOpen] = useState(false)
  const [invoiceCursor, setInvoiceCursor] = useState<string | undefined>()
  const [allInvoices, setAllInvoices] = useState<Invoice[]>([])
  const [portalError, setPortalError] = useState<string | null>(null)

  const {
    data: sub,
    isLoading: subLoading,
    isError: subError,
    refetch: refetchSub,
  } = useQuery({ queryKey: ['billing', 'subscription'], queryFn: billingApi.getSubscription })

  const { data: usage, isLoading: usageLoading } = useQuery({
    queryKey: ['billing', 'usage'],
    queryFn: billingApi.getUsage,
  })

  const {
    data: invoicePage,
    isLoading: invoicesLoading,
    isFetching: invoicesFetching,
  } = useQuery({
    queryKey: ['billing', 'invoices', invoiceCursor],
    queryFn: () => billingApi.getInvoices(invoiceCursor),
    placeholderData: (prev) => prev,
  })

  // Accumulate invoice pages
  const invoiceData = invoicePage?.data ?? []
  const invoices: Invoice[] =
    allInvoices.length === 0 && invoiceCursor === undefined ? invoiceData : allInvoices

  function handleLoadMore() {
    if (!invoicePage?.nextCursor) return
    setAllInvoices((prev) => [...prev, ...(invoicePage.data ?? [])])
    setInvoiceCursor(invoicePage.nextCursor ?? undefined)
  }

  const cancelMutation = useMutation({
    mutationFn: billingApi.cancelSubscription,
    onSuccess: () => {
      setCancelOpen(false)
      qc.invalidateQueries({ queryKey: ['billing', 'subscription'] })
    },
  })

  const portalMutation = useMutation({
    mutationFn: billingApi.openPortal,
    onSuccess: (data) => {
      window.location.href = data.url
    },
    onError: () => {
      setPortalError('Unable to open billing portal. Please try again.')
      setTimeout(() => setPortalError(null), 4000)
    },
  })

  const planLabel: Record<string, string> = {
    free: 'Free',
    pro: 'Pro',
    enterprise: 'Enterprise',
    legacy: 'Legacy',
  }

  const periodEnd = sub?.currentPeriodEnd
    ? new Date(sub.currentPeriodEnd).toLocaleDateString('en-US', {
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      })
    : '—'

  if (subError) {
    return (
      <div className="space-y-6">
        <PageHeader />
        <div
          className="flex flex-col items-center gap-3 rounded-xl p-8 text-center"
          style={{ border: '1px solid var(--border-default)', background: 'var(--surface-card)' }}
        >
          <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
            Unable to load billing info. Please try again.
          </p>
          <Button variant="outline" size="sm" onClick={() => refetchSub()}>
            <RefreshCw className="h-3.5 w-3.5" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader />

      {/* ── Current Plan Card ── */}
      <section
        className="rounded-xl p-5"
        style={{
          border: sub?.planTier === 'pro' ? '2px solid var(--color-primary)' : '1px solid var(--border-default)',
          background: 'var(--surface-card)',
          boxShadow: 'var(--shadow-card)',
        }}
      >
        {subLoading ? (
          <SkeletonLines count={3} />
        ) : (
          <>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span
                    className="text-lg font-bold uppercase tracking-wide"
                    style={{ color: 'var(--text-primary)' }}
                  >
                    {planLabel[sub?.planTier ?? 'free']} Plan
                  </span>
                  <PlanStatusBadge status={sub?.status ?? 'active'} cancelled={sub?.cancelAtPeriodEnd ?? false} />
                </div>
                {sub?.planTier !== 'free' && sub?.planTier !== 'enterprise' && (
                  <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                    {sub?.period === 'annual' ? 'Annual plan' : '$49/month'} · Renews {periodEnd}
                  </p>
                )}
                {sub?.cancelAtPeriodEnd && (
                  <p className="text-sm" style={{ color: '#EF4444' }}>
                    Cancels on {periodEnd}
                  </p>
                )}
                {sub?.trialEnd && (
                  <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                    Free trial until{' '}
                    {new Date(sub.trialEnd).toLocaleDateString('en-US', {
                      month: 'long',
                      day: 'numeric',
                      year: 'numeric',
                    })}
                  </p>
                )}
              </div>
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button asChild size="sm" style={{ background: 'var(--color-primary)', color: '#fff' }}>
                <Link to="/settings/billing/plans">Upgrade Plan</Link>
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={portalMutation.isPending}
                onClick={() => portalMutation.mutate()}
                style={{ borderColor: 'var(--border-default)', color: 'var(--text-primary)' }}
              >
                <CreditCard className="h-3.5 w-3.5" />
                {portalMutation.isPending ? 'Opening…' : 'Manage Payment Method'}
                <ExternalLink className="h-3 w-3 opacity-50" />
              </Button>
            </div>
            {portalError && (
              <p className="mt-2 text-xs" style={{ color: '#EF4444' }} role="alert">
                {portalError}
              </p>
            )}
          </>
        )}
      </section>

      {/* ── Usage Meters ── */}
      <section
        className="rounded-xl p-5 space-y-4"
        style={{ border: '1px solid var(--border-default)', background: 'var(--surface-card)', boxShadow: 'var(--shadow-card)' }}
      >
        <h2
          className="text-xs font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          Usage this period
        </h2>
        {usageLoading ? (
          <SkeletonLines count={3} />
        ) : usage ? (
          <div className="space-y-4">
            <UsageMeter label="Seats" used={usage.seats.used} limit={usage.seats.limit} />
            <UsageMeter label="Tickets" used={usage.tickets.used} limit={usage.tickets.limit} />
            <UsageMeter label="API Calls" used={usage.apiCalls.used} limit={usage.apiCalls.limit} />
          </div>
        ) : null}
      </section>

      {/* ── Invoice History ── */}
      <section
        className="rounded-xl overflow-hidden"
        style={{ border: '1px solid var(--border-default)', background: 'var(--surface-card)', boxShadow: 'var(--shadow-card)' }}
      >
        <div className="px-5 py-4 border-b" style={{ borderColor: 'var(--border-default)' }}>
          <h2
            className="text-xs font-semibold uppercase tracking-widest"
            style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
          >
            Invoice History
          </h2>
        </div>

        {invoicesLoading ? (
          <div className="p-5">
            <SkeletonLines count={4} />
          </div>
        ) : invoices.length === 0 ? (
          <div className="p-8 text-center">
            <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
              No invoices yet. Your first invoice will appear here after your first billing cycle.
            </p>
          </div>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border-default)' }}>
                  {['Invoice', 'Date', 'Amount', 'Status', ''].map((h) => (
                    <th
                      key={h}
                      className="px-5 py-3 text-left text-xs font-semibold uppercase"
                      style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
                      scope="col"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => (
                  <InvoiceRow key={inv.id} invoice={inv} />
                ))}
              </tbody>
            </table>

            {invoicePage?.hasMore && (
              <div className="flex justify-center px-5 py-4" style={{ borderTop: '1px solid var(--border-default)' }}>
                <button
                  type="button"
                  onClick={handleLoadMore}
                  disabled={invoicesFetching}
                  className="text-sm font-medium transition-opacity"
                  style={{ color: 'var(--color-primary)' }}
                >
                  {invoicesFetching ? 'Loading…' : 'Load more invoices'}
                  {!invoicesFetching && <ChevronDown className="ml-1 inline h-3.5 w-3.5" />}
                </button>
              </div>
            )}
          </>
        )}
      </section>

      {/* ── Cancel Subscription ── */}
      {sub && sub.status !== 'cancelled' && !sub.cancelAtPeriodEnd && (
        <div className="pt-2">
          <button
            type="button"
            onClick={() => setCancelOpen(true)}
            className="text-sm transition-opacity hover:opacity-80"
            style={{ color: '#EF4444' }}
          >
            Cancel Subscription
          </button>
        </div>
      )}

      {cancelOpen && sub && (
        <CancelDialog
          periodEnd={sub.currentPeriodEnd}
          onConfirm={() => cancelMutation.mutate()}
          onClose={() => setCancelOpen(false)}
          loading={cancelMutation.isPending}
        />
      )}
    </div>
  )
}

// ── Sub-components ───────────────────────────────────────────────────────────

function PageHeader() {
  return (
    <div>
      <h1 className="text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>
        Billing.
      </h1>
      <p className="mt-1 text-sm" style={{ color: 'var(--text-secondary)' }}>
        Subscription, usage, and invoices
      </p>
    </div>
  )
}

function PlanStatusBadge({
  status,
  cancelled,
}: {
  status: string
  cancelled: boolean
}) {
  if (cancelled) {
    return (
      <span
        className="rounded-full px-2 py-0.5 text-xs font-semibold uppercase"
        style={{ background: '#FEF3C7', color: '#D97706', letterSpacing: '0.04em' }}
      >
        Cancelling
      </span>
    )
  }
  if (status === 'cancelled') {
    return (
      <span
        className="rounded-full px-2 py-0.5 text-xs font-semibold uppercase"
        style={{ background: '#FEF3C7', color: '#D97706', letterSpacing: '0.04em' }}
      >
        Cancelled
      </span>
    )
  }
  if (status === 'trialing') {
    return (
      <span
        className="rounded-full px-2 py-0.5 text-xs font-semibold uppercase"
        style={{ background: '#DBEAFE', color: '#1D4ED8', letterSpacing: '0.04em' }}
      >
        Trial
      </span>
    )
  }
  return (
    <span
      className="rounded-full px-2 py-0.5 text-xs font-semibold uppercase"
      style={{ background: '#DCFCE7', color: '#15803D', letterSpacing: '0.04em' }}
    >
      Active
    </span>
  )
}

function InvoiceRow({ invoice }: { invoice: Invoice }) {
  const date = new Date(invoice.date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })

  return (
    <tr
      className="transition-colors hover:bg-[#F7F8FA]"
      style={{ borderBottom: '1px solid var(--border-subtle)' }}
    >
      <td
        className="px-5 py-3 font-mono text-xs"
        style={{ color: 'var(--text-primary)' }}
      >
        {invoice.identifier}
      </td>
      <td className="px-5 py-3 text-sm" style={{ color: 'var(--text-secondary)' }}>
        {date}
      </td>
      <td className="px-5 py-3 text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
        {formatPrice(invoice.amountCents, invoice.currency)}
      </td>
      <td className="px-5 py-3">
        <div className="flex items-center gap-2">
          <InvoiceStatusBadge status={invoice.status} />
          {invoice.status === 'failed' && (
            <button
              type="button"
              className="text-xs underline"
              style={{ color: 'var(--color-primary)' }}
            >
              Retry Payment
            </button>
          )}
        </div>
      </td>
      <td className="px-5 py-3 text-right">
        {invoice.downloadUrl ? (
          <a
            href={invoice.downloadUrl}
            download
            aria-label={`Download invoice ${invoice.identifier} as PDF`}
            className="inline-flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-70"
            style={{ color: 'var(--color-primary)' }}
          >
            <Download className="h-3.5 w-3.5" />
            PDF
          </a>
        ) : null}
      </td>
    </tr>
  )
}

function SkeletonLines({ count }: { count: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="animate-pulse h-4 rounded"
          style={{ background: 'var(--border-default)', width: `${60 + (i % 3) * 15}%` }}
        />
      ))}
    </div>
  )
}
