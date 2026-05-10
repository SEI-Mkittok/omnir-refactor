import { useEffect, useState, type FormEvent } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useCompanySettings, useUpdateCompanySettings } from '@/hooks/useAdminSettings'
import type { CompanySettings } from '@/api/adminSettings'

const EMPTY_FORM: Required<CompanySettings> = {
  company_name: '',
  company_logo_url: '',
  company_website: '',
  company_email: '',
  company_phone: '',
  company_address_line1: '',
  company_address_line2: '',
  company_city: '',
  company_state: '',
  company_postal_code: '',
  company_country: '',
}

function normalizePayload(form: Required<CompanySettings>): CompanySettings {
  return Object.entries(form).reduce<CompanySettings>((acc, [key, value]) => {
    acc[key as keyof CompanySettings] = value.trim()
    return acc
  }, {})
}

export function CompanyProfileSettingsPage() {
  const { data, isLoading, isError } = useCompanySettings()
  const updateCompany = useUpdateCompanySettings()
  const [form, setForm] = useState<Required<CompanySettings>>(EMPTY_FORM)
  const [status, setStatus] = useState('')

  useEffect(() => {
    if (!data) return
    setForm({
      company_name: data.company_name ?? '',
      company_logo_url: data.company_logo_url ?? '',
      company_website: data.company_website ?? '',
      company_email: data.company_email ?? '',
      company_phone: data.company_phone ?? '',
      company_address_line1: data.company_address_line1 ?? '',
      company_address_line2: data.company_address_line2 ?? '',
      company_city: data.company_city ?? '',
      company_state: data.company_state ?? '',
      company_postal_code: data.company_postal_code ?? '',
      company_country: data.company_country ?? '',
    })
    setStatus('')
  }, [data])

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setStatus('')
    updateCompany.mutate(normalizePayload(form), {
      onSuccess: () => setStatus('Company profile saved.'),
      onError: () => setStatus('Unable to save company profile.'),
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">Admin Settings</p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Company Profile</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Manage tenant-safe company details used across customer-facing surfaces.
        </p>
      </div>

      {isError && <p className="text-sm text-red-600">Failed to load company profile settings.</p>}

      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm">
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            placeholder="Company name"
            value={form.company_name}
            onChange={(e) => setForm((s) => ({ ...s, company_name: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Company website"
            value={form.company_website}
            onChange={(e) => setForm((s) => ({ ...s, company_website: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Company email"
            value={form.company_email}
            onChange={(e) => setForm((s) => ({ ...s, company_email: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Phone"
            value={form.company_phone}
            onChange={(e) => setForm((s) => ({ ...s, company_phone: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Logo URL"
            value={form.company_logo_url}
            onChange={(e) => setForm((s) => ({ ...s, company_logo_url: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Address line 1"
            value={form.company_address_line1}
            onChange={(e) => setForm((s) => ({ ...s, company_address_line1: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Address line 2"
            value={form.company_address_line2}
            onChange={(e) => setForm((s) => ({ ...s, company_address_line2: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="City"
            value={form.company_city}
            onChange={(e) => setForm((s) => ({ ...s, company_city: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="State"
            value={form.company_state}
            onChange={(e) => setForm((s) => ({ ...s, company_state: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Postal code"
            value={form.company_postal_code}
            onChange={(e) => setForm((s) => ({ ...s, company_postal_code: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
          <Input
            placeholder="Country"
            value={form.company_country}
            onChange={(e) => setForm((s) => ({ ...s, company_country: e.target.value }))}
            disabled={isLoading || updateCompany.isPending}
          />
        </div>

        <div className="flex items-center justify-between gap-3">
          <p className={updateCompany.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>{status}</p>
          <Button type="submit" disabled={isLoading || updateCompany.isPending}>
            <Save className="h-4 w-4" />
            {updateCompany.isPending ? 'Saving...' : 'Save Company Profile'}
          </Button>
        </div>
      </form>
    </div>
  )
}
