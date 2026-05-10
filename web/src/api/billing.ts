import apiClient from './client'

// ── Types ──────────────────────────────────────────────────────────────────

export type PlanTier = 'free' | 'pro' | 'enterprise' | 'legacy'
export type BillingPeriod = 'monthly' | 'annual'
export type InvoiceStatus = 'paid' | 'failed' | 'pending'
export type PlanStatus = 'active' | 'cancelled' | 'trialing'

export interface Plan {
  tier: PlanTier
  name: string
  priceMonthly: number | null  // null = custom (enterprise)
  priceAnnual: number | null
  seats: number | null          // null = unlimited
  ticketsPerMonth: number | null
  apiCallsPerMonth: number | null
  features: {
    slaRules: boolean
    reports: boolean
    apiAccess: boolean
    customFields: boolean
    whiteLabel: boolean
  }
}

export interface Subscription {
  id: string
  planTier: PlanTier
  status: PlanStatus
  period: BillingPeriod
  currentPeriodEnd: string   // ISO date
  cancelAtPeriodEnd: boolean
  trialEnd: string | null
  savedCard: SavedCard | null
}

export interface SavedCard {
  last4: string
  brand: string
  expMonth: number
  expYear: number
}

export interface UsageMeters {
  seats: { used: number; limit: number | null }
  tickets: { used: number; limit: number | null }
  apiCalls: { used: number; limit: number | null }
}

export interface Invoice {
  id: string
  identifier: string
  date: string               // ISO date
  amountCents: number
  currency: string
  status: InvoiceStatus
  downloadUrl: string | null
}

export interface InvoiceListResponse {
  data: Invoice[]
  hasMore: boolean
  nextCursor: string | null
}

interface UsageMeterShape {
  used?: unknown
  limit?: unknown
}

interface LegacyUsageResponse {
  usage?: {
    users?: UsageMeterShape
    contacts?: UsageMeterShape
    storage_mb?: UsageMeterShape
  }
}

interface LegacySubscriptionResponse {
  id?: unknown
  plan?: unknown
  status?: unknown
  current_period_end?: unknown
  period?: unknown
  cancel_at_period_end?: unknown
  trial_end?: unknown
  saved_card?: unknown
}

interface LegacyInvoiceResponse {
  id?: unknown
  stripe_invoice_id?: unknown
  identifier?: unknown
  date?: unknown
  created_at?: unknown
  amountCents?: unknown
  amount_cents?: unknown
  currency?: unknown
  status?: unknown
  downloadUrl?: unknown
  pdf_url?: unknown
}

interface LegacyInvoiceListResponse {
  data?: unknown
  meta?: {
    page?: unknown
    total_pages?: unknown
  }
  hasMore?: unknown
  nextCursor?: unknown
  next_cursor?: unknown
}

function toNumber(value: unknown, fallback = 0): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

function toNullableNumber(value: unknown): number | null {
  if (value === null) return null
  if (typeof value === 'number' && Number.isFinite(value)) return value
  return null
}

function normalizeUsageMeter(raw: UsageMeterShape | undefined): { used: number; limit: number | null } {
  return {
    used: toNumber(raw?.used, 0),
    limit: toNullableNumber(raw?.limit),
  }
}

function normalizeUsage(raw: unknown): UsageMeters {
  const usage = raw as Partial<UsageMeters> & LegacyUsageResponse

  const seatsRaw = usage.seats ?? usage.usage?.users
  const ticketsRaw = usage.tickets ?? usage.usage?.contacts
  const apiCallsRaw = usage.apiCalls ?? usage.usage?.storage_mb

  return {
    seats: normalizeUsageMeter(seatsRaw),
    tickets: normalizeUsageMeter(ticketsRaw),
    apiCalls: normalizeUsageMeter(apiCallsRaw),
  }
}

function normalizeSubscriptionStatus(raw: unknown): PlanStatus {
  switch (raw) {
    case 'active':
      return 'active'
    case 'trialing':
      return 'trialing'
    case 'cancelled':
    case 'canceled':
      return 'cancelled'
    default:
      return 'active'
  }
}

function normalizePlanTier(raw: unknown): PlanTier {
  switch (raw) {
    case 'free':
    case 'pro':
    case 'enterprise':
    case 'legacy':
      return raw
    default:
      return 'free'
  }
}

function normalizeBillingPeriod(raw: unknown): BillingPeriod {
  return raw === 'annual' ? 'annual' : 'monthly'
}

function normalizeSavedCard(raw: unknown): SavedCard | null {
  if (!raw || typeof raw !== 'object') return null
  const card = raw as Partial<SavedCard>
  if (
    typeof card.last4 !== 'string' ||
    typeof card.brand !== 'string' ||
    typeof card.expMonth !== 'number' ||
    typeof card.expYear !== 'number'
  ) {
    return null
  }
  return {
    last4: card.last4,
    brand: card.brand,
    expMonth: card.expMonth,
    expYear: card.expYear,
  }
}

function normalizeSubscription(raw: unknown): Subscription {
  const data = raw as Partial<Subscription> & LegacySubscriptionResponse
  return {
    id: typeof data.id === 'string' ? data.id : 'subscription',
    planTier: normalizePlanTier(data.planTier ?? data.plan),
    status: normalizeSubscriptionStatus(data.status),
    period: normalizeBillingPeriod(data.period),
    currentPeriodEnd:
      typeof data.currentPeriodEnd === 'string'
        ? data.currentPeriodEnd
        : typeof data.current_period_end === 'string'
          ? data.current_period_end
          : new Date().toISOString(),
    cancelAtPeriodEnd:
      typeof data.cancelAtPeriodEnd === 'boolean'
        ? data.cancelAtPeriodEnd
        : Boolean(data.cancel_at_period_end),
    trialEnd:
      typeof data.trialEnd === 'string'
        ? data.trialEnd
        : typeof data.trial_end === 'string'
          ? data.trial_end
          : null,
    savedCard: normalizeSavedCard(data.savedCard ?? data.saved_card),
  }
}

function normalizeInvoiceStatus(raw: unknown): InvoiceStatus {
  switch (raw) {
    case 'paid':
      return 'paid'
    case 'failed':
    case 'uncollectible':
    case 'void':
      return 'failed'
    default:
      return 'pending'
  }
}

function normalizeInvoice(raw: unknown): Invoice {
  const invoice = raw as LegacyInvoiceResponse
  const id =
    typeof invoice.id === 'string'
      ? invoice.id
      : typeof invoice.stripe_invoice_id === 'string'
        ? invoice.stripe_invoice_id
        : 'invoice'

  const identifier =
    typeof invoice.identifier === 'string'
      ? invoice.identifier
      : typeof invoice.stripe_invoice_id === 'string'
        ? invoice.stripe_invoice_id
        : id

  const date =
    typeof invoice.date === 'string'
      ? invoice.date
      : typeof invoice.created_at === 'string'
        ? invoice.created_at
        : new Date().toISOString()

  const amountCents =
    typeof invoice.amountCents === 'number'
      ? invoice.amountCents
      : typeof invoice.amount_cents === 'number'
        ? invoice.amount_cents
        : 0

  return {
    id,
    identifier,
    date,
    amountCents,
    currency: typeof invoice.currency === 'string' ? invoice.currency.toUpperCase() : 'USD',
    status: normalizeInvoiceStatus(invoice.status),
    downloadUrl:
      typeof invoice.downloadUrl === 'string'
        ? invoice.downloadUrl
        : typeof invoice.pdf_url === 'string'
          ? invoice.pdf_url
          : null,
  }
}

function normalizeInvoiceList(raw: unknown): InvoiceListResponse {
  const data = raw as LegacyInvoiceListResponse
  const rows = Array.isArray(data.data) ? data.data : []
  const invoices = rows.map(normalizeInvoice)

  const explicitHasMore = typeof data.hasMore === 'boolean' ? data.hasMore : null
  const explicitNext =
    typeof data.nextCursor === 'string'
      ? data.nextCursor
      : typeof data.next_cursor === 'string'
        ? data.next_cursor
        : null

  if (explicitHasMore !== null || explicitNext !== null) {
    return {
      data: invoices,
      hasMore: explicitHasMore ?? explicitNext !== null,
      nextCursor: explicitNext,
    }
  }

  const page = typeof data.meta?.page === 'number' ? data.meta.page : 1
  const totalPages = typeof data.meta?.total_pages === 'number' ? data.meta.total_pages : 1
  const hasMore = page < totalPages

  return {
    data: invoices,
    hasMore,
    nextCursor: hasMore ? String(page + 1) : null,
  }
}

// ── API ────────────────────────────────────────────────────────────────────

export const billingApi = {
  getPlans: () =>
    apiClient.get<Plan[]>('/billing/plans').then((r) => r.data),

  getSubscription: () =>
    apiClient.get<unknown>('/billing/subscription').then((r) => normalizeSubscription(r.data)),

  getUsage: () =>
    apiClient.get<unknown>('/billing/usage').then((r) => normalizeUsage(r.data)),

  getInvoices: (cursor?: string) =>
    apiClient
      .get<unknown>('/billing/invoices', {
        params: cursor
          ? {
              cursor,
              ...(Number.isFinite(Number(cursor)) ? { page: Number(cursor) } : {}),
            }
          : undefined,
      })
      .then((r) => normalizeInvoiceList(r.data)),

  createUpgradeCheckout: (payload: { tier: PlanTier; period: BillingPeriod }) =>
    apiClient
      .post<{ clientSecret: string; subscriptionId: string }>('/billing/checkout', payload)
      .then((r) => r.data),

  cancelSubscription: () =>
    apiClient.post<Subscription>('/billing/subscription/cancel').then((r) => r.data),

  openPortal: () =>
    apiClient.post<{ url: string }>('/billing/portal').then((r) => r.data),
}

// ── Helpers ────────────────────────────────────────────────────────────────

export const PLANS: Plan[] = [
  {
    tier: 'free',
    name: 'Free',
    priceMonthly: 0,
    priceAnnual: 0,
    seats: 2,
    ticketsPerMonth: 100,
    apiCallsPerMonth: null,
    features: { slaRules: false, reports: false, apiAccess: false, customFields: false, whiteLabel: false },
  },
  {
    tier: 'pro',
    name: 'Pro',
    priceMonthly: 49,
    priceAnnual: 44,
    seats: 25,
    ticketsPerMonth: null,
    apiCallsPerMonth: 10000,
    features: { slaRules: true, reports: true, apiAccess: true, customFields: false, whiteLabel: false },
  },
  {
    tier: 'enterprise',
    name: 'Enterprise',
    priceMonthly: null,
    priceAnnual: null,
    seats: null,
    ticketsPerMonth: null,
    apiCallsPerMonth: null,
    features: { slaRules: true, reports: true, apiAccess: true, customFields: true, whiteLabel: true },
  },
]

export function usageStatus(used: number, limit: number | null): 'normal' | 'warning' | 'critical' {
  if (limit === null) return 'normal'
  const pct = used / limit
  if (pct >= 1) return 'critical'
  if (pct >= 0.8) return 'warning'
  return 'normal'
}

export function usagePercent(used: number, limit: number | null): number {
  if (limit === null) return 0
  return Math.min(100, Math.round((used / limit) * 100))
}

export function formatPrice(cents: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency, minimumFractionDigits: 2 }).format(
    cents / 100
  )
}
