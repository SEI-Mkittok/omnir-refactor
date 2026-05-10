import { useEffect, useState, type FormEvent } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  useOutgoingServerSettings,
  useUpdateOutgoingServerSettings,
} from '@/hooks/useAdminSettings'

export function OutgoingServerSettingsPage() {
  const { data, isLoading, isError } = useOutgoingServerSettings()
  const updateOutgoing = useUpdateOutgoingServerSettings()

  const [smtpHost, setSMTPHost] = useState('')
  const [smtpPort, setSMTPPort] = useState('587')
  const [smtpUsername, setSMTPUsername] = useState('')
  const [smtpFromEmail, setSMTPFromEmail] = useState('')
  const [smtpFromName, setSMTPFromName] = useState('')
  const [smtpSecurity, setSMTPSecurity] = useState('starttls')
  const [smtpAuthType, setSMTPAuthType] = useState('password')
  const [smtpPassword, setSMTPPassword] = useState('')
  const [clearPassword, setClearPassword] = useState(false)
  const [status, setStatus] = useState('')

  useEffect(() => {
    if (!data) return
    setSMTPHost(data.smtp_host ?? '')
    setSMTPPort(String(data.smtp_port ?? 587))
    setSMTPUsername(data.smtp_username ?? '')
    setSMTPFromEmail(data.smtp_from_email ?? '')
    setSMTPFromName(data.smtp_from_name ?? '')
    setSMTPSecurity(data.smtp_security ?? 'starttls')
    setSMTPAuthType(data.smtp_auth_type ?? 'password')
    setClearPassword(false)
    setSMTPPassword('')
    setStatus('')
  }, [data])

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setStatus('')

    const parsedPort = Number(smtpPort)
    if (!Number.isInteger(parsedPort) || parsedPort < 1 || parsedPort > 65535) {
      setStatus('SMTP port must be between 1 and 65535.')
      return
    }

    updateOutgoing.mutate(
      {
        smtp_host: smtpHost.trim(),
        smtp_port: parsedPort,
        smtp_username: smtpUsername.trim(),
        smtp_from_email: smtpFromEmail.trim(),
        smtp_from_name: smtpFromName.trim(),
        smtp_security: smtpSecurity.trim().toLowerCase(),
        smtp_auth_type: smtpAuthType.trim().toLowerCase(),
        smtp_password: smtpPassword.trim() || undefined,
        clear_password: clearPassword,
      },
      {
        onSuccess: () => {
          setSMTPPassword('')
          setClearPassword(false)
          setStatus('Outgoing server settings saved.')
        },
        onError: () => setStatus('Unable to save outgoing server settings.'),
      }
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">Admin Settings</p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Outgoing Server</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Configure organization SMTP delivery and sender defaults.
        </p>
      </div>

      {isError && <p className="text-sm text-red-600">Failed to load outgoing server settings.</p>}

      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm">
        <div className="grid gap-3 md:grid-cols-2">
          <Input value={smtpHost} onChange={(e) => setSMTPHost(e.target.value)} placeholder="SMTP host" />
          <Input
            type="number"
            min={1}
            max={65535}
            value={smtpPort}
            onChange={(e) => setSMTPPort(e.target.value)}
            placeholder="SMTP port"
          />
          <Input
            value={smtpUsername}
            onChange={(e) => setSMTPUsername(e.target.value)}
            placeholder="SMTP username"
          />
          <Input
            value={smtpFromEmail}
            onChange={(e) => setSMTPFromEmail(e.target.value)}
            placeholder="From email"
          />
          <Input
            value={smtpFromName}
            onChange={(e) => setSMTPFromName(e.target.value)}
            placeholder="From name"
          />
          <Input
            value={smtpSecurity}
            onChange={(e) => setSMTPSecurity(e.target.value)}
            placeholder="Security (none/starttls/tls)"
          />
          <Input
            value={smtpAuthType}
            onChange={(e) => setSMTPAuthType(e.target.value)}
            placeholder="Auth type (password/oauth2)"
          />
          <Input
            type="password"
            value={smtpPassword}
            onChange={(e) => setSMTPPassword(e.target.value)}
            placeholder={(data?.smtp_password_set ?? false) ? 'Password stored (enter to rotate)' : 'SMTP password'}
          />
        </div>

        <label className="flex items-center gap-2 text-sm text-[#374151]">
          <input type="checkbox" checked={clearPassword} onChange={(e) => setClearPassword(e.target.checked)} />
          Clear saved password
        </label>

        <div className="flex items-center justify-between gap-3">
          <p className={updateOutgoing.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>{status}</p>
          <Button type="submit" disabled={isLoading || updateOutgoing.isPending}>
            <Save className="h-4 w-4" />
            {updateOutgoing.isPending ? 'Saving...' : 'Save Outgoing Server'}
          </Button>
        </div>
      </form>
    </div>
  )
}
