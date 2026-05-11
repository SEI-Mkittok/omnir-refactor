import type { UserRole } from '@/api/types'

type UserAssignmentForm = {
  role?: UserRole
  role_id?: string
  profile_id?: string
}

export function applyPlatformRoleChange<T extends UserAssignmentForm>(form: T, role: UserRole): T {
  if (form.role === role) return form

  const next = { ...form, role }
  delete next.role_id
  delete next.profile_id
  return next
}
