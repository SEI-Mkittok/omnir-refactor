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

// ── API ────────────────────────────────────────────────────────────────────

export const billingApi = {
  getPlans: () =>
    apiClient.get<Plan[]>('/billing/plans').then((r) => r.data),

  getSubscription: () =>
    apiClient.get<Subscription>('/billing/subscription').then((r) => r.data),

  getUsage: () =>
    apiClient.get<UsageMeters>('/billing/usage').then((r) => r.data),

  getInvoices: (cursor?: string) =>
    apiClient
      .get<InvoiceListResponse>('/billing/invoices', { params: cursor ? { cursor } : undefined })
      .then((r) => r.data),

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
