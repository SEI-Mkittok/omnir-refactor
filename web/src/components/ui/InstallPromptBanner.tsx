import { Download, X } from 'lucide-react'
import { useInstallPrompt } from '@/hooks/useInstallPrompt'

export function InstallPromptBanner() {
  const { canInstall, showPrompt, dismiss } = useInstallPrompt()

  if (!canInstall) return null

  return (
    <div
      className="flex items-center justify-between gap-3 px-4 py-2 text-sm font-medium text-white"
      style={{ background: 'var(--color-primary)' }}
    >
      <div className="flex items-center gap-2">
        <Download className="h-4 w-4 shrink-0" />
        <span>Install PraestOS for faster access</span>
      </div>
      <div className="flex items-center gap-3">
        <button
          onClick={showPrompt}
          className="rounded px-3 py-1 text-xs font-semibold transition-colors"
          style={{ background: 'var(--color-primary-light)', color: 'var(--color-primary)' }}
        >
          Add to Home Screen
        </button>
        <button
          onClick={dismiss}
          className="opacity-70 hover:opacity-100"
          aria-label="Dismiss install banner"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
