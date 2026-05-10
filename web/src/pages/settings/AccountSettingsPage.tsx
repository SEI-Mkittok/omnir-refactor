import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { User, Shield, Check, AlertCircle, SlidersHorizontal } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import { preferencesApi, type UserPreferences } from '@/api/preferences'
import { currenciesApi } from '@/api/currencies'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

function toPreferenceForm(prefs?: UserPreferences) {
  return {
    default_currency: prefs?.default_currency ?? '',
    number_format: prefs?.number_format ?? '',
    default_record_view: prefs?.default_record_view ?? '',
    landing_page: prefs?.landing_page ?? '',
    address_line1: prefs?.address_line1 ?? '',
    city: prefs?.city ?? '',
    state: prefs?.state ?? '',
    postal_code: prefs?.postal_code ?? '',
    country: prefs?.country ?? '',
  }
}

export function AccountSettingsPage() {
  const qc = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)

  const [name, setName] = useState(user?.name ?? '')
  const [profileToast, setProfileToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [prefsToast, setPrefsToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [prefForm, setPrefForm] = useState(toPreferenceForm())

  const preferencesQuery = useQuery({
    queryKey: ['settings', 'preferences'],
    queryFn: preferencesApi.get,
  })
  const currenciesQuery = useQuery({
    queryKey: ['settings', 'currencies'],
    queryFn: currenciesApi.get,
  })

  useEffect(() => {
    if (!preferencesQuery.data) return
    setPrefForm(toPreferenceForm(preferencesQuery.data))
    setPrefsToast(null)
  }, [preferencesQuery.data])

  const updateNameMutation = useMutation({
    mutationFn: async () => {
      if (!user) return null
      return usersApi.update(user.id, { name: name.trim() })
    },
    onSuccess: (updated) => {
      if (!updated) return
      setUser(updated)
      setProfileToast({ type: 'success', text: 'Profile updated.' })
    },
    onError: () => setProfileToast({ type: 'error', text: 'Failed to save changes. Please try again.' }),
  })

  const updatePreferencesMutation = useMutation({
    mutationFn: () => preferencesApi.update(prefForm),
    onSuccess: async () => {
      setPrefsToast({ type: 'success', text: 'Preferences saved.' })
      await qc.invalidateQueries({ queryKey: ['settings', 'preferences'] })
    },
    onError: () => setPrefsToast({ type: 'error', text: 'Unable to save preferences.' }),
  })

  const roleLabel: Record<string, string> = {
    super_admin: 'Super Admin',
    admin: 'Admin',
    agent: 'Agent',
    client: 'Client',
  }

  const currencyOptions = useMemo(
    () => (currenciesQuery.data?.currencies ?? []).map((row) => row.code),
    [currenciesQuery.data?.currencies]
  )

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">My Account</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Update your personal information and profile defaults.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start gap-4 pb-5 border-b border-slate-100">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
            <User className="h-5 w-5 text-slate-600" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-slate-900">Profile</h2>
            <p className="mt-0.5 text-sm text-slate-500">Your name, email, and role.</p>
          </div>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            if (!user || name.trim() === user.name) return
            setProfileToast(null)
            updateNameMutation.mutate()
          }}
          className="mt-5 space-y-4 max-w-md"
        >
          {profileToast && (
            <div
              className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm ${
                profileToast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
              }`}
            >
              {profileToast.type === 'success' ? <Check className="h-4 w-4 shrink-0" /> : <AlertCircle className="h-4 w-4 shrink-0" />}
              {profileToast.text}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1" htmlFor="name">
              Full name
            </label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Email address</label>
            <Input value={user?.email ?? ''} disabled className="bg-slate-50 text-slate-500 cursor-not-allowed" />
            <p className="mt-1 text-xs text-slate-400">Contact your administrator to change your email.</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Role</label>
            <Input
              value={user?.role ? (roleLabel[user.role] ?? user.role) : ''}
              disabled
              className="bg-slate-50 text-slate-500 cursor-not-allowed"
            />
          </div>

          <div className="pt-1">
            <Button type="submit" disabled={updateNameMutation.isPending || name.trim() === user?.name}>
              {updateNameMutation.isPending ? 'Saving...' : 'Save Profile'}
            </Button>
          </div>
        </form>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start gap-4 pb-5 border-b border-slate-100">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
            <SlidersHorizontal className="h-5 w-5 text-slate-600" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-slate-900">Preferences</h2>
            <p className="mt-0.5 text-sm text-slate-500">Default values used in forms and list experiences.</p>
          </div>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            setPrefsToast(null)
            updatePreferencesMutation.mutate()
          }}
          className="mt-5 grid gap-4 md:grid-cols-2"
        >
          {prefsToast && (
            <div
              className={`md:col-span-2 flex items-center gap-2 rounded-md px-3 py-2 text-sm ${
                prefsToast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
              }`}
            >
              {prefsToast.type === 'success' ? <Check className="h-4 w-4 shrink-0" /> : <AlertCircle className="h-4 w-4 shrink-0" />}
              {prefsToast.text}
            </div>
          )}

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Default currency</label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={prefForm.default_currency}
              onChange={(e) => setPrefForm((current) => ({ ...current, default_currency: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            >
              <option value="">Use org default</option>
              {currencyOptions.map((code) => (
                <option key={code} value={code}>
                  {code}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Number format</label>
            <Input
              value={prefForm.number_format}
              placeholder="#,###.##"
              onChange={(e) => setPrefForm((current) => ({ ...current, number_format: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Default record view</label>
            <select
              className="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm"
              value={prefForm.default_record_view}
              onChange={(e) => setPrefForm((current) => ({ ...current, default_record_view: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            >
              <option value="">System default</option>
              <option value="list">List</option>
              <option value="kanban">Kanban</option>
              <option value="detail">Detail</option>
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Landing page</label>
            <Input
              value={prefForm.landing_page}
              placeholder="/dashboard"
              onChange={(e) => setPrefForm((current) => ({ ...current, landing_page: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div className="md:col-span-2">
            <label className="mb-1 block text-sm font-medium text-slate-700">Address line</label>
            <Input
              value={prefForm.address_line1}
              placeholder="Address"
              onChange={(e) => setPrefForm((current) => ({ ...current, address_line1: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">City</label>
            <Input
              value={prefForm.city}
              onChange={(e) => setPrefForm((current) => ({ ...current, city: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">State</label>
            <Input
              value={prefForm.state}
              onChange={(e) => setPrefForm((current) => ({ ...current, state: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Postal code</label>
            <Input
              value={prefForm.postal_code}
              onChange={(e) => setPrefForm((current) => ({ ...current, postal_code: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Country</label>
            <Input
              value={prefForm.country}
              onChange={(e) => setPrefForm((current) => ({ ...current, country: e.target.value }))}
              disabled={preferencesQuery.isLoading || updatePreferencesMutation.isPending}
            />
          </div>

          <div className="md:col-span-2 flex justify-end">
            <Button type="submit" disabled={updatePreferencesMutation.isPending}>
              {updatePreferencesMutation.isPending ? 'Saving...' : 'Save Preferences'}
            </Button>
          </div>
        </form>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
              <Shield className="h-5 w-5 text-slate-600" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-slate-900">Security</h2>
              <p className="mt-0.5 text-sm text-slate-500">
                Two-factor authentication and account security settings.
              </p>
            </div>
          </div>
          <Link
            to="/settings/security"
            className="shrink-0 rounded-md border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
          >
            Manage
          </Link>
        </div>
      </div>
    </div>
  )
}
