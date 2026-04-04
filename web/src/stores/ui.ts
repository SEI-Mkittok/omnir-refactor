import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UIState {
  sidebarCollapsed: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (v: boolean) => void
  onboardingOpen: boolean
  setOnboardingOpen: (v: boolean) => void
  onboardingDismissed: boolean
  setOnboardingDismissed: (v: boolean) => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
      setSidebarCollapsed: (v) => set({ sidebarCollapsed: v }),
      onboardingOpen: false,
      setOnboardingOpen: (v) => set({ onboardingOpen: v }),
      onboardingDismissed: false,
      setOnboardingDismissed: (v) => set({ onboardingDismissed: v }),
    }),
    {
      name: 'praestos_ui',
      partialize: (state) => ({ sidebarCollapsed: state.sidebarCollapsed, onboardingDismissed: state.onboardingDismissed }),
    }
  )
)
