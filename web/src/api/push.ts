import { apiClient } from './client'

export const pushApi = {
  subscribe: async (subscription: PushSubscriptionJSON): Promise<void> => {
    await apiClient.post('/push/subscribe', subscription)
  },

  unsubscribe: async (): Promise<void> => {
    await apiClient.delete('/push/subscribe')
  },
}
