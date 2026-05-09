import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { numberingApi, type UpdateNumberingSettingsRequest } from '@/api/numbering'

export const numberingKeys = {
  all: ['numbering-settings'] as const,
  detail: () => [...numberingKeys.all, 'detail'] as const,
}

export function useNumberingSettings() {
  return useQuery({
    queryKey: numberingKeys.detail(),
    queryFn: numberingApi.get,
    staleTime: 60_000,
  })
}

export function useUpdateNumberingSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: UpdateNumberingSettingsRequest) => numberingApi.update(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: numberingKeys.detail() })
    },
  })
}
