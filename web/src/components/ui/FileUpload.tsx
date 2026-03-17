import { useRef, useState, useCallback } from 'react'
import { Upload, X, File as FileIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './Button'

interface FileUploadProps {
  onFiles: (files: File[]) => void
  accept?: string
  multiple?: boolean
  maxSizeMb?: number
  disabled?: boolean
  className?: string
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileUpload({
  onFiles,
  accept,
  multiple = true,
  maxSizeMb = 10,
  disabled = false,
  className,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [pendingFiles, setPendingFiles] = useState<File[]>([])
  const [errors, setErrors] = useState<string[]>([])

  const maxBytes = maxSizeMb * 1024 * 1024

  const processFiles = useCallback(
    (rawFiles: FileList | File[]) => {
      const valid: File[] = []
      const errs: string[] = []
      Array.from(rawFiles).forEach((f) => {
        if (f.size > maxBytes) {
          errs.push(`${f.name} exceeds ${maxSizeMb} MB limit`)
        } else {
          valid.push(f)
        }
      })
      setErrors(errs)
      if (valid.length > 0) {
        const next = multiple ? [...pendingFiles, ...valid] : [valid[0]]
        setPendingFiles(next)
        onFiles(next)
      }
    },
    [maxBytes, maxSizeMb, multiple, onFiles, pendingFiles]
  )

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    if (!disabled) setIsDragging(true)
  }

  const handleDragLeave = () => setIsDragging(false)

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    if (disabled) return
    processFiles(e.dataTransfer.files)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) processFiles(e.target.files)
    e.target.value = ''
  }

  const removeFile = (index: number) => {
    const next = pendingFiles.filter((_, i) => i !== index)
    setPendingFiles(next)
    onFiles(next)
  }

  return (
    <div className={cn('space-y-2', className)}>
      {/* Drop zone */}
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => !disabled && inputRef.current?.click()}
        className={cn(
          'flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed px-4 py-8 text-center transition-colors',
          isDragging
            ? 'border-indigo-400 bg-indigo-50'
            : 'border-slate-300 bg-slate-50 hover:border-slate-400 hover:bg-slate-100',
          disabled && 'cursor-not-allowed opacity-50'
        )}
      >
        <Upload className="h-6 w-6 text-slate-400" />
        <div>
          <p className="text-sm font-medium text-slate-700">
            Drop files here or <span className="text-indigo-600">browse</span>
          </p>
          <p className="text-xs text-slate-500 mt-0.5">
            {accept ? accept.replace(/,/g, ', ') : 'Any file type'} · max {maxSizeMb} MB
          </p>
        </div>
      </div>

      <input
        ref={inputRef}
        type="file"
        accept={accept}
        multiple={multiple}
        className="hidden"
        onChange={handleInputChange}
        disabled={disabled}
      />

      {/* Errors */}
      {errors.length > 0 && (
        <ul className="space-y-1">
          {errors.map((err, i) => (
            <li key={i} className="text-xs text-red-600">
              {err}
            </li>
          ))}
        </ul>
      )}

      {/* Pending files list */}
      {pendingFiles.length > 0 && (
        <ul className="space-y-1">
          {pendingFiles.map((f, i) => (
            <li
              key={i}
              className="flex items-center justify-between gap-2 rounded-md border border-slate-200 bg-white px-3 py-2"
            >
              <div className="flex items-center gap-2 min-w-0">
                <FileIcon className="h-4 w-4 shrink-0 text-slate-400" />
                <span className="truncate text-sm text-slate-700">{f.name}</span>
                <span className="shrink-0 text-xs text-slate-500">{formatBytes(f.size)}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0"
                onClick={(e) => {
                  e.stopPropagation()
                  removeFile(i)
                }}
              >
                <X className="h-3 w-3" />
              </Button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
