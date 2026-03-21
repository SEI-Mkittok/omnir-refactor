import axios from 'axios'
import apiClient from './client'
import type { User } from './types'

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

export interface SSOConfig {
  enabled: boolean
  provider?: 'google' | 'microsoft' | 'oidc'
  orgSlug?: string
}

export interface TOTPSetupResult {
  otp_auth_uri: string
  qr_data_url: string
}

export interface TOTPVerifyResult {
  backup_codes: string[]
}

export const ssoApi = {
  getConfig: async (orgSlug: string): Promise<SSOConfig> => {
    const res = await axios.get(`${AUTH_BASE}/auth/sso/config`, {
      params: { orgSlug },
    })
    return res.data
  },

  /** Initiates SSO login by redirecting the browser to the OIDC provider. */
  beginLogin: (orgSlug: string) => {
    window.location.href = `${AUTH_BASE}/auth/sso/${orgSlug}/login`
  },

  /** Called after SSO callback — fetches the current user via cookie. */
  getMe: async (): Promise<User> => {
    const { data } = await apiClient.get('/users/me')
    return data
  },
}

export const totpApi = {
  /** Starts TOTP setup: returns QR data URL + otpauth URI. */
  setup: async (): Promise<TOTPSetupResult> => {
    const { data } = await apiClient.post('/users/me/2fa/setup')
    return data
  },

  /** Confirms TOTP code after scanning QR; activates 2FA and returns backup codes. */
  verify: async (code: string): Promise<TOTPVerifyResult> => {
    const { data } = await apiClient.post('/users/me/2fa/verify', { code })
    return data
  },

  /** Disables TOTP 2FA — requires the user's password for confirmation. */
  disable: async (password: string): Promise<void> => {
    await apiClient.post('/users/me/2fa/disable', { password })
  },

  /**
   * Verifies a TOTP code during the login flow (when requires_2fa=true).
   * Requires the pre_auth_token returned by the login endpoint.
   */
  verifyLogin: async (
    params: { code?: string; backup_code?: string },
    preAuthToken: string,
  ): Promise<User> => {
    const res = await axios.post(
      `${AUTH_BASE}/v1/auth/2fa/verify`,
      params,
      {
        headers: { Authorization: `Bearer ${preAuthToken}` },
        withCredentials: true,
      },
    )
    return res.data.user
  },
}
