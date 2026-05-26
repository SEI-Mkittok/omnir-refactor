import { useQuery } from '@tanstack/react-query'
import { schedulerApi } from '@/api/scheduler'

export function useSchedulerJobs() {
  return useQuery({
    queryKey: ['scheduler', 'jobs'],
    queryFn: () => schedulerApi.jobs(),
    staleTime: 15_000,
  })
}

export function useSchedulerRuns(jobKey: string) {
  return useQuery({
    queryKey: ['scheduler', 'runs', jobKey],
    queryFn: () => schedulerApi.runs(jobKey),
    enabled: !!jobKey,
    staleTime: 15_000,
  })
}
