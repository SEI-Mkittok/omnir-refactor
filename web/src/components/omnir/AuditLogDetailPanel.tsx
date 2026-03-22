import { LogIn, FileOutput } from 'lucide-react'
import { SidePanel } from '@/components/ui/SidePanel'
import { useAuditLogEntry } from '@/hooks/useAuditLog'
import { formatDate } from '@/lib/utils'
import type { AuditAction, AuditChanges } from '@/api/types'

const ACTION_COLORS: Record<AuditAction, string> = {
  created: 'bg-green-100 text-green-700',
  updated: 'bg-blue-100 text-blue-700',
  deleted: 'bg-red-100 text-red-700',
  converted: 'bg-teal-100 text-teal-700',
  login: 'bg-slate-100 text-slate-700',
  export: 'bg-amber-100 text-amber-700',
}

const ACTION_ICONS: Partial<Record<AuditAction, React.ReactNode>> = {
  login: <LogIn className="h-3 w-3" />,
  export: <FileOutput className="h-3 w-3" />,
}

interface AuditLogDetailPanelProps {
  entryId: string | null
  onClose: () => void
}

function DiffView({ changes }: { changes: AuditChanges }) {
  const fields = Object.keys(changes)
  if (fields.length === 0) {
    return <p className="text-sm text-slate-500 italic">No field changes recorded.</p>
  }

  return (
    <div className="space-y-3">
      {fields.map((field) => {
        const { from, to } = changes[field]
        return (
          <div key={field} className="rounded-md border border-slate-200 overflow-hidden text-sm">
            <div className="bg-slate-100 px-3 py-1.5 font-medium text-slate-700 text-xs uppercase tracking-wide">
              {field.replace(/_/g, ' ')}
            </div>
            <div className="grid grid-cols-2 divide-x divide-slate-200">
              <div className="px-3 py-2 bg-red-50">
                <p className="text-xs font-medium text-red-600 mb-1">Before</p>
                <pre className="text-xs text-red-800 whitespace-pre-wrap break-all">
                  {from == null ? <span className="italic text-slate-400">—</span> : JSON.stringify(from, null, 2)}
                </pre>
              </div>
              <div className="px-3 py-2 bg-green-50">
                <p className="text-xs font-medium text-green-600 mb-1">After</p>
                <pre className="text-xs text-green-800 whitespace-pre-wrap break-all">
                  {to == null ? <span className="italic text-slate-400">—</span> : JSON.stringify(to, null, 2)}
                </pre>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function Field({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-medium text-slate-500 uppercase tracking-wide">{label}</dt>
      <dd className="mt-0.5 text-sm text-slate-900">{value ?? <span className="italic text-slate-400">—</span>}</dd>
    </div>
  )
}

export function AuditLogDetailPanel({ entryId, onClose }: AuditLogDetailPanelProps) {
  const { data: entry, isLoading } = useAuditLogEntry(entryId)

  return (
    <SidePanel open={!!entryId} onClose={onClose} title="Audit Log Entry" width="lg">
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
        </div>
      )}

      {!isLoading && entry && (
        <div className="space-y-6">
          {/* Meta */}
          <dl className="grid grid-cols-2 gap-4">
            <Field label="Timestamp" value={formatDate(entry.created_at)} />
            <Field label="Action" value={
              <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium capitalize ${ACTION_COLORS[entry.action] ?? 'bg-slate-100 text-slate-700'}`}>
                {ACTION_ICONS[entry.action]}
                {entry.action}
              </span>
            } />
            <Field label="Entity Type" value={
              <span className="capitalize">{entry.entity_type}</span>
            } />
            <Field label="Entity ID" value={entry.entity_id ?? '—'} />
            <Field label="Entity Name" value={entry.entity_name} />
            <Field label="Actor" value={entry.user_id ?? entry.agent_id ?? 'System'} />
            <Field label="IP Address" value={entry.ip_address} />
          </dl>

          {/* Diff */}
          {entry.changes && Object.keys(entry.changes).length > 0 && (
            <div>
              <h3 className="text-sm font-semibold text-slate-900 mb-3">Field Changes</h3>
              <DiffView changes={entry.changes} />
            </div>
          )}

          {!entry.changes || Object.keys(entry.changes).length === 0 ? (
            <p className="text-sm text-slate-500">No field-level changes recorded for this event.</p>
          ) : null}
        </div>
      )}
    </SidePanel>
  )
}
