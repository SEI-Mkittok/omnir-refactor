import apiClient from './client'
import type {
  EmailSequence,
  SequenceEnrollment,
  SequenceAnalytics,
  CreateSequenceRequest,
  UpdateSequenceRequest,
  EnrollRequest,
  SequenceStatus,
} from './types'

export interface ListSequencesParams {
  page?: number
  limit?: number
  status?: SequenceStatus
}

export const sequencesApi = {
  list(params?: ListSequencesParams): Promise<{ data: EmailSequence[]; total: number }> {
    return apiClient.get('/sequences', { params }).then((r) => r.data)
  },

  get(id: string): Promise<EmailSequence> {
    return apiClient.get(`/sequences/${id}`).then((r) => r.data)
  },

  create(req: CreateSequenceRequest): Promise<EmailSequence> {
    return apiClient.post('/sequences', req).then((r) => r.data)
  },

  update(id: string, req: UpdateSequenceRequest): Promise<EmailSequence> {
    return apiClient.patch(`/sequences/${id}`, req).then((r) => r.data)
  },

  delete(id: string): Promise<void> {
    return apiClient.delete(`/sequences/${id}`).then(() => undefined)
  },

  enroll(id: string, req: EnrollRequest): Promise<{ enrolled: number }> {
    return apiClient.post(`/sequences/${id}/enroll`, req).then((r) => r.data)
  },

  listEnrollments(id: string): Promise<{ data: SequenceEnrollment[] }> {
    return apiClient.get(`/sequences/${id}/enrollments`).then((r) => r.data)
  },

  updateEnrollment(sequenceId: string, enrollmentId: string, status: string): Promise<void> {
    return apiClient
      .patch(`/sequences/${sequenceId}/enrollments/${enrollmentId}`, { status })
      .then(() => undefined)
  },

  analytics(id: string): Promise<SequenceAnalytics> {
    return apiClient.get(`/sequences/${id}/analytics`).then((r) => r.data)
  },
}
