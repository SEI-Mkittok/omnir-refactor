import { useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useCreateAccount } from '@/hooks/useAccounts'
import type { Account, CreateAccountRequest } from '@/api/types'

const INDUSTRY_OPTIONS = [
  'Technology', 'Finance', 'Healthcare', 'Retail', 'Manufacturing',
  'Education', 'Real Estate', 'Media', 'Consulting', 'Other',
]

const SIZE_OPTIONS = [
  '1-10', '11-50', '51-200', '201-500', '501-1000', '1000+',
]

interface AccountFormProps {
  onClose: () => void
  initialValues?: Partial<CreateAccountRequest>
  onCreated?: (account: Account) => void
}

export function AccountForm({ onClose, initialValues, onCreated }: AccountFormProps) {
  const { mutateAsync: createAccount, isPending } = useCreateAccount()
  const [form, setForm] = useState<CreateAccountRequest>({
    name: initialValues?.name ?? '',
    domain: initialValues?.domain,
    industry: initialValues?.industry,
    size: initialValues?.size,
    phone: initialValues?.phone,
    address: initialValues?.address,
    website: initialValues?.website,
    owner_id: initialValues?.owner_id,
  })
  const [nameError, setNameError] = useState('')

  const set = (field: keyof CreateAccountRequest) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value || undefined }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name.trim()) {
      setNameError('Required')
      return
    }
    setNameError('')
    const account = await createAccount({ ...form, name: form.name.trim() })
    onCreated?.(account)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-xl bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-slate-200 px-4 sm:px-6 py-4">
          <h2 className="text-lg font-semibold text-slate-900">New Account</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 p-4 sm:p-6">
          <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="account-name">
              Company name <span className="text-red-500">*</span>
            </label>
            <input
              id="account-name"
              type="text"
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="Acme Corp"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
            />
            {nameError && <p className="text-xs text-red-600">{nameError}</p>}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="account-industry">
                Industry
              </label>
              <select
                id="account-industry"
                value={form.industry ?? ''}
                onChange={set('industry')}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              >
                <option value="">Select…</option>
                {INDUSTRY_OPTIONS.map((o) => (
                  <option key={o} value={o}>{o}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="account-size">
                Company size
              </label>
              <select
                id="account-size"
                value={form.size ?? ''}
                onChange={set('size')}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              >
                <option value="">Select…</option>
                {SIZE_OPTIONS.map((o) => (
                  <option key={o} value={o}>{o}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="account-domain">
                Domain
              </label>
              <input
                id="account-domain"
                type="text"
                value={form.domain ?? ''}
                onChange={set('domain')}
                placeholder="acme.com"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700" htmlFor="account-phone">
                Phone
              </label>
              <input
                id="account-phone"
                type="tel"
                value={form.phone ?? ''}
                onChange={set('phone')}
                placeholder="+1 555 000 0000"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="block text-sm font-medium text-slate-700" htmlFor="account-website">
              Website
            </label>
            <input
              id="account-website"
              type="url"
              value={form.website ?? ''}
              onChange={set('website')}
              placeholder="https://acme.com"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-[var(--border-focus)] focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create Account'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
