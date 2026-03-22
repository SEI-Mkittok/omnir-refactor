// Replays queued offline mutations when connectivity is restored.
// Two triggers:
//   1. navigator.onLine events (cross-browser fallback)
//   2. Service Worker REPLAY_MUTATIONS message (Background Sync)

import { useEffect } from 'react'
import axios from 'axios'
import {
  getMutationQueue,
  removeMutationFromQueue,
  type QueuedMutation,
} from '@/lib/mutationQueue'

async function replayMutation(mutation: QueuedMutation): Promise<void> {
  await axios({
    method: mutation.method,
    url: mutation.url,
    data: mutation.data,
    withCredentials: true,
  })
}

async function replayAll(): Promise<void> {
  const queue = await getMutationQueue()
  if (queue.length === 0) return

  for (const mutation of queue) {
    try {
      await replayMutation(mutation)
      await removeMutationFromQueue(mutation.id)
    } catch {
      // Leave failed items in the queue for the next retry
      break
    }
  }
}

export function useMutationQueue() {
  useEffect(() => {
    // Replay on online event (covers browsers without Background Sync)
    const handleOnline = () => {
      replayAll()
    }
    window.addEventListener('online', handleOnline)

    // Replay on SW Background Sync message
    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === 'REPLAY_MUTATIONS') {
        replayAll()
      }
    }
    navigator.serviceWorker?.addEventListener('message', handleMessage)

    return () => {
      window.removeEventListener('online', handleOnline)
      navigator.serviceWorker?.removeEventListener('message', handleMessage)
    }
  }, [])
}
