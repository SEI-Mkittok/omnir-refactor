import { useState } from 'react'
import { GitBranch, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import {
  useACLRoles,
  useCreateACLRole,
  useDeleteACLRole,
  useMoveACLRole,
} from '@/hooks/useAccessSettings'
import type { ACLRole } from '@/api/types'

function parentName(role: ACLRole, roles: ACLRole[]) {
  if (!role.parent_id) return 'Top level'
  return roles.find((candidate) => candidate.id === role.parent_id)?.name ?? 'Unknown parent'
}

export function RolesPage() {
  const { data: roles = [], isLoading } = useACLRoles()
  const createRole = useCreateACLRole()
  const moveRole = useMoveACLRole()
  const deleteRole = useDeleteACLRole()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [parentId, setParentId] = useState('')
  const [error, setError] = useState('')

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (!name.trim()) {
      setError('Role name is required.')
      return
    }
    createRole.mutate(
      { name: name.trim(), description: description.trim() || undefined, parent_id: parentId || undefined },
      {
        onSuccess: () => {
          setName('')
          setDescription('')
          setParentId('')
        },
        onError: () => setError('Unable to create role. Check for duplicate names or invalid parent.'),
      }
    )
  }

  function handleDelete(role: ACLRole) {
    if (role.system_key) return
    if (!confirm(`Delete role "${role.name}"? Users must be reassigned separately.`)) return
    deleteRole.mutate(role.id)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Roles</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Maintain the org role hierarchy used by sharing rules and record visibility.
        </p>
      </div>

      <form onSubmit={handleCreate} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-3 flex items-center gap-2">
          <GitBranch className="h-4 w-4 text-slate-500" />
          <h2 className="text-sm font-semibold text-slate-900">Create Role</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-[1fr_1fr_220px_auto]">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Role name" />
          <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
          <select
            className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
            value={parentId}
            onChange={(e) => setParentId(e.target.value)}
          >
            <option value="">Top level</option>
            {roles.map((role) => (
              <option key={role.id} value={role.id}>{role.name}</option>
            ))}
          </select>
          <Button type="submit" disabled={createRole.isPending}>
            <Plus className="h-4 w-4" />
            Add
          </Button>
        </div>
        {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
      </form>

      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-200 px-4 py-3">
          <h2 className="text-sm font-semibold text-slate-900">
            {isLoading ? 'Loading roles...' : `${roles.length} roles`}
          </h2>
        </div>
        <div className="divide-y divide-slate-100">
          {roles.map((role) => (
            <div key={role.id} className="grid gap-3 px-4 py-3 md:grid-cols-[1fr_220px_auto] md:items-center">
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium text-slate-900">{role.name}</span>
                  {role.system_key && <Badge variant="gray">System</Badge>}
                </div>
                <p className="mt-0.5 text-sm text-slate-500">{role.description || parentName(role, roles)}</p>
              </div>
              <select
                className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                value={role.parent_id ?? ''}
                onChange={(e) => moveRole.mutate({ id: role.id, parent_id: e.target.value || null })}
              >
                <option value="">Top level</option>
                {roles
                  .filter((candidate) => candidate.id !== role.id)
                  .map((candidate) => (
                    <option key={candidate.id} value={candidate.id}>{candidate.name}</option>
                  ))}
              </select>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-red-600 hover:bg-red-50"
                disabled={Boolean(role.system_key)}
                onClick={() => handleDelete(role)}
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </Button>
            </div>
          ))}
          {!isLoading && roles.length === 0 && (
            <p className="px-4 py-8 text-center text-sm text-slate-500">No roles configured yet.</p>
          )}
        </div>
      </div>
    </div>
  )
}
