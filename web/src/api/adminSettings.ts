import apiClient from './client'

export interface CompanySettings {
  company_name?: string
  company_logo_url?: string
  company_website?: string
  company_email?: string
  company_phone?: string
  company_address_line1?: string
  company_address_line2?: string
  company_city?: string
  company_state?: string
  company_postal_code?: string
  company_country?: string
}

export interface PortalSettings {
  portal_enabled: boolean
  portal_display_name?: string
  portal_announcement?: string
  portal_default_assignee_id?: string
  portal_menu: string[]
  portal_shortcuts: string[]
  portal_recent_widget_limit: number
}

export interface OutgoingServerSettings {
  smtp_host?: string
  smtp_port?: number
  smtp_username?: string
  smtp_from_email?: string
  smtp_from_name?: string
  smtp_security?: string
  smtp_auth_type?: string
  smtp_password_set: boolean
}

export interface UpdateOutgoingServerSettingsRequest {
  smtp_host?: string
  smtp_port?: number
  smtp_username?: string
  smtp_from_email?: string
  smtp_from_name?: string
  smtp_security?: string
  smtp_auth_type?: string
  smtp_password?: string
  clear_password?: boolean
}

export interface ConfigEditorSettings {
  config_support_email?: string
  config_upload_max_mb?: number
  config_default_page_size?: number
  config_list_preview_chars?: number
}

export interface MenuConfigSettings {
  menu_config: Record<string, boolean>
}

export const adminSettingsApi = {
  getCompany: async (): Promise<CompanySettings> => {
    const { data } = await apiClient.get('/settings/company')
    return data
  },

  updateCompany: async (payload: CompanySettings): Promise<CompanySettings> => {
    const { data } = await apiClient.patch('/settings/company', payload)
    return data
  },

  getPortal: async (): Promise<PortalSettings> => {
    const { data } = await apiClient.get('/settings/portal')
    return data
  },

  updatePortal: async (payload: Partial<PortalSettings>): Promise<PortalSettings> => {
    const { data } = await apiClient.patch('/settings/portal', payload)
    return data
  },

  getOutgoingServer: async (): Promise<OutgoingServerSettings> => {
    const { data } = await apiClient.get('/settings/outgoing-server')
    return data
  },

  updateOutgoingServer: async (
    payload: UpdateOutgoingServerSettingsRequest
  ): Promise<OutgoingServerSettings> => {
    const { data } = await apiClient.patch('/settings/outgoing-server', payload)
    return data
  },

  getConfigEditor: async (): Promise<ConfigEditorSettings> => {
    const { data } = await apiClient.get('/settings/config-editor')
    return data
  },

  updateConfigEditor: async (payload: ConfigEditorSettings): Promise<ConfigEditorSettings> => {
    const { data } = await apiClient.patch('/settings/config-editor', payload)
    return data
  },

  getMenuConfig: async (): Promise<MenuConfigSettings> => {
    const { data } = await apiClient.get('/settings/menu')
    return data
  },

  updateMenuConfig: async (payload: MenuConfigSettings): Promise<MenuConfigSettings> => {
    const { data } = await apiClient.patch('/settings/menu', payload)
    return data
  },
}
