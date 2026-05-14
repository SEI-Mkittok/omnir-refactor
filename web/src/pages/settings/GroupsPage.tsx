import { useState } from 'react'
import { Plus, Trash2, Users } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  useACLGroups,
  useCreateACLGroup,
  useDeleteACLGroup,
  useGroupMemberCandidates,
  useSaveGroupMembers,
} from '@/hooks/useAccessSettings'
import type { ACLGroup } from '@/api/types'

function selectedOptions(select: HTMLSelectElement): string[] {
  return Array.from(select.selectedOptions).map((option) => option.value)
}

export function GroupsPage() {
  const { data: rawGroups, isLoading } = useACLGroups()
  const { data: usersData } = useGroupMemberCandidates()
  const createGroup = useCreateACLGroup()
  const saveMembers = useSaveGroupMembers()
  const deleteGroup = useDeleteACLGroup()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [draftMembers, setDraftMembers] = useState<Record<string, string[]>>({})

  const groups = rawGroups ?? []
  const users = usersData?.data ?? []

  function memberIDs(group: ACLGroup) {
    return draftMembers[group.id] ?? group.user_ids ?? []
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    createGroup.mutate(
      { name: name.trim(), description: description.trim() || undefined, user_ids: [] },
      {
        onSuccess: () => {
          setName('')
          setDescription('')
        },
      }
    )
  }

  function handleDelete(group: ACLGroup) {
    if (!confirm(`Delete group "${group.name}"? Sharing rules for this group will also be removed.`)) return
    deleteGroup.mutate(group.id)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[#1A1D23]">Groups</h1>
        <p className="mt-1 text-sm text-[#6B7280]">
          Group users for sharing-rule grants without changing the role hierarchy.
        </p>
      </div>

      <form onSubmit={handleCreate} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-3 flex items-center gap-2">
          <Users className="h-4 w-4 text-slate-500" />
          <h2 className="text-sm font-semibold text-slate-900">Create Group</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Group name" />
          <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
          <Button type="submit" disabled={createGroup.isPending}>
            <Plus className="h-4 w-4" />
            Add
          </Button>
        </div>
      </form>

      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-200 px-4 py-3">
          <h2 className="text-sm font-semibold text-slate-900">
            {isLoading ? 'Loading groups...' : `${groups.length} groups`}
          </h2>
        </div>
        <div className="divide-y divide-slate-100">
          {groups.map((group) => (
            <div key={group.id} className="grid gap-3 px-4 py-4 lg:grid-cols-[1fr_360px_auto] lg:items-start">
              <div>
                <h3 className="font-medium text-slate-900">{group.name}</h3>
                <p className="mt-0.5 text-sm text-slate-500">{group.description || 'No description'}</p>
                <p className="mt-2 text-xs font-medium uppercase tracking-wide text-slate-400">
                  {memberIDs(group).length} members
                </p>
              </div>
              <select
                multiple
                className="min-h-28 rounded-md border border-slate-200 bg-white px-3 py-2 text-sm"
                value={memberIDs(group)}
                onChange={(e) =>
                  setDraftMembers((current) => ({ ...current, [group.id]: selectedOptions(e.currentTarget) }))
                }
              >
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name} ({user.email})
                  </option>
                ))}
              </select>
              <div className="flex gap-2 lg:flex-col">
                <Button
                  type="button"
                  size="sm"
                  onClick={() => saveMembers.mutate({ id: group.id, user_ids: memberIDs(group) })}
                >
                  Save Members
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  className="text-red-600 hover:bg-red-50"
                  onClick={() => handleDelete(group)}
                >
                  <Trash2 className="h-4 w-4" />
                  Delete
                </Button>
              </div>
            </div>
          ))}
          {!isLoading && groups.length === 0 && (
            <p className="px-4 py-8 text-center text-sm text-slate-500">No groups configured yet.</p>
          )}
        </div>
      </div>
    </div>
  )
}
