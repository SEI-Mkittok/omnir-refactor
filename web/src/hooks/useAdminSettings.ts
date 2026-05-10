import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  adminSettingsApi,
  type CompanySettings,
  type ConfigEditorSettings,
  type MenuConfigSettings,
  type PortalSettings,
  type UpdateOutgoingServerSettingsRequest,
} from '@/api/adminSettings'

export const adminSettingsKeys = {
  base: ['admin-settings'] as const,
  company: () => [...adminSettingsKeys.base, 'company'] as const,
  portal: () => [...adminSettingsKeys.base, 'portal'] as const,
  outgoingServer: () => [...adminSettingsKeys.base, 'outgoing-server'] as const,
  configEditor: () => [...adminSettingsKeys.base, 'config-editor'] as const,
  menu: () => [...adminSettingsKeys.base, 'menu'] as const,
}

function useInvalidateOnSuccess() {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: adminSettingsKeys.base })
}

export function useCompanySettings() {
  return useQuery({
    queryKey: adminSettingsKeys.company(),
    queryFn: adminSettingsApi.getCompany,
  })
}

export function useUpdateCompanySettings() {
  const invalidate = useInvalidateOnSuccess()
  return useMutation({
    mutationFn: (payload: CompanySettings) => adminSettingsApi.updateCompany(payload),
    onSuccess: invalidate,
  })
}

export function usePortalSettings() {
  return useQuery({
    queryKey: adminSettingsKeys.portal(),
    queryFn: adminSettingsApi.getPortal,
  })
}

export function useUpdatePortalSettings() {
  const invalidate = useInvalidateOnSuccess()
  return useMutation({
    mutationFn: (payload: Partial<PortalSettings>) => adminSettingsApi.updatePortal(payload),
    onSuccess: invalidate,
  })
}

export function useOutgoingServerSettings() {
  return useQuery({
    queryKey: adminSettingsKeys.outgoingServer(),
    queryFn: adminSettingsApi.getOutgoingServer,
  })
}

export function useUpdateOutgoingServerSettings() {
  const invalidate = useInvalidateOnSuccess()
  return useMutation({
    mutationFn: (payload: UpdateOutgoingServerSettingsRequest) =>
      adminSettingsApi.updateOutgoingServer(payload),
    onSuccess: invalidate,
  })
}

export function useConfigEditorSettings() {
  return useQuery({
    queryKey: adminSettingsKeys.configEditor(),
    queryFn: adminSettingsApi.getConfigEditor,
  })
}

export function useUpdateConfigEditorSettings() {
  const invalidate = useInvalidateOnSuccess()
  return useMutation({
    mutationFn: (payload: ConfigEditorSettings) => adminSettingsApi.updateConfigEditor(payload),
    onSuccess: invalidate,
  })
}

export function useMenuConfigSettings(enabled = true) {
  return useQuery({
    queryKey: adminSettingsKeys.menu(),
    queryFn: adminSettingsApi.getMenuConfig,
    enabled,
  })
}

export function useUpdateMenuConfigSettings() {
  const invalidate = useInvalidateOnSuccess()
  return useMutation({
    mutationFn: (payload: MenuConfigSettings) => adminSettingsApi.updateMenuConfig(payload),
    onSuccess: invalidate,
  })
}
