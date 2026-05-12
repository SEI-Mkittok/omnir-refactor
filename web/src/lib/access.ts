import type { ACLAction, ACLModule, User } from '@/api/types'

export const adminRouteModules: Record<string, ACLModule> = {
  '/users': 'users',
  '/settings/roles': 'settings',
  '/settings/profiles': 'settings',
  '/settings/sharing-rules': 'settings',
  '/settings/groups': 'settings',
  '/settings/custom-fields': 'custom_fields',
  '/settings/numbering': 'settings',
  '/settings/company': 'settings',
  '/settings/portal': 'settings',
  '/settings/outgoing-server': 'settings',
  '/settings/config-editor': 'settings',
  '/settings/menu': 'settings',
  '/settings/currencies': 'settings',
  '/settings/picklists': 'settings',
  '/settings/picklist-dependencies': 'settings',
  '/settings/lead-conversion-mapping': 'settings',
  '/api-keys': 'api_keys',
  '/settings/sla': 'sla',
  '/settings/webhooks': 'webhooks',
  '/settings/billing': 'billing',
  '/settings/billing/plans': 'billing',
  '/settings/integrations': 'integrations',
  '/admin/audit': 'audit_log',
}

export function isPlatformAdmin(user: User | null | undefined) {
  return user?.role === 'admin' || user?.role === 'super_admin'
}

export function hasAclPermission(
  user: User | null | undefined,
  module: ACLModule,
  action: ACLAction
) {
  if (user?.role === 'super_admin') return true
  const actions = user?.permissions?.[module]
  if (!actions) return false
  if (actions.admin || actions[action]) return true
  return action === 'export' && Boolean(actions.read)
}

export function canAccessAdminModule(
  user: User | null | undefined,
  module: ACLModule,
  action: ACLAction = 'admin'
) {
  return isPlatformAdmin(user) || user?.profile_name === 'Administrator' || hasAclPermission(user, module, action)
}

export function isWorkspaceAdmin(user: User | null | undefined) {
  return (
    isPlatformAdmin(user)
    || user?.profile_name === 'Administrator'
    || hasAclPermission(user, 'users', 'admin')
    || hasAclPermission(user, 'settings', 'admin')
  )
}

export function canAccessAdminPath(user: User | null | undefined, path: string) {
  const module = adminRouteModules[path]
  return module ? canAccessAdminModule(user, module) : isWorkspaceAdmin(user)
}
