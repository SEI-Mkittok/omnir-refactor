import { create } from 'zustand'
import type { TicketStatus, TicketPriority } from '@/api/types'

interface TicketFilterState {
  status: TicketStatus | ''
  priority: TicketPriority | ''
  assignedTo: string
  search: string
  page: number

  setStatus: (status: TicketStatus | '') => void
  setPriority: (priority: TicketPriority | '') => void
  setAssignedTo: (assignedTo: string) => void
  setSearch: (search: string) => void
  setPage: (page: number) => void
  reset: () => void
}

const defaultState = {
  status: '' as TicketStatus | '',
  priority: '' as TicketPriority | '',
  assignedTo: '',
  search: '',
  page: 1,
}

export const useTicketFilterStore = create<TicketFilterState>((set) => ({
  ...defaultState,

  setStatus: (status) => set({ status, page: 1 }),
  setPriority: (priority) => set({ priority, page: 1 }),
  setAssignedTo: (assignedTo) => set({ assignedTo, page: 1 }),
  setSearch: (search) => set({ search, page: 1 }),
  setPage: (page) => set({ page }),
  reset: () => set(defaultState),
}))
