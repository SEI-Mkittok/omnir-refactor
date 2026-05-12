import type { UserRole } from '@/api/types'

type UserAssignmentForm = {
  role?: UserRole
  role_id?: string | null
  profile_id?: string | null
}

export function applyPlatformRoleChange<T extends UserAssignmentForm>(form: T, role: UserRole): T {
  if (form.role === role) return form

  const next = { ...form, role }
  delete next.role_id
  delete next.profile_id
  return next
}
