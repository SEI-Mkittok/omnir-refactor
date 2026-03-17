import axios from 'axios'
import type { User } from './types'

const BASE_URL = import.meta.env.VITE_AUTH_URL || 'http://localhost:8080/api'

export interface SetupStatusResponse {
  setupRequired: boolean
}

export interface SetupRequest {
  companyName: string
  adminName: string
  email: string
  password: string
}

export interface SetupResponse {
  access_token: string
  refresh_token: string
  user: User
}

export async function getSetupStatus(): Promise<SetupStatusResponse> {
  const res = await axios.get(`${BASE_URL}/setup/status`)
  return res.data
}

export async function submitSetup(data: SetupRequest): Promise<SetupResponse> {
  const res = await axios.post(`${BASE_URL}/setup`, data)
  return res.data
}
