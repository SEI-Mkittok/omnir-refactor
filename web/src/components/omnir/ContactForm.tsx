import { useState, useEffect, useRef } from 'react'
import { Loader2, Sparkles, X } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useCreateContact } from '@/hooks/useContacts'
import { useDomainLookup } from '@/hooks/useContacts'
import { useCustomFieldDefinitions } from '@/hooks/useCustomFields'
import { CustomFieldFormSection } from '@/components/omnir/CustomFieldRenderer'
import type { CreateContactRequest, CustomFieldValues } from '@/api/types'

interface ContactFormProps {
  open: boolean
  onClose: () => void
}

const INITIAL: CreateContactRequest = {
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  title: '',
  department: '',
  stage: 'prospect',
}

function domainFromEmail(email: string): string | null {
  const parts = email.split('@')
  if (parts.length !== 2 || !parts[1].includes('.')) return null
  return parts[1].toLowerCase().trim()
}

export function ContactForm({ open, onClose }: ContactFormProps) {
  const [form, setForm] = useState<CreateContactRequest>(INITIAL)
  const [errors, setErrors] = useState<Partial<Record<keyof CreateContactRequest, string>>>({})
  const [customFieldValues, setCustomFieldValues] = useState<CustomFieldValues>({})
  const [bannerDismissed, setBannerDismissed] = useState(false)
  const [lookupDomain, setLookupDomain] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const createContact = useCreateContact()
  const { data: customFields = [] } = useCustomFieldDefinitions('contact')
  const { data: enrichment, isFetching: enrichFetching } = useDomainLookup(lookupDomain)

  const set = (field: keyof CreateContactRequest) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const value = e.target.value
    setForm((f) => ({ ...f, [field]: value }))

    if (field === 'email') {
      setBannerDismissed(false)
      if (debounceRef.current) clearTimeout(debounceRef.current)
      debounceRef.current = setTimeout(() => {
        const domain = domainFromEmail(value)
        setLookupDomain(domain)
      }, 600)
    }
  }

  // Reset when dialog closes
  useEffect(() => {
    if (!open) {
      setForm(INITIAL)
      setErrors({})
      setCustomFieldValues({})
      setBannerDismissed(false)
      setLookupDomain(null)
    }
  }, [open])

  const validate = (): boolean => {
    const errs: typeof errors = {}
    if (!form.first_name.trim()) errs.first_name = 'Required'
    if (!form.last_name.trim()) errs.last_name = 'Required'
    if (!form.email.trim()) errs.email = 'Required'
    else if (!/\S+@\S+\.\S+/.test(form.email)) errs.email = 'Invalid email'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    await createContact.mutateAsync({
      ...form,
      phone: form.phone || undefined,
      title: form.title || undefined,
      department: form.department || undefined,
      custom_fields: Object.keys(customFieldValues).length ? customFieldValues : undefined,
    } as CreateContactRequest & { custom_fields?: CustomFieldValues })
    setForm(INITIAL)
    setErrors({})
    setCustomFieldValues({})
    setBannerDismissed(false)
    setLookupDomain(null)
    onClose()
  }

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setForm(INITIAL)
      setErrors({})
      setCustomFieldValues({})
      setBannerDismissed(false)
      setLookupDomain(null)
      onClose()
    }
  }

  const applyEnrichment = () => {
    if (!enrichment?.data) return
    setForm((f) => ({
      ...f,
      department: f.department || enrichment.data.industry || f.department,
    }))
    setBannerDismissed(true)
  }

  const showBanner =
    !bannerDismissed &&
    !enrichFetching &&
    !!enrichment?.data?.company_name &&
    !!lookupDomain

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>New Contact</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Enrichment banner */}
          {enrichFetching && lookupDomain && (
            <div
              className="flex items-center gap-2 rounded-md px-3 py-2 text-xs"
              style={{
                background: 'var(--color-primary-light)',
                color: 'var(--color-primary)',
                border: '1px solid var(--border-subtle)',
              }}
            >
              <Loader2 className="h-3 w-3 animate-spin shrink-0" />
              Looking up {lookupDomain}…
            </div>
          )}

          {showBanner && (
            <div
              className="flex items-start gap-2 rounded-md px-3 py-2.5"
              style={{
                background: 'var(--color-primary-light)',
                border: '1px solid var(--color-primary)',
                color: 'var(--text-primary)',
              }}
              role="status"
              aria-label="Company enrichment suggestion"
            >
              <Sparkles
                className="h-4 w-4 mt-0.5 shrink-0"
                style={{ color: 'var(--color-primary)' }}
              />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-semibold" style={{ color: 'var(--color-primary)' }}>
                  Company data found for {lookupDomain}
                </p>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-secondary)' }}>
                  <strong>{enrichment!.data.company_name}</strong>
                  {enrichment!.data.industry && ` · ${enrichment!.data.industry}`}
                  {enrichment!.data.size && ` · ${enrichment!.data.size}`}
                </p>
              </div>
              <div className="flex items-center gap-1.5 shrink-0">
                {enrichment?.data.industry && !form.department && (
                  <button
                    type="button"
                    onClick={applyEnrichment}
                    className="text-xs font-semibold px-2 py-1 rounded"
                    style={{
                      background: 'var(--color-primary)',
                      color: '#FFFFFF',
                      border: 'none',
                      cursor: 'pointer',
                    }}
                  >
                    Apply
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => setBannerDismissed(true)}
                  aria-label="Dismiss"
                  style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}
                >
                  <X className="h-3.5 w-3.5" style={{ color: 'var(--text-label)' }} />
                </button>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">
                First name <span className="text-red-500">*</span>
              </label>
              <Input
                value={form.first_name}
                onChange={set('first_name')}
                placeholder="Jane"
                aria-invalid={!!errors.first_name}
              />
              {errors.first_name && (
                <p className="mt-0.5 text-xs text-red-500">{errors.first_name}</p>
              )}
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">
                Last name <span className="text-red-500">*</span>
              </label>
              <Input
                value={form.last_name}
                onChange={set('last_name')}
                placeholder="Smith"
                aria-invalid={!!errors.last_name}
              />
              {errors.last_name && (
                <p className="mt-0.5 text-xs text-red-500">{errors.last_name}</p>
              )}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-slate-700">
              Email <span className="text-red-500">*</span>
            </label>
            <Input
              type="email"
              value={form.email}
              onChange={set('email')}
              placeholder="jane@example.com"
              aria-invalid={!!errors.email}
            />
            {errors.email && <p className="mt-0.5 text-xs text-red-500">{errors.email}</p>}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Phone</label>
              <Input
                type="tel"
                value={form.phone ?? ''}
                onChange={set('phone')}
                placeholder="+1 555 000 0000"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Stage</label>
              <select
                value={form.stage}
                onChange={set('stage')}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-indigo-500"
              >
                <option value="lead">Lead</option>
                <option value="prospect">Prospect</option>
                <option value="customer">Customer</option>
                <option value="churned">Churned</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Title</label>
              <Input
                value={form.title ?? ''}
                onChange={set('title')}
                placeholder="VP of Sales"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-slate-700">Department</label>
              <Input
                value={form.department ?? ''}
                onChange={set('department')}
                placeholder="Sales"
              />
            </div>
          </div>

          <CustomFieldFormSection
            fields={customFields}
            values={customFieldValues}
            onChange={setCustomFieldValues}
          />

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createContact.isPending}>
              {createContact.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
              Create Contact
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
