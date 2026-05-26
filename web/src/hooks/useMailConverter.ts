import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { mailConverterApi } from '@/api/mailConverter'
import type { MailConverterRule } from '@/api/types'

const keys = {
  all: ['mail-converter'] as const,
  rules: (params?: unknown) => [...keys.all, 'rules', params] as const,
  runs: (id: string) => [...keys.all, 'runs', id] as const,
}

export function useMailConverterRules(params?: { page?: number; limit?: number; status?: string }) {
  return useQuery({
    queryKey: keys.rules(params),
    queryFn: () => mailConverterApi.listRules(params),
    staleTime: 30_000,
  })
}

export function useCreateMailConverterRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<MailConverterRule>) => mailConverterApi.createRule(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useUpdateMailConverterRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<MailConverterRule> }) => mailConverterApi.updateRule(id, payload),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: keys.all })
      qc.invalidateQueries({ queryKey: keys.runs(id) })
    },
  })
}

export function useDeleteMailConverterRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => mailConverterApi.deleteRule(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function usePreviewMailConverterRule() {
  return useMutation({
    mutationFn: (id: string) => mailConverterApi.previewRule(id),
  })
}

export function useScanMailConverterRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => mailConverterApi.scanRule(id),
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: keys.all })
      qc.invalidateQueries({ queryKey: keys.runs(id) })
    },
  })
}

export function useMailConverterRuns(id: string) {
  return useQuery({
    queryKey: keys.runs(id),
    queryFn: () => mailConverterApi.runs(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}
