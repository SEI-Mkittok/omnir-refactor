// Offline mutation queue — stores failed write operations in IndexedDB
// for later replay when connectivity is restored.

import { get, set } from 'idb-keyval'

const QUEUE_KEY = 'omnir-mutation-queue'

export interface QueuedMutation {
  id: string
  method: 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  url: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data?: any
  queuedAt: string
}

export async function enqueueMutation(mutation: Omit<QueuedMutation, 'id' | 'queuedAt'>): Promise<void> {
  const queue = await getMutationQueue()
  const entry: QueuedMutation = {
    ...mutation,
    id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
    queuedAt: new Date().toISOString(),
  }
  await set(QUEUE_KEY, [...queue, entry])
}

export async function getMutationQueue(): Promise<QueuedMutation[]> {
  return (await get<QueuedMutation[]>(QUEUE_KEY)) ?? []
}

export async function clearMutationQueue(): Promise<void> {
  await set(QUEUE_KEY, [])
}

export async function removeMutationFromQueue(id: string): Promise<void> {
  const queue = await getMutationQueue()
  await set(QUEUE_KEY, queue.filter((m) => m.id !== id))
}
