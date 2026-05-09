import { useState } from 'react'
import { Link } from 'react-router-dom'
import { User, Shield, Check, AlertCircle } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

export function AccountSettingsPage() {
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)

  const [name, setName] = useState(user?.name ?? '')
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    if (!user || name.trim() === user.name) return
    setSaving(true)
    setToast(null)
    try {
      const updated = await usersApi.update(user.id, { name: name.trim() })
      setUser(updated)
      setToast({ type: 'success', text: 'Profile updated.' })
    } catch {
      setToast({ type: 'error', text: 'Failed to save changes. Please try again.' })
    } finally {
      setSaving(false)
    }
  }

  const roleLabel: Record<string, string> = {
    super_admin: 'Super Admin',
    admin: 'Admin',
    agent: 'Agent',
    client: 'Client',
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">My Account</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Update your personal information and account details.
        </p>
      </div>

      {/* Profile card */}
      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start gap-4 pb-5 border-b border-slate-100">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
            <User className="h-5 w-5 text-slate-600" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-slate-900">Profile</h2>
            <p className="mt-0.5 text-sm text-slate-500">Your name and email address.</p>
          </div>
        </div>

        <form onSubmit={handleSave} className="mt-5 space-y-4 max-w-md">
          {toast && (
            <div
              className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm ${
                toast.type === 'success'
                  ? 'bg-green-50 text-green-700'
                  : 'bg-red-50 text-red-700'
              }`}
            >
              {toast.type === 'success' ? (
                <Check className="h-4 w-4 shrink-0" />
              ) : (
                <AlertCircle className="h-4 w-4 shrink-0" />
              )}
              {toast.text}
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
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Email address
            </label>
            <Input
              value={user?.email ?? ''}
              disabled
              className="bg-slate-50 text-slate-500 cursor-not-allowed"
            />
            <p className="mt-1 text-xs text-slate-400">
              Contact your administrator to change your email.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Role
            </label>
            <Input
              value={user?.role ? (roleLabel[user.role] ?? user.role) : ''}
              disabled
              className="bg-slate-50 text-slate-500 cursor-not-allowed"
            />
          </div>

          <div className="pt-1">
            <Button
              type="submit"
              disabled={saving || name.trim() === user?.name}
              className="bg-[#1B3A4B] text-white hover:bg-[#16303f]"
            >
              {saving ? 'Saving…' : 'Save changes'}
            </Button>
          </div>
        </form>
      </div>

      {/* Security card */}
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
            Manage →
          </Link>
        </div>
      </div>
    </div>
  )
}
