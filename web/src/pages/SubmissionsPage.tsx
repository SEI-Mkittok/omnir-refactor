import { Inbox, Link as LinkIcon } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { useSubmissions } from '@/hooks/useWebforms'

function formatDate(iso: string) {
  return new Date(iso).toLocaleString()
}

export function SubmissionsPage() {
  const submissions = useSubmissions()
  const rows = submissions.data?.data ?? []

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Submissions</h1>
          <p className="mt-1 text-sm text-[#6B7280]">Unified intake log for public webforms and campaign attribution.</p>
        </div>
        <div className="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm">
          <span className="block text-xs text-slate-500">Total</span>
          <span className="font-semibold text-slate-900">{submissions.data?.meta.total ?? rows.length}</span>
        </div>
      </div>

      <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
        <div className="grid border-b border-slate-100 bg-slate-50 px-5 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500 md:grid-cols-[1fr_0.8fr_1fr_1fr]">
          <span>Submission</span>
          <span>Target</span>
          <span>Created record</span>
          <span>Received</span>
        </div>
        {rows.length === 0 ? (
          <div className="flex items-center gap-2 px-5 py-10 text-sm text-slate-500">
            <Inbox className="h-4 w-4" />
            No submissions have been received yet.
          </div>
        ) : (
          rows.map((submission) => (
            <div key={submission.id} className="grid gap-2 border-b border-slate-100 px-5 py-4 text-sm md:grid-cols-[1fr_0.8fr_1fr_1fr]">
              <div className="min-w-0">
                <p className="truncate font-medium text-slate-900">{submission.id}</p>
                <p className="mt-1 truncate text-xs text-slate-500">{submission.ip_address || 'No IP'} · {submission.user_agent || 'No user agent'}</p>
              </div>
              <span><Badge variant="blue">{submission.target_module}</Badge></span>
              <div className="min-w-0">
                {submission.created_record_id ? (
                  <p className="flex items-center gap-1 truncate text-slate-700">
                    <LinkIcon className="h-3.5 w-3.5" />
                    {submission.created_record_type}: {submission.created_record_id}
                  </p>
                ) : (
                  <span className="text-slate-500">No record</span>
                )}
              </div>
              <span className="text-slate-600">{formatDate(submission.created_at)}</span>
            </div>
          ))
        )}
      </section>
    </div>
  )
}
