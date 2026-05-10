import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useMenuConfigSettings, useUpdateMenuConfigSettings } from '@/hooks/useAdminSettings'

const MODULES = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'deals', label: 'Pipeline' },
  { key: 'contacts', label: 'Contacts' },
  { key: 'accounts', label: 'Accounts' },
  { key: 'leads', label: 'Leads' },
  { key: 'quotes', label: 'Quotes' },
  { key: 'sequences', label: 'Sequences' },
  { key: 'inbox', label: 'Email Inbox' },
  { key: 'tickets', label: 'Tickets' },
  { key: 'kb', label: 'Knowledge Base' },
  { key: 'reports', label: 'Reports' },
  { key: 'dashboards', label: 'Dashboards' },
  { key: 'automations', label: 'Automations' },
  { key: 'calendar', label: 'Calendar' },
]

function defaultConfig(): Record<string, boolean> {
  return MODULES.reduce<Record<string, boolean>>((acc, mod) => {
    acc[mod.key] = true
    return acc
  }, {})
}

export function MenuConfigurationPage() {
  const { data, isLoading, isError } = useMenuConfigSettings()
  const updateMenu = useUpdateMenuConfigSettings()
  const [menuConfig, setMenuConfig] = useState<Record<string, boolean>>(defaultConfig)
  const [status, setStatus] = useState('')

  useEffect(() => {
    if (!data?.menu_config) return
    setMenuConfig((current) => ({ ...current, ...data.menu_config }))
    setStatus('')
  }, [data])

  const enabledCount = useMemo(
    () => Object.values(menuConfig).filter(Boolean).length,
    [menuConfig]
  )

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setStatus('')
    updateMenu.mutate(
      { menu_config: menuConfig },
      {
        onSuccess: () => setStatus('Main menu configuration saved.'),
        onError: () => setStatus('Unable to save main menu configuration.'),
      }
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">Admin Settings</p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Main Menu Configuration</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Select which CRM modules are enabled in navigation for this organization.
        </p>
      </div>

      {isError && <p className="text-sm text-red-600">Failed to load main menu configuration.</p>}

      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm">
        <p className="text-xs text-[#6B7280]">Enabled modules: {enabledCount}</p>

        <div className="grid gap-2 md:grid-cols-2">
          {MODULES.map((module) => (
            <label key={module.key} className="flex items-center gap-2 rounded-md border border-[#E5E7EB] px-3 py-2 text-sm">
              <input
                type="checkbox"
                checked={menuConfig[module.key] ?? true}
                onChange={(e) =>
                  setMenuConfig((current) => ({
                    ...current,
                    [module.key]: e.target.checked,
                  }))
                }
                disabled={isLoading || updateMenu.isPending}
              />
              {module.label}
            </label>
          ))}
        </div>

        <div className="flex items-center justify-between gap-3">
          <p className={updateMenu.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>{status}</p>
          <Button type="submit" disabled={isLoading || updateMenu.isPending}>
            <Save className="h-4 w-4" />
            {updateMenu.isPending ? 'Saving...' : 'Save Menu Configuration'}
          </Button>
        </div>
      </form>
    </div>
  )
}
