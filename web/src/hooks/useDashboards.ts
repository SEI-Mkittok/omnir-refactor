import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listDashboards,
  getDashboard,
  createDashboard,
  updateDashboard,
  deleteDashboard,
  runDashboard,
  listSchedules,
  createSchedule,
  updateSchedule,
  deleteSchedule,
} from '@/api/dashboards'
import type { Widget } from '@/api/dashboards'

export const dashboardKeys = {
  all: ['dashboards'] as const,
  detail: (id: string) => ['dashboards', id] as const,
  run: (id: string) => ['dashboards', id, 'run'] as const,
  schedules: ['dashboards', 'schedules'] as const,
}

// ── Dashboards ───────────────────────────────────────────────────────────────

export function useDashboards() {
  return useQuery({
    queryKey: dashboardKeys.all,
    queryFn: listDashboards,
    staleTime: 60 * 1000,
  })
}

export function useDashboard(id: string | null) {
  return useQuery({
    queryKey: dashboardKeys.detail(id ?? ''),
    queryFn: () => getDashboard(id!),
    enabled: !!id,
    staleTime: 60 * 1000,
  })
}

export function useRunDashboard(id: string | null) {
  return useQuery({
    queryKey: dashboardKeys.run(id ?? ''),
    queryFn: () => runDashboard(id!),
    enabled: !!id,
    staleTime: 30 * 1000,
  })
}

export function useCreateDashboard() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { name: string; widgets: Widget[] }) => createDashboard(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: dashboardKeys.all }),
  })
}

export function useUpdateDashboard() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: { name?: string; widgets?: Widget[] } }) =>
      updateDashboard(id, patch),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all })
      qc.invalidateQueries({ queryKey: dashboardKeys.detail(id) })
      qc.invalidateQueries({ queryKey: dashboardKeys.run(id) })
    },
  })
}

export function useDeleteDashboard() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteDashboard(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: dashboardKeys.all }),
  })
}

// ── Schedules ────────────────────────────────────────────────────────────────

export function useSchedules() {
  return useQuery({
    queryKey: dashboardKeys.schedules,
    queryFn: listSchedules,
    staleTime: 60 * 1000,
  })
}

export function useCreateSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: createSchedule,
    onSuccess: () => qc.invalidateQueries({ queryKey: dashboardKeys.schedules }),
  })
}

export function useUpdateSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: { dashboard_id?: string; schedule?: string; recipients?: string[] } }) =>
      updateSchedule(id, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: dashboardKeys.schedules }),
  })
}

export function useDeleteSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteSchedule(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: dashboardKeys.schedules }),
  })
}
