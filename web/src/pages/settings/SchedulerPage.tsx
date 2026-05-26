import { useEffect, useMemo, useState } from 'react'
import { CheckCircle, Clock, RefreshCw, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { useSchedulerJobs, useSchedulerRuns } from '@/hooks/useScheduler'

function formatDate(iso?: string) {
  if (!iso) return 'Not yet'
  return new Date(iso).toLocaleString()
}

function resultBadge(result?: string) {
  if (result === 'succeeded') return <Badge variant="green">Succeeded</Badge>
  if (result === 'failed') return <Badge variant="red">Failed</Badge>
  if (result === 'running') return <Badge variant="blue">Running</Badge>
  return <Badge variant="gray">No run</Badge>
}

export function SchedulerPage() {
  const [selected, setSelected] = useState<string>('')
  const jobs = useSchedulerJobs()
  const runs = useSchedulerRuns(selected)
  const rows = useMemo(() => jobs.data?.data ?? [], [jobs.data?.data])
  const selectedJob = rows.find((job) => job.key === selected) ?? rows[0]

  useEffect(() => {
    if (!selected && rows.length > 0) setSelected(rows[0].key)
  }, [rows, selected])

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Scheduler</h1>
          <p className="mt-1 text-sm text-[#6B7280]">Read-only status for background jobs and intake workers.</p>
        </div>
        <Button variant="outline" onClick={() => jobs.refetch()} disabled={jobs.isFetching}>
          <RefreshCw className={jobs.isFetching ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
          Refresh
        </Button>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="grid border-b border-slate-100 bg-slate-50 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500 md:grid-cols-[1.4fr_0.7fr_0.7fr_1fr_1fr]">
          <span>Job</span>
          <span>Interval</span>
          <span>State</span>
          <span>Last finish</span>
          <span>Next run</span>
        </div>
        {rows.length === 0 ? (
          <div className="px-4 py-10 text-sm text-slate-500">No scheduler jobs reported.</div>
        ) : (
          rows.map((job) => (
            <button
              key={job.key}
              onClick={() => setSelected(job.key)}
              className={`grid w-full gap-2 border-b border-slate-100 px-4 py-3 text-left text-sm transition-colors md:grid-cols-[1.4fr_0.7fr_0.7fr_1fr_1fr] ${
                selectedJob?.key === job.key ? 'bg-[#F4F7FA]' : 'hover:bg-slate-50'
              }`}
            >
              <span>
                <span className="font-medium text-slate-900">{job.name}</span>
                <span className="mt-0.5 block text-xs text-slate-500">{job.key}</span>
              </span>
              <span className="text-slate-700">{job.interval}</span>
              <span>{job.enabled ? <Badge variant="green">Enabled</Badge> : <Badge variant="gray">Disabled</Badge>}</span>
              <span className="text-slate-600">{formatDate(job.last_finished_at)}</span>
              <span className="text-slate-600">{formatDate(job.next_run_at)}</span>
            </button>
          ))
        )}
      </section>

      {selectedJob && (
        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h2 className="text-base font-semibold text-slate-900">{selectedJob.name}</h2>
              <p className="mt-1 text-sm text-slate-500">Last started: {formatDate(selectedJob.last_started_at)}</p>
            </div>
            {resultBadge(selectedJob.last_result)}
          </div>
          {selectedJob.error_text && (
            <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{selectedJob.error_text}</div>
          )}
          <div className="mt-5 divide-y divide-slate-100">
            {(runs.data?.data ?? []).length === 0 ? (
              <p className="py-4 text-sm text-slate-500">No recent run history for this job.</p>
            ) : (
              runs.data?.data.map((run) => (
                <div key={run.id} className="flex items-center gap-3 py-3 text-sm">
                  {run.status === 'succeeded' ? <CheckCircle className="h-4 w-4 text-green-500" /> : run.status === 'failed' ? <XCircle className="h-4 w-4 text-red-500" /> : <Clock className="h-4 w-4 text-blue-500" />}
                  <span className="font-medium text-slate-900">{run.status}</span>
                  <span className="text-slate-500">{formatDate(run.finished_at ?? run.started_at)}</span>
                  {run.error_text && <span className="text-red-600">{run.error_text}</span>}
                </div>
              ))
            )}
          </div>
        </section>
      )}
    </div>
  )
}
