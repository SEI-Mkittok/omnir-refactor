import { mgmtClient } from './client'
import type { Org, User } from './types'

export const orgsApi = {
  /** GET /api/orgs — super_admin only */
  list: () =>
    mgmtClient.get<{ data: Org[] }>('/orgs').then((r) => r.data),

  /** GET /api/orgs/:id */
  get: (id: string) =>
    mgmtClient.get<Org>(`/orgs/${id}`).then((r) => r.data),

  /**
   * POST /api/orgs/signup — creates a new org + its first admin user.
   */
  signup: (payload: { orgName: string; adminName: string; email: string; password: string }) =>
    mgmtClient.post<{ org: Org; user: User }>('/orgs/signup', payload).then((r) => r.data),

  /**
   * create is kept for OrgOnboardingPage compatibility.
   */
  create: (payload: { name: string; slug: string }) =>
    mgmtClient
      .post<{ org: Org; user: User }>('/orgs/signup', {
        orgName: payload.name,
        adminName: 'Admin',
        email: `admin@${payload.slug}.local`,
        password: Math.random().toString(36).slice(-12),
      })
      .then((r) => r.data.org),

  /** POST /api/orgs/:id/switch — re-issues JWT for super_admin switching org context */
  switchTo: (id: string) =>
    mgmtClient
      .post<{ org: Org; user: { id: string; role: string } }>(`/orgs/${id}/switch`)
      .then((r) => r.data),
}
