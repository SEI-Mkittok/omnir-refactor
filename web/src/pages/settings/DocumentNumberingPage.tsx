import { useEffect, useMemo, useState, type ElementType, type FormEvent } from 'react'
import { AlertCircle, BookOpen, FileText, Hash, Receipt, Save, Ticket } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useNumberingSettings, useUpdateNumberingSettings } from '@/hooks/useNumbering'
import type { NumberingSettings, UpdateNumberingSettingsRequest } from '@/api/numbering'

type NumberingStartField =
  | 'quote_number_start'
  | 'ticket_number_start'
  | 'kb_article_number_start'
  | 'invoice_number_start'

type NumberingPrefixField =
  | 'quote_number_prefix'
  | 'ticket_number_prefix'
  | 'kb_article_number_prefix'
  | 'invoice_number_prefix'

type NumberingCurrentField =
  | 'quote_number_current'
  | 'ticket_number_current'
  | 'kb_article_number_current'
  | 'invoice_number_current'

type NumberingField = NumberingStartField | NumberingPrefixField | NumberingCurrentField

interface FieldMeta {
  startKey: NumberingStartField
  prefixKey: NumberingPrefixField
  currentKey: NumberingCurrentField
  label: string
  description: string
  fallbackPrefix: string
  icon: ElementType
}

const FIELDS: FieldMeta[] = [
  {
    startKey: 'quote_number_start',
    prefixKey: 'quote_number_prefix',
    currentKey: 'quote_number_current',
    label: 'Quotes',
    description: 'First number used when allocating new quote IDs.',
    fallbackPrefix: 'QUO',
    icon: FileText,
  },
  {
    startKey: 'ticket_number_start',
    prefixKey: 'ticket_number_prefix',
    currentKey: 'ticket_number_current',
    label: 'Tickets',
    description: 'First number used when allocating new support ticket IDs.',
    fallbackPrefix: 'TCK',
    icon: Ticket,
  },
  {
    startKey: 'kb_article_number_start',
    prefixKey: 'kb_article_number_prefix',
    currentKey: 'kb_article_number_current',
    label: 'KB Articles',
    description: 'First number used when allocating new knowledge base article IDs.',
    fallbackPrefix: 'KB',
    icon: BookOpen,
  },
  {
    startKey: 'invoice_number_start',
    prefixKey: 'invoice_number_prefix',
    currentKey: 'invoice_number_current',
    label: 'Invoices',
    description: 'First number used when allocating new invoice IDs.',
    fallbackPrefix: 'INV',
    icon: Receipt,
  },
]

function formFromSettings(settings?: NumberingSettings): Record<NumberingField, string> {
  return {
    quote_number_start: String(settings?.quote_number_start ?? 1),
    ticket_number_start: String(settings?.ticket_number_start ?? 1),
    kb_article_number_start: String(settings?.kb_article_number_start ?? 1),
    invoice_number_start: String(settings?.invoice_number_start ?? 1),
    quote_number_prefix: settings?.quote_number_prefix ?? 'QUO',
    ticket_number_prefix: settings?.ticket_number_prefix ?? 'TCK',
    kb_article_number_prefix: settings?.kb_article_number_prefix ?? 'KB',
    invoice_number_prefix: settings?.invoice_number_prefix ?? 'INV',
    quote_number_current: String(settings?.quote_number_current ?? 0),
    ticket_number_current: String(settings?.ticket_number_current ?? 0),
    kb_article_number_current: String(settings?.kb_article_number_current ?? 0),
    invoice_number_current: String(settings?.invoice_number_current ?? 0),
  }
}

function validate(form: Record<NumberingField, string>) {
  const errors: Partial<Record<NumberingField, string>> = {}
  for (const field of FIELDS) {
    const startRaw = form[field.startKey].trim()
    const startParsed = Number(startRaw)
    if (!startRaw || !Number.isInteger(startParsed) || startParsed < 1) {
      errors[field.startKey] = 'Enter a positive whole number.'
    }

    const currentRaw = form[field.currentKey].trim()
    const currentParsed = Number(currentRaw)
    if (!currentRaw || !Number.isInteger(currentParsed) || currentParsed < 0) {
      errors[field.currentKey] = 'Enter zero or a positive whole number.'
    }

    const prefixRaw = form[field.prefixKey].trim().toUpperCase()
    if (!prefixRaw) {
      errors[field.prefixKey] = 'Prefix is required.'
    }
    if (prefixRaw.length > 12) {
      errors[field.prefixKey] = 'Prefix must be 12 characters or fewer.'
    }
  }
  return errors
}

export function DocumentNumberingPage() {
  const { data, isLoading, isError } = useNumberingSettings()
  const updateNumbering = useUpdateNumberingSettings()
  const [form, setForm] = useState<Record<NumberingField, string>>(formFromSettings())
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<NumberingField, string>>>({})
  const [statusMessage, setStatusMessage] = useState('')

  useEffect(() => {
    if (data) {
      setForm(formFromSettings(data))
      setFieldErrors({})
      setStatusMessage('')
    }
  }, [data])

  const preview = useMemo(() => {
    return FIELDS.reduce<Record<NumberingStartField, string>>((acc, field) => {
      const start = Number(form[field.startKey])
      const current = Number(form[field.currentKey])
      const prefix = form[field.prefixKey].trim().toUpperCase() || field.fallbackPrefix
      const next = Number.isInteger(current) && current >= 0
        ? current + 1
        : Number.isInteger(start) && start > 0
          ? start
          : 1
      acc[field.startKey] = `${prefix}-${String(next).padStart(5, '0')}`
      return acc
    }, {} as Record<NumberingStartField, string>)
  }, [form])

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setStatusMessage('')
    const errors = validate(form)
    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) {
      return
    }

    const payload = FIELDS.reduce<UpdateNumberingSettingsRequest>((acc, field) => {
      acc[field.startKey] = Number(form[field.startKey])
      acc[field.prefixKey] = form[field.prefixKey].trim().toUpperCase()
      acc[field.currentKey] = Number(form[field.currentKey])
      return acc
    }, {} as UpdateNumberingSettingsRequest)

    updateNumbering.mutate(payload, {
      onSuccess: () => setStatusMessage('Document numbering saved.'),
      onError: () => setStatusMessage('Unable to save document numbering. Please try again.'),
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Document Numbering</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Configure starting values for newly allocated document numbers. Existing numbers are not
          rewritten.
        </p>
      </div>

      {isError && (
        <div
          role="alert"
          className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          Failed to load document numbering settings.
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <div className="grid gap-3 lg:grid-cols-2">
          {FIELDS.map((field) => {
            const Icon = field.icon
            const startError = fieldErrors[field.startKey]
            const prefixError = fieldErrors[field.prefixKey]
            const currentError = fieldErrors[field.currentKey]
            return (
              <section
                key={field.startKey}
                className="rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm"
                aria-labelledby={`${field.startKey}-label`}
              >
                <div className="flex items-start gap-4">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#E8EDF2]">
                    <Icon className="h-5 w-5 text-[#1B3A4B]" aria-hidden="true" />
                  </div>
                  <div className="min-w-0 flex-1 space-y-3">
                    <div>
                      <h2 id={`${field.startKey}-label`} className="text-sm font-bold text-[#1A1D23]">
                        {field.label}
                      </h2>
                      <p className="mt-0.5 text-xs text-[#6B7280]">{field.description}</p>
                    </div>
                    <div className="grid gap-2 sm:grid-cols-[1fr_1fr_1fr_auto] sm:items-end">
                      <div>
                        <label
                          htmlFor={field.prefixKey}
                          className="mb-1 block text-[11px] font-bold uppercase tracking-widest text-[#6B7280]"
                        >
                          Prefix
                        </label>
                        <Input
                          id={field.prefixKey}
                          value={form[field.prefixKey]}
                          maxLength={12}
                          onChange={(e) => {
                            setForm((current) => ({ ...current, [field.prefixKey]: e.target.value.toUpperCase() }))
                            setFieldErrors((current) => ({ ...current, [field.prefixKey]: undefined }))
                            setStatusMessage('')
                          }}
                          aria-invalid={!!prefixError}
                          aria-describedby={prefixError ? `${field.prefixKey}-error` : undefined}
                          disabled={isLoading || updateNumbering.isPending}
                        />
                        {prefixError && (
                          <p id={`${field.prefixKey}-error`} className="mt-1 text-xs text-red-600">
                            {prefixError}
                          </p>
                        )}
                      </div>
                      <div>
                        <label
                          htmlFor={field.startKey}
                          className="mb-1 block text-[11px] font-bold uppercase tracking-widest text-[#6B7280]"
                        >
                          Starting number
                        </label>
                        <Input
                          id={field.startKey}
                          inputMode="numeric"
                          min={1}
                          step={1}
                          type="number"
                          value={form[field.startKey]}
                          onChange={(e) => {
                            setForm((current) => ({ ...current, [field.startKey]: e.target.value }))
                            setFieldErrors((current) => ({ ...current, [field.startKey]: undefined }))
                            setStatusMessage('')
                          }}
                          aria-invalid={!!startError}
                          aria-describedby={startError ? `${field.startKey}-error` : `${field.startKey}-preview`}
                          disabled={isLoading || updateNumbering.isPending}
                        />
                        {startError && (
                          <p id={`${field.startKey}-error`} className="mt-1 text-xs text-red-600">
                            {startError}
                          </p>
                        )}
                      </div>
                      <div>
                        <label
                          htmlFor={field.currentKey}
                          className="mb-1 block text-[11px] font-bold uppercase tracking-widest text-[#6B7280]"
                        >
                          Current sequence
                        </label>
                        <Input
                          id={field.currentKey}
                          inputMode="numeric"
                          min={0}
                          step={1}
                          type="number"
                          value={form[field.currentKey]}
                          onChange={(e) => {
                            setForm((current) => ({ ...current, [field.currentKey]: e.target.value }))
                            setFieldErrors((current) => ({ ...current, [field.currentKey]: undefined }))
                            setStatusMessage('')
                          }}
                          aria-invalid={!!currentError}
                          aria-describedby={currentError ? `${field.currentKey}-error` : `${field.startKey}-preview`}
                          disabled={isLoading || updateNumbering.isPending}
                        />
                        {currentError && (
                          <p id={`${field.currentKey}-error`} className="mt-1 text-xs text-red-600">
                            {currentError}
                          </p>
                        )}
                      </div>
                      <div
                        id={`${field.startKey}-preview`}
                        className="rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-2 text-xs font-semibold text-[#1B3A4B]"
                      >
                        {preview[field.startKey]}
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            )
          })}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3">
          {statusMessage ? (
            <p
              role={updateNumbering.isError ? 'alert' : 'status'}
              className={updateNumbering.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}
            >
              {statusMessage}
            </p>
          ) : (
            <span />
          )}
          <Button type="submit" disabled={isLoading || updateNumbering.isPending}>
            <Save className="h-4 w-4" />
            {updateNumbering.isPending ? 'Saving...' : 'Save Numbering'}
          </Button>
        </div>
      </form>
    </div>
  )
}
