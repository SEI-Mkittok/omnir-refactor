import apiClient from './client'

export type ImportEntity = 'contacts' | 'accounts' | 'leads'
export type ExportEntity = 'contacts' | 'accounts' | 'deals'

export interface ImportRowError {
  row: number
  field?: string
  message: string
}

export interface ImportResult {
  rows_processed: number
  created: number
  updated: number
  failed: number
  errors: ImportRowError[]
}

export interface FieldMapping {
  csv_column: string
  crm_field: string
}

export interface ImportRequest {
  field_mappings: FieldMapping[]
}

/**
 * Upload a CSV file and field mappings to the import endpoint.
 * Returns a structured result summary.
 */
export async function importCsv(
  entity: ImportEntity,
  file: File,
  fieldMappings: FieldMapping[]
): Promise<ImportResult> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('field_mappings', JSON.stringify(fieldMappings))

  const { data } = await apiClient.post<ImportResult>(`/import/${entity}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

/**
 * Parse the first row of a CSV file to get its column headers.
 * Runs entirely in the browser — no server round-trip.
 */
export function parseCsvHeaders(file: File): Promise<string[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target?.result as string
      const firstLine = text.split('\n')[0] ?? ''
      // Handle quoted columns
      const cols = firstLine
        .split(',')
        .map((c) => c.trim().replace(/^"|"$/g, ''))
        .filter(Boolean)
      resolve(cols)
    }
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsText(file)
  })
}

/**
 * Trigger a CSV download by navigating to the export URL with current filters.
 */
export function downloadExportCsv(entity: ExportEntity, params: Record<string, string> = {}) {
  const base = (import.meta.env.VITE_API_URL || '/api/v1') + `/export/${entity}`
  const qs = new URLSearchParams(params).toString()
  const url = qs ? `${base}?${qs}` : base
  // Open in same window so the browser triggers a file download
  window.location.href = url
}
