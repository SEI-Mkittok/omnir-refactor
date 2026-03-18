import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { attachmentsApi } from '@/api/attachments'
import type { EntityAttachmentEntityType } from '@/api/types'

const attachmentKeys = {
  list: (entityType: EntityAttachmentEntityType, entityId: string) =>
    ['attachments', entityType, entityId] as const,
}

export function useEntityAttachments(entityType: EntityAttachmentEntityType, entityId: string) {
  return useQuery({
    queryKey: attachmentKeys.list(entityType, entityId),
    queryFn: () => attachmentsApi.list(entityType, entityId),
    enabled: !!entityId,
    staleTime: 30_000,
  })
}

export function useUploadAttachment(entityType: EntityAttachmentEntityType, entityId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      file,
      onProgress,
    }: {
      file: File
      onProgress?: (pct: number) => void
    }) => attachmentsApi.upload(entityType, entityId, file, onProgress),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: attachmentKeys.list(entityType, entityId) })
    },
  })
}

export function useDeleteAttachment(entityType: EntityAttachmentEntityType, entityId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (attachmentId: string) =>
      attachmentsApi.delete(entityType, entityId, attachmentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: attachmentKeys.list(entityType, entityId) })
    },
  })
}
