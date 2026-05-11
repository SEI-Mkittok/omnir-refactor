import { useState, useCallback } from 'react'
import axios from 'axios'
import { Plus, UserCog } from 'lucide-react'
import { useDebounce } from '@/hooks/useDebounce'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/hooks/useUsers'
import { useACLRoles, useACLProfiles } from '@/hooks/useAccessSettings'
import { useAuthStore } from '@/stores/auth'
import { FilterBar } from '@/components/ui/FilterBar'
import { Table, type Column } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/Dialog'
import { formatDate } from '@/lib/utils'
import { canAccessAdminModule } from '@/lib/access'
import { applyPlatformRoleChange } from '@/lib/userAssignments'
import type { ACLProfile, ACLRole, User, UserRole, CreateUserRequest, UpdateUserRequest } from '@/api/types'

const ROLE_OPTIONS = [
  { label: 'All Roles', value: '' },
  { label: 'Super Admin', value: 'super_admin' },
  { label: 'Admin', value: 'admin' },
  { label: 'Agent', value: 'agent' },
  { label: 'Client', value: 'client' },
]

const ASSIGNABLE_ROLE_OPTIONS: Array<{ label: string; value: UserRole }> = [
  { label: 'Admin', value: 'admin' },
  { label: 'Agent', value: 'agent' },
  { label: 'Client', value: 'client' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'created_at:desc' },
  { label: 'Oldest', value: 'created_at:asc' },
  { label: 'Name A–Z', value: 'name:asc' },
  { label: 'Name Z–A', value: 'name:desc' },
]

const roleBadgeVariant: Record<UserRole, 'default' | 'blue' | 'yellow' | 'green' | 'red' | 'gray' | 'indigo' | 'purple' | 'orange'> = {
  super_admin: 'purple',
  admin: 'red',
  agent: 'indigo',
  client: 'green',
}

const roleLabel: Record<UserRole, string> = {
  super_admin: 'Super Admin',
  admin: 'Admin',
  agent: 'Agent',
  client: 'Client',
}

function getApiErrorMessage(err: unknown, fallback: string): string {
  if (!axios.isAxiosError(err)) return fallback
  const payload = err.response?.data as { error?: unknown; message?: unknown } | undefined
  if (typeof payload?.error === 'string' && payload.error.trim()) return payload.error
  if (typeof payload?.message === 'string' && payload.message.trim()) return payload.message
  return fallback
}

// ---- Create User Dialog ----

interface CreateUserDialogProps {
  open: boolean
  onClose: () => void
  aclRoles: ACLRole[]
  profiles: ACLProfile[]
}

function CreateUserDialog({ open, onClose, aclRoles, profiles }: CreateUserDialogProps) {
  const { mutate: createUser, isPending } = useCreateUser()
  const [form, setForm] = useState<CreateUserRequest>({
    name: '',
    email: '',
    password: '',
    role: 'agent',
  })
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.name || !form.email || !form.password) {
      setError('Name, email, and password are required.')
      return
    }
    createUser(form, {
      onSuccess: () => {
        setForm({ name: '', email: '', password: '', role: 'agent' })
        onClose()
      },
      onError: (err) =>
        setError(getApiErrorMessage(err, 'Failed to create user. Email may already be in use.')),
    })
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add User</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Name</label>
            <Input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="Full name"
              autoFocus
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Email</label>
            <Input
              type="email"
              value={form.email}
              onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
              placeholder="email@example.com"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Password</label>
            <Input
              type="password"
              value={form.password}
              onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              placeholder="Min 8 characters"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Platform Role</label>
            <select
              className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
              value={form.role}
              onChange={(e) => setForm((f) => applyPlatformRoleChange(f, e.target.value as UserRole))}
            >
              {ASSIGNABLE_ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Org Role</label>
              <select
                className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
                value={form.role_id ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, role_id: e.target.value || undefined }))}
              >
                <option value="">Default from platform role</option>
                {aclRoles.map((role) => (
                  <option key={role.id} value={role.id}>{role.name}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Profile</label>
              <select
                className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
                value={form.profile_id ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, profile_id: e.target.value || undefined }))}
              >
                <option value="">Default from platform role</option>
                {profiles.map((profile) => (
                  <option key={profile.id} value={profile.id}>{profile.name}</option>
                ))}
              </select>
            </div>
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create User'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ---- Edit User Dialog ----

interface EditUserDialogProps {
  user: User | null
  currentUser: User | null
  currentUserId: string
  onClose: () => void
  aclRoles: ACLRole[]
  profiles: ACLProfile[]
}

function EditUserDialog({ user, currentUser, currentUserId, onClose, aclRoles, profiles }: EditUserDialogProps) {
  const { mutate: updateUser, isPending } = useUpdateUser()
  const [form, setForm] = useState<UpdateUserRequest>({})
  const [error, setError] = useState('')

  const isAdmin = canAccessAdminModule(currentUser, 'users')
  const isSelf = user?.id === currentUserId

  // Populate form when user changes
  const handleOpen = () => {
    if (user) {
      setForm({
        name: user.name,
        email: user.email,
        role: user.role,
        role_id: user.role_id,
        profile_id: user.profile_id,
      })
      setError('')
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!user) return
    setError('')
    updateUser(
      { id: user.id, payload: form },
      {
        onSuccess: onClose,
        onError: (err) => setError(getApiErrorMessage(err, 'Failed to update user.')),
      }
    )
  }

  return (
    <Dialog open={!!user} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent onOpenAutoFocus={handleOpen}>
        <DialogHeader>
          <DialogTitle>Edit User</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Name</label>
            <Input
              value={form.name ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="Full name"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-slate-700">Email</label>
            <Input
              type="email"
              value={form.email ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
              placeholder="email@example.com"
              disabled={!isAdmin && !isSelf}
            />
          </div>
          {isAdmin && (
            <div className="grid gap-3 md:grid-cols-3">
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-slate-700">Platform Role</label>
                <select
                  className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
                  value={form.role ?? 'agent'}
                  onChange={(e) => setForm((f) => applyPlatformRoleChange(f, e.target.value as UserRole))}
                >
                  {form.role === 'super_admin' && (
                    <option value="super_admin" disabled>
                      Super Admin
                    </option>
                  )}
                  {ASSIGNABLE_ROLE_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-slate-700">Org Role</label>
                <select
                  className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
                  value={form.role_id ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, role_id: e.target.value || undefined }))}
                >
                  <option value="">Default</option>
                  {aclRoles.map((role) => (
                    <option key={role.id} value={role.id}>{role.name}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-slate-700">Profile</label>
                <select
                  className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--border-focus)]"
                  value={form.profile_id ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, profile_id: e.target.value || undefined }))}
                >
                  <option value="">Default</option>
                  {profiles.map((profile) => (
                    <option key={profile.id} value={profile.id}>{profile.name}</option>
                  ))}
                </select>
              </div>
            </div>
          )}
          {error && <p className="text-sm text-red-600">{error}</p>}
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Saving…' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ---- Main page ----

export function UsersPage() {
  const currentUser = useAuthStore((s) => s.user)
  const isAdmin = canAccessAdminModule(currentUser, 'users')

  const [search, setSearch] = useState('')
  const [role, setRole] = useState('')
  const [sortKey, setSortKey] = useState('created_at:desc')
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)
  const [editUser, setEditUser] = useState<User | null>(null)

  const debouncedSearch = useDebounce(search, 300)
  const [sort, order] = sortKey.split(':') as [string, 'asc' | 'desc']

  const { data, isLoading } = useUsers({
    page,
    limit: 20,
    q: debouncedSearch || undefined,
    role: (role as UserRole) || undefined,
    sort,
    order,
  })

  const { mutate: deleteUser } = useDeleteUser()
  const { data: aclRoles = [] } = useACLRoles()
  const { data: profiles = [] } = useACLProfiles()

  const users = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const totalPages = data?.meta?.total_pages ?? 1

  const handleSort = useCallback((key: string) => {
    setSortKey((prev) => {
      const [prevKey, prevDir] = prev.split(':')
      if (prevKey === key) return `${key}:${prevDir === 'asc' ? 'desc' : 'asc'}`
      return `${key}:asc`
    })
  }, [])

  const handleDelete = (u: User) => {
    if (!confirm(`Delete user "${u.name}"? This cannot be undone.`)) return
    deleteUser(u.id)
  }

  const columns: Column<User>[] = [
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (u) => (
        <span className="font-medium text-slate-900">{u.name}</span>
      ),
    },
    {
      key: 'email',
      header: 'Email',
      render: (u) => <span className="text-slate-600">{u.email}</span>,
    },
    {
      key: 'role',
      header: 'Platform Role',
      render: (u) => (
        <Badge variant={roleBadgeVariant[u.role]}>{roleLabel[u.role]}</Badge>
      ),
    },
    {
      key: 'profile_name',
      header: 'Access',
      hideOnMobile: true,
      render: (u) => (
        <span className="text-sm text-slate-600">
          {u.role_name ?? 'Default'} / {u.profile_name ?? 'Default'}
        </span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      hideOnMobile: true,
      render: (u) => <span className="text-slate-500 text-xs">{formatDate(u.created_at)}</span>,
    },
    {
      key: 'id',
      header: '',
      render: (u) => (
        <div className="flex items-center justify-end gap-2">
          {(isAdmin || u.id === currentUser?.id) && (
            <Button
              variant="outline"
              size="sm"
              onClick={(e) => { e.stopPropagation(); setEditUser(u) }}
            >
              Edit
            </Button>
          )}
          {isAdmin && u.id !== currentUser?.id && (
            <Button
              variant="outline"
              size="sm"
              className="text-red-600 hover:bg-red-50 hover:border-red-300"
              onClick={(e) => { e.stopPropagation(); handleDelete(u) }}
            >
              Delete
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Users</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${total} total`}
          </p>
        </div>
        {isAdmin && (
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" />
            Add User
          </Button>
        )}
      </div>

      {/* Filters */}
      <FilterBar
        searchValue={search}
        onSearchChange={(v) => { setSearch(v); setPage(1) }}
        searchPlaceholder="Search users…"
        filters={[
          {
            label: 'Role',
            value: role,
            options: ROLE_OPTIONS,
            onChange: (v) => { setRole(v); setPage(1) },
          },
          {
            label: 'Sort',
            value: sortKey,
            options: SORT_OPTIONS,
            onChange: (v) => { setSortKey(v); setPage(1) },
          },
        ]}
      />

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={users}
          isLoading={isLoading}
          sortBy={sort}
          sortDir={order}
          onSort={handleSort}
          emptyIcon={UserCog}
          emptyTitle="No users found"
          emptyDescription="Try adjusting your search or filters."
          keyExtractor={(u) => u.id}
        />

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 border-t border-slate-200 px-4 py-3">
            <p className="text-sm text-slate-500">
              Page {page} of {totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Dialogs */}
      {isAdmin && (
        <CreateUserDialog
          open={showCreate}
          onClose={() => setShowCreate(false)}
          aclRoles={aclRoles}
          profiles={profiles}
        />
      )}
      <EditUserDialog
        user={editUser}
        currentUser={currentUser}
        currentUserId={currentUser?.id ?? ''}
        aclRoles={aclRoles}
        profiles={profiles}
        onClose={() => setEditUser(null)}
      />
    </div>
  )
}
