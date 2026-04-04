import { create } from 'zustand'
import type { User, Org } from '@/api/types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  activeOrg: Org | null
  isInitializing: boolean

  setUser: (user: User) => void
  setActiveOrg: (org: Org | null) => void
  setInitializing: (v: boolean) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  isAuthenticated: false,
  activeOrg: null,
  isInitializing: true,

  setUser: (user) => set({ user, isAuthenticated: true }),
  setActiveOrg: (org) => set({ activeOrg: org }),
  setInitializing: (v) => set({ isInitializing: v }),
  logout: () => set({ user: null, isAuthenticated: false, activeOrg: null }),
}))
