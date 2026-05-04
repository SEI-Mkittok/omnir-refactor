import { Plus, Star, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'

export type RelationshipRole =
  | 'primary'
  | 'billing'
  | 'technical'
  | 'executive'
  | 'decision-maker'

export interface RelationshipRow {
  id: string
  entityId?: string
  label: string
  meta?: string
  role: RelationshipRole
  isPrimary: boolean
}

export const RELATIONSHIP_ROLE_OPTIONS: { value: RelationshipRole; label: string }[] = [
  { value: 'primary', label: 'Primary' },
  { value: 'billing', label: 'Billing' },
  { value: 'technical', label: 'Technical' },
  { value: 'executive', label: 'Executive' },
  { value: 'decision-maker', label: 'Decision-maker' },
]

function createRelationshipId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `relationship-${Math.random().toString(36).slice(2, 10)}`
}

export function normalizeRelationshipRows(rows: RelationshipRow[]): RelationshipRow[] {
  if (rows.length === 0) return []

  const explicitPrimaryIndex = rows.findIndex((row) => row.isPrimary)
  const rolePrimaryIndex = rows.findIndex((row) => row.role === 'primary')
  const safePrimaryIndex =
    explicitPrimaryIndex >= 0
      ? explicitPrimaryIndex
      : rolePrimaryIndex >= 0
        ? rolePrimaryIndex
        : 0

  return rows.map((row, index) => {
    const isPrimary = index === safePrimaryIndex

    return {
      ...row,
      role: isPrimary ? 'primary' : row.role === 'primary' ? 'billing' : row.role,
      isPrimary,
    }
  })
}

export function createRelationshipRow(initial?: Partial<RelationshipRow>): RelationshipRow {
  return normalizeRelationshipRows([
    {
      id: initial?.id ?? createRelationshipId(),
      entityId: initial?.entityId,
      label: initial?.label ?? '',
      meta: initial?.meta ?? '',
      role: initial?.role ?? 'primary',
      isPrimary: initial?.isPrimary ?? true,
    },
  ])[0]
}

interface RelationshipEditorProps {
  value: RelationshipRow[]
  onChange: (rows: RelationshipRow[]) => void
  entityLabel: string
  title?: string
  emptyMessage?: string
  addLabel?: string
}

export function RelationshipEditor({
  value,
  onChange,
  entityLabel,
  title = 'Relationships',
  emptyMessage,
  addLabel,
}: RelationshipEditorProps) {
  const rows = normalizeRelationshipRows(value)

  const handleAdd = () => {
    const nextRows = normalizeRelationshipRows([
      ...rows,
      createRelationshipRow({
        role: rows.length === 0 ? 'primary' : 'billing',
        isPrimary: rows.length === 0,
      }),
    ])
    onChange(nextRows)
  }

  const handleUpdate = (rowId: string, patch: Partial<RelationshipRow>) => {
    const nextRows = normalizeRelationshipRows(
      rows.map((row) =>
        row.id === rowId
          ? {
              ...row,
              ...patch,
              role: patch.isPrimary ? 'primary' : patch.role ?? row.role,
              isPrimary: patch.isPrimary ?? row.isPrimary,
            }
          : row
      )
    )
    onChange(nextRows)
  }

  const handlePrimaryChange = (rowId: string) => {
    const nextRows = normalizeRelationshipRows(
      rows.map((row) => ({
        ...row,
        isPrimary: row.id === rowId,
        role: row.id === rowId ? 'primary' : row.role,
      }))
    )
    onChange(nextRows)
  }

  const handleRemove = (rowId: string) => {
    const nextRows = normalizeRelationshipRows(rows.filter((row) => row.id !== rowId))
    onChange(nextRows)
  }

  return (
    <section
      className="rounded-xl border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
      aria-label={title}
    >
      <div className="mb-4 flex items-center gap-2">
        <p
          className="text-[11px] font-semibold uppercase tracking-widest"
          style={{ color: 'var(--text-label)', letterSpacing: 'var(--letter-spacing-label)' }}
        >
          {title}
        </p>
        <Button variant="ghost" size="sm" className="ml-auto h-8 px-2 text-xs" onClick={handleAdd}>
          <Plus className="h-3.5 w-3.5" />
          {addLabel ?? `Add ${entityLabel}`}
        </Button>
      </div>

      {rows.length === 0 ? (
        <p className="text-sm" style={{ color: 'var(--text-label)' }}>
          {emptyMessage ?? `No ${entityLabel.toLowerCase()} relationships yet.`}
        </p>
      ) : (
        <div className="space-y-3">
          {rows.map((row, index) => (
            <div
              key={row.id}
              className="rounded-xl border p-3"
              style={{ borderColor: row.isPrimary ? 'var(--color-primary)' : 'var(--border-default)' }}
            >
              <div className="flex flex-wrap items-start gap-3">
                <div className="min-w-0 flex-1 space-y-3">
                  <div>
                    <label
                      htmlFor={`relationship-label-${row.id}`}
                      className="mb-1 block text-xs font-medium"
                      style={{ color: 'var(--text-secondary)' }}
                    >
                      {entityLabel}
                    </label>
                    <input
                      id={`relationship-label-${row.id}`}
                      value={row.label}
                      onChange={(event) => handleUpdate(row.id, { label: event.target.value })}
                      placeholder={`Enter ${entityLabel.toLowerCase()} name`}
                      className="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
                      style={{
                        borderColor: 'var(--border-default)',
                        background: 'var(--surface-app)',
                        color: 'var(--text-primary)',
                      }}
                    />
                  </div>

                  <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_180px]">
                    <div>
                      <label
                        htmlFor={`relationship-meta-${row.id}`}
                        className="mb-1 block text-xs font-medium"
                        style={{ color: 'var(--text-secondary)' }}
                      >
                        Details
                      </label>
                      <input
                        id={`relationship-meta-${row.id}`}
                        value={row.meta ?? ''}
                        onChange={(event) => handleUpdate(row.id, { meta: event.target.value })}
                        placeholder="Title, email, or context"
                        className="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
                        style={{
                          borderColor: 'var(--border-default)',
                          background: 'var(--surface-app)',
                          color: 'var(--text-primary)',
                        }}
                      />
                    </div>

                    <div>
                      <label
                        htmlFor={`relationship-role-${row.id}`}
                        className="mb-1 block text-xs font-medium"
                        style={{ color: 'var(--text-secondary)' }}
                      >
                        Role
                      </label>
                      <select
                        id={`relationship-role-${row.id}`}
                        value={row.role}
                        onChange={(event) =>
                          handleUpdate(row.id, {
                            role: event.target.value as RelationshipRole,
                            isPrimary: event.target.value === 'primary',
                          })
                        }
                        className="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
                        style={{
                          borderColor: 'var(--border-default)',
                          background: 'var(--surface-app)',
                          color: 'var(--text-primary)',
                        }}
                      >
                        {RELATIONSHIP_ROLE_OPTIONS.map((option) => (
                          <option key={option.value} value={option.value}>
                            {option.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handlePrimaryChange(row.id)}
                    className="inline-flex items-center gap-1 rounded-full border px-3 py-1.5 text-xs font-semibold transition-colors"
                    style={{
                      borderColor: row.isPrimary ? 'var(--color-primary)' : 'var(--border-default)',
                      background: row.isPrimary ? 'var(--color-primary-light)' : 'transparent',
                      color: row.isPrimary ? 'var(--color-primary)' : 'var(--text-secondary)',
                    }}
                    aria-pressed={row.isPrimary}
                    aria-label={row.isPrimary ? `${row.label || entityLabel} is primary` : `Make ${row.label || entityLabel} primary`}
                  >
                    <Star className="h-3.5 w-3.5" />
                    {row.isPrimary ? 'Primary' : 'Set primary'}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleRemove(row.id)}
                    className="rounded-md p-2 transition-colors hover:bg-[var(--surface-app)]"
                    aria-label={`Remove ${entityLabel.toLowerCase()} relationship ${index + 1}`}
                  >
                    <Trash2 className="h-4 w-4" style={{ color: 'var(--text-label)' }} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
