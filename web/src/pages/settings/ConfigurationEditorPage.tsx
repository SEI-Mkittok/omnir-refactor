import { useEffect, useState, type FormEvent } from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  useConfigEditorSettings,
  useUpdateConfigEditorSettings,
} from '@/hooks/useAdminSettings'

export function ConfigurationEditorPage() {
  const { data, isLoading, isError } = useConfigEditorSettings()
  const updateConfig = useUpdateConfigEditorSettings()

  const [supportEmail, setSupportEmail] = useState('')
  const [uploadMaxMB, setUploadMaxMB] = useState('25')
  const [defaultPageSize, setDefaultPageSize] = useState('25')
  const [listPreviewChars, setListPreviewChars] = useState('120')
  const [status, setStatus] = useState('')

  useEffect(() => {
    if (!data) return
    setSupportEmail(data.config_support_email ?? '')
    setUploadMaxMB(String(data.config_upload_max_mb ?? 25))
    setDefaultPageSize(String(data.config_default_page_size ?? 25))
    setListPreviewChars(String(data.config_list_preview_chars ?? 120))
    setStatus('')
  }, [data])

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setStatus('')

    const upload = Number(uploadMaxMB)
    const pageSize = Number(defaultPageSize)
    const preview = Number(listPreviewChars)

    if (!Number.isInteger(upload) || upload < 1 || upload > 1024) {
      setStatus('Upload max must be between 1 and 1024 MB.')
      return
    }
    if (!Number.isInteger(pageSize) || pageSize < 1 || pageSize > 500) {
      setStatus('Default page size must be between 1 and 500.')
      return
    }
    if (!Number.isInteger(preview) || preview < 20 || preview > 2000) {
      setStatus('List preview chars must be between 20 and 2000.')
      return
    }

    updateConfig.mutate(
      {
        config_support_email: supportEmail.trim(),
        config_upload_max_mb: upload,
        config_default_page_size: pageSize,
        config_list_preview_chars: preview,
      },
      {
        onSuccess: () => setStatus('Configuration editor settings saved.'),
        onError: () => setStatus('Unable to save configuration editor settings.'),
      }
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">Admin Settings</p>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Configuration Editor</h1>
        <p className="mt-1 max-w-2xl text-sm text-[#6B7280]">
          Tune shared runtime defaults used by admin and module interfaces.
        </p>
      </div>

      {isError && <p className="text-sm text-red-600">Failed to load configuration editor settings.</p>}

      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm">
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            placeholder="Support email"
            value={supportEmail}
            onChange={(e) => setSupportEmail(e.target.value)}
            disabled={isLoading || updateConfig.isPending}
          />
          <Input
            placeholder="Upload max (MB)"
            type="number"
            min={1}
            max={1024}
            value={uploadMaxMB}
            onChange={(e) => setUploadMaxMB(e.target.value)}
            disabled={isLoading || updateConfig.isPending}
          />
          <Input
            placeholder="Default page size"
            type="number"
            min={1}
            max={500}
            value={defaultPageSize}
            onChange={(e) => setDefaultPageSize(e.target.value)}
            disabled={isLoading || updateConfig.isPending}
          />
          <Input
            placeholder="List preview chars"
            type="number"
            min={20}
            max={2000}
            value={listPreviewChars}
            onChange={(e) => setListPreviewChars(e.target.value)}
            disabled={isLoading || updateConfig.isPending}
          />
        </div>

        <div className="flex items-center justify-between gap-3">
          <p className={updateConfig.isError ? 'text-sm text-red-600' : 'text-sm text-green-700'}>{status}</p>
          <Button type="submit" disabled={isLoading || updateConfig.isPending}>
            <Save className="h-4 w-4" />
            {updateConfig.isPending ? 'Saving...' : 'Save Configuration'}
          </Button>
        </div>
      </form>
    </div>
  )
}
