import { useState } from 'react'
import { Upload, CheckCircle, XCircle, ArrowRight, RotateCcw } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { FileUpload } from '@/components/ui/FileUpload'
import { Spinner } from '@/components/ui/Spinner'
import {
  importCsv,
  parseCsvHeaders,
  type ImportEntity,
  type FieldMapping,
  type ImportResult,
} from '@/api/importExport'

// Known CRM fields per entity
const CRM_FIELDS: Record<ImportEntity, { value: string; label: string }[]> = {
  contacts: [
    { value: 'first_name', label: 'First name *' },
    { value: 'last_name', label: 'Last name *' },
    { value: 'email', label: 'Email *' },
    { value: 'phone', label: 'Phone' },
    { value: 'title', label: 'Title' },
    { value: 'department', label: 'Department' },
    { value: 'stage', label: 'Stage' },
    { value: 'account_id', label: 'Account ID' },
    { value: 'tags', label: 'Tags (semicolon-separated)' },
    { value: '__skip__', label: '— Skip this column —' },
  ],
  accounts: [
    { value: 'name', label: 'Name *' },
    { value: 'domain', label: 'Domain' },
    { value: 'industry', label: 'Industry' },
    { value: 'size', label: 'Size' },
    { value: 'tags', label: 'Tags (semicolon-separated)' },
    { value: '__skip__', label: '— Skip this column —' },
  ],
  leads: [
    { value: 'first_name', label: 'First name *' },
    { value: 'last_name', label: 'Last name *' },
    { value: 'email', label: 'Email *' },
    { value: 'phone', label: 'Phone' },
    { value: 'company', label: 'Company' },
    { value: 'lead_source', label: 'Lead source' },
    { value: 'status', label: 'Status' },
    { value: '__skip__', label: '— Skip this column —' },
  ],
}

/** Attempt to auto-match a CSV header to a known CRM field */
function autoMatch(header: string, fields: { value: string }[]): string {
  const normalized = header.toLowerCase().replace(/[\s_-]/g, '_')
  const match = fields.find((f) => f.value === normalized || f.value === header.toLowerCase())
  return match?.value ?? '__skip__'
}

type Step = 'upload' | 'mapping' | 'importing' | 'results'

interface ImportModalProps {
  open: boolean
  onClose: () => void
  entity: ImportEntity
  entityLabel: string
}

export function ImportModal({ open, onClose, entity, entityLabel }: ImportModalProps) {
  const [step, setStep] = useState<Step>('upload')
  const [file, setFile] = useState<File | null>(null)
  const [csvHeaders, setCsvHeaders] = useState<string[]>([])
  const [mappings, setMappings] = useState<Record<string, string>>({})
  const [result, setResult] = useState<ImportResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const crmFields = CRM_FIELDS[entity]

  const handleFiles = async (files: File[]) => {
    const f = files[0]
    if (!f) return
    setFile(f)
    setError(null)
    try {
      const headers = await parseCsvHeaders(f)
      setCsvHeaders(headers)
      // Auto-map columns
      const auto: Record<string, string> = {}
      headers.forEach((h) => { auto[h] = autoMatch(h, crmFields) })
      setMappings(auto)
    } catch {
      setError('Failed to read CSV headers. Make sure the file is a valid CSV.')
    }
  }

  const handleStartImport = async () => {
    if (!file) return
    setStep('importing')
    setError(null)
    try {
      const fieldMappings: FieldMapping[] = Object.entries(mappings)
        .filter(([, crm]) => crm !== '__skip__')
        .map(([csv_column, crm_field]) => ({ csv_column, crm_field }))

      const res = await importCsv(entity, file, fieldMappings)
      setResult(res)
      setStep('results')
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : 'Import failed. Please check your file and try again.'
      setError(msg)
      setStep('mapping')
    }
  }

  const handleReset = () => {
    setStep('upload')
    setFile(null)
    setCsvHeaders([])
    setMappings({})
    setResult(null)
    setError(null)
  }

  const handleClose = () => {
    handleReset()
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Upload className="h-5 w-5 text-[var(--color-primary)]" />
            Import {entityLabel}
          </DialogTitle>
        </DialogHeader>

        {/* Step indicator */}
        <div className="flex items-center gap-1.5 text-xs text-slate-500 -mt-2">
          <span className={step === 'upload' ? 'text-[var(--color-primary)] font-medium' : ''}>1. Upload</span>
          <ArrowRight className="h-3 w-3" />
          <span className={step === 'mapping' ? 'text-[var(--color-primary)] font-medium' : ''}>2. Map fields</span>
          <ArrowRight className="h-3 w-3" />
          <span className={step === 'results' ? 'text-[var(--color-primary)] font-medium' : ''}>3. Results</span>
        </div>

        {/* ── Step 1: Upload ── */}
        {step === 'upload' && (
          <div className="space-y-3">
            <p className="text-sm text-slate-600">
              Upload a CSV file to import {entityLabel.toLowerCase()} in bulk. The first row must be a header row.
            </p>
            <FileUpload
              onFiles={handleFiles}
              accept=".csv"
              multiple={false}
              maxSizeMb={20}
            />
            {error && <p className="text-xs text-red-600">{error}</p>}
          </div>
        )}

        {/* ── Step 2: Field mapping ── */}
        {step === 'mapping' && (
          <div className="space-y-3">
            <p className="text-sm text-slate-600">
              Map each CSV column to a {entityLabel.slice(0, -1).toLowerCase()} field. Columns set to "Skip" will be ignored.
            </p>
            <div className="max-h-72 overflow-y-auto rounded-lg border border-slate-200 divide-y divide-slate-100">
              {csvHeaders.map((header) => (
                <div key={header} className="flex items-center gap-3 px-3 py-2 bg-white">
                  <span className="w-1/2 truncate text-sm font-medium text-slate-700" title={header}>
                    {header}
                  </span>
                  <select
                    className="w-1/2 rounded-md border border-slate-200 bg-white px-2 py-1 text-sm text-slate-900 focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                    value={mappings[header] ?? '__skip__'}
                    onChange={(e) =>
                      setMappings((m) => ({ ...m, [header]: e.target.value }))
                    }
                  >
                    {crmFields.map((f) => (
                      <option key={f.value} value={f.value}>
                        {f.label}
                      </option>
                    ))}
                  </select>
                </div>
              ))}
            </div>
            {error && <p className="text-xs text-red-600">{error}</p>}
          </div>
        )}

        {/* ── Step 3: Importing (progress) ── */}
        {step === 'importing' && (
          <div className="flex flex-col items-center gap-4 py-8">
            <Spinner size="lg" />
            <p className="text-sm text-slate-600">Importing {entityLabel.toLowerCase()}… please wait.</p>
          </div>
        )}

        {/* ── Step 4: Results ── */}
        {step === 'results' && result && (
          <div className="space-y-4">
            {/* Summary */}
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <Stat label="Processed" value={result.rows_processed} />
              <Stat label="Created" value={result.created} color="green" />
              <Stat label="Updated" value={result.updated} color="blue" />
              <Stat label="Failed" value={result.failed} color={result.failed > 0 ? 'red' : undefined} />
            </div>

            {/* Row-level errors */}
            {result.errors.length > 0 && (
              <div className="space-y-1.5">
                <p className="text-xs font-semibold text-slate-700 uppercase tracking-wide">
                  Row errors
                </p>
                <div className="max-h-48 overflow-y-auto rounded-lg border border-red-100 bg-red-50 divide-y divide-red-100">
                  {result.errors.map((e, i) => (
                    <div key={i} className="flex items-start gap-2 px-3 py-2">
                      <XCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-red-500" />
                      <p className="text-xs text-red-700">
                        Row {e.row}
                        {e.field ? ` · ${e.field}` : ''}: {e.message}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {result.failed === 0 && (
              <div className="flex items-center gap-2 text-sm text-green-700">
                <CheckCircle className="h-4 w-4" />
                Import completed successfully.
              </div>
            )}
          </div>
        )}

        {/* Footer actions */}
        <DialogFooter>
          {step === 'upload' && (
            <>
              <Button variant="outline" onClick={handleClose}>Cancel</Button>
              <Button
                disabled={!file || !!error}
                onClick={() => setStep('mapping')}
              >
                Next: Map fields
              </Button>
            </>
          )}
          {step === 'mapping' && (
            <>
              <Button variant="outline" onClick={() => setStep('upload')}>Back</Button>
              <Button onClick={handleStartImport}>
                Start import
              </Button>
            </>
          )}
          {step === 'results' && (
            <>
              <Button variant="outline" onClick={handleReset}>
                <RotateCcw className="h-4 w-4" />
                Import another file
              </Button>
              <Button onClick={handleClose}>Done</Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function Stat({
  label,
  value,
  color,
}: {
  label: string
  value: number
  color?: 'green' | 'blue' | 'red'
}) {
  const colorClass =
    color === 'green'
      ? 'text-green-700 bg-green-50 border-green-100'
      : color === 'blue'
        ? 'text-blue-700 bg-blue-50 border-blue-100'
        : color === 'red'
          ? 'text-red-700 bg-red-50 border-red-100'
          : 'text-slate-700 bg-slate-50 border-slate-200'

  return (
    <div className={`rounded-lg border px-3 py-2.5 text-center ${colorClass}`}>
      <p className="text-xl font-bold">{value}</p>
      <p className="text-xs">{label}</p>
    </div>
  )
}
