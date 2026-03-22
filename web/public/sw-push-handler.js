// PraestOS — Service Worker: Push Notifications + Background Sync
// Imported via workbox importScripts from the generated SW.

// ── Push Notifications ────────────────────────────────────────────────────────

self.addEventListener('push', (event) => {
  let data = {}
  try {
    data = event.data?.json() ?? {}
  } catch {
    data = { title: 'PraestOS', body: event.data?.text() ?? '' }
  }

  const title = data.title || 'PraestOS'
  const options = {
    body: data.body || '',
    icon: '/icons/pwa-192x192.png',
    badge: '/icons/pwa-192x192.png',
    tag: data.tag || 'praestos-notification',
    renotify: true,
    data: { url: data.url || '/dashboard' },
  }

  event.waitUntil(self.registration.showNotification(title, options))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = event.notification.data?.url || '/dashboard'

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      // Focus an existing window if one is open at this origin
      for (const client of windowClients) {
        if (client.url.startsWith(self.location.origin) && 'focus' in client) {
          client.postMessage({ type: 'NAVIGATE', url })
          return client.focus()
        }
      }
      return clients.openWindow(url)
    })
  )
})

// ── Background Sync — Mutation Replay ────────────────────────────────────────
// When the browser fires a sync event (network restored after queued mutations),
// notify all open windows to replay their IndexedDB mutation queue.

self.addEventListener('sync', (event) => {
  if (event.tag === 'omnir-mutations') {
    event.waitUntil(notifyClientsToReplay())
  }
})

async function notifyClientsToReplay() {
  const windowClients = await clients.matchAll({
    type: 'window',
    includeUncontrolled: true,
  })
  for (const client of windowClients) {
    client.postMessage({ type: 'REPLAY_MUTATIONS' })
  }
}
