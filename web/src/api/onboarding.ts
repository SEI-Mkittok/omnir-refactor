import apiClient from './client'

export type InviteRole = 'admin' | 'member' | 'viewer'

export interface OnboardingState {
  id: string
  completedSteps: string[]
  completed: boolean
  orgName?: string
}

export interface InviteRow {
  email: string
  role: InviteRole
}

export interface SlaConfig {
  critical_first_response: number
  critical_resolution: number
  high_first_response: number
  high_resolution: number
  medium_first_response: number
  medium_resolution: number
  low_first_response: number
  low_resolution: number
}

export interface OnboardingStatus {
  completed: boolean
  completedSteps: string[]
  invites: Array<{ email: string; role: string; accepted: boolean }>
  orgName?: string
}

export async function getOnboardingState(): Promise<OnboardingState> {
  const { data } = await apiClient.get<OnboardingState>('/onboarding')
  return data
}

export async function updateOnboarding(payload: {
  completedSteps?: string[]
  completed?: boolean
  orgName?: string
  sla?: SlaConfig
}): Promise<OnboardingState> {
  const { data } = await apiClient.patch<OnboardingState>('/onboarding', payload)
  return data
}

export async function sendInvite(invite: InviteRow): Promise<void> {
  await apiClient.post('/onboarding/invite', invite)
}

export async function getOnboardingStatus(): Promise<OnboardingStatus> {
  const { data } = await apiClient.get<OnboardingStatus>('/onboarding/status')
  return data
}
