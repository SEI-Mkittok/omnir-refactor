import axios from 'axios'

type ApiErrorDetail = {
  field?: string
  message?: string
}

type ApiErrorPayload = {
  error?: string
  code?: string
  details?: ApiErrorDetail[]
}

export type CrmLinkErrorKind = 'account_contact_mismatch' | 'unknown'

export type CrmLinkError = {
  kind: CrmLinkErrorKind
  message: string
  code?: string
  details: ApiErrorDetail[]
}

export function mapCrmLinkError(error: unknown): CrmLinkError {
  if (!axios.isAxiosError(error)) {
    return {
      kind: 'unknown',
      message: 'Could not complete link update. Please try again.',
      details: [],
    }
  }

  const payload = (error.response?.data ?? {}) as ApiErrorPayload
  const message = payload.error ?? error.message ?? 'Could not complete link update. Please try again.'
  const details = Array.isArray(payload.details) ? payload.details : []
  const code = payload.code
  const normalized = `${message} ${details.map((detail) => `${detail.field ?? ''} ${detail.message ?? ''}`).join(' ')}`.toLowerCase()

  if (
    (code === 'validation_error' || error.response?.status === 422) &&
    (
      normalized.includes('contact_id is not related to account_id') ||
      normalized.includes('contact_id is not related to deal_id account')
    )
  ) {
    return {
      kind: 'account_contact_mismatch',
      message,
      code,
      details,
    }
  }

  return {
    kind: 'unknown',
    message,
    code,
    details,
  }
}
