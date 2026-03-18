import { useState } from 'react'
import {
  FileText,
  Image,
  File as FileIcon,
  Download,
  Trash2,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { FileUpload } from '@/components/ui/FileUpload'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { useEntityAttachments, useUploadAttachment, useDeleteAttachment } from '@/hooks/useAttachments'
import type { EntityAttachment, EntityAttachmentEntityType } from '@/api/types'
import { formatDate } from '@/lib/utils'

function formatBytes(bytes?: number): string {
  if (bytes === undefined || bytes === null) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function AttachmentIcon({ contentType }: { contentType: string }) {
  if (contentType.startsWith('image/')) return <Image className="h-4 w-4 text-indigo-500" />
  if (contentType === 'application/pdf') return <FileText className="h-4 w-4 text-red-500" />
  return <FileIcon className="h-4 w-4 text-slate-400" />
}

function AttachmentRow({
  attachment,
  onDelete,
  isDeleting,
}: {
  attachment: EntityAttachment
  onDelete: () => void
  isDeleting: boolean
}) {
  const [previewOpen, setPreviewOpen] = useState(false)
  const isImage = attachment.content_type.startsWith('image/')
  const isPdf = attachment.content_type === 'application/pdf'
  const canPreview = isImage || isPdf

  return (
    <li className="rounded-lg border border-slate-200 bg-white">
      <div className="flex items-center gap-3 px-3 py-2.5">
        <AttachmentIcon contentType={attachment.content_type} />

        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-slate-800">{attachment.filename}</p>
          <p className="text-xs text-slate-400">
            {formatBytes(attachment.size_bytes)}
            {attachment.size_bytes ? ' · ' : ''}
            {formatDate(attachment.created_at)}
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-1">
          {canPreview && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Toggle preview"
              onClick={() => setPreviewOpen((o) => !o)}
            >
              {previewOpen ? (
                <ChevronDown className="h-3.5 w-3.5" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5" />
              )}
            </Button>
          )}

          <a
            href={attachment.url}
            target="_blank"
            rel="noopener noreferrer"
            download={attachment.filename}
            className="inline-flex h-7 w-7 items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
            title="Download"
          >
            <Download className="h-3.5 w-3.5" />
          </a>

          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-slate-400 hover:text-red-600 hover:bg-red-50"
            title="Delete"
            disabled={isDeleting}
            onClick={() => {
              if (confirm(`Delete "${attachment.filename}"?`)) onDelete()
            }}
          >
            {isDeleting ? <Spinner size="sm" /> : <Trash2 className="h-3.5 w-3.5" />}
          </Button>
        </div>
      </div>

      {previewOpen && canPreview && (
        <div className="border-t border-slate-100 px-3 pb-3 pt-2">
          {isImage ? (
            <img
              src={attachment.url}
              alt={attachment.filename}
              className="max-h-64 rounded object-contain"
            />
          ) : (
            <iframe
              src={attachment.url}
              title={attachment.filename}
              className="h-64 w-full rounded border border-slate-200"
            />
          )}
        </div>
      )}
    </li>
  )
}

interface AttachmentsPanelProps {
  entityType: EntityAttachmentEntityType
  entityId: string
}

export function AttachmentsPanel({ entityType, entityId }: AttachmentsPanelProps) {
  const [uploadingIds, setUploadingIds] = useState<Set<string>>(new Set())
  const [uploadProgress, setUploadProgress] = useState<Record<string, number>>({})

  const { data: attachments = [], isLoading } = useEntityAttachments(entityType, entityId)
  const upload = useUploadAttachment(entityType, entityId)
  const deleteAttachment = useDeleteAttachment(entityType, entityId)

  const handleFiles = async (files: File[]) => {
    for (const file of files) {
      const tempId = `${file.name}-${Date.now()}`
      setUploadingIds((prev) => new Set(prev).add(tempId))
      setUploadProgress((prev) => ({ ...prev, [tempId]: 0 }))

      try {
        await upload.mutateAsync({
          file,
          onProgress: (pct) => setUploadProgress((prev) => ({ ...prev, [tempId]: pct })),
        })
      } finally {
        setUploadingIds((prev) => {
          const next = new Set(prev)
          next.delete(tempId)
          return next
        })
        setUploadProgress((prev) => {
          const next = { ...prev }
          delete next[tempId]
          return next
        })
      }
    }
  }

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-slate-700">Attachments</h3>

      <FileUpload
        onFiles={handleFiles}
        maxSizeMb={25}
        multiple
        disabled={uploadingIds.size > 0}
      />

      {/* In-progress uploads */}
      {uploadingIds.size > 0 && (
        <ul className="space-y-1">
          {Array.from(uploadingIds).map((id) => (
            <li key={id} className="rounded-md border border-indigo-100 bg-indigo-50 px-3 py-2">
              <div className="flex items-center justify-between text-xs text-indigo-700">
                <span>Uploading…</span>
                <span>{uploadProgress[id] ?? 0}%</span>
              </div>
              <div className="mt-1.5 h-1 w-full overflow-hidden rounded-full bg-indigo-200">
                <div
                  className="h-full bg-indigo-500 transition-all"
                  style={{ width: `${uploadProgress[id] ?? 0}%` }}
                />
              </div>
            </li>
          ))}
        </ul>
      )}

      {/* Attachment list */}
      {isLoading ? (
        <div className="flex justify-center py-4">
          <Spinner />
        </div>
      ) : attachments.length === 0 ? (
        <p className="text-xs text-slate-400">No attachments yet.</p>
      ) : (
        <ul className="space-y-1.5">
          {attachments.map((a) => (
            <AttachmentRow
              key={a.id}
              attachment={a}
              isDeleting={deleteAttachment.isPending && deleteAttachment.variables === a.id}
              onDelete={() => deleteAttachment.mutate(a.id)}
            />
          ))}
        </ul>
      )}
    </div>
  )
}
