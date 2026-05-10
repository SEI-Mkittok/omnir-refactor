import { useEffect, useState, type FormEvent } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { usePortalSettings, useUpdatePortalSettings } from '@/hooks/useAdminSettings'

function splitCSV(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

export function PortalConfigurationPage() {
  const { data, isLoading, isError } = usePortalSettings()
  const updatePortal = useUpdatePortalSettings()

  const [portalEnabled, setPortalEnabled] = useState(false)
  const [displayName, setDisplayName] = useState('')
  const [announcement, setAnnouncement] = useState('')
  const [menu, setMenu] = useState('')
  const [shortcuts, setShortcuts] = useState('')
  const [recentLimit, setRecentLimit] = useState('5')
  const [status, setStatus] = useState('')

  useEffect(() => {
    if (!data) return
    setPortalEnabled(data.portal_enabled)
    setDisplayName(data.portal_display_name ?? '')
    setAnnouncement(data.portal_announcement ?? '')
    setMenu((data.portal_menu ?? []).join(', '))
    setShortcuts((data.portal_shortcuts ?? []).join(', '))
    setRecentLimit(String(data.portal_recent_widget_limit ?? 5))
    setStatus('')
  }, [data])

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setStatus('')

    const limit = Number(recentLimit)
    if (!Number.isInteger(limit) || limit < 0 || limit > 50) {
      setStatus('Recent activity limit must be between 0 and 50.')
      return
    }

    updatePortal.mutate(
      {
        portal_enabled: portalEnabled,
        portal_display_name: displayName.trim(),
        portal_announcement: announcement.trim(),
        portal_menu: splitCSV(menu),
        portal_shortcuts: splitCSV(shortcuts),
        portal_recent_widget_limit: limit,
      },
      {
        onSuccess: () => setStatus('Portal configuration saved.'),
        onError: () => setStatus('Unable to save portal configuration.'),
      }
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">Admin Settings</p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Portal Configuration</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Configure customer portal visibility, shortcuts, and feed behavior.
        </p>
      </div>

      {isError && <p className="text-sm text-red-600">Failed to load portal settings.</p>}

      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm">
        <label className="flex items-center gap-2 text-sm text-[#1A1D23]">
          <input
            type="checkbox"
            checked={portalEnabled}
            onChange={(e) => setPortalEnabled(e.target.checked)}
            disabled={isLoading || updatePortal.isPending}
          />
          Enable customer portal
        </label>

        <div className="grid gap-3 md:grid-cols-2">
          <Input
            placeholder="Portal display name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            disabled={isLoading || updatePortal.isPending}
          />
          <Input
            placeholder="Recent widget limit"
            type="number"
            min={0}
            max={50}
            value={recentLimit}
            onChange={(e) => setRecentLimit(e.target.value)}
            disabled={isLoading || updatePortal.isPending}
          />
        </div>

        <label className="block text-sm text-[#374151]">
          Announcement
          <textarea
            className="mt-1 w-full rounded-md border border-[#D1D5DB] px-3 py-2 text-sm"
            rows={3}
            value={announcement}
            onChange={(e) => setAnnouncement(e.target.value)}
            disabled={isLoading || updatePortal.isPending}
          />
        </label>

        <div className="grid gap-3 md:grid-cols-2">
          <Input
            placeholder="Portal menu items (comma separated)"
            value={menu}
            onChange={(e) => setMenu(e.target.value)}
            disabled={isLoading || updatePortal.isPending}
          />
          <Input
            placeholder="Portal shortcuts (comma separated)"
            value={shortcuts}
            onChange={(e) => setShortcuts(e.target.value)}
            disabled={isLoading || updatePortal.isPending}
          />
        </div>

        <div className="flex items-center justify-between gap-3">
          <p className={updatePortal.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>{status}</p>
          <Button type="submit" disabled={isLoading || updatePortal.isPending}>
            <Save className="h-4 w-4" />
            {updatePortal.isPending ? 'Saving...' : 'Save Portal Settings'}
          </Button>
        </div>
      </form>
    </div>
  )
}
