// Tracks page visits and captures the browser's beforeinstallprompt event.
// After 3 visits, exposes a `showPrompt` function that triggers the native
// "Add to Home Screen" dialog.

import { useState, useEffect, useCallback } from 'react'

const VISIT_COUNT_KEY = 'omnir-pwa-visit-count'
const DISMISSED_KEY = 'omnir-pwa-install-dismissed'
const THRESHOLD = 3

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

interface UseInstallPromptResult {
  canInstall: boolean
  showPrompt: () => void
  dismiss: () => void
}

export function useInstallPrompt(): UseInstallPromptResult {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [visitThresholdMet, setVisitThresholdMet] = useState(false)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    // Increment visit counter
    const count = parseInt(localStorage.getItem(VISIT_COUNT_KEY) ?? '0', 10) + 1
    localStorage.setItem(VISIT_COUNT_KEY, String(count))
    if (count >= THRESHOLD) setVisitThresholdMet(true)

    // Check if user previously dismissed
    if (localStorage.getItem(DISMISSED_KEY) === '1') setDismissed(true)

    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
    }

    window.addEventListener('beforeinstallprompt', handler)
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const showPrompt = useCallback(() => {
    if (!deferredPrompt) return
    deferredPrompt.prompt()
    deferredPrompt.userChoice.then((choice) => {
      if (choice.outcome === 'accepted') {
        setDeferredPrompt(null)
      }
    })
  }, [deferredPrompt])

  const dismiss = useCallback(() => {
    localStorage.setItem(DISMISSED_KEY, '1')
    setDismissed(true)
  }, [])

  return {
    canInstall: !!deferredPrompt && visitThresholdMet && !dismissed,
    showPrompt,
    dismiss,
  }
}
