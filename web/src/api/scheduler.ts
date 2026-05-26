import apiClient from './client'
import type { SchedulerJob, SchedulerJobRun } from './types'

export const schedulerApi = {
  jobs(): Promise<{ data: SchedulerJob[] }> {
    return apiClient.get('/settings/scheduler/jobs').then((r) => r.data)
  },

  runs(jobKey: string): Promise<{ data: SchedulerJobRun[] }> {
    return apiClient.get(`/settings/scheduler/jobs/${jobKey}/runs`).then((r) => r.data)
  },
}
