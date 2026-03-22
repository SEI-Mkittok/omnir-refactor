import { useState, useCallback } from 'react'
import { Plus, Trash2, Send, Save, X } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { useProducts } from '@/hooks/useProducts'
import { useCreateQuote, useUpdateQuote, useSendQuote } from '@/hooks/useQuotes'
import { formatCurrency } from '@/lib/utils'
import type { Quote, QuoteLineItemInput, CreateQuoteRequest, UpdateQuoteRequest, Product } from '@/api/types'

interface LineItemRow extends QuoteLineItemInput {
  _key: number
}

interface QuoteBuilderProps {
  dealId?: string
  contactId?: string
  contactEmail?: string
  quote?: Quote | null
  onClose: () => void
  onSaved?: (quote: Quote) => void
}

function centsToAmount(cents: number): string {
  return (cents / 100).toFixed(2)
}

function amountToCents(s: string): number {
  const n = parseFloat(s)
  return isNaN(n) ? 0 : Math.round(n * 100)
}

function lineTotal(li: QuoteLineItemInput): number {
  const gross = (li.unit_price_cents / 100) * li.quantity
  const discount = gross * (li.discount_pct / 100)
  return Math.round((gross - discount) * 100) / 100
}

function quoteTotal(items: QuoteLineItemInput[]): number {
  return items.reduce((sum, li) => sum + lineTotal(li), 0)
}

let keyCounter = 0

export function QuoteBuilder({ dealId, contactId, contactEmail, quote, onClose, onSaved }: QuoteBuilderProps) {
  const isEdit = !!quote

  const [title, setTitle] = useState(quote?.title ?? '')
  const [currency, setCurrency] = useState(quote?.currency ?? 'USD')
  const [validUntil, setValidUntil] = useState(quote?.valid_until?.slice(0, 10) ?? '')
  const [notes, setNotes] = useState(quote?.notes ?? '')
  const [lineItems, setLineItems] = useState<LineItemRow[]>(() => {
    if (quote?.line_items?.length) {
      return quote.line_items.map((li) => ({
        _key: ++keyCounter,
        product_id: li.product_id,
        product_name: li.product_name,
        description: li.description,
        quantity: li.quantity,
        unit_price_cents: li.unit_price_cents,
        discount_pct: li.discount_pct,
        sort_order: li.sort_order,
      }))
    }
    return [newBlankLine()]
  })

  const [showProductPicker, setShowProductPicker] = useState<number | null>(null)
  const [showSendModal, setShowSendModal] = useState(false)
  const [sendTo, setSendTo] = useState(contactEmail ?? '')
  const [sendSubject, setSendSubject] = useState(title ? `Quote: ${title}` : 'Your Quote')
  const [sendMessage, setSendMessage] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: productsResult } = useProducts({ active: true, limit: 200 })
  const products = productsResult?.data ?? []

  const createQuote = useCreateQuote()
  const updateQuote = useUpdateQuote()
  const sendQuote = useSendQuote()

  function newBlankLine(): LineItemRow {
    return {
      _key: ++keyCounter,
      product_name: '',
      quantity: 1,
      unit_price_cents: 0,
      discount_pct: 0,
      sort_order: 0,
    }
  }

  const addLine = useCallback(() => {
    setLineItems((prev) => [...prev, newBlankLine()])
  }, [])

  const removeLine = useCallback((key: number) => {
    setLineItems((prev) => prev.filter((li) => li._key !== key))
  }, [])

  const updateLine = useCallback((key: number, updates: Partial<LineItemRow>) => {
    setLineItems((prev) => prev.map((li) => li._key === key ? { ...li, ...updates } : li))
  }, [])

  const applyProduct = useCallback((key: number, product: Product) => {
    updateLine(key, {
      product_id: product.id,
      product_name: product.name,
      unit_price_cents: product.unit_price_cents,
      description: product.description,
    })
    setShowProductPicker(null)
  }, [updateLine])

  function validate(): boolean {
    const errs: Record<string, string> = {}
    if (!title.trim()) errs.title = 'Required'
    if (lineItems.length === 0) errs.lines = 'Add at least one line item'
    lineItems.forEach((li, i) => {
      if (!li.product_name.trim()) errs[`line_${i}_name`] = 'Product name required'
    })
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  function buildPayload(): CreateQuoteRequest | UpdateQuoteRequest {
    return {
      title: title.trim(),
      currency,
      ...(validUntil ? { valid_until: validUntil } : {}),
      ...(notes.trim() ? { notes: notes.trim() } : {}),
      ...(dealId ? { deal_id: dealId } : {}),
      ...(contactId ? { contact_id: contactId } : {}),
      line_items: lineItems.map((li, idx) => ({
        product_id: li.product_id,
        product_name: li.product_name,
        description: li.description,
        quantity: li.quantity,
        unit_price_cents: li.unit_price_cents,
        discount_pct: li.discount_pct,
        sort_order: idx,
      })),
    }
  }

  async function handleSave() {
    if (!validate()) return
    const payload = buildPayload()
    let saved: Quote
    if (isEdit && quote) {
      saved = await updateQuote.mutateAsync({ id: quote.id, payload })
    } else {
      saved = await createQuote.mutateAsync(payload as CreateQuoteRequest)
    }
    onSaved?.(saved)
    onClose()
  }

  async function handleSend() {
    if (!validate()) return
    if (!sendTo.trim()) {
      setErrors((prev) => ({ ...prev, sendTo: 'Recipient email required' }))
      return
    }

    // Save first if new
    let quoteId = quote?.id
    if (!quoteId) {
      const saved = await createQuote.mutateAsync(buildPayload() as CreateQuoteRequest)
      quoteId = saved.id
    } else if (isEdit) {
      await updateQuote.mutateAsync({ id: quoteId, payload: buildPayload() })
    }

    await sendQuote.mutateAsync({
      id: quoteId,
      payload: { to: sendTo.trim(), subject: sendSubject, message: sendMessage },
    })
    setShowSendModal(false)
    onClose()
  }

  const isPending = createQuote.isPending || updateQuote.isPending || sendQuote.isPending
  const total = quoteTotal(lineItems)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative z-10 w-full max-w-4xl max-h-[90vh] flex flex-col rounded-xl bg-white shadow-xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">
            {isEdit ? 'Edit Quote' : 'New Quote'}
          </h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Basic fields */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="sm:col-span-2 space-y-1">
              <label className="block text-sm font-medium text-slate-700">
                Quote title <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Q-2026-001 — Acme Corp"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
              {errors.title && <p className="text-xs text-red-600">{errors.title}</p>}
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Currency</label>
              <select
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              >
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
                <option value="GBP">GBP</option>
                <option value="CAD">CAD</option>
                <option value="AUD">AUD</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Valid until</label>
              <input
                type="date"
                value={validUntil}
                onChange={(e) => setValidUntil(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>

            <div className="sm:col-span-2 space-y-1">
              <label className="block text-sm font-medium text-slate-700">Notes</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={2}
                placeholder="Payment terms, delivery details…"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>
          </div>

          {/* Line items */}
          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-slate-900">Line Items</h3>
              <Button variant="outline" size="sm" onClick={addLine}>
                <Plus className="h-3.5 w-3.5 mr-1" /> Add item
              </Button>
            </div>

            {errors.lines && <p className="mb-2 text-xs text-red-600">{errors.lines}</p>}

            <div className="space-y-2">
              {/* Column headers */}
              <div className="hidden sm:grid sm:grid-cols-12 gap-2 text-xs font-medium text-slate-500 px-1">
                <span className="col-span-4">Product</span>
                <span className="col-span-2">Qty</span>
                <span className="col-span-2">Unit price</span>
                <span className="col-span-2">Disc %</span>
                <span className="col-span-1 text-right">Total</span>
                <span className="col-span-1" />
              </div>

              {lineItems.map((li, idx) => (
                <div key={li._key} className="grid grid-cols-12 gap-2 items-start">
                  {/* Product name / picker */}
                  <div className="col-span-12 sm:col-span-4 relative">
                    <input
                      type="text"
                      value={li.product_name}
                      onChange={(e) => updateLine(li._key, { product_name: e.target.value, product_id: undefined })}
                      onFocus={() => setShowProductPicker(li._key)}
                      placeholder="Product or description"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                    />
                    {errors[`line_${idx}_name`] && (
                      <p className="text-xs text-red-600">{errors[`line_${idx}_name`]}</p>
                    )}

                    {/* Product picker dropdown */}
                    {showProductPicker === li._key && products.length > 0 && (
                      <div className="absolute top-full left-0 z-20 mt-1 w-64 rounded-lg border border-slate-200 bg-white shadow-lg">
                        <div className="max-h-48 overflow-y-auto py-1">
                          {products
                            .filter((p) =>
                              !li.product_name || p.name.toLowerCase().includes(li.product_name.toLowerCase())
                            )
                            .map((p) => (
                              <button
                                key={p.id}
                                type="button"
                                className="flex w-full items-center justify-between px-3 py-2 text-sm hover:bg-slate-50"
                                onClick={() => applyProduct(li._key, p)}
                              >
                                <span className="font-medium text-slate-900">{p.name}</span>
                                <span className="text-slate-500 ml-2">
                                  {formatCurrency(p.unit_price_cents / 100, p.currency)}
                                </span>
                              </button>
                            ))}
                        </div>
                        <button
                          type="button"
                          className="w-full px-3 py-1.5 text-xs text-slate-500 border-t border-slate-100 hover:bg-slate-50 text-left"
                          onClick={() => setShowProductPicker(null)}
                        >
                          Close
                        </button>
                      </div>
                    )}
                  </div>

                  {/* Qty */}
                  <div className="col-span-4 sm:col-span-2">
                    <input
                      type="number"
                      value={li.quantity}
                      min={0.001}
                      step="0.001"
                      onChange={(e) => updateLine(li._key, { quantity: parseFloat(e.target.value) || 1 })}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                    />
                  </div>

                  {/* Unit price */}
                  <div className="col-span-4 sm:col-span-2">
                    <input
                      type="number"
                      value={centsToAmount(li.unit_price_cents)}
                      min={0}
                      step="0.01"
                      onChange={(e) => updateLine(li._key, { unit_price_cents: amountToCents(e.target.value) })}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                    />
                  </div>

                  {/* Discount % */}
                  <div className="col-span-3 sm:col-span-2">
                    <input
                      type="number"
                      value={li.discount_pct}
                      min={0}
                      max={100}
                      step="0.1"
                      onChange={(e) => updateLine(li._key, { discount_pct: parseFloat(e.target.value) || 0 })}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                    />
                  </div>

                  {/* Line total */}
                  <div className="col-span-1 sm:col-span-1 flex items-center justify-end pt-2 sm:pt-0">
                    <span className="text-sm font-medium text-slate-900 whitespace-nowrap">
                      {formatCurrency(lineTotal(li), currency)}
                    </span>
                  </div>

                  {/* Remove */}
                  <div className="col-span-12 sm:col-span-1 flex justify-end sm:justify-center">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="text-red-500 hover:text-red-700 h-8 w-8"
                      onClick={() => removeLine(li._key)}
                      disabled={lineItems.length === 1}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>

            {/* Totals */}
            <div className="mt-4 flex justify-end">
              <div className="space-y-1 text-right min-w-48">
                <div className="flex justify-between gap-8 text-sm text-slate-500">
                  <span>Subtotal</span>
                  <span>{formatCurrency(total, currency)}</span>
                </div>
                <div className="flex justify-between gap-8 text-base font-bold text-slate-900 border-t border-slate-200 pt-1 mt-1">
                  <span>Total</span>
                  <span>{formatCurrency(total, currency)}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Approval status badge for existing quotes */}
          {quote && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">Status:</span>
              <QuoteStatusBadge status={quote.status} />
              {quote.sent_at && (
                <span className="text-xs text-slate-400">
                  Sent {new Date(quote.sent_at).toLocaleDateString()}
                </span>
              )}
              {quote.approved_at && (
                <span className="text-xs text-green-600">
                  Approved {new Date(quote.approved_at).toLocaleDateString()}
                </span>
              )}
              {quote.rejected_at && (
                <span className="text-xs text-red-600">
                  Rejected {new Date(quote.rejected_at).toLocaleDateString()}
                </span>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="border-t border-slate-200 px-6 py-4 flex justify-between gap-3">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={handleSave}
              disabled={isPending}
            >
              <Save className="h-4 w-4 mr-1.5" />
              Save draft
            </Button>
            <Button
              onClick={() => {
                if (!validate()) return
                setSendSubject(title ? `Quote: ${title}` : 'Your Quote')
                setShowSendModal(true)
              }}
              disabled={isPending}
            >
              <Send className="h-4 w-4 mr-1.5" />
              Send quote
            </Button>
          </div>
        </div>
      </div>

      {/* Send modal */}
      {showSendModal && (
        <div className="fixed inset-0 z-60 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setShowSendModal(false)} />
          <div className="relative z-10 w-full max-w-md rounded-xl bg-white shadow-xl p-6 space-y-4">
            <h3 className="text-lg font-semibold text-slate-900">Send Quote</h3>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">
                Recipient email <span className="text-red-500">*</span>
              </label>
              <input
                type="email"
                value={sendTo}
                onChange={(e) => setSendTo(e.target.value)}
                placeholder="contact@example.com"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
              {errors.sendTo && <p className="text-xs text-red-600">{errors.sendTo}</p>}
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Subject</label>
              <input
                type="text"
                value={sendSubject}
                onChange={(e) => setSendSubject(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Message</label>
              <textarea
                value={sendMessage}
                onChange={(e) => setSendMessage(e.target.value)}
                rows={3}
                placeholder="Optional message to the recipient…"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button variant="outline" onClick={() => setShowSendModal(false)}>Cancel</Button>
              <Button onClick={handleSend} disabled={isPending}>
                <Send className="h-4 w-4 mr-1.5" />
                {sendQuote.isPending ? 'Sending…' : 'Send'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export function QuoteStatusBadge({ status }: { status: string }) {
  const map: Record<string, 'default' | 'blue' | 'green' | 'red' | 'yellow' | 'indigo'> = {
    draft: 'default',
    sent: 'blue',
    approved: 'green',
    rejected: 'red',
    expired: 'yellow',
  }
  const labels: Record<string, string> = {
    draft: 'Draft',
    sent: 'Sent',
    approved: 'Approved',
    rejected: 'Rejected',
    expired: 'Expired',
  }
  return <Badge variant={map[status] ?? 'default'}>{labels[status] ?? status}</Badge>
}
