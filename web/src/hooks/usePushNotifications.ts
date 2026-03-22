// Manages Web Push subscription: request permission, subscribe via VAPID,
// and register the subscription with the backend.

import { useEffect, useRef } from 'react'
import { pushApi } from '@/api/push'

// VAPID public key is injected at build time.
// Falls back to empty string so the hook is a no-op in environments where it
// isn't configured, rather than crashing.
const VAPID_PUBLIC_KEY = import.meta.env.VITE_VAPID_PUBLIC_KEY ?? ''

function urlBase64ToUint8Array(base64String: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = atob(base64)
  const buffer = new ArrayBuffer(rawData.length)
  const view = new Uint8Array(buffer)
  for (let i = 0; i < rawData.length; i++) {
    view[i] = rawData.charCodeAt(i)
  }
  return view
}

export function usePushNotifications() {
  const subscribedRef = useRef(false)

  useEffect(() => {
    if (subscribedRef.current) return
    if (!VAPID_PUBLIC_KEY) return
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) return

    async function subscribe() {
      try {
        const registration = await navigator.serviceWorker.ready

        // Check existing subscription first to avoid duplicate POSTs
        const existing = await registration.pushManager.getSubscription()
        if (existing) {
          subscribedRef.current = true
          return
        }

        // Request notification permission
        const permission = await Notification.requestPermission()
        if (permission !== 'granted') return

        // Create new push subscription
        const subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
        })

        // Register with backend
        await pushApi.subscribe(subscription.toJSON())
        subscribedRef.current = true
      } catch {
        // Non-fatal: push is an enhancement, not a critical path
      }
    }

    subscribe()
  }, [])
}
