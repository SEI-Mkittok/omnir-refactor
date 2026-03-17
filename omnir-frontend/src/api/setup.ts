import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

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
  user: {
    id: string
    email: string
    name: string
    role: 'admin' | 'user' | 'viewer'
    created_at: string
    updated_at: string
  }
}

export async function getSetupStatus(): Promise<SetupStatusResponse> {
  const res = await axios.get(`${BASE_URL}/setup/status`)
  return res.data
}

export async function submitSetup(data: SetupRequest): Promise<SetupResponse> {
  const res = await axios.post(`${BASE_URL}/setup`, data)
  return res.data
}
