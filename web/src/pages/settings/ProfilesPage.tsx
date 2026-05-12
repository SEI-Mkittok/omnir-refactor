import { useEffect, useMemo, useState } from 'react'
import { Plus, Save, ShieldCheck, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import {
  useACLProfiles,
  useCreateACLProfile,
  usePermissionCatalog,
  useProfilePermissions,
  useSaveProfilePermissions,
} from '@/hooks/useAccessSettings'
import type { ACLAction, ACLModule, ACLProfileFieldPermission, ACLProfilePermission } from '@/api/types'

const ACTION_LABELS: Record<ACLAction, string> = {
  read: 'Read',
  create: 'Create',
  update: 'Update',
  delete: 'Delete',
  export: 'Export',
  admin: 'Admin',
}

function moduleLabel(module: ACLModule) {
  return module.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

export function upsertFieldPermissionOverride(
  fieldPermissions: ACLProfileFieldPermission[],
  override: ACLProfileFieldPermission
) {
  const normalizedOverride = {
    ...override,
    field_name: override.field_name.trim(),
  }
  let replaced = false

  const updated = fieldPermissions.map((field) => {
    if (field.module === normalizedOverride.module && field.field_name === normalizedOverride.field_name) {
      replaced = true
      return normalizedOverride
    }
    return field
  })

  return replaced ? updated : [...fieldPermissions, normalizedOverride]
}

export function canSaveProfilePermissions({
  selectedProfileId,
  editableProfileId,
  isLoading,
  isFetching,
  isSaving,
}: {
  selectedProfileId: string
  editableProfileId: string
  isLoading: boolean
  isFetching: boolean
  isSaving: boolean
}) {
  return Boolean(selectedProfileId) &&
    selectedProfileId === editableProfileId &&
    !isLoading &&
    !isFetching &&
    !isSaving
}

export function ProfilesPage() {
  const { data: profiles = [], isLoading } = useACLProfiles()
  const { data: catalog } = usePermissionCatalog()
  const createProfile = useCreateACLProfile()
  const savePermissions = useSaveProfilePermissions()

  const [selectedProfileId, setSelectedProfileId] = useState('')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [permissions, setPermissions] = useState<ACLProfilePermission[]>([])
  const [fieldPermissions, setFieldPermissions] = useState<ACLProfileFieldPermission[]>([])
  const [editableProfileId, setEditableProfileId] = useState('')
  const [newField, setNewField] = useState<{ module: ACLModule | ''; field_name: string; can_write: boolean }>({
    module: '',
    field_name: '',
    can_write: false,
  })

  const selectedProfile = profiles.find((profile) => profile.id === selectedProfileId)
  const permissionsQuery = useProfilePermissions(selectedProfileId)

  useEffect(() => {
    if (!selectedProfileId && profiles.length > 0) {
      setSelectedProfileId(profiles[0].id)
    }
  }, [profiles, selectedProfileId])

  useEffect(() => {
    setEditableProfileId('')
    setPermissions([])
    setFieldPermissions([])
  }, [selectedProfileId])

  useEffect(() => {
    if (!selectedProfileId || !permissionsQuery.data) return
    setPermissions(permissionsQuery.data?.permissions ?? [])
    setFieldPermissions(permissionsQuery.data?.field_permissions ?? [])
    setEditableProfileId(selectedProfileId)
  }, [permissionsQuery.data, selectedProfileId])

  const permissionLookup = useMemo(() => {
    const lookup = new Map<string, boolean>()
    for (const permission of permissions) {
      lookup.set(`${permission.module}:${permission.action}`, permission.allowed)
    }
    return lookup
  }, [permissions])

  function isAllowed(module: ACLModule, action: ACLAction) {
    return permissionLookup.get(`${module}:${action}`) ?? false
  }

  function togglePermission(module: ACLModule, action: ACLAction) {
    setPermissions((current) => {
      const exists = current.find((item) => item.module === module && item.action === action)
      if (exists) {
        return current.map((item) =>
          item.module === module && item.action === action
            ? { ...item, allowed: !item.allowed }
            : item
        )
      }
      return [...current, { module, action, allowed: true }]
    })
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    createProfile.mutate(
      { name: name.trim(), description: description.trim() || undefined },
      {
        onSuccess: (profile) => {
          setName('')
          setDescription('')
          setSelectedProfileId(profile.id)
        },
      }
    )
  }

  function addFieldPermission() {
    const fieldName = newField.field_name.trim()
    if (!newField.module || !fieldName) return
    setFieldPermissions((current) =>
      upsertFieldPermissionOverride(current, {
        module: newField.module as ACLModule,
        field_name: fieldName,
        can_write: newField.can_write,
      })
    )
    setNewField({ module: '', field_name: '', can_write: false })
  }

  function handleSave() {
    if (!canSave) return
    savePermissions.mutate({
      profileId: selectedProfileId,
      permissions,
      field_permissions: fieldPermissions,
    })
  }

  const modules = catalog?.modules ?? []
  const actions = catalog?.actions ?? []
  const canSave = canSaveProfilePermissions({
    selectedProfileId,
    editableProfileId,
    isLoading: permissionsQuery.isLoading,
    isFetching: permissionsQuery.isFetching,
    isSaving: savePermissions.isPending,
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Profiles</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Configure module actions and write-protected fields for each user profile.
        </p>
      </div>

      <form onSubmit={handleCreate} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-3 flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-slate-500" />
          <h2 className="text-sm font-semibold text-slate-900">Create Profile</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Profile name" />
          <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
          <Button type="submit" disabled={createProfile.isPending}>
            <Plus className="h-4 w-4" />
            Add
          </Button>
        </div>
      </form>

      <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
        <aside className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-200 px-4 py-3">
            <h2 className="text-sm font-semibold text-slate-900">
              {isLoading ? 'Loading profiles...' : `${profiles.length} profiles`}
            </h2>
          </div>
          <div className="divide-y divide-slate-100">
            {profiles.map((profile) => (
              <button
                key={profile.id}
                type="button"
                onClick={() => setSelectedProfileId(profile.id)}
                className={`block w-full px-4 py-3 text-left transition-colors ${
                  selectedProfileId === profile.id ? 'bg-slate-50' : 'hover:bg-slate-50'
                }`}
              >
                <div className="flex items-center gap-2">
                  <span className="font-medium text-slate-900">{profile.name}</span>
                  {profile.system_key && <Badge variant="gray">System</Badge>}
                </div>
                <p className="mt-0.5 text-sm text-slate-500">{profile.description || 'No description'}</p>
              </button>
            ))}
          </div>
        </aside>

        <section className="space-y-4">
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div>
              <h2 className="text-lg font-semibold text-slate-900">
                {selectedProfile?.name ?? 'Select a profile'}
              </h2>
              <p className="text-sm text-slate-500">
                {selectedProfile?.description ?? 'Choose a profile to edit permissions.'}
              </p>
            </div>
            <Button onClick={handleSave} disabled={!canSave}>
              <Save className="h-4 w-4" />
              {savePermissions.isPending ? 'Saving...' : 'Save Permissions'}
            </Button>
          </div>

          <div className="overflow-auto rounded-xl border border-slate-200 bg-white shadow-sm">
            <table className="min-w-full text-sm">
              <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-4 py-3">Module</th>
                  {actions.map((action) => (
                    <th key={action} className="px-4 py-3">{ACTION_LABELS[action]}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {modules.map((module) => (
                  <tr key={module}>
                    <td className="whitespace-nowrap px-4 py-3 font-medium text-slate-900">
                      {moduleLabel(module)}
                    </td>
                    {actions.map((action) => (
                      <td key={action} className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={isAllowed(module, action)}
                          onChange={() => togglePermission(module, action)}
                          className="h-4 w-4 rounded border-slate-300"
                        />
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <h2 className="text-sm font-semibold text-slate-900">Field Write Controls</h2>
            <p className="mt-1 text-sm text-slate-500">
              Add only fields that need an explicit write allow/deny override. Unlisted fields remain writable.
            </p>
            <div className="mt-3 grid gap-3 md:grid-cols-[180px_1fr_140px_auto]">
              <select
                className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                value={newField.module}
                onChange={(e) => setNewField((field) => ({ ...field, module: e.target.value as ACLModule }))}
              >
                <option value="">Module</option>
                {modules.map((module) => (
                  <option key={module} value={module}>{moduleLabel(module)}</option>
                ))}
              </select>
              <Input
                value={newField.field_name}
                onChange={(e) => setNewField((field) => ({ ...field, field_name: e.target.value }))}
                placeholder="field_name"
              />
              <select
                className="h-9 rounded-md border border-slate-200 bg-white px-3 text-sm"
                value={newField.can_write ? 'true' : 'false'}
                onChange={(e) => setNewField((field) => ({ ...field, can_write: e.target.value === 'true' }))}
              >
                <option value="false">Deny write</option>
                <option value="true">Allow write</option>
              </select>
              <Button type="button" variant="outline" onClick={addFieldPermission}>Add Field</Button>
            </div>
            <div className="mt-3 divide-y divide-slate-100">
              {fieldPermissions.map((field, index) => (
                <div key={`${field.module}-${field.field_name}-${index}`} className="flex items-center justify-between gap-3 py-2">
                  <span className="text-sm text-slate-700">
                    <strong>{moduleLabel(field.module)}</strong> / {field.field_name}: {field.can_write ? 'allow write' : 'deny write'}
                  </span>
                  <button
                    type="button"
                    className="text-red-600 hover:text-red-700"
                    onClick={() => setFieldPermissions((current) => current.filter((_, i) => i !== index))}
                    aria-label="Remove field permission"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
              {fieldPermissions.length === 0 && (
                <p className="py-3 text-sm text-slate-500">No field-level overrides configured.</p>
              )}
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}
