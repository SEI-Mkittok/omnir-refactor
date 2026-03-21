import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User, Org } from '@/api/types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  activeOrg: Org | null

  setUser: (user: User) => void
  setActiveOrg: (org: Org | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      activeOrg: null,

      setUser: (user) => set({ user, isAuthenticated: true }),

      setActiveOrg: (org) => set({ activeOrg: org }),

      logout: () =>
        set({
          user: null,
          isAuthenticated: false,
          activeOrg: null,
        }),
    }),
    {
      name: 'omnir-auth',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        activeOrg: state.activeOrg,
      }),
    }
  )
)
