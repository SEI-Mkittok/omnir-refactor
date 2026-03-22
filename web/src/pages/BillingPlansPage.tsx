import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, X, AlertCircle, CheckCircle2, RefreshCw, Lock } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { billingApi, PLANS, type Plan, type PlanTier, type BillingPeriod } from '@/api/billing'

// ── Feature rows shown on each card ─────────────────────────────────────────

const FEATURE_ROWS: Array<{ label: string; key: keyof Plan['features'] | 'seats' | 'tickets' | 'apiCalls' }> = [
  { label: 'Seats', key: 'seats' },
  { label: 'Tickets / month', key: 'tickets' },
  { label: 'SLA rules', key: 'slaRules' },
  { label: 'Reports', key: 'reports' },
  { label: 'API access', key: 'apiAccess' },
  { label: 'Custom fields', key: 'customFields' },
  { label: 'White-label', key: 'whiteLabel' },
]

function getFeatureValue(plan: Plan, key: string): string | boolean {
  if (key === 'seats') return plan.seats === null ? 'Unlimited' : `${plan.seats}`
  if (key === 'tickets') return plan.ticketsPerMonth === null ? 'Unlimited' : `${plan.ticketsPerMonth}`
  if (key === 'apiCalls') return plan.apiCallsPerMonth === null ? 'Unlimited' : plan.apiCallsPerMonth.toLocaleString()
  return plan.features[key as keyof Plan['features']]
}

// ── Upgrade Modal ────────────────────────────────────────────────────────────

interface UpgradeModalProps {
  plan: Plan
  period: BillingPeriod
  savedCard: { last4: string; brand: string } | null
  onClose: () => void
  onSuccess: () => void
}

function UpgradeModal({ plan, period, savedCard, onClose, onSuccess }: UpgradeModalProps) {
  const [useNewCard, setUseNewCard] = useState(!savedCard)
  const [succeeded, setSucceeded] = useState(false)
  const [stripeUnavailable, setStripeUnavailable] = useState(false)
  const closeRef = useRef<HTMLButtonElement>(null)

  const price = period === 'annual' ? plan.priceAnnual : plan.priceMonthly
  const nextRenewal = new Date()
  nextRenewal.setMonth(nextRenewal.getMonth() + (period === 'annual' ? 12 : 1))
  const renewalLabel = nextRenewal.toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })

  const checkoutMutation = useMutation({
    mutationFn: () => billingApi.createUpgradeCheckout({ tier: plan.tier, period }),
    onSuccess: () => {
      setSucceeded(true)
    },
    onError: () => {
      setStripeUnavailable(true)
    },
  })

  // Restore focus on close
  useEffect(() => {
    return () => {
      // Trigger close-button focus restore handled by caller
    }
  }, [])

  if (succeeded) {
    return (
      <ModalShell onClose={onClose} labelId="upgrade-modal-title">
        <div className="flex flex-col items-center gap-4 py-6 text-center">
          <CheckCircle2 className="h-12 w-12" style={{ color: '#22C55E' }} />
          <div>
            <h2
              id="upgrade-modal-title"
              className="text-lg font-bold"
              style={{ color: 'var(--text-primary)' }}
            >
              You're now on {plan.name}!
            </h2>
            <p className="mt-1 text-sm" style={{ color: 'var(--text-secondary)' }}>
              A receipt will be emailed to you.
            </p>
          </div>
          <Button
            onClick={() => { onSuccess(); onClose() }}
            style={{ background: 'var(--color-primary)', color: '#fff' }}
          >
            Close
          </Button>
        </div>
      </ModalShell>
    )
  }

  return (
    <ModalShell onClose={onClose} labelId="upgrade-modal-title">
      <div className="flex items-center justify-between mb-5">
        <h2
          id="upgrade-modal-title"
          className="text-base font-semibold"
          style={{ color: 'var(--text-primary)' }}
        >
          Upgrade to {plan.name}
        </h2>
        <button
          ref={closeRef}
          type="button"
          aria-label="Close upgrade dialog"
          onClick={onClose}
          disabled={checkoutMutation.isPending}
          className="rounded p-1 transition-colors hover:bg-[#F7F8FA]"
          style={{ color: 'var(--text-secondary)' }}
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Plan Summary */}
      <div
        className="rounded-lg p-4 mb-5"
        style={{ background: 'var(--surface-app)', border: '1px solid var(--border-default)' }}
      >
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
            {plan.name} Plan · {period === 'annual' ? 'Annual' : 'Monthly'}
          </span>
          <span className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
            {price !== null ? `$${price}.00 / month` : 'Custom pricing'}
          </span>
        </div>
        {price !== null && (
          <p className="mt-1 text-xs" style={{ color: 'var(--text-secondary)' }}>
            Billed today: ${price}.00 · Next renewal: {renewalLabel}
          </p>
        )}
      </div>

      {/* Payment Method */}
      {stripeUnavailable ? (
        <div
          className="rounded-lg p-4 mb-5 text-sm"
          style={{ background: '#FEF3C7', color: '#92400E', border: '1px solid #FDE68A' }}
        >
          Payment system unavailable. Please try again later.
        </div>
      ) : (
        <div className="mb-5 space-y-3">
          <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
            Payment Method
          </p>

          {savedCard && (
            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="card-select"
                  checked={!useNewCard}
                  onChange={() => setUseNewCard(false)}
                  className="accent-[#1B3A4B]"
                />
                <span className="text-sm" style={{ color: 'var(--text-primary)' }}>
                  Use saved card ···· ···· ···· {savedCard.last4}
                </span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="card-select"
                  checked={useNewCard}
                  onChange={() => setUseNewCard(true)}
                  className="accent-[#1B3A4B]"
                />
                <span className="text-sm" style={{ color: 'var(--text-primary)' }}>
                  Add new card
                </span>
              </label>
            </div>
          )}

          {(!savedCard || useNewCard) && (
            <div
              className="rounded-lg p-3 text-xs"
              style={{ border: '1px solid var(--border-default)', color: 'var(--text-secondary)' }}
            >
              {/* Stripe Elements iframe placeholder — in production, mount Stripe Elements here */}
              <div className="flex items-center justify-center h-10 rounded" style={{ background: 'var(--surface-app)' }}>
                <Lock className="h-3.5 w-3.5 mr-1.5" style={{ color: 'var(--text-label)' }} />
                <span style={{ color: 'var(--text-label)' }}>Stripe card input</span>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Error */}
      {checkoutMutation.isError && !stripeUnavailable && (
        <div
          className="flex items-center gap-2 mb-4 rounded-lg p-3 text-sm"
          role="alert"
          style={{ background: '#FEF2F2', color: '#DC2626', border: '1px solid #FECACA' }}
        >
          <AlertCircle className="h-4 w-4 shrink-0" />
          Payment failed. Please try again.
        </div>
      )}

      {/* CTA */}
      <Button
        className="w-full text-sm font-semibold"
        style={{ background: 'var(--color-primary)', color: '#fff' }}
        disabled={checkoutMutation.isPending || stripeUnavailable}
        onClick={() => checkoutMutation.mutate()}
      >
        {checkoutMutation.isPending ? (
          <span className="flex items-center gap-2">
            <RefreshCw className="h-4 w-4 animate-spin" /> Processing…
          </span>
        ) : (
          `Confirm & Pay${price !== null ? ` $${price}.00` : ''}`
        )}
      </Button>

      <p className="mt-3 flex items-center justify-center gap-1 text-xs" style={{ color: 'var(--text-secondary)' }}>
        <Lock className="h-3 w-3" />
        Secured by Stripe
      </p>
    </ModalShell>
  )
}

function ModalShell({
  children,
  onClose,
  labelId,
}: {
  children: React.ReactNode
  onClose: () => void
  labelId: string
}) {
  // Trap focus inside modal
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const focusable = el.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    first?.focus()

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key !== 'Tab') return
      if (e.shiftKey) {
        if (document.activeElement === first) { e.preventDefault(); last?.focus() }
      } else {
        if (document.activeElement === last) { e.preventDefault(); first?.focus() }
      }
    }
    el.addEventListener('keydown', handleKeyDown)
    return () => el.removeEventListener('keydown', handleKeyDown)
  }, [])

  return (
    <div
      className="fixed inset-0 z-[var(--z-modal)] flex items-center justify-center p-4"
      style={{ background: 'var(--surface-overlay)', backdropFilter: 'blur(2px)' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose() }}
    >
      <div
        ref={ref}
        role="dialog"
        aria-modal="true"
        aria-labelledby={labelId}
        className="w-full max-w-md rounded-xl p-6"
        style={{
          background: 'var(--surface-card)',
          boxShadow: 'var(--shadow-modal)',
        }}
      >
        {children}
      </div>
    </div>
  )
}

// ── Downgrade Confirmation ───────────────────────────────────────────────────

function DowngradeDialog({
  targetPlan,
  lostFeatures,
  onConfirm,
  onClose,
}: {
  targetPlan: string
  lostFeatures: string[]
  onConfirm: () => void
  onClose: () => void
}) {
  return (
    <ModalShell onClose={onClose} labelId="downgrade-dialog-title">
      <h2
        id="downgrade-dialog-title"
        className="text-base font-semibold mb-3"
        style={{ color: 'var(--text-primary)' }}
      >
        Downgrade to {targetPlan}?
      </h2>
      <p className="text-sm mb-3" style={{ color: 'var(--text-secondary)' }}>
        You will lose access to the following features:
      </p>
      <ul className="mb-5 space-y-1">
        {lostFeatures.map((f) => (
          <li key={f} className="flex items-center gap-2 text-sm" style={{ color: '#EF4444' }}>
            <X className="h-3.5 w-3.5 shrink-0" />
            {f}
          </li>
        ))}
      </ul>
      <div className="flex gap-2">
        <Button
          className="border-none text-white"
          style={{ background: '#EF4444' }}
          onClick={onConfirm}
        >
          Confirm Downgrade
        </Button>
        <Button variant="outline" autoFocus onClick={onClose}>
          Keep Current Plan
        </Button>
      </div>
    </ModalShell>
  )
}

// ── Plan Card ────────────────────────────────────────────────────────────────

interface PlanCardProps {
  plan: Plan
  isCurrent: boolean
  isPopular: boolean
  period: BillingPeriod
  onUpgrade: (plan: Plan) => void
  onDowngrade: (plan: Plan) => void
}

function PlanCard({ plan, isCurrent, isPopular, period, onUpgrade, onDowngrade }: PlanCardProps) {
  const price = period === 'annual' ? plan.priceAnnual : plan.priceMonthly
  const isEnterprise = plan.tier === 'enterprise'

  return (
    <div
      className="relative flex flex-col rounded-xl p-6"
      role="region"
      aria-label={`${plan.name} plan`}
      style={{
        background: 'var(--surface-card)',
        border: isCurrent ? '2px solid var(--color-primary)' : '1px solid var(--border-default)',
        boxShadow: 'var(--shadow-card)',
      }}
    >
      {/* Badges */}
      {isPopular && !isCurrent && (
        <span
          className="absolute -top-3 right-4 rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase"
          style={{ background: '#F59E0B', color: '#fff', letterSpacing: '0.06em' }}
        >
          Most Popular
        </span>
      )}
      {isCurrent && (
        <span
          className="absolute -top-3 left-4 rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase"
          style={{ background: '#22C55E', color: '#fff', letterSpacing: '0.06em' }}
        >
          Current Plan
        </span>
      )}

      {/* Plan name & price */}
      <div className="mb-5">
        <h3 className="text-base font-bold uppercase tracking-wide" style={{ color: 'var(--text-primary)' }}>
          {plan.name}
        </h3>
        {isEnterprise ? (
          <p className="mt-1 text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>
            Custom
          </p>
        ) : (
          <p className="mt-1">
            <span className="text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>
              ${price}
            </span>
            <span className="text-sm ml-1" style={{ color: 'var(--text-secondary)' }}>
              /mo
            </span>
            {period === 'annual' && plan.tier !== 'free' && (
              <span
                className="ml-2 rounded-full px-1.5 py-0.5 text-xs font-medium"
                style={{ background: '#DCFCE7', color: '#15803D' }}
              >
                SAVE 10%
              </span>
            )}
          </p>
        )}
      </div>

      {/* Feature list */}
      <ul className="mb-6 flex-1 space-y-2.5" aria-label="Included features">
        {FEATURE_ROWS.map(({ label, key }) => {
          const val = getFeatureValue(plan, key)
          const included = typeof val === 'boolean' ? val : true
          const displayVal = typeof val === 'string' ? val : null

          return (
            <li key={key} className="flex items-center gap-2 text-sm">
              {included ? (
                <Check className="h-4 w-4 shrink-0" style={{ color: '#22C55E' }} aria-hidden="true" />
              ) : (
                <X className="h-4 w-4 shrink-0" style={{ color: 'var(--text-secondary)' }} aria-hidden="true" />
              )}
              <span style={{ color: included ? 'var(--text-primary)' : 'var(--text-secondary)' }}>
                {displayVal ? `${displayVal} ${label.toLowerCase()}` : label}
              </span>
            </li>
          )
        })}
      </ul>

      {/* CTA */}
      {isEnterprise ? (
        <a
          href="mailto:sales@praestos.io"
          className="block w-full rounded-lg border px-4 py-2.5 text-center text-sm font-semibold transition-colors"
          style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
          aria-label="Contact us about Enterprise plan"
        >
          Contact Us
        </a>
      ) : isCurrent ? (
        <button
          type="button"
          disabled
          className="w-full rounded-lg px-4 py-2.5 text-sm font-semibold opacity-50 cursor-not-allowed"
          style={{ background: 'var(--border-default)', color: 'var(--text-secondary)' }}
        >
          Current Plan
        </button>
      ) : (
        <Button
          className="w-full text-sm font-semibold"
          style={{ background: 'var(--color-primary)', color: '#fff' }}
          onClick={() => {
            // Simple heuristic: free is a downgrade from pro
            if (plan.tier === 'free') onDowngrade(plan)
            else onUpgrade(plan)
          }}
          aria-label={`Upgrade to ${plan.name} plan`}
        >
          Upgrade to {plan.name}
        </Button>
      )}
    </div>
  )
}

// ── Main Page ────────────────────────────────────────────────────────────────

export function BillingPlansPage() {
  const qc = useQueryClient()
  const [period, setPeriod] = useState<BillingPeriod>('monthly')
  const [upgradeTarget, setUpgradeTarget] = useState<Plan | null>(null)
  const [downgradeTarget, setDowngradeTarget] = useState<Plan | null>(null)
  const [billingError] = useState<string | null>(null)

  const { data: sub, isLoading: subLoading, isError: subError, refetch } = useQuery({
    queryKey: ['billing', 'subscription'],
    queryFn: billingApi.getSubscription,
  })

  const currentTier: PlanTier = sub?.planTier ?? 'free'

  // Legacy plan — add a read-only card before the 3-tier grid
  const isLegacy = currentTier === 'legacy'

  function handleUpgradeSuccess() {
    qc.invalidateQueries({ queryKey: ['billing', 'subscription'] })
  }

  if (subError) {
    return (
      <div className="space-y-6">
        <PlansHeader period={period} onPeriodChange={setPeriod} />
        <div
          className="flex flex-col items-center gap-3 rounded-xl p-8 text-center"
          style={{ border: '1px solid var(--border-default)', background: 'var(--surface-card)' }}
        >
          <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
            Unable to load billing info. Try again.
          </p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className="h-3.5 w-3.5" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PlansHeader period={period} onPeriodChange={setPeriod} />

      {billingError && (
        <div
          className="rounded-lg px-4 py-3 text-sm"
          role="alert"
          style={{ background: '#FEF2F2', color: '#DC2626', border: '1px solid #FECACA' }}
        >
          {billingError}
        </div>
      )}

      {/* Legacy plan card */}
      {isLegacy && (
        <div
          className="rounded-xl p-5"
          style={{ border: '2px solid var(--color-primary)', background: 'var(--surface-card)' }}
          role="region"
          aria-label="Legacy plan"
        >
          <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
            Legacy Plan <span className="ml-2 rounded-full px-2 py-0.5 text-xs" style={{ background: '#E5E7EB', color: 'var(--text-secondary)' }}>Current</span>
          </p>
          <p className="mt-1 text-xs" style={{ color: 'var(--text-secondary)' }}>
            You're on a grandfathered plan. Your current pricing and features are preserved.
          </p>
        </div>
      )}

      {/* 3-tier grid */}
      {subLoading ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className="animate-pulse h-96 rounded-xl"
              style={{ background: 'var(--border-default)' }}
            />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          {PLANS.map((plan) => (
            <PlanCard
              key={plan.tier}
              plan={plan}
              isCurrent={!isLegacy && plan.tier === currentTier}
              isPopular={plan.tier === 'pro'}
              period={period}
              onUpgrade={setUpgradeTarget}
              onDowngrade={setDowngradeTarget}
            />
          ))}
        </div>
      )}

      {/* Upgrade modal */}
      {upgradeTarget && (
        <UpgradeModal
          plan={upgradeTarget}
          period={period}
          savedCard={sub?.savedCard ?? null}
          onClose={() => setUpgradeTarget(null)}
          onSuccess={handleUpgradeSuccess}
        />
      )}

      {/* Downgrade confirmation */}
      {downgradeTarget && (
        <DowngradeDialog
          targetPlan={downgradeTarget.name}
          lostFeatures={['SLA rules', 'Reports', 'API access']}
          onConfirm={() => {
            // In production, call downgrade endpoint; here just close
            setDowngradeTarget(null)
          }}
          onClose={() => setDowngradeTarget(null)}
        />
      )}
    </div>
  )
}

// ── Sub-components ───────────────────────────────────────────────────────────

function PlansHeader({
  period,
  onPeriodChange,
}: {
  period: BillingPeriod
  onPeriodChange: (p: BillingPeriod) => void
}) {
  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 className="text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>
          Billing.
        </h1>
        <p className="mt-1 text-sm" style={{ color: 'var(--text-secondary)' }}>
          Manage your subscription and invoices
        </p>
      </div>

      {/* Monthly / Annual toggle */}
      <div
        role="group"
        aria-label="Billing period"
        className="flex items-center rounded-full p-0.5"
        style={{ border: '1px solid var(--border-default)', background: 'var(--surface-app)' }}
      >
        {(['monthly', 'annual'] as BillingPeriod[]).map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => onPeriodChange(p)}
            className="rounded-full px-4 py-1.5 text-sm font-medium transition-all"
            style={
              period === p
                ? { background: 'var(--color-primary)', color: '#fff' }
                : { color: 'var(--text-secondary)' }
            }
          >
            {p === 'monthly' ? 'Monthly' : 'Annual (10% off)'}
          </button>
        ))}
      </div>
    </div>
  )
}
