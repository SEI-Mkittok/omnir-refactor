import apiClient from './client'
import type { EntityAttachment, EntityAttachmentEntityType } from './types'

const entityPath = (entityType: EntityAttachmentEntityType, entityId: string) =>
  `/${entityType}s/${entityId}/attachments`

export const attachmentsApi = {
  list: async (entityType: EntityAttachmentEntityType, entityId: string): Promise<EntityAttachment[]> => {
    const { data } = await apiClient.get<EntityAttachment[]>(entityPath(entityType, entityId))
    return data
  },

  upload: async (
    entityType: EntityAttachmentEntityType,
    entityId: string,
    file: File,
    onUploadProgress?: (pct: number) => void
  ): Promise<EntityAttachment> => {
    const form = new FormData()
    form.append('file', file)
    const { data } = await apiClient.post<EntityAttachment>(entityPath(entityType, entityId), form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (evt) => {
        if (onUploadProgress && evt.total) {
          onUploadProgress(Math.round((evt.loaded * 100) / evt.total))
        }
      },
    })
    return data
  },

  delete: async (
    entityType: EntityAttachmentEntityType,
    entityId: string,
    attachmentId: string
  ): Promise<void> => {
    await apiClient.delete(`${entityPath(entityType, entityId)}/${attachmentId}`)
  },
}
