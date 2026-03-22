/**
 * TOTP / SSO auth API
 *
 * Backend endpoints are stubbed with mock responses where not yet implemented.
 * Remove stubs once Tyr's OMN-505 backend lands on develop.
 */

import apiClient from './client'
import type { User } from './types'

// ---- Types ----

export interface TotpSetupResponse {
  qr_url: string
  secret: string
}

export interface TotpBackupCodesResponse {
  codes: string[]
}

export interface TotpVerifyResponse {
  ok: boolean
  error?: string
  user?: User
}

// ---- TOTP enrollment ----

export async function getTotpSetup(): Promise<TotpSetupResponse> {
  try {
    const { data } = await apiClient.get('/auth/totp/setup')
    return data
  } catch {
    // Stub: backend not yet available
    return {
      qr_url: 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=otpauth://totp/Omnir%3Auser@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Omnir',
      secret: 'JBSWY3DPEHPK3PXP',
    }
  }
}

export async function verifyTotpSetup(code: string): Promise<TotpVerifyResponse> {
  try {
    const { data } = await apiClient.post('/auth/totp/verify-setup', { code })
    return data
  } catch {
    // Stub: accept any 6-digit code
    if (code.length === 6) return { ok: true }
    return { ok: false, error: 'Invalid code' }
  }
}

export async function getTotpBackupCodes(): Promise<TotpBackupCodesResponse> {
  try {
    const { data } = await apiClient.get('/auth/totp/backup-codes')
    return data
  } catch {
    // Stub
    return {
      codes: [
        'a1b2-c3d4',
        'e5f6-g7h8',
        'i9j0-k1l2',
        'm3n4-o5p6',
        'q7r8-s9t0',
        'u1v2-w3x4',
        'y5z6-a7b8',
        'c9d0-e1f2',
      ],
    }
  }
}

// ---- TOTP verification at login ----

export async function verifyTotp(code: string): Promise<TotpVerifyResponse> {
  try {
    const { data } = await apiClient.post('/auth/totp/verify', { code })
    return data
  } catch {
    return { ok: false, error: 'Invalid code' }
  }
}

export async function verifyTotpBackup(code: string): Promise<TotpVerifyResponse> {
  try {
    const { data } = await apiClient.post('/auth/totp/verify-backup', { code })
    return data
  } catch {
    return { ok: false, error: 'Invalid backup code' }
  }
}
