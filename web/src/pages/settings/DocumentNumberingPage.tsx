import { useEffect, useMemo, useState, type ElementType, type FormEvent } from 'react'
import { AlertCircle, BookOpen, FileText, Hash, Receipt, Save, Ticket } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useNumberingSettings, useUpdateNumberingSettings } from '@/hooks/useNumbering'
import type { NumberingSettings, UpdateNumberingSettingsRequest } from '@/api/numbering'

type NumberingField = keyof UpdateNumberingSettingsRequest

interface FieldMeta {
  key: NumberingField
  label: string
  description: string
  prefix: string
  icon: ElementType
}

const FIELDS: FieldMeta[] = [
  {
    key: 'quote_number_start',
    label: 'Quotes',
    description: 'First number used when allocating new quote IDs.',
    prefix: 'QUO',
    icon: FileText,
  },
  {
    key: 'ticket_number_start',
    label: 'Tickets',
    description: 'First number used when allocating new support ticket IDs.',
    prefix: 'TCK',
    icon: Ticket,
  },
  {
    key: 'kb_article_number_start',
    label: 'KB Articles',
    description: 'First number used when allocating new knowledge base article IDs.',
    prefix: 'KB',
    icon: BookOpen,
  },
  {
    key: 'invoice_number_start',
    label: 'Invoices',
    description: 'First number used when allocating new invoice IDs.',
    prefix: 'INV',
    icon: Receipt,
  },
]

function formFromSettings(settings?: NumberingSettings): Record<NumberingField, string> {
  return {
    quote_number_start: String(settings?.quote_number_start ?? 1),
    ticket_number_start: String(settings?.ticket_number_start ?? 1),
    kb_article_number_start: String(settings?.kb_article_number_start ?? 1),
    invoice_number_start: String(settings?.invoice_number_start ?? 1),
  }
}

function validate(form: Record<NumberingField, string>) {
  const errors: Partial<Record<NumberingField, string>> = {}
  for (const field of FIELDS) {
    const raw = form[field.key].trim()
    const parsed = Number(raw)
    if (!raw || !Number.isInteger(parsed) || parsed < 1) {
      errors[field.key] = 'Enter a positive whole number.'
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
    return FIELDS.reduce<Record<NumberingField, string>>((acc, field) => {
      const parsed = Number(form[field.key])
      acc[field.key] = Number.isInteger(parsed) && parsed > 0
        ? `${field.prefix}-${String(parsed).padStart(5, '0')}`
        : `${field.prefix}-00001`
      return acc
    }, {} as Record<NumberingField, string>)
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
      acc[field.key] = Number(form[field.key])
      return acc
    }, {})

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
            const error = fieldErrors[field.key]
            return (
              <section
                key={field.key}
                className="rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm"
                aria-labelledby={`${field.key}-label`}
              >
                <div className="flex items-start gap-4">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#E8EDF2]">
                    <Icon className="h-5 w-5 text-[#1B3A4B]" aria-hidden="true" />
                  </div>
                  <div className="min-w-0 flex-1 space-y-3">
                    <div>
                      <h2 id={`${field.key}-label`} className="text-sm font-bold text-[#1A1D23]">
                        {field.label}
                      </h2>
                      <p className="mt-0.5 text-xs text-[#6B7280]">{field.description}</p>
                    </div>
                    <div className="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-end">
                      <div>
                        <label
                          htmlFor={field.key}
                          className="mb-1 block text-[11px] font-bold uppercase tracking-widest text-[#6B7280]"
                        >
                          Starting number
                        </label>
                        <Input
                          id={field.key}
                          inputMode="numeric"
                          min={1}
                          step={1}
                          type="number"
                          value={form[field.key]}
                          onChange={(e) => {
                            setForm((current) => ({ ...current, [field.key]: e.target.value }))
                            setFieldErrors((current) => ({ ...current, [field.key]: undefined }))
                            setStatusMessage('')
                          }}
                          aria-invalid={!!error}
                          aria-describedby={error ? `${field.key}-error` : `${field.key}-preview`}
                          disabled={isLoading || updateNumbering.isPending}
                        />
                        {error && (
                          <p id={`${field.key}-error`} className="mt-1 text-xs text-red-600">
                            {error}
                          </p>
                        )}
                      </div>
                      <div
                        id={`${field.key}-preview`}
                        className="rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-2 text-xs font-semibold text-[#1B3A4B]"
                      >
                        {preview[field.key]}
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
